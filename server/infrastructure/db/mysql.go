package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/go-sql-driver/mysql"
)

// NewDBConnection はCloud SQL (MySQL) への新しいデータベース接続を確立します。
// 接続情報は環境変数から取得することを想定しています。
//
//	DB_USER, DB_PASS, DB_NAME, INSTANCE_CONNECTION_NAME
func NewDBConnection() (*sql.DB, error) {
	var (
		dbUser                 = os.Getenv("DB_USER")
		dbPwd                  = os.Getenv("DB_PASS")
		dbName                 = os.Getenv("DB_NAME")
		instanceConnectionName = os.Getenv("INSTANCE_CONNECTION_NAME")
		// usePrivate             = os.Getenv("PRIVATE_IP") // 必要に応じてプライベートIP接続も考慮
	)

	if dbUser == "" || dbPwd == "" || dbName == "" || instanceConnectionName == "" {
		return nil, fmt.Errorf("データベース接続情報 (DB_USER, DB_PASS, DB_NAME, INSTANCE_CONNECTION_NAME) が環境変数に設定されていません")
	}

	d, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}

	mysql.RegisterDialContext("cloudsqlconn",
		func(ctx context.Context, addr string) (net.Conn, error) {
			return d.Dial(ctx, instanceConnectionName)
		})

	dbURI := fmt.Sprintf("%s:%s@cloudsqlconn(localhost:3306)/%s?parseTime=true", dbUser, dbPwd, dbName)
	log.Println("接続中:", dbURI) // パスワードはログに出力しないように注意

	conn, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	return conn, nil
}
