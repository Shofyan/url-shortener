package base62

import (
	"log"
	"math"
	"strings"

	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generator implements the ShortKeyGenerator interface using Base62 encoding
type Generator struct{}

// NewGenerator creates a new Base62 generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateFromID converts an ID to a Base62 encoded short key
func (g *Generator) GenerateFromID(id int64) (*valueobject.ShortKey, error) {
	if id == 0 {
		return valueobject.NewShortKey("0")
	}

	encoded := g.encode(id)
	log.Printf("[Base62] Generated ID: %d, Encoded: %s, Length: %d", id, encoded, len(encoded))

	shortKey, err := valueobject.NewShortKey(encoded)
	if err != nil {
		log.Printf("[Base62] Error creating short key from encoded string '%s': %v", encoded, err)
		return nil, err
	}

	return shortKey, nil
}

// DecodeToID decodes a Base62 short key back to an ID
func (g *Generator) DecodeToID(shortKey *valueobject.ShortKey) (int64, error) {
	return g.decode(shortKey.Value()), nil
}

// encode converts a number to Base62
func (g *Generator) encode(num int64) string {
	if num == 0 {
		return "0"
	}

	var result strings.Builder
	base := int64(len(base62Chars))

	for num > 0 {
		remainder := num % base
		result.WriteByte(base62Chars[remainder])
		num = num / base
	}

	// Reverse the string
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// decode converts a Base62 string to a number
func (g *Generator) decode(encoded string) int64 {
	var num int64
	base := int64(len(base62Chars))
	length := len(encoded)

	for i, char := range encoded {
		power := length - i - 1
		index := strings.IndexRune(base62Chars, char)
		num += int64(index) * int64(math.Pow(float64(base), float64(power)))
	}

	return num
}
