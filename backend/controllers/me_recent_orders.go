package controllers

import (
	"bakeflow/models"
	"encoding/json"
	"net/http"
	"strconv"
)

func MeRecentOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	psid, ok := requirePSIDFromToken(w, r)
	if !ok {
		return
	}

	limit := 5
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	orders, err := models.GetRecentOrdersBySenderID(psid, limit)
	if err != nil {
		http.Error(w, "failed to load recent orders", http.StatusInternalServerError)
		return
	}
	for i := range orders {
		orders[i].SenderID = ""
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"orders": orders})
}

func MeActiveOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	psid, ok := requirePSIDFromToken(w, r)
	if !ok {
		return
	}

	// SIMPLIFIED: Return no active order for blocking purposes
	// The ordering logic is now handled entirely in /api/chat/orders endpoint
	// This endpoint returning nil means:
	// - Regular orders no longer block new orders (users can have unlimited simultaneous regular orders)
	// - Choice dialog suggestions are shown in /api/chat/orders instead
	// - Custom cake order blocking is still enforced at /api/chat/orders level

	statuses := []string{"pending", "confirmed", "preparing", "ready", "delivering", "scheduled"}
	allOrders, err := models.GetAllOrdersBySenderIDAndStatuses(psid, statuses)
	if err != nil {
		http.Error(w, "failed to load active order", http.StatusInternalServerError)
		return
	}

	// For now, return nil - no "active order" blocking
	// All order management is done at /api/chat/orders endpoint level
	var order *models.Order
	if len(allOrders) == 0 {
		order = nil
	} else {
		// If we had custom order checking logic here, it would go here
		// But for now, we just return nil to not block anything
		order = nil
	}

	editable := false

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order": order, "editable": editable})
}
