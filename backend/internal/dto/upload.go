package dto

import "github.com/google/uuid"

// UploadTrackResponse identifies a successfully imported and saved track.
type UploadTrackResponse struct {
	ID uuid.UUID `json:"id"`
}
