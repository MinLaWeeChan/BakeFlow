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
