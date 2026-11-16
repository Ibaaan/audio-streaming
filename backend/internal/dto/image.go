package dto

import "github.com/google/uuid"

// Image contains artwork dimensions and URLs for each documented rendition.
type Image struct {
	ID       uuid.UUID     `json:"id"`
	Width    *int          `json:"width"`
	Height   *int          `json:"height"`
	Variants ImageVariants `json:"variants"`
}

type ImageVariants struct {
	Size300  string `json:"300"`
	Size640  string `json:"640"`
	Size1200 string `json:"1200"`
}
