package controllers

import (
	"bakeflow/configs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ChatOrderSchedule struct {
	Type string `json:"type"`
	Date string `json:"date"`
	Time string `json:"time"`
}

type ChatOrderRequest struct {
	UserID        string             `json:"user_id"`
	Items         []ChatOrderItem    `json:"items"`
	Channel       string             `json:"channel"`
	Notes         string             `json:"notes"`
	CustomerName  string             `json:"customer_name"`
	CustomerPhone string             `json:"customer_phone"`
	DeliveryType  string             `json:"delivery_type"`
	Address       string             `json:"address"`
	Schedule      *ChatOrderSchedule `json:"schedule"`
}

type ChatOrderItem struct {
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

type ChatOrderResponse struct {
	Success bool   `json:"success"`
	OrderID int    `json:"order_id"`
	Message string `json:"message"`
}

// CreateChatOrder handles orders from the mini webview
func CreateChatOrder(w http.ResponseWriter, r *http.Request) {
	var req ChatOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Invalid request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Determine user identity.
	// Preferred: signed token (?t=...) which encodes PSID.
	userID := strings.TrimSpace(req.UserID)
	if tok := strings.TrimSpace(r.URL.Query().Get("t")); tok != "" {
		if psid, err := VerifyWebviewToken(tok); err == nil {
			userID = psid
		} else {
			log.Printf("⚠️  Invalid webview token: %v", err)
		}
	}
	if userID == "" {
		http.Error(w, "missing user identity", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	log.Printf("📦 Creating order for user %s with %d items", userID, len(req.Items))

	// Parse schedule (if provided)
	var scheduledFor *time.Time
	scheduleType := ""
	orderStatus := "pending"
	if req.Schedule != nil {
		scheduleType = strings.TrimSpace(req.Schedule.Type)
		dateStr := strings.TrimSpace(req.Schedule.Date)
		timeStr := strings.TrimSpace(req.Schedule.Time)
		if dateStr != "" && timeStr != "" {
			if t, err := time.Parse("2006-01-02T15:04:05", dateStr+"T"+timeStr+":00"); err == nil {
				scheduledFor = &t
				if t.After(time.Now()) {
					orderStatus = "scheduled"
				}
			}
		}
	}

	// Calculate total and item count
	var total float64
	var totalItems int
	for _, item := range req.Items {
		total += item.Price * float64(item.Qty)
		totalItems += item.Qty
	}

	// Combine customer info into customer_name field
	customerInfo := req.CustomerName
	if req.CustomerPhone != "" {
		customerInfo += " (" + req.CustomerPhone + ")"
	}

	// Insert order into database
	var orderID int

	insertWithSchedule := `
		INSERT INTO orders (customer_name, delivery_type, address, status, total_items, subtotal, delivery_fee, total_amount, sender_id, scheduled_for, schedule_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $6, $7, $8, $9, NOW())
		RETURNING id
	`

	insertLegacy := `
		INSERT INTO orders (customer_name, delivery_type, address, status, total_items, subtotal, delivery_fee, total_amount, sender_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $6, $7, NOW())
		RETURNING id
	`

	err := configs.DB.QueryRow(insertWithSchedule, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, total, userID, scheduledFor, scheduleType).Scan(&orderID)
	if err != nil {
		// Backwards compatible fallback if DB columns aren't migrated yet.
		msg := err.Error()
		if strings.Contains(msg, "scheduled_for") || strings.Contains(msg, "schedule_type") {
			err = configs.DB.QueryRow(insertLegacy, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, total, userID).Scan(&orderID)
		}
	}

	if err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	// Insert order items
	for _, item := range req.Items {
		_, err = configs.DB.Exec(`
			INSERT INTO order_items (order_id, product, quantity, price, created_at)
			VALUES ($1, $2, $3, $4, NOW())
		`, orderID, item.Name, item.Qty, item.Price)

		if err != nil {
			log.Printf("⚠️  Failed to insert item %s: %v", item.Name, err)
		}
	}

	log.Printf("✅ Order #%d created successfully", orderID)

	// Send response
	resp := ChatOrderResponse{
		Success: true,
		OrderID: orderID,
		Message: "Order placed successfully!",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// Send confirmation message to user via Messenger (async)
	go func() {
		defer func() { _ = recover() }()

		itemsList := ""
		for i, item := range req.Items {
			if i >= 3 {
				break
			}
			itemsList += fmt.Sprintf("%s × %d\n", item.Name, item.Qty)
		}
		if len(req.Items) > 3 {
			itemsList += "...and more\n"
		}

		schedLine := ""
		if scheduledFor != nil {
			schedLine = fmt.Sprintf("\nScheduled: %s %s\n", req.Schedule.Date, req.Schedule.Time)
		}
		statusLine := "Status: ⏳ Pending\n\n"
		if orderStatus == "scheduled" {
			statusLine = "Status: 🗓 Scheduled\n\n"
		}

		msg := "🎉 Order Confirmed!\n\n" +
			"Order #" + strconv.Itoa(orderID) + "\n" +
			itemsList +
			fmt.Sprintf("\nTotal: $%.2f\n", total) +
			schedLine +
			statusLine +
			"We'll start preparing your order soon!"

		SendMessage(userID, msg)
	}()
}
