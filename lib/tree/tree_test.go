package tree

import (
	"testing"

	"stellaris-data-parser/lib/models"
)

func createTestTechnologies() map[string]*models.Technology {
	return map[string]*models.Technology{
		"tech_root_1": {
			Key:           "tech_root_1",
			Cost:          0,
			Area:          "physics",
			Tier:          0,
			Category:      []string{"computing"},
			Prerequisites: []string{},
			IsStartTech:   true,
		},
		"tech_root_2": {
			Key:           "tech_root_2",
			Cost:          0,
			Area:          "society",
			Tier:          0,
			Category:      []string{"biology"},
			Prerequisites: []string{},
			IsStartTech:   true,
		},
		"tech_level_1": {
			Key:           "tech_level_1",
			Cost:          1000,
			Area:          "physics",
			Tier:          1,
			Category:      []string{"computing"},
			Prerequisites: []string{"tech_root_1"},
		},
		"tech_level_2": {
			Key:           "tech_level_2",
			Cost:          2000,
			Area:          "physics",
			Tier:          2,
			Category:      []string{"computing", "materials"},
			Prerequisites: []string{"tech_level_1"},
		},
		"tech_multi_prereq": {
			Key:           "tech_multi_prereq",
			Cost:          3000,
			Area:          "engineering",
			Tier:          2,
			Category:      []string{"voidcraft"},
			Prerequisites: []string{"tech_level_1", "tech_root_2"},
		},
		"tech_rare": {
			Key:           "tech_rare",
			Cost:          5000,
			Area:          "physics",
			Tier:          3,
			Category:      []string{"particles"},
			Prerequisites: []string{"tech_level_2"},
			IsRare:        true,
		},
		"tech_dangerous": {
			Key:           "tech_dangerous",
			Cost:          10000,
			Area:          "physics",
			Tier:          4,
			Category:      []string{"particles"},
			Prerequisites: []string{"tech_rare"},
			IsDangerous:   true,
		},
	}
}

func TestNewTechTree(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	if tree == nil {
		t.Fatal("Expected tree to be created, got nil")
	}

	if len(tree.GetAllNodes()) != len(technologies) {
		t.Errorf("Expected %d nodes, got %d", len(technologies), len(tree.GetAllNodes()))
	}
}

func TestGetRootNodes(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	rootNodes := tree.GetRootNodes()

	if len(rootNodes) != 2 {
		t.Errorf("Expected 2 root nodes, got %d", len(rootNodes))
	}

	// Verify root nodes have no dependencies
	for _, node := range rootNodes {
		if len(node.Dependencies) != 0 {
			t.Errorf("Root node '%s' should have no dependencies, got %d", node.Tech.Key, len(node.Dependencies))
		}
	}
}

func TestCalculateLevels(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	// Check level calculations
	tests := []struct {
		key           string
		expectedLevel int
	}{
		{"tech_root_1", 0},
		{"tech_root_2", 0},
		{"tech_level_1", 1},
		{"tech_level_2", 2},
		{"tech_multi_prereq", 2}, // max(level of tech_level_1, level of tech_root_2) + 1 = max(1, 0) + 1 = 2
		{"tech_rare", 3},
		{"tech_dangerous", 4},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			node, exists := tree.GetNode(tt.key)
			if !exists {
				t.Fatalf("Expected to find node '%s'", tt.key)
			}

			if node.Level != tt.expectedLevel {
				t.Errorf("Expected level %d for '%s', got %d", tt.expectedLevel, tt.key, node.Level)
			}
		})
	}
}

func TestGetMaxLevel(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	maxLevel := tree.GetMaxLevel()

	if maxLevel != 4 {
		t.Errorf("Expected max level 4, got %d", maxLevel)
	}
}

func TestGetNode(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	// Test existing node
	node, exists := tree.GetNode("tech_level_1")
	if !exists {
		t.Error("Expected to find tech_level_1")
	}
	if node.Tech.Key != "tech_level_1" {
		t.Errorf("Expected key 'tech_level_1', got '%s'", node.Tech.Key)
	}

	// Test non-existing node
	_, exists = tree.GetNode("tech_nonexistent")
	if exists {
		t.Error("Expected tech_nonexistent to not exist")
	}
}

func TestDependencies(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	// Test single dependency
	node, _ := tree.GetNode("tech_level_1")
	if len(node.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(node.Dependencies))
	}
	if node.Dependencies[0].Tech.Key != "tech_root_1" {
		t.Errorf("Expected dependency 'tech_root_1', got '%s'", node.Dependencies[0].Tech.Key)
	}

	// Test multiple dependencies
	nodeMulti, _ := tree.GetNode("tech_multi_prereq")
	if len(nodeMulti.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(nodeMulti.Dependencies))
	}
}

func TestDependents(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	// Test that tech_level_1 has 2 dependents (tech_level_2 and tech_multi_prereq)
	node, _ := tree.GetNode("tech_level_1")
	if len(node.Dependents) != 2 {
		t.Errorf("Expected 2 dependents, got %d", len(node.Dependents))
	}

	// Test that tech_dangerous has no dependents
	nodeDangerous, _ := tree.GetNode("tech_dangerous")
	if len(nodeDangerous.Dependents) != 0 {
		t.Errorf("Expected 0 dependents for tech_dangerous, got %d", len(nodeDangerous.Dependents))
	}
}

func TestGetAreas(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	areas := tree.GetAreas()

	if len(areas) != 3 {
		t.Errorf("Expected 3 areas, got %d", len(areas))
	}

	expectedAreas := map[string]bool{
		"physics":     true,
		"society":     true,
		"engineering": true,
	}

	for _, area := range areas {
		if !expectedAreas[area] {
			t.Errorf("Unexpected area '%s'", area)
		}
	}
}

func TestGetTiers(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	tiers := tree.GetTiers()

	if len(tiers) != 5 {
		t.Errorf("Expected 5 tiers (0-4), got %d", len(tiers))
	}

	// Tiers should be sorted
	for i := 0; i < len(tiers)-1; i++ {
		if tiers[i] >= tiers[i+1] {
			t.Error("Expected tiers to be sorted in ascending order")
		}
	}
}

func TestGetCategories(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	categories := tree.GetCategories()

	if len(categories) == 0 {
		t.Error("Expected categories to be found")
	}

	expectedCategories := map[string]bool{
		"computing":  true,
		"biology":    true,
		"materials":  true,
		"voidcraft":  true,
		"particles":  true,
	}

	for _, category := range categories {
		if !expectedCategories[category] {
			t.Errorf("Unexpected category '%s'", category)
		}
	}
}

func TestGetNodesByArea(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	physicsNodes := tree.GetNodesByArea("physics")

	if len(physicsNodes) == 0 {
		t.Error("Expected physics nodes to be found")
	}

	// Verify all nodes are physics
	for _, node := range physicsNodes {
		if node.Tech.Area != "physics" {
			t.Errorf("Expected area 'physics', got '%s'", node.Tech.Area)
		}
	}
}

func TestGetNodesByTier(t *testing.T) {
	technologies := createTestTechnologies()
	tree := NewTechTree(technologies)

	tier0Nodes := tree.GetNodesByTier(0)

	if len(tier0Nodes) != 2 {
		t.Errorf("Expected 2 tier 0 nodes, got %d", len(tier0Nodes))
	}

	// Verify all nodes are tier 0
	for _, node := range tier0Nodes {
		if node.Tech.Tier != 0 {
			t.Errorf("Expected tier 0, got %d", node.Tech.Tier)
		}
	}
}

func TestUnknownPrerequisite(t *testing.T) {
	technologies := map[string]*models.Technology{
		"tech_with_missing_prereq": {
			Key:           "tech_with_missing_prereq",
			Cost:          1000,
			Area:          "physics",
			Tier:          1,
			Prerequisites: []string{"tech_nonexistent"},
		},
	}

	// This should not panic, but print a warning
	tree := NewTechTree(technologies)

	node, _ := tree.GetNode("tech_with_missing_prereq")
	if len(node.Dependencies) != 0 {
		t.Errorf("Expected 0 dependencies (missing prereq), got %d", len(node.Dependencies))
	}
}

func TestEmptyTechTree(t *testing.T) {
	technologies := make(map[string]*models.Technology)
	tree := NewTechTree(technologies)

	if tree == nil {
		t.Fatal("Expected tree to be created, got nil")
	}

	if len(tree.GetAllNodes()) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(tree.GetAllNodes()))
	}

	if tree.GetMaxLevel() != 0 {
		t.Errorf("Expected max level 0, got %d", tree.GetMaxLevel())
	}

	if len(tree.GetRootNodes()) != 0 {
		t.Errorf("Expected 0 root nodes, got %d", len(tree.GetRootNodes()))
	}
}

func TestComplexDependencyChain(t *testing.T) {
	technologies := map[string]*models.Technology{
		"tech_a": {
			Key:           "tech_a",
			Prerequisites: []string{},
		},
		"tech_b": {
			Key:           "tech_b",
			Prerequisites: []string{"tech_a"},
		},
		"tech_c": {
			Key:           "tech_c",
			Prerequisites: []string{"tech_a"},
		},
		"tech_d": {
			Key:           "tech_d",
			Prerequisites: []string{"tech_b", "tech_c"},
		},
		"tech_e": {
			Key:           "tech_e",
			Prerequisites: []string{"tech_d"},
		},
	}

	tree := NewTechTree(technologies)

	// tech_e should be at level 3 (tech_a=0, tech_b/c=1, tech_d=2, tech_e=3)
	node, _ := tree.GetNode("tech_e")
	if node.Level != 3 {
		t.Errorf("Expected level 3 for tech_e, got %d", node.Level)
	}

	// tech_d should have 2 dependencies
	nodeD, _ := tree.GetNode("tech_d")
	if len(nodeD.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies for tech_d, got %d", len(nodeD.Dependencies))
	}
}
