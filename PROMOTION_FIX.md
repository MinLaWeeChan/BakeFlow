# Promotion System Fix - Implementation Guide

## Problem Identified
Promotions were only being displayed in the banner but not actually applied during checkout/order submission. The discount calculated by the `/checkout` endpoint was never being saved to the database or reflected in the final order total.

## Root Cause
The frontend was calculating the promotion discount via the `/checkout` API endpoint and storing it in `window.currentCheckout`, but this information was **never sent** to the backend when creating the order via `/api/chat/orders`.

## Solution Implemented

### 1. Database Schema Update
Created migration file: `migrations/012_add_promotion_to_orders.sql`

Added two new columns to the `orders` table:
- `promotion_id INT REFERENCES promotions(id)` - Foreign key to track which promotion was applied
- `discount DECIMAL(10, 2)` - The discount amount applied to the order

### 2. Backend Changes

#### Updated Models (`models/order.go`)
Added fields to the `Order` struct:
```go
PromotionID   *int        `json:"promotion_id,omitempty"`
Discount      float64     `json:"discount"`
```

Updated database queries to include the new fields with proper backwards compatibility fallbacks.

#### Updated Controllers (`controllers/chat_order.go`)
- Extended `ChatOrderRequest` struct to include:
  ```go
  AppliedPromotion  *AppliedPromotion  `json:"appliedPromotion,omitempty"`
  Discount          float64            `json:"discount"`
  ```

- Modified `CreateChatOrder` function to:
  1. Extract promotion info from the request
  2. Calculate the final total: `total = subtotal - discount`
  3. Save `promotion_id` and `discount` to the database
  4. Display the discount in the order confirmation message to the user

- Updated insertion logic with multiple fallbacks for different database schema versions

### 3. Frontend Changes

#### Updated `public/js/order-form.js`
Modified `submitOrder()` function to include promotion data:
```javascript
// Include promotion data if one was applied
if (window.currentCheckout) {
    orderData.discount = window.currentCheckout.discount || 0;
    orderData.appliedPromotion = window.currentCheckout.appliedPromotion || null;
}
```

#### Updated `public/js/order.js`
Modified `processOrderSubmission()` function with the same promotion data inclusion:
```javascript
// Include promotion data if one was applied
if (window.currentCheckout) {
    orderData.discount = window.currentCheckout.discount || 0;
    orderData.appliedPromotion = window.currentCheckout.appliedPromotion || null;
}
```

## How It Works Now

### Order Flow
1. **Frontend calculates promotion**: User adds items to cart → `updateCart()` calls `calculateCheckoutWithPromotions()`
2. **Promotion displayed**: Cart bar shows strikethrough subtotal and discounted total
3. **Order creation**: When user clicks "Place Order", both the promotion data AND cart items are sent to backend
4. **Backend saves order**: Order is created with:
   - Subtotal: sum of all item prices
   - Discount: applied promotion amount
   - Total: subtotal - discount
   - Promotion ID: reference to which promotion was used
5. **Confirmation message**: User receives Messenger notification showing the discount applied

### Example Order Data Sent
```json
{
  "user_id": "1234567890",
  "items": [
    { "product_id": 1, "name": "Chocolate Cake", "qty": 1, "price": 25.00 }
  ],
  "customer_name": "John Doe",
  "discount": 5.00,
  "appliedPromotion": {
    "id": 1,
    "name": "20% Off Cakes",
    "type": "PERCENT_OFF",
    "description": "20% off all products"
  }
}
```

## Database Backwards Compatibility
The implementation includes fallback logic for different database schema versions:
1. **New schema**: With both promotion and scheduling columns
2. **Mid-version**: With scheduling but no promotion columns
3. **Legacy**: Neither promotion nor scheduling columns

## Testing Checklist
- [ ] Apply migration: `psql "$DATABASE_URL" -f migrations/012_add_promotion_to_orders.sql`
- [ ] Test creating order with promotion applied
- [ ] Verify `orders.promotion_id` is populated
- [ ] Verify `orders.discount` is saved correctly
- [ ] Check Messenger confirmation message shows discount
- [ ] Test with multiple promotions (should apply highest discount)
- [ ] Verify orders without promotions have `discount = 0`
- [ ] Test backwards compatibility if database not fully migrated

## Files Modified
1. ✅ `backend/migrations/012_add_promotion_to_orders.sql` (new)
2. ✅ `backend/models/order.go`
3. ✅ `backend/controllers/chat_order.go`
4. ✅ `frontend/public/js/order-form.js`
5. ✅ `frontend/public/js/order.js`

## Next Steps
1. Apply the migration to your database
2. Restart the backend server
3. Clear browser cache
4. Test creating an order with an active promotion
