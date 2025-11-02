package tree

import (
	"fmt"
	"sort"

	"stellaris-research-tree/lib/models"
)

// TechNode represents a node in the technology tree
type TechNode struct {
	Tech         *models.Technology
	Dependencies []*TechNode
	Dependents   []*TechNode
	Level        int
	Visited      bool
}

// TechTree represents the complete technology dependency tree
type TechTree struct {
	nodes      map[string]*TechNode
	rootNodes  []*TechNode
	maxLevel   int
	byArea     map[string][]*TechNode
	byTier     map[int][]*TechNode
	byCategory map[string][]*TechNode
}

// NewTechTree creates a new technology tree from parsed technologies
func NewTechTree(technologies map[string]*models.Technology) *TechTree {
	tree := &TechTree{
		nodes:      make(map[string]*TechNode),
		rootNodes:  []*TechNode{},
		byArea:     make(map[string][]*TechNode),
		byTier:     make(map[int][]*TechNode),
		byCategory: make(map[string][]*TechNode),
	}

	// Create nodes for all technologies
	for key, tech := range technologies {
		node := &TechNode{
			Tech:         tech,
			Dependencies: []*TechNode{},
			Dependents:   []*TechNode{},
		}
		tree.nodes[key] = node
	}

	// Build dependencies
	for key, node := range tree.nodes {
		for _, prereqKey := range node.Tech.Prerequisites {
			if prereqNode, exists := tree.nodes[prereqKey]; exists {
				node.Dependencies = append(node.Dependencies, prereqNode)
				prereqNode.Dependents = append(prereqNode.Dependents, node)
			} else {
				fmt.Printf("Warning: technology '%s' has unknown prerequisite '%s'\n", key, prereqKey)
			}
		}
	}

	// Find root nodes (technologies with no prerequisites)
	for _, node := range tree.nodes {
		if len(node.Dependencies) == 0 {
			tree.rootNodes = append(tree.rootNodes, node)
		}
	}

	// Calculate levels
	tree.calculateLevels()

	// Organize by area, tier, and category
	tree.organizeByAttributes()

	return tree
}

// calculateLevels determines the level of each node in the tree
func (t *TechTree) calculateLevels() {
	// Reset all visited flags
	for _, node := range t.nodes {
		node.Visited = false
		node.Level = 0
	}

	// BFS to calculate levels
	queue := make([]*TechNode, len(t.rootNodes))
	copy(queue, t.rootNodes)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node.Visited {
			continue
		}
		node.Visited = true

		// Calculate level as max of all dependencies + 1
		maxDepLevel := -1
		allDepsVisited := true
		for _, dep := range node.Dependencies {
			if !dep.Visited {
				allDepsVisited = false
				break
			}
			if dep.Level > maxDepLevel {
				maxDepLevel = dep.Level
			}
		}

		if !allDepsVisited {
			// Re-queue if dependencies aren't all processed
			queue = append(queue, node)
			continue
		}

		node.Level = maxDepLevel + 1
		if node.Level > t.maxLevel {
			t.maxLevel = node.Level
		}

		// Add dependents to queue
		queue = append(queue, node.Dependents...)
	}
}

// organizeByAttributes organizes nodes by area, tier, and category
func (t *TechTree) organizeByAttributes() {
	for _, node := range t.nodes {
		// By area
		if node.Tech.Area != "" {
			t.byArea[node.Tech.Area] = append(t.byArea[node.Tech.Area], node)
		}

		// By tier
		t.byTier[node.Tech.Tier] = append(t.byTier[node.Tech.Tier], node)

		// By category
		for _, category := range node.Tech.Category {
			t.byCategory[category] = append(t.byCategory[category], node)
		}
	}
}

// GetRootNodes returns all root nodes (no prerequisites)
func (t *TechTree) GetRootNodes() []*TechNode {
	return t.rootNodes
}

// GetNode returns a specific node by technology key
func (t *TechTree) GetNode(key string) (*TechNode, bool) {
	node, exists := t.nodes[key]
	return node, exists
}

// GetAllNodes returns all nodes in the tree
func (t *TechTree) GetAllNodes() map[string]*TechNode {
	return t.nodes
}

// GetNodesByArea returns nodes filtered by research area
func (t *TechTree) GetNodesByArea(area string) []*TechNode {
	return t.byArea[area]
}

// GetNodesByTier returns nodes filtered by tier
func (t *TechTree) GetNodesByTier(tier int) []*TechNode {
	return t.byTier[tier]
}

// GetMaxLevel returns the maximum depth of the tree
func (t *TechTree) GetMaxLevel() int {
	return t.maxLevel
}

// GetAreas returns all unique research areas
func (t *TechTree) GetAreas() []string {
	areas := make([]string, 0, len(t.byArea))
	for area := range t.byArea {
		areas = append(areas, area)
	}
	sort.Strings(areas)
	return areas
}

// GetTiers returns all unique tiers
func (t *TechTree) GetTiers() []int {
	tiers := make([]int, 0, len(t.byTier))
	for tier := range t.byTier {
		tiers = append(tiers, tier)
	}
	sort.Ints(tiers)
	return tiers
}

// GetCategories returns all unique categories
func (t *TechTree) GetCategories() []string {
	categories := make([]string, 0, len(t.byCategory))
	for category := range t.byCategory {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories
}
