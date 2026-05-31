package httpapi

import (
	"net/http"

	"loombloom/internal/views"
)

func (api API) dashboard(w http.ResponseWriter, r *http.Request) {
	data, err := api.store.Dashboard(r.Context())
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
