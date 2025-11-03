package models

// Technology represents a single research technology in Stellaris
type Technology struct {
	Key           string
	Name          string
	Description   string
	Cost          int
	Area          string
	Tier          int
	Category      []string
	Prerequisites []string
	Weight        int
	BaseWeight    float64
	SourceFile    string // The filename this technology was parsed from
	Icon          string // Icon filename (without extension), defaults to tech key if not specified
	IsStartTech   bool
	IsDangerous   bool
	IsRare        bool
	IsEvent       bool
	IsRepeatable  bool
	Levels        int // For repeatable technologies
	// Empire type restrictions
	IsGestalt          bool
	IsMegacorp         bool
	IsMachineEmpire    bool
	IsHiveEmpire       bool
	IsDriveAssimilator bool
	IsRogueServitor    bool
	// Additional fields
	FeatureUnlocks   []string
	WeightModifiers  []WeightModifier
	Potential        *Condition
	AIUpdateType string
	Gateway      string
	IsReverse    bool
}

// WeightModifier represents a modifier that affects technology weight
type WeightModifier struct {
	Factor     float64
	Add        float64
	Conditions []Condition
}

// Condition represents a conditional statement in Stellaris scripting
type Condition struct {
	Type     string                 // AND, OR, NOT, or specific condition type
	Key      string                 // The condition key (e.g., "has_technology")
	Value    interface{}            // The condition value
	Operator string                 // Comparison operator (=, >, <, etc.)
	Children []Condition            // Nested conditions
	Raw      map[string]interface{} // Raw data for complex structures
}

// Modifier represents a game effect or modifier
type Modifier struct {
	Type  string
	Value interface{}
}
