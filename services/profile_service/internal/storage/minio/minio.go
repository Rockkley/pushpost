package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	PublicBaseURL   string // e.g. "http://localhost:9000"
}

type ObjectStorage struct {
	client     *minio.Client
	bucketName string
	publicBase string
}

func New(cfg Config) (*ObjectStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: new client: %w", err)
	}

	return &ObjectStorage{
		client:     client,
		bucketName: cfg.BucketName,
		publicBase: cfg.PublicBaseURL,
	}, nil
}

// EnsureBucket создаёт бакет с публичным чтением, если его нет.
// Вызывается при старте приложения.
func (s *ObjectStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("minio: check bucket: %w", err)
	}

	if exists {
		return nil
	}

	if err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("minio: make bucket: %w", err)
	}

	// Публичное чтение — аватары открыты всем
	policy := fmt.Sprintf(`{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Principal":{"AWS":["*"]},
			"Action":["s3:GetObject"],
			"Resource":["arn:aws:s3:::%s/*"]
		}]
	}`, s.bucketName)

	if err = s.client.SetBucketPolicy(ctx, s.bucketName, policy); err != nil {
		return fmt.Errorf("minio: set bucket policy: %w", err)
	}

	return nil
}

func (s *ObjectStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio: put object %q: %w", key, err)
	}

	return fmt.Sprintf("%s/%s/%s", s.publicBase, s.bucketName, key), nil
}

func (s *ObjectStorage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: remove object %q: %w", key, err)
	}

	return nil
}

// KeyFromURL извлекает ключ объекта из публичного URL.
// Используется для удаления старого аватара перед загрузкой нового.
func (s *ObjectStorage) KeyFromURL(avatarURL string) string {
	prefix := s.publicBase + "/" + s.bucketName + "/"
	if len(avatarURL) > len(prefix) && avatarURL[:len(prefix)] == prefix {
		return avatarURL[len(prefix):]
	}

	return ""
}
