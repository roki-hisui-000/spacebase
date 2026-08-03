package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDashboardJSON_Unmarshal(t *testing.T) {
	inputJSON := `{
  "workerStatus": "running",
  "metrics": {
    "currentStock": 0,
    "queueLength": 65,
    "dbOrderCount": 35
  },
  "recentOrders": [
    { "orderId": "ord_89a3f", "userId": "user_42", "price": 150.0, "status": "completed", "requestId": "req_1", "createdAt": "2026-07-04T09:55:02Z" },
    { "orderId": "ord_12c7b", "userId": "user_11", "price": 200.0, "status": "completed", "requestId": "req_2", "createdAt": "2026-07-04T09:55:01Z" }
  ]
}`

	var db Dashboard
	err := json.Unmarshal([]byte(inputJSON), &db)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if db.WorkerStatus != "running" {
		t.Errorf("expected workerStatus running, got %s", db.WorkerStatus)
	}

	if db.Metrics.CurrentStock != 0 || db.Metrics.QueueLength != 65 || db.Metrics.DbOrderCount != 35 {
		t.Errorf("metrics values mismatch: %+v", db.Metrics)
	}

	if len(db.RecentOrders) != 2 {
		t.Fatalf("expected 2 recent orders, got %d", len(db.RecentOrders))
	}

	order1 := db.RecentOrders[0]
	if order1.OrderID != "ord_89a3f" {
		t.Errorf("expected orderId ord_89a3f, got %s", order1.OrderID)
	}
	if order1.Status != StatusCompleted {
		t.Errorf("expected status completed (1), got %d", order1.Status)
	}
	if order1.Price != 150.0 {
		t.Errorf("expected price 150.0, got %f", order1.Price)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-07-04T09:55:02Z")
	if !order1.CreatedAt.Equal(expectedTime) {
		t.Errorf("expected createdAt %v, got %v", expectedTime, order1.CreatedAt)
	}
}

func TestDashboardJSON_Marshal(t *testing.T) {
	createdAt, _ := time.Parse(time.RFC3339, "2026-07-04T09:55:02Z")
	db := Dashboard{
		WorkerStatus: "running",
		Metrics: DashboardMetrics{
			CurrentStock: 10,
			QueueLength:  5,
			DbOrderCount: 100,
		},
		RecentOrders: []RecentOrder{
			{
				OrderID:   "ord_999",
				UserID:    "user_abc",
				Price:     500.50,
				Status:    StatusCompleted,
				RequestID: "req_xyz",
				CreatedAt: createdAt,
			},
		},
	}

	data, err := json.Marshal(db)
	if err != nil {
		t.Fatalf("failed to marshal Dashboard: %v", err)
	}

	// Unmarshal back into a map to check actual JSON representation of status
	var raw map[string]interface{}
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("failed to unmarshal into map: %v", err)
	}

	recentOrders, ok := raw["recentOrders"].([]interface{})
	if !ok || len(recentOrders) != 1 {
		t.Fatalf("recentOrders not serialized properly")
	}

	order := recentOrders[0].(map[string]interface{})
	status, ok := order["status"].(string)
	if !ok {
		t.Fatalf("status should be a string in JSON, but was %v", order["status"])
	}

	if status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", status)
	}
}
