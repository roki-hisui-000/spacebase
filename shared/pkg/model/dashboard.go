package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type OrderStatus int

const (
	StatusCompleted OrderStatus = 1
	StatusOrder     OrderStatus = 2
	StatusReject    OrderStatus = 3
	StatusError     OrderStatus = 4
)

var OrderStatusToString = map[OrderStatus]string{
	StatusCompleted: "completed",
	StatusOrder:     "order",
	StatusReject:    "reject",
	StatusError:     "error",
}

var StringToOrderStatus = map[string]OrderStatus{
	"completed": StatusCompleted,
	"order":     StatusOrder,
	"reject":    StatusReject,
	"error":     StatusError,
}

func (s OrderStatus) String() string {
	if val, ok := OrderStatusToString[s]; ok {
		return val
	}
	return fmt.Sprintf("unknown(%d)", s)
}

// RecentOrder represents the order history.
type RecentOrder struct {
	OrderID   string      `json:"orderId"`
	UserID    string      `json:"userId"`
	Price     float64     `json:"price"`
	Status    OrderStatus `json:"status"`
	RequestID string      `json:"requestId"`
	CreatedAt time.Time   `json:"createdAt"`
}

// MarshalJSON customizes JSON representation of RecentOrder to output Status as string.
func (r RecentOrder) MarshalJSON() ([]byte, error) {
	type Alias RecentOrder
	return json.Marshal(&struct {
		Status string `json:"status"`
		Alias
	}{
		Status: r.Status.String(),
		Alias:  (Alias)(r),
	})
}

// UnmarshalJSON customizes JSON decoding of RecentOrder to accept Status as string.
func (r *RecentOrder) UnmarshalJSON(data []byte) error {
	type Alias RecentOrder
	aux := &struct {
		Status string `json:"status"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if val, ok := StringToOrderStatus[aux.Status]; ok {
		r.Status = val
	} else {
		var statusInt int
		if _, err := fmt.Sscanf(aux.Status, "%d", &statusInt); err == nil {
			r.Status = OrderStatus(statusInt)
		} else {
			r.Status = 0
		}
	}
	return nil
}

// DashboardMetrics represents metrics information.
type DashboardMetrics struct {
	CurrentStock int `json:"currentStock"`
	QueueLength  int `json:"queueLength"`
	DbOrderCount int `json:"dbOrderCount"`
}

// Dashboard represents the overall dashboard model.
type Dashboard struct {
	WorkerStatus string           `json:"workerStatus"`
	Metrics      DashboardMetrics `json:"metrics"`
	RecentOrders []RecentOrder    `json:"recentOrders"`
}
