package store

import "context"

func (s Store) CreateWorker(ctx context.Context, worker Worker) (Worker, error) {
	const query = `
		INSERT INTO workers (organization_id, name, phone, average_production, close_per_shift, number_of_shifts)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, name, phone, average_production, close_per_shift, number_of_shifts, created_at`

	err := s.db.QueryRowContext(ctx, query,
		worker.OrganizationID,
		worker.Name,
		worker.Phone,
		worker.AverageProduction,
		worker.ClosePerShift,
		worker.NumberOfShifts,
	).Scan(
		&worker.ID,
		&worker.OrganizationID,
		&worker.Name,
		&worker.Phone,
		&worker.AverageProduction,
		&worker.ClosePerShift,
		&worker.NumberOfShifts,
		&worker.CreatedAt,
	)
	return worker, err
}

func (s Store) ListWorkers(ctx context.Context, organizationID string) ([]Worker, error) {
	const query = `
		SELECT id, organization_id, name, phone, average_production, close_per_shift, number_of_shifts, created_at
		FROM workers
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workers []Worker
	for rows.Next() {
		var worker Worker
		if err := rows.Scan(&worker.ID, &worker.OrganizationID, &worker.Name, &worker.Phone, &worker.AverageProduction, &worker.ClosePerShift, &worker.NumberOfShifts, &worker.CreatedAt); err != nil {
			return nil, err
		}
		workers = append(workers, worker)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if workers == nil {
		return []Worker{}, nil
	}
	return workers, nil
}
