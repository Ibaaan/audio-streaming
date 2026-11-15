package handlers

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const streamExpiry = time.Hour

type trackChecker interface {
	HasTrack(ctx context.Context, id uuid.UUID) (bool, error)
}

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
func StreamTrack(lib trackChecker, blobs blobFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("track"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "track must be a UUID")
			return
		}
		present, err := lib.HasTrack(r.Context(), id)
		if err != nil {
			internalError(w, "checking track", err)
			return
		}
		if !present {
			writeError(w, http.StatusNotFound, "RESOURCE_NOT_FOUND", "no such track")
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

// audioKey finds the blob the pipeline stored for id. Its extension is
// that of the uploaded file, so it is looked up by prefix.
func audioKey(ctx context.Context, blobs blobFinder, id uuid.UUID) (string, error) {
	prefix := "audio/" + id.String()
	keys, err := blobs.List(ctx, prefix)
	if err != nil {
		return "", err
	}
	for _, k := range keys {
		if strings.TrimSuffix(k, path.Ext(k)) == prefix {
			return k, nil
		}
	}
	return "", fmt.Errorf("no blob for track %s", id)
}
