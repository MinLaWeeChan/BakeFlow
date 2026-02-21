package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// UI helper functions moved to `ui_helpers.go`.

// showAbout displays company information and help instructions in user's language
// showAbout moved to `ui_helpers.go`.

// showLanguageSelection shows language choice at the beginning
// showLanguageSelection moved to `ui_helpers.go`.

// startOrderingFlow begins the ordering process with welcome message and simple menu
// startOrderingFlow moved to `ui_helpers.go`.

// showMainMenu displays main menu as cards (like your screenshot)
// showMainMenu moved to `ui_helpers.go`.

// showProducts displays the product catalog
// showProducts moved to `ui_helpers.go`.

// askQuantity asks how many items the user wants
// askQuantity moved to `ui_helpers.go`.

// askName asks for the customer's name
// askName moved to `ui_helpers.go`.

// addToCart adds the current product to the cart
// addToCart moved to `ui_helpers.go`.

// askAddMore asks if customer wants to add more items or checkout
// askAddMore moved to `ui_helpers.go`.

// showCart displays current cart contents
// showCart moved to `ui_helpers.go`.

// VerifyWebhook handles Facebook Messenger webhook verification (GET requests)
//
// Facebook sends a GET request with these query parameters:
// - hub.mode=subscribe
// - hub.verify_token=<your_verify_token>
// - hub.challenge=<random_string>
//
// COMMON VERIFICATION FAILURES AND HOW TO FIX:
// 1. "Callback URL or verify token couldn't be validated"
//   - VERIFY_TOKEN in .env doesn't match the one in Meta Developer Console
//   - .env file not loaded properly (check godotenv.Load() in main.go)
//   - ngrok URL is wrong or expired (get new URL with `ngrok http 8080`)
//   - Server not running on the correct port
//
// 2. "URL is not available"
//   - Server not running (run `go run main.go`)
//   - Firewall blocking port 8080
//   - ngrok not forwarding to localhost:8080
//
// 3. "The URL couldn't be validated"
//   - Server returning wrong status code
//   - Not returning the hub.challenge value
//   - HTTPS required (ngrok provides this automatically)
func VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	// Get the verify token from environment
	verifyToken := os.Getenv("VERIFY_TOKEN")

	// Debug logging (helps troubleshoot verification issues)
	log.Printf("========== WEBHOOK VERIFICATION ATTEMPT ==========")
	log.Printf("Mode: %s", mode)
	log.Printf("Token received: %s", token)
	log.Printf("Token expected: %s", verifyToken)
	log.Printf("Challenge: %s", challenge)
	log.Printf("Full URL: %s", r.URL.String())

	// Check if verify token is loaded from .env
	if verifyToken == "" {
		log.Println("❌ ERROR: VERIFY_TOKEN is empty! Check your .env file")
		http.Error(w, "Server configuration error", http.StatusInternalServerError)
		return
	}

	// Verify that mode and token are correct
	if mode == "subscribe" && token == verifyToken {
		log.Println("✅ Webhook verified successfully!")

		// Respond with the challenge token from the request
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	// Verification failed
	log.Println("❌ Webhook verification FAILED")
	if mode != "subscribe" {
		log.Printf("   - Wrong mode: got '%s', expected 'subscribe'", mode)
	}
	if token != verifyToken {
		log.Printf("   - Token mismatch!")
		log.Printf("   - Received: '%s'", token)
		log.Printf("   - Expected: '%s'", verifyToken)
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
}

// ReceiveWebhook handles incoming messages from Facebook Messenger (POST requests)
//
// Facebook sends POST requests when users message your page with this structure:
//
//	{
//	  "object": "page",
//	  "entry": [{
//	    "id": "page_id",
//	    "time": 1234567890,
//	    "messaging": [{
//	      "sender": {"id": "user_id"},
//	      "recipient": {"id": "page_id"},
//	      "timestamp": 1234567890,
//	      "message": {
//	        "mid": "message_id",
//	        "text": "Hello!"
//	      }
//	    }]
//	  }]
//	}
func ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Println("========== INCOMING WEBHOOK POST ==========")
	log.Printf("Raw body: %s", string(body))

	// Parse the webhook payload
	var webhook WebhookPayload
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		// Still return 200 OK to Facebook so they don't retry
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("EVENT_RECEIVED"))
		return
	}

	// Verify this is a page subscription
	if webhook.Object != "page" {
		log.Printf("⚠️  Not a page webhook: %s", webhook.Object)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("EVENT_RECEIVED"))
		return
	}

	// Process each entry
	for _, entry := range webhook.Entry {
		log.Printf("Processing entry from page ID: %s", entry.ID)

		// Process each messaging event
		for _, event := range entry.Messaging {
			senderID := event.Sender.ID

			// Check if this is a quick reply (button click from quick reply)
			if event.Message.QuickReply != nil && event.Message.QuickReply.Payload != "" {
				log.Printf("⚡ Quick Reply from %s: %s", senderID, event.Message.QuickReply.Payload)
				handlePostback(senderID, event.Message.QuickReply.Payload)
				continue
			}

			// Check if this is a message event (text input)
			if event.Message.Text != "" {
				log.Printf("📨 Message from %s: %s", senderID, event.Message.Text)
				handleMessage(senderID, strings.TrimSpace(event.Message.Text))
				continue
			}

			// Check for postback (button clicks from structured messages)
			if event.Postback.Payload != "" {
				log.Printf("🔘 Postback from %s: %s", senderID, event.Postback.Payload)
				handlePostback(senderID, event.Postback.Payload)
			}
		}
	}

	// Always return 200 OK to Facebook within 20 seconds
	// Otherwise Facebook will retry the webhook multiple times
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("EVENT_RECEIVED"))
}

// The rest of the webhook logic (message/postback handlers and helpers)
// has been moved to `flow.go` for clarity and easier maintenance.
