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
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/log"
)

// ExportLimit max rows in one TSV export
const ExportLimit = 100000

// ActivityLogAdminService reads the log for the admin UI. Its constructor also attaches the
// object / user lookups to the writer (see NewActivityLogService).
type ActivityLogAdminService struct {
	logService    *ActivityLogService
	repo          ActivityLogRepo
	objectService *object_info.ObjService
	userCommon    *usercommon.UserCommon
	configService *config.ConfigService
}

// NewActivityLogAdminService new admin service
func NewActivityLogAdminService(
	logService *ActivityLogService,
	repo ActivityLogRepo,
	objectService *object_info.ObjService,
	userCommon *usercommon.UserCommon,
	configService *config.ConfigService,
) *ActivityLogAdminService {
	logService.Attach(objectService, userCommon)
	return &ActivityLogAdminService{logService: logService, repo: repo, objectService: objectService,
		userCommon: userCommon, configService: configService}
}

// buildQuery turns the request into repo filters (usernames / free text resolved to user ids)
func (s *ActivityLogAdminService) buildQuery(ctx context.Context, req *schema.ActivityLogPageReq) (*Query, error) {
	q := &Query{}
	if req.From > 0 {
		q.From = time.Unix(req.From, 0)
	}
	if req.To > 0 {
		q.To = time.Unix(req.To, 0)
	}
	if req.Username != "" {
		info, exist, err := s.userCommon.GetUserBasicInfoByUserName(ctx, strings.TrimSpace(req.Username))
		if err != nil {
			return nil, err
		}
		if !exist {
			q.UserIDs = []string{"-1"} // unknown user → no rows
		} else {
			q.UserIDs = []string{info.ID}
		}
	}
	if req.Action != "" {
		for _, a := range strings.Split(req.Action, ",") {
			if a = strings.TrimSpace(a); a != "" {
				q.Actions = append(q.Actions, a)
			}
		}
	}
	if text := strings.TrimSpace(req.Q); text != "" {
		q.Text = text
		ids, err := s.repo.SearchUserIDs(ctx, text, 50)
		if err != nil {
			return nil, err
		}
		q.TextUserIDs = ids
		if id := uid.DeShortID(text); isNumeric(id) {
			q.TextObjectID = id
		}
	}
	return q, nil
}

// Page one page of enriched rows
func (s *ActivityLogAdminService) Page(ctx context.Context, req *schema.ActivityLogPageReq) (*pager.PageModel, error) {
	q, err := s.buildQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repo.Page(ctx, q, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := s.decorate(ctx, rows)
	return pager.NewPageModel(total, items), nil
}

// Actions action keys with counts for the filter
func (s *ActivityLogAdminService) Actions(ctx context.Context, req *schema.ActivityLogPageReq) ([]*schema.ActivityLogActionCount, error) {
	q, err := s.buildQuery(ctx, &schema.ActivityLogPageReq{From: req.From, To: req.To})
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.ActionCounts(ctx, q)
	if err != nil {
		return nil, err
	}
	out := make([]*schema.ActivityLogActionCount, 0, len(counts))
	for k, v := range counts {
		out = append(out, &schema.ActivityLogActionCount{Action: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Action < out[j].Action })
	return out, nil
}

// PageView records a page view sent by the UI
func (s *ActivityLogAdminService) PageView(ctx context.Context, req *schema.ActivityLogPageViewReq) {
	s.logService.LogPageView(ctx, req.UserID, req.Path, req.Title)
}

// Daily counts per local day for the dashboard (today first)
func (s *ActivityLogAdminService) Daily(ctx context.Context, req *schema.ActivityLogDailyReq) ([]*schema.ActivityLogDailyRow, error) {
	days := req.Days
	if days <= 0 {
		days = 14
	}
	loc := time.FixedZone("browser", req.TzOffset*60)
	today := time.Now().In(loc)
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(days - 1))
	rows, err := s.repo.ListActionsSince(ctx, start)
	if err != nil {
		return nil, err
	}
	byDay := map[string]*schema.ActivityLogDailyRow{}
	out := make([]*schema.ActivityLogDailyRow, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := start.AddDate(0, 0, i)
		row := &schema.ActivityLogDailyRow{Date: d.Format("2006-01-02"), From: d.Unix(), To: d.AddDate(0, 0, 1).Unix(), Actions: map[string]int64{}}
		byDay[row.Date] = row
		out = append(out, row)
	}
	for _, r := range rows {
		row, ok := byDay[r.CreatedAt.In(loc).Format("2006-01-02")]
		if !ok {
			continue
		}
		row.Actions[r.Action]++
		switch r.Action {
		case "question.create":
			row.Questions++
		case "answer.create":
			row.Answers++
		case "comment.create", ActionCommentReply:
			row.Comments++
		case ActionReviewQueued:
			row.Reviews++
		case ActionBadgeAward:
			row.Badges++
		case ActionPageView:
			row.Views++
		case ActionUserLogin:
			row.Logins++
		}
	}
	return out, nil
}

// decorate resolves users and object titles for a batch of rows
func (s *ActivityLogAdminService) decorate(ctx context.Context, rows []*entity.ActivityLog) []*schema.ActivityLogItem {
	userIDs := make([]string, 0, len(rows)*2)
	seen := map[string]bool{}
	for _, r := range rows {
		for _, id := range []string{r.UserID, r.TargetUserID} {
			if id != "" && id != "0" && !seen[id] {
				seen[id] = true
				userIDs = append(userIDs, id)
			}
		}
	}
	users := map[string]*schema.UserBasicInfo{}
	if len(userIDs) > 0 {
		if m, err := s.userCommon.BatchUserBasicInfoByID(ctx, userIDs); err == nil {
			users = m
		} else {
			log.Error(err)
		}
	}
	toUser := func(id string, actor bool) *schema.ActivityLogUser {
		if id == "" || (id == UserSystem && !actor) {
			return nil // no target
		}
		if id == UserSystem {
			return &schema.ActivityLogUser{ID: UserSystem, Username: "system", DisplayName: "System"}
		}
		if u, ok := users[id]; ok && u != nil {
			return &schema.ActivityLogUser{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, Avatar: u.Avatar, Status: u.Status}
		}
		return &schema.ActivityLogUser{ID: id, Username: "", DisplayName: "#" + id}
	}

	items := make([]*schema.ActivityLogItem, 0, len(rows))
	for _, r := range rows {
		detail := DecodeDetail(r.Detail)
		item := &schema.ActivityLogItem{
			ID: r.ID, CreatedAt: r.CreatedAt.Unix(), Action: r.Action, User: toUser(r.UserID, true), Target: toUser(r.TargetUserID, false),
			ObjectType: r.ObjectType, ObjectID: zeroToEmpty(r.ObjectID), QuestionID: zeroToEmpty(r.QuestionID), AnswerID: zeroToEmpty(r.AnswerID),
			RankDelta: r.RankDelta, Detail: detail, IP: r.IP,
		}
		if t, ok := detail["title"].(string); ok {
			item.Title = t
		}
		// reputation rows store the activity config id (written inside a transaction) → resolve the key here
		if id, ok := detail["activity_type"].(float64); ok && id > 0 && s.configService != nil {
			if cfg, err := s.configService.GetConfigByID(ctx, int(id)); err == nil && cfg != nil {
				detail["activity"] = cfg.Key
			}
			delete(detail, "activity_type")
		}
		// reputation rows carry only the object id → look the title up
		if item.Title == "" && item.ObjectID != "" && r.ObjectType != "page" && r.ObjectType != "user" && r.ObjectType != "badge_award" && s.objectService != nil {
			if info, err := s.objectService.GetInfo(ctx, item.ObjectID); err == nil && info != nil {
				item.Title = info.Title
				if item.QuestionID == "" {
					item.QuestionID = info.QuestionID
				}
				if item.AnswerID == "" {
					item.AnswerID = info.AnswerID
				}
			}
		}
		if item.Title != "" {
			item.UrlTitle = htmltext.UrlTitle(item.Title)
		}
		items = append(items, item)
	}
	return items
}

// Export writes matching rows as TSV (UTF-8 with BOM so Excel opens it correctly)
func (s *ActivityLogAdminService) Export(ctx context.Context, req *schema.ActivityLogPageReq, labels map[string]string, w io.Writer) error {
	q, err := s.buildQuery(ctx, req)
	if err != nil {
		return err
	}
	if _, err = io.WriteString(w, "\ufeff"); err != nil {
		return err
	}
	header := []string{"czas", "login", "użytkownik", "akcja", "akcja_opis", "typ_obiektu", "id_obiektu", "tytuł", "cel_login", "cel_użytkownik", "punkty", "szczegóły", "ip"}
	if _, err = io.WriteString(w, strings.Join(header, "\t")+"\r\n"); err != nil {
		return err
	}
	var writeErr error
	err = s.repo.Iterate(ctx, q, ExportLimit, func(rows []*entity.ActivityLog) bool {
		for _, it := range s.decorate(ctx, rows) {
			// every column except the numeric delta is user-influenced text → escape for spreadsheets
			cols := []string{
				time.Unix(it.CreatedAt, 0).Format("2006-01-02 15:04:05"),
				tsvText(userCol(it.User, true)), tsvText(userCol(it.User, false)),
				tsvText(it.Action), tsvText(labels[it.Action]),
				tsvText(it.ObjectType), tsvText(it.ObjectID), tsvText(it.Title),
				tsvText(userCol(it.Target, true)), tsvText(userCol(it.Target, false)),
				fmt.Sprintf("%d", it.RankDelta),
				tsvText(detailText(it.Detail)), tsvText(it.IP),
			}
			if _, writeErr = io.WriteString(w, strings.Join(cols, "\t")+"\r\n"); writeErr != nil {
				return false
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	return writeErr
}

// zeroToEmpty "0" (no object) → "" so the UI does not build links to /questions/0
func zeroToEmpty(id string) string {
	if id == "0" {
		return ""
	}
	return id
}

func userCol(u *schema.ActivityLogUser, login bool) string {
	if u == nil {
		return ""
	}
	if login {
		return u.Username
	}
	return u.DisplayName
}

// detailText flattens the detail map to "k=v; k=v" (title is its own column)
func detailText(d map[string]any) string {
	keys := make([]string, 0, len(d))
	for k := range d {
		if k != "title" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, d[k]))
	}
	return strings.Join(parts, "; ")
}

// tsvText makes a text cell safe for a TSV opened in a spreadsheet: no tabs / line breaks, and a
// leading apostrophe when the value would otherwise be interpreted as a formula (=, +, -, @, or a
// tab / CR that some spreadsheets strip before parsing) — CSV/TSV formula injection.
func tsvText(s string) string {
	s = strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(s)
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '|', '%':
		return "'" + s
	}
	return s
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
