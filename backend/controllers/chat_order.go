package controllers

import (
	"bakeflow/configs"
	"bakeflow/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

type OrderChoiceRequest struct {
	UserID       string          `json:"user_id"`
	Choice       string          `json:"choice"`   // "add_to_existing" or "new_order"
	OrderID      int             `json:"order_id"` // for add_to_existing choice
	Items        []ChatOrderItem `json:"items"`
	CustomerName string          `json:"customer_name"`
	DeliveryType string          `json:"delivery_type"`
	Address      string          `json:"address"`
}

// HandleOrderChoice processes the user's decision to add to existing or create new order
func HandleOrderChoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify webview token
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

	var req OrderChoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Use the verified user ID from token, not from request body
	userID := psid

	choice := strings.ToLower(strings.TrimSpace(req.Choice))

	if choice == "add_to_existing" {
		// Add items to existing order with smart merging
		addItemsToExistingOrder(w, userID, req.OrderID, req.Items)
	} else if choice == "new_order" {
		// Create a new order (treat as a fresh order)
		createNewOrderAfterChoice(w, userID, req)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "invalid_choice",
			"message": "Choice must be 'add_to_existing' or 'new_order'",
		})
	}
}

// addItemsToExistingOrder adds items to existing order, merging duplicate products
func addItemsToExistingOrder(w http.ResponseWriter, userID string, orderID int, newItems []ChatOrderItem) {
	// Get the existing order
	existingOrder, err := models.GetOrderByID(orderID)
	if err != nil || existingOrder == nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	var totalNewItems int
	var totalNewSubtotal float64

	// Get existing items
	existingItems, err := models.GetOrderItems(orderID)
	if err != nil {
		http.Error(w, "failed to fetch order items", http.StatusInternalServerError)
		return
	}

	// Create a map of product names to existing item IDs for easy lookup
	productItemMap := make(map[string]int)     // product_name -> item_id
	productQuantityMap := make(map[string]int) // product_name -> current quantity

	for _, item := range existingItems {
		productItemMap[item.Product] = item.ID
		productQuantityMap[item.Product] = item.Quantity
	}

	// Process new items - merge or insert
	for _, newItem := range newItems {
		if existingItemID, exists := productItemMap[newItem.Name]; exists {
			// Product already exists in order - increase its quantity
			currentQty := productQuantityMap[newItem.Name]
			newQty := currentQty + newItem.Qty

			_, err = configs.DB.Exec(`
				UPDATE order_items
				SET quantity = $1
				WHERE id = $2
			`, newQty, existingItemID)

			if err != nil {
				log.Printf("⚠️  Failed to update item %s: %v", newItem.Name, err)
			} else {
				log.Printf("✅ Updated quantity for '%s': %d → %d", newItem.Name, currentQty, newQty)
			}

			totalNewItems += newItem.Qty
			totalNewSubtotal += newItem.Price * float64(newItem.Qty)
		} else {
			// New product - insert as new item
			_, err = configs.DB.Exec(`
				INSERT INTO order_items (order_id, product, quantity, price, note, image_url, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
			`, orderID, newItem.Name, newItem.Qty, newItem.Price, newItem.Note, newItem.ImageURL)

			if err != nil {
				log.Printf("⚠️  Failed to insert item %s: %v", newItem.Name, err)
			} else {
				log.Printf("✅ Inserted new item '%s' (qty: %d)", newItem.Name, newItem.Qty)
			}

			totalNewItems += newItem.Qty
			totalNewSubtotal += newItem.Price * float64(newItem.Qty)

			// Handle stock reservation if product has ID
			if newItem.ProductID > 0 {
				if err := models.AtomicPurchase(newItem.ProductID, newItem.Qty, orderID); err != nil {
					log.Printf("⚠️ Atomic purchase failed for product %d: %v", newItem.ProductID, err)
				}
			}
		}
	}

	// Update order totals
	_, _ = configs.DB.Exec(`
		UPDATE orders
		SET total_items = total_items + $1,
		    subtotal = subtotal + $2,
		    total_amount = total_amount + $3
		WHERE id = $4
	`, totalNewItems, totalNewSubtotal, totalNewSubtotal, orderID)

	_ = models.UpsertCustomerPhone(userID, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orderID": orderID,
		"message": fmt.Sprintf("✅ Added %d items to your order #BF-%d", totalNewItems, orderID),
		"action":  "items_merged",
	})

	// Send Messenger notification
	go func() {
		defer func() { _ = recover() }()
		if !isValidMessengerRecipientID(userID) {
			return
		}

		// Fetch updated order for notification
		updatedOrder, _ := models.GetOrderByID(orderID)
		if updatedOrder == nil {
			return
		}

		itemSummary := "Items added"
		if len(newItems) > 0 {
			first := newItems[0]
			itemSummary = fmt.Sprintf("%s × %d", first.Name, first.Qty)
			if len(newItems) > 1 {
				itemSummary = fmt.Sprintf("%s + %d more", itemSummary, len(newItems)-1)
			}
		}

		productImage := "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=400&h=200&fit=crop"
		if len(newItems) > 0 && strings.TrimSpace(newItems[0].ImageURL) != "" {
			imgURL := strings.TrimSpace(newItems[0].ImageURL)
			if strings.HasPrefix(imgURL, "https://") {
				productImage = imgURL
			}
		}

		title := "Items Added & Merged"
		subtitle := fmt.Sprintf("Order #BF-%d • %s • Total: $%.2f", updatedOrder.ID, itemSummary, updatedOrder.TotalAmount)
		buttons := []Button{
			{Type: "postback", Title: "Track Order", Payload: fmt.Sprintf("TRACK_ORDER_%d", updatedOrder.ID)},
			{Type: "postback", Title: "View Cart", Payload: fmt.Sprintf("VIEW_CART_%d", updatedOrder.ID)},
		}
		_ = SendOrderCardWithTag(userID, updatedOrder.ID, title, subtitle, productImage, buttons, "POST_PURCHASE_UPDATE")
	}()
}

// createNewOrderAfterChoice creates a completely new order
func createNewOrderAfterChoice(w http.ResponseWriter, userID string, req OrderChoiceRequest) {
	var orderID int

	// Calculate totals
	var subtotal float64
	totalItems := 0
	for _, item := range req.Items {
		subtotal += item.Price * float64(item.Qty)
		totalItems += item.Qty
	}

	customerInfo := strings.TrimSpace(req.CustomerName)
	if customerInfo == "" {
		customerInfo = "Customer"
	}

	status := "pending"

	// Insert new order
	insertQuery := `
		INSERT INTO orders (customer_name, delivery_type, address, status, total_items, subtotal, delivery_fee, total_amount, sender_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $6, $7, NOW())
		RETURNING id
	`

	err := configs.DB.QueryRow(insertQuery, customerInfo, req.DeliveryType, req.Address, status, totalItems, subtotal, userID).Scan(&orderID)
	if err != nil {
		log.Printf("❌ Failed to create new order: %v", err)
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	log.Printf("📦 New order #%d created", orderID)

	// Insert order items
	for _, item := range req.Items {
		_, err = configs.DB.Exec(`
			INSERT INTO order_items (order_id, product, quantity, price, note, image_url, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
		`, orderID, item.Name, item.Qty, item.Price, item.Note, item.ImageURL)

		if err != nil {
			log.Printf("⚠️  Failed to insert item %s: %v", item.Name, err)
		}

		// Handle stock reservation
		if item.ProductID > 0 {
			if err := models.AtomicPurchase(item.ProductID, item.Qty, orderID); err != nil {
				log.Printf("⚠️ Atomic purchase failed for product %d: %v", item.ProductID, err)
			}
		}
	}

	_ = models.UpsertCustomerPhone(userID, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"orderID": orderID,
		"message": fmt.Sprintf("✅ New order #BF-%d created with %d items", orderID, totalItems),
		"action":  "new_order_created",
	})

	// Send Messenger notification
	go func() {
		defer func() { _ = recover() }()
		if !isValidMessengerRecipientID(userID) {
			return
		}

		itemSummary := "Order placed"
		if len(req.Items) > 0 {
			first := req.Items[0]
			itemSummary = fmt.Sprintf("%s × %d", first.Name, first.Qty)
			if len(req.Items) > 1 {
				itemSummary = fmt.Sprintf("%s + %d more", itemSummary, len(req.Items)-1)
			}
		}

		productImage := "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=400&h=200&fit=crop"
		if len(req.Items) > 0 && strings.TrimSpace(req.Items[0].ImageURL) != "" {
			imgURL := strings.TrimSpace(req.Items[0].ImageURL)
			if strings.HasPrefix(imgURL, "https://") {
				productImage = imgURL
			}
		}

		title := "New Order Placed"
		subtitle := fmt.Sprintf("Order #BF-%d • %s • Total: $%.2f", orderID, itemSummary, subtotal)
		buttons := []Button{
			{Type: "postback", Title: "Track Order", Payload: fmt.Sprintf("TRACK_ORDER_%d", orderID)},
			{Type: "postback", Title: "Need Help?", Payload: "CONTACT_SUPPORT"},
		}
		_ = SendOrderCardWithTag(userID, orderID, title, subtitle, productImage, buttons, "POST_PURCHASE_UPDATE")

		// Send Payment Link Card
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000" // Default fallback
		}

		paymentLink := fmt.Sprintf("%s/order/%d", frontendURL, orderID)

		paymentElement := Element{
			Title:    fmt.Sprintf("💳 Pay for Order #%d", orderID),
			Subtitle: fmt.Sprintf("Total: $%.2f. Tap to pay securely.", subtotal),                      // subtotal used in this func, logic matches CreateChatOrder
			ImageURL: "https://images.unsplash.com/photo-1556742049-0cfed4f7a07d?w=300&h=200&fit=crop", // Payment image
			Buttons: []Button{
				{Type: "web_url", Title: "Pay Now", URL: paymentLink},
			},
		}

		SendGenericTemplate(userID, []Element{paymentElement})

	}()
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
	existingOrders, _ := models.GetAllOrdersBySenderIDAndStatuses(userID, activeStatuses)
	log.Printf("🔍 [OrderChoice] Checking for existing orders. UserID: %s, Found: %d orders", userID, len(existingOrders))

	// Filter out delivered/cancelled orders and check if any are available for adding items
	var availableOrders []*models.Order
	if len(existingOrders) > 0 {
		for _, order := range existingOrders {
			status := strings.ToLower(strings.TrimSpace(order.Status))
			allowAdd := status != "delivered" && status != "cancelled"
			if allowAdd {
				availableOrders = append(availableOrders, order)
				log.Printf("🔍 [OrderChoice] Available order found: #%d, Status: %s, Items: %d", order.ID, order.Status, order.TotalItems)
			} else {
				log.Printf("🔍 [OrderChoice] Skipping order #%d with status: %s (not available for adding)", order.ID, order.Status)
			}
		}
	}

	log.Printf("🔍 [OrderChoice] Total available orders after filtering: %d", len(availableOrders))

	if len(availableOrders) > 0 { // Build summary of available orders
		type OrderSummary struct {
			ID      int     `json:"id"`
			Status  string  `json:"status"`
			Items   int     `json:"items"`
			Amount  float64 `json:"amount"`
			Summary string  `json:"summary"`
		}

		var orderSummaries []OrderSummary
		for _, order := range availableOrders {
			summary := OrderSummary{
				ID:      order.ID,
				Status:  order.Status,
				Items:   order.TotalItems,
				Amount:  order.TotalAmount,
				Summary: fmt.Sprintf("Order #BF-%d (%s) • %d items • $%.2f", order.ID, order.Status, order.TotalItems, order.TotalAmount),
			}
			orderSummaries = append(orderSummaries, summary)
		}

		// Return choice with all available orders
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"action":   "ask_user_choice",
			"message":  "You have active orders. Which one would you like to add items to?",
			"orders":   orderSummaries,
			"newItems": req.Items,
		})
		return
	}

	// Check for orders by phone if no user orders found
	if len(existingOrders) == 0 && req.CustomerPhone != "" {
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
		log.Printf("🔔 [Order] Attempting to send notification to '%s' for order #%d", userID, orderID)

		if !isValidMessengerRecipientID(userID) {
			log.Printf("❌ [Order] Skipping Messenger notification: invalid recipient id '%s'", userID)
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

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:3000"
		}

		buttons := []Button{
			{Type: "web_url", Title: "Pay Now", URL: fmt.Sprintf("%s/order/%d", frontendURL, orderID)},
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
	// Allow any non-empty ID for now to prevent silent failures during testing
	return true
}
