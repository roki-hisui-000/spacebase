package gateway

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/roki-hisui/work/spacebase/config"
	processingpb "github.com/roki-hisui/work/spacebase/internal/processing"
	"github.com/roki-hisui/work/spacebase/internal/space"
	"github.com/roki-hisui/work/spacebase/pkg/model"
	sharedmodel "github.com/roki-hisui/work/spacebase/shared/pkg/model"

	"google.golang.org/grpc"
)

//go:embed web/*.html
var webFS embed.FS

// Gateway はパブリックAPIを受け、適切なProcessingUnitへルーティングするインターフェース
type Gateway interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// SyncGateway はgRPCを使って同期呼び出しを行う実装
type SyncGateway struct {
	client processingpb.ProcessingUnitClient
	addr   string
	sp     space.Space
}

// NewSyncGateway でgRPCクライアントを初期化
func NewSyncGateway(sp space.Space) (*SyncGateway, error) {
	addr := os.Getenv("PROCESSING_UNIT_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	client := processingpb.NewProcessingUnitClient(conn)
	return &SyncGateway{client: client, addr: addr, sp: sp}, nil
}

// ServeHTTP はリクエストのURLによってAPI処理と画面（WebUI）配信にルーティングします
func (g *SyncGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[HTTP Sync Request] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// 1. パスが "/admin" で始まる場合はBasic認証をかける
	if strings.HasPrefix(r.URL.Path, "/admin") {
		username, password, ok := r.BasicAuth()
		if !ok || username != config.AdminUser() || password != config.AdminPass() {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		g.handleAdminWebUI(w, r)
		return
	}

	// 2. パスが "/api" で始まる場合はAPIリクエストとして処理
	if strings.HasPrefix(r.URL.Path, "/api") {
		g.handleAPI(w, r)
		return
	}

	// 3. それ以外の場合は画面（静的ファイル）配信として処理
	g.handleWebUI(w, r)
}

// handleAPI は各種バックエンドAPIの呼び出しを処理します
func (g *SyncGateway) handleAPI(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch r.URL.Path {
	case "/api/register":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handleRegister(w, r, ctx)

	case "/api/profiles":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handleGetProfiles(w, r, ctx)

	case "/api/dashboard/orders":
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handlePostDashboardOrder(w, r, ctx)

	case "/api/dashboard":
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		g.handleGetDashboard(w, r, ctx)

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// handleRegister は JSON パースを行い、gRPC RegisterUser へフォワードします
func (g *SyncGateway) handleRegister(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var payload struct {
		Email     string   `json:"email"`
		Status    string   `json:"status"`
		Languages []string `json:"languages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// メールアドレスの重複チェック
	if g.sp != nil {
		keys, err := g.sp.Keys(ctx, "user:*:profile")
		if err == nil {
			for _, key := range keys {
				tuple, err := g.sp.Get(ctx, key)
				if err != nil {
					continue
				}

				var existingProfile model.UserProfile
				if err := json.Unmarshal(tuple.Value, &existingProfile); err == nil {
					if strings.EqualFold(existingProfile.Email, payload.Email) {
						http.Error(w, "メールアドレスは既に登録されています。", http.StatusConflict)
						return
					}
				} else {
					// プレーンテキストで保存されている場合を考慮
					if strings.EqualFold(string(tuple.Value), payload.Email) {
						http.Error(w, "メールアドレスは既に登録されています。", http.StatusConflict)
						return
					}
				}
			}
		}
	}

	req := &processingpb.RegisterUserRequest{
		Name:      payload.Email,
		Status:    payload.Status,
		Languages: payload.Languages,
	}
	res, err := g.client.RegisterUser(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res.Message))
}

// handleGetProfiles は Space(Valkey) から全てのユーザーデータを検索・取得して返します
func (g *SyncGateway) handleGetProfiles(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if g.sp == nil {
		http.Error(w, "Space adapter is not configured", http.StatusInternalServerError)
		return
	}

	// "user:*:profile" パターンに一致するキー一覧を検索
	keys, err := g.sp.Keys(ctx, "user:*:profile")
	if err != nil {
		http.Error(w, "Failed to retrieve keys: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var profiles []model.UserProfile
	for _, key := range keys {
		tuple, err := g.sp.Get(ctx, key)
		if err != nil {
			continue // 一部取得エラーはスキップ
		}

		var profile model.UserProfile
		// JSON形式でのパースを試みる
		if err := json.Unmarshal(tuple.Value, &profile); err != nil {
			// パースに失敗した場合は、プレーンテキストとして扱い、Languagesは"error"とする
			parts := strings.Split(key, ":")
			id := ""
			if len(parts) >= 2 {
				id = parts[1]
			}
			profile = model.UserProfile{
				ID:        id,
				Email:     string(tuple.Value),
				Status:    "active",
				Languages: []string{"error"},
			}
		} else {
			// JSONパースに成功したが、IDが空の場合はキーから補完
			if profile.ID == "" {
				parts := strings.Split(key, ":")
				if len(parts) >= 2 {
					profile.ID = parts[1]
				}
			}
		}
		profiles = append(profiles, profile)
	}

	if profiles == nil {
		profiles = []model.UserProfile{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(profiles); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// handlePostDashboardOrder handles POST /api/dashboard/orders (creates a new order to persist)
func (g *SyncGateway) handlePostDashboardOrder(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if g.sp == nil {
		http.Error(w, "Space adapter is not configured", http.StatusInternalServerError)
		return
	}

	var payload struct {
		UserID    string `json:"userId"`
		Price     uint64 `json:"price"`
		Status    string `json:"status"`
		RequestID string `json:"requestId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderID := "ord_" + uuid.New().String()

	statusVal := sharedmodel.StatusUnknown
	if val, ok := sharedmodel.StringToOrderStatus[payload.Status]; ok {
		statusVal = val
	}

	order := sharedmodel.RecentOrder{
		OrderID:   orderID,
		UserID:    payload.UserID,
		Price:     payload.Price,
		Status:    statusVal,
		RequestID: payload.RequestID,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(order)
	if err != nil {
		http.Error(w, "failed to marshal order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = g.sp.Put(ctx, space.Tuple{
		Key:   "order:" + orderID,
		Value: data,
	})
	if err != nil {
		http.Error(w, "failed to save order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// handleGetDashboard handles GET /api/dashboard (returns current dashboard metrics and recent orders)
// Supports Server-Sent Events (SSE) streaming if stream=true query parameter or Accept: text/event-stream header is specified.
func (g *SyncGateway) handleGetDashboard(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	if g.sp == nil {
		http.Error(w, "Space adapter is not configured", http.StatusInternalServerError)
		return
	}

	stream := r.URL.Query().Get("stream") == "true" || strings.Contains(r.Header.Get("Accept"), "text/event-stream")

	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// 初回のデータを即時に送信
		db, err := g.fetchDashboardData(ctx)
		if err == nil {
			if data, err := json.Marshal(db); err == nil {
				_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
				flusher.Flush()
			}
		}

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				db, err := g.fetchDashboardData(ctx)
				if err != nil {
					continue
				}
				data, err := json.Marshal(db)
				if err != nil {
					continue
				}
				_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
				flusher.Flush()
			}
		}
	} else {
		db, err := g.fetchDashboardData(ctx)
		if err != nil {
			http.Error(w, "failed to fetch dashboard data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(db); err != nil {
			http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (g *SyncGateway) fetchDashboardData(ctx context.Context) (sharedmodel.Dashboard, error) {
	var db sharedmodel.Dashboard

	workerStatusVal := "running"
	if tuple, err := g.sp.Get(ctx, "dashboard:worker_status"); err == nil && len(tuple.Value) > 0 {
		workerStatusVal = string(tuple.Value)
	}
	db.WorkerStatus = workerStatusVal

	metricsVal := sharedmodel.DashboardMetrics{
		CurrentStock: 0,
		QueueLength:  0,
		DbOrderCount: 0,
	}
	if tuple, err := g.sp.Get(ctx, "dashboard:metrics"); err == nil && len(tuple.Value) > 0 {
		var m sharedmodel.DashboardMetrics
		if err := json.Unmarshal(tuple.Value, &m); err == nil {
			metricsVal = m
		}
	}
	db.Metrics = metricsVal

	keys, err := g.sp.Keys(ctx, "order:*")
	if err != nil {
		return db, err
	}

	var orders []sharedmodel.RecentOrder
	for _, key := range keys {
		tuple, err := g.sp.Get(ctx, key)
		if err != nil {
			continue
		}
		var order sharedmodel.RecentOrder
		if err := json.Unmarshal(tuple.Value, &order); err == nil {
			orders = append(orders, order)
		}
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	if orders == nil {
		orders = []sharedmodel.RecentOrder{}
	}
	db.RecentOrders = orders

	return db, nil
}

// handleWebUI は埋め込まれた Web 画面を返却します
func (g *SyncGateway) handleWebUI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// 埋め込みFS内のパスに調整 (web/index.html 形式)
	embedPath := "web" + path

	// ファイルが存在するか確認して配信
	data, err := webFS.ReadFile(embedPath)
	if err != nil {
		// profiles 画面への対応
		if path == "/profiles" || path == "/profiles/" {
			data, err = webFS.ReadFile("web/profiles.html")
			if err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
				return
			}
		}
		http.NotFound(w, r)
		return
	}

	contentType := "text/html; charset=utf-8"
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// handleAdminWebUI は管理画面（静的HTML）を返却します
func (g *SyncGateway) handleAdminWebUI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	var template string
	switch path {
	case "/admin", "/admin/":
		template = "web/admin.html"
	case "/admin/profiles", "/admin/profiles/":
		template = "web/admin_profiles.html"
	case "/admin/profiles/regist", "/admin/profiles/regist/":
		template = "web/admin_regist.html"
	default:
		// デフォルトは admin.html (topページ)
		template = "web/admin.html"
	}

	data, err := webFS.ReadFile(template)
	if err != nil {
		http.Error(w, "Admin template not found: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// NewGateway creates a Gateway based on GATEWAY_MODE env var.
func NewGateway(sp space.Space) (Gateway, error) {
	mode := os.Getenv("GATEWAY_MODE")
	switch mode {
	case "async":
		g := NewAsyncGateway()
		return g, nil
	default:
		return NewSyncGateway(sp)
	}
}
