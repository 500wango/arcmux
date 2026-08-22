package model

import (
	"fmt"
	"testing"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetChannelExcludingKeepsHighestAvailablePriority(t *testing.T) {
	modelName := fmt.Sprintf("db-failover-model-%s", t.Name())
	highPriority := int64(10)
	lowPriority := int64(1)
	weight := uint(100)
	channels := []struct {
		id       int
		priority *int64
	}{
		{id: 9901, priority: &highPriority},
		{id: 9902, priority: &highPriority},
		{id: 9903, priority: &lowPriority},
	}

	require.NoError(t, DB.Where("model = ?", modelName).Delete(&Ability{}).Error)
	require.NoError(t, DB.Where("models = ?", modelName).Delete(&Channel{}).Error)
	t.Cleanup(func() {
		DB.Where("model = ?", modelName).Delete(&Ability{})
		DB.Where("models = ?", modelName).Delete(&Channel{})
	})

	for _, channel := range channels {
		require.NoError(t, DB.Create(&Channel{
			Id:       channel.id,
			Type:     constant.ChannelTypeOpenAI,
			Key:      fmt.Sprintf("key-%d", channel.id),
			Status:   common.ChannelStatusEnabled,
			Name:     fmt.Sprintf("channel-%d", channel.id),
			Weight:   &weight,
			Models:   modelName,
			Group:    "default",
			Priority: channel.priority,
		}).Error)
		require.NoError(t, DB.Create(&Ability{
			Group:     "default",
			Model:     modelName,
			ChannelId: channel.id,
			Enabled:   true,
			Priority:  channel.priority,
			Weight:    weight,
		}).Error)
	}

	originalMemoryCacheEnabled := common.MemoryCacheEnabled
	common.MemoryCacheEnabled = false
	t.Cleanup(func() { common.MemoryCacheEnabled = originalMemoryCacheEnabled })

	first, err := GetChannelExcluding("default", modelName, 0, "/v1/chat/completions", nil)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Contains(t, []int{9901, 9902}, first.Id)

	second, err := GetChannelExcluding("default", modelName, 1, "/v1/chat/completions", map[int]struct{}{first.Id: {}})
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Contains(t, []int{9901, 9902}, second.Id)
	assert.NotEqual(t, first.Id, second.Id)

	third, err := GetChannelExcluding("default", modelName, 2, "/v1/chat/completions", map[int]struct{}{first.Id: {}, second.Id: {}})
	require.NoError(t, err)
	require.NotNil(t, third)
	assert.Equal(t, 9903, third.Id)
}

// Regression: PostgreSQL sorts NULLs first on DESC while MySQL/SQLite sort them
// last. NULL priority must behave like the lowest tier on every supported
// database, matching the legacy MAX(priority) retry path that ignores NULLs.
func TestGetChannelExcludingTreatsNullPriorityAsLowestTier(t *testing.T) {
	modelName := fmt.Sprintf("db-null-priority-model-%s", t.Name())
	highPriority := int64(10)
	lowPriority := int64(1)
	weight := uint(100)
	channels := []struct {
		id       int
		priority *int64
	}{
		{id: 9911, priority: &highPriority},
		{id: 9912, priority: &lowPriority},
		{id: 9913, priority: nil},
	}

	require.NoError(t, DB.Where("model = ?", modelName).Delete(&Ability{}).Error)
	require.NoError(t, DB.Where("models = ?", modelName).Delete(&Channel{}).Error)
	t.Cleanup(func() {
		DB.Where("model = ?", modelName).Delete(&Ability{})
		DB.Where("models = ?", modelName).Delete(&Channel{})
	})

	for _, channel := range channels {
		require.NoError(t, DB.Create(&Channel{
			Id:       channel.id,
			Type:     constant.ChannelTypeOpenAI,
			Key:      fmt.Sprintf("key-%d", channel.id),
			Status:   common.ChannelStatusEnabled,
			Name:     fmt.Sprintf("channel-%d", channel.id),
			Weight:   &weight,
			Models:   modelName,
			Group:    "default",
			Priority: channel.priority,
		}).Error)
		require.NoError(t, DB.Create(&Ability{
			Group:     "default",
			Model:     modelName,
			ChannelId: channel.id,
			Enabled:   true,
			Priority:  channel.priority,
			Weight:    weight,
		}).Error)
	}

	originalMemoryCacheEnabled := common.MemoryCacheEnabled
	common.MemoryCacheEnabled = false
	t.Cleanup(func() { common.MemoryCacheEnabled = originalMemoryCacheEnabled })

	excluded := map[int]struct{}{}
	expectations := []struct {
		retry           int
		expectedChannel int
	}{
		{retry: 0, expectedChannel: 9911},
		{retry: 1, expectedChannel: 9912},
		{retry: 2, expectedChannel: 9913},
	}
	for _, tc := range expectations {
		channel, err := GetChannelExcluding("default", modelName, tc.retry, "/v1/chat/completions", excluded)
		require.NoError(t, err)
		require.NotNil(t, channel, "retry %d", tc.retry)
		assert.Equal(t, tc.expectedChannel, channel.Id, "retry %d", tc.retry)
		excluded[channel.Id] = struct{}{}
	}
}
