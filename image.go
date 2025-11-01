package imgx

import (
	"image"
	"time"
)

// Image represents an image with processing metadata
type Image struct {
	data     *image.NRGBA
	metadata *ProcessingMetadata
}

// ProcessingMetadata contains information about image processing operations
type ProcessingMetadata struct {
	SourcePath  string
	Operations  []OperationRecord
	Software    string // Fixed: "imgx"
	Version     string // Fixed: version from load.go
	Author      string // Customizable: artist/creator name
	ProjectURL  string // Fixed: project URL
	AddMetadata bool
}

// OperationRecord represents a single image processing operation
type OperationRecord struct {
	Action     string
	Parameters string
	Timestamp  time.Time
}

// Clone creates a deep copy of ProcessingMetadata
func (m *ProcessingMetadata) Clone() *ProcessingMetadata {
	ops := make([]OperationRecord, len(m.Operations))
	copy(ops, m.Operations)
	return &ProcessingMetadata{
		SourcePath:  m.SourcePath,
		Operations:  ops,
		Software:    m.Software,
		Version:     m.Version,
		Author:      m.Author,
		ProjectURL:  m.ProjectURL,
		AddMetadata: m.AddMetadata,
	}
}

// AddOperation adds a new operation record to the metadata
func (m *ProcessingMetadata) AddOperation(action, parameters string) {
	m.Operations = append(m.Operations, OperationRecord{
		Action:     action,
		Parameters: parameters,
		Timestamp:  time.Now(),
	})
}

// ToNRGBA returns the underlying NRGBA image data
func (img *Image) ToNRGBA() *image.NRGBA {
	return img.data
}

// Bounds returns the bounds of the image
func (img *Image) Bounds() image.Rectangle {
	return img.data.Bounds()
}

// GetMetadata returns the processing metadata
func (img *Image) GetMetadata() *ProcessingMetadata {
	return img.metadata
}

// SetAuthor sets the artist/creator name for the image metadata
// This overrides the default author but keeps creator_tool unchanged
func (img *Image) SetAuthor(author string) *Image {
	img.metadata.Author = author
	return img
}
