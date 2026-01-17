package service

import (
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// IDGenerator defines the interface for ID generation.
type IDGenerator interface {
	// Generate generates a unique ID
	Generate() (int64, error)
}

// ShortKeyGenerator defines the interface for short key generation.
type ShortKeyGenerator interface {
	// GenerateFromID generates a short key from an ID
	GenerateFromID(id int64) (*valueobject.ShortKey, error)

	// DecodeToID decodes a short key back to an ID
	DecodeToID(shortKey *valueobject.ShortKey) (int64, error)
}

// GeneratorService combines ID and short key generation.
type GeneratorService struct {
	idGenerator       IDGenerator
	shortKeyGenerator ShortKeyGenerator
}

// NewGeneratorService creates a new GeneratorService.
func NewGeneratorService(idGen IDGenerator, shortKeyGen ShortKeyGenerator) *GeneratorService {
	return &GeneratorService{
		idGenerator:       idGen,
		shortKeyGenerator: shortKeyGen,
	}
}

// GenerateShortKey generates a new short key.
func (s *GeneratorService) GenerateShortKey() (*valueobject.ShortKey, int64, error) {
	id, err := s.idGenerator.Generate()
	if err != nil {
		return nil, 0, err
	}

	shortKey, err := s.shortKeyGenerator.GenerateFromID(id)
	if err != nil {
		return nil, 0, err
	}

	return shortKey, id, nil
}

// GenerateID generates a new unique ID.
func (s *GeneratorService) GenerateID() (int64, error) {
	return s.idGenerator.Generate()
}
