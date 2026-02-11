package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bakeflow/configs"
	"bakeflow/models"

	"github.com/gorilla/mux"
)

func GetPreorderSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	settings, err := models.GetPreorderSettings()
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder settings"}`, http.StatusInternalServerError)
		return
	}
	var updatedAt interface{}
	if settings != nil {
		updatedAt = settings.UpdatedAt
	}

	if settings == nil || !settings.Enabled {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled":     false,
			"product_ids": []int{},
			"products":    []models.Product{},
			"updated_at":  updatedAt,
		})
		return
	}

	products, err := models.GetProductsByIDs(settings.ProductIDs, true)
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder products"}`, http.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []models.Product{}
	}
	filtered := make([]models.Product, 0, len(products))
	filteredIDs := make([]int, 0, len(products))
	for _, product := range products {
		if strings.ToLower(strings.TrimSpace(product.Category)) != "cakes" {
			continue
		}
		if strings.ToLower(strings.TrimSpace(product.Status)) != "active" {
			continue
		}
		filtered = append(filtered, product)
		filteredIDs = append(filteredIDs, product.ID)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":     settings.Enabled,
		"product_ids": filteredIDs,
		"products":    filtered,
		"updated_at":  updatedAt,
	})
}

func AdminGetPreorderSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	settings, err := models.GetPreorderSettings()
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder settings"}`, http.StatusInternalServerError)
		return
	}

	if settings == nil {
		settings = &models.PreorderSettings{Enabled: true, ProductIDs: []int{}}
	}
	updatedAt := settings.UpdatedAt

	products, err := models.GetProductsByIDs(settings.ProductIDs, false)
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder products"}`, http.StatusInternalServerError)
		return
	}
	if products == nil {
		products = []models.Product{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":     settings.Enabled,
		"product_ids": settings.ProductIDs,
		"products":    products,
		"updated_at":  updatedAt,
	})
}

func AdminUpdatePreorderSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Enabled    *bool `json:"enabled"`
		ProductIDs []int `json:"product_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Enabled == nil {
		http.Error(w, `{"error":"Missing enabled flag"}`, http.StatusBadRequest)
		return
	}

	normalized := normalizeUniqueIDs(req.ProductIDs)
	allowedIDs := normalized
	if len(normalized) > 0 {
		products, err := models.GetProductsByIDs(normalized, false)
		if err != nil {
			http.Error(w, `{"error":"Failed to load preorder products"}`, http.StatusInternalServerError)
			return
		}
		filtered := make([]int, 0, len(products))
		for _, product := range products {
			if strings.ToLower(strings.TrimSpace(product.Category)) != "cakes" {
				continue
			}
			if strings.ToLower(strings.TrimSpace(product.Status)) != "active" {
				continue
			}
			filtered = append(filtered, product.ID)
		}
		allowedIDs = filtered
	}
	settings, err := models.UpdatePreorderSettings(*req.Enabled, allowedIDs)
	if err != nil {
		http.Error(w, `{"error":"Failed to save preorder settings"}`, http.StatusInternalServerError)
		return
	}

	products, err := models.GetProductsByIDs(settings.ProductIDs, false)
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder products"}`, http.StatusInternalServerError)
		return
	}
	if products == nil {
		products = []models.Product{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"enabled":     settings.Enabled,
		"product_ids": settings.ProductIDs,
		"products":    products,
		"updated_at":  settings.UpdatedAt,
	})
}

func GetPreorderProductSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(idStr)
	if err != nil || productID <= 0 {
		http.Error(w, `{"error":"Invalid product id"}`, http.StatusBadRequest)
		return
	}

	settings, err := models.GetPreorderProductSettings(productID)
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder settings"}`, http.StatusInternalServerError)
		return
	}

	if settings == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"product_id":   productID,
			"enabled":      false,
			"start_date":   "",
			"end_date":     "",
			"sizes":        []string{},
			"layers":       []string{},
			"creams":       []string{},
			"flavors":      []string{},
			"size_prices":  map[string]float64{},
			"layer_prices": map[string]float64{},
			"cream_prices": map[string]float64{},
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id":   settings.ProductID,
		"enabled":      settings.Enabled,
		"start_date":   formatDate(settings.StartDate),
		"end_date":     formatDate(settings.EndDate),
		"sizes":        settings.Sizes,
		"layers":       settings.Layers,
		"creams":       settings.Creams,
		"flavors":      settings.Flavors,
		"size_prices":  settings.SizePrices,
		"layer_prices": settings.LayerPrices,
		"cream_prices": settings.CreamPrices,
		"updated_at":   settings.UpdatedAt,
	})
}

func AdminGetPreorderProductSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(idStr)
	if err != nil || productID <= 0 {
		http.Error(w, `{"error":"Invalid product id"}`, http.StatusBadRequest)
		return
	}

	product, err := models.GetProductByID(configs.DB, productID)
	if err != nil {
		http.Error(w, `{"error":"Failed to load product"}`, http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, `{"error":"Product not found"}`, http.StatusNotFound)
		return
	}
	if strings.ToLower(strings.TrimSpace(product.Category)) != "cakes" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"product_id":   productID,
			"enabled":      false,
			"start_date":   "",
			"end_date":     "",
			"sizes":        []string{},
			"layers":       []string{},
			"creams":       []string{},
			"flavors":      []string{},
			"size_prices":  map[string]float64{},
			"layer_prices": map[string]float64{},
			"cream_prices": map[string]float64{},
		})
		return
	}

	settings, err := models.GetPreorderProductSettings(productID)
	if err != nil {
		http.Error(w, `{"error":"Failed to load preorder settings"}`, http.StatusInternalServerError)
		return
	}
	if settings == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"product_id":   productID,
			"enabled":      true,
			"start_date":   "",
			"end_date":     "",
			"sizes":        []string{},
			"layers":       []string{},
			"creams":       []string{},
			"flavors":      []string{},
			"size_prices":  map[string]float64{},
			"layer_prices": map[string]float64{},
			"cream_prices": map[string]float64{},
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id":   settings.ProductID,
		"enabled":      settings.Enabled,
		"start_date":   formatDate(settings.StartDate),
		"end_date":     formatDate(settings.EndDate),
		"sizes":        settings.Sizes,
		"layers":       settings.Layers,
		"creams":       settings.Creams,
		"flavors":      settings.Flavors,
		"size_prices":  settings.SizePrices,
		"layer_prices": settings.LayerPrices,
		"cream_prices": settings.CreamPrices,
		"updated_at":   settings.UpdatedAt,
	})
}

func AdminUpdatePreorderProductSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	idStr := mux.Vars(r)["id"]
	productID, err := strconv.Atoi(idStr)
	if err != nil || productID <= 0 {
		http.Error(w, `{"error":"Invalid product id"}`, http.StatusBadRequest)
		return
	}

	product, err := models.GetProductByID(configs.DB, productID)
	if err != nil {
		http.Error(w, `{"error":"Failed to load product"}`, http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, `{"error":"Product not found"}`, http.StatusNotFound)
		return
	}
	if strings.ToLower(strings.TrimSpace(product.Category)) != "cakes" {
		http.Error(w, `{"error":"Preorder settings are only available for cakes"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Enabled     *bool              `json:"enabled"`
		StartDate   *string            `json:"start_date"`
		EndDate     *string            `json:"end_date"`
		Sizes       []string           `json:"sizes"`
		Layers      []string           `json:"layers"`
		Creams      []string           `json:"creams"`
		Flavors     []string           `json:"flavors"`
		SizePrices  map[string]float64 `json:"size_prices"`
		LayerPrices map[string]float64 `json:"layer_prices"`
		CreamPrices map[string]float64 `json:"cream_prices"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Enabled == nil {
		http.Error(w, `{"error":"Missing enabled flag"}`, http.StatusBadRequest)
		return
	}

	startDate, err := parseDateString(req.StartDate)
	if err != nil {
		http.Error(w, `{"error":"Invalid start date"}`, http.StatusBadRequest)
		return
	}
	endDate, err := parseDateString(req.EndDate)
	if err != nil {
		http.Error(w, `{"error":"Invalid end date"}`, http.StatusBadRequest)
		return
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		http.Error(w, `{"error":"End date must be after start date"}`, http.StatusBadRequest)
		return
	}

	sizes := normalizeOptionList(req.Sizes)
	layers := normalizeOptionList(req.Layers)
	creams := normalizeOptionList(req.Creams)
	flavors := normalizeOptionList(req.Flavors)
	sizePrices := normalizePriceMap(sizes, req.SizePrices)
	layerPrices := normalizePriceMap(layers, req.LayerPrices)
	creamPrices := normalizePriceMap(creams, req.CreamPrices)

	settings, err := models.UpsertPreorderProductSettings(productID, *req.Enabled, startDate, endDate, sizes, layers, creams, flavors, sizePrices, layerPrices, creamPrices)
	if err != nil {
		http.Error(w, `{"error":"Failed to save preorder settings"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"product_id":   settings.ProductID,
		"enabled":      settings.Enabled,
		"start_date":   formatDate(settings.StartDate),
		"end_date":     formatDate(settings.EndDate),
		"sizes":        settings.Sizes,
		"layers":       settings.Layers,
		"creams":       settings.Creams,
		"flavors":      settings.Flavors,
		"size_prices":  settings.SizePrices,
		"layer_prices": settings.LayerPrices,
		"cream_prices": settings.CreamPrices,
		"updated_at":   settings.UpdatedAt,
	})
}

func normalizeUniqueIDs(ids []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func normalizeOptionList(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizePriceMap(options []string, prices map[string]float64) map[string]float64 {
	if len(options) == 0 {
		return map[string]float64{}
	}
	index := map[string]string{}
	for _, opt := range options {
		key := strings.ToLower(strings.TrimSpace(opt))
		if key == "" {
			continue
		}
		index[key] = opt
	}
	out := map[string]float64{}
	for rawKey, value := range prices {
		key := strings.ToLower(strings.TrimSpace(rawKey))
		if key == "" {
			continue
		}
		opt, ok := index[key]
		if !ok {
			continue
		}
		if value < 0 {
			value = 0
		}
		out[opt] = value
	}
	return out
}

func parseDateString(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}
