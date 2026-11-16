package handlers

import (
	"context"
	"errors"
	"net/http"

	"audio-streaming/internal/dto"
	"audio-streaming/internal/service"
)

type searcher interface {
	Search(context.Context, dto.SearchRequest) (dto.SearchResponse, error)
}

// Search handles GET /v1/search.
func Search(lib searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := searchRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
			return
		}
		result, err := lib.Search(r.Context(), req)
		if errors.Is(err, service.ErrInvalidSearch) {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
			return
		}
		if err != nil {
			internalError(w, "searching catalog", err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}
