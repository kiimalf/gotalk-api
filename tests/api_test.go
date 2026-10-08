package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gotalk-api/app/model"
	"gotalk-api/app/service"
	"gotalk-api/config"
	"gotalk-api/helper"
	"gotalk-api/route"
)

func TestAppHealthCheck(t *testing.T) {
	logger := config.NewLogger()
	app := config.NewApp(logger, route.Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable)
}

func TestCursorEncodingDecoding(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	var id int64 = 42

	encoded := helper.EncodeCursor(now, id)
	assert.NotEmpty(t, encoded)

	cursor, err := helper.DecodeCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, id, cursor.ID)
	assert.Equal(t, now.UnixNano(), cursor.CreatedAt.UnixNano())
}

func TestPasswordHashingAndVerification(t *testing.T) {
	plain := "SecretPass123"
	hashed, err := helper.HashPassword(plain)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)

	assert.True(t, helper.VerifyPassword(hashed, plain))
	assert.False(t, helper.VerifyPassword(hashed, "WrongPass123"))
}

func TestPermissionSetEvaluation(t *testing.T) {
	raw := map[string][]string{
		"admin": {"user:list", "user:create", "user:read:any"},
		"user":  {"chat:message:send"},
	}
	perms := helper.NewPermissionSet(raw)

	assert.True(t, perms.Can("admin", "user:list"))
	assert.False(t, perms.Can("user", "user:list"))
	assert.True(t, perms.Can("user", "chat:message:send"))
	assert.False(t, perms.Can("guest", "chat:message:send"))
}

func TestUserRules(t *testing.T) {
	// Tes IsEmptyPatch
	emptyReq := model.PatchUserRequest{}
	assert.True(t, service.IsEmptyPatch(emptyReq))

	newUsername := "johndoe"
	nonEmptyReq := model.PatchUserRequest{Username: &newUsername}
	assert.False(t, service.IsEmptyPatch(nonEmptyReq))

	// Tes ApplyPatch
	u := model.User{ID: 1, Username: "old", Email: "old@mail.com", Role: "user", IsActive: true}
	service.ApplyPatch(&u, nonEmptyReq)
	assert.Equal(t, "johndoe", u.Username)
	assert.Equal(t, "old@mail.com", u.Email)

	// Tes CanAccessUser
	authSelf := model.AuthUser{UserID: 1, Username: "johndoe", Role: "user"}
	authOther := model.AuthUser{UserID: 2, Username: "other", Role: "user"}
	authAdmin := model.AuthUser{UserID: 3, Username: "admin", Role: "admin"}

	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"user:read:any"},
	})

	assert.True(t, service.CanAccessUser(authSelf, 1, perms, "user:read:any"))
	assert.False(t, service.CanAccessUser(authOther, 1, perms, "user:read:any"))
	assert.True(t, service.CanAccessUser(authAdmin, 1, perms, "user:read:any"))
}
