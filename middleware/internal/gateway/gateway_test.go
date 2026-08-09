package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// MockSpaceWithData holds data in memory for dashboard testing.
type MockSpaceWithData struct {
	data map[string][]byte
}

func NewMockSpaceWithData() *MockSpaceWithData {
	return &MockSpaceWithData{
		data: make(map[string][]byte),
	}
}

func (ms *MockSpaceWithData) Put(ctx context.Context, tuple space.Tuple) error {
	ms.data[tuple.Key] = tuple.Value
	return nil
}

func (ms *MockSpaceWithData) Get(ctx context.Context, key string) (space.Tuple, error) {
	val, ok := ms.data[key]
	if !ok {
		return space.Tuple{}, http.ErrMissingBoundary // エラーを返す (何のエラーでもよい)
	}
	return space.Tuple{Key: key, Value: val}, nil
}

func (ms *MockSpaceWithData) Keys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	prefix := strings.TrimSuffix(pattern, "*")
	for k := range ms.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func TestServeHTTPAdminAuthAndRouting(t *testing.T) {
	// configパッケージのAdminUser/AdminPassをテスト用に一時的に設定
	originalCfg := config.GetTestConfig()
	defer config.SetTestConfig(originalCfg)

	config.SetTestConfig(config.Config{
		AdminUser: "testadmin",
		AdminPass: "testpass",
	})

	gateway := &SyncGateway{
		client: nil,
		addr:   "localhost:50051",
		sp:     &MockSpace{},
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
	gateway := &SyncGateway{
		client: nil,
		addr:   "localhost:50051",
		sp:     &MockSpace{},
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

func TestDashboardAPI_PostAndGet(t *testing.T) {
	mockSpace := NewMockSpaceWithData()
	gateway := &SyncGateway{
		client: nil,
		addr:   "localhost:50051",
		sp:     mockSpace,
	}

	// 1. POST /api/dashboard/orders で注文を永続化
	orderPayload := `{"userId": "user_123", "price": 500, "status": "completed", "requestId": "req_abc"}`
	reqPost := httptest.NewRequest("POST", "/api/dashboard/orders", strings.NewReader(orderPayload))
	rrPost := httptest.NewRecorder()

	gateway.ServeHTTP(rrPost, reqPost)

	if rrPost.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rrPost.Code)
	}

	var createdOrder map[string]interface{}
	if err := json.Unmarshal(rrPost.Body.Bytes(), &createdOrder); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if createdOrder["userId"] != "user_123" || createdOrder["price"] != float64(500) || createdOrder["status"] != "completed" {
		t.Errorf("Saved order details mismatch: %+v", createdOrder)
	}
	orderID, ok := createdOrder["orderId"].(string)
	if !ok || !strings.HasPrefix(orderID, "ord_") {
		t.Errorf("Invalid orderId generated: %s", orderID)
	}

	// 2. GET /api/dashboard でダッシュボード情報を取得
	reqGet := httptest.NewRequest("GET", "/api/dashboard", nil)
	rrGet := httptest.NewRecorder()

	gateway.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rrGet.Code)
	}

	var dashboard map[string]interface{}
	if err := json.Unmarshal(rrGet.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("Failed to unmarshal dashboard JSON: %v", err)
	}

	if dashboard["workerStatus"] != "running" {
		t.Errorf("Expected workerStatus running, got %v", dashboard["workerStatus"])
	}

	recentOrders, ok := dashboard["recentOrders"].([]interface{})
	if !ok || len(recentOrders) != 1 {
		t.Fatalf("Expected 1 recent order, got %+v", dashboard["recentOrders"])
	}

	fetchedOrder := recentOrders[0].(map[string]interface{})
	if fetchedOrder["orderId"] != orderID {
		t.Errorf("Fetched orderId mismatch. Expected %s, got %v", orderID, fetchedOrder["orderId"])
	}
}

func TestDashboardAPI_Streaming(t *testing.T) {
	mockSpace := NewMockSpaceWithData()
	gateway := &SyncGateway{
		client: nil,
		addr:   "localhost:50051",
		sp:     mockSpace,
	}

	// ストリーミング接続のシミュレーション。context をキャンセルして無限ループを防止
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reqStream := httptest.NewRequest("GET", "/api/dashboard?stream=true", nil).WithContext(ctx)
	rrStream := httptest.NewRecorder()

	// 1秒後にストリーム接続を終了するゴルーチン
	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	gateway.ServeHTTP(rrStream, reqStream)

	// SSE レスポンスヘッダの検証
	contentType := rrStream.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("Expected Content-Type to contain 'text/event-stream', got '%s'", contentType)
	}

	bodyStr := rrStream.Body.String()
	if !strings.Contains(bodyStr, "data:") {
		t.Errorf("Expected body to contain event stream data format, got:\n%s", bodyStr)
	}
}
