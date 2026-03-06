package storage

import (
	"context"
	"io"
	"github.com/minio/minio-go/v7"
)

type MinIOStorage struct {
	client *minio.Client
}

func NewMinIOStorage(client *minio.Client) *MinIOStorage {
	return &MinIOStorage{client: client}
}

func (s *MinIOStorage) UploadFile(ctx context.Context, bucket string, objectName string, fileStream io.Reader, fileSize int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, objectName, fileStream, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}