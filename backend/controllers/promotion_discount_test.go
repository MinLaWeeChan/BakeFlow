package controllers

import (
	"testing"

	"bakeflow/models"
)

func TestCalculateBuyXGetYDiscount_NewMode_Free_PicksBestGetItem(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:        2,
		GetQty:        1,
		BuyProductIDs: []int{1},
		GetProductIDs: []int{2, 3},
		DiscountType:  "FREE",
	}

	cart := []CheckoutCartItem{
		{ProductID: 1, Qty: 2, UnitPrice: 10},
		{ProductID: 2, Qty: 1, UnitPrice: 5},
		{ProductID: 3, Qty: 1, UnitPrice: 8},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cart)
	if discount != 8 {
		t.Fatalf("discount=%v, want %v", discount, 8.0)
	}
}

func TestCalculateBuyXGetYDiscount_NewMode_PercentOff(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:          2,
		GetQty:          1,
		BuyProductIDs:   []int{1},
		GetProductIDs:   []int{3},
		DiscountType:    "PERCENT_OFF",
		DiscountPercent: 50,
	}

	cart := []CheckoutCartItem{
		{ProductID: 1, Qty: 2, UnitPrice: 10},
		{ProductID: 3, Qty: 1, UnitPrice: 8},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cart)
	if discount != 4 {
		t.Fatalf("discount=%v, want %v", discount, 4.0)
	}
}

func TestCalculateBuyXGetYDiscount_NewMode_FixedPrice(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:        2,
		GetQty:        1,
		BuyProductIDs: []int{1},
		GetProductIDs: []int{3},
		DiscountType:  "FIXED_PRICE",
		FixedPrice:    3,
	}

	cart := []CheckoutCartItem{
		{ProductID: 1, Qty: 2, UnitPrice: 10},
		{ProductID: 3, Qty: 1, UnitPrice: 8},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cart)
	if discount != 5 {
		t.Fatalf("discount=%v, want %v", discount, 5.0)
	}
}

func TestCalculateBuyXGetYDiscount_NewMode_RequiresGetItems(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:        2,
		GetQty:        1,
		BuyProductIDs: []int{1},
		GetProductIDs: []int{3},
		DiscountType:  "FREE",
	}

	cart := []CheckoutCartItem{
		{ProductID: 1, Qty: 2, UnitPrice: 10},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cart)
	if discount != 0 {
		t.Fatalf("discount=%v, want %v", discount, 0.0)
	}
}

func TestCalculateBuyXGetYDiscount_Legacy_BogoSameProduct(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:     1,
		GetQty:     1,
		ProductIDs: []int{1},
	}

	cart := []CheckoutCartItem{
		{ProductID: 1, Qty: 2, UnitPrice: 10},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cart)
	if discount != 10 {
		t.Fatalf("discount=%v, want %v", discount, 10.0)
	}
}
