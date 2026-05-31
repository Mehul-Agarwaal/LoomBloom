package store

import "context"

func (s Store) CreateMachine(ctx context.Context, machine Machine) (Machine, error) {
	const query = `
		INSERT INTO machines (organization_id, machine_number, size, type, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, machine_number, size, type, status, created_at`

	err := s.db.QueryRowContext(ctx, query,
		machine.OrganizationID,
		machine.MachineNumber,
		machine.Size,
		machine.Type,
		machine.Status,
	).Scan(
		&machine.ID,
		&machine.OrganizationID,
		&machine.MachineNumber,
		&machine.Size,
		&machine.Type,
		&machine.Status,
		&machine.CreatedAt,
	)
	return machine, err
}

func (s Store) ListMachines(ctx context.Context, organizationID string) ([]Machine, error) {
	const query = `
		SELECT id, organization_id, machine_number, size, type, status, created_at
		FROM machines
		WHERE organization_id = $1
		ORDER BY machine_number`

	rows, err := s.db.QueryContext(ctx, query, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []Machine
	for rows.Next() {
		var machine Machine
		if err := rows.Scan(&machine.ID, &machine.OrganizationID, &machine.MachineNumber, &machine.Size, &machine.Type, &machine.Status, &machine.CreatedAt); err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if machines == nil {
		return []Machine{}, nil
	}
	return machines, nil
}
