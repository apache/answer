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

package tag_common

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_queue"
	"github.com/apache/answer/internal/service/revision_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type TagCommonRepo interface {
	AddTagList(ctx context.Context, tagList []*entity.Tag) (err error)
	GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity.Tag, err error)
	GetTagBySlugName(ctx context.Context, slugName string) (tagInfo *entity.Tag, exist bool, err error)
	GetTagListByName(ctx context.Context, name string, recommend, reserved bool) (tagList []*entity.Tag, err error)
	GetTagListByNames(ctx context.Context, names []string) (tagList []*entity.Tag, err error)
	GetTagByID(ctx context.Context, tagID string, includeDeleted bool) (tag *entity.Tag, exist bool, err error)
	GetTagPage(ctx context.Context, page, pageSize int, tag *entity.Tag, queryCond string) (tagList []*entity.Tag, total int64, err error)
	GetRecommendTagList(ctx context.Context) (tagList []*entity.Tag, err error)
	GetReservedTagList(ctx context.Context) (tagList []*entity.Tag, err error)
	UpdateTagsAttribute(ctx context.Context, tags []string, attribute string, value bool) (err error)
	UpdateTagQuestionCount(ctx context.Context, tagID string, questionCount int) (err error)
}

type TagRepo interface {
	RemoveTag(ctx context.Context, tagID string) (err error)
	UpdateTag(ctx context.Context, tag *entity.Tag) (err error)
	RecoverTag(ctx context.Context, tagID string) (err error)
	MustGetTagByNameOrID(ctx context.Context, tagID, slugName string) (tag *entity.Tag, exist bool, err error)
	UpdateTagSynonym(ctx context.Context, tagSlugNameList []string, mainTagID int64, mainTagSlugName string) (err error)
	GetTagSynonymCount(ctx context.Context, tagID string) (count int64, err error)
	GetIDsByMainTagId(ctx context.Context, mainTagID string) (tagIDs []string, err error)
	GetTagList(ctx context.Context, tag *entity.Tag) (tagList []*entity.Tag, err error)
}

type TagRelRepo interface {
	AddTagRelList(ctx context.Context, tagList []*entity.TagRel) (err error)
	RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	HideTagRelListByObjectID(ctx context.Context, objectID string) (err error)
	RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error)
	EnableTagRelByIDs(ctx context.Context, ids []int64, hide bool) (err error)
	GetObjectTagRelWithoutStatus(ctx context.Context, objectId, tagID string) (tagRel *entity.TagRel, exist bool, err error)
	GetObjectTagRelList(ctx context.Context, objectId string) (tagListList []*entity.TagRel, err error)
	BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity.TagRel, err error)
	CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error)
	GetTagRelDefaultStatusByObjectID(ctx context.Context, objectID string) (status int, err error)
	MigrateTagObjects(ctx context.Context, sourceTagId, targetTagId string) error
}

// TagCommonService user service
type TagCommonService struct {
	revisionService      *revision_common.RevisionService
	tagCommonRepo        TagCommonRepo
	tagRelRepo           TagRelRepo
	tagRepo              TagRepo
	siteInfoService      siteinfo_common.SiteInfoCommonService
	activityQueueService activity_queue.ActivityQueueService
}

// NewTagCommonService new tag service
func NewTagCommonService(
	tagCommonRepo TagCommonRepo,
	tagRelRepo TagRelRepo,
	tagRepo TagRepo,
	revisionService *revision_common.RevisionService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	activityQueueService activity_queue.ActivityQueueService,
) *TagCommonService {
	return &TagCommonService{
		tagCommonRepo:        tagCommonRepo,
		tagRelRepo:           tagRelRepo,
		tagRepo:              tagRepo,
		revisionService:      revisionService,
		siteInfoService:      siteInfoService,
		activityQueueService: activityQueueService,
	}
}

// SearchTagLike get tag list all
func (ts *TagCommonService) SearchTagLike(ctx context.Context, req *schema.SearchTagLikeReq) (resp []schema.GetTagBasicResp, err error) {
	tags, err := ts.tagCommonRepo.GetTagListByName(ctx, req.Tag, len(req.Tag) == 0, false)
	if err != nil {
		return
	}
	ts.TagsFormatRecommendAndReserved(ctx, tags)
	mainTagId := make([]string, 0)
	for _, tag := range tags {
		if tag.MainTagID != 0 {
			mainTagId = append(mainTagId, converter.IntToString(tag.MainTagID))
		}
	}
	mainTagMap := make(map[string]*entity.Tag)
	if len(mainTagId) > 0 {
		mainTagList, err := ts.tagCommonRepo.GetTagListByIDs(ctx, mainTagId)
		if err != nil {
			return nil, err
		}
		for _, tag := range mainTagList {
			mainTagMap[tag.ID] = tag
		}
	}
	for _, tag := range tags {
		if tag.MainTagID == 0 {
			continue
		}
		mainTagID := converter.IntToString(tag.MainTagID)
		if _, ok := mainTagMap[mainTagID]; ok {
			tag.ID = mainTagMap[mainTagID].ID
			tag.SlugName = mainTagMap[mainTagID].SlugName
			tag.DisplayName = mainTagMap[mainTagID].DisplayName
			tag.Reserved = mainTagMap[mainTagID].Reserved
			tag.Recommend = mainTagMap[mainTagID].Recommend
		}
	}
	resp = make([]schema.GetTagBasicResp, 0)
	repetitiveTag := make(map[string]bool)
	for _, tag := range tags {
		if _, ok := repetitiveTag[tag.SlugName]; !ok {
			item := schema.GetTagBasicResp{}
			item.TagID = tag.ID
			item.SlugName = tag.SlugName
			item.DisplayName = tag.DisplayName
			item.Recommend = tag.Recommend
			item.Reserved = tag.Reserved
			resp = append(resp, item)
			repetitiveTag[tag.SlugName] = true
		}
	}
	return resp, nil
}

func (ts *TagCommonService) GetSiteWriteRecommendTag(ctx context.Context) (tags []*schema.SiteWriteTag, err error) {
	tags = make([]*schema.SiteWriteTag, 0)
	list, err := ts.tagCommonRepo.GetRecommendTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, &schema.SiteWriteTag{
			SlugName:    item.SlugName,
			DisplayName: item.DisplayName,
		})
	}
	return tags, nil
}

func (ts *TagCommonService) GetSiteWriteReservedTag(ctx context.Context) (tags []*schema.SiteWriteTag, err error) {
	tags = make([]*schema.SiteWriteTag, 0)
	list, err := ts.tagCommonRepo.GetReservedTagList(ctx)
	if err != nil {
		return tags, err
	}
	for _, item := range list {
		tags = append(tags, &schema.SiteWriteTag{
			SlugName:    item.SlugName,
			DisplayName: item.DisplayName,
		})
	}
	return tags, nil
}

func (ts *TagCommonService) SetSiteWriteTag(ctx context.Context, recommendTags, reservedTags []string, userID string) (
	errFields []*validator.FormErrorField, err error) {
	recommendErr := ts.CheckTag(ctx, recommendTags, userID)
	reservedErr := ts.CheckTag(ctx, reservedTags, userID)
	if recommendErr != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "recommend_tags",
			ErrorMsg:   recommendErr.Error(),
		})
		err = recommendErr
	}
	if reservedErr != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "reserved_tags",
			ErrorMsg:   reservedErr.Error(),
		})
		err = reservedErr
	}
	if len(errFields) > 0 {
		return errFields, err
	}

	err = ts.SetTagsAttribute(ctx, recommendTags, "recommend")
	if err != nil {
		return nil, err
	}
	err = ts.SetTagsAttribute(ctx, reservedTags, "reserved")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// SetTagsAttribute
func (ts *TagCommonService) SetTagsAttribute(ctx context.Context, tags []string, attribute string) (err error) {
	var oldTags []*entity.Tag
	switch attribute {
	case "recommend":
		oldTags, err = ts.tagCommonRepo.GetRecommendTagList(ctx)
	case "reserved":
		oldTags, err = ts.tagCommonRepo.GetReservedTagList(ctx)
	default:
		return
	}
	if err != nil {
		return err
	}
	oldTagSlugNameList := make([]string, 0)
	for _, tag := range oldTags {
		oldTagSlugNameList = append(oldTagSlugNameList, tag.SlugName)
	}

	err = ts.tagCommonRepo.UpdateTagsAttribute(ctx, oldTagSlugNameList, attribute, false)
	if err != nil {
		return err
	}
	err = ts.tagCommonRepo.UpdateTagsAttribute(ctx, tags, attribute, true)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TagCommonService) GetTagListByNames(ctx context.Context, tagNames []string) ([]*entity.Tag, error) {
	for k, tagname := range tagNames {
		tagNames[k] = strings.ToLower(tagname)
	}
	tagList, err := ts.tagCommonRepo.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return tagList, nil
}

func (ts *TagCommonService) ExistRecommend(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	taginfo, err := ts.siteInfoService.GetSiteWrite(ctx)
	if err != nil {
		return false, err
	}
	if !taginfo.RequiredTag || len(taginfo.RecommendTags) == 0 {
		return true, nil
	}
	tagNames := make([]string, 0)
	for _, item := range tags {
		item.SlugName = strings.ReplaceAll(item.SlugName, " ", "-")
		tagNames = append(tagNames, item.SlugName)
	}
	list, err := ts.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return false, err
	}
	for _, item := range list {
		if item.Recommend {
			return true, nil
		}
	}
	return false, nil
}

func (ts *TagCommonService) HasNewTag(ctx context.Context, tags []*schema.TagItem) (bool, error) {
	tagNames := make([]string, 0)
	tagMap := make(map[string]bool)
	for _, item := range tags {
		item.SlugName = strings.ReplaceAll(item.SlugName, " ", "-")
		tagNames = append(tagNames, item.SlugName)
		tagMap[item.SlugName] = false
	}
	list, err := ts.GetTagListByNames(ctx, tagNames)
	if err != nil {
		return true, err
	}
	for _, item := range list {
		_, ok := tagMap[item.SlugName]
		if ok {
			tagMap[item.SlugName] = true
		}
	}
	for _, has := range tagMap {
		if !has {
			return true, nil
		}
	}
	return false, nil
}

// GetObjectTag get object tag
func (ts *TagCommonService) GetObjectTag(ctx context.Context, objectId string) (objTags []*schema.TagResp, err error) {
	tagsInfoList, err := ts.GetObjectEntityTag(ctx, objectId)
	if err != nil {
		return nil, err
	}
	return ts.TagFormat(ctx, tagsInfoList)
}

// AddTag get object tag
func (ts *TagCommonService) AddTag(ctx context.Context, req *schema.AddTagReq) (resp *schema.AddTagResp, err error) {
	_, exist, err := ts.GetTagBySlugName(ctx, req.SlugName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.BadRequest(reason.TagAlreadyExist)
	}
	slugName := strings.ReplaceAll(req.SlugName, " ", "-")
	slugName = strings.ToLower(slugName)
	tagInfo := &entity.Tag{
		SlugName:     slugName,
		DisplayName:  req.DisplayName,
		OriginalText: req.OriginalText,
		ParsedText:   req.ParsedText,
		Status:       entity.TagStatusAvailable,
		UserID:       req.UserID,
	}
	tagList := []*entity.Tag{tagInfo}
	err = ts.tagCommonRepo.AddTagList(ctx, tagList)
	if err != nil {
		return nil, err
	}
	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: tagInfo.ID,
		Title:    tagInfo.SlugName,
	}
	tagInfoJson, _ := json.Marshal(tagInfo)
	revisionDTO.Content = string(tagInfoJson)
	revisionID, err := ts.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return nil, err
	}
	ts.activityQueueService.Send(ctx, &schema.ActivityMsg{
		UserID:           req.UserID,
		ObjectID:         tagInfo.ID,
		OriginalObjectID: tagInfo.ID,
		ActivityTypeKey:  constant.ActTagCreated,
		RevisionID:       revisionID,
	})
	return &schema.AddTagResp{SlugName: tagInfo.SlugName}, nil
}

// AddTagList get object tag
func (ts *TagCommonService) AddTagList(ctx context.Context, tagList []*entity.Tag) (err error) {
	return ts.tagCommonRepo.AddTagList(ctx, tagList)
}

// GetTagByID get object tag
func (ts *TagCommonService) GetTagByID(ctx context.Context, tagID string) (tag *entity.Tag, exist bool, err error) {
	tag, exist, err = ts.tagCommonRepo.GetTagByID(ctx, tagID, false)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagIDsByMainTagID get object tag
func (ts *TagCommonService) GetTagIDsByMainTagID(ctx context.Context, tagID string) (tagIDs []string, err error) {
	tagIDs, err = ts.tagRepo.GetIDsByMainTagId(ctx, tagID)
	return
}

// GetTagBySlugName get object tag
func (ts *TagCommonService) GetTagBySlugName(ctx context.Context, slugName string) (tag *entity.Tag, exist bool, err error) {
	tag, exist, err = ts.tagCommonRepo.GetTagBySlugName(ctx, slugName)
	if !exist {
		return
	}
	ts.tagFormatRecommendAndReserved(ctx, tag)
	return
}

// GetTagListByIDs get object tag
func (ts *TagCommonService) GetTagListByIDs(ctx context.Context, ids []string) (tagList []*entity.Tag, err error) {
	tagList, err = ts.tagCommonRepo.GetTagListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

// GetTagPage get object tag
func (ts *TagCommonService) GetTagPage(ctx context.Context, page, pageSize int, tag *entity.Tag, queryCond string) (
	tagList []*entity.Tag, total int64, err error) {
	tagList, total, err = ts.tagCommonRepo.GetTagPage(ctx, page, pageSize, tag, queryCond)
	if err != nil {
		return nil, 0, err
	}
	ts.TagsFormatRecommendAndReserved(ctx, tagList)
	return
}

func (ts *TagCommonService) GetObjectEntityTag(ctx context.Context, objectId string) (objTags []*entity.Tag, err error) {
	tagList, err := ts.tagRelRepo.GetObjectTagRelList(ctx, objectId)
	if err != nil {
		return nil, err
	}
	tagIDList := make([]string, 0)
	for _, tag := range tagList {
		tagIDList = append(tagIDList, tag.TagID)
	}
	objTags, err = ts.GetTagListByIDs(ctx, tagIDList)
	if err != nil {
		return nil, err
	}
	return objTags, nil
}

func (ts *TagCommonService) TagFormat(ctx context.Context, tags []*entity.Tag) (objTags []*schema.TagResp, err error) {
	objTags = make([]*schema.TagResp, 0)
	for _, tagInfo := range tags {
		objTags = append(objTags, &schema.TagResp{
			SlugName:        tagInfo.SlugName,
			DisplayName:     tagInfo.DisplayName,
			MainTagSlugName: tagInfo.MainTagSlugName,
			Recommend:       tagInfo.Recommend,
			Reserved:        tagInfo.Reserved,
		})
	}
	return objTags, nil
}

func (ts *TagCommonService) TagsFormatRecommendAndReserved(ctx context.Context, tagList []*entity.Tag) {
	if len(tagList) == 0 {
		return
	}
	tagConfig, err := ts.siteInfoService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	if !tagConfig.RequiredTag {
		for _, tag := range tagList {
			tag.Recommend = false
		}
	}
}

func (ts *TagCommonService) tagFormatRecommendAndReserved(ctx context.Context, tag *entity.Tag) {
	if tag == nil {
		return
	}
	tagConfig, err := ts.siteInfoService.GetSiteWrite(ctx)
	if err != nil {
		log.Error(err)
		return
	}
	if !tagConfig.RequiredTag {
		tag.Recommend = false
	}
}

// BatchGetObjectTag batch get object tag
func (ts *TagCommonService) BatchGetObjectTag(ctx context.Context, objectIds []string) (map[string][]*schema.TagResp, error) {
	objectIDTagMap := make(map[string][]*schema.TagResp)
	if len(objectIds) == 0 {
		return objectIDTagMap, nil
	}
	objectTagRelList, err := ts.tagRelRepo.BatchGetObjectTagRelList(ctx, objectIds)
	if err != nil {
		return objectIDTagMap, err
	}
	tagIDList := make([]string, 0)
	for _, tag := range objectTagRelList {
		tagIDList = append(tagIDList, tag.TagID)
	}
	tagsInfoList, err := ts.GetTagListByIDs(ctx, tagIDList)
	if err != nil {
		return objectIDTagMap, err
	}
	tagsInfoMapping := make(map[string]*entity.Tag)
	tagsRank := make(map[string]int) // Used for sorting
	for idx, item := range tagsInfoList {
		tagsInfoMapping[item.ID] = item
		tagsRank[item.ID] = idx
	}

	for _, item := range objectTagRelList {
		_, ok := tagsInfoMapping[item.TagID]
		if ok {
			tagInfo := tagsInfoMapping[item.TagID]
			t := &schema.TagResp{
				ID:              tagInfo.ID,
				SlugName:        tagInfo.SlugName,
				DisplayName:     tagInfo.DisplayName,
				MainTagSlugName: tagInfo.MainTagSlugName,
				Recommend:       tagInfo.Recommend,
				Reserved:        tagInfo.Reserved,
			}
			objectIDTagMap[item.ObjectID] = append(objectIDTagMap[item.ObjectID], t)
		}
	}
	// The sorting in tagsRank is correct, object tags should be sorted by tagsRank
	for _, objectTags := range objectIDTagMap {
		sort.SliceStable(objectTags, func(i, j int) bool {
			return tagsRank[objectTags[i].ID] < tagsRank[objectTags[j].ID]
		})
	}
	return objectIDTagMap, nil
}

func (ts *TagCommonService) CheckTag(ctx context.Context, tags []string, userID string) (err error) {
	if len(tags) == 0 {
		return nil
	}

	// find tags name
	tagListInDb, err := ts.GetTagListByNames(ctx, tags)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity.Tag)
	checktags := make([]string, 0)

	for _, tag := range tagListInDb {
		if tag.MainTagID != 0 {
			checktags = append(checktags, fmt.Sprintf("\"%s\"", tag.SlugName))
		}
		tagInDbMapping[tag.SlugName] = tag
	}
	if len(checktags) > 0 {
		err = errors.BadRequest(reason.TagNotContainSynonym).WithMsg(fmt.Sprintf("Should not contain synonym tags %s", strings.Join(checktags, ",")))
		return err
	}

	addTagList := make([]*entity.Tag, 0)
	addTagMsgList := make([]string, 0)
	for _, tag := range tags {
		_, ok := tagInDbMapping[tag]
		if ok {
			continue
		}
		item := &entity.Tag{}
		item.SlugName = tag
		item.DisplayName = tag
		item.OriginalText = ""
		item.ParsedText = ""
		item.Status = entity.TagStatusAvailable
		item.UserID = userID
		addTagList = append(addTagList, item)
		addTagMsgList = append(addTagMsgList, tag)
	}

	if len(addTagList) > 0 {
		err = errors.BadRequest(reason.TagNotFound).WithMsg(fmt.Sprintf("tag [%s] does not exist",
			strings.Join(addTagMsgList, ",")))
		return err

	}

	return nil
}

// CheckTagsIsChange
func (ts *TagCommonService) CheckTagsIsChange(ctx context.Context, tagNameList, oldtagNameList []string) bool {
	check := make(map[string]bool)
	if len(tagNameList) != len(oldtagNameList) {
		return true
	}
	for _, item := range tagNameList {
		check[item] = false
	}
	for _, item := range oldtagNameList {
		_, ok := check[item]
		if !ok {
			return true
		}
		check[item] = true
	}
	for _, value := range check {
		if !value {
			return true
		}
	}
	return false
}

func (ts *TagCommonService) CheckChangeReservedTag(ctx context.Context, oldobjectTagData, objectTagData []*entity.Tag) (bool, bool, []string, []string) {
	reservedTagsMap := make(map[string]bool)
	needTagsMap := make([]string, 0)
	notNeedTagsMap := make([]string, 0)
	for _, tag := range objectTagData {
		if tag.Reserved {
			reservedTagsMap[tag.SlugName] = true
		}
	}
	for _, tag := range oldobjectTagData {
		if tag.Reserved {
			_, ok := reservedTagsMap[tag.SlugName]
			if !ok {
				needTagsMap = append(needTagsMap, tag.SlugName)
			} else {
				reservedTagsMap[tag.SlugName] = false
			}
		}
	}

	for k, v := range reservedTagsMap {
		if v {
			notNeedTagsMap = append(notNeedTagsMap, k)
		}
	}

	if len(needTagsMap) > 0 {
		return false, true, needTagsMap, []string{}
	}

	if len(notNeedTagsMap) > 0 {
		return true, false, []string{}, notNeedTagsMap
	}

	return true, true, []string{}, []string{}
}

// ObjectChangeTag change object tag list
func (ts *TagCommonService) ObjectChangeTag(ctx context.Context, objectTagData *schema.TagChange) (err error) {
	if len(objectTagData.Tags) == 0 {
		return nil
	}

	thisObjTagNameList := make([]string, 0)
	thisObjTagIDList := make([]string, 0)
	for _, t := range objectTagData.Tags {
		t.SlugName = strings.ToLower(t.SlugName)
		thisObjTagNameList = append(thisObjTagNameList, t.SlugName)
	}

	// find tags name
	tagListInDb, err := ts.tagCommonRepo.GetTagListByNames(ctx, thisObjTagNameList)
	if err != nil {
		return err
	}

	tagInDbMapping := make(map[string]*entity.Tag)
	for _, tag := range tagListInDb {
		tagInDbMapping[strings.ToLower(tag.SlugName)] = tag
		thisObjTagIDList = append(thisObjTagIDList, tag.ID)
	}

	addTagList := make([]*entity.Tag, 0)
	for _, tag := range objectTagData.Tags {
		_, ok := tagInDbMapping[strings.ToLower(tag.SlugName)]
		if ok {
			continue
		}
		item := &entity.Tag{}
		item.SlugName = strings.ReplaceAll(tag.SlugName, " ", "-")
		item.DisplayName = tag.DisplayName
		item.OriginalText = tag.OriginalText
		item.ParsedText = tag.ParsedText
		item.Status = entity.TagStatusAvailable
		item.UserID = objectTagData.UserID
		addTagList = append(addTagList, item)
	}

	if len(addTagList) > 0 {
		err = ts.tagCommonRepo.AddTagList(ctx, addTagList)
		if err != nil {
			return err
		}
		for _, tag := range addTagList {
			thisObjTagIDList = append(thisObjTagIDList, tag.ID)
			revisionDTO := &schema.AddRevisionDTO{
				UserID:   objectTagData.UserID,
				ObjectID: tag.ID,
				Title:    tag.SlugName,
			}
			tagInfoJson, _ := json.Marshal(tag)
			revisionDTO.Content = string(tagInfoJson)
			revisionID, err := ts.revisionService.AddRevision(ctx, revisionDTO, true)
			if err != nil {
				return err
			}
			ts.activityQueueService.Send(ctx, &schema.ActivityMsg{
				UserID:           objectTagData.UserID,
				ObjectID:         tag.ID,
				OriginalObjectID: tag.ID,
				ActivityTypeKey:  constant.ActTagCreated,
				RevisionID:       revisionID,
			})
		}
	}

	err = ts.CreateOrUpdateTagRelList(ctx, objectTagData.ObjectID, thisObjTagIDList)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TagCommonService) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	return ts.tagRelRepo.CountTagRelByTagID(ctx, tagID)
}

// RefreshTagQuestionCount refresh tag question count
func (ts *TagCommonService) RefreshTagQuestionCount(ctx context.Context, tagIDs []string) (err error) {
	for _, tagID := range tagIDs {
		count, err := ts.tagRelRepo.CountTagRelByTagID(ctx, tagID)
		if err != nil {
			return err
		}
		err = ts.tagCommonRepo.UpdateTagQuestionCount(ctx, tagID, int(count))
		if err != nil {
			return err
		}
		log.Debugf("tag count updated %s %d", tagID, count)
	}
	return nil
}

func (ts *TagCommonService) RefreshTagCountByQuestionID(ctx context.Context, questionID string) (err error) {
	tagListList, err := ts.tagRelRepo.GetObjectTagRelList(ctx, questionID)
	if err != nil {
		return err
	}
	tagIDs := make([]string, 0)
	for _, item := range tagListList {
		tagIDs = append(tagIDs, item.TagID)
	}
	err = ts.RefreshTagQuestionCount(ctx, tagIDs)
	if err != nil {
		return err
	}
	return nil
}

// RemoveTagRelListByObjectID remove tag relation by object id
func (ts *TagCommonService) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return ts.tagRelRepo.RemoveTagRelListByObjectID(ctx, objectID)
}

// RecoverTagRelListByObjectID recover tag relation by object id
func (ts *TagCommonService) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return ts.tagRelRepo.RecoverTagRelListByObjectID(ctx, objectID)
}

func (ts *TagCommonService) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return ts.tagRelRepo.HideTagRelListByObjectID(ctx, objectID)
}

func (ts *TagCommonService) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	return ts.tagRelRepo.ShowTagRelListByObjectID(ctx, objectID)
}

// CreateOrUpdateTagRelList if tag relation is exists update status, if not create it
func (ts *TagCommonService) CreateOrUpdateTagRelList(ctx context.Context, objectId string, tagIDs []string) (err error) {
	addTagIDMapping := make(map[string]struct{})
	for _, t := range tagIDs {
		addTagIDMapping[t] = struct{}{}
	}

	// get all old relation
	oldTagRelList, err := ts.tagRelRepo.GetObjectTagRelList(ctx, objectId)
	if err != nil {
		return err
	}
	var deleteTagRel []int64
	needRefreshTagIDs := make([]string, 0, len(oldTagRelList)+len(tagIDs))
	needRefreshTagIDs = append(needRefreshTagIDs, tagIDs...)
	for _, rel := range oldTagRelList {
		if _, ok := addTagIDMapping[rel.TagID]; !ok {
			deleteTagRel = append(deleteTagRel, rel.ID)
			needRefreshTagIDs = append(needRefreshTagIDs, rel.TagID)
		}
	}

	addTagRelList := make([]*entity.TagRel, 0)
	enableTagRelList := make([]int64, 0)
	defaultTagRelStatus, err := ts.tagRelRepo.GetTagRelDefaultStatusByObjectID(ctx, objectId)
	if err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		rel, exist, err := ts.tagRelRepo.GetObjectTagRelWithoutStatus(ctx, objectId, tagID)
		if err != nil {
			return err
		}
		// if not exist add tag relation
		if !exist {
			addTagRelList = append(addTagRelList, &entity.TagRel{
				TagID: tagID, ObjectID: objectId, Status: defaultTagRelStatus,
			})
		}
		// if exist and has been removed, that should be enabled
		if exist && rel.Status != entity.TagRelStatusAvailable && rel.Status != entity.TagRelStatusHide {
			enableTagRelList = append(enableTagRelList, rel.ID)
		}
	}

	if len(deleteTagRel) > 0 {
		if err = ts.tagRelRepo.RemoveTagRelListByIDs(ctx, deleteTagRel); err != nil {
			return err
		}
	}
	if len(addTagRelList) > 0 {
		if err = ts.tagRelRepo.AddTagRelList(ctx, addTagRelList); err != nil {
			return err
		}
	}
	if len(enableTagRelList) > 0 {
		if err = ts.tagRelRepo.EnableTagRelByIDs(ctx, enableTagRelList, defaultTagRelStatus == entity.TagRelStatusHide); err != nil {
			return err
		}
	}

	err = ts.RefreshTagQuestionCount(ctx, needRefreshTagIDs)
	if err != nil {
		log.Error(err)
	}
	return nil
}

func (ts *TagCommonService) UpdateTag(ctx context.Context, req *schema.UpdateTagReq) (err error) {
	var canUpdate bool
	_, existUnreviewed, err := ts.revisionService.ExistUnreviewedByObjectID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if existUnreviewed {
		err = errors.BadRequest(reason.AnswerCannotUpdate)
		return err
	}

	tagInfo, exist, err := ts.GetTagByID(ctx, req.TagID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.TagNotFound)
	}

	//Adding equivalent slug formatting for tag update
	slugName := strings.ReplaceAll(req.SlugName, " ", "-")
	slugName = strings.ToLower(slugName)

	//If the content is the same, ignore it
	if tagInfo.OriginalText == req.OriginalText &&
		tagInfo.DisplayName == req.DisplayName &&
		tagInfo.SlugName == slugName {
		return nil
	}

	tagInfo.SlugName = slugName
	tagInfo.DisplayName = req.DisplayName
	tagInfo.OriginalText = req.OriginalText
	tagInfo.ParsedText = req.ParsedText

	revisionDTO := &schema.AddRevisionDTO{
		UserID:   req.UserID,
		ObjectID: tagInfo.ID,
		Title:    tagInfo.SlugName,
		Log:      req.EditSummary,
	}

	if req.NoNeedReview {
		canUpdate = true
		err = ts.tagRepo.UpdateTag(ctx, tagInfo)
		if err != nil {
			return err
		}
		if tagInfo.MainTagID == 0 && len(req.SlugName) > 0 {
			log.Debugf("tag %s update slug_name", tagInfo.SlugName)
			tagList, err := ts.tagRepo.GetTagList(ctx, &entity.Tag{MainTagID: converter.StringToInt64(tagInfo.ID)})
			if err != nil {
				return err
			}
			updateTagSlugNames := make([]string, 0)
			for _, tag := range tagList {
				updateTagSlugNames = append(updateTagSlugNames, tag.SlugName)
			}
			err = ts.tagRepo.UpdateTagSynonym(ctx, updateTagSlugNames, converter.StringToInt64(tagInfo.ID), tagInfo.MainTagSlugName)
			if err != nil {
				return err
			}
		}
		revisionDTO.Status = entity.RevisionReviewPassStatus
	} else {
		revisionDTO.Status = entity.RevisionUnreviewedStatus
	}

	tagInfoJson, _ := json.Marshal(tagInfo)
	revisionDTO.Content = string(tagInfoJson)
	revisionID, err := ts.revisionService.AddRevision(ctx, revisionDTO, true)
	if err != nil {
		return err
	}
	if canUpdate {
		ts.activityQueueService.Send(ctx, &schema.ActivityMsg{
			UserID:           req.UserID,
			ObjectID:         tagInfo.ID,
			OriginalObjectID: tagInfo.ID,
			ActivityTypeKey:  constant.ActTagEdited,
			RevisionID:       revisionID,
		})
	}

	return
}

// MigrateTagQuestions migrate tag question
func (ts *TagCommonService) MigrateTagQuestions(ctx context.Context, sourceTagID, targetTagID string) (err error) {
	return ts.tagRelRepo.MigrateTagObjects(ctx, sourceTagID, targetTagID)
}
