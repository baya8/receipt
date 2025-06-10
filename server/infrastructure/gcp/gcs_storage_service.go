package gcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// GCSStorageService は domain.StorageService のGCS実装です。
type GCSStorageService struct {
	client     *storage.Client
	bucketName string
}

// NewGCSStorageService は新しいGCSStorageServiceを生成します。
// GCS_BUCKET_NAME 環境変数からバケット名を取得します。
func NewGCSStorageService(ctx context.Context) (*GCSStorageService, error) {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("環境変数 GCS_BUCKET_NAME が設定されていません")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCSクライアントの作成に失敗しました: %w", err)
	}

	return &GCSStorageService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadImage は画像をGCSにアップロードし、公開URLを返します。
// objectName はGCS上のファイル名（パスを含む）です。
func (s *GCSStorageService) UploadImage(ctx context.Context, file io.Reader, originalFilename string) (string, error) {
	// GCS上のオブジェクト名を生成 (例: receipts/2023/10/uuid.jpg)
	objectName := fmt.Sprintf("receipts/%s/%s%s", time.Now().Format("2006/01"), uuid.NewString(), filepath.Ext(originalFilename))

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)

	wc := obj.NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("GCSへのファイルコピーに失敗しました: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("GCSライターのクローズに失敗しました: %w", err)
	}
	// バケットとオブジェクトが公開されていれば、この形式のURLでアクセス可能
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName)
	return publicURL, nil
}

// Close はGCSクライアントをクローズします。
func (s *GCSStorageService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
