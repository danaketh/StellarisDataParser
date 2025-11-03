package generator

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"stellaris-data-parser/lib/models"
	"stellaris-data-parser/lib/tree"
)

func createTestTree() *tree.TechTree {
	technologies := map[string]*models.Technology{
		"tech_test_1": {
			Key:           "tech_test_1",
			Cost:          0,
			Area:          "physics",
			Tier:          0,
			Category:      []string{"computing"},
			Prerequisites: []string{},
			Weight:        100,
			IsStartTech:   true,
		},
		"tech_test_2": {
			Key:           "tech_test_2",
			Cost:          1000,
			Area:          "physics",
			Tier:          1,
			Category:      []string{"materials"},
			Prerequisites: []string{"tech_test_1"},
			Weight:        85,
			IsRare:        true,
		},
		"tech_test_3": {
			Key:           "tech_test_3",
			Cost:          2000,
			Area:          "engineering",
			Tier:          2,
			Category:      []string{"voidcraft"},
			Prerequisites: []string{"tech_test_2"},
			Weight:        75,
			IsDangerous:   true,
		},
	}

	return tree.NewTechTree(technologies)
}

func TestNewJSONGenerator(t *testing.T) {
	testTree := createTestTree()
	generator := NewJSONGenerator(testTree)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.tree == nil {
		t.Error("Expected tree to be set")
	}
}

func TestGenerate(t *testing.T) {
	testTree := createTestTree()
	generator := NewJSONGenerator(testTree)

	// Create temp directory for output files
	tmpDir := t.TempDir()

	// Generate JSON files
	err := generator.Generate(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate JSON: %v", err)
	}

	// Verify JSON files were created
	metadataFile := tmpDir + "/metadata.json"
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Expected metadata.json to be created")
	}

	// Verify area-specific JSON files
	physicsFile := tmpDir + "/research-physics.json"
	if _, err := os.Stat(physicsFile); os.IsNotExist(err) {
		t.Error("Expected research-physics.json to be created")
	}

	engineeringFile := tmpDir + "/research-engineering.json"
	if _, err := os.Stat(engineeringFile); os.IsNotExist(err) {
		t.Error("Expected research-engineering.json to be created")
	}

	// Read and verify physics JSON file
	jsonContent, err := os.ReadFile(physicsFile)
	if err != nil {
		t.Fatalf("Failed to read physics JSON file: %v", err)
	}

	jsonStr := string(jsonContent)
	if !strings.Contains(jsonStr, "tech_test_1") {
		t.Error("Expected tech_test_1 in physics JSON data")
	}

	if !strings.Contains(jsonStr, "tech_test_2") {
		t.Error("Expected tech_test_2 in physics JSON data")
	}

	// Read and verify engineering JSON file
	engContent, err := os.ReadFile(engineeringFile)
	if err != nil {
		t.Fatalf("Failed to read engineering JSON file: %v", err)
	}

	engStr := string(engContent)
	if !strings.Contains(engStr, "tech_test_3") {
		t.Error("Expected tech_test_3 in engineering JSON data")
	}
}

func TestGenerateJSONFiles(t *testing.T) {
	testTree := createTestTree()
	generator := NewJSONGenerator(testTree)

	tmpDir := t.TempDir()

	err := generator.GenerateJSONFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate JSON files: %v", err)
	}

	// Check metadata file
	metadataContent, err := os.ReadFile(tmpDir + "/metadata.json")
	if err != nil {
		t.Fatalf("Failed to read metadata.json: %v", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataContent, &metadata); err != nil {
		t.Fatalf("Failed to parse metadata.json: %v", err)
	}

	// Check areas
	areas, ok := metadata["areas"].([]interface{})
	if !ok {
		t.Fatal("Expected areas to be array")
	}

	if len(areas) == 0 {
		t.Error("Expected areas to be populated")
	}

	// Check tiers
	tiers, ok := metadata["tiers"].([]interface{})
	if !ok {
		t.Fatal("Expected tiers to be array")
	}

	if len(tiers) == 0 {
		t.Error("Expected tiers to be populated")
	}

	// Check max level
	maxLevel, ok := metadata["maxLevel"].(float64)
	if !ok {
		t.Fatal("Expected maxLevel to be number")
	}

	if maxLevel < 0 {
		t.Errorf("Expected non-negative max level, got %f", maxLevel)
	}

	// Check technology area files exist
	if _, err := os.Stat(tmpDir + "/research-physics.json"); os.IsNotExist(err) {
		t.Error("Expected research-physics.json to be created")
	}
}

func TestFormatTechName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with tech_ prefix", "tech_basic_science", "Basic Science"},
		{"without prefix", "basic_science", "Basic Science"},
		{"multiple words", "tech_powered_exoskeletons", "Powered Exoskeletons"},
		{"single word", "tech_physics", "Physics"},
		{"already formatted", "Physics", "Physics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTechName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateWithComplexTech(t *testing.T) {
	technologies := map[string]*models.Technology{
		"tech_complex": {
			Key:           "tech_complex",
			Cost:          5000,
			Area:          "society",
			Tier:          3,
			Category:      []string{"psionics", "biology"},
			Prerequisites: []string{},
			Weight:        50,
			BaseWeight:    1.5,
			IsStartTech:   false,
			IsRare:        true,
			IsDangerous:   false,
			IsEvent:       true,
			IsReverse:     false,
			IsGestalt:     true,
			IsMegacorp:    false,
			FeatureUnlocks: []string{"feature_1", "feature_2"},
			WeightModifiers: []models.WeightModifier{
				{Factor: 2.0, Add: 100},
			},
		},
	}

	testTree := tree.NewTechTree(technologies)
	generator := NewJSONGenerator(testTree)

	tmpDir := t.TempDir()

	err := generator.Generate(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate JSON: %v", err)
	}

	// Verify society JSON file was created and contains complex properties
	jsonFile := tmpDir + "/research-society.json"
	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	jsonStr := string(jsonContent)

	// Verify complex properties are in the JSON
	if !strings.Contains(jsonStr, "isEvent") {
		t.Error("Expected isEvent property in JSON")
	}

	if !strings.Contains(jsonStr, "isGestalt") {
		t.Error("Expected isGestalt property in JSON")
	}

	if !strings.Contains(jsonStr, "weight") {
		t.Error("Expected weight property in JSON")
	}
}

func TestGenerateInvalidPath(t *testing.T) {
	testTree := createTestTree()
	generator := NewJSONGenerator(testTree)

	// Try to generate to an invalid path
	err := generator.Generate("/invalid/path/that/does/not/exist/output.html")
	if err == nil {
		t.Error("Expected error when generating to invalid path")
	}
}

func TestTechnologyFieldsInJSON(t *testing.T) {
	testTree := createTestTree()
	generator := NewJSONGenerator(testTree)

	tmpDir := t.TempDir()

	err := generator.GenerateJSONFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate JSON files: %v", err)
	}

	// Read physics technologies file
	content, err := os.ReadFile(tmpDir + "/research-physics.json")
	if err != nil {
		t.Fatalf("Failed to read technologies file: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check technologies array
	techs, ok := data["technologies"].([]interface{})
	if !ok {
		t.Fatal("Expected technologies to be array")
	}

	if len(techs) > 0 {
		tech := techs[0].(map[string]interface{})

		requiredFields := []string{
			"key", "name", "cost", "area", "tier", "level",
			"category", "prerequisites", "weight", "sourceFile",
			"isStartTech", "isDangerous", "isRare",
			"isEvent", "isReverse", "isRepeatable", "levels",
			"isGestalt", "isMegacorp",
		}

		for _, field := range requiredFields {
			if _, exists := tech[field]; !exists {
				t.Errorf("Expected field '%s' to exist in technology data", field)
			}
		}
	}
}

func TestEmptyTreeGeneration(t *testing.T) {
	technologies := make(map[string]*models.Technology)
	testTree := tree.NewTechTree(technologies)
	generator := NewJSONGenerator(testTree)

	tmpDir := t.TempDir()

	err := generator.Generate(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate JSON for empty tree: %v", err)
	}

	// Verify metadata file was created
	if _, err := os.Stat(tmpDir + "/metadata.json"); os.IsNotExist(err) {
		t.Error("Expected metadata.json file to be created")
	}
}
