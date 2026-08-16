package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (s Store) CreateOrganization(ctx context.Context, org Organization) (Organization, error) {
	const query = `
		INSERT INTO organizations (name, gst_number, owner_name, owner_email, phone, subscription_plan, phone_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at`

	err := s.db.QueryRowContext(ctx, query, org.Name, org.GSTNumber, org.OwnerName, org.OwnerEmail, org.Phone, org.SubscriptionPlan, org.PhoneVerified).Scan(
		&org.ID,
		&org.Name,
		&org.GSTNumber,
		&org.OwnerName,
		&org.OwnerEmail,
		&org.Phone,
		&org.PhoneVerified,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) ListOrganizations(ctx context.Context) ([]Organization, error) {
	const query = `
		SELECT id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at
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
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.GSTNumber,
			&org.OwnerName,
			&org.OwnerEmail,
			&org.Phone,
			&org.PhoneVerified,
			&org.SubscriptionPlan,
			&org.CreatedAt,
		)
		if err != nil {
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
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (s Store) GetFirstOrganization(ctx context.Context) (Organization, error) {
	const query = `
		SELECT id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at
		FROM organizations
		ORDER BY created_at ASC
		LIMIT 1`
	var org Organization
	err := s.db.QueryRowContext(ctx, query).Scan(
		&org.ID,
		&org.Name,
		&org.GSTNumber,
		&org.OwnerName,
		&org.OwnerEmail,
		&org.Phone,
		&org.PhoneVerified,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) GetOrganization(ctx context.Context, id string) (Organization, error) {
	const query = `
		SELECT id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at
		FROM organizations
		WHERE id = $1`
	var org Organization
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.GSTNumber,
		&org.OwnerName,
		&org.OwnerEmail,
		&org.Phone,
		&org.PhoneVerified,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) GetOrganizationByEmail(ctx context.Context, email string) (Organization, error) {
	const query = `
		SELECT id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at
		FROM organizations
		WHERE owner_email = $1
		LIMIT 1`
	var org Organization
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&org.ID,
		&org.Name,
		&org.GSTNumber,
		&org.OwnerName,
		&org.OwnerEmail,
		&org.Phone,
		&org.PhoneVerified,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) GetOrganizationByPhone(ctx context.Context, phone string) (Organization, error) {
	const query = `
		SELECT id, name, gst_number, owner_name, owner_email, phone, phone_verified, subscription_plan, created_at
		FROM organizations
		WHERE phone = $1
		LIMIT 1`
	var org Organization
	err := s.db.QueryRowContext(ctx, query, phone).Scan(
		&org.ID,
		&org.Name,
		&org.GSTNumber,
		&org.OwnerName,
		&org.OwnerEmail,
		&org.Phone,
		&org.PhoneVerified,
		&org.SubscriptionPlan,
		&org.CreatedAt,
	)
	return org, err
}

func (s Store) UpdateOTP(ctx context.Context, orgID string, otp string, expiresAt time.Time) error {
	const query = `
		UPDATE organizations
		SET otp_code = $1, otp_expires_at = $2
		WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query, otp, expiresAt, orgID)
	return err
}

func (s Store) VerifyOTP(ctx context.Context, orgID string, otp string) (bool, error) {
	const query = `
		SELECT otp_code, otp_expires_at
		FROM organizations
		WHERE id = $1`
	var storedOTP string
	var expiresAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, orgID).Scan(&storedOTP, &expiresAt)
	if err != nil {
		return false, err
	}

	if storedOTP == "" || storedOTP != otp {
		return false, nil
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return false, nil
	}

	// Update organization as phone verified
	const updateQuery = `
		UPDATE organizations
		SET phone_verified = true, otp_code = '', otp_expires_at = NULL
		WHERE id = $1`
	_, err = s.db.ExecContext(ctx, updateQuery, orgID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s Store) UpdateSubscriptionPlan(ctx context.Context, orgID string, plan string) error {
	const query = `
		UPDATE organizations
		SET subscription_plan = $1
		WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, plan, orgID)
	return err
}

// Session management methods
func (s Store) CreateSession(ctx context.Context, orgID string, token string, expiresAt time.Time) (Session, error) {
	const query = `
		INSERT INTO organization_sessions (organization_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, token, expires_at, created_at`
	var sess Session
	err := s.db.QueryRowContext(ctx, query, orgID, token, expiresAt).Scan(
		&sess.ID,
		&sess.OrganizationID,
		&sess.Token,
		&sess.ExpiresAt,
		&sess.CreatedAt,
	)
	return sess, err
}

func (s Store) GetSession(ctx context.Context, token string) (Session, error) {
	const query = `
		SELECT id, organization_id, token, expires_at, created_at
		FROM organization_sessions
		WHERE token = $1 AND expires_at > NOW()`
	var sess Session
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&sess.ID,
		&sess.OrganizationID,
		&sess.Token,
		&sess.ExpiresAt,
		&sess.CreatedAt,
	)
	return sess, err
}

func (s Store) DeleteSession(ctx context.Context, token string) error {
	const query = `DELETE FROM organization_sessions WHERE token = $1`
	_, err := s.db.ExecContext(ctx, query, token)
	return err
}


