package appointment

import (
	"net/http"
)

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, StatusConfirmed)
}
