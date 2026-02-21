package models

import (
	"bakeflow/configs"
	"database/sql"
	"errors"
	"time"
)

// ProductRating represents a customer rating for a product
type ProductRating struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	OrderID   int       `json:"order_id"`
	UserID    string    `json:"user_id"`
	Stars     int       `json:"stars"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ProductRatingSummary shows rating breakdown for a product
type ProductRatingSummary struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	AvgRating   float64 `json:"avg_rating"`
	RatingCount int     `json:"rating_count"`
	FiveStar    int     `json:"five_star"`
	FourStar    int     `json:"four_star"`
	ThreeStar   int     `json:"three_star"`
	TwoStar     int     `json:"two_star"`
	OneStar     int     `json:"one_star"`
}

// CreateProductRating adds a new rating for a product
func CreateProductRating(rating *ProductRating) error {
	if rating.Stars < 1 || rating.Stars > 5 {
		return errors.New("stars must be between 1 and 5")
	}

	query := `
		INSERT INTO product_ratings (product_id, order_id, user_id, stars, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (product_id, order_id, user_id) 
		DO UPDATE SET stars = $4, comment = $5
		RETURNING id, created_at
	`

	err := configs.DB.QueryRow(
		query,
		rating.ProductID,
		rating.OrderID,
		rating.UserID,
		rating.Stars,
		rating.Comment,
	).Scan(&rating.ID, &rating.CreatedAt)

	return err
}

// GetProductRating retrieves a specific rating
func GetProductRating(productID, orderID int, userID string) (*ProductRating, error) {
	var r ProductRating
	var comment sql.NullString

	err := configs.DB.QueryRow(`
		SELECT id, product_id, order_id, user_id, stars, comment, created_at
		FROM product_ratings
		WHERE product_id = $1 AND order_id = $2 AND user_id = $3
	`, productID, orderID, userID).Scan(
		&r.ID, &r.ProductID, &r.OrderID, &r.UserID,
		&r.Stars, &comment, &r.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if comment.Valid {
		r.Comment = comment.String
	}

	return &r, nil
}

// GetProductAvgRating returns the average rating for a product
func GetProductAvgRating(productID int) (float64, int, error) {
	var avgRating sql.NullFloat64
	var count int

	err := configs.DB.QueryRow(`
		SELECT avg_rating, rating_count FROM products WHERE id = $1
	`, productID).Scan(&avgRating, &count)

	if err != nil {
		return 0, 0, err
	}

	if avgRating.Valid {
		return avgRating.Float64, count, nil
	}

	return 0, 0, nil
}

// GetProductRatings returns all ratings for a product
func GetProductRatings(productID int, limit int) ([]ProductRating, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := configs.DB.Query(`
		SELECT id, product_id, order_id, user_id, stars, COALESCE(comment, ''), created_at
		FROM product_ratings
		WHERE product_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, productID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []ProductRating
	for rows.Next() {
		var r ProductRating
		if err := rows.Scan(&r.ID, &r.ProductID, &r.OrderID, &r.UserID, &r.Stars, &r.Comment, &r.CreatedAt); err != nil {
			continue
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

// GetUserRatingsForOrder returns all ratings a user has given for an order
func GetUserRatingsForOrder(orderID int, userID string) ([]ProductRating, error) {
	rows, err := configs.DB.Query(`
		SELECT id, product_id, order_id, user_id, stars, COALESCE(comment, ''), created_at
		FROM product_ratings
		WHERE order_id = $1 AND user_id = $2
		ORDER BY product_id
	`, orderID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []ProductRating
	for rows.Next() {
		var r ProductRating
		if err := rows.Scan(&r.ID, &r.ProductID, &r.OrderID, &r.UserID, &r.Stars, &r.Comment, &r.CreatedAt); err != nil {
			continue
		}
		ratings = append(ratings, r)
	}

	return ratings, nil
}

// GetProductRatingSummary returns detailed rating breakdown
func GetProductRatingSummary(productID int) (*ProductRatingSummary, error) {
	var s ProductRatingSummary

	err := configs.DB.QueryRow(`
		SELECT 
			product_id, product_name, 
			COALESCE(avg_rating, 0), COALESCE(rating_count, 0),
			COALESCE(five_star, 0), COALESCE(four_star, 0), 
			COALESCE(three_star, 0), COALESCE(two_star, 0), COALESCE(one_star, 0)
		FROM product_ratings_summary
		WHERE product_id = $1
	`, productID).Scan(
		&s.ProductID, &s.ProductName,
		&s.AvgRating, &s.RatingCount,
		&s.FiveStar, &s.FourStar, &s.ThreeStar, &s.TwoStar, &s.OneStar,
	)

	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetAllProductRatings returns ratings summary for all products (for admin)
func GetAllProductRatings() ([]ProductRatingSummary, error) {
	rows, err := configs.DB.Query(`
		SELECT 
			product_id, product_name, 
			COALESCE(avg_rating, 0), COALESCE(rating_count, 0),
			COALESCE(five_star, 0), COALESCE(four_star, 0), 
			COALESCE(three_star, 0), COALESCE(two_star, 0), COALESCE(one_star, 0)
		FROM product_ratings_summary
		ORDER BY rating_count DESC, avg_rating DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ProductRatingSummary
	for rows.Next() {
		var s ProductRatingSummary
		if err := rows.Scan(
			&s.ProductID, &s.ProductName,
			&s.AvgRating, &s.RatingCount,
			&s.FiveStar, &s.FourStar, &s.ThreeStar, &s.TwoStar, &s.OneStar,
		); err != nil {
			continue
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// CanUserRateOrder checks if a user can rate an order (must be delivered)
func CanUserRateOrder(orderID int, userID string) (bool, error) {
	var status string
	var senderID sql.NullString

	err := configs.DB.QueryRow(`
		SELECT status, sender_id FROM orders WHERE id = $1
	`, orderID).Scan(&status, &senderID)

	if err != nil {
		return false, err
	}

	// Must be delivered
	if status != "delivered" {
		return false, nil
	}

	// Must be the order owner
	if !senderID.Valid || senderID.String != userID {
		return false, nil
	}

	return true, nil
}

// GetUnratedItemsForOrder returns order items that haven't been rated yet
func GetUnratedItemsForOrder(orderID int, userID string) ([]OrderItem, error) {
	rows, err := configs.DB.Query(`
		SELECT 
			oi.id, oi.order_id, COALESCE(p.id, 0) as product_id, oi.product, oi.quantity, oi.price, 
			COALESCE(oi.note, ''), COALESCE(oi.image_url, ''), oi.created_at
		FROM order_items oi
		LEFT JOIN products p ON LOWER(p.name) = LOWER(oi.product)
		LEFT JOIN product_ratings pr 
			ON pr.order_id = oi.order_id 
			AND pr.user_id = $2
			AND pr.product_id = p.id
		WHERE oi.order_id = $1 AND pr.id IS NULL
	`, orderID, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		var pid sql.NullInt64
		if err := rows.Scan(&item.ID, &item.OrderID, &pid, &item.Product, &item.Quantity,
			&item.Price, &item.Note, &item.ImageURL, &item.CreatedAt); err != nil {
			continue
		}
		if pid.Valid && pid.Int64 > 0 {
			v := int(pid.Int64)
			item.ProductID = &v
		}
		items = append(items, item)
	}

	return items, nil
}
