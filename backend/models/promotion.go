package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"bakeflow/configs"
)

// Promotion represents a discount promotion
type Promotion struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`  // PERCENT_OFF or BUY_X_GET_Y
	Rules     json.RawMessage `json:"rules"` // JSON rules
	Active    bool            `json:"active"`
	StartAt   time.Time       `json:"start_at"`
	EndAt     time.Time       `json:"end_at"`
	Priority  int             `json:"priority"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// PromotionRules represents the parsed rules JSON
type PromotionRules struct {
	// For PERCENT_OFF
	Percent    float64 `json:"percent,omitempty"`
	ProductIDs []int   `json:"productIds,omitempty"` // Empty = all products

	// For BUY_X_GET_Y
	BuyQty int `json:"buyQty,omitempty"`
	GetQty int `json:"getQty,omitempty"`

	BuyProductIDs   []int   `json:"buyProductIds,omitempty"`
	GetProductIDs   []int   `json:"getProductIds,omitempty"`
	DiscountType    string  `json:"discountType,omitempty"` // FREE, PERCENT_OFF, FIXED_PRICE
	DiscountPercent float64 `json:"discountPercent,omitempty"`
	FixedPrice      float64 `json:"fixedPrice,omitempty"`
}

// IsActive checks if promotion is currently active (within date range and active flag)
func (p *Promotion) IsActive() bool {
	if !p.Active {
		return false
	}
	now := time.Now()
	return now.After(p.StartAt) && now.Before(p.EndAt)
}

// ParseRules parses the rules JSON into PromotionRules
func (p *Promotion) ParseRules() (*PromotionRules, error) {
	var rules PromotionRules
	if err := json.Unmarshal(p.Rules, &rules); err != nil {
		return nil, err
	}
	return &rules, nil
}

// GetActivePromotions returns all currently active promotions, sorted by priority
func GetActivePromotions() ([]Promotion, error) {
	now := time.Now()

	query := `
		SELECT id, name, type, rules, active, start_at, end_at, priority, created_at, updated_at
		FROM promotions
		WHERE active = true
		  AND start_at <= $1
		  AND end_at > $1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := configs.DB.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []Promotion
	for rows.Next() {
		var p Promotion
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.Rules, &p.Active,
			&p.StartAt, &p.EndAt, &p.Priority, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			continue
		}
		promotions = append(promotions, p)
	}

	return promotions, nil
}

// GetPromotionByID returns a single promotion by ID
func GetPromotionByID(id int) (*Promotion, error) {
	var p Promotion
	query := `
		SELECT id, name, type, rules, active, start_at, end_at, priority, created_at, updated_at
		FROM promotions
		WHERE id = $1
	`

	err := configs.DB.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Type, &p.Rules, &p.Active,
		&p.StartAt, &p.EndAt, &p.Priority, &p.CreatedAt, &p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// CreatePromotion creates a new promotion
func CreatePromotion(p *Promotion) error {
	if p.Type != "PERCENT_OFF" && p.Type != "BUY_X_GET_Y" {
		return errors.New("invalid promotion type")
	}

	query := `
		INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	return configs.DB.QueryRow(
		query, p.Name, p.Type, p.Rules, p.Active, p.StartAt, p.EndAt, p.Priority,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

// UpdatePromotion updates an existing promotion
func UpdatePromotion(p *Promotion) error {
	query := `
		UPDATE promotions
		SET name = $2, type = $3, rules = $4, active = $5, 
		    start_at = $6, end_at = $7, priority = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	return configs.DB.QueryRow(
		query, p.ID, p.Name, p.Type, p.Rules, p.Active, p.StartAt, p.EndAt, p.Priority,
	).Scan(&p.UpdatedAt)
}

// GetAllPromotions returns all promotions (for admin)
func GetAllPromotions() ([]Promotion, error) {
	query := `
		SELECT id, name, type, rules, active, start_at, end_at, priority, created_at, updated_at
		FROM promotions
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := configs.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []Promotion
	for rows.Next() {
		var p Promotion
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Type, &p.Rules, &p.Active,
			&p.StartAt, &p.EndAt, &p.Priority, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			continue
		}
		promotions = append(promotions, p)
	}

	return promotions, nil
}

// DeletePromotion deletes a promotion by ID
func DeletePromotion(id int) error {
	_, err := configs.DB.Exec(`DELETE FROM promotions WHERE id = $1`, id)
	return err
}
