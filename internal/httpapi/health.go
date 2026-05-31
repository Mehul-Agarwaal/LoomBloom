package httpapi

import (
	"context"
	"net/http"
	"time"
)

func (api API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api API) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := api.store.DB().PingContext(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (api API) modules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]string{
		"modules": {
			"organizations",
			"subscriptions",
			"machines",
			"workers",
			"stock",
			"spare_parts",
			"raw_materials",
			"payments",
			"production",
		},
	})
}
