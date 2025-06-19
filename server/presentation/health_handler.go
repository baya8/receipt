package presentation

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// HealthCheckResponse はヘルスチェックAPIのレスポンス構造体です。
type HealthCheckResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// HealthHandler はヘルスチェックリクエストを処理するためのハンドラです。
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler は新しいHealthHandlerを生成します。
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// ServeHTTP はヘルスチェックリクエストを処理します。
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second) // 5秒のタイムアウト
	defer cancel()

	dbStatus := "ok"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "error: " + err.Error()
		log.Printf("ヘルスチェック: データベース接続エラー: %v", err)
	}

	httpStatus := http.StatusOK
	overallStatus := "ok"

	if dbStatus != "ok" {
		httpStatus = http.StatusServiceUnavailable
		overallStatus = "error"
	}

	response := HealthCheckResponse{
		Status: overallStatus,
		Checks: map[string]string{
			"database": dbStatus,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("ヘルスチェックレスポンスのエンコードに失敗しました: %v", err)
	}
}
