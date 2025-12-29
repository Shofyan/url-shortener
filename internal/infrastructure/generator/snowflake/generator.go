package snowflake

import (
	"sync"

	"github.com/bwmarrin/snowflake"
)

// Generator implements the IDGenerator interface using Snowflake
type Generator struct {
	node *snowflake.Node
	mu   sync.Mutex
}

// NewGenerator creates a new Snowflake ID generator
func NewGenerator(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}

	return &Generator{
		node: node,
	}, nil
}

// Generate generates a unique Snowflake ID
func (g *Generator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := g.node.Generate()
	return id.Int64(), nil
}
