package services_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLocalStorageProvider_PutGetDelete(t *testing.T) {
	dir := t.TempDir()
	logger := zap.NewNop()
	provider, err := services.NewLocalStorageProvider(dir, "secret-key", logger)
	require.NoError(t, err)

	ctx := context.Background()
	key := "reports/test-job-123.csv"
	content := []byte("User ID,Score\nuser1,85\nuser2,92\n")

	// Put
	err = provider.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "text/csv")
	assert.NoError(t, err)

	// Get
	rc, size, err := provider.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	rc.Close()
	assert.Equal(t, content, got)

	// SignedURL valid
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	url, err := provider.SignedURL(ctx, key, 15*time.Minute, expiresAt)
	assert.NoError(t, err)
	assert.Contains(t, url, key)
	assert.Contains(t, url, "sig=")

	// SignedURL for expired report fails with ErrReportExpired
	expiredAt := time.Now().UTC().Add(-1 * time.Minute)
	_, err = provider.SignedURL(ctx, key, 15*time.Minute, expiredAt)
	assert.Equal(t, services.ErrReportExpired, err)

	// Delete
	err = provider.Delete(ctx, key)
	assert.NoError(t, err)

	// Get after Delete fails
	_, _, err = provider.Get(ctx, key)
	assert.Error(t, err)
}

func TestValidateCronExpr(t *testing.T) {
	// Valid standard 5-field expressions
	assert.NoError(t, services.ValidateCronExpr("0 8 * * *"))
	assert.NoError(t, services.ValidateCronExpr("0 0 * * 1"))
	assert.NoError(t, services.ValidateCronExpr("*/15 * * * *"))

	// Invalid expressions
	assert.Error(t, services.ValidateCronExpr("invalid cron"))
	assert.Error(t, services.ValidateCronExpr("1 2 3 4"))
}

func TestValidateIANATimezone(t *testing.T) {
	assert.NoError(t, services.ValidateIANATimezone("UTC"))
	assert.NoError(t, services.ValidateIANATimezone("Asia/Kolkata"))
	assert.NoError(t, services.ValidateIANATimezone("America/New_York"))

	assert.Error(t, services.ValidateIANATimezone("Invalid/Zone"))
	assert.Error(t, services.ValidateIANATimezone(""))
}
