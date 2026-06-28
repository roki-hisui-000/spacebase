package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/roki-hisui/work/spacebase/internal/service"
	"github.com/roki-hisui/work/spacebase/pkg/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	service *service.ProfileService
}

func NewHandler(rdb *redis.Client) *Handler {
	svc := service.NewProfileService(rdb)
	return &Handler{
		service: svc,
	}
}

// CreateProfileHandler handles POST /profiles
func (h *Handler) CreateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var profile model.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "不正なリクエストデータです", http.StatusBadRequest)
		return
	}
	profile.ID = uuid.New().String()

	ctx := context.Background()
	if err := h.service.SaveProfile(ctx, &profile); err != nil {
		http.Error(w, "プロファイルの保存に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

// GetProfileHandler handles GET /profiles/{id}
func (h *Handler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	ctx := context.Background()
	profile, err := h.service.GetProfile(ctx, id)
	if err != nil {
		http.Error(w, "プロファイルが見つかりません", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(profile)
}
