package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	instructorAnalyticsCacheTTL          = 60 * time.Second
	instructorAnalyticsVersionKeyFmt     = "analytics:instructor:%s:version"
	instructorAnalyticsCacheKeyFmt       = "analytics:instructor:%s:v%s:%s"
	quizAnalyticsVersionKeyFmt           = "analytics:quiz:%s:version"
	quizAnalyticsCacheKeyFmt             = "analytics:quiz:%s:v%s:%s"
)

type InstructorAnalyticsCache struct {
	client *goredis.Client
	logger *zap.Logger
}

func NewInstructorAnalyticsCache(client *goredis.Client, logger *zap.Logger) *InstructorAnalyticsCache {
	return &InstructorAnalyticsCache{client: client, logger: logger}
}

func (c *InstructorAnalyticsCache) available() bool {
	return c != nil && c.client != nil
}

func (c *InstructorAnalyticsCache) BumpInstructorVersion(userID string) {
	if !c.available() || userID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := c.client.Incr(ctx, fmt.Sprintf(instructorAnalyticsVersionKeyFmt, userID)).Err(); err != nil && c.logger != nil {
		c.logger.Warn("instructor analytics cache version bump failed", zap.String("user_id", userID), zap.Error(err))
	}
}

func (c *InstructorAnalyticsCache) BumpQuizVersion(quizID string) {
	if !c.available() || quizID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := c.client.Incr(ctx, fmt.Sprintf(quizAnalyticsVersionKeyFmt, quizID)).Err(); err != nil && c.logger != nil {
		c.logger.Warn("quiz analytics cache version bump failed", zap.String("quiz_id", quizID), zap.Error(err))
	}
}

func (c *InstructorAnalyticsCache) instructorVersion(ctx context.Context, userID string) string {
	if !c.available() {
		return "0"
	}
	val, err := c.client.Get(ctx, fmt.Sprintf(instructorAnalyticsVersionKeyFmt, userID)).Result()
	if err == goredis.Nil {
		return "0"
	}
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("instructor analytics cache version read failed", zap.Error(err))
		}
		return "0"
	}
	return val
}

func (c *InstructorAnalyticsCache) quizVersion(ctx context.Context, quizID string) string {
	if !c.available() {
		return "0"
	}
	val, err := c.client.Get(ctx, fmt.Sprintf(quizAnalyticsVersionKeyFmt, quizID)).Result()
	if err == goredis.Nil {
		return "0"
	}
	if err != nil {
		if c.logger != nil {
			c.logger.Warn("quiz analytics cache version read failed", zap.Error(err))
		}
		return "0"
	}
	return val
}

func (c *InstructorAnalyticsCache) GetPortfolioJSON(ctx context.Context, userID, hash string, dest interface{}) bool {
	if !c.available() {
		return false
	}
	version := c.instructorVersion(ctx, userID)
	key := fmt.Sprintf(instructorAnalyticsCacheKeyFmt, userID, version, hash)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != goredis.Nil && c.logger != nil {
			c.logger.Warn("instructor analytics cache get failed", zap.Error(err))
		}
		return false
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false
	}
	return true
}

func (c *InstructorAnalyticsCache) SetPortfolioJSON(ctx context.Context, userID, hash string, value interface{}) {
	if !c.available() {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	version := c.instructorVersion(ctx, userID)
	key := fmt.Sprintf(instructorAnalyticsCacheKeyFmt, userID, version, hash)
	if err := c.client.Set(ctx, key, raw, instructorAnalyticsCacheTTL).Err(); err != nil && c.logger != nil {
		c.logger.Warn("instructor analytics cache set failed", zap.Error(err))
	}
}

func (c *InstructorAnalyticsCache) GetQuizJSON(ctx context.Context, quizID, hash string, dest interface{}) bool {
	if !c.available() {
		return false
	}
	version := c.quizVersion(ctx, quizID)
	key := fmt.Sprintf(quizAnalyticsCacheKeyFmt, quizID, version, hash)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err != goredis.Nil && c.logger != nil {
			c.logger.Warn("quiz analytics cache get failed", zap.Error(err))
		}
		return false
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return false
	}
	return true
}

func (c *InstructorAnalyticsCache) SetQuizJSON(ctx context.Context, quizID, hash string, value interface{}) {
	if !c.available() {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	version := c.quizVersion(ctx, quizID)
	key := fmt.Sprintf(quizAnalyticsCacheKeyFmt, quizID, version, hash)
	if err := c.client.Set(ctx, key, raw, instructorAnalyticsCacheTTL).Err(); err != nil && c.logger != nil {
		c.logger.Warn("quiz analytics cache set failed", zap.Error(err))
	}
}

func CanonicalEndpointHash(parts ...string) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}
