package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/willbon-dev/UniSub/internal/config"
	"github.com/willbon-dev/UniSub/internal/service"
)

func New(cfg *config.Config, svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/subscribe", handleSubscribe(svc))
	return mux
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleSubscribe(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only GET is supported")
			return
		}

		secret := strings.TrimSpace(r.URL.Query().Get("secret"))
		if secret == "" {
			writeError(w, http.StatusBadRequest, "missing_secret", "secret is required")
			return
		}

		refresh := r.URL.Query().Get("refresh") == "1"
		result, err := svc.RenderSubscription(r.Context(), secret, r.URL.Query().Get("platform"), refresh)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrNotFound):
				writeError(w, http.StatusNotFound, "not_found", err.Error())
			default:
				writeError(w, http.StatusBadGateway, "upstream_error", err.Error())
			}
			return
		}

		w.Header().Set("Content-Type", result.ContentType)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(result.Body)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
