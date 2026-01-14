package models

import (
	"database/sql"
	"time"

	"bakeflow/configs"
)

func GetRecentOrdersBySenderID(senderID string, limit int) ([]Order, error) {
	if configs.DB == nil {
		return nil, sql.ErrConnDone
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}

	rows, err := configs.DB.Query(`
		SELECT id, customer_name,
		       COALESCE(delivery_type, 'pickup') as delivery_type,
		       COALESCE(address, '') as address,
		       status, total_items,
		       COALESCE(subtotal, 0), COALESCE(delivery_fee, 0), COALESCE(total_amount, 0),
		       reordered_from, rating_id, COALESCE(sender_id, '') as sender_id, created_at, completed_at
		FROM orders
		WHERE sender_id = $1
		ORDER BY id DESC
		LIMIT $2
	`, senderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	for rows.Next() {
		var o Order
		var completedAt sql.NullTime
		var reorderedFrom sql.NullInt64
		var ratingID sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(
			&o.ID, &o.CustomerName, &o.DeliveryType, &o.Address, &o.Status, &o.TotalItems,
			&o.Subtotal, &o.DeliveryFee, &o.TotalAmount,
			&reorderedFrom, &ratingID, &o.SenderID, &createdAt, &completedAt,
		)
		if err != nil {
			return nil, err
		}
		o.CreatedAt = createdAt
		if completedAt.Valid {
			o.CompletedAt = &completedAt.Time
		}
		if reorderedFrom.Valid {
			v := int(reorderedFrom.Int64)
			o.ReorderedFrom = &v
		}
		if ratingID.Valid {
			v := int(ratingID.Int64)
			o.RatingID = &v
		}

		items, err := GetOrderItems(o.ID)
		if err == nil {
			o.Items = items
		}
		orders = append(orders, o)
	}

	return orders, nil
}
