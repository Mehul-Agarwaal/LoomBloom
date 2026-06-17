package httpapi

import (
	"log/slog"
	"net/http"

	"loombloom/internal/store"
)

type API struct {
	store  store.Store
	logger *slog.Logger
}

func New(store store.Store, logger *slog.Logger) API {
	return API{store: store, logger: logger}
}

func (api API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", api.dashboard)
	mux.HandleFunc("GET /machines", api.machinesPage)
	mux.HandleFunc("POST /machines", api.machinesPage)
	mux.HandleFunc("GET /workers", api.workersPage)
	mux.HandleFunc("POST /workers", api.workersPage)
	mux.HandleFunc("GET /stock", api.stockPage)
	mux.HandleFunc("POST /stock", api.stockPage)
	mux.HandleFunc("GET /production", api.productionPage)
	mux.HandleFunc("POST /production", api.productionPage)

	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("GET /readyz", api.ready)
	mux.HandleFunc("GET /api/v1/modules", api.modules)

	mux.HandleFunc("GET /api/v1/organizations", api.listOrganizations)
	mux.HandleFunc("POST /api/v1/organizations", api.createOrganization)

	mux.HandleFunc("GET /api/v1/organizations/{organizationID}/machines", api.listMachines)
	mux.HandleFunc("POST /api/v1/organizations/{organizationID}/machines", api.createMachine)
	mux.HandleFunc("GET /api/v1/organizations/{organizationID}/workers", api.listWorkers)
	mux.HandleFunc("POST /api/v1/organizations/{organizationID}/workers", api.createWorker)
	mux.HandleFunc("GET /api/v1/organizations/{organizationID}/stock", api.listStockItems)
	mux.HandleFunc("POST /api/v1/organizations/{organizationID}/stock", api.createStockItem)
	mux.HandleFunc("GET /api/v1/organizations/{organizationID}/production", api.listProductionEntries)
	mux.HandleFunc("POST /api/v1/organizations/{organizationID}/production", api.createProductionEntry)

	return api.withMiddleware(mux)
}

func (api API) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
