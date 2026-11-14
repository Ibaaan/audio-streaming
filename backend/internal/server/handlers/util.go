package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// apiError is the error body from docs/openapi.yaml.
type apiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Reason  string `json:"reason"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encoding response", "error", err)
	}
}

func writeError(w http.ResponseWriter, code int, reason, msg string) {
	var e apiError
	e.Error.Code, e.Error.Reason, e.Error.Message = code, reason, msg
	writeJSON(w, code, e)
}

func internalError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
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
