package handlers

import (
	"errors"
	"net/http"
	"receipt/server/internal/service"

	"github.com/gin-gonic/gin"
)

// serviceErrorMapping Service層のエラーとHTTPステータスコード・メッセージのマッピング
var serviceErrorMapping = []struct {
	err    error
	status int
	msg    string
}{
	// User
	{service.ErrInvalidCredentials, http.StatusUnauthorized, "Invalid email or password"},
	{service.ErrUserNotFound, http.StatusUnauthorized, "User record not found"},

	// Group
	{service.ErrGroupNotFound, http.StatusNotFound, "Group not found"},
	{service.ErrNotOwner, http.StatusForbidden, "Only group owner can perform this action"},
	{service.ErrInviteUserNotFound, http.StatusNotFound, service.ErrInviteUserNotFound.Error()},
	{service.ErrAlreadyMember, http.StatusBadRequest, "User is already a member of this group"},
	{service.ErrOwnerCannotBeRemoved, http.StatusBadRequest, "Owner cannot be removed from the group"},
	{service.ErrMemberNotFound, http.StatusNotFound, "User to remove not found"},

	// Receipt
	{service.ErrReceiptNotFound, http.StatusNotFound, "Receipt not found"},
	{service.ErrNotCreator, http.StatusForbidden, "Only the creator can modify this receipt"},
	{service.ErrAlreadySettled, http.StatusForbidden, "精算済みのレシートは変更できません"},
	{service.ErrInvalidAmount, http.StatusBadRequest, "金額は1円以上にしてください"},

	// Settlement
	{service.ErrInvalidSettlementAmount, http.StatusBadRequest, "精算金額は1円以上にしてください"},
}

// respondWithServiceError Service層のエラーに応じたHTTPレスポンスを返す。
// マッピングに一致するものがある場合はそのステータスコードとメッセージで返し、
// 一致しない場合は 500 Internal Server Error を返す。
// 戻り値：エラーがマッピングに一致した場合はtrue
func respondWithServiceError(c *gin.Context, err error) bool {
	for _, m := range serviceErrorMapping {
		if errors.Is(err, m.err) {
			c.JSON(m.status, gin.H{"error": m.msg})
			return true
		}
	}
	return false
}

// respondInternalError 汎用的な500エラーレスポンスを返す
func respondInternalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
