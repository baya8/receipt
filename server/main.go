package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"receipt/application"
	"receipt/infrastructure/db"
	"receipt/infrastructure/gcp"
	"receipt/infrastructure/repository"
	"receipt/presentation"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

func main() {
	// デフォルトの.envファイルを最初に読み込む試み。
	// プロファイル固有のファイルで上書きされる可能性がある。
	// ファイルが存在しなくてもエラーにはしない。
	if err := godotenv.Load(); err != nil {
		log.Println("情報: デフォルトの.envファイルが見つからないか、読み込みに失敗しました。")
	}

	// コマンドライン引数からプロファイルを読み込む
	profile := pflag.String("profile", "", "Specify the environment profile (e.g., local, production)")
	pflag.Parse()

	// プロファイルが指定されていれば、対応する.env.<profile>ファイルを読み込み、
	// 既存の環境変数を上書きする。
	loadProfileEnvFile(*profile)

	log.Println("サーバーを http://localhost:8080 で起動します")
	log.Println("利用可能なエンドポイント:")

	// データベース接続の初期化
	// 環境変数 (DB_USER, DB_PASS, DB_NAME, DB_HOST, DB_PORT など) の設定が必要です。
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
	defer storageService.Close() // GCSStorageServiceにCloseメソッドがあれば呼び出す

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

	// ヘルスチェックハンドラの初期化とエンドポイントの登録
	healthHandler := presentation.NewHealthHandler(sqlDB)
	http.HandleFunc("/health", healthHandler.ServeHTTP)
	log.Println("GET /health")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}

// loadProfileEnvFile は指定されたプロファイルに基づいて.env.<profile>ファイルを読み込み、
// 既存の環境変数を上書きします。
func loadProfileEnvFile(profile string) {
	if profile == "" {
		log.Println("情報: プロファイルが指定されていません。デフォルトの.envまたはシステム環境変数を使用します。")
		return
	}

	filename := fmt.Sprintf(".env.%s", profile)
	log.Printf("情報: プロファイル '%s' に基づいて %s を読み込み、既存の環境変数を上書きします。", profile, filename)

	// godotenv.Overload は指定されたファイルを読み込み、既存の環境変数を上書きします。
	// ファイルが存在しない場合はエラーを返します。
	if err := godotenv.Overload(filename); err != nil {
		log.Printf("警告: %s の読み込みに失敗しました: %v。デフォルトの.envまたはシステム環境変数が使用されます。", filename, err)
	}
}
