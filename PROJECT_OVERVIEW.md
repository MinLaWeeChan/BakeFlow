# BakeFlow - Complete Project Overview

## 🎯 Project Summary

**BakeFlow** is a comprehensive Facebook Messenger chatbot system for a bakery business that enables customers to:
- Browse products dynamically from a database
- Place orders through conversational interface
- Track order status
- Rate orders and products
- View order history
- Use a webview-based order form for enhanced UX

The system includes:
- **Backend**: Go-based REST API with Facebook Messenger webhook integration
- **Frontend**: Next.js admin dashboard for order/product management
- **Database**: PostgreSQL with comprehensive schema for orders, products, ratings, stock management

---

## 🏗️ Architecture Overview

```
┌─────────────────┐
│  Facebook       │
│  Messenger      │
└────────┬────────┘
         │ Webhook (GET/POST)
         │
┌────────▼────────┐         ┌──────────────┐
│  Go Backend     │◄────────┤  PostgreSQL  │
│  (Port 8080)    │         │   Database   │
└────────┬────────┘         └──────────────┘
         │
         │ API Calls
         │
┌────────▼────────┐
│  Next.js        │
│  Frontend       │
│  (Port 3000)    │
└─────────────────┘
```

---

## 📁 Project Structure

### Backend (`/backend`)

#### Core Files
- **`main.go`**: Entry point, initializes DB, sets up routes, starts server
- **`configs/config.go`**: Database connection management
- **`routes/routes.go`**: HTTP route definitions and middleware

#### Controllers (Business Logic)
- **`webhook.go`**: Facebook webhook verification (GET) and message reception (POST)
- **`message_handler.go`**: Processes text messages with natural language understanding
- **`postback_handler.go`**: Handles button clicks and structured interactions
- **`order_service.go`**: Order creation, confirmation, reordering logic
- **`flow.go`**: Order flow management (cart, checkout, confirmation)
- **`ui_helpers.go`**: UI generation for Messenger (product cards, menus, receipts)
- **`messenger.go`**: Facebook Graph API integration (sending messages, templates)
- **`product_controller.go`**: Product CRUD operations
- **`rating_controller.go`**: Product/order rating system
- **`stock_api.go`**: Real-time stock checking and validation
- **`persistent_menu.go`**: Facebook Messenger persistent menu setup
- **`types.go`**: Shared data structures (CartItem, UserState, Product, etc.)

#### Models (Data Layer)
- **`order.go`**: Order and OrderItem models, database operations
- **`product.go`**: Product model, validation, analytics
- **`rating.go`**: Rating model and operations
- **`stock.go`**: Stock reservation system
- **`saved_order.go`**: Saved orders for quick reordering
- **`recent_orders.go`**: Recent orders retrieval

#### Database Migrations (`/migrations`)
- `002_add_order_items.sql`: Order items table
- `003_add_sender_id.sql`: Messenger sender ID tracking
- `004_create_products_system.sql`: Products, analytics, admin roles
- `005_create_saved_orders.sql`: Saved orders feature
- `006_add_order_scheduling.sql`: Order scheduling
- `009_stock_reservation_system.sql`: Stock reservation with expiration
- `010_product_ratings.sql`: Product ratings system

### Frontend (`/frontend`)

#### Pages (`/src/pages`)
- **`admin/dashboard.js`**: Main admin dashboard with orders, stats, charts
- **`admin/orders.js`**: Order management page
- **`admin/products.js`**: Product listing page
- **`admin/products/[id].js`**: Product edit page
- **`admin/products/new.js`**: New product creation

#### Components (`/src/components`)
- **`Sidebar.js`**: Navigation sidebar
- **`TopNavbar.js`**: Top navigation with notifications
- **`SummaryCards.js`**: Dashboard statistics cards
- **`RecentOrdersTable.js`**: Orders table with status updates
- **`SalesChart.js`**: Revenue chart (Recharts)
- **`PopularItems.js`**: Most ordered products
- **`NotificationPreviewCard.js`**: New order notifications

#### Public Files (`/public`)
- **`order-form.html`**: Webview order form (original)
- **`order-form-v2.html`**: Enhanced modular order form
- **`saved-orders.html`**: Saved orders webview
- **`order-details.html`**: Order details webview
- **`rate-order.html`**: Product rating webview
- **`/js/`**: JavaScript modules for webviews
- **`/css/`**: Stylesheets for webviews

---

## 🔄 Core Workflows

### 1. User Ordering Flow (Messenger Bot)

```
User sends message
    ↓
ReceiveWebhook() → Parse JSON
    ↓
handleMessage() or handlePostback()
    ↓
State Machine:
    language_selection → greeting → awaiting_product → 
    awaiting_quantity → awaiting_cart_decision → 
    awaiting_name → awaiting_delivery_type → 
    awaiting_address (if delivery) → confirming
    ↓
confirmOrder() → CreateOrder() → Save to DB
    ↓
Send confirmation message
```

**Key States:**
- `language_selection`: Choose English/Myanmar
- `greeting`: Welcome message
- `awaiting_product`: Show product catalog
- `awaiting_quantity`: Ask for quantity (1-5)
- `awaiting_cart_decision`: Add more or checkout
- `awaiting_name`: Get customer name
- `awaiting_delivery_type`: Pickup or delivery
- `awaiting_address`: Delivery address (if delivery)
- `confirming`: Show order summary, confirm

### 2. Product Display Flow

```
showProducts() → getProductElements()
    ↓
models.GetActiveProducts() → Query DB
    ↓
Map to Messenger Element format
    ↓
SendGenericTemplate() → Facebook Graph API
    ↓
User sees product carousel
```

### 3. Order Confirmation Flow

```
confirmOrder()
    ↓
Calculate totals (subtotal, delivery fee)
    ↓
models.CreateOrder() → Transaction
    ↓
Insert order + order_items
    ↓
Send confirmation message with:
    - Order ID
    - Cart items
    - Pricing breakdown
    - Delivery info
    ↓
Reset user state
```

### 4. Admin Dashboard Flow

```
Frontend: dashboard.js
    ↓
fetchOrders() → GET /api/admin/orders
    ↓
Backend: AdminGetOrders() → models.GetAllOrders()
    ↓
Return orders with items
    ↓
Frontend: Display in table, calculate stats
    ↓
Poll every 10 seconds for new orders
    ↓
Show notification for new pending orders
```

---

## 🗄️ Database Schema

### Core Tables

#### `orders`
- Order information (customer, delivery type, address, status)
- Financial data (subtotal, delivery_fee, total_amount)
- Tracking (sender_id for Messenger notifications)
- Scheduling (scheduled_for, schedule_type)

#### `order_items`
- Individual items in each order
- Product name, quantity, price
- Optional notes and image_url

#### `products`
- Product catalog (name, description, category, price)
- Stock management (stock count)
- Status (draft, active, inactive, archived)
- Soft delete (deleted_at)

#### `product_analytics`
- View counts
- Purchase counts
- Last viewed/purchased timestamps

#### `ratings`
- Order ratings (1-5 stars)
- Product ratings (linked to order items)
- Comments

#### `stock_reservations`
- Temporary stock holds during checkout
- Expiration timestamps
- Auto-cleanup job runs every minute

#### `saved_orders`
- Quick reorder functionality
- Linked to Messenger sender_id

---

## 🔌 API Endpoints

### Webhook
- `GET /webhook`: Facebook webhook verification
- `POST /webhook`: Receive messages from Messenger

### Orders
- `GET /orders`: Get all orders (legacy)
- `GET /api/admin/orders`: Admin order list
- `PUT /api/admin/orders/{id}/status`: Update order status
- `POST /api/admin/orders/{id}/cancel`: Cancel order
- `POST /api/chat/orders`: Create order from webview

### Products
- `GET /api/products`: List products (with filters)
- `GET /api/products/{id}`: Get single product
- `POST /api/products`: Create product
- `PUT /api/products/{id}`: Update product
- `DELETE /api/products/{id}`: Archive product
- `PATCH /api/products/{id}/status`: Update status
- `PATCH /api/products/{id}/stock`: Update stock
- `GET /api/products/low-stock`: Low stock alerts
- `GET /api/products/{id}/logs`: Audit logs

### Stock
- `GET /api/stock/{id}`: Get product stock status
- `POST /api/stock/bulk`: Bulk stock check
- `POST /api/stock/validate-cart`: Validate cart stock

### Ratings
- `GET /api/products/{id}/ratings`: Get product ratings
- `POST /api/ratings`: Submit rating
- `POST /api/ratings/bulk`: Bulk rating submission
- `GET /api/orders/{id}/rating-status`: Check if order rated
- `GET /api/admin/ratings`: All ratings (admin)

### User APIs (Webview)
- `GET /api/me/saved-orders`: Get saved orders (requires token)
- `POST /api/me/saved-orders`: Save order
- `DELETE /api/me/saved-orders/{id}`: Delete saved order
- `GET /api/me/recent-orders`: Recent orders

### Uploads
- `POST /api/uploads/cloudinary`: Upload product images

---

## 🎨 Key Features

### 1. Natural Language Processing
- Understands product names in English and Myanmar
- Handles quantity requests ("two", "2", "နှစ်")
- Recognizes commands ("menu", "cancel", "help")

### 2. Dynamic Product Catalog
- Products loaded from database (not hardcoded)
- Real-time stock checking
- Category-based filtering
- Image support with fallbacks

### 3. Stock Management
- Real-time stock validation
- Reservation system (prevents overselling)
- Auto-expiring reservations (cleanup job)
- Low stock alerts

### 4. Order Scheduling
- Schedule orders for future dates/times
- Different schedule types (asap, scheduled, etc.)

### 5. Rating System
- Order-level ratings
- Per-product ratings (via webview)
- Rating analytics

### 6. Saved Orders
- Quick reorder functionality
- Linked to Messenger user ID

### 7. Webview Integration
- Enhanced order form (HTML/JS)
- Product rating interface
- Order details view
- Saved orders management

### 8. Admin Dashboard
- Real-time order monitoring
- Order status management
- Product CRUD operations
- Analytics (sales charts, popular items)
- Notification system for new orders

---

## 🔐 Security & Configuration

### Environment Variables (`.env`)
```
VERIFY_TOKEN=verifyme123          # Facebook webhook verification
PAGE_ACCESS_TOKEN=EAA...          # Facebook Graph API token
DATABASE_URL=postgresql://...     # PostgreSQL connection string
PORT=8080                          # Server port (default: 8080)
WEBVIEW_BASE_URL=https://...      # Base URL for webviews (ngrok/production)
```

### Authentication
- Webview APIs use signed tokens (`?t=...`)
- Admin APIs (future: JWT/session-based)
- Facebook webhook signature verification (TODO)

---

## 🚀 Deployment Flow

### Development
1. **Backend**: `cd backend && go run main.go`
2. **Frontend**: `cd frontend && npm run dev`
3. **ngrok**: `ngrok http 8080` (for webhook)
4. **ngrok**: `ngrok http 3000` (for webviews)

### Production (Recommended)
- Backend: Deploy to cloud (Heroku, Railway, AWS, etc.)
- Frontend: Deploy to Vercel/Netlify
- Database: Managed PostgreSQL (Neon, Supabase, etc.)
- Set environment variables in hosting platform

---

## 📊 Data Flow Examples

### Example 1: User Places Order

```
1. User: "menu"
   → handleMessage() → showProducts()
   → getProductElements() → DB query
   → SendGenericTemplate() → Product carousel

2. User clicks "Order" on product
   → handlePostback("ORDER_PRODUCT_123")
   → Set state.CurrentProduct, state.State = "awaiting_quantity"
   → askQuantity() → Quick replies (1-5)

3. User clicks "2"
   → handlePostback("QTY_2")
   → state.CurrentQuantity = 2
   → addToCart() → Add to state.Cart
   → askAddMore() → "Add more" or "Checkout"

4. User clicks "Checkout"
   → handlePostback("CHECKOUT")
   → showCart() → askName()

5. User types name
   → handleMessage() → state.CustomerName = "John"
   → askDeliveryType() → Quick replies

6. User clicks "Delivery"
   → handlePostback("DELIVERY")
   → state.DeliveryType = "delivery"
   → Ask for address

7. User types address
   → handleMessage() → state.Address = "123 Main St"
   → showOrderSummary() → Generic template with confirmation

8. User clicks "Confirm"
   → handlePostback("CONFIRM_ORDER")
   → confirmOrder()
   → models.CreateOrder() → DB transaction
   → Send confirmation message
   → ResetUserState()
```

### Example 2: Admin Updates Order Status

```
1. Admin clicks "Mark as Delivered" in dashboard
   → Frontend: PUT /api/admin/orders/123/status
   → Body: { "status": "delivered" }

2. Backend: AdminUpdateOrderStatus()
   → models.UpdateOrderStatus(123, "delivered")
   → UPDATE orders SET status = 'delivered' WHERE id = 123

3. Backend: Send notification to customer (if sender_id exists)
   → SendOrderCardWithTag(senderID, orderID, ...)
   → Facebook Graph API → Customer receives update

4. Frontend: Refresh orders list
   → Order status updated in UI
```

---

## 🧪 Testing

### Manual Testing
1. **Webhook Verification**: `curl "http://localhost:8080/webhook?hub.mode=subscribe&hub.verify_token=verifyme123&hub.challenge=test"`
2. **Send Test Message**: Use Facebook Messenger to message your page
3. **Admin Dashboard**: Open `http://localhost:3000/admin/dashboard`

### Key Test Scenarios
- ✅ Order complete flow (product → quantity → checkout → confirm)
- ✅ Stock validation (out of stock products)
- ✅ Order status updates
- ✅ Rating submission
- ✅ Saved orders
- ✅ Natural language understanding

---

## 📝 Key Design Decisions

1. **In-Memory State**: User states stored in memory (use Redis/DB for production)
2. **State Machine**: Explicit state management for order flow
3. **Transaction Safety**: Order creation uses DB transactions
4. **Stock Reservations**: Temporary holds prevent overselling
5. **Webview for Complex UI**: Rating interface uses webview for better UX
6. **Polling for Admin**: Dashboard polls every 10s (consider WebSockets for production)

---

## 🔮 Future Enhancements

- [ ] Webhook signature verification
- [ ] Redis for user state persistence
- [ ] WebSocket for real-time admin updates
- [ ] Payment integration
- [ ] Order notifications via email/SMS
- [ ] Advanced analytics
- [ ] Multi-language support expansion
- [ ] Product variants (size, flavor)
- [ ] Discount/promotion system
- [ ] Inventory management
- [ ] Customer profiles

---

## 📚 Documentation Files

- `README.md`: Backend setup and webhook configuration
- `ARCHITECTURE.md`: System architecture diagrams
- `BOT_PRODUCT_FLOW.md`: Dynamic product integration guide
- `PRODUCT_MANAGEMENT.md`: Product system documentation
- `WEBVIEW_SETUP.md`: Webview configuration
- `IMPLEMENTATION_SUMMARY.md`: Implementation details
- `TROUBLESHOOTING.md`: Common issues and solutions

---

## 🎓 Learning Resources

- Facebook Messenger Platform: https://developers.facebook.com/docs/messenger-platform
- Go Web Development: Standard library + gorilla/mux
- Next.js: React framework for admin dashboard
- PostgreSQL: Relational database for orders/products

---

**Last Updated**: 2025-01-XX
**Project Status**: Production-Ready (with noted enhancements for scale)
