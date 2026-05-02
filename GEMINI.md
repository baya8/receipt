# プロジェクト進捗状況 (2026-05-02)

## 概要
夫婦間で生活費を共有・精算するスマホ向けWebアプリの開発。

## 実装済み事項

### インフラ・ベース構成
- **モノレポ構成**: `client/` (Next.js), `server/` (Go) のディレクトリ構成を構築。
- **Docker**: MySQL 8.0 と Go サーバーのコンテナ環境を整備。
- **CORS**: フロントエンド (localhost:3000) とバックエンド (localhost:8080) の通信許可設定。

### バックエンド (Go / Gin / GORM)
- **DBモデル**: `users`, `groups`, `receipts`, `settlements` のテーブル定義とオートマイグレーションの実装。
- **認証**: 
  - ユーザー登録 (`POST /auth/register`)
  - ログイン・JWT発行 (`POST /auth/login`)
  - パスワードの bcrypt ハッシュ化。
- **レシート管理**: 
  - 登録 (`POST /api/receipts`)
  - 一覧取得 (`GET /api/receipts`)
  - 詳細取得 (`GET /api/receipts/:id`)
  - 更新 (`PUT /api/receipts/:id`)
  - 削除 (`DELETE /api/receipts/:id`)
- **グループ管理**: 
  - 作成 (`POST /api/groups`)
  - 所属一覧取得 (`GET /api/groups`)
- **サマリー・精算**: 
  - 月次集計 API (`GET /api/summary`) の実装。
  - 精算完了 API (`POST /api/settle`) の実装（一括でレシートを精算済みに更新）。
- **Gemini AI 解析**:
  - `POST /api/receipts/analyze` エンドポイントの実装。
  - `gemini-flash-latest` モデルを使用したレシート画像解析（店名・金額・日付・品名の抽出）。

### フロントエンド (Next.js / Tailwind CSS / shadcn/ui)
- **基本レイアウト**: モバイルサイズに最適化した外枠とボトムナビゲーション。
- **認証系画面**:
  - ログイン画面 (`/login`) の実装とAPI結合。
  - 新規登録画面 (`/signup`) の実装とAPI結合。
  - 全画面へのログインガード（未ログイン時のリダイレクト）追加。
- **メイン機能画面**:
  - レシート一覧画面 (`/`) のAPI結合（実データの表示）。
  - レシート登録画面 (`/register`) のAPI結合。
  - レシート詳細画面 (`/receipt/[id]`) のAPI結合（閲覧・更新・削除）。
  - サマリー画面 (`/summary`) のAPI結合（月次集計・支払いバランス表示・精算確定機能）。
- **Gemini AI 連携**:
  - レシート登録画面でのカメラ起動・画像アップロード機能。
  - 解析中のローディング表示と、解析結果のフォーム自動入力。
- **UI/UX 改善**:
  - `lib/api.ts` のエラーハンドリング強化（「Body has already been consumed」エラーの修正）。
  - 入力フィールドの文字色を `text-gray-900` に固定。

## 次にやるべきこと

1. **UX/UIのブラッシュアップ**
   - 保存・削除・精算成功時のトースト通知。
   - グループ作成・招待の UI 強化（現在はgroup_id=1固定のため）。
2. **テストと検証**
   - 実際のスマホ（ブラウザ）からの操作確認。
3. **データ永続化とバックアップ**
   - Dockerボリュームの設定確認とバックアップ手順の検討。

# ルール

何かを実装するとき、先に説明をしたうえで、実装に進むこと。
