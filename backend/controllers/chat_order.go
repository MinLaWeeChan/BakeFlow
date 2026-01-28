package controllers

import (
	"bakeflow/configs"
	"bakeflow/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type ChatOrderSchedule struct {
	Type string `json:"type"`
	Date string `json:"date"`
	Time string `json:"time"`
}

type ChatOrderRequest struct {
	UserID           string             `json:"user_id"`
	Items            []ChatOrderItem    `json:"items"`
	Channel          string             `json:"channel"`
	Notes            string             `json:"notes"`
	CustomerName     string             `json:"customer_name"`
	CustomerPhone    string             `json:"customer_phone"`
	DeliveryType     string             `json:"delivery_type"`
	Address          string             `json:"address"`
	Schedule         *ChatOrderSchedule `json:"schedule"`
	AppliedPromotion *AppliedPromotion  `json:"appliedPromotion,omitempty"`
	Discount         float64            `json:"discount"`
}

type ChatOrderItem struct {
	ProductID int     `json:"product_id"`
	Name      string  `json:"name"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
	Note      string  `json:"note"`
	ImageURL  string  `json:"image_url"`
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

	tok := strings.TrimSpace(r.URL.Query().Get("t"))
	if tok == "" {
		http.Error(w, "missing auth token", http.StatusUnauthorized)
		return
	}
	psid, errTok := VerifyWebviewToken(tok)
	if errTok != nil {
		http.Error(w, "invalid auth token", http.StatusUnauthorized)
		return
	}
	userID := psid

	if len(req.Items) == 0 {
		http.Error(w, "cart is empty", http.StatusBadRequest)
		return
	}

	normalizedPhone, ok := NormalizeMyanmarPhoneE164(req.CustomerPhone)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "invalid_phone",
			"message": "Enter 09xxxxxxxxx or +959xxxxxxxxx",
		})
		return
	}
	req.CustomerPhone = normalizedPhone

	blocked, err := models.IsIdentityBlocked("psid", userID)
	if err != nil {
		http.Error(w, "failed to verify customer", http.StatusInternalServerError)
		return
	}
	if blocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "blocked",
			"message": "This account is currently restricted from placing new orders. Please contact the bakery if you believe this is a mistake.",
		})
		return
	}

	blocked, err = models.IsIdentityBlocked("phone", req.CustomerPhone)
	if err != nil {
		http.Error(w, "failed to verify customer", http.StatusInternalServerError)
		return
	}
	if blocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "blocked",
			"message": "This phone number is currently restricted from placing new orders. Please contact the bakery if you believe this is a mistake.",
		})
		return
	}

	// PRODUCTION-GRADE: Validate stock availability using atomic check
	// This checks AVAILABLE stock (total - reserved) to prevent overselling
	for _, item := range req.Items {
		if item.ProductID > 0 {
			stockStatus, err := models.GetProductStockStatus(item.ProductID)
			if err != nil {
				log.Printf("⚠️ Product %d not found or inactive", item.ProductID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error":   "product_unavailable",
					"message": fmt.Sprintf("%s is no longer available", item.Name),
				})
				return
			}

			// Check AVAILABLE stock (accounts for reservations)
			if stockStatus.AvailableStock < item.Qty {
				log.Printf("⚠️ Insufficient stock for %s: requested %d, available %d (reserved: %d)",
					stockStatus.ProductName, item.Qty, stockStatus.AvailableStock, stockStatus.ReservedStock)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success":   false,
					"error":     "insufficient_stock",
					"message":   fmt.Sprintf("Sorry, only %d %s available", stockStatus.AvailableStock, stockStatus.ProductName),
					"product":   stockStatus.ProductName,
					"requested": item.Qty,
					"available": stockStatus.AvailableStock,
				})
				return
			}
		}
	}

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

	// Calculate subtotal from items
	subtotal := 0.0
	var totalItems int
	for _, item := range req.Items {
		subtotal += item.Price * float64(item.Qty)
		totalItems += item.Qty
	}

	// Apply discount if promotion was applied
	discount := req.Discount
	total := subtotal - discount
	if total < 0 {
		total = 0
	}

	// Extract promotion ID from applied promotion
	var promotionID *int
	if req.AppliedPromotion != nil {
		promotionID = &req.AppliedPromotion.ID
	}

	// Combine customer info into customer_name field
	customerInfo := req.CustomerName
	if req.CustomerPhone != "" {
		customerInfo += " (" + req.CustomerPhone + ")"
	}

	activeStatuses := []string{"pending", "confirmed", "preparing", "ready", "delivering", "scheduled"}
	existingOrder, _ := models.GetLatestOrderBySenderIDAndStatuses(userID, activeStatuses)
	if existingOrder == nil && req.CustomerPhone != "" {
		phoneOrder, _ := models.GetLatestOrderByPhoneAndStatuses(req.CustomerPhone, activeStatuses)
		if phoneOrder != nil {
			status := strings.ToLower(strings.TrimSpace(phoneOrder.Status))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "order_locked",
				"message": "We found an active order for this phone number that can’t be modified from this session. Please continue from the original Messenger chat or wait until it completes.",
				"status":  status,
				"order":   phoneOrder.ID,
			})
			return
		}
	}
	if existingOrder != nil {
		status := strings.ToLower(strings.TrimSpace(existingOrder.Status))
		allowAdd := status == "pending" || status == "confirmed"
		if !allowAdd {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "order_locked",
				"message": "Your order is already being prepared and can’t be modified. Please place a new order for additional items.",
				"status":  status,
				"order":   existingOrder.ID,
			})
			return
		}
		{
			newCustomerName := strings.TrimSpace(req.CustomerName)
			newCustomerInfo := newCustomerName
			if newCustomerInfo == "" {
				newCustomerInfo = existingOrder.CustomerName
			} else if req.CustomerPhone != "" {
				newCustomerInfo += " (" + req.CustomerPhone + ")"
			}

			newDeliveryType := strings.ToLower(strings.TrimSpace(req.DeliveryType))
			if newDeliveryType == "" {
				newDeliveryType = existingOrder.DeliveryType
			}

			newAddress := strings.TrimSpace(req.Address)
			if newDeliveryType != "delivery" {
				newAddress = "Pickup at store"
			}
			if newAddress == "" {
				newAddress = existingOrder.Address
			}

			_, _ = configs.DB.Exec(`
				UPDATE orders
				SET customer_name = $1,
				    delivery_type = $2,
				    address = $3
				WHERE id = $4
			`, newCustomerInfo, newDeliveryType, newAddress, existingOrder.ID)
		}

		var addedItems int
		for _, item := range req.Items {
			addedItems += item.Qty
		}
		addedSubtotal := subtotal
		addedTotal := subtotal
		for _, item := range req.Items {
			_, err = configs.DB.Exec(`
				INSERT INTO order_items (order_id, product, quantity, price, note, image_url, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
			`, existingOrder.ID, item.Name, item.Qty, item.Price, item.Note, item.ImageURL)
			if err != nil {
				log.Printf("⚠️  Failed to insert item %s: %v", item.Name, err)
			}
			if item.ProductID > 0 {
				if err := models.AtomicPurchase(item.ProductID, item.Qty, existingOrder.ID); err != nil {
					log.Printf("⚠️ Atomic purchase failed for product %d: %v", item.ProductID, err)
				}
			}
		}
		_, _ = configs.DB.Exec(`
			UPDATE orders
			SET total_items = total_items + $1,
			    subtotal = subtotal + $2,
			    total_amount = total_amount + $3
			WHERE id = $4
		`, addedItems, addedSubtotal, addedTotal, existingOrder.ID)
		_ = models.UpsertCustomerPhone(userID, req.CustomerPhone)

		resp := ChatOrderResponse{
			Success: true,
			OrderID: existingOrder.ID,
			Message: "Items added to your existing order",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		go func() {
			defer func() { _ = recover() }()
			if !isValidMessengerRecipientID(userID) {
				return
			}
			productImage := "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=400&h=200&fit=crop"
			if len(req.Items) > 0 && strings.TrimSpace(req.Items[0].ImageURL) != "" {
				imgURL := strings.TrimSpace(req.Items[0].ImageURL)
				if strings.HasPrefix(imgURL, "https://") {
					productImage = imgURL
				}
			}
			itemSummary := "Items added"
			if len(req.Items) > 0 {
				first := req.Items[0]
				itemSummary = fmt.Sprintf("%s × %d", first.Name, first.Qty)
				if len(req.Items) > 1 {
					itemSummary = fmt.Sprintf("%s + %d more", itemSummary, len(req.Items)-1)
				}
			}
			newTotal := existingOrder.TotalAmount + addedTotal
			totalLabel := fmt.Sprintf("New total: $%.2f", newTotal)
			subtitle := fmt.Sprintf("Order #BF-%d • %s • %s", existingOrder.ID, itemSummary, totalLabel)
			title := "Items Added"
			buttons := []Button{
				{Type: "postback", Title: "Track Order", Payload: fmt.Sprintf("TRACK_ORDER_%d", existingOrder.ID)},
				{Type: "postback", Title: "Need Help?", Payload: "CONTACT_SUPPORT"},
			}
			_ = SendOrderCardWithTag(userID, existingOrder.ID, title, subtitle, productImage, buttons, "POST_PURCHASE_UPDATE")
		}()
		return
	}

	limited, retryAfter, err := checkOrderRateLimit(userID)
	if err != nil {
		http.Error(w, "failed to verify customer", http.StatusInternalServerError)
		return
	}
	if limited {
		retryMinutes := int(retryAfter.Minutes())
		if retryMinutes < 1 {
			retryMinutes = 1
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":             false,
			"error":               "rate_limited",
			"message":             fmt.Sprintf("Too many recent orders. Please wait %d minutes before ordering again.", retryMinutes),
			"retry_after_minutes": retryMinutes,
		})
		return
	}

	log.Printf("📦 Creating order for user %s with %d items", userID, len(req.Items))

	var orderID int

	// New insert with promotion fields
	insertWithPromotion := `
		INSERT INTO orders (customer_name, delivery_type, address, status, total_items, subtotal, delivery_fee, total_amount, sender_id, promotion_id, discount, scheduled_for, schedule_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8, $9, $10, $11, $12, NOW())
		RETURNING id
	`

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

	// Try insert with promotion columns first
	err = configs.DB.QueryRow(insertWithPromotion, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, subtotal, total, userID, promotionID, discount, scheduledFor, scheduleType).Scan(&orderID)
	if err != nil {
		// Backwards compatible fallback if DB columns aren't migrated yet.
		msg := err.Error()
		if strings.Contains(msg, "promotion_id") || strings.Contains(msg, "discount") {
			// Try with scheduling columns
			err = configs.DB.QueryRow(insertWithSchedule, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, subtotal, userID, scheduledFor, scheduleType).Scan(&orderID)
			if err != nil {
				// Check for scheduling columns not existing
				if strings.Contains(err.Error(), "scheduled_for") || strings.Contains(err.Error(), "schedule_type") {
					// If the user attempted to schedule, don't silently drop scheduling.
					if scheduledFor != nil {
						http.Error(w, "Scheduling is not enabled on the database yet. Please apply migration 006_add_order_scheduling.sql", http.StatusBadRequest)
						return
					}
					err = configs.DB.QueryRow(insertLegacy, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, subtotal, userID).Scan(&orderID)
				}
			}
		} else if strings.Contains(msg, "scheduled_for") || strings.Contains(msg, "schedule_type") {
			// If the user attempted to schedule, don't silently drop scheduling.
			if scheduledFor != nil {
				http.Error(w, "Scheduling is not enabled on the database yet. Please apply migration 006_add_order_scheduling.sql", http.StatusBadRequest)
				return
			}
			err = configs.DB.QueryRow(insertLegacy, customerInfo, req.DeliveryType, req.Address, orderStatus, totalItems, subtotal, userID).Scan(&orderID)
		}
	}

	if err != nil {
		log.Printf("❌ Failed to create order: %v", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	if err := models.UpsertCustomerPhone(userID, req.CustomerPhone); err != nil {
		log.Printf("⚠️ Failed to link phone for %s: %v", userID, err)
	}

	// Insert order items and deduct stock using ATOMIC operations
	// This prevents race conditions where two users order the last item simultaneously
	for _, item := range req.Items {
		_, err = configs.DB.Exec(`
			INSERT INTO order_items (order_id, product, quantity, price, note, image_url, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, orderID, item.Name, item.Qty, item.Price, item.Note, item.ImageURL)

		if err != nil {
			log.Printf("⚠️  Failed to insert item %s: %v", item.Name, err)
		}

		// PRODUCTION-GRADE: Use atomic purchase with row locking
		if item.ProductID > 0 {
			err = models.AtomicPurchase(item.ProductID, item.Qty, orderID)
			if err != nil {
				// Stock became unavailable between check and purchase (rare race condition)
				// In production, you might want to handle partial fulfillment here
				log.Printf("⚠️ Atomic purchase failed for product %d: %v", item.ProductID, err)
				// Note: Order is already created, so we log the issue
				// A more robust system would use a transaction for the entire order
			} else {
				log.Printf("📦 Atomically deducted %d stock from product #%d", item.Qty, item.ProductID)
			}
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
		if !isValidMessengerRecipientID(userID) {
			log.Printf("[Order] Skipping Messenger notification: invalid recipient id '%s'", userID)
			return
		}

		// Build item summary (max 3 items shown)
		itemsList := ""
		for i, item := range req.Items {
			if i >= 3 {
				itemsList += fmt.Sprintf("...and %d more item(s)", len(req.Items)-3)
				break
			}
			itemsList += fmt.Sprintf("• %s × %d\n", item.Name, item.Qty)
		}

		// Get product image from first item (must be valid https URL)
		productImage := "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=400&h=200&fit=crop"
		if len(req.Items) > 0 && req.Items[0].ImageURL != "" {
			imgURL := strings.TrimSpace(req.Items[0].ImageURL)
			// Only use if it's a valid https URL
			if strings.HasPrefix(imgURL, "https://") {
				productImage = imgURL
			}
		}

		itemSummary := "Items"
		if len(req.Items) > 0 {
			first := req.Items[0]
			itemSummary = fmt.Sprintf("%s × %d", first.Name, first.Qty)
			if len(req.Items) > 1 {
				itemSummary = fmt.Sprintf("%s + %d more", itemSummary, len(req.Items)-1)
			}
		}

		totalLabel := fmt.Sprintf("Total: $%.2f", total)
		if discount > 0 {
			totalLabel = fmt.Sprintf("Total: $%.2f (saved $%.2f)", total, discount)
		}

		subtitle := fmt.Sprintf("Order #BF-%d • %s • %s", orderID, itemSummary, totalLabel)
		if scheduledFor != nil {
			subtitle = fmt.Sprintf("%s • Scheduled %s %s", subtitle, req.Schedule.Date, req.Schedule.Time)
		}

		title := "Order Confirmed"

		// Buttons for the card - include order ID so Track Order shows this specific order
		trackPayload := fmt.Sprintf("TRACK_ORDER_%d", orderID)
		log.Printf("[Order] Creating card with Track payload: %s", trackPayload)
		buttons := []Button{
			{Type: "postback", Title: "Track Order", Payload: trackPayload},
			{Type: "postback", Title: "Need Help?", Payload: "CONTACT_SUPPORT"},
		}

		// Send modern card with POST_PURCHASE_UPDATE tag
		err := SendOrderCardWithTag(userID, orderID, title, subtitle, productImage, buttons, "POST_PURCHASE_UPDATE")
		if err != nil {
			log.Printf("[Order] Card failed for order #%d: %v", orderID, err)
			// Fallback to plain text
			fallbackMsg := fmt.Sprintf("✅ Order #BF-%d confirmed!\n\n%s\nTotal: $%.2f\n\nWe'll notify you when it's ready!", orderID, itemsList, total)
			SendMessageWithTag(userID, fallbackMsg, "POST_PURCHASE_UPDATE")
		}
	}()
}

func isValidMessengerRecipientID(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
