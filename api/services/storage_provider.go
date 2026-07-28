package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// ErrReportExpired is returned when a signed URL is requested for an expired report.
var ErrReportExpired = errors.New("report has expired; signed URL cannot be issued")

// StorageProvider is a generic interface for storing and retrieving report files.
// All implementations must cap signed URL TTL to min(ttl, expires-now).
type StorageProvider interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, key string) error
	// SignedURL returns a pre-authenticated URL valid for at most min(ttl, expires-now).
	// Returns ErrReportExpired if the report has already expired.
	SignedURL(ctx context.Context, key string, ttl time.Duration, expires time.Time) (string, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// LocalStorageProvider
// ─────────────────────────────────────────────────────────────────────────────

// LocalStorageProvider writes files to a configurable base directory.
type LocalStorageProvider struct {
	baseDir string
	secret  []byte // HMAC secret for local signed URLs
	logger  *zap.Logger
}

// NewLocalStorageProvider creates a local storage provider.
func NewLocalStorageProvider(baseDir, hmacSecret string, logger *zap.Logger) (*LocalStorageProvider, error) {
	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return nil, fmt.Errorf("storage: cannot create base dir %s: %w", baseDir, err)
	}
	return &LocalStorageProvider{
		baseDir: baseDir,
		secret:  []byte(hmacSecret),
		logger:  logger,
	}, nil
}

func (p *LocalStorageProvider) absPath(key string) string {
	// Sanitise key — prevent directory traversal.
	safe := filepath.Join(p.baseDir, filepath.FromSlash(filepath.Clean("/"+key)))
	return safe
}

func (p *LocalStorageProvider) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	path := p.absPath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("storage put mkdir: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("storage put create: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("storage put copy: %w", err)
	}
	return nil
}

func (p *LocalStorageProvider) Get(_ context.Context, key string) (io.ReadCloser, int64, error) {
	path := p.absPath(key)
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("storage get open: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}

func (p *LocalStorageProvider) Delete(_ context.Context, key string) error {
	err := os.Remove(p.absPath(key))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage delete: %w", err)
	}
	return nil
}

func (p *LocalStorageProvider) SignedURL(_ context.Context, key string, ttl time.Duration, expires time.Time) (string, error) {
	now := time.Now().UTC()
	remaining := expires.Sub(now)
	if remaining <= 0 {
		return "", ErrReportExpired
	}
	if ttl > remaining {
		ttl = remaining
	}
	validUntil := now.Add(ttl).Unix()
	payload := fmt.Sprintf("%s:%d", key, validUntil)
	mac := hmac.New(sha256.New, p.secret)
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("/api/v1/internal/storage/download?key=%s&expires=%d&sig=%s",
		key, validUntil, sig), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// S3StorageProvider
// ─────────────────────────────────────────────────────────────────────────────

// S3StorageProvider stores files in an S3-compatible bucket (supports MinIO).
type S3StorageProvider struct {
	client *s3.Client
	bucket string
	logger *zap.Logger
}

// NewS3StorageProvider creates an S3 storage provider.
func NewS3StorageProvider(ctx context.Context, bucket, region, endpoint, accessKey, secretKey string, logger *zap.Logger) (*S3StorageProvider, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if accessKey != "" && secretKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage s3 config: %w", err)
	}
	var clientOpts []func(*s3.Options)
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}
	client := s3.NewFromConfig(cfg, clientOpts...)
	return &S3StorageProvider{client: client, bucket: bucket, logger: logger}, nil
}

func (p *S3StorageProvider) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("s3 put object: %w", err)
	}
	return nil
}

func (p *S3StorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	out, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("s3 get object: %w", err)
	}
	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return out.Body, size, nil
}

func (p *S3StorageProvider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (p *S3StorageProvider) SignedURL(ctx context.Context, key string, ttl time.Duration, expires time.Time) (string, error) {
	now := time.Now().UTC()
	remaining := expires.Sub(now)
	if remaining <= 0 {
		return "", ErrReportExpired
	}
	if ttl > remaining {
		ttl = remaining
	}
	presigner := s3.NewPresignClient(p.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("s3 presign: %w", err)
	}
	return req.URL, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor helper
// ─────────────────────────────────────────────────────────────────────────────

// ContentTypeForFormat returns the MIME type for the given export format extension.
func ContentTypeForFormat(format string) string {
	switch strings.ToUpper(format) {
	case "CSV":
		return "text/csv; charset=utf-8"
	case "XLSX":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "PDF":
		return "application/pdf"
	default:
		return http.DetectContentType(nil)
	}
}
