package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

type albumGetter interface {
	GetAlbum(context.Context, uuid.UUID, []string) (dto.Album, error)
}

// GetAlbum handles GET /v1/albums/{album}, optionally including its tracklist.
func GetAlbum(lib albumGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("album"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "album must be a UUID")
			return
		}
		include, err := albumInclude(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
			return
		}
		album, err := lib.GetAlbum(r.Context(), id, include)
		if errors.Is(err, catalog.ErrNotFound) {
			writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "no such album")
			return
		}
		if err != nil {
			internalError(w, "getting album", err)
			return
		}
		writeJSON(w, http.StatusOK, album)
	}
}
