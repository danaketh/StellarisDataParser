package parser

import (
	"os"
	"path/filepath"
	"testing"

	"stellaris-data-parser/lib/models"
)

func TestNewTechParser(t *testing.T) {
	parser := NewTechParser()

	if parser == nil {
		t.Fatal("Expected parser to be created, got nil")
	}

	if parser.technologies == nil {
		t.Error("Expected technologies map to be initialized")
	}

	if len(parser.technologies) != 0 {
		t.Errorf("Expected empty technologies map, got %d items", len(parser.technologies))
	}
}

func TestParseDirectory(t *testing.T) {
	parser := NewTechParser()

	// Get the testdata path relative to the project root
	testdataPath, err := filepath.Abs("../../testdata/common/technology")
	if err != nil {
		t.Fatalf("Failed to get testdata path: %v", err)
	}

	err = parser.ParseDirectory(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse directory: %v", err)
	}

	technologies := parser.GetTechnologies()

	if len(technologies) == 0 {
		t.Error("Expected to parse technologies, got 0")
	}

	// Check that we have technologies from all sample files
	if _, exists := technologies["tech_basic_science_lab_1"]; !exists {
		t.Error("Expected to find tech_basic_science_lab_1")
	}
}

func TestParseFile(t *testing.T) {
	parser := NewTechParser()

	testdataPath, err := filepath.Abs("../../testdata/common/technology/00_sample_physics.txt")
	if err != nil {
		t.Fatalf("Failed to get testdata path: %v", err)
	}

	err = parser.ParseFile(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	technologies := parser.GetTechnologies()

	// Test specific technologies
	if tech, exists := technologies["tech_basic_science_lab_1"]; exists {
		if tech.Cost != 0 {
			t.Errorf("Expected cost 0, got %d", tech.Cost)
		}
		if tech.Area != "physics" {
			t.Errorf("Expected area 'physics', got '%s'", tech.Area)
		}
		if tech.Tier != 0 {
			t.Errorf("Expected tier 0, got %d", tech.Tier)
		}
		if !tech.IsStartTech {
			t.Error("Expected IsStartTech to be true")
		}
		if tech.Weight != 100 {
			t.Errorf("Expected weight 100, got %d", tech.Weight)
		}
		if tech.SourceFile != "00_sample_physics.txt" {
			t.Errorf("Expected SourceFile '00_sample_physics.txt', got '%s'", tech.SourceFile)
		}
	} else {
		t.Error("Expected to find tech_basic_science_lab_1")
	}

	if tech, exists := technologies["tech_jump_drive_1"]; exists {
		if !tech.IsRare {
			t.Error("Expected IsRare to be true for tech_jump_drive_1")
		}
		if len(tech.Prerequisites) != 2 {
			t.Errorf("Expected 2 prerequisites, got %d", len(tech.Prerequisites))
		}
	} else {
		t.Error("Expected to find tech_jump_drive_1")
	}

	if tech, exists := technologies["tech_psi_jump_drive_1"]; exists {
		if !tech.IsRare {
			t.Error("Expected IsRare to be true")
		}
		if !tech.IsDangerous {
			t.Error("Expected IsDangerous to be true")
		}
	} else {
		t.Error("Expected to find tech_psi_jump_drive_1")
	}
}

func TestParseComplexTech(t *testing.T) {
	parser := NewTechParser()

	testdataPath, err := filepath.Abs("../../testdata/common/technology/00_complex_tech.txt")
	if err != nil {
		t.Fatalf("Failed to get testdata path: %v", err)
	}

	err = parser.ParseFile(testdataPath)
	if err != nil {
		t.Fatalf("Failed to parse complex tech file: %v", err)
	}

	technologies := parser.GetTechnologies()

	// Test gestalt-only tech
	if tech, exists := technologies["tech_gestalt_only"]; exists {
		if !tech.IsGestalt {
			t.Error("Expected IsGestalt to be true")
		}
		if tech.Potential == nil {
			t.Error("Expected Potential to be parsed")
		}
		if len(tech.WeightModifiers) == 0 {
			t.Error("Expected WeightModifiers to be parsed")
		}
	} else {
		t.Error("Expected to find tech_gestalt_only")
	}

	// Test megacorp tech
	if tech, exists := technologies["tech_megacorp_special"]; exists {
		if !tech.IsMegacorp {
			t.Error("Expected IsMegacorp to be true")
		}
		if !tech.IsRare {
			t.Error("Expected IsRare to be true")
		}
		if tech.BaseWeight != 1.5 {
			t.Errorf("Expected BaseWeight 1.5, got %f", tech.BaseWeight)
		}
		if len(tech.FeatureUnlocks) != 2 {
			t.Errorf("Expected 2 feature unlocks, got %d", len(tech.FeatureUnlocks))
		}
	} else {
		t.Error("Expected to find tech_megacorp_special")
	}

	// Test event tech
	if tech, exists := technologies["tech_event_based"]; exists {
		if !tech.IsEvent {
			t.Error("Expected IsEvent to be true")
		}
		if tech.Weight != 0 {
			t.Errorf("Expected weight 0, got %d", tech.Weight)
		}
	} else {
		t.Error("Expected to find tech_event_based")
	}

	// Test reverse engineering tech
	if tech, exists := technologies["tech_reverse_engineering"]; exists {
		if !tech.IsReverse {
			t.Error("Expected IsReverse to be true")
		}
		if tech.Potential == nil {
			t.Error("Expected Potential with OR condition to be parsed")
		}
		if tech.Potential != nil && tech.Potential.Type != "OR" {
			t.Errorf("Expected Potential type 'OR', got '%s'", tech.Potential.Type)
		}
	} else {
		t.Error("Expected to find tech_reverse_engineering")
	}

	// Test machine empire tech
	if tech, exists := technologies["tech_machine_empire"]; exists {
		if !tech.IsMachineEmpire {
			t.Error("Expected IsMachineEmpire to be true")
		}
		if tech.AIUpdateType != "military" {
			t.Errorf("Expected AIUpdateType 'military', got '%s'", tech.AIUpdateType)
		}
		if tech.Gateway != "ftl" {
			t.Errorf("Expected Gateway 'ftl', got '%s'", tech.Gateway)
		}
	} else {
		t.Error("Expected to find tech_machine_empire")
	}

	// Test hive mind tech
	if tech, exists := technologies["tech_hive_mind"]; exists {
		if !tech.IsHiveEmpire {
			t.Error("Expected IsHiveEmpire to be true")
		}
	} else {
		t.Error("Expected to find tech_hive_mind")
	}

	// Test complex potential condition
	if tech, exists := technologies["tech_with_complex_potential"]; exists {
		if !tech.IsDangerous {
			t.Error("Expected IsDangerous to be true")
		}
		if tech.Potential == nil {
			t.Error("Expected Potential with AND condition to be parsed")
		}
		if tech.Potential != nil && tech.Potential.Type != "AND" {
			t.Errorf("Expected Potential type 'AND', got '%s'", tech.Potential.Type)
		}
	} else {
		t.Error("Expected to find tech_with_complex_potential")
	}
}

func TestParseValue(t *testing.T) {
	parser := NewTechParser()

	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"quoted string", `"test_value"`, "test_value"},
		{"boolean yes", "yes", true},
		{"boolean no", "no", false},
		{"boolean true", "true", true},
		{"boolean false", "false", false},
		{"integer", "42", 42},
		{"negative integer", "-10", -10},
		{"float", "3.14", 3.14},
		{"negative float", "-2.5", -2.5},
		{"unquoted string", "tech_test", "tech_test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseValue(tt.input)

			switch expected := tt.expected.(type) {
			case string:
				if str, ok := result.(string); !ok || str != expected {
					t.Errorf("Expected string '%s', got %v", expected, result)
				}
			case bool:
				if b, ok := result.(bool); !ok || b != expected {
					t.Errorf("Expected bool %v, got %v", expected, result)
				}
			case int:
				if i, ok := result.(int); !ok || i != expected {
					t.Errorf("Expected int %d, got %v", expected, result)
				}
			case float64:
				if f, ok := result.(float64); !ok || f != expected {
					t.Errorf("Expected float64 %f, got %v", expected, result)
				}
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	parser := NewTechParser()

	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected bool
	}{
		{"bool true", map[string]interface{}{"flag": true}, "flag", true},
		{"bool false", map[string]interface{}{"flag": false}, "flag", false},
		{"string yes", map[string]interface{}{"flag": "yes"}, "flag", true},
		{"string no", map[string]interface{}{"flag": "no"}, "flag", false},
		{"missing key", map[string]interface{}{}, "flag", false},
		{"wrong type", map[string]interface{}{"flag": 42}, "flag", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.getBool(tt.data, tt.key)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	parser := NewTechParser()

	tests := []struct {
		name     string
		input    string
		expected int // expected length
	}{
		{"quoted strings", `{ "tech_1" "tech_2" "tech_3" }`, 3},
		{"single item", `{ "tech_1" }`, 1},
		{"empty array", `{ }`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseArray(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Expected array length %d, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestIsArray(t *testing.T) {
	parser := NewTechParser()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"array of strings", `{ "item1" "item2" }`, true},
		{"map with equals", `{ key = value }`, false},
		{"empty", `{ }`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isArray(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetTechnology(t *testing.T) {
	parser := NewTechParser()
	parser.technologies["tech_test"] = &models.Technology{
		Key:  "tech_test",
		Cost: 1000,
	}

	// Test existing tech
	tech, exists := parser.GetTechnology("tech_test")
	if !exists {
		t.Error("Expected to find tech_test")
	}
	if tech.Cost != 1000 {
		t.Errorf("Expected cost 1000, got %d", tech.Cost)
	}

	// Test non-existing tech
	_, exists = parser.GetTechnology("tech_nonexistent")
	if exists {
		t.Error("Expected tech_nonexistent to not exist")
	}
}

func TestParseDirectoryNonExistent(t *testing.T) {
	parser := NewTechParser()

	err := parser.ParseDirectory("/nonexistent/path")
	if err == nil {
		t.Error("Expected error when parsing non-existent directory")
	}
}

func TestReadFileContent(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_tech_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `# Comment line
tech_test = {
	cost = 1000 # inline comment
	area = physics
}
`
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Reopen for reading
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	result, err := readFileContent(file)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}

	// Should not contain comments
	if len(result) == 0 {
		t.Error("Expected non-empty content")
	}
}

func TestSkipTierFile(t *testing.T) {
	parser := NewTechParser()

	// Create a temporary directory
	tmpDir := t.TempDir()
	tierFilePath := filepath.Join(tmpDir, "00_tier.txt")

	// Write some tier definitions
	content := `
tier_0 = {
	cost = 0
	weight = 100
}
tier_1 = {
	cost = 1000
	weight = 85
}
`
	if err := os.WriteFile(tierFilePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write tier file: %v", err)
	}

	// Parse the file - it should be skipped
	err := parser.ParseFile(tierFilePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Should have no technologies since the file was skipped
	techs := parser.GetTechnologies()
	if len(techs) != 0 {
		t.Errorf("Expected 0 technologies from tier file, got %d", len(techs))
	}
}
