package store

import (
	"context"
)

type Dashboard struct {
	Organization      Organization
	MachineCount      int
	WorkerCount       int
	StockQuantity     float64
	RawMaterialCount  int
	SparePartCount    int
	PaymentTotal      float64
	TodayProduction   float64
	RecentMachines    []Machine
	RecentWorkers     []Worker
	RecentStockItems  []StockItem
	RecentProductions []ProductionEntry
}

func (s Store) Dashboard(ctx context.Context, org Organization) (Dashboard, error) {
	var dashboard Dashboard
	dashboard.Organization = org
	organizationID := org.ID

	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM machines WHERE organization_id = $1`, organizationID).Scan(&dashboard.MachineCount); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workers WHERE organization_id = $1`, organizationID).Scan(&dashboard.WorkerCount); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity), 0) FROM stock_items WHERE organization_id = $1`, organizationID).Scan(&dashboard.StockQuantity); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM raw_materials WHERE organization_id = $1`, organizationID).Scan(&dashboard.RawMaterialCount); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM spare_parts WHERE organization_id = $1`, organizationID).Scan(&dashboard.SparePartCount); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE organization_id = $1`, organizationID).Scan(&dashboard.PaymentTotal); err != nil {
		return dashboard, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity), 0) FROM production_entries WHERE organization_id = $1 AND entry_date = CURRENT_DATE`, organizationID).Scan(&dashboard.TodayProduction); err != nil {
		return dashboard, err
	}

	machines, err := s.ListMachines(ctx, organizationID)
	if err != nil {
		return dashboard, err
	}
	workers, err := s.ListWorkers(ctx, organizationID)
	if err != nil {
		return dashboard, err
	}
	stockItems, err := s.ListStockItems(ctx, organizationID)
	if err != nil {
		return dashboard, err
	}
	productions, err := s.ListProductionEntries(ctx, organizationID)
	if err != nil {
		return dashboard, err
	}

	dashboard.RecentMachines = limitMachines(machines, 4)
	dashboard.RecentWorkers = limitWorkers(workers, 4)
	dashboard.RecentStockItems = limitStockItems(stockItems, 4)
	dashboard.RecentProductions = limitProductionEntries(productions, 5)

	return dashboard, nil
}

func limitMachines(items []Machine, limit int) []Machine {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitWorkers(items []Worker, limit int) []Worker {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitStockItems(items []StockItem, limit int) []StockItem {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitProductionEntries(items []ProductionEntry, limit int) []ProductionEntry {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
