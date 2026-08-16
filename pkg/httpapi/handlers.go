package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"loombloom/pkg/store"
)

func (api API) listOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := api.store.ListOrganizations(r.Context())
	if err != nil {
		api.logger.Error("list organizations", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

func (api API) createOrganization(w http.ResponseWriter, r *http.Request) {
	var org store.Organization
	if err := decodeJSON(r, &org); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := required(org.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if org.SubscriptionPlan == "" {
		org.SubscriptionPlan = "starter"
	}

	created, err := api.store.CreateOrganization(r.Context(), org)
	if err != nil {
		api.logger.Error("create organization", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api API) listMachines(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	machines, err := api.store.ListMachines(r.Context(), organizationID)
	if err != nil {
		api.logger.Error("list machines", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, machines)
}

func (api API) createMachine(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	var machine store.Machine
	if err := decodeJSON(r, &machine); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := required(machine.MachineNumber, "machine_number"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if machine.Status == "" {
		machine.Status = "active"
	}
	machine.OrganizationID = organizationID

	created, err := api.store.CreateMachine(r.Context(), machine)
	if err != nil {
		api.logger.Error("create machine", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api API) listWorkers(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	workers, err := api.store.ListWorkers(r.Context(), organizationID)
	if err != nil {
		api.logger.Error("list workers", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, workers)
}

func (api API) createWorker(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	var worker store.Worker
	if err := decodeJSON(r, &worker); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := required(worker.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	worker.OrganizationID = organizationID

	created, err := api.store.CreateWorker(r.Context(), worker)
	if err != nil {
		api.logger.Error("create worker", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api API) listStockItems(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	items, err := api.store.ListStockItems(r.Context(), organizationID)
	if err != nil {
		api.logger.Error("list stock items", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (api API) createStockItem(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	var item store.StockItem
	if err := decodeJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := required(item.Name, "name"); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if item.Unit == "" {
		item.Unit = "meters"
	}
	item.OrganizationID = organizationID

	created, err := api.store.CreateStockItem(r.Context(), item)
	if err != nil {
		api.logger.Error("create stock item", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api API) listProductionEntries(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	entries, err := api.store.ListProductionEntries(r.Context(), organizationID)
	if err != nil {
		api.logger.Error("list production entries", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (api API) createProductionEntry(w http.ResponseWriter, r *http.Request) {
	organizationID := r.PathValue("organizationID")
	if ok := api.ensureOrganization(w, r, organizationID); !ok {
		return
	}

	var entry store.ProductionEntry
	if err := decodeJSON(r, &entry); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if entry.MachineID == "" {
		writeError(w, http.StatusBadRequest, errors.New("machine_id is required"))
		return
	}
	if entry.EntryDate.IsZero() {
		entry.EntryDate = time.Now().UTC()
	}
	if entry.Unit == "" {
		entry.Unit = "meters"
	}
	entry.OrganizationID = organizationID

	created, err := api.store.CreateProductionEntry(r.Context(), entry)
	if err != nil {
		api.logger.Error("create production entry", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (api API) ensureOrganization(w http.ResponseWriter, r *http.Request, organizationID string) bool {
	if organizationID == "" {
		writeError(w, http.StatusBadRequest, errors.New("organization id is required"))
		return false
	}

	exists, err := api.store.OrganizationExists(r.Context(), organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("organization not found"))
			return false
		}
		api.logger.Error("check organization", "error", err)
		writeError(w, http.StatusInternalServerError, err)
		return false
	}
	if !exists {
		writeError(w, http.StatusNotFound, errors.New("organization not found"))
		return false
	}
	return true
}
