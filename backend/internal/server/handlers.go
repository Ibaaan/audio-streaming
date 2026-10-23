package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"audio-streaming/internal/catalog"
	"audio-streaming/internal/pipeline"
	"audio-streaming/internal/service"
)

const (
	maxUploadSize = 1 << 30
	streamExpiry  = time.Hour
)

// apiError is the error body from docs/openapi.yaml.
type apiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, code int, reason, msg string) {
	var e apiError
	e.Error.Code, e.Error.Reason, e.Error.Message = code, reason, msg
	writeJSON(w, code, e)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encoding response", "error", err)
	}
}

func internalError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}

func ListTracks(repo catalog.TrackLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, err := repo.ListTracks(r.Context())
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

// uploadTrack takes a multipart form with the audio in its "file" field,
// copies it to a temporary file and runs the pipeline on it.
func UploadTrack(p *pipeline.Pipeline, tmpDir string) http.HandlerFunc {
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

		err = p.Run(r.Context(), tmp)
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
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

// receiveFile streams the "file" part of r's multipart body into a new
// file in dir and returns its path. The file keeps the uploaded name's
// extension, which becomes the extension of the stored blob.
func receiveFile(r *http.Request, dir string) (string, error) {
	mr, err := r.MultipartReader()
	if err != nil {
		return "", err
	}
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return "", errors.New(`missing "file" field`)
		}
		if err != nil {
			return "", err
		}
		if part.FormName() != "file" {
			continue
		}

		ext := strings.ToLower(filepath.Ext(path.Base(part.FileName())))
		f, err := os.CreateTemp(dir, "upload-*"+ext)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(f, part)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			_ = os.Remove(f.Name())
			return "", err
		}
		return f.Name(), nil
	}
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

// streamTrack returns a presigned URL for the track's audio. The browser
// streams from storage directly, with Range requests for seeking.
func StreamTrack(repo catalog.TrackRepository, blobs blobFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("track"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PARAMETER", "track must be a UUID")
			return
		}
		_, err = repo.GetTrack(r.Context(), id)
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
