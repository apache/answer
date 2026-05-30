package fake_username

import (
	"context"

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/checker"
	"github.com/segmentfault/pacman/log"
)

type AnonymityService struct {
	fakeUsernameRepo FakeUsernameRepo
}

func NewAnonymityService(fakeUsernameRepo FakeUsernameRepo) *AnonymityService {
	return &AnonymityService{
		fakeUsernameRepo: fakeUsernameRepo,
	}
}

func (s *AnonymityService) AnonymizeUserData(
	ctx context.Context, userIDs []string, questionID, forUserID string,
) (anonymizeInfo map[string]*schema.UserBasicInfo, err error) {
	anonymizeInfo = map[string]*schema.UserBasicInfo{}

	userIDs = checker.FilterEmptyString(userIDs)
	if len(userIDs) == 0 {
		return
	}

	filteredUserIDs := make([]string, 0, len(userIDs))
	if forUserID != "" {
		for _, id := range userIDs {
			if id == forUserID {
				continue
			}
			filteredUserIDs = append(filteredUserIDs, id)
		}
	} else {
		filteredUserIDs = userIDs
	}

	fakeUsernames, err := s.fakeUsernameRepo.BatchGetByUserIDs(ctx, filteredUserIDs, questionID)
	if err != nil {
		return
	}

	for _, item := range fakeUsernames {
		anonymizeInfo[item.UserID] = &schema.UserBasicInfo{DisplayName: item.FakeName}
	}
	log.Errorf("fake usernames for %s: %d", questionID, len(fakeUsernames))

	return
}
