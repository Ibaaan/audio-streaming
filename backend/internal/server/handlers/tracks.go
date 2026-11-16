package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/service"
)

type trackLister interface {
	ListTracks(ctx context.Context) ([]catalog.TrackSummary, error)
}

func ListTracks(lib trackLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, err := lib.ListTracks(r.Context())
		if err != nil {
			internalError(w, "listing tracks", err)
			return
		}
		if tracks == nil {
			tracks = []catalog.TrackSummary{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
	}
}

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
