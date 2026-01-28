package controllers

import (
	"encoding/json"
	"math"
	"testing"

	"bakeflow/models"
)

func TestCalculateBuyXGetYDiscount_SameItemTargetsEligibleOnly(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	}

	cartItems := []CheckoutCartItem{
		{ProductID: 1, Qty: 1, UnitPrice: 100},
		{ProductID: 2, Qty: 2, UnitPrice: 20},
	}

	discount, _ := calculateBuyXGetYDiscount(rules, cartItems)
	if math.Abs(discount-20) > 1e-9 {
		t.Fatalf("expected discount=20, got=%v", discount)
	}
}

func TestCalculateBuyXGetYDiscount_SameItemQuantityRules(t *testing.T) {
	rules := &models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	}

	tests := []struct {
		name     string
		qty      int
		expected float64
	}{
		{name: "qty1_no_free", qty: 1, expected: 0},
		{name: "qty2_one_free", qty: 2, expected: 20},
		{name: "qty3_one_free", qty: 3, expected: 20},
		{name: "qty4_two_free", qty: 4, expected: 40},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cartItems := []CheckoutCartItem{
				{ProductID: 2, Qty: tc.qty, UnitPrice: 20},
			}
			discount, _ := calculateBuyXGetYDiscount(rules, cartItems)
			if math.Abs(discount-tc.expected) > 1e-9 {
				t.Fatalf("expected discount=%v, got=%v", tc.expected, discount)
			}
		})
	}
}

func TestSelectBestPromotions_StacksPercentOffAndBogoWhenNoOverlap(t *testing.T) {
	percentRulesJSON, err := json.Marshal(&models.PromotionRules{
		Percent:    20,
		ProductIDs: []int{1},
	})
	if err != nil {
		t.Fatalf("marshal percent rules: %v", err)
	}

	bogoRulesJSON, err := json.Marshal(&models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	})
	if err != nil {
		t.Fatalf("marshal bogo rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 101, Name: "20% off Test cake", Type: "PERCENT_OFF", Rules: percentRulesJSON, Priority: 1},
		{ID: 202, Name: "Buy 1 Get 1 Free Chocolate Cake", Type: "BUY_X_GET_Y", Rules: bogoRulesJSON, Priority: 2},
	}

	cartItems := []CheckoutCartItem{
		{ProductID: 1, Qty: 1, UnitPrice: 100},
		{ProductID: 2, Qty: 2, UnitPrice: 20},
	}
	subtotal := 140.0

	discount, appliedPromo, appliedPromos := selectBestPromotions(promos, cartItems, subtotal)
	if math.Abs(discount-40) > 1e-9 {
		t.Fatalf("expected discount=40, got=%v", discount)
	}
	if appliedPromo == nil {
		t.Fatalf("expected appliedPromo to be set")
	}
	if appliedPromo.ID != 202 {
		t.Fatalf("expected appliedPromo.ID=202, got=%v", appliedPromo.ID)
	}
	if len(appliedPromos) != 2 {
		t.Fatalf("expected 2 appliedPromos, got=%d", len(appliedPromos))
	}
}

func TestAllocatePromotionsToLineItems_SameProductPercentAndBogo_Qty2(t *testing.T) {
	bogoRulesJSON, err := json.Marshal(&models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	})
	if err != nil {
		t.Fatalf("marshal bogo rules: %v", err)
	}

	percentRulesJSON, err := json.Marshal(&models.PromotionRules{
		Percent:    20,
		ProductIDs: []int{2},
	})
	if err != nil {
		t.Fatalf("marshal percent rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 10, Name: "BOGO Chocolate", Type: "BUY_X_GET_Y", Rules: bogoRulesJSON, Priority: 2},
		{ID: 20, Name: "20% off Chocolate", Type: "PERCENT_OFF", Rules: percentRulesJSON, Priority: 1},
	}

	lineItems, discountTotal, _, _ := allocatePromotionsToLineItems(promos, []CheckoutCartItem{
		{ClientLineID: "l1", ProductID: 2, Qty: 2, UnitPrice: 10},
	})

	if len(lineItems) != 1 {
		t.Fatalf("expected 1 lineItem, got=%d", len(lineItems))
	}
	if lineItems[0].PaidQty != 1 || lineItems[0].FreeQty != 1 {
		t.Fatalf("expected paidQty=1 freeQty=1, got paidQty=%d freeQty=%d", lineItems[0].PaidQty, lineItems[0].FreeQty)
	}
	if math.Abs(discountTotal-12) > 1e-9 {
		t.Fatalf("expected total discount=12, got=%v", discountTotal)
	}
}

func TestAllocatePromotionsToLineItems_SameProductPercentAndBogo_Qty3(t *testing.T) {
	bogoRulesJSON, err := json.Marshal(&models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	})
	if err != nil {
		t.Fatalf("marshal bogo rules: %v", err)
	}

	percentRulesJSON, err := json.Marshal(&models.PromotionRules{
		Percent:    20,
		ProductIDs: []int{2},
	})
	if err != nil {
		t.Fatalf("marshal percent rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 10, Name: "BOGO Chocolate", Type: "BUY_X_GET_Y", Rules: bogoRulesJSON, Priority: 2},
		{ID: 20, Name: "20% off Chocolate", Type: "PERCENT_OFF", Rules: percentRulesJSON, Priority: 1},
	}

	lineItems, discountTotal, _, _ := allocatePromotionsToLineItems(promos, []CheckoutCartItem{
		{ClientLineID: "l1", ProductID: 2, Qty: 3, UnitPrice: 10},
	})

	if len(lineItems) != 1 {
		t.Fatalf("expected 1 lineItem, got=%d", len(lineItems))
	}
	if lineItems[0].PaidQty != 2 || lineItems[0].FreeQty != 1 {
		t.Fatalf("expected paidQty=2 freeQty=1, got paidQty=%d freeQty=%d", lineItems[0].PaidQty, lineItems[0].FreeQty)
	}
	if math.Abs(discountTotal-14) > 1e-9 {
		t.Fatalf("expected total discount=14, got=%v", discountTotal)
	}
}

func TestAllocatePromotionsToLineItems_BuyXGetYPlusPercentOnGetProduct(t *testing.T) {
	bxgyRulesJSON, err := json.Marshal(&models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{1},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	})
	if err != nil {
		t.Fatalf("marshal bxgy rules: %v", err)
	}

	percentRulesJSON, err := json.Marshal(&models.PromotionRules{
		Percent:    20,
		ProductIDs: []int{2},
	})
	if err != nil {
		t.Fatalf("marshal percent rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 10, Name: "Buy 1 A get 1 B", Type: "BUY_X_GET_Y", Rules: bxgyRulesJSON, Priority: 2},
		{ID: 20, Name: "20% off B", Type: "PERCENT_OFF", Rules: percentRulesJSON, Priority: 1},
	}

	lineItems, discountTotal, _, _ := allocatePromotionsToLineItems(promos, []CheckoutCartItem{
		{ClientLineID: "buy", ProductID: 1, Qty: 1, UnitPrice: 100},
		{ClientLineID: "get", ProductID: 2, Qty: 2, UnitPrice: 10},
	})

	if len(lineItems) != 2 {
		t.Fatalf("expected 2 lineItems, got=%d", len(lineItems))
	}

	if lineItems[0].ProductID != 1 || lineItems[0].FreeQty != 0 {
		t.Fatalf("expected buy product to have freeQty=0, got productId=%d freeQty=%d", lineItems[0].ProductID, lineItems[0].FreeQty)
	}
	if lineItems[1].ProductID != 2 || lineItems[1].PaidQty != 1 || lineItems[1].FreeQty != 1 {
		t.Fatalf("expected get product paidQty=1 freeQty=1, got productId=%d paidQty=%d freeQty=%d", lineItems[1].ProductID, lineItems[1].PaidQty, lineItems[1].FreeQty)
	}
	if math.Abs(discountTotal-12) > 1e-9 {
		t.Fatalf("expected total discount=12, got=%v", discountTotal)
	}
}

func TestAllocatePromotionsToLineItems_NonEligibleProductsUnaffected(t *testing.T) {
	percentRulesJSON, err := json.Marshal(&models.PromotionRules{
		Percent:    20,
		ProductIDs: []int{1},
	})
	if err != nil {
		t.Fatalf("marshal percent rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 20, Name: "20% off A", Type: "PERCENT_OFF", Rules: percentRulesJSON, Priority: 1},
	}

	lineItems, discountTotal, _, _ := allocatePromotionsToLineItems(promos, []CheckoutCartItem{
		{ClientLineID: "x", ProductID: 99, Qty: 2, UnitPrice: 10},
	})

	if len(lineItems) != 1 {
		t.Fatalf("expected 1 lineItem, got=%d", len(lineItems))
	}
	if lineItems[0].PaidQty != 2 || lineItems[0].FreeQty != 0 || len(lineItems[0].Discounts) != 0 {
		t.Fatalf("expected non-eligible product to have no discounts; got paidQty=%d freeQty=%d discounts=%d", lineItems[0].PaidQty, lineItems[0].FreeQty, len(lineItems[0].Discounts))
	}
	if math.Abs(discountTotal-0) > 1e-9 {
		t.Fatalf("expected total discount=0, got=%v", discountTotal)
	}
}

func TestAllocatePromotionsToLineItems_WrongFreeItemBugPrevention(t *testing.T) {
	bogoRulesJSON, err := json.Marshal(&models.PromotionRules{
		BuyQty:        1,
		GetQty:        1,
		BuyProductIDs: []int{2},
		GetProductIDs: []int{2},
		DiscountType:  "FREE",
	})
	if err != nil {
		t.Fatalf("marshal bogo rules: %v", err)
	}

	promos := []models.Promotion{
		{ID: 10, Name: "BOGO Chocolate", Type: "BUY_X_GET_Y", Rules: bogoRulesJSON, Priority: 1},
	}

	lineItems, discountTotal, _, _ := allocatePromotionsToLineItems(promos, []CheckoutCartItem{
		{ClientLineID: "test", ProductID: 1, Qty: 1, UnitPrice: 100},
		{ClientLineID: "choc", ProductID: 2, Qty: 2, UnitPrice: 20},
	})

	if len(lineItems) != 2 {
		t.Fatalf("expected 2 lineItems, got=%d", len(lineItems))
	}
	if lineItems[0].ProductID != 1 || lineItems[0].FreeQty != 0 || len(lineItems[0].Discounts) != 0 {
		t.Fatalf("expected Test cake to have no free/discounts; got productId=%d freeQty=%d discounts=%d", lineItems[0].ProductID, lineItems[0].FreeQty, len(lineItems[0].Discounts))
	}
	if lineItems[1].ProductID != 2 || lineItems[1].PaidQty != 1 || lineItems[1].FreeQty != 1 {
		t.Fatalf("expected Chocolate to have paidQty=1 freeQty=1; got paidQty=%d freeQty=%d", lineItems[1].PaidQty, lineItems[1].FreeQty)
	}
	if math.Abs(discountTotal-20) > 1e-9 {
		t.Fatalf("expected total discount=20, got=%v", discountTotal)
	}
}
