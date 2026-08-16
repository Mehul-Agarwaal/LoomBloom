package httpapi

import (
	"log/slog"
	"net/http"

	"loombloom/pkg/store"
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

	// Dashboard
	mux.HandleFunc("GET /", api.dashboard)

	// Production
	mux.HandleFunc("GET /production", api.productionList)
	mux.HandleFunc("GET /production/new", api.productionForm)
	mux.HandleFunc("POST /production/new", api.productionCreate)

	// Machines
	mux.HandleFunc("GET /machines", api.machinesList)
	mux.HandleFunc("GET /machines/new", api.machineForm)
	mux.HandleFunc("POST /machines", api.machineCreate)
	mux.HandleFunc("GET /machines/{id}", api.machineDetail)

	// Workers
	mux.HandleFunc("GET /workers", api.workersList)
	mux.HandleFunc("GET /workers/new", api.workerForm)
	mux.HandleFunc("POST /workers", api.workerCreate)
	mux.HandleFunc("GET /workers/{id}", api.workerDetail)

	// Inventory
	mux.HandleFunc("GET /inventory", api.inventoryList)
	mux.HandleFunc("GET /inventory/inward", api.inventoryInwardForm)
	mux.HandleFunc("POST /inventory/inward", api.inventoryInwardCreate)
	mux.HandleFunc("GET /inventory/outward", api.inventoryOutwardForm)
	mux.HandleFunc("POST /inventory/outward", api.inventoryOutwardCreate)

	// Settings
	mux.HandleFunc("GET /settings", api.settingsPage)

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

	mux.HandleFunc("GET /login", api.loginPageHandler)
	mux.HandleFunc("POST /login", api.loginPageHandler)
	mux.HandleFunc("GET /setup", api.setupPageHandler)
	mux.HandleFunc("POST /setup", api.setupPageHandler)
	mux.HandleFunc("GET /verify", api.verifyPageHandler)
	mux.HandleFunc("POST /verify", api.verifyPageHandler)
	mux.HandleFunc("POST /verify/resend", api.verifyResendHandler)
	mux.HandleFunc("GET /plans", api.plansPageHandler)
	mux.HandleFunc("POST /checkout", api.checkoutHandler)
	mux.HandleFunc("GET /payment/checkout", api.mockCheckoutPageHandler)
	mux.HandleFunc("POST /payment/callback", api.mockPaymentCallbackHandler)
	mux.HandleFunc("GET /logout", api.logoutHandler)

	return api.withMiddleware(api.OnboardingMiddleware(mux))
}

func (api API) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
