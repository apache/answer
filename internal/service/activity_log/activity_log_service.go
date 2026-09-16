/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package activity_log

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/queue"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/comment_common"
	"github.com/apache/answer/internal/service/eventqueue"
	"github.com/apache/answer/internal/service/object_info"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

// Action keys. Content actions follow the event naming (object.action); the rest are ours.
const (
	ActionUserLogin        = "user.login"
	ActionUserUpdate       = "user.update"
	ActionUserShare        = "user.share"
	ActionUserRole         = "user.role"
	ActionUserSuspend      = "user.suspend"
	ActionUserUnsuspend    = "user.unsuspend"
	ActionUserDelete       = "user.delete"
	ActionPageView         = "page.view"
	ActionBadgeAward       = "badge.award"
	ActionBadgeRevoke      = "badge.revoke"
	ActionReviewQueued     = "review.queued"
	ActionReviewApprove    = "review.approve"
	ActionReviewReject     = "review.reject"
	ActionReputationChange = "reputation.change"
	ActionCommentReply     = "comment.reply"
)

// UserSystem is the actor of automatic actions (badge rules, reviewer bot)
const UserSystem = "0"

// KnownActions every action key the log can contain (labels come from i18n ui.admin.activity_log.action.*)
var KnownActions = []string{
	ActionUserLogin, ActionUserUpdate, ActionUserShare, ActionUserRole, ActionUserSuspend, ActionUserUnsuspend, ActionUserDelete,
	ActionPageView, ActionBadgeAward, ActionBadgeRevoke, ActionReviewQueued, ActionReviewApprove, ActionReviewReject,
	ActionReputationChange, ActionCommentReply,
	"question.create", "question.update", "question.delete", "question.vote_up", "question.vote_down", "question.vote_cancel",
	"question.accept", "question.flag", "question.react",
	"answer.create", "answer.update", "answer.delete", "answer.vote_up", "answer.vote_down", "answer.vote_cancel",
	"answer.flag", "answer.react",
	"comment.create", "comment.update", "comment.delete", "comment.vote_up", "comment.vote_down", "comment.vote_cancel",
	"comment.flag",
}

const (
	pageViewDedupWindow = 10 * time.Second
	excerptLen          = 160
)

// Query filters for listing / export
type Query struct {
	From, To     time.Time
	UserIDs      []string // actor or target
	Actions      []string
	Text         string   // free text: matched against detail/action/ip …
	TextUserIDs  []string // … and against users whose name/username matched Text (resolved by the service)
	TextObjectID string   // … and against object/question id when Text looks like an id
}

// ActivityLogRepo persistence
type ActivityLogRepo interface {
	Add(ctx context.Context, entry *entity.ActivityLog) error
	ExistsSourceRef(ctx context.Context, sourceRef string) (bool, error)
	Page(ctx context.Context, q *Query, page, pageSize int) ([]*entity.ActivityLog, int64, error)
	Iterate(ctx context.Context, q *Query, limit int, fn func(rows []*entity.ActivityLog) bool) error
	ActionCounts(ctx context.Context, q *Query) (map[string]int64, error)
	DeletePageViewsBefore(ctx context.Context, t time.Time) (int64, error)
	// SearchUserIDs ids of users whose username / display name / e-mail contains text
	SearchUserIDs(ctx context.Context, text string, limit int) ([]string, error)
}

// Detail is the free-form part of an entry (stored as JSON)
type Detail map[string]any

// ActivityLogService records what users do and serves the admin log
type ActivityLogService struct {
	repo          ActivityLogRepo
	objectService *object_info.ObjService
	userCommon    *usercommon.UserCommon
	commentRepo   comment_common.CommentCommonRepo
	writeQueue    queue.Service[*entity.ActivityLog]

	pvMu       sync.Mutex
	pvLastSeen map[string]time.Time // userID+path → last page view (dedup)
}

// NewActivityLogService new service; subscribes to the event queue next to the badge handler.
// Object / user lookups are attached later (Attach) because the repos that log into this
// service (user, rank) sit below object_info / user_common in the dependency graph.
func NewActivityLogService(
	repo ActivityLogRepo,
	commentRepo comment_common.CommentCommonRepo,
	eventQueueService eventqueue.Service,
) *ActivityLogService {
	s := &ActivityLogService{
		repo:        repo,
		commentRepo: commentRepo,
		writeQueue:  queue.New[*entity.ActivityLog]("activity_log", 1024),
		pvLastSeen:  make(map[string]time.Time),
	}
	s.writeQueue.RegisterHandler(func(ctx context.Context, entry *entity.ActivityLog) error {
		return s.repo.Add(ctx, entry)
	})
	eventQueueService.RegisterHandler(s.handleEvent)
	return s
}

// Attach wires the lookups used to enrich entries (called once by the admin service)
func (s *ActivityLogService) Attach(objectService *object_info.ObjService, userCommon *usercommon.UserCommon) {
	s.objectService = objectService
	s.userCommon = userCommon
}

// Log records an entry asynchronously. IP and user agent are taken from the gin context when available.
func (s *ActivityLogService) Log(ctx context.Context, entry *entity.ActivityLog) {
	if s == nil || entry == nil || entry.Action == "" {
		return
	}
	if ginCtx, ok := ctx.(*gin.Context); ok && ginCtx != nil && ginCtx.Request != nil {
		if entry.IP == "" {
			entry.IP = ginCtx.ClientIP()
		}
		if entry.UserAgent == "" {
			entry.UserAgent = truncate(ginCtx.Request.UserAgent(), 512)
		}
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	// bigint columns: PostgreSQL rejects "" (SQLite does not)
	for _, f := range []*string{&entry.UserID, &entry.ObjectID, &entry.QuestionID, &entry.AnswerID, &entry.TargetUserID} {
		if *f == "" {
			*f = "0"
		}
	}
	s.writeQueue.Send(context.Background(), entry)
}

// LogDetail is Log with a detail map
func (s *ActivityLogService) LogDetail(ctx context.Context, entry *entity.ActivityLog, detail Detail) {
	if s == nil || entry == nil {
		return
	}
	entry.Detail = encodeDetail(detail)
	s.Log(ctx, entry)
}

// LogPageView records a page view, skipping repeats of the same path within a short window
func (s *ActivityLogService) LogPageView(ctx context.Context, userID, path, title string) (recorded bool) {
	if userID == "" || path == "" {
		return false
	}
	key := userID + "|" + path
	now := time.Now()
	s.pvMu.Lock()
	last, seen := s.pvLastSeen[key]
	if seen && now.Sub(last) < pageViewDedupWindow {
		s.pvMu.Unlock()
		return false
	}
	s.pvLastSeen[key] = now
	if len(s.pvLastSeen) > 10000 { // keep the dedup map bounded
		for k, t := range s.pvLastSeen {
			if now.Sub(t) > pageViewDedupWindow {
				delete(s.pvLastSeen, k)
			}
		}
	}
	s.pvMu.Unlock()

	entry := &entity.ActivityLog{UserID: userID, Action: ActionPageView, ObjectType: "page", ObjectID: "0"}
	// link the view to the object when the path is a question/user/tag page
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		switch parts[0] {
		case "questions":
			entry.ObjectType = constant.QuestionObjectType
			entry.ObjectID = parts[1]
			entry.QuestionID = parts[1]
			if len(parts) >= 4 {
				entry.AnswerID = parts[3]
			}
		case "users":
			entry.ObjectType = constant.UserObjectType
		case "tags":
			entry.ObjectType = constant.TagObjectType
		}
	}
	s.LogDetail(ctx, entry, Detail{"path": path, "title": truncate(title, 200)})
	return true
}

// CleanupPageViews deletes page views older than the retention period
func (s *ActivityLogService) CleanupPageViews(ctx context.Context, days int) {
	if days <= 0 {
		return
	}
	n, err := s.repo.DeletePageViewsBefore(ctx, time.Now().AddDate(0, 0, -days))
	if err != nil {
		log.Errorf("activity_log: cleanup page views: %v", err)
		return
	}
	if n > 0 {
		log.Infof("activity_log: removed %d page views older than %d days", n, days)
	}
}

// handleEvent maps an event from the shared event queue to a log entry
func (s *ActivityLogService) handleEvent(ctx context.Context, msg *schema.EventMsg) error {
	if msg == nil {
		return nil
	}
	entry := &entity.ActivityLog{UserID: msg.UserID, Action: string(msg.EventType), ObjectID: msg.TriggerObjectID,
		QuestionID: msg.QuestionID, AnswerID: msg.AnswerID}
	detail := Detail{}
	if msg.EventType == constant.EventUserUpdate || msg.EventType == constant.EventUserShare {
		entry.ObjectType = constant.UserObjectType
		entry.ObjectID = msg.UserID
		for k, v := range msg.ExtraInfo {
			detail[k] = v
		}
		s.LogDetail(ctx, entry, detail)
		return nil
	}

	objectType, _, _ := strings.Cut(string(msg.EventType), ".")
	entry.ObjectType = objectType
	if entry.ObjectID == "" {
		switch objectType {
		case constant.QuestionObjectType:
			entry.ObjectID = msg.QuestionID
		case constant.AnswerObjectType:
			entry.ObjectID = msg.AnswerID
		case constant.CommentObjectType:
			entry.ObjectID = msg.CommentID
		}
	}
	// the object author is the "target" of votes, flags, reactions, accepts …
	switch objectType {
	case constant.QuestionObjectType:
		entry.TargetUserID = msg.QuestionUserID
	case constant.AnswerObjectType:
		entry.TargetUserID = msg.AnswerUserID
	case constant.CommentObjectType:
		entry.TargetUserID = msg.CommentUserID
	}
	if entry.TargetUserID == entry.UserID {
		entry.TargetUserID = ""
	}

	if info := s.objectInfo(ctx, entry.ObjectID); info != nil {
		if entry.QuestionID == "" {
			entry.QuestionID = info.QuestionID
		}
		if entry.AnswerID == "" {
			entry.AnswerID = info.AnswerID
		}
		if info.Title != "" {
			detail["title"] = truncate(info.Title, 200)
		}
		if objectType != constant.QuestionObjectType || msg.EventType == constant.EventQuestionCreate ||
			msg.EventType == constant.EventQuestionUpdate {
			if ex := excerpt(info.Content); ex != "" {
				detail["excerpt"] = ex
			}
		}
		if entry.TargetUserID == "" && info.ObjectCreatorUserID != entry.UserID {
			entry.TargetUserID = info.ObjectCreatorUserID
		}
	}

	// votes: direction / cancel come from the extra info added by the vote service (added with the activity log)
	if strings.HasSuffix(string(msg.EventType), ".vote") {
		dir := msg.ExtraInfo["vote_direction"]
		if dir == "" {
			dir = "up"
		}
		if msg.ExtraInfo["vote_cancel"] == "1" {
			entry.Action = objectType + ".vote_cancel"
			detail["direction"] = dir
		} else {
			entry.Action = objectType + ".vote_" + dir
		}
		for _, k := range []string{"vote_up_amount", "vote_down_amount"} {
			if v, ok := msg.ExtraInfo[k]; ok {
				detail[k] = v
			}
		}
	}
	// a comment answering another comment is logged as a reply
	if msg.EventType == constant.EventCommentCreate && msg.CommentID != "" && s.commentRepo != nil {
		if c, exist, err := s.commentRepo.GetComment(ctx, msg.CommentID); err == nil && exist && c.GetReplyCommentID() != "" && c.GetReplyCommentID() != "0" {
			entry.Action = ActionCommentReply
			detail["reply_to"] = c.GetReplyCommentID()
			// the person being answered is the target, not the post author
			if parent, ok, err := s.commentRepo.GetComment(ctx, c.GetReplyCommentID()); err == nil && ok && parent.UserID != entry.UserID {
				entry.TargetUserID = parent.UserID
			}
		}
	}
	for k, v := range msg.ExtraInfo {
		if _, done := detail[k]; !done && k != "vote_direction" && k != "vote_cancel" {
			detail[k] = v
		}
	}
	s.LogDetail(ctx, entry, detail)
	return nil
}

func (s *ActivityLogService) objectInfo(ctx context.Context, objectID string) *schema.SimpleObjectInfo {
	if s.objectService == nil || objectID == "" || objectID == "0" {
		return nil
	}
	info, err := s.objectService.GetInfo(ctx, objectID)
	if err != nil {
		log.Debugf("activity_log: object %s: %v", objectID, err)
		return nil
	}
	return info
}

func encodeDetail(d Detail) string {
	if len(d) == 0 {
		return ""
	}
	b, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(b)
}

// DecodeDetail parses the stored JSON detail (empty map on error)
func DecodeDetail(s string) Detail {
	d := Detail{}
	if s == "" {
		return d
	}
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return Detail{"raw": s}
	}
	return d
}

func excerpt(content string) string {
	text := strings.TrimSpace(htmltext.ClearText(content))
	return truncate(text, excerptLen)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
