package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

type artistGetter interface {
	GetArtist(context.Context, uuid.UUID, []string) (dto.Artist, error)
}

// GetArtist handles GET /v1/artists/{artist}, with optional albums and tracks.
func GetArtist(lib artistGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("artist"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "artist must be a UUID")
			return
		}
		include, err := artistInclude(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
			return
		}
		artist, err := lib.GetArtist(r.Context(), id, include)
		if errors.Is(err, catalog.ErrNotFound) {
			writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "no such artist")
			return
		}
		if err != nil {
			internalError(w, "getting artist", err)
			return
		}
		writeJSON(w, http.StatusOK, artist)
	}
}
