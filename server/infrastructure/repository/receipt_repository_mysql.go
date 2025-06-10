package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"receipt/domain"

	"github.com/google/uuid"
)

// MySQLReceiptRepository は domain.ReceiptRepository のMySQL実装です。
type MySQLReceiptRepository struct {
	db *sql.DB
}

// NewMySQLReceiptRepository は新しいMySQLReceiptRepositoryを生成します。
func NewMySQLReceiptRepository(db *sql.DB) *MySQLReceiptRepository {
	return &MySQLReceiptRepository{db: db}
}

// Save はレシート情報をデータベースに保存します。
func (r *MySQLReceiptRepository) Save(ctx context.Context, receipt *domain.Receipt) error {
	receipt.ID = uuid.NewString() // 新規作成時にUUIDを生成
	receipt.CreatedAt = time.Now()
	receipt.UpdatedAt = time.Now()

	query := `INSERT INTO receipts (id, date, store, items, total_amount, payer, payment_method, image_url, created_at, updated_at)
	           VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, receipt.ID, receipt.Date, receipt.Store, receipt.Items, receipt.TotalAmount, receipt.Payer, receipt.PaymentMethod, receipt.ImageURL, receipt.CreatedAt, receipt.UpdatedAt)
	if err != nil {
		return fmt.Errorf("レシートの保存に失敗しました (ExecContext): %w", err)
	}
	return nil
}
