package parser

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"stellaris-data-parser/lib/models"
)

// TechParser handles parsing of Stellaris technology files
type TechParser struct {
	technologies map[string]*models.Technology
}

// NewTechParser creates a new technology parser
func NewTechParser() *TechParser {
	return &TechParser{
		technologies: make(map[string]*models.Technology),
	}
}

// ParseDirectory parses all technology files in a directory
func (p *TechParser) ParseDirectory(path string) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .txt files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			if err := p.ParseFile(filePath); err != nil {
				fmt.Printf("Warning: failed to parse %s: %v\n", filePath, err)
			}
		}
		return nil
	})
}

// ParseFile parses a single technology file
func (p *TechParser) ParseFile(path string) error {
	// Get just the filename (not the full path)
	filename := filepath.Base(path)

	// Skip tier definition files
	if filename == "00_tier.txt" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := readFileContent(file)
	if err != nil {
		return err
	}

	techs := p.parseContent(content, filename)
	for key, tech := range techs {
		p.technologies[key] = tech
	}

	return nil
}

// readFileContent reads and preprocesses file content
func readFileContent(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	var content strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		// Remove comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line != "" {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	return content.String(), scanner.Err()
}

// parseContent parses the preprocessed content
func (p *TechParser) parseContent(content string, filename string) map[string]*models.Technology {
	techs := make(map[string]*models.Technology)

	// Split into top-level blocks
	blocks := p.extractTopLevelBlocks(content)

	for key, blockContent := range blocks {
		tech := p.parseTechnologyBlock(key, blockContent)
		tech.SourceFile = filename
		techs[key] = tech
	}

	return techs
}

// extractTopLevelBlocks extracts technology definition blocks
func (p *TechParser) extractTopLevelBlocks(content string) map[string]string {
	blocks := make(map[string]string)

	// Pattern to match tech_name = { ... }
	pattern := regexp.MustCompile(`(\w+)\s*=\s*\{`)

	lines := strings.Split(content, "\n")
	var currentKey string
	var currentBlock strings.Builder
	braceDepth := 0
	inBlock := false

	for _, line := range lines {
		if matches := pattern.FindStringSubmatch(line); matches != nil && braceDepth == 0 {
			// Save previous block if exists
			if inBlock && currentKey != "" {
				blocks[currentKey] = currentBlock.String()
			}

			currentKey = matches[1]
			currentBlock.Reset()
			inBlock = true

			// Count braces in this line
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
		} else if inBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			if braceDepth == 0 {
				blocks[currentKey] = currentBlock.String()
				inBlock = false
				currentKey = ""
				currentBlock.Reset()
			}
		}
	}

	// Save last block if exists
	if inBlock && currentKey != "" {
		blocks[currentKey] = currentBlock.String()
	}

	return blocks
}

// parseTechnologyBlock parses a single technology block
func (p *TechParser) parseTechnologyBlock(key, content string) *models.Technology {
	tech := &models.Technology{
		Key:             key,
		Prerequisites:   []string{},
		Category:        []string{},
		FeatureUnlocks:  []string{},
		WeightModifiers: []models.WeightModifier{},
	}

	// Parse the block as a map
	data := p.parseBlock(content)

	// Extract simple fields
	if cost, ok := data["cost"].(int); ok {
		tech.Cost = cost
	}
	if area, ok := data["area"].(string); ok {
		tech.Area = area
	}
	if tier, ok := data["tier"].(int); ok {
		tech.Tier = tier
	}
	if weight, ok := data["weight"].(int); ok {
		tech.Weight = weight
	}
	if baseWeight, ok := data["base_weight"].(float64); ok {
		tech.BaseWeight = baseWeight
	}

	// Boolean flags
	tech.IsStartTech = p.getBool(data, "start_tech")
	tech.IsDangerous = p.getBool(data, "is_dangerous")
	tech.IsRare = p.getBool(data, "is_rare")
	tech.IsEvent = p.getBool(data, "is_event_tech")
	tech.IsReverse = p.getBool(data, "is_reverse_engineerable")
	tech.IsRepeatable = p.getBool(data, "is_repeatable")
	tech.IsGestalt = p.getBool(data, "is_gestalt")
	tech.IsMegacorp = p.getBool(data, "is_megacorp")
	tech.IsMachineEmpire = p.getBool(data, "is_machine_empire")
	tech.IsHiveEmpire = p.getBool(data, "is_hive_empire")
	tech.IsDriveAssimilator = p.getBool(data, "is_drive_assimilator")
	tech.IsRogueServitor = p.getBool(data, "is_rogue_servitor")

	// Repeatable tech levels
	if levels, ok := data["levels"].(int); ok {
		tech.Levels = levels
	}

	// String fields
	if aiUpdateType, ok := data["ai_update_type"].(string); ok {
		tech.AIUpdateType = aiUpdateType
	}
	if gateway, ok := data["gateway"].(string); ok {
		tech.Gateway = gateway
	}
	if icon, ok := data["icon"].(string); ok {
		tech.Icon = icon
	} else {
		// Default to technology key if no icon specified
		tech.Icon = key
	}

	// Array fields
	if prereqs, ok := data["prerequisites"].([]interface{}); ok {
		for _, p := range prereqs {
			if str, ok := p.(string); ok {
				tech.Prerequisites = append(tech.Prerequisites, str)
			}
		}
	}

	if categories, ok := data["category"].([]interface{}); ok {
		for _, c := range categories {
			if str, ok := c.(string); ok {
				tech.Category = append(tech.Category, str)
			}
		}
	}

	if features, ok := data["feature_unlocks"].([]interface{}); ok {
		for _, f := range features {
			if str, ok := f.(string); ok {
				tech.FeatureUnlocks = append(tech.FeatureUnlocks, str)
			}
		}
	}

	// Parse weight_modifiers
	if modifiers, ok := data["weight_modifiers"].(map[string]interface{}); ok {
		tech.WeightModifiers = p.parseWeightModifiers(modifiers)
	}

	// Parse potential
	if potential, ok := data["potential"].(map[string]interface{}); ok {
		tech.Potential = p.parseCondition(potential)
	}

	return tech
}

// parseBlock parses a block of content into a map
func (p *TechParser) parseBlock(content string) map[string]interface{} {
	result := make(map[string]interface{})

	lines := strings.Split(content, "\n")
	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || line == "}" {
			i++
			continue
		}

		// Check for key = value or key = { block }
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			i++
			continue
		}

		key := strings.TrimSpace(parts[0])
		valuePart := strings.TrimSpace(parts[1])

		// Check if it's a block
		if strings.HasPrefix(valuePart, "{") {
			// Extract the block
			blockContent, newIndex := p.extractBlock(lines, i)
			i = newIndex

			// Parse the block
			if p.isArray(blockContent) {
				result[key] = p.parseArray(blockContent)
			} else {
				result[key] = p.parseBlock(blockContent)
			}
		} else {
			// Simple value
			result[key] = p.parseValue(valuePart)
			i++
		}
	}

	return result
}

// extractBlock extracts a { ... } block starting from the current line
// Returns the content WITHOUT the outer braces
func (p *TechParser) extractBlock(lines []string, startIndex int) (string, int) {
	var block strings.Builder
	braceDepth := 0
	started := false
	firstBrace := true

	for i := startIndex; i < len(lines); i++ {
		line := lines[i]

		for _, char := range line {
			if char == '{' {
				braceDepth++
				started = true
				// Skip the first opening brace
				if firstBrace {
					firstBrace = false
					continue
				}
			} else if char == '}' {
				braceDepth--
				// Skip the last closing brace
				if braceDepth == 0 {
					return block.String(), i + 1
				}
			}

			if started && braceDepth > 0 {
				block.WriteRune(char)
			}
		}

		if started && braceDepth > 0 {
			block.WriteRune('\n')
		}
	}

	return block.String(), len(lines)
}

// isArray checks if a block represents an array
func (p *TechParser) isArray(content string) bool {
	// Remove braces and whitespace
	content = strings.Trim(content, "{} \n\t")

	// If it contains = it's likely a map, not an array
	return !strings.Contains(content, "=")
}

// parseArray parses an array block
func (p *TechParser) parseArray(content string) []interface{} {
	var result []interface{}

	// Remove outer braces
	content = strings.Trim(content, "{} \n\t")

	// Split by quotes and spaces
	stringPattern := regexp.MustCompile(`"([^"]+)"`)
	matches := stringPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			result = append(result, match[1])
		}
	}

	// If no quoted strings found, try splitting by whitespace
	if len(result) == 0 {
		parts := strings.Fields(content)
		for _, part := range parts {
			result = append(result, p.parseValue(part))
		}
	}

	return result
}

// parseValue parses a single value
func (p *TechParser) parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Remove trailing punctuation
	value = strings.TrimRight(value, ",")

	// String
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return strings.Trim(value, "\"")
	}

	// Boolean
	if value == "yes" || value == "true" {
		return true
	}
	if value == "no" || value == "false" {
		return false
	}

	// Integer
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Default to string
	return value
}

// getBool safely gets a boolean value from the map
func (p *TechParser) getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
		if s, ok := val.(string); ok {
			return s == "yes" || s == "true"
		}
	}
	return false
}

// parseWeightModifiers parses weight_modifiers block
func (p *TechParser) parseWeightModifiers(data map[string]interface{}) []models.WeightModifier {
	var modifiers []models.WeightModifier

	// Weight modifiers can have factor, add, and various conditions
	if factor, ok := data["factor"]; ok {
		mod := models.WeightModifier{}
		if f, ok := factor.(float64); ok {
			mod.Factor = f
		} else if i, ok := factor.(int); ok {
			mod.Factor = float64(i)
		}
		modifiers = append(modifiers, mod)
	}

	if add, ok := data["add"]; ok {
		mod := models.WeightModifier{}
		if a, ok := add.(float64); ok {
			mod.Add = a
		} else if i, ok := add.(int); ok {
			mod.Add = float64(i)
		}
		modifiers = append(modifiers, mod)
	}

	return modifiers
}

// parseCondition parses a condition block
func (p *TechParser) parseCondition(data map[string]interface{}) *models.Condition {
	condition := &models.Condition{
		Children: []models.Condition{},
		Raw:      data,
	}

	// Check for logical operators
	if andBlock, ok := data["AND"].(map[string]interface{}); ok {
		condition.Type = "AND"
		for key, val := range andBlock {
			child := &models.Condition{
				Key:   key,
				Value: val,
			}
			condition.Children = append(condition.Children, *child)
		}
	} else if orBlock, ok := data["OR"].(map[string]interface{}); ok {
		condition.Type = "OR"
		for key, val := range orBlock {
			child := &models.Condition{
				Key:   key,
				Value: val,
			}
			condition.Children = append(condition.Children, *child)
		}
	} else if notBlock, ok := data["NOT"].(map[string]interface{}); ok {
		condition.Type = "NOT"
		for key, val := range notBlock {
			child := &models.Condition{
				Key:   key,
				Value: val,
			}
			condition.Children = append(condition.Children, *child)
		}
	} else {
		// Simple condition
		for key, val := range data {
			condition.Key = key
			condition.Value = val
			break
		}
	}

	return condition
}

// GetTechnologies returns all parsed technologies
func (p *TechParser) GetTechnologies() map[string]*models.Technology {
	return p.technologies
}

// GetTechnology returns a specific technology by key
func (p *TechParser) GetTechnology(key string) (*models.Technology, bool) {
	tech, exists := p.technologies[key]
	return tech, exists
}
