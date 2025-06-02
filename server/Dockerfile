# ビルドステージ: Goのビルド環境
FROM golang:1.22-alpine AS builder

# アプリケーションの作業ディレクトリを設定
WORKDIR /app

# go.mod と go.sum をコピーして依存関係を解決
# これにより、ソースコードが変更された場合のみ依存関係のダウンロードが再実行される
COPY go.mod go.sum ./
RUN go mod download

# アプリケーションのソースコードをコピー
# serverディレクトリ内のコードをビルド対象とする
COPY ./server ./server
COPY ./handlers ./handlers
COPY ./message ./message

# アプリケーションをビルド
# CGO_ENABLED=0: CGOを無効にし、静的リンクバイナリを生成
# -ldflags="-s -w": デバッグ情報を削除し、バイナリサイズを削減
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /app/receipt-server ./server/main.go

# ランタイムステージ: 軽量なAlpine Linuxイメージ
FROM alpine:latest

# ca-certificates をインストール (HTTPS通信などに必要)
RUN apk --no-cache add ca-certificates

# ビルドステージからコンパイル済みバイナリをコピー
COPY --from=builder /app/receipt-server /receipt-server

# アプリケーションがリッスンするポート (main.goで指定したポートと一致させる)
EXPOSE 8080

# コンテナ起動時に実行するコマンド
CMD ["/receipt-server"]