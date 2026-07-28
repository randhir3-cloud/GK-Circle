package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	learnerAnalyticsCacheTTL     = 60 * time.Second
	learnerAnalyticsVersionKeyFmt = "analytics:learner:%s:version"
	learnerAnalyticsCacheKeyFmt   = "analytics:learner:%s:v%s:%s"
)

// LearnerAnalyticsCache provides versioned Redis caching with fail-open behaviour.
type LearnerAnalyticsCache struct {
	client *goredis.Client
	logger *zap.Logger
}

func NewLearnerAnalyticsCache(client *goredis.Client, logger *zap.Logger) *LearnerAnalyticsCache {
	return &LearnerAnalyticsCache{client: client, logger: logger}
}

func (c *LearnerAnalyticsCache) available() bool {
	return c != nil && c.client != nil
}

func (c *LearnerAnalyticsCache) BumpVersion(userID string) {
	if !c.available() || userID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := c.client.Incr(ctx, fmt.Sprintf(learnerAnalyticsVersionKeyFmt, userID)).Err(); err != nil && c.logger != nil {
		c.logger.Warn("learner analytics cache version bump failed", zap.String("user_id", userID), zap.Error(err))
	}
}

func (c *LearnerAnalyticsCache) BumpVersions(userIDs []string) {
	seen := map[string]struct{}{}
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		c.BumpVersion(userID)
	}
}

func (c *LearnerAnalyticsCache) version(ctx context.Context, userID string) string {
	if !c.available() {
		return "0"
	}
	val, err := c.client.Get(ctx, fmt.Sprintf(learnerAnalyticsVersionKeyFmt, userID)).Result()
	if err == goredis.Nil {
		return "0"
	}
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("learner analytics cache version read failed", zap.Error(err))
		}
		return "0"
	}
	return val
}

func endpointHash(parts ...string) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (c *LearnerAnalyticsCache) GetJSON(ctx context.Context, userID, hash string, dest interface{}) bool {
	if !c.available() {
		return false
	}
	version := c.version(ctx, userID)
	key := fmt.Sprintf(learnerAnalyticsCacheKeyFmt, userID, version, hash)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != goredis.Nil && c.logger != nil {
			c.logger.Warn("learner analytics cache get failed", zap.Error(err))
		}
		return false
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false
	}
	return true
}

func (c *LearnerAnalyticsCache) SetJSON(ctx context.Context, userID, hash string, value interface{}) {
	if !c.available() {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	version := c.version(ctx, userID)
	key := fmt.Sprintf(learnerAnalyticsCacheKeyFmt, userID, version, hash)
	if err := c.client.Set(ctx, key, raw, learnerAnalyticsCacheTTL).Err(); err != nil && c.logger != nil {
		c.logger.Warn("learner analytics cache set failed", zap.Error(err))
	}
}

// Exported for tests.
func LearnerAnalyticsEndpointHash(parts ...string) string {
	return endpointHash(parts...)
}

func LearnerAnalyticsVersionKey(userID string) string {
	return fmt.Sprintf(learnerAnalyticsVersionKeyFmt, userID)
}

func LearnerAnalyticsCacheKey(userID, version, hash string) string {
	return fmt.Sprintf(learnerAnalyticsCacheKeyFmt, userID, version, hash)
}

func ParseIntOrZero(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}
