// Package base62 provides Base62 encoding for generating URL short codes.
//
// Base62 encoding uses a character set of [0-9A-Za-z] to create compact,
// URL-safe identifiers from numeric values. This implementation generates
// short keys that are both human-readable and suitable for use in URLs
// without requiring additional encoding.
//
// Features:
//   - Bijective Base62 encoding and decoding
//   - Configurable minimum length for generated codes
//   - Character set optimization for URL safety
//   - High-performance encoding algorithms
//   - Collision-resistant code generation
//   - Customizable alphabet for specific requirements
//
// The Base62 generator is ideal for:
//   - Human-readable short codes
//   - Sequential ID encoding
//   - Compact representation of large numbers
//   - URL-safe identifier generation
//
// This implementation ensures generated codes are consistent,
// deterministic, and suitable for web-based URL shortening services.
package base62
