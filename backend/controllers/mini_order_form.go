package controllers

import (
	"bakeflow/configs"
	"bakeflow/models"
	"fmt"
	"log"
	"strings"
)

// ShowMiniOrderForm displays an interactive quick-order interface
// Uses Messenger's Generic Template to show product picker with buttons
func ShowMiniOrderForm(userID string) {
	state := GetUserState(userID)
	state.State = "quick_ordering"

	// Build product elements with add/remove buttons
	elements := []Element{}

	// Load recent products from DB (all statuses)
	products, err := models.GetRecentProducts(configs.DB, 12, 0)
	if err != nil {
		log.Printf("❌ Failed to load active products: %v", err)
		products = []models.Product{}
	}

	for _, p := range products {
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

		buttons := []Button{
			{
				Type:    "postback",
				Title:   "➕ +1",
				Payload: fmt.Sprintf("QUICK_ADD_PRODUCT_%d", p.ID),
			},
			{
				Type:    "postback",
				Title:   "🛒 Order",
				Payload: fmt.Sprintf("ORDER_PRODUCT_%d", p.ID),
			},
		}

		element := Element{
			Title: fmt.Sprintf("%s %s", emoji, p.Name),
			ImageURL: func() string {
				if p.ImageURL != "" {
					return p.ImageURL
				}
				return "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop"
			}(),
			Subtitle: fmt.Sprintf("%s • $%.2f", p.Description, p.Price),
			Buttons:  buttons,
		}
		elements = append(elements, element)
	}

	// Add action buttons
	elements = append(elements, Element{
		Title:    "📋 My Cart",
		Subtitle: "Review items & checkout",
		Buttons: []Button{
			{
				Type:    "postback",
				Title:   "View Cart",
				Payload: "QUICK_SHOW_CART",
			},
			{
				Type:    "postback",
				Title:   "Proceed",
				Payload: "QUICK_CHECKOUT",
			},
		},
	})

	// Send generic template
	SendGenericTemplate(userID, elements)
}

// CartItem in types.go already defines: Product, ProductEmoji, Quantity
// This function adds Emoji and Price to the existing CartItem for quick orders

// handleQuickAddProduct quickly adds a product without quantity dialog
func handleQuickAddProduct(userID, productKey string) {
	state := GetUserState(userID)

	// Map product keys to products
	productMap := map[string]struct {
		Name  string
		Price float64
		Emoji string
	}{
		"QUICK_ORDER_CAKE":      {"Chocolate Cake", 25.00, "🍰"},
		"QUICK_ORDER_VANILLA":   {"Vanilla Cake", 22.00, "🎂"},
		"QUICK_ORDER_CROISSANT": {"Croissant", 8.00, "🥐"},
		"QUICK_ORDER_CINNAMON":  {"Cinnamon Roll", 12.00, "🌀"},
	}

	prod, exists := productMap[productKey]
	if !exists {
		SendMessage(userID, "❌ Product not found")
		return
	}

	// Add 1 unit to cart
	if state.Cart == nil {
		state.Cart = make([]CartItem, 0)
	}

	// Check if product already in cart, if yes increment qty
	found := false
	for i, item := range state.Cart {
		if item.Product == prod.Name {
			state.Cart[i].Quantity++
			found = true
			break
		}
	}

	// If not in cart, add it
	if !found {
		state.Cart = append(state.Cart, CartItem{
			Product:      prod.Name,
			ProductEmoji: prod.Emoji,
			Quantity:     1,
		})
	}

	// Confirm addition
	msg := fmt.Sprintf("✅ Added %s %s to cart!", prod.Emoji, prod.Name)
	if state.Language == "my" {
		msg = fmt.Sprintf("✅ %s %s စတုံအိုးသို့ ထည့်သွင်းပြီး!", prod.Emoji, prod.Name)
	}

	SendMessage(userID, msg)
	showQuickCartSummary(userID)
}

// handleQuickAddProductByID adds +1 of a DB product to cart
func handleQuickAddProductByID(userID string, productID int) {
	state := GetUserState(userID)

	p, err := models.GetProductByID(configs.DB, productID)
	if err != nil || p == nil {
		SendMessage(userID, "❌ Product not found")
		return
	}

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

	if state.Cart == nil {
		state.Cart = make([]CartItem, 0)
	}

	// Check if product already in cart
	found := false
	for i, item := range state.Cart {
		if item.ProductID == p.ID || item.Product == p.Name {
			state.Cart[i].Quantity++
			found = true
			break
		}
	}

	if !found {
		state.Cart = append(state.Cart, CartItem{
			Product:      p.Name,
			ProductEmoji: emoji,
			Quantity:     1,
			ProductID:    p.ID,
			Price:        p.Price,
		})
	}

	msg := fmt.Sprintf("✅ Added %s %s to cart!", emoji, p.Name)
	SendMessage(userID, msg)
	showQuickCartSummary(userID)
}

// showQuickCartSummary shows a compact cart view with action buttons
func showQuickCartSummary(userID string) {
	state := GetUserState(userID)

	if len(state.Cart) == 0 {
		SendMessage(userID, "🛒 Cart is empty. Add items to get started!")
		return
	}

	// Build cart summary - show items with simple prices
	summary := "🛒 **Your Quick Cart:**\n\n"
	total := 0.0

	for _, item := range state.Cart {
		price := item.Price
		if price <= 0 {
			price = 20.00
		}
		subtotal := price * float64(item.Quantity)
		total += subtotal
		summary += fmt.Sprintf("%s %s × %d = $%.2f\n",
			item.ProductEmoji, item.Product, item.Quantity, subtotal)
	}

	summary += fmt.Sprintf("\n**Total: $%.2f**\n\n", total)
	summary += "What would you like to do?"

	if state.Language == "my" {
		summary = "🛒 **သင်၏စတုံအိုး:**\n\n"
		for _, item := range state.Cart {
			price := item.Price
			if price <= 0 {
				price = 20.00
			}
			subtotal := price * float64(item.Quantity)
			total += subtotal
			summary += fmt.Sprintf("%s %s × %d = $%.2f\n",
				item.ProductEmoji, item.Product, item.Quantity, subtotal)
		}
		summary += fmt.Sprintf("\n**စုစုပေါင်း: $%.2f**\n\n", total)
		summary += "ဘာလုပ်မည်လဲ?"
	}

	quickReplies := []QuickReply{
		{ContentType: "text", Title: "➕ Add More", Payload: "QUICK_ADD_MORE"},
		{ContentType: "text", Title: "🛒 Review", Payload: "QUICK_SHOW_CART"},
		{ContentType: "text", Title: "✅ Checkout", Payload: "QUICK_CHECKOUT"},
		{ContentType: "text", Title: "❌ Clear", Payload: "QUICK_CLEAR_CART"},
	}

	SendQuickReplies(userID, summary, quickReplies)
}

// handleQuickCheckout processes the cart and begins checkout
func handleQuickCheckout(userID string) {
	state := GetUserState(userID)

	if len(state.Cart) == 0 {
		SendMessage(userID, "❌ Cart is empty. Please add items first!")
		ShowMiniOrderForm(userID)
		return
	}

	// Move to name entry
	state.State = "awaiting_name"
	quickReplies := []QuickReply{
		{ContentType: "text", Title: "⬅️ Back", Payload: "QUICK_ADD_MORE"},
		{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
	}

	msg := "📝 What's your name?"
	if state.Language == "my" {
		msg = "📝 သင်၏နာမည်ကဘာလဲ?"
	}

	SendQuickReplies(userID, msg, quickReplies)
}

// handleQuickClearCart empties the cart
func handleQuickClearCart(userID string) {
	state := GetUserState(userID)
	state.Cart = []CartItem{}

	msg := "🗑️ Cart cleared!"
	if state.Language == "my" {
		msg = "🗑️ စတုံအိုးအလွတ်ပြီး!"
	}
	SendMessage(userID, msg)

	// Show mini form again
	quickReplies := []QuickReply{
		{ContentType: "text", Title: "🛍️ Shop", Payload: "QUICK_SHOP"},
		{ContentType: "text", Title: "🏠 Home", Payload: "MENU_ORDER"},
	}
	SendQuickReplies(userID, "What next?", quickReplies)
}

// LogCartState logs the current cart for debugging
func LogCartState(userID string, state *UserState) {
	if len(state.Cart) == 0 {
		log.Printf("📋 [Cart %s] Empty", userID)
		return
	}

	total := 0.0
	priceMap := map[string]float64{
		"Chocolate Cake": 25.00,
		"Vanilla Cake":   22.00,
		"Croissant":      8.00,
		"Cinnamon Roll":  12.00,
	}
	log.Printf("📋 [Cart %s] Contents:", userID)
	for i, item := range state.Cart {
		price, exists := priceMap[item.Product]
		if !exists {
			price = 20.00
		}
		subtotal := price * float64(item.Quantity)
		total += subtotal
		log.Printf("   %d) %s × %d @ $%.2f = $%.2f",
			i+1, item.Product, item.Quantity, price, subtotal)
	}
	log.Printf("   Total: $%.2f", total)
}
