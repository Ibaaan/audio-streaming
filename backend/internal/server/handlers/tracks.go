package handlers

import (
	"context"
	"net/http"

	"audio-streaming/internal/catalog"
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
