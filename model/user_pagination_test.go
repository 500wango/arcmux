package model

import (
	"fmt"
	"testing"

	"github.com/500wango/arcmux/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertUsersForPaginationTest(t *testing.T, total int) {
	t.Helper()
	for id := 1; id <= total; id++ {
		user := &User{
			Id:          id,
			Username:    fmt.Sprintf("user%02d", id),
			Password:    "password123",
			DisplayName: fmt.Sprintf("User %02d", id),
			Email:       fmt.Sprintf("user%02d@example.com", id),
			Role:        common.RoleCommonUser,
			Status:      common.UserStatusEnabled,
			Group:       "default",
			AffCode:     fmt.Sprintf("aff%02d", id),
		}
		require.NoError(t, DB.Create(user).Error)
	}
}

func collectUserIDs(users []*User) []int {
	ids := make([]int, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.Id)
	}
	return ids
}

func resetQuotaDataCacheForUserTest(t *testing.T) {
	t.Helper()
	CacheQuotaDataLock.Lock()
	CacheQuotaData = make(map[string]*QuotaData)
	CacheQuotaDataLock.Unlock()
	t.Cleanup(func() {
		CacheQuotaDataLock.Lock()
		CacheQuotaData = make(map[string]*QuotaData)
		CacheQuotaDataLock.Unlock()
	})
}

func TestGetAllUsersSortsBeforePagination(t *testing.T) {
	truncateTables(t)
	insertUsersForPaginationTest(t, 42)

	pageOne, total, err := GetAllUsers(&common.PageInfo{Page: 1, PageSize: 20}, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	assert.Equal(t, int64(42), total)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, collectUserIDs(pageOne))

	pageTwo, total, err := GetAllUsers(&common.PageInfo{Page: 2, PageSize: 20}, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	assert.Equal(t, int64(42), total)
	assert.Equal(t, []int{21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}, collectUserIDs(pageTwo))

	pageThree, total, err := GetAllUsers(&common.PageInfo{Page: 3, PageSize: 20}, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	assert.Equal(t, int64(42), total)
	assert.Equal(t, []int{41, 42}, collectUserIDs(pageThree))
}

func TestSearchUsersSortsBeforePagination(t *testing.T) {
	truncateTables(t)
	insertUsersForPaginationTest(t, 42)

	users, total, err := SearchUsers("user", "", nil, nil, 20, 20, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	assert.Equal(t, int64(42), total)
	assert.Equal(t, []int{21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}, collectUserIDs(users))
}

func TestGetAllUsersIncludesPersistedAndPendingUsedTokens(t *testing.T) {
	truncateTables(t)
	resetQuotaDataCacheForUserTest(t)
	insertUsersForPaginationTest(t, 2)
	require.NoError(t, DB.Create(&[]QuotaData{
		{UserID: 1, Username: "user01", ModelName: "gpt-a", CreatedAt: 1000, TokenUsed: 120},
		{UserID: 1, Username: "user01", ModelName: "gpt-b", CreatedAt: 2000, TokenUsed: 230},
		{UserID: 2, Username: "user02", ModelName: "gpt-a", CreatedAt: 1000, TokenUsed: 70},
	}).Error)
	LogQuotaData(QuotaDataLogParams{
		UserID:    1,
		Username:  "user01",
		ModelName: "gpt-a",
		CreatedAt: 3000,
		TokenUsed: 25,
	})

	users, total, err := GetAllUsers(&common.PageInfo{Page: 1, PageSize: 20}, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, users, 2)
	assert.Equal(t, int64(375), users[0].UsedTokens)
	assert.Equal(t, int64(70), users[1].UsedTokens)

	SaveQuotaDataCache()
	users, _, err = GetAllUsers(&common.PageInfo{Page: 1, PageSize: 20}, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, int64(375), users[0].UsedTokens)
	assert.Equal(t, int64(70), users[1].UsedTokens)
}

func TestSearchUsersIncludesUsedTokensForMatchedUsers(t *testing.T) {
	truncateTables(t)
	resetQuotaDataCacheForUserTest(t)
	insertUsersForPaginationTest(t, 2)
	require.NoError(t, DB.Create(&[]QuotaData{
		{UserID: 1, Username: "user01", ModelName: "gpt-a", CreatedAt: 1000, TokenUsed: 40},
		{UserID: 2, Username: "user02", ModelName: "gpt-a", CreatedAt: 1000, TokenUsed: 90},
	}).Error)

	users, total, err := SearchUsers("user02", "", nil, nil, 0, 20, NewUserSortOptions("id", "asc"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
	assert.Equal(t, 2, users[0].Id)
	assert.Equal(t, int64(90), users[0].UsedTokens)
}
