package main

import (
	"context"
	"log"
	"net/http"

	"receipt/application"
	"receipt/infrastructure/db"
	"receipt/infrastructure/gcp"
	"receipt/infrastructure/repository"
	"receipt/presentation"
)

func main() {
	log.Println("サーバーを http://localhost:8080 で起動します")
	log.Println("利用可能なエンドポイント:")

	// データベース接続の初期化
	// 環境変数 (DB_USER, DB_PASS, DB_NAME, INSTANCE_CONNECTION_NAME) の設定が必要です。
	sqlDB, err := db.NewDBConnection()
	if err != nil {
		log.Fatalf("データベース接続の初期化に失敗しました: %v", err)
	}
	defer sqlDB.Close()

	// GCSストレージサービスの初期化
	// 環境変数 (GCS_BUCKET_NAME) の設定が必要です。
	storageService, err := gcp.NewGCSStorageService(context.Background())
	if err != nil {
		log.Fatalf("GCSストレージサービスの初期化に失敗しました: %v", err)
	}
	// defer storageService.Close() // GCSStorageServiceにCloseメソッドがあれば呼び出す

	// インフラストラクチャ層のサービスとリポジトリを初期化
	receiptRepo := repository.NewMySQLReceiptRepository(sqlDB)
	geminiSvc := gcp.NewGeminiClient() // TODO: Geminiクライアントの初期化パラメータを設定

	// アプリケーション層のサービスを初期化
	receiptService := application.NewReceiptService(receiptRepo, geminiSvc, storageService)

	// プレゼンテーション層のハンドラを初期化し、サービスを注入
	receiptHandler := presentation.NewReceiptHandler(receiptService)

	// ルーター設定
	http.HandleFunc("/receipts", receiptHandler.CreateReceiptHandler)
	log.Println("POST /receipts (date, payer, paymentMethod, receiptImage)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
