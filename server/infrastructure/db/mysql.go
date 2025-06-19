package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // MySQLドライバをインポート
)

// NewDBConnection はデータベースへの新しい接続を確立します。
// 接続情報は環境変数から取得することを想定しています。
//
// プライベートIP経由でCloud SQLに接続する場合:
// DB_USER, DB_PASS, DB_NAME, DB_HOST (プライベートIP), DB_PORT
//
// (従来のcloudsqlconnを使用する場合: DB_USER, DB_PASS, DB_NAME, INSTANCE_CONNECTION_NAME)
func NewDBConnection() (*sql.DB, error) {
	var (
		dbUser = os.Getenv("DB_USER")
		dbPwd  = os.Getenv("DB_PASS")
		dbName = os.Getenv("DB_NAME")
		dbHost = os.Getenv("DB_HOST") // Cloud SQLのプライベートIPアドレス
		dbPort = os.Getenv("DB_PORT") // Cloud SQLのポート (通常は "3306")
	)

	if dbUser == "" || dbPwd == "" || dbName == "" || dbHost == "" || dbPort == "" {
		return nil, fmt.Errorf("データベース接続情報 (DB_USER, DB_PASS, DB_NAME, DB_HOST, DB_PORT) が環境変数に設定されていません")
	}

	// 標準的なMySQL DSN (Data Source Name) を使用
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPwd, dbHost, dbPort, dbName)
	log.Println("データベースに接続します:", fmt.Sprintf("%s:****@tcp(%s:%s)/%s?parseTime=true", dbUser, dbHost, dbPort, dbName))

	conn, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	return conn, nil
}
