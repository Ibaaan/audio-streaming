package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"

	"audio-streaming/internal/dto"

	"audio-streaming/internal/pipeline"
	"audio-streaming/internal/service"
)

const maxUploadSize = 1 << 30

type trackImporter interface {
	Run(ctx context.Context, path string) (uuid.UUID, error)
}

// UploadTrack takes a multipart form with the audio in its "file" field,
// copies it to a temporary file and runs the pipeline on it.
func UploadTrack(p trackImporter, tmpDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		tmp, err := receiveFile(r, tmpDir)
		if err != nil {
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				writeError(w, http.StatusRequestEntityTooLarge,
					"INVALID_PARAMETER", "file is larger than 1 GiB")
				return
			}
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
			return
		}
		defer func() { _ = os.Remove(tmp) }()

		id, err := p.Run(r.Context(), tmp)
		switch {
		case err == nil:
			w.Header().Set("Location", "/v1/tracks/"+id.String())
			writeJSON(w, http.StatusCreated, dto.UploadTrackResponse{ID: id})
		case errors.Is(err, pipeline.ErrInvalidSource):
			slog.Info("rejected upload", "error", err)
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "not a readable audio file")
		case errors.Is(err, service.ErrInvalidRecord):
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
		case errors.Is(err, service.ErrTrackExists):
			writeError(w, http.StatusConflict, "ALREADY_EXISTS", err.Error())
		default:
			internalError(w, "importing track", err)
		}
	}
}
