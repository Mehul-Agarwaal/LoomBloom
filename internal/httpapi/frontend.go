package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"loombloom/internal/store"
	"loombloom/internal/views"
)

func (api API) dashboard(w http.ResponseWriter, r *http.Request) {
	org, ok := GetOrg(r.Context())
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
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

// --- Machines Page ---

func (api API) machinesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	var errStr string
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			errStr = "Failed to parse form: " + err.Error()
		} else {
			machineNum := r.FormValue("machine_number")
			size := r.FormValue("size")
			machineType := r.FormValue("type")
			status := r.FormValue("status")

			if machineNum == "" {
				errStr = "Machine number is required."
			} else {
				_, err := api.store.CreateMachine(ctx, store.Machine{
					OrganizationID: org.ID,
					MachineNumber:  machineNum,
					Size:           size,
					Type:           machineType,
					Status:         status,
				})
				if err != nil {
					api.logger.Error("create machine from web", "error", err)
					errStr = "Failed to create machine: " + err.Error()
				} else {
					http.Redirect(w, r, "/machines", http.StatusSeeOther)
					return
				}
			}
		}
	}

	machines, err := api.store.ListMachines(ctx, org.ID)
	if err != nil {
		api.logger.Error("list machines", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.MachinesPage(org, machines, errStr).Render(ctx, w); err != nil {
		api.logger.Error("render machines page", "error", err)
	}
}

// --- Workers Page ---

func (api API) workersPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	var errStr string
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			errStr = "Failed to parse form: " + err.Error()
		} else {
			name := r.FormValue("name")
			phone := r.FormValue("phone")
			avgProdStr := r.FormValue("average_production")
			closesStr := r.FormValue("close_per_shift")
			shiftsStr := r.FormValue("number_of_shifts")

			avgProd, _ := strconv.ParseFloat(avgProdStr, 64)
			closes, _ := strconv.Atoi(closesStr)
			shifts, _ := strconv.Atoi(shiftsStr)

			if name == "" {
				errStr = "Worker name is required."
			} else {
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
					errStr = "Failed to create worker: " + err.Error()
				} else {
					http.Redirect(w, r, "/workers", http.StatusSeeOther)
					return
				}
			}
		}
	}

	workers, err := api.store.ListWorkers(ctx, org.ID)
	if err != nil {
		api.logger.Error("list workers", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.WorkersPage(org, workers, errStr).Render(ctx, w); err != nil {
		api.logger.Error("render workers page", "error", err)
	}
}

// --- Stock Page ---

func (api API) stockPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	var errStr string
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			errStr = "Failed to parse form: " + err.Error()
		} else {
			name := r.FormValue("name")
			quality := r.FormValue("quality")
			storedOn := r.FormValue("stored_on")
			qtyStr := r.FormValue("quantity")
			unit := r.FormValue("unit")

			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				errStr = "Quantity must be a valid number."
			} else if name == "" {
				errStr = "Stock item name is required."
			} else {
				if unit == "" {
					unit = "meters"
				}
				_, err := api.store.CreateStockItem(ctx, store.StockItem{
					OrganizationID: org.ID,
					Name:           name,
					Quality:        quality,
					StoredOn:       storedOn,
					Quantity:       qty,
					Unit:           unit,
				})
				if err != nil {
					api.logger.Error("create stock item from web", "error", err)
					errStr = "Failed to create stock item: " + err.Error()
				} else {
					http.Redirect(w, r, "/stock", http.StatusSeeOther)
					return
				}
			}
		}
	}

	stock, err := api.store.ListStockItems(ctx, org.ID)
	if err != nil {
		api.logger.Error("list stock items", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.StockPage(org, stock, errStr).Render(ctx, w); err != nil {
		api.logger.Error("render stock page", "error", err)
	}
}

// --- Production Page ---

func (api API) productionPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	org, ok := GetOrg(ctx)
	if !ok {
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	var errStr string
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			errStr = "Failed to parse form: " + err.Error()
		} else {
			machineID := r.FormValue("machine_id")
			workerID := r.FormValue("worker_id")
			dateStr := r.FormValue("entry_date")
			shift := r.FormValue("shift")
			qtyStr := r.FormValue("quantity")
			unit := r.FormValue("unit")

			qty, err := strconv.ParseFloat(qtyStr, 64)
			if err != nil {
				errStr = "Quantity must be a valid number."
			} else if machineID == "" {
				errStr = "Loom / Machine is required."
			} else {
				entryDate := time.Now().UTC()
				if dateStr != "" {
					if parsedDate, err := time.Parse("2006-01-02", dateStr); err == nil {
						entryDate = parsedDate
					}
				}

				if unit == "" {
					unit = "meters"
				}

				_, err := api.store.CreateProductionEntry(ctx, store.ProductionEntry{
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
					errStr = "Failed to create production entry: " + err.Error()
				} else {
					http.Redirect(w, r, "/production", http.StatusSeeOther)
					return
				}
			}
		}
	}

	entries, err := api.store.ListProductionEntries(ctx, org.ID)
	if err != nil {
		api.logger.Error("list production entries", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	machines, err := api.store.ListMachines(ctx, org.ID)
	if err != nil {
		api.logger.Error("list machines for production page", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	workers, err := api.store.ListWorkers(ctx, org.ID)
	if err != nil {
		api.logger.Error("list workers for production page", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.ProductionPage(org, entries, machines, workers, errStr).Render(ctx, w); err != nil {
		api.logger.Error("render production page", "error", err)
	}
}
