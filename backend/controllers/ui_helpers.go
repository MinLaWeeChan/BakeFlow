package controllers

import (
	"bakeflow/configs"
	"bakeflow/models"
	"fmt"
	"strconv"
	"strings"
)

// getProductElements returns product carousel elements from the database
func getProductElements() []Element {
	products, err := models.GetActiveProducts(configs.DB, 10, 0, "", "")
	if err != nil {
		return []Element{}
	}
	var elements []Element
	for _, p := range products {
		price := fmt.Sprintf("$%.2f", p.Price)
		img := p.ImageURL
		if img == "" {
			img = "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop"
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

		// Get product rating
		avgRating, ratingCount, _ := models.GetProductAvgRating(p.ID)
		ratingText := ""
		if ratingCount > 0 {
			stars := strings.Repeat("⭐", int(avgRating))
			ratingText = fmt.Sprintf("\n%s %.1f (%d reviews)", stars, avgRating, ratingCount)
		} else {
			ratingText = "\n⭐ No ratings yet"
		}

		subtitle := fmt.Sprintf("%s • %s%s", p.Description, price, ratingText)

		// Add buttons: Order and Rate
		buttons := []Button{
			{Type: "postback", Title: "🛒 Order", Payload: fmt.Sprintf("ORDER_PRODUCT_%d", p.ID)},
			{Type: "postback", Title: "⭐ Rate", Payload: fmt.Sprintf("RATE_PRODUCT_%d", p.ID)},
		}

		elements = append(elements, Element{
			Title:    emoji + " " + p.Name,
			ImageURL: img,
			Subtitle: subtitle,
			Buttons:  buttons,
		})
	}
	return elements
}

// showAbout displays company information and help instructions in user's language
func showAbout(userID string) {
	state := GetUserState(userID)
	var aboutMsg, helpMsg string

	if state.Language == "my" {
		aboutMsg = "🏪 ကျွန်ုပ်တို့အကြောင်း\n\n" +
			"BakeFlow သည် လတ်ဆတ်သော မုန့်များကို နေ့စဉ် ဖုတ်လုပ်သော မုန့်ဆိုင်ဖြစ်ပါသည်။\n\n" +
			"🎂 ကျွန်ုပ်တို့၏ အထူးမုန့်များ:\n" +
			"• ချောကလက် ကိတ်မုန့်\n" +
			"• ဗနီလာ ကိတ်မုန့်\n" +
			"• ဆော့ဘီ ကိတ်မုန့်\n" +
			"• ချိစ်ကိတ်မုန့်\n" +
			"• နီမုန့်\n" +
			"• ချောကလက် ကွတ်ကီး\n" +
			"• ဗာတာကွတ်ကီး\n" +
			"• အာလုမွန့်\n\n" +
			"📍 တည်နေရာ: ရန်ကုန်မြို့\n" +
			"⏰ ဖွင့်ချိန်: နံနက် 8:00 - ညနေ 8:00\n" +
			"📞 ဆက်သွယ်ရန်: +95 9 XXX XXX XXX"

		helpMsg = "\n\n❓ အသုံးပြုနည်း\n\n" +
			"သဘာဝဘာသာစကားဖြင့် ရိုက်နိုင်ပါတယ်:\n\n" +
			"• \"မီနူး\" သို့မဟုတ် \"မုန့်များ\"\n" +
			"• \"ချောကလက်ကိတ်မုန့်လိုချင်တယ်\"\n" +
			"• \"နှစ်ခု\" သို့မဟုတ် \"၂\"\n" +
			"• \"ပို့ပေးပါ\" သို့မဟုတ် \"ကိုယ်တိုင်ယူမယ်\"\n" +
			"• \"ပယ်ဖျက်\" သို့မဟုတ် \"အစကနေစမယ်\"\n\n" +
			"🛒 အော်ဒါမှာရန် 'မီနူး' လို့ရိုက်ပါ!"
	} else {
		aboutMsg = "🏪 About Us\n\n" +
			"BakeFlow is your neighborhood bakery, baking fresh daily!\n\n" +
			"🎂 Our Specialties:\n" +
			"• Chocolate Cake\n" +
			"• Vanilla Cake\n" +
			"• Strawberry Cake\n" +
			"• Cheesecake\n" +
			"• Red Velvet Cake\n" +
			"• Chocolate Cookies\n" +
			"• Butter Cookies\n" +
			"• Almond Croissant\n\n" +
			"📍 Location: Yangon, Myanmar\n" +
			"⏰ Hours: 8:00 AM - 8:00 PM\n" +
			"📞 Contact: +95 9 XXX XXX XXX"

		helpMsg = "\n\n❓ How to Use\n\n" +
			"You can type naturally:\n\n" +
			"• \"menu\" or \"show products\"\n" +
			"• \"I want chocolate cake\"\n" +
			"• \"two\" or \"2\"\n" +
			"• \"delivery please\" or \"pickup\"\n" +
			"• \"cancel\" or \"start over\"\n\n" +
			"🛒 Type 'menu' to start ordering!"
	}

	SendMessage(userID, aboutMsg+helpMsg)
}

// showLanguageSelection shows language choice at the beginning
func showLanguageSelection(userID string) {
	state := GetUserState(userID)
	state.State = "language_selection"

	welcomeMsg := "Hi there! 👋 မင်္ဂလာပါ! 👋\n\n" +
		"I'm BakeFlow Bot, your virtual bakery assistant (Beta). " +
		"I'm still learning, so I might not have all the answers yet, but I'll try to assist you the best I can! 🍰\n\n" +
		"ကျွန်တော် BakeFlow Bot ပါ၊ သင့်ရဲ့ မုန့်ဆိုင် အကူအညီပေး စက်ရုပ်ပါ (စမ်းသပ်ဗားရှင်း)。 " +
		"ကျွန်တော် ယခုတော့ သင်ယူနေဆဲဖြစ်တဲ့အတွက် အားလုံးကို မဖြေနိုင်သေးပေမယ့် တတ်နိုင်သမျှ အကောင်းဆုံး ကူညီပေးပါမယ်နော်! 🍰\n\n" +
		"Please select your language to get started.\n" +
		"စတင်ဖို့ ဘာသာစကားကို ရွေးချယ်ပါ။"

	SendMessage(userID, welcomeMsg)

	quickReplies := []QuickReply{
		{ContentType: "text", Title: "🇬🇧 English", Payload: "LANG_EN"},
		{ContentType: "text", Title: "🇲🇲 မြန်မာ", Payload: "LANG_MY"},
	}
	SendQuickReplies(userID, "Choose your language / ဘာသာစကား ရွေးပါ:", quickReplies)
}

// startOrderingFlow begins the ordering process with welcome message and simple menu
func startOrderingFlow(userID string) {
	state := GetUserState(userID)
	state.State = "main_menu"

	// Send welcome message with simple button menu
	if state.Language == "my" {
		SendMessage(userID, "🍰 BakeFlow မှ ကြိုဆိုပါတယ်!")
		showMainMenuSimple(userID)
	} else {
		SendMessage(userID, "🍰 Welcome to BakeFlow!")
		showMainMenuSimple(userID)
	}
}

// showMainMenu displays main menu as cards (like your screenshot)
func showMainMenu(userID string) {
	state := GetUserState(userID)

	var elements []Element

	if state.Language == "my" {
		elements = []Element{
			{
				Title:    "🛒 အော်ဒါမှာမယ်",
				Subtitle: "ကျွန်ုပ်တို့၏ လတ်ဆတ်သော မုန့်များကို ကြည့်ရှုပါ",
				ImageURL: "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "လုပ်ဆောင်မည်", Payload: "MENU_ORDER_PRODUCTS"}},
			},
			{
				Title:    "ℹ️ အကြောင်းနှင့်အကူအညီ",
				Subtitle: "ကျွန်ုပ်တို့အကြောင်းနှင့် အသုံးပြုနည်း",
				ImageURL: "https://images.unsplash.com/photo-1556910103-1c02745aae4d?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "ဖတ်ရှုမည်", Payload: "MENU_ABOUT"}},
			},
			{
				Title:    "🌐 ဘာသာပြောင်းမယ်",
				Subtitle: "English သို့ ပြောင်းလဲရန်",
				ImageURL: "https://images.unsplash.com/photo-1523050854058-8df90110c9f1?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "ပြောင်းမည်", Payload: "MENU_CHANGE_LANG"}},
			},
		}
	} else {
		elements = []Element{
			{
				Title:    "� Order Now",
				Subtitle: "Browse our fresh baked goods",
				ImageURL: "https://images.unsplash.com/photo-1578985545062-69928b1d9587?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "Start Order", Payload: "MENU_ORDER_PRODUCTS"}},
			},
			{
				Title:    "ℹ️ About & Help",
				Subtitle: "Learn about us and how to order",
				ImageURL: "https://images.unsplash.com/photo-1556910103-1c02745aae4d?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "Learn More", Payload: "MENU_ABOUT"}},
			},
			{
				Title:    "🌐 Change Language",
				Subtitle: "Switch to Myanmar language",
				ImageURL: "https://images.unsplash.com/photo-1523050854058-8df90110c9f1?w=300&h=200&fit=crop",
				Buttons:  []Button{{Type: "postback", Title: "Switch", Payload: "MENU_CHANGE_LANG"}},
			},
		}
	}

	SendGenericTemplate(userID, elements)
}

// showProducts displays the product catalog
func showProducts(userID string) {
	// Check business hours before showing products
	if !checkBusinessHours(userID) {
		return
	}

	// Show active promotions first
	showActivePromotions(userID)

	state := GetUserState(userID)
	state.State = "awaiting_product"
	SendTypingIndicator(userID, true)
	SendGenericTemplate(userID, getProductElements())
}

// showActivePromotions displays active promotions to the user
func showActivePromotions(userID string) {
	promotions, err := models.GetActivePromotions()
	if err != nil || len(promotions) == 0 {
		return // No promotions, silently skip
	}

	state := GetUserState(userID)

	// Show highest priority promotion
	promo := promotions[0]
	rules, err := promo.ParseRules()
	if err != nil {
		return
	}

	var promoMsg string
	if promo.Type == "PERCENT_OFF" && rules.Percent > 0 {
		if len(rules.ProductIDs) == 0 {
			promoMsg = fmt.Sprintf("🎉 **Special Promotion!**\n\n%d%% OFF on all products!\n\nValid until %s",
				int(rules.Percent), promo.EndAt.Format("Jan 2, 2006"))
		} else {
			promoMsg = fmt.Sprintf("🎉 **Special Promotion!**\n\n%d%% OFF on selected products!\n\nValid until %s",
				int(rules.Percent), promo.EndAt.Format("Jan 2, 2006"))
		}
	} else if promo.Type == "BUY_X_GET_Y" && rules.BuyQty > 0 && rules.GetQty > 0 {
		promoMsg = fmt.Sprintf("🎉 **Special Promotion!**\n\nBuy %d Get %d Free!\n\nValid until %s",
			rules.BuyQty, rules.GetQty, promo.EndAt.Format("Jan 2, 2006"))
	} else {
		promoMsg = fmt.Sprintf("🎉 **Special Promotion!**\n\n%s\n\nValid until %s",
			promo.Name, promo.EndAt.Format("Jan 2, 2006"))
	}

	if state.Language == "my" {
		if promo.Type == "PERCENT_OFF" && rules.Percent > 0 {
			promoMsg = fmt.Sprintf("🎉 **အထူးလျှော့စျေး!**\n\nထုတ်ကုန်အားလုံးတွင် %d%% လျှော့စျေး!\n\n%s အထိ စျေးနှုန်းများ",
				int(rules.Percent), promo.EndAt.Format("Jan 2, 2006"))
		} else if promo.Type == "BUY_X_GET_Y" {
			promoMsg = fmt.Sprintf("🎉 **အထူးလျှော့စျေး!**\n\n%d ခု ဝယ်ရင် %d ခု အလကားရ!**\n\n%s အထိ",
				rules.BuyQty, rules.GetQty, promo.EndAt.Format("Jan 2, 2006"))
		}
	}

	SendMessage(userID, promoMsg)
}

// askQuantity asks how many items the user wants
func askQuantity(userID string) {
	state := GetUserState(userID)

	quickReplies := []QuickReply{
		{ContentType: "text", Title: "1", Payload: "QTY_1"},
		{ContentType: "text", Title: "2", Payload: "QTY_2"},
		{ContentType: "text", Title: "3", Payload: "QTY_3"},
		{ContentType: "text", Title: "4", Payload: "QTY_4"},
		{ContentType: "text", Title: "5", Payload: "QTY_5"},
		{ContentType: "text", Title: "⬅️ Back", Payload: "GO_BACK"},
		{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
	}
	SendQuickReplies(userID, fmt.Sprintf("How many %s %s would you like?", state.CurrentEmoji, state.CurrentProduct), quickReplies)
}

// askName asks for the customer's name
func askName(userID string) {
	state := GetUserState(userID)
	state.State = "awaiting_name"

	// Send a message with quick reply options to go back
	quickReplies := []QuickReply{
		{ContentType: "text", Title: "⬅️ Back to Cart", Payload: "GO_BACK"},
		{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
	}
	SendQuickReplies(userID, "Great! What's your name?", quickReplies)
}

// addToCart adds the current product to the cart
func addToCart(userID string) {
	state := GetUserState(userID)

	// Add current product to cart
	cartItem := CartItem{
		Product:      state.CurrentProduct,
		ProductEmoji: state.CurrentEmoji,
		Quantity:     state.CurrentQuantity,
	}
	state.Cart = append(state.Cart, cartItem)

	// Clear current product
	state.CurrentProduct = ""
	state.CurrentEmoji = ""
	state.CurrentQuantity = 0

	// Ask if they want to add more
	askAddMore(userID)
}

// askAddMore asks if customer wants to add more items or checkout
func askAddMore(userID string) {
	state := GetUserState(userID)

	// Calculate total items in cart
	totalItems := 0
	for _, item := range state.Cart {
		totalItems += item.Quantity
	}

	// Show what was just added
	lastItem := state.Cart[len(state.Cart)-1]
	message := fmt.Sprintf("✅ %d× %s %s added\n\nCart: %d items",
		lastItem.Quantity, lastItem.ProductEmoji, lastItem.Product, totalItems)

	quickReplies := []QuickReply{
		{ContentType: "text", Title: "Add More", Payload: "ADD_MORE_ITEMS"},
		{ContentType: "text", Title: fmt.Sprintf("Checkout (%d)", totalItems), Payload: "CHECKOUT"},
		{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
	}

	state.State = "awaiting_cart_decision"
	SendQuickReplies(userID, message, quickReplies)
}

// showCart displays current cart contents
func showCart(userID string) {
	state := GetUserState(userID)

	if len(state.Cart) == 0 {
		SendMessage(userID, "🛒 Your cart is empty!\n\nLet's start ordering!")
		startOrderingFlow(userID)
		return
	}

	// Build cart display
	cartDisplay := "🛒 **Your Cart:**\n\n"
	totalItems := 0

	for _, item := range state.Cart {
		cartDisplay += fmt.Sprintf("• %d× %s %s\n", item.Quantity, item.ProductEmoji, item.Product)
		totalItems += item.Quantity
	}

	cartDisplay += fmt.Sprintf("\n**Total Items:** %d", totalItems)

	SendMessage(userID, cartDisplay)
}

// showOrderSummary displays the order summary and asks for confirmation
func showOrderSummary(userID string) {
	state := GetUserState(userID)

	deliveryIcon := "🏠"
	if state.DeliveryType == "delivery" {
		deliveryIcon = "🚚"
	}

	// Build cart items display with pricing
	cartDisplay := ""
	totalItems := 0
	for _, item := range state.Cart {
		itemPrice := 0.00
		if product, exists := ProductCatalog[item.Product]; exists {
			priceStr := strings.ReplaceAll(product.Price, "$", "")
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				itemPrice = price * float64(item.Quantity)
			}
		}
		cartDisplay += fmt.Sprintf("• %d× %s %s - $%.2f\n", item.Quantity, item.ProductEmoji, item.Product, itemPrice)
		totalItems += item.Quantity
	}

	// Calculate totals
	subtotal, deliveryFee, totalAmount := calculateOrderTotals(state.Cart, state.DeliveryType, state.Address)

	// Pricing breakdown
	pricingInfo := fmt.Sprintf(
		"\n💰 **Pricing:**\n"+
			"Subtotal: $%.2f\n"+
			"Delivery Fee: $%.2f\n"+
			"━━━━━━━━━━━━\n"+
			"**Total: $%.2f**",
		subtotal,
		deliveryFee,
		totalAmount,
	)

	summary := fmt.Sprintf(
		"📋 **Order Summary**\n\n"+
			"🛒 **Your Items:**\n"+
			"%s"+
			"%s\n\n"+
			"👤 **Customer:** %s\n"+
			"%s **%s**\n"+
			"📍 **Address:** %s\n\n"+
			"Everything look good?",
		cartDisplay,
		pricingInfo,
		state.CustomerName,
		deliveryIcon, strings.Title(state.DeliveryType),
		state.Address,
	)

	quickReplies := []QuickReply{
		{ContentType: "text", Title: "✅ Confirm Order", Payload: "CONFIRM_ORDER"},
		{ContentType: "text", Title: "❌ Cancel", Payload: "CANCEL_ORDER"},
	}
	SendQuickReplies(userID, summary, quickReplies)
}

// showMenu displays the product menu as text then shows product cards
func showMenu(userID string) {
	menu := "🍰 **BakeFlow Menu**\n\n" +
		"🎂 **Cakes**\n" +
		"  • Chocolate Cake - $25\n" +
		"  • Vanilla Cake - $24\n" +
		"  • Red Velvet Cake - $28\n\n" +
		"🥐 **Pastries**\n" +
		"  • Croissant - $4.50\n" +
		"  • Cinnamon Roll - $5\n\n" +
		"🧁 **Others**\n" +
		"  • Chocolate Cupcake - $3.50\n" +
		"  • Fresh Bread - $6\n" +
		"  • Coffee - $5\n\n" +
		"👇 Click the buttons below to order!"

	SendMessage(userID, menu)
	showProducts(userID)
}
