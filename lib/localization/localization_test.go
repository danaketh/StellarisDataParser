package localization

import (
	"testing"
)

func TestResolveVariables(t *testing.T) {
	// Create a test parser with sample data
	parser := NewLocalizationParser()
	parser.data.Languages["english"] = &LanguageData{
		Translations: map[string]string{
			"building_affluence_emporium": "Affluence Emporium",
			"building_micro_forge":        "Micro Forge",
			"BOARDING_CABLES":             "Boarding Cables",
			"MANDIBLE_2":                  "Mandible II",
			"MANDIBLE_3":                  "Mandible III",
			"pc_ringworld_habitable":      "Ringworld",
			"building_fe_lab_1":           "Advanced Lab",
			"clue":                        "Clue",
			// Test nested resolution
			"nested_ref":                  "$building_micro_forge$",
			"double_nested":               "$nested_ref$",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple variable resolution",
			input:    "$BOARDING_CABLES$",
			expected: "Boarding Cables",
		},
		{
			name:     "Variable in text",
			input:    "Technology unlocks $building_affluence_emporium$",
			expected: "Technology unlocks Affluence Emporium",
		},
		{
			name:     "Multiple variables",
			input:    "$MANDIBLE_2$ and $MANDIBLE_3$",
			expected: "Mandible II and Mandible III",
		},
		{
			name:     "Nested variable resolution",
			input:    "$nested_ref$",
			expected: "Micro Forge",
		},
		{
			name:     "Double nested variable resolution",
			input:    "$double_nested$",
			expected: "Micro Forge",
		},
		{
			name:     "Non-existent variable",
			input:    "$non_existent_var$",
			expected: "$non_existent_var$",
		},
		{
			name:     "No variables",
			input:    "Plain text without variables",
			expected: "Plain text without variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.resolveVariables(tt.input, "english")
			if result != tt.expected {
				t.Errorf("resolveVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetLocalizedNameWithVariables(t *testing.T) {
	// Create a test parser with sample data
	parser := NewLocalizationParser()
	parser.data.Languages["english"] = &LanguageData{
		Translations: map[string]string{
			"tech_boarding_cables":        "$BOARDING_CABLES$",
			"BOARDING_CABLES":             "Boarding Cables",
			"tech_fe_affluence_1":         "$building_affluence_emporium$",
			"building_affluence_emporium": "Affluence Emporium",
		},
	}

	tests := []struct {
		name     string
		techKey  string
		expected string
	}{
		{
			name:     "Tech with variable reference",
			techKey:  "tech_boarding_cables",
			expected: "Boarding Cables",
		},
		{
			name:     "Tech with building reference",
			techKey:  "tech_fe_affluence_1",
			expected: "Affluence Emporium",
		},
		{
			name:     "Non-existent tech",
			techKey:  "non_existent_tech",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetLocalizedName(tt.techKey, "english")
			if result != tt.expected {
				t.Errorf("GetLocalizedName() = %q, want %q", result, tt.expected)
			}
		})
	}
}
