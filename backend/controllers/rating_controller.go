package controllers

import (
	"bakeflow/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// RatingRequest represents a rating submission
type RatingRequest struct {
	ProductID int    `json:"product_id"`
	OrderID   int    `json:"order_id"`
	Stars     int    `json:"stars"`
	Comment   string `json:"comment,omitempty"`
}

// BulkRatingRequest for rating multiple products at once
type BulkRatingRequest struct {
	OrderID int `json:"order_id"`
	Ratings []struct {
		ProductID int    `json:"product_id"`
		Stars     int    `json:"stars"`
		Comment   string `json:"comment,omitempty"`
	} `json:"ratings"`
}

// SubmitRating handles POST /api/ratings
// Requires signed token for user authentication
func SubmitRating(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user from signed token
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		// Try to get from token
		tok := r.URL.Query().Get("t")
		if tok != "" {
			if psid, err := VerifyWebviewToken(tok); err == nil {
				userID = psid
			}
		}
	}

	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req RatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate stars
	if req.Stars < 1 || req.Stars > 5 {
		http.Error(w, `{"error": "Stars must be between 1 and 5"}`, http.StatusBadRequest)
		return
	}

	// Check if user can rate this order
	canRate, err := models.CanUserRateOrder(req.OrderID, userID)
	if err != nil {
		log.Printf("Error checking rating eligibility: %v", err)
		http.Error(w, `{"error": "Order not found"}`, http.StatusNotFound)
		return
	}

	if !canRate {
		http.Error(w, `{"error": "Cannot rate this order. It must be delivered first."}`, http.StatusForbidden)
		return
	}

	// Create rating
	rating := &models.ProductRating{
		ProductID: req.ProductID,
		OrderID:   req.OrderID,
		UserID:    userID,
		Stars:     req.Stars,
		Comment:   req.Comment,
	}

	if err := models.CreateProductRating(rating); err != nil {
		log.Printf("Error creating rating: %v", err)
		http.Error(w, `{"error": "Failed to save rating"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("⭐ User %s rated product #%d with %d stars for order #%d", userID, req.ProductID, req.Stars, req.OrderID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rating":  rating,
	})
}

// SubmitBulkRatings handles POST /api/ratings/bulk
// Rate multiple products from an order at once
func SubmitBulkRatings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user from token
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		tok := r.URL.Query().Get("t")
		if tok != "" {
			if psid, err := VerifyWebviewToken(tok); err == nil {
				userID = psid
			}
		}
	}

	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req BulkRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Check if user can rate this order
	canRate, err := models.CanUserRateOrder(req.OrderID, userID)
	if err != nil || !canRate {
		http.Error(w, `{"error": "Cannot rate this order"}`, http.StatusForbidden)
		return
	}

	// Save each rating
	saved := 0
	for _, r := range req.Ratings {
		if r.Stars < 1 || r.Stars > 5 {
			continue
		}

		rating := &models.ProductRating{
			ProductID: r.ProductID,
			OrderID:   req.OrderID,
			UserID:    userID,
			Stars:     r.Stars,
			Comment:   r.Comment,
		}

		if err := models.CreateProductRating(rating); err == nil {
			saved++
		}
	}

	log.Printf("⭐ User %s rated %d products for order #%d", userID, saved, req.OrderID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"saved":   saved,
	})
}

// GetProductRatings returns ratings for a specific product
// GET /api/products/{id}/ratings
func GetProductRatings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error": "Invalid product ID"}`, http.StatusBadRequest)
		return
	}

	// Get summary
	summary, err := models.GetProductRatingSummary(productID)
	if err != nil {
		// No ratings yet - return empty
		json.NewEncoder(w).Encode(map[string]interface{}{
			"product_id":   productID,
			"avg_rating":   0,
			"rating_count": 0,
			"ratings":      []interface{}{},
		})
		return
	}

	// Get recent ratings
	ratings, _ := models.GetProductRatings(productID, 20)
	if ratings == nil {
		ratings = []models.ProductRating{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"summary": summary,
		"ratings": ratings,
	})
}

// GetOrderRatingStatus checks if an order can be rated and which items are unrated
// GET /api/orders/{id}/rating-status
func GetOrderRatingStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, `{"error": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	// Get user from token
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		tok := r.URL.Query().Get("t")
		if tok != "" {
			if psid, err := VerifyWebviewToken(tok); err == nil {
				userID = psid
			}
		}
	}

	if userID == "" {
		http.Error(w, `{"error": "Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Check if can rate
	canRate, _ := models.CanUserRateOrder(orderID, userID)

	// Get existing ratings
	existingRatings, _ := models.GetUserRatingsForOrder(orderID, userID)
	if existingRatings == nil {
		existingRatings = []models.ProductRating{}
	}

	// Get unrated items
	unratedItems, _ := models.GetUnratedItemsForOrder(orderID, userID)
	if unratedItems == nil {
		unratedItems = []models.OrderItem{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id":         orderID,
		"can_rate":         canRate,
		"existing_ratings": existingRatings,
		"unrated_items":    unratedItems,
		"all_rated":        len(unratedItems) == 0 && len(existingRatings) > 0,
	})
}

// AdminGetAllRatings returns all product ratings for admin dashboard
// GET /api/admin/ratings
func AdminGetAllRatings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	summaries, err := models.GetAllProductRatings()
	if err != nil {
		log.Printf("Error fetching ratings: %v", err)
		http.Error(w, `{"error": "Failed to fetch ratings"}`, http.StatusInternalServerError)
		return
	}

	if summaries == nil {
		summaries = []models.ProductRatingSummary{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ratings": summaries,
	})
}
