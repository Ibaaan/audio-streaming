package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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

func paginationValue(r *http.Request, name string, fallback int) (int, error) {
	values, ok := r.URL.Query()[name]
	if !ok {
		return fallback, nil
	}
	if len(values) != 1 {
		return 0, errors.New(name + " must be specified once")
	}
	value, err := strconv.Atoi(values[0])
	if err != nil || value < 0 {
		return 0, errors.New(name + " must be a non-negative integer")
	}
	return value, nil
}
