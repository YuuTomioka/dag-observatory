package minio

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Secure    bool
	Bucket    string
}

type Presigner struct {
	client *minio.Client
	bucket string
}

func NewPresigner(cfg Config) *Presigner {
	if cfg.Endpoint == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil
	}
	return &Presigner{client: client, bucket: cfg.Bucket}
}

func (p *Presigner) PresignGet(ctx context.Context, objectKey string, expires time.Duration) (*url.URL, error) {
	if p == nil || p.client == nil {
		return nil, fmt.Errorf("minio: presigner not configured")
	}
	return p.client.PresignedGetObject(ctx, p.bucket, objectKey, expires, url.Values{})
}

func (p *Presigner) PresignGetWithBucket(ctx context.Context, bucket, objectKey string, expires time.Duration) (*url.URL, error) {
	if p == nil || p.client == nil {
		return nil, fmt.Errorf("minio: presigner not configured")
	}
	if bucket == "" {
		bucket = p.bucket
	}
	return p.client.PresignedGetObject(ctx, bucket, objectKey, expires, url.Values{})
}
