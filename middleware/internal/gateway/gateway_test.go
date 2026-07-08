package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/roki-hisui/work/spacebase/config"
	"github.com/roki-hisui/work/spacebase/internal/space"
)

// MockSpace for testing purposes.
type MockSpace struct{}

func (ms *MockSpace) Put(ctx context.Context, tuple space.Tuple) error {
	return nil
}

func (ms *MockSpace) Get(ctx context.Context, key string) (space.Tuple, error) {
	return space.Tuple{}, nil
}

func (ms *MockSpace) Keys(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}

func TestServeHTTPAdminAuthAndRouting(t *testing.T) {
	// configパッケージのAdminUser/AdminPassをテスト用に一時的に設定
	originalCfg := config.GetTestConfig()
	defer config.SetTestConfig(originalCfg)

	config.SetTestConfig(config.Config{
		AdminUser: "testadmin",
		AdminPass: "testpass",
	})

	gateway, err := NewSyncGateway(&MockSpace{})
	if err != nil {
		t.Fatalf("Failed to create SyncGateway: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		username       string
		password       string
		expectedStatus int
		expectedBody   string // Body content to check for
	}{
		{
			name:           "Admin path without auth",
			path:           "/admin",
			username:       "",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "Admin path with wrong auth",
			path:           "/admin/profiles",
			username:       "wrong",
			password:       "password",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized",
		},
		{
			name:           "Admin path with correct auth",
			path:           "/admin",
			username:       "testadmin",
			password:       "testpass",
			expectedStatus: http.StatusOK,
			expectedBody:   "<title>Spacebase - 管理画面</title>", // admin.html の一部
		},
		{
			name:           "Admin profiles path with correct auth",
			path:           "/admin/profiles",
			username:       "testadmin",
			password:       "testpass",
			expectedStatus: http.StatusOK,
			expectedBody:   "<title>Spacebase - 管理画面 (一覧)</title>", // admin_profiles.html の一部
		},
		{
			name:           "Admin regist path with correct auth",
			path:           "/admin/profiles/regist",
			username:       "testadmin",
			password:       "testpass",
			expectedStatus: http.StatusOK,
			expectedBody:   "<title>Spacebase - 管理画面 (登録)</title>", // admin_regist.html の一部
		},
		{
			name:           "Non-admin path without auth",
			path:           "/",
			username:       "",
			password:       "",
			expectedStatus: http.StatusOK,
			expectedBody:   "<title>Spacebase - ユーザー登録</title>", // web/index.html の一部
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.username != "" || tt.password != "" {
				req.SetBasicAuth(tt.username, tt.password)
			}
			rr := httptest.NewRecorder()

			gateway.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Path: %s, Expected status %d, got %d", tt.path, tt.expectedStatus, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("Path: %s, Expected body to contain '%s', got body:\n%s", tt.path, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestHandleAdminWebUI_Fallback(t *testing.T) {
	gateway, err := NewSyncGateway(&MockSpace{})
	if err != nil {
		t.Fatalf("Failed to create SyncGateway: %v", err)
	}

	reqNotFound := httptest.NewRequest("GET", "/admin/path/not/defined/in/switch", nil)
	rrNotFound := httptest.NewRecorder()
	gateway.handleAdminWebUI(rrNotFound, reqNotFound)

	// デフォルトのadmin.htmlが返されることを確認
	if rrNotFound.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rrNotFound.Code)
	}
	if !strings.Contains(rrNotFound.Body.String(), "<title>Spacebase - 管理画面</title>") {
		t.Errorf("Expected body to contain admin.html content, got:\n%s", rrNotFound.Body.String())
	}
}
