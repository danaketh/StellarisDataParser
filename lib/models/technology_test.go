package models

import (
	"testing"
)

func TestTechnologyStruct(t *testing.T) {
	tech := &Technology{
		Key:         "tech_test",
		Name:        "Test Technology",
		Description: "A test technology",
		Cost:        1000,
		Area:        "physics",
		Tier:        2,
		Category:    []string{"computing", "materials"},
		Prerequisites: []string{"tech_prereq_1", "tech_prereq_2"},
		Weight:      75,
		BaseWeight:  1.5,
		IsStartTech: false,
		IsDangerous: false,
		IsRare:      true,
		IsEvent:     false,
	}

	// Test basic fields
	if tech.Key != "tech_test" {
		t.Errorf("Expected Key to be 'tech_test', got '%s'", tech.Key)
	}

	if tech.Cost != 1000 {
		t.Errorf("Expected Cost to be 1000, got %d", tech.Cost)
	}

	if tech.Area != "physics" {
		t.Errorf("Expected Area to be 'physics', got '%s'", tech.Area)
	}

	if tech.Tier != 2 {
		t.Errorf("Expected Tier to be 2, got %d", tech.Tier)
	}

	// Test array fields
	if len(tech.Category) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(tech.Category))
	}

	if len(tech.Prerequisites) != 2 {
		t.Errorf("Expected 2 prerequisites, got %d", len(tech.Prerequisites))
	}

	// Test boolean fields
	if tech.IsStartTech {
		t.Error("Expected IsStartTech to be false")
	}

	if !tech.IsRare {
		t.Error("Expected IsRare to be true")
	}
}

func TestWeightModifier(t *testing.T) {
	wm := WeightModifier{
		Factor: 2.0,
		Add:    100,
		Conditions: []Condition{
			{
				Type:  "simple",
				Key:   "has_technology",
				Value: "tech_prereq",
			},
		},
	}

	if wm.Factor != 2.0 {
		t.Errorf("Expected Factor to be 2.0, got %f", wm.Factor)
	}

	if wm.Add != 100 {
		t.Errorf("Expected Add to be 100, got %f", wm.Add)
	}

	if len(wm.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(wm.Conditions))
	}
}

func TestCondition(t *testing.T) {
	// Test simple condition
	simpleCondition := Condition{
		Type:  "simple",
		Key:   "is_gestalt",
		Value: true,
	}

	if simpleCondition.Type != "simple" {
		t.Errorf("Expected Type to be 'simple', got '%s'", simpleCondition.Type)
	}

	// Test nested condition
	nestedCondition := Condition{
		Type: "AND",
		Children: []Condition{
			{
				Key:   "has_technology",
				Value: "tech_test",
			},
			{
				Key:   "is_gestalt",
				Value: false,
			},
		},
	}

	if nestedCondition.Type != "AND" {
		t.Errorf("Expected Type to be 'AND', got '%s'", nestedCondition.Type)
	}

	if len(nestedCondition.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(nestedCondition.Children))
	}
}

func TestConditionLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		condType string
	}{
		{"AND condition", "AND"},
		{"OR condition", "OR"},
		{"NOT condition", "NOT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := Condition{
				Type: tt.condType,
				Children: []Condition{
					{Key: "test", Value: true},
				},
			}

			if condition.Type != tt.condType {
				t.Errorf("Expected Type to be '%s', got '%s'", tt.condType, condition.Type)
			}
		})
	}
}

func TestEmpireTypeRestrictions(t *testing.T) {
	tech := &Technology{
		Key:                "tech_test",
		IsGestalt:          true,
		IsMegacorp:         false,
		IsMachineEmpire:    true,
		IsHiveEmpire:       false,
		IsDriveAssimilator: true,
		IsRogueServitor:    false,
	}

	if !tech.IsGestalt {
		t.Error("Expected IsGestalt to be true")
	}

	if !tech.IsMachineEmpire {
		t.Error("Expected IsMachineEmpire to be true")
	}

	if !tech.IsDriveAssimilator {
		t.Error("Expected IsDriveAssimilator to be true")
	}

	if tech.IsMegacorp {
		t.Error("Expected IsMegacorp to be false")
	}
}

func TestModifier(t *testing.T) {
	mod := Modifier{
		Type:  "pop_growth_speed",
		Value: 0.15,
	}

	if mod.Type != "pop_growth_speed" {
		t.Errorf("Expected Type to be 'pop_growth_speed', got '%s'", mod.Type)
	}

	if mod.Value != 0.15 {
		t.Errorf("Expected Value to be 0.15, got %v", mod.Value)
	}
}
