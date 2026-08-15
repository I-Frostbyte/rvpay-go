package webhook

import (
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HighLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := h.service.ReadBody(r)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	err = h.service.HandleWebhook(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return immediately after storing the event.
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
