package controllers

import (
	"bakeflow/models"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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

	// Only return non-custom (regular) active orders so custom cake orders
	// don't block the regular cart flow
	statuses := []string{"pending", "confirmed", "preparing", "ready", "delivering", "scheduled"}
	allOrders, err := models.GetAllOrdersBySenderIDAndStatuses(psid, statuses)
	if err != nil {
		http.Error(w, "failed to load active order", http.StatusInternalServerError)
		return
	}

	// Find the latest non-custom, non-scheduled order
	// Scheduled orders are independent — they shouldn't block "order now"
	var order *models.Order
	for _, o := range allOrders {
		if isCustomOrder(o.ID) {
			continue
		}
		if strings.ToLower(strings.TrimSpace(o.Status)) == "scheduled" {
			continue
		}
		order = o
		break
	}

	editable := false
	if order != nil {
		switch order.Status {
		case "pending", "confirmed", "preparing":
			editable = true
		}
		order.SenderID = ""
		order.Items = nil
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"order": order, "editable": editable})
}
