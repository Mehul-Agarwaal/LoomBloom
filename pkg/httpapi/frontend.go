package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"loombloom/pkg/store"
	"loombloom/pkg/views"
)

func (api API) dashboard(w http.ResponseWriter, r *http.Request) {
	org, ok := GetOrg(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data, err := api.store.Dashboard(r.Context(), org)
	if err != nil {
		api.logger.Error("load dashboard", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardPage(data).Render(r.Context(), w); err != nil {
		api.logger.Error("render dashboard", "error", err)
	}
}

// --- Production Handlers ---

func (api API) productionList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	entries, err := api.store.ListProductionEntries(ctx, org.ID)
	if err != nil {
		api.logger.Error("list production entries", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Currently using existing ProductionPage which expects (org, entries, machines, workers, errStr)
	machines, _ := api.store.ListMachines(ctx, org.ID)
	workers, _ := api.store.ListWorkers(ctx, org.ID)
	if err := views.ProductionPage(org, entries, machines, workers, "").Render(ctx, w); err != nil {
		api.logger.Error("render production page", "error", err)
	}
}

func (api API) productionForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	machines, _ := api.store.ListMachines(ctx, org.ID)
	workers, _ := api.store.ListWorkers(ctx, org.ID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.ProductionFormPage(org, machines, workers, "").Render(ctx, w); err != nil {
		api.logger.Error("render production form", "error", err)
	}
}

func (api API) productionCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	machineID := r.FormValue("machine_id")
	workerID := r.FormValue("worker_id")
	dateStr := r.FormValue("entry_date")
	shift := r.FormValue("shift")
	qtyStr := r.FormValue("quantity")
	unit := r.FormValue("unit")

	qty, err := strconv.ParseFloat(qtyStr, 64)
	if err != nil {
		http.Error(w, "Quantity must be a valid number.", http.StatusBadRequest)
		return
	}
	if machineID == "" {
		http.Error(w, "Loom / Machine is required.", http.StatusBadRequest)
		return
	}

	entryDate := time.Now().UTC()
	if dateStr != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
			entryDate = parsedDate
		}
	}

	if unit == "" {
		unit = "meters"
	}

	_, err = api.store.CreateProductionEntry(ctx, store.ProductionEntry{
		OrganizationID: org.ID,
		MachineID:      machineID,
		WorkerID:       workerID,
		EntryDate:      entryDate,
		Shift:          shift,
		Quantity:       qty,
		Unit:           unit,
	})
	if err != nil {
		api.logger.Error("create production entry from web", "error", err)
		http.Error(w, "Failed to create production entry: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/production", http.StatusSeeOther)
}

// --- Machine Handlers ---

func (api API) machinesList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	machines, err := api.store.ListMachines(ctx, org.ID)
	if err != nil {
		api.logger.Error("list machines", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.MachinesPage(org, machines, "").Render(ctx, w); err != nil {
		api.logger.Error("render machines page", "error", err)
	}
}

func (api API) machineForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.MachineFormPage(org, "").Render(ctx, w); err != nil {
		api.logger.Error("render machine form", "error", err)
	}
}

func (api API) machineCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	machineNum := r.FormValue("machine_number")
	size := r.FormValue("size")
	machineType := r.FormValue("type")
	status := r.FormValue("status")

	if machineNum == "" {
		http.Error(w, "Machine number is required.", http.StatusBadRequest)
		return
	}
	_, err := api.store.CreateMachine(ctx, store.Machine{
		OrganizationID: org.ID,
		MachineNumber:  machineNum,
		Size:           size,
		Type:           machineType,
		Status:         status,
	})
	if err != nil {
		api.logger.Error("create machine from web", "error", err)
		http.Error(w, "Failed to create machine: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/machines", http.StatusSeeOther)
}

func (api API) machineDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// Mock implementation for detail view, we would typically fetch the machine by ID
	// machineID := r.PathValue("id")
	// For now we will just list machines and pick the first one to simulate it
	machines, err := api.store.ListMachines(ctx, org.ID)
	if err != nil || len(machines) == 0 {
		http.Redirect(w, r, "/machines", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.MachineDetailPage(org, machines[0]).Render(ctx, w); err != nil {
		api.logger.Error("render machine detail", "error", err)
	}
}

// --- Worker Handlers ---

func (api API) workersList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	workers, err := api.store.ListWorkers(ctx, org.ID)
	if err != nil {
		api.logger.Error("list workers", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.WorkersPage(org, workers, "").Render(ctx, w); err != nil {
		api.logger.Error("render workers page", "error", err)
	}
}

func (api API) workerForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.WorkerFormPage(org, "").Render(ctx, w); err != nil {
		api.logger.Error("render worker form", "error", err)
	}
}

func (api API) workerCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	avgProdStr := r.FormValue("average_production")
	closesStr := r.FormValue("close_per_shift")
	shiftsStr := r.FormValue("number_of_shifts")

	avgProd, _ := strconv.ParseFloat(avgProdStr, 64)
	closes, _ := strconv.Atoi(closesStr)
	shifts, _ := strconv.Atoi(shiftsStr)

	if name == "" {
		http.Error(w, "Worker name is required.", http.StatusBadRequest)
		return
	}
	if shifts < 1 {
		shifts = 1
	}
	_, err := api.store.CreateWorker(ctx, store.Worker{
		OrganizationID:    org.ID,
		Name:              name,
		Phone:             phone,
		AverageProduction: avgProd,
		ClosePerShift:     closes,
		NumberOfShifts:    shifts,
	})
	if err != nil {
		api.logger.Error("create worker from web", "error", err)
		http.Error(w, "Failed to create worker: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/workers", http.StatusSeeOther)
}

func (api API) workerDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Mock implementation for detail view
	workers, err := api.store.ListWorkers(ctx, org.ID)
	if err != nil || len(workers) == 0 {
		http.Redirect(w, r, "/workers", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.WorkerDetailPage(org, workers[0]).Render(ctx, w); err != nil {
		api.logger.Error("render worker detail", "error", err)
	}
}

// --- Inventory Handlers ---

func (api API) inventoryList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	stock, err := api.store.ListStockItems(ctx, org.ID)
	if err != nil {
		api.logger.Error("list stock items", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Using InventoryPage
	if err := views.InventoryPage(org, stock, "").Render(ctx, w); err != nil {
		api.logger.Error("render inventory page", "error", err)
	}
}

func (api API) inventoryInwardForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.InventoryInwardFormPage(org, "").Render(ctx, w); err != nil {
		api.logger.Error("render inventory form", "error", err)
	}
}

func (api API) inventoryInwardCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	quality := r.FormValue("quality")
	storedOn := r.FormValue("stored_on")
	qtyStr := r.FormValue("quantity")
	unit := r.FormValue("unit")

	qty, err := strconv.ParseFloat(qtyStr, 64)
	if err != nil {
		http.Error(w, "Quantity must be a valid number.", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(w, "Stock item name is required.", http.StatusBadRequest)
		return
	}
	if unit == "" {
		unit = "meters"
	}
	_, err = api.store.CreateStockItem(ctx, store.StockItem{
		OrganizationID: org.ID,
		Name:           name,
		Quality:        quality,
		StoredOn:       storedOn,
		Quantity:       qty,
		Unit:           unit,
	})
	if err != nil {
		api.logger.Error("create stock item from web", "error", err)
		http.Error(w, "Failed to create stock item: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/inventory", http.StatusSeeOther)
}

func (api API) inventoryOutwardForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Inventory Outward Form - Not Implemented Yet")
}

func (api API) inventoryOutwardCreate(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Inventory Outward Create - Not Implemented Yet")
}

// --- Settings Handler ---

func (api API) settingsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.SettingsPage(org, "").Render(ctx, w); err != nil {
		api.logger.Error("render settings page", "error", err)
	}
}
