package handlers

import (
	"net/http"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/service"
)

func ListTracks(lib *service.Library) http.HandlerFunc {
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
