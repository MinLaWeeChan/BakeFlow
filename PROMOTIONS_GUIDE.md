# Promotions System - User Guide

## Overview

The promotions system allows admins to create discounts that are automatically applied at checkout. Promotions do NOT change product base prices - they only apply discounts during checkout calculation.

## Features

- ✅ **PERCENT_OFF**: Percentage discount (e.g., 20% off)
- ✅ **BUY_X_GET_Y**: Buy X Get Y Free (e.g., Buy 1 Get 1)
- ✅ Automatic application at checkout (highest priority wins)
- ✅ Date-based activation (start_at, end_at)
- ✅ Product-specific or all products
- ✅ Frontend display (banner, badges)
- ✅ Chatbot integration

## Database Setup

Run the migration to create the promotions table:

```bash
cd backend
psql "$DATABASE_URL" -f migrations/011_create_promotions.sql
```

## API Endpoints

### Get Active Promotions
```
GET /promotions/active
```

Returns all currently active promotions (within date range and active=true).

**Response:**
```json
{
  "promotions": [
    {
      "id": 1,
      "name": "20% Off Everything",
      "type": "PERCENT_OFF",
      "rules": {
        "percent": 20,
        "productIds": []
      },
      "active": true,
      "start_at": "2025-01-01T00:00:00Z",
      "end_at": "2025-01-31T23:59:59Z",
      "priority": 10
    }
  ]
}
```

### Calculate Checkout with Promotions
```
POST /checkout
Content-Type: application/json

{
  "cartItems": [
    {
      "productId": 1,
      "qty": 2,
      "unitPrice": 25.99
    }
  ]
}
```

**Response:**
```json
{
  "subtotal": 51.98,
  "discount": 10.40,
  "appliedPromotion": {
    "id": 1,
    "name": "20% Off Everything",
    "type": "PERCENT_OFF",
    "description": "20% off all products"
  },
  "total": 41.58
}
```

### Admin Endpoints

- `GET /api/admin/promotions` - List all promotions
- `POST /api/admin/promotions` - Create new promotion
- `PUT /api/admin/promotions/{id}` - Update promotion

## Creating Promotions

### Example 1: 20% Off All Products

```sql
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  '20% Off Everything',
  'PERCENT_OFF',
  '{"percent": 20, "productIds": []}'::jsonb,
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP + INTERVAL '30 days',
  10
);
```

### Example 2: 15% Off Specific Products

```sql
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  '15% Off Cakes',
  'PERCENT_OFF',
  '{"percent": 15, "productIds": [1, 2, 3]}'::jsonb,
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP + INTERVAL '7 days',
  5
);
```

### Example 3: Buy 1 Get 1 Free

```sql
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  'BOGO Chocolate Cake',
  'BUY_X_GET_Y',
  '{"buyQty": 1, "getQty": 1, "productIds": [1]}'::jsonb,
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP + INTERVAL '14 days',
  8
);
```

### Example 4: Buy 2 Get 1 Free

```sql
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  'Buy 2 Get 1 Free',
  'BUY_X_GET_Y',
  '{"buyQty": 2, "getQty": 1, "productIds": [1, 2, 3]}'::jsonb,
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP + INTERVAL '7 days',
  7
);
```

## Promotion Rules

### PERCENT_OFF Rules
```json
{
  "percent": 20,           // Discount percentage (1-100)
  "productIds": []         // Empty array = all products, or [1,2,3] for specific products
}
```

### BUY_X_GET_Y Rules
```json
{
  "buyQty": 1,             // Quantity to buy
  "getQty": 1,             // Quantity to get free
  "productIds": [1]        // Required: specific product IDs
}
```

## Priority System

- Higher `priority` number = higher priority
- If multiple promotions are eligible, the highest priority one is applied
- Only ONE promotion is applied per checkout

## Frontend Integration

### Promotion Banner

The promotion banner automatically appears at the top of the order form when active promotions exist.

### Checkout Calculation

The frontend calls `/checkout` API which:
1. Calculates subtotal from cart items
2. Finds all active promotions
3. Checks eligibility for each promotion
4. Applies the highest priority eligible promotion
5. Returns subtotal, discount, and final total

**Frontend MUST NOT calculate discounts** - always use the backend API.

### Display in Checkout

When a promotion is applied, the checkout shows:
- Original subtotal (strikethrough)
- Discount amount
- Final total (highlighted)
- Promotion name/description

## Chatbot Integration

When users view products (`showProducts`), active promotions are automatically displayed as a message before showing the product catalog.

## Testing

### Test Active Promotions API
```bash
curl http://localhost:8080/promotions/active
```

### Test Checkout Calculation
```bash
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "cartItems": [
      {"productId": 1, "qty": 2, "unitPrice": 25.99}
    ]
  }'
```

## Business Rules

1. ✅ Promotions do NOT change product base prices
2. ✅ Discounts are calculated ONLY at checkout
3. ✅ Only ONE promotion applies per checkout (highest priority)
4. ✅ Promotions must be active AND within date range
5. ✅ Frontend displays promotions but doesn't calculate discounts
6. ✅ Backend validates eligibility and applies discounts

## Troubleshooting

### Promotion Not Showing
- Check `active = true` in database
- Verify `start_at <= NOW() AND end_at > NOW()`
- Check frontend is calling `/promotions/active`

### Discount Not Applied
- Verify promotion is active and within date range
- Check product IDs match (if product-specific)
- Verify priority is set correctly
- Check backend logs for eligibility errors

### Multiple Promotions
- Only highest priority promotion applies
- Lower priority promotions are ignored if higher one is eligible

## Admin Dashboard (Future)

Create an admin interface to:
- View all promotions
- Create/edit/delete promotions
- Toggle active status
- Set dates and priority
- Preview promotion rules

## Notes

- Product IDs in `productIds` array must match actual product IDs in database
- Empty `productIds` array means "all products"
- `BUY_X_GET_Y` requires specific product IDs (cannot be "all products")
- Date ranges use UTC timestamps
- Priority can be any integer (higher = more important)
