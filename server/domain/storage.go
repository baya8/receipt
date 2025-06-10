package domain

import (
	"context"
	"io"
)

// StorageService はファイルストレージへのアクセスを抽象化するインターフェースです。
type StorageService interface {
	// UploadImage は画像をストレージにアップロードし、アクセス可能なURLを返します。
	UploadImage(ctx context.Context, file io.Reader, objectName string) (string, error)
}
