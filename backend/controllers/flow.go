package controllers

import (
	"fmt"
	"log"
	"strings"

	"bakeflow/models"
)

// handleMessage processes text messages from users
// handleMessage moved to `message_handler.go`.

// handlePostback processes button clicks (postback payloads)
// Full implementation moved to `postback_handler.go`.

// showHelp displays help information
// showHelp displays ordering instructions
func showHelp(userID string) {
	help := "🆘 *How to Order* / *မှာယူနည်း*\n\n" +
		"1️⃣ Choose what you'd like to order\n" +
		"    လိုချင်တဲ့ပစ္စည်းကို ရွေးပါ\n" +
		"2️⃣ Select quantity / အရေအတွက် ရွေးပါ\n" +
		"3️⃣ Enter your name / နာမည် ထည့်ပါ\n" +
		"4️⃣ Choose pickup or delivery\n" +
		"    ကိုယ်တိုင်ယူမလား ပို့မလား ရွေးပါ\n" +
		"5️⃣ Confirm your order / အတည်ပြုပါ\n\n" +
		"*You can type naturally:* / *သဘာဝအတိုင်း စာရိုက်နိုင်ပါတယ်*\n" +
		"• \"I want chocolate cake\" / \"ချောကလက်ကိတ်လိုချင်တယ်\"\n" +
		"• \"Give me 2\" / \"2 ခု ပေးပါ\"\n" +
		"• \"I want to cancel\" / \"ပယ်ဖျက်ချင်တယ်\"\n" +
		"• \"Show menu\" / \"မီနူး ပြပါ\"\n\n" +
		"*Quick Commands:*\n" +
		"• 'menu' - View products\n" +
		"• 'cancel' - Start over\n" +
		"• 'help' - Show this message"

	SendMessage(userID, help)

	// After showing help, open webview order form
	ShowWebviewOrderForm(userID)
}

// showOrderHistory displays recent orders (mock implementation)
// goBack handles the "Go Back" navigation
func goBack(userID string) {
	// Chat-based ordering has been removed
	// All navigation now goes through webview form
	ShowWebviewOrderForm(userID)
}

// ========== NEW FEATURES ==========

// Business logic moved to `order_service.go`.

// showOrderHistory displays user's past orders with beautiful card design
func showOrderHistory(userID string) {
	// Get all orders (in future, filter by userID)
	orders, err := models.GetAllOrders()
	if err != nil {
		log.Printf("❌ Error fetching orders: %v", err)
		SendMessage(userID, "😞 Sorry, couldn't load your order history. Please try again later.")
		return
	}

	// Check if empty
	if len(orders) == 0 {
		state := GetUserState(userID)
		emptyMsg := "🛒 **No Orders Yet!**\n\n" +
			"You haven't placed any orders with us.\n\n" +
			"Ready to try our delicious baked goods?\n\n" +
			"Type 'menu' to start ordering! 🍰"

		if state.Language == "my" {
			emptyMsg = "🛒 **မှာထားမှုမရှိသေးပါ!**\n\n" +
				"သင် ကျွန်ုပ်တို့နှင့် မှာထားမှုမလုပ်ရသေးပါ။\n\n" +
				"ကျွန်ုပ်တို့ရဲ့ အရသာရှိတဲ့ မုန့်တွေကို စမ်းကြည့်ဖို့ အဆင်သင့်လား?\n\n" +
				"'မီနူး' လို့ရိုက်ပြီး မှာယူလိုက်ပါ! 🍰"
		}

		SendMessage(userID, emptyMsg)
		return
	}

	// Show recent orders as cards (limit to 5 most recent)
	displayOrders := orders
	if len(orders) > 5 {
		displayOrders = orders[:5]
	}

	var elements []Element
	for _, order := range displayOrders {
		// Build items list
		itemsList := ""
		for i, item := range order.Items {
			if i < 3 {
				emoji := "🍰"
				if product, exists := ProductCatalog[item.Product]; exists {
					emoji = product.Emoji
				}
				itemsList += fmt.Sprintf("%d× %s %s\n", item.Quantity, emoji, item.Product)
			}
		}
		if len(order.Items) > 3 {
			itemsList += fmt.Sprintf("...and %d more items\n", len(order.Items)-3)
		}

		// Status badge
		statusEmoji := "⏳"
		statusText := "Pending"
		switch order.Status {
		case "pending":
			statusEmoji = "⏳"
			statusText = "Pending"
		case "preparing":
			statusEmoji = "👨‍🍳"
			statusText = "Preparing"
		case "ready":
			statusEmoji = "✅"
			statusText = "Ready"
		case "delivered":
			statusEmoji = "🎉"
			statusText = "Delivered"
		case "completed":
			statusEmoji = "✔️"
			statusText = "Completed"
		}

		// Delivery icon
		deliveryIcon := "🏠"
		if order.DeliveryType == "delivery" {
			deliveryIcon = "🚚"
		}

		// Format date
		dateStr := order.CreatedAt.Format("Jan 2, 3:04 PM")

		// Build subtitle
		subtitle := fmt.Sprintf("%s %s • %s %s\n%s\nTotal: $%.2f",
			statusEmoji, statusText,
			deliveryIcon, strings.Title(order.DeliveryType),
			dateStr,
			order.TotalAmount)

		element := Element{
			Title:    fmt.Sprintf("Order #%d - %s", order.ID, order.CustomerName),
			Subtitle: subtitle + "\n\n" + itemsList,
			Buttons: []Button{
				{
					Type:    "postback",
					Title:   "🔄 Reorder",
					Payload: fmt.Sprintf("REORDER_%d", order.ID),
				},
				{
					Type:    "postback",
					Title:   "⭐ Rate",
					Payload: fmt.Sprintf("RATE_ORDER_%d", order.ID),
				},
			},
		}

		elements = append(elements, element)
	}

	SendMessage(userID, fmt.Sprintf("📋 **Your Recent Orders** (Showing %d of %d)", len(displayOrders), len(orders)))
	SendGenericTemplate(userID, elements)
}

// showOrderTracking displays the status of a specific order or the most recent one
// If orderID is 0, shows the most recent order
func showOrderTracking(userID string, orderID int) {
	var order *models.Order
	var err error

	if orderID > 0 {
		// Get specific order by ID
		order, err = models.GetOrderByID(orderID)
		if err != nil || order == nil {
			SendMessage(userID, "📦 Order not found.\n\nTap the button below to place an order!")
			SendQuickReplies(userID, "Would you like to order?", []QuickReply{
				{ContentType: "text", Title: "🛒 Order Now", Payload: "ORDER_NOW"},
			})
			return
		}
	} else {
		// Get the user's most recent order
		orders, err := models.GetRecentOrdersBySenderID(userID, 1)
		if err != nil || len(orders) == 0 {
			SendMessage(userID, "📦 No recent orders found.\n\nTap the button below to place an order!")
			SendQuickReplies(userID, "Would you like to order?", []QuickReply{
				{ContentType: "text", Title: "🛒 Order Now", Payload: "ORDER_NOW"},
			})
			return
		}
		order = &orders[0]
	}

	statusInfo := map[string]struct {
		emoji   string
		message string
	}{
		"pending":    {"⏳", "Your order is being reviewed"},
		"confirmed":  {"✅", "Your order has been confirmed"},
		"preparing":  {"👨‍🍳", "Your order is being prepared"},
		"ready":      {"📦", "Your order is ready for pickup or delivery"},
		"delivering": {"🚗", "Your order is out for delivery"},
		"delivered":  {"✅", "Your order has been delivered"},
		"cancelled":  {"❌", "This order was cancelled"},
		"scheduled":  {"📅", "Your order is scheduled"},
	}

	statusKey := strings.ToLower(strings.TrimSpace(order.Status))
	info, ok := statusInfo[statusKey]
	if !ok {
		info = statusInfo["pending"]
		statusKey = "pending"
	}

	// Build items list
	itemsList := ""
	for i, item := range order.Items {
		if i >= 3 {
			itemsList += fmt.Sprintf("  ...and %d more\n", len(order.Items)-3)
			break
		}
		itemsList += fmt.Sprintf("  • %s × %d\n", item.Product, item.Quantity)
	}

	steps := []string{"pending", "preparing", "ready", "delivered"}
	stepLabels := map[string]string{
		"pending":    "Pending",
		"preparing":  "Preparing",
		"ready":      "Ready",
		"delivered":  "Delivered",
		"delivering": "Out for delivery",
	}
	if order.DeliveryType == "delivery" {
		steps = []string{"pending", "preparing", "delivering", "delivered"}
	}

	stepStatus := "pending"
	switch statusKey {
	case "pending", "confirmed", "scheduled":
		stepStatus = "pending"
	case "preparing":
		stepStatus = "preparing"
	case "ready", "delivering":
		if order.DeliveryType == "delivery" {
			stepStatus = "delivering"
		} else {
			stepStatus = "ready"
		}
	case "delivered":
		stepStatus = "delivered"
	}

	currentStepIndex := 0
	for i, s := range steps {
		if s == stepStatus {
			currentStepIndex = i
			break
		}
	}

	timelineLines := make([]string, 0, len(steps))
	for i, s := range steps {
		icon := "○"
		if i < currentStepIndex {
			icon = "✅"
		} else if i == currentStepIndex {
			icon = "🟡"
		}
		label := stepLabels[s]
		timelineLines = append(timelineLines, fmt.Sprintf("%s %s", icon, label))
	}
	progress := strings.Join(timelineLines, "\n")

	msg := fmt.Sprintf("📍 *Order Tracking*\n\n"+
		"Order #BF-%d\n"+
		"━━━━━━━━━━━━━━━\n\n"+
		"%s *%s*\n"+
		"%s\n\n"+
		"*Items:*\n%s\n"+
		"*Total:* $%.2f\n"+
		"*Type:* %s\n\n"+
		"*Progress:*\n%s",
		order.ID,
		info.emoji, strings.Title(statusKey),
		info.message,
		itemsList,
		order.TotalAmount,
		strings.Title(order.DeliveryType),
		progress)

	if order.DeliveryType == "delivery" && order.Address != "" {
		msg += fmt.Sprintf("\n*Deliver to:* %s", order.Address)
	}

	SendMessage(userID, msg)
}

// Rating handling moved to `order_service.go`.

// Business-hour checks moved to `order_service.go`.
