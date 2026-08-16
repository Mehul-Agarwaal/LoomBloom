package store

import "context"

func (s Store) CreateProductionEntry(ctx context.Context, entry ProductionEntry) (ProductionEntry, error) {
	const query = `
		INSERT INTO production_entries (organization_id, machine_id, worker_id, entry_date, shift, quantity, unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, organization_id, machine_id, worker_id, entry_date, shift, quantity, unit, created_at`

	err := s.db.QueryRowContext(ctx, query,
		entry.OrganizationID,
		entry.MachineID,
		entry.WorkerID,
		entry.EntryDate,
		entry.Shift,
		entry.Quantity,
		entry.Unit,
	).Scan(
		&entry.ID,
		&entry.OrganizationID,
		&entry.MachineID,
		&entry.WorkerID,
		&entry.EntryDate,
		&entry.Shift,
		&entry.Quantity,
		&entry.Unit,
		&entry.CreatedAt,
	)
	return entry, err
}

func (s Store) ListProductionEntries(ctx context.Context, organizationID string) ([]ProductionEntry, error) {
	const query = `
		SELECT id, organization_id, machine_id, worker_id, entry_date, shift, quantity, unit, created_at
		FROM production_entries
		WHERE organization_id = $1
		ORDER BY entry_date DESC, created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ProductionEntry
	for rows.Next() {
		var entry ProductionEntry
		if err := rows.Scan(&entry.ID, &entry.OrganizationID, &entry.MachineID, &entry.WorkerID, &entry.EntryDate, &entry.Shift, &entry.Quantity, &entry.Unit, &entry.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if entries == nil {
		return []ProductionEntry{}, nil
	}
	return entries, nil
}
