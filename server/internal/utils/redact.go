package utils

import "regexp"

// 一部の外部APIクライアント（GeminiのSDK等）はAPIキーをリクエストURLの
// クエリパラメータに含めるため、タイムアウト等のエラーメッセージにURLごと
// 出力されそのままログに残ってしまうことがある。ログに残す前に必ず除去する。
var secretQueryParamPattern = regexp.MustCompile(`([?&](?:key|access_token)=)[^&\s"]+`)

// RedactSecrets エラーメッセージからAPIキー等の秘密情報を除去した文字列を返す
func RedactSecrets(err error) string {
	return secretQueryParamPattern.ReplaceAllString(err.Error(), "${1}[REDACTED]")
}
