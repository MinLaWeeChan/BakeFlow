# Testing Promotions System - Step by Step

## Step 1: Verify Database Setup

```bash
cd backend
psql "$DATABASE_URL" -f migrations/verify_promotions.sql
```

You should see:
- ✅ Table structure with all columns
- ✅ `check_dates` constraint
- ✅ 4 indexes (active, dates, priority, type)
- ✅ `update_promotions_updated_at` trigger
- ✅ 2 example promotions (inactive)

## Step 2: Start Your Backend Server

```bash
cd backend
go run main.go
```

Make sure it starts without errors. You should see:
```
✅ Connected to PostgreSQL!
🚀 Server starting on port 8080...
```

## Step 3: Test Active Promotions API

In a new terminal:

```bash
# Should return empty array (no active promotions yet)
curl http://localhost:8080/promotions/active
```

Expected response:
```json
{
  "promotions": []
}
```

## Step 4: Create a Test Promotion

### Option A: Via SQL

```bash
psql "$DATABASE_URL" <<EOF
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  '20% Off Everything',
  'PERCENT_OFF',
  '{"percent": 20, "productIds": []}'::jsonb,
  true,
  NOW(),
  NOW() + INTERVAL '7 days',
  10
);
EOF
```

### Option B: Via API (if admin endpoint is ready)

```bash
curl -X POST http://localhost:8080/api/admin/promotions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "20% Off Everything",
    "type": "PERCENT_OFF",
    "rules": {"percent": 20, "productIds": []},
    "active": true,
    "start_at": "2025-01-23T00:00:00Z",
    "end_at": "2025-01-30T23:59:59Z",
    "priority": 10
  }'
```

## Step 5: Verify Promotion is Active

```bash
curl http://localhost:8080/promotions/active
```

Should now return:
```json
{
  "promotions": [
    {
      "id": 1,
      "name": "20% Off Everything",
      "type": "PERCENT_OFF",
      "rules": {"percent": 20, "productIds": []},
      "active": true,
      "start_at": "...",
      "end_at": "...",
      "priority": 10
    }
  ]
}
```

## Step 6: Test Checkout Calculation

```bash
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "cartItems": [
      {"productId": 1, "qty": 2, "unitPrice": 25.99}
    ]
  }'
```

Expected response:
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

## Step 7: Test Frontend Display

1. Start your frontend (if using webview):
   ```bash
   cd frontend
   npm run dev
   ```

2. Open the order form in browser or Messenger webview

3. You should see:
   - ✅ Promotion banner at the top (if active promotions exist)
   - ✅ Checkout shows discount when items are in cart

## Step 8: Test Chatbot Integration

1. Send a message to your Facebook Page: "menu" or "show products"

2. You should see:
   - ✅ Promotion message before products are shown
   - ✅ Product catalog with ratings

## Step 9: Test BUY_X_GET_Y Promotion

Create a BOGO promotion:

```bash
psql "$DATABASE_URL" <<EOF
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  'Buy 1 Get 1 Free - Chocolate Cake',
  'BUY_X_GET_Y',
  '{"buyQty": 1, "getQty": 1, "productIds": [1]}'::jsonb,
  true,
  NOW(),
  NOW() + INTERVAL '7 days',
  8
);
EOF
```

Test checkout with 2 items:

```bash
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "cartItems": [
      {"productId": 1, "qty": 2, "unitPrice": 25.99}
    ]
  }'
```

Expected: One item should be free (discount = 25.99)

## Step 10: Test Priority System

Create two promotions with different priorities:

```bash
psql "$DATABASE_URL" <<EOF
-- Lower priority (5)
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  '10% Off',
  'PERCENT_OFF',
  '{"percent": 10, "productIds": []}'::jsonb,
  true,
  NOW(),
  NOW() + INTERVAL '7 days',
  5
);

-- Higher priority (15) - should win
INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
VALUES (
  '20% Off',
  'PERCENT_OFF',
  '{"percent": 20, "productIds": []}'::jsonb,
  true,
  NOW(),
  NOW() + INTERVAL '7 days',
  15
);
EOF
```

Test checkout - should apply 20% (higher priority):

```bash
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "cartItems": [
      {"productId": 1, "qty": 1, "unitPrice": 25.99}
    ]
  }'
```

Should show 20% discount, not 10%.

## Troubleshooting

### API returns 404
- Check server is running on port 8080
- Verify routes are registered in `routes.go`

### No promotions returned
- Check `active = true` in database
- Verify dates: `start_at <= NOW() AND end_at > NOW()`
- Check timezone (uses UTC)

### Discount not applied
- Verify promotion is active
- Check product IDs match (if product-specific)
- Verify priority is set
- Check backend logs for errors

### Frontend not showing promotions
- Check browser console for errors
- Verify `/promotions/active` API works
- Check `promotions.js` is loaded
- Verify banner element exists in HTML

## Next Steps

1. ✅ Database migration complete
2. ✅ API endpoints working
3. ✅ Create real promotions for your bakery
4. ✅ Test with actual products
5. ✅ Monitor promotion performance

## Quick Reference

```bash
# View all promotions
psql "$DATABASE_URL" -c "SELECT id, name, type, active, priority FROM promotions;"

# Activate a promotion
psql "$DATABASE_URL" -c "UPDATE promotions SET active = true WHERE id = 1;"

# Deactivate a promotion
psql "$DATABASE_URL" -c "UPDATE promotions SET active = false WHERE id = 1;"

# Delete a promotion
psql "$DATABASE_URL" -c "DELETE FROM promotions WHERE id = 1;"
```
