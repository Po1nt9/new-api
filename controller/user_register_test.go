package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/api/user/register", Register)
	return r
}

func TestRegisterWithInvitationScenarios(t *testing.T) {
	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousType := common.MainDatabaseType()
	previousRedis := common.RedisEnabled
	previousMemoryCache := common.MemoryCacheEnabled
	previousSecret := common.SessionSecret

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Invitation{}, &model.Log{}))
	model.DB = db
	model.LOG_DB = db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.MemoryCacheEnabled = false
	common.SessionSecret = "test-secret-12345678"

	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.SetMainDatabaseType(previousType)
		common.RedisEnabled = previousRedis
		common.MemoryCacheEnabled = previousMemoryCache
		common.SessionSecret = previousSecret
		common.InvitationCodeRequired = false
		common.RegisterEnabled = true
		common.PasswordRegisterEnabled = true
		common.EmailVerificationEnabled = false
	})

	common.RegisterEnabled = true
	common.PasswordRegisterEnabled = true
	common.EmailVerificationEnabled = false
	common.QuotaForNewUser = 1000

	r := setupTestRouter()
	now := common.GetTimestamp()

	// Seed invitation codes
	invValid := model.Invitation{
		Key:         "INV-VALID-01",
		Status:      common.InvitationCodeStatusEnabled,
		Quota:       5000,
		Group:       "vip",
		CreatedTime: now,
	}
	require.NoError(t, invValid.Insert())

	invExpired := model.Invitation{
		Key:         "INV-EXPIRED-01",
		Status:      common.InvitationCodeStatusEnabled,
		Quota:       5000,
		ExpiredTime: now - 100,
		CreatedTime: now - 200,
	}
	require.NoError(t, invExpired.Insert())

	invDisabled := model.Invitation{
		Key:         "INV-DISABLED-01",
		Status:      common.InvitationCodeStatusDisabled,
		Quota:       5000,
		CreatedTime: now,
	}
	require.NoError(t, invDisabled.Insert())

	invUsed := model.Invitation{
		Key:         "INV-USED-01",
		Status:      common.InvitationCodeStatusUsed,
		Quota:       5000,
		CreatedTime: now,
	}
	require.NoError(t, invUsed.Insert())

	t.Run("Mandatory mode: missing invitation code rejected", func(t *testing.T) {
		common.InvitationCodeRequired = true
		defer func() { common.InvitationCodeRequired = false }()

		body, _ := json.Marshal(map[string]string{
			"username": "user_no_code",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("Mandatory mode: invalid invitation code rejected", func(t *testing.T) {
		common.InvitationCodeRequired = true
		defer func() { common.InvitationCodeRequired = false }()

		body, _ := json.Marshal(map[string]string{
			"username":        "user_bad_code",
			"password":        "password123",
			"invitation_code": "INV-NOT-EXIST",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("Mandatory mode: expired invitation code rejected", func(t *testing.T) {
		common.InvitationCodeRequired = true
		defer func() { common.InvitationCodeRequired = false }()

		body, _ := json.Marshal(map[string]string{
			"username":        "user_exp_code",
			"password":        "password123",
			"invitation_code": "INV-EXPIRED-01",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("Mandatory mode: already-used invitation code rejected", func(t *testing.T) {
		common.InvitationCodeRequired = true
		defer func() { common.InvitationCodeRequired = false }()

		body, _ := json.Marshal(map[string]string{
			"username":        "user_used_code",
			"password":        "password123",
			"invitation_code": "INV-USED-01",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("Valid invitation code succeeds and provisions quota & group", func(t *testing.T) {
		common.InvitationCodeRequired = true
		defer func() { common.InvitationCodeRequired = false }()

		body, _ := json.Marshal(map[string]string{
			"username":        "user_vip_registered",
			"password":        "password123",
			"invitation_code": "INV-VALID-01",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])

		// Verify user properties
		var createdUser model.User
		require.NoError(t, model.DB.Where("username = ?", "user_vip_registered").First(&createdUser).Error)
		assert.Equal(t, 1000+5000, createdUser.Quota)
		assert.Equal(t, "vip", createdUser.Group)

		// Verify invitation code marked used
		invCheck, err := model.GetInvitationByKey("INV-VALID-01")
		require.NoError(t, err)
		assert.Equal(t, common.InvitationCodeStatusUsed, invCheck.Status)
		assert.Equal(t, createdUser.Id, invCheck.UsedUserId)
	})

	t.Run("Optional mode without code succeeds with defaults", func(t *testing.T) {
		common.InvitationCodeRequired = false

		body, _ := json.Marshal(map[string]string{
			"username": "user_optional_open",
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])

		var createdUser model.User
		require.NoError(t, model.DB.Where("username = ?", "user_optional_open").First(&createdUser).Error)
		assert.Equal(t, 1000, createdUser.Quota)
		assert.Equal(t, "default", createdUser.Group)
	})

	t.Run("Affiliate link registration binds inviter without failing on invitation code", func(t *testing.T) {
		common.InvitationCodeRequired = false

		// Seed inviter
		inviter := model.User{
			Username:    "parent_inviter",
			Password:    "password123",
			DisplayName: "Parent",
			AffCode:     "AFF_PARENT_CODE",
		}
		require.NoError(t, model.DB.Create(&inviter).Error)

		body, _ := json.Marshal(map[string]string{
			"username": "invited_child_user",
			"password": "password123",
			"aff_code": "AFF_PARENT_CODE",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])

		var createdUser model.User
		require.NoError(t, model.DB.Where("username = ?", "invited_child_user").First(&createdUser).Error)
		assert.Equal(t, inviter.Id, createdUser.InviterId)
	})
}
