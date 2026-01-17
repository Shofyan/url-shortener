// Package snowflake provides Snowflake ID generation for distributed URL shortener instances.
//
// Snowflake IDs are 64-bit integers that guarantee uniqueness across distributed
// systems without requiring coordination between nodes. Each ID contains
// timestamp, machine identifier, and sequence information, making them
// naturally ordered and collision-free.
//
// Snowflake ID structure:
//   - 1 bit: Reserved (always 0)
//   - 41 bits: Timestamp (milliseconds since custom epoch)
//   - 10 bits: Machine/Node ID
//   - 12 bits: Sequence number
//
// Features:
//   - Guaranteed uniqueness across distributed instances
//   - Time-ordered IDs for natural sorting
//   - High throughput (up to 4096 IDs per millisecond per node)
//   - No network coordination required
//   - Configurable node ID and epoch settings
//   - Built-in clock drift protection
//
// This generator is ideal for distributed deployments where multiple
// URL shortener instances need to generate unique IDs without conflicts.
package snowflake
