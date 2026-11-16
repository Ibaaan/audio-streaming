package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/service"
)

func GetTrack(lib *service.Library) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("trackId"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "trackId must be a UUID")
			return
		}

		track, err := lib.GetTrack(r.Context(), id)
		if errors.Is(err, catalog.ErrNotFound) {
			writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "no such track")
			return
		}
		if err != nil {
			internalError(w, "getting track", err)
			return
		}

		writeJSON(w, http.StatusOK, track)
	}
}
