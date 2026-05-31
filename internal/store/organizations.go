package store

import (
	"context"
	"database/sql"
)

func (s Store) CreateOrganization(ctx context.Context, org Organization) (Organization, error) {
	const query = `
		INSERT INTO organizations (name, owner_name, phone, subscription_plan)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, owner_name, phone, subscription_plan, created_at`

	err := s.db.QueryRowContext(ctx, query, org.Name, org.OwnerName, org.Phone, org.SubscriptionPlan).Scan(
		&org.ID,
		&org.Name,
		&org.OwnerName,
		&org.Phone,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) ListOrganizations(ctx context.Context) ([]Organization, error) {
	const query = `
		SELECT id, name, owner_name, phone, subscription_plan, created_at
		FROM organizations
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.OwnerName, &org.Phone, &org.SubscriptionPlan, &org.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if orgs == nil {
		return []Organization{}, nil
	}
	return orgs, nil
}

func (s Store) OrganizationExists(ctx context.Context, id string) (bool, error) {
	const query = `SELECT 1 FROM organizations WHERE id = $1`

	var exists int
	err := s.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}
