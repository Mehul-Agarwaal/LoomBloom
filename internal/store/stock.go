package store

import "context"

func (s Store) CreateStockItem(ctx context.Context, item StockItem) (StockItem, error) {
	const query = `
		INSERT INTO stock_items (organization_id, name, quality, stored_on, quantity, unit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, name, quality, stored_on, quantity, unit, created_at`

	err := s.db.QueryRowContext(ctx, query,
		item.OrganizationID,
		item.Name,
		item.Quality,
		item.StoredOn,
		item.Quantity,
		item.Unit,
	).Scan(
		&item.ID,
		&item.OrganizationID,
		&item.Name,
		&item.Quality,
		&item.StoredOn,
		&item.Quantity,
		&item.Unit,
		&item.CreatedAt,
	)
	return item, err
}

func (s Store) ListStockItems(ctx context.Context, organizationID string) ([]StockItem, error) {
	const query = `
		SELECT id, organization_id, name, quality, stored_on, quantity, unit, created_at
		FROM stock_items
		WHERE organization_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StockItem
	for rows.Next() {
		var item StockItem
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Quality, &item.StoredOn, &item.Quantity, &item.Unit, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		return []StockItem{}, nil
	}
	return items, nil
}
