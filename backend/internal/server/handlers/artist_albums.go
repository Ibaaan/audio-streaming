package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/dto"
)

type artistAlbumLister interface {
	ListArtistAlbums(context.Context, uuid.UUID, int, int) (dto.ListAlbumsResponse, error)
}

// ListArtistAlbums handles GET /v1/artists/{artist}/albums.
func ListArtistAlbums(lib artistAlbumLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("artist"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "artist must be a UUID")
			return
		}
		pageSize, err := paginationValue(r, "page_size", 20)
		if err != nil {
			writeError(w, 400, "INVALID_PARAMETER", err.Error())
			return
		}
		offset, err := paginationValue(r, "offset", 0)
		if err != nil {
			writeError(w, 400, "INVALID_PARAMETER", err.Error())
			return
		}
		if pageSize == 0 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}
		page, err := lib.ListArtistAlbums(r.Context(), id, pageSize, offset)
		if errors.Is(err, catalog.ErrNotFound) {
			writeError(w, 404, "RESOURCE_NOT_FOUND", "no such artist")
			return
		}
		if err != nil {
			internalError(w, "listing artist albums", err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	}
}
