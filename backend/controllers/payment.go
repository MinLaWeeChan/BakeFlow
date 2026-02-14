package controllers

import (
	"bakeflow/configs"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cloudinary "github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gorilla/mux"
	qrcode "github.com/skip2/go-qrcode"
)

// InitiatePaymentHandler creates a payment record (optional, can be implicit)
func InitiatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID int     `json:"order_id"`
		Amount  float64 `json:"amount"`
		Method  string  `json:"method"` // "kpay", "wave", etc.
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// In a real app, you might create a "Payment" record here with status "pending"
	// For this simple version, we'll just acknowledge it.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment initiated",
	})
}

// GetQRCodeHandler generates a QR code for a specific order amount
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"] // "order_123.png" -> we need to parse this if complex, or just use ID

	// Simply parse "order_{id}.png"
	var orderID int
	fmt.Sscanf(orderIDStr, "order_%d.png", &orderID)

	// In a real app, fetch order total to embed in QR
	// For now, generating a generic payment QR
	// You could embed a deep link like: "kpay://pay?amount=..."

	png, err := qrcode.Encode(fmt.Sprintf("https://bakeflow.com/pay/order/%d", orderID), qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

// UploadPaymentImageHandler handles the receipt screenshot upload via Cloudinary
func UploadPaymentImageHandler(w http.ResponseWriter, r *http.Request) {
	// 10MB limit
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate image type
	if err := validateImageHeader(header); err != nil {
		http.Error(w, "Unsupported image type", http.StatusBadRequest)
		return
	}

	orderIDStr := r.FormValue("order_id")
	orderID, _ := strconv.Atoi(orderIDStr)

	// --- Upload to Cloudinary ---
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		log.Println("❌ Cloudinary env vars not configured")
		http.Error(w, "Upload service not configured", http.StatusInternalServerError)
		return
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Printf("❌ Cloudinary init failed: %v", err)
		http.Error(w, "Upload service error", http.StatusInternalServerError)
		return
	}

	uploadParams := uploader.UploadParams{
		Folder:       "bakeflow/payments",
		ResourceType: "image",
		PublicID:     fmt.Sprintf("payment_%d_%d", orderID, time.Now().Unix()),
	}

	res, err := cld.Upload.Upload(r.Context(), file, uploadParams)
	if err != nil {
		log.Printf("❌ Cloudinary upload failed: %v", err)
		http.Error(w, "Image upload failed", http.StatusInternalServerError)
		return
	}

	proofURL := res.SecureURL

	// Save payment record to DB
	_, err = configs.DB.Exec(`
        INSERT INTO payments (order_id, user_id, amount, method, status, proof_url, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, orderID, r.FormValue("user_id"), 0, "manual_upload", "pending", proofURL)

	if err != nil {
		log.Printf("⚠️ Failed to insert payment record: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"url":     proofURL,
		"message": "Receipt uploaded successfully",
	})
}

// GetPaymentStatusHandler checks the status
func GetPaymentStatusHandler(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.URL.Query().Get("order_id")
	orderID, _ := strconv.Atoi(orderIDStr)

	// Check payments table
	var status string
	var proofURL sql.NullString
	err := configs.DB.QueryRow("SELECT status, proof_url FROM payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1", orderID).Scan(&status, &proofURL)
	if err == sql.ErrNoRows {
		status = "none"
	} else if err != nil {
		status = "error"
	}

	response := map[string]interface{}{
		"status": status,
	}
	if proofURL.Valid {
		response["proof_url"] = proofURL.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminGetPayments returns a list of payments, optionally filtered by status
func AdminGetPayments(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	query := `
		SELECT id, order_id, user_id, amount, method, status, proof_url, created_at
		FROM payments
	`
	var args []interface{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT 50"

	rows, err := configs.DB.Query(query, args...)
	if err != nil {
		log.Printf("❌ Error fetching payments: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to fetch payments",
		})
		return
	}
	defer rows.Close()

	type Payment struct {
		ID        int       `json:"id"`
		OrderID   int       `json:"order_id"`
		UserID    string    `json:"user_id"`
		Amount    float64   `json:"amount"`
		Method    string    `json:"method"`
		Status    string    `json:"status"`
		ProofURL  string    `json:"proof_url"`
		CreatedAt time.Time `json:"created_at"`
	}

	var payments []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Method, &p.Status, &p.ProofURL, &p.CreatedAt); err != nil {
			log.Printf("❌ Error scanning payment: %v", err)
			continue
		}
		payments = append(payments, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

// AdminVerifyPayment updates the payment status (approve/reject)
func AdminVerifyPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentIDStr := vars["id"]
	paymentID, _ := strconv.Atoi(paymentIDStr)

	var req struct {
		Status string `json:"status"` // "verified", "rejected"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid request",
		})
		return
	}

	if req.Status != "verified" && req.Status != "rejected" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invalid status",
		})
		return
	}

	// Begin transaction
	tx, err := configs.DB.Begin()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Database error",
		})
		return
	}
	defer tx.Rollback()

	// Update payment status
	_, err = tx.Exec("UPDATE payments SET status = $1 WHERE id = $2", req.Status, paymentID)
	if err != nil {
		log.Printf("❌ Failed to update payment status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update payment",
		})
		return
	}

	// If verified, update order status to 'confirmed' or 'preparing'
	if req.Status == "verified" {
		var orderID int
		err := tx.QueryRow("SELECT order_id FROM payments WHERE id = $1", paymentID).Scan(&orderID)
		if err == nil {
			// Update order status
			_, _ = tx.Exec("UPDATE orders SET status = 'confirmed' WHERE id = $1", orderID)
		}
	}

	if err := tx.Commit(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to commit transaction",
		})
		return
	}

	// Post-commit notifications (best effort)
	if req.Status == "verified" {
		// Fetch payment details to notify user
		var userID string
		var orderID int
		configs.DB.QueryRow("SELECT user_id, order_id FROM payments WHERE id = $1", paymentID).Scan(&userID, &orderID)

		if userID != "" {
			msg := fmt.Sprintf("✅ **Payment Verified!**\n\nYour payment for Order #%d has been confirmed. We are preparing your order now! 🍰", orderID)
			SendMessageWithTag(userID, msg, "POST_PURCHASE_UPDATE")
		}
	} else if req.Status == "rejected" {
		var userID string
		var orderID int
		configs.DB.QueryRow("SELECT user_id, order_id FROM payments WHERE id = $1", paymentID).Scan(&userID, &orderID)

		if userID != "" {
			msg := fmt.Sprintf("❌ **Payment Issue**\n\nWe couldn't verify your payment for Order #%d. Please upload a clear receipt or contact support.", orderID)
			SendMessageWithTag(userID, msg, "POST_PURCHASE_UPDATE")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Payment status updated",
	})
}
