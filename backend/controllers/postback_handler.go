package controllers

import (
	"bakeflow/configs"
	"bakeflow/models"
	"log"
	"strconv"
	"strings"
)

// handlePostback processes button clicks (postback payloads)
func handlePostback(userID, payload string) {
	log.Printf("[Postback] User %s clicked: %s", userID, payload)
	state := GetUserState(userID)

	switch payload {
	// Language selection
	case "LANG_EN":
		state.Language = "en"
		state.State = "greeting"
		SendMessage(userID, "✅ English selected!")
		startOrderingFlow(userID)

	case "LANG_MY":
		state.Language = "my"
		state.State = "greeting"
		SendMessage(userID, "✅ မြန်မာဘာသာ ရွေးချယ်ပြီးပါပြီ!")
		startOrderingFlow(userID)

	// Persistent Menu Actions (from ☰ menu)
	case "MENU_ORDER":
		startOrderingFlow(userID)

	case "MENU_ORDER_HISTORY":
		showOrderHistory(userID)

	case "MENU_ABOUT":
		showAbout(userID) // Shows both About and Help combined

	case "MENU_CHANGE_LANG":
		showLanguageSelection(userID)

	// Main Menu Actions (from card buttons)
	case "MENU_ORDER_PRODUCTS":
		showProducts(userID)

	case "MENU_HELP":
		showHelp(userID)

	case "GET_STARTED":
		showLanguageSelection(userID)

	// Order Now / Order Again button
	case "ORDER_NOW":
		startOrderingFlow(userID)

	// Track Order button - show recent order status (legacy, shows latest)
	case "TRACK_ORDER":
		showOrderTracking(userID, 0)

	// Contact Support / Need Help button
	case "CONTACT_SUPPORT":
		showHelp(userID)

	case "VERIFY_CUSTOMER_CONFIRM":
		if confirmMessengerVerification(userID) {
			return
		}

	// Product selection
	case "ORDER_CHOCOLATE_CAKE":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Chocolate Cake"].Name
		state.CurrentEmoji = ProductCatalog["Chocolate Cake"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_VANILLA_CAKE":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Vanilla Cake"].Name
		state.CurrentEmoji = ProductCatalog["Vanilla Cake"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_RED_VELVET":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Red Velvet"].Name
		state.CurrentEmoji = ProductCatalog["Red Velvet"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_CROISSANT":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Croissant"].Name
		state.CurrentEmoji = ProductCatalog["Croissant"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_CINNAMON_ROLL":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Cinnamon Roll"].Name
		state.CurrentEmoji = ProductCatalog["Cinnamon Roll"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_CUPCAKE":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Chocolate Cupcake"].Name
		state.CurrentEmoji = ProductCatalog["Chocolate Cupcake"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_COFFEE":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Coffee"].Name
		state.CurrentEmoji = ProductCatalog["Coffee"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_BREAD":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Bread"].Name
		state.CurrentEmoji = ProductCatalog["Bread"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	case "ORDER_CHOCOLATE_CUPCAKE":
		if state.State != "awaiting_product" {
			SendMessage(userID, "⚠️ Please complete your current step first, or type 'cancel' to start over.")
			return
		}
		state.CurrentProduct = ProductCatalog["Chocolate Cupcake"].Name
		state.CurrentEmoji = ProductCatalog["Chocolate Cupcake"].Emoji
		state.State = "awaiting_quantity"
		SendTypingIndicator(userID, true)
		askQuantity(userID)

	// Quantity selection
	case "QTY_1":
		if state.State != "awaiting_quantity" {
			SendMessage(userID, "⚠️ Please select a product first!")
			return
		}
		state.CurrentQuantity = 1
		SendTypingIndicator(userID, true)
		addToCart(userID)

	case "QTY_2":
		if state.State != "awaiting_quantity" {
			SendMessage(userID, "⚠️ Please select a product first!")
			return
		}
		state.CurrentQuantity = 2
		SendTypingIndicator(userID, true)
		addToCart(userID)

	case "QTY_3":
		if state.State != "awaiting_quantity" {
			SendMessage(userID, "⚠️ Please select a product first!")
			return
		}
		state.CurrentQuantity = 3
		SendTypingIndicator(userID, true)
		addToCart(userID)

	case "QTY_4":
		if state.State != "awaiting_quantity" {
			SendMessage(userID, "⚠️ Please select a product first!")
			return
		}
		state.CurrentQuantity = 4
		SendTypingIndicator(userID, true)
		addToCart(userID)

	case "QTY_5":
		if state.State != "awaiting_quantity" {
			SendMessage(userID, "⚠️ Please select a product first!")
			return
		}
		state.CurrentQuantity = 5
		SendTypingIndicator(userID, true)
		addToCart(userID)

	// Cart actions
	case "ADD_MORE_ITEMS":
		showProducts(userID)

	case "CHECKOUT":
		// Show cart and ask for name
		showCart(userID)
		SendTypingIndicator(userID, true)
		askName(userID)

	// Navigation
	case "GO_BACK":
		goBack(userID)

	case "MAIN_MENU":
		ResetUserState(userID)
		startOrderingFlow(userID)

	// Delivery type
	case "PICKUP":
		state.DeliveryType = "pickup"
		state.Address = "Pickup at store"
		state.State = "confirming"
		SendTypingIndicator(userID, true)
		showOrderSummary(userID)

	case "DELIVERY":
		state.DeliveryType = "delivery"
		state.State = "awaiting_address"

		// Add navigation options when asking for address
		quickReplies := []QuickReply{
			{ContentType: "text", Title: "⬅️ Back", Payload: "GO_BACK"},
			{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
		}
		SendQuickReplies(userID, "Perfect! Please type your delivery address:\n(Street, City, ZIP)", quickReplies)

	// Order confirmation
	case "CONFIRM_ORDER":
		SendTypingIndicator(userID, true)
		confirmOrder(userID)

	case "CANCEL_ORDER":
		ResetUserState(userID)
		SendMessage(userID, "❌ Order cancelled.")
		SendMessage(userID, "━━━━━━━━━━━━━━━━━")
		SendMessage(userID, "Ready to start fresh? Type 'menu' to see our products!")

	// Mini Quick Order Form
	case "QUICK_SHOP":
		state.State = "quick_ordering"
		ShowWebviewOrderForm(userID) // Opens mini web app inside Messenger

	case "QUICK_SHOW_CART":
		showQuickCartSummary(userID)

	case "QUICK_CHECKOUT":
		handleQuickCheckout(userID)

	case "QUICK_ADD_MORE":
		state.State = "quick_ordering"
		ShowMiniOrderForm(userID)

	case "QUICK_CLEAR_CART":
		handleQuickClearCart(userID)

	// Special actions
	case "SHOW_MENU":
		showMenu(userID)

	// Rating actions (order-level)
	case "RATING_1":
		handleRating(userID, 1)
	case "RATING_2":
		handleRating(userID, 2)
	case "RATING_3":
		handleRating(userID, 3)
	case "RATING_4":
		handleRating(userID, 4)
	case "RATING_5":
		handleRating(userID, 5)
	case "SKIP_RATING":
		SendMessage(userID, "No problem! Feel free to rate us anytime.\n\nType 'menu' to order again! 🍰")
		ResetUserState(userID)

	// Product rating actions
	case "PRODUCT_RATING_1":
		handleProductRating(userID, 1)
	case "PRODUCT_RATING_2":
		handleProductRating(userID, 2)
	case "PRODUCT_RATING_3":
		handleProductRating(userID, 3)
	case "PRODUCT_RATING_4":
		handleProductRating(userID, 4)
	case "PRODUCT_RATING_5":
		handleProductRating(userID, 5)
	case "SKIP_PRODUCT_RATING":
		SendMessage(userID, "No problem! Feel free to rate products anytime.\n\nType 'menu' to see products! 🍰")
		ResetUserState(userID)

	default:
		// Dynamic product ordering by ID
		if strings.HasPrefix(payload, "ORDER_PRODUCT_") {
			idStr := strings.TrimPrefix(payload, "ORDER_PRODUCT_")
			if pid, err := strconv.Atoi(idStr); err == nil {
				if !checkBusinessHours(userID) {
					return
				}
				if p, err := models.GetProductByID(configs.DB, pid); err == nil && p != nil {
					emoji := "🍰"
					switch strings.ToLower(p.Category) {
					case "cakes":
						emoji = "🎂"
					case "cupcakes":
						emoji = "🧁"
					case "coffee":
						emoji = "☕"
					case "bread":
						emoji = "🍞"
					case "muffins":
						emoji = "🧁"
					case "tarts":
						emoji = "🥧"
					case "pastries":
						emoji = "🥐"
					}
					state.CurrentProduct = p.Name
					state.CurrentEmoji = emoji
					state.State = "awaiting_quantity"
					SendTypingIndicator(userID, true)
					askQuantity(userID)
					return
				}
			}
		}

		// Quick add by product ID (dynamic)
		if strings.HasPrefix(payload, "QUICK_ADD_PRODUCT_") {
			idStr := strings.TrimPrefix(payload, "QUICK_ADD_PRODUCT_")
			if pid, err := strconv.Atoi(idStr); err == nil {
				handleQuickAddProductByID(userID, pid)
				return
			}
		}

		// Quick add/view from webview
		if strings.HasPrefix(payload, "QUICK_ADD_") {
			productKey := strings.TrimPrefix(payload, "QUICK_ADD_")
			handleQuickAddProduct(userID, productKey)
			return
		}

		if strings.HasPrefix(payload, "QUICK_VIEW_") {
			productKey := strings.TrimPrefix(payload, "QUICK_VIEW_")
			handleQuickAddProduct(userID, productKey)
			return
		}

		// Check for dynamic payloads (REORDER_123, RATE_ORDER_123)
		if strings.HasPrefix(payload, "REORDER_") {
			orderIDStr := strings.TrimPrefix(payload, "REORDER_")
			if orderID, err := strconv.Atoi(orderIDStr); err == nil {
				if !checkBusinessHours(userID) {
					return
				}
				handleReorder(userID, orderID)
				return
			}
		}

		if strings.HasPrefix(payload, "RATE_ORDER_") {
			orderIDStr := strings.TrimPrefix(payload, "RATE_ORDER_")
			if orderID, err := strconv.Atoi(orderIDStr); err == nil {
				askForRating(userID, orderID)
				return
			}
		}

		// Rate a specific product (RATE_PRODUCT_123)
		if strings.HasPrefix(payload, "RATE_PRODUCT_") {
			productIDStr := strings.TrimPrefix(payload, "RATE_PRODUCT_")
			if productID, err := strconv.Atoi(productIDStr); err == nil {
				showProductRatingOptions(userID, productID)
				return
			}
		}

		// Rate product from order (RATE_PRODUCT_ORDER_123_456 = productID_orderID)
		if strings.HasPrefix(payload, "RATE_PRODUCT_ORDER_") {
			parts := strings.Split(strings.TrimPrefix(payload, "RATE_PRODUCT_ORDER_"), "_")
			if len(parts) == 2 {
				if productID, err1 := strconv.Atoi(parts[0]); err1 == nil {
					if orderID, err2 := strconv.Atoi(parts[1]); err2 == nil {
						askForProductRating(userID, productID, orderID)
						return
					}
				}
			}
		}

		// Track specific order by ID (TRACK_ORDER_123)
		if strings.HasPrefix(payload, "TRACK_ORDER_") {
			orderIDStr := strings.TrimPrefix(payload, "TRACK_ORDER_")
			if orderID, err := strconv.Atoi(orderIDStr); err == nil {
				showOrderTracking(userID, orderID)
				return
			}
		}

		SendMessage(userID, "Sorry, I didn't understand that. Let's start over!")
		ResetUserState(userID)
	}
}
