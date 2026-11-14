package handlers

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/service"
)

const streamExpiry = time.Hour

type blobFinder interface {
	List(ctx context.Context, prefix string) ([]string, error)
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type streamURL struct {
	URL         string    `json:"url"`
	ContentType string    `json:"content_type"`
	ExpireTime  time.Time `json:"expire_time"`
}

// StreamTrack returns a presigned URL for the track's audio. The browser
// streams from storage directly, with Range requests for seeking.
func StreamTrack(lib *service.Library, blobs blobFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("track"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "track must be a UUID")
			return
		}
		_, err = lib.GetTrack(r.Context(), id)
		if errors.Is(err, catalog.ErrNotFound) {
			writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "no such track")
			return
		}
		if err != nil {
			internalError(w, "getting track", err)
			return
		}

		key, err := audioKey(r.Context(), blobs, id)
		if err != nil {
			internalError(w, "finding audio", err)
			return
		}
		url, err := blobs.PresignedGetURL(r.Context(), key, streamExpiry)
		if err != nil {
			internalError(w, "presigning audio URL", err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, streamURL{
			URL:         url,
			ContentType: mime.TypeByExtension(path.Ext(key)),
			ExpireTime:  time.Now().Add(streamExpiry).UTC(),
		})
	}
}
