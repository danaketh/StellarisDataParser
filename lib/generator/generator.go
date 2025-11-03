package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"stellaris-data-parser/lib/tree"
)

// JSONGenerator generates JSON data files and icons for Docusaurus
type JSONGenerator struct {
	tree    *tree.TechTree
	gameDir string // Game directory for finding icons
}

// NewJSONGenerator creates a new JSON generator
func NewJSONGenerator(techTree *tree.TechTree) *JSONGenerator {
	return &JSONGenerator{
		tree: techTree,
	}
}

// SetGameDir sets the game directory path for icon extraction
func (g *JSONGenerator) SetGameDir(gameDir string) {
	g.gameDir = gameDir
}

// Generate creates JSON data files and converts icons
func (g *JSONGenerator) Generate(outputPath string) error {
	// outputPath is now the output directory
	outputDir := outputPath

	// Generate separate JSON files
	if err := g.GenerateJSONFiles(outputDir); err != nil {
		return fmt.Errorf("failed to generate JSON files: %w", err)
	}

	// Convert and copy icon files if game directory is set
	if g.gameDir != "" {
		if err := g.ConvertIcons(outputDir); err != nil {
			// Don't fail generation if icons can't be converted
			// Just log a warning
			fmt.Printf("⚠ Warning: Failed to convert some icons: %v\n", err)
		}
	}

	return nil
}

// GenerateJSONFiles creates separate JSON files for technologies by area
func (g *JSONGenerator) GenerateJSONFiles(outputDir string) error {
	// Prepare all data
	allNodes := g.tree.GetAllNodes()
	techsByArea := make(map[string][]map[string]interface{})

	// Process all technologies
	for key, node := range allNodes {
		// Prepare tech data with English localization
		deps := make([]string, len(node.Dependencies))
		for i, dep := range node.Dependencies {
			deps[i] = dep.Tech.Key
		}

		// Use localized name if available, otherwise format from key
		name := node.Tech.Name
		if name == "" {
			name = formatTechName(key)
		}

		techData := map[string]interface{}{
			"key":           key,
			"name":          name,
			"description":   node.Tech.Description,
			"cost":          node.Tech.Cost,
			"area":          node.Tech.Area,
			"tier":          node.Tech.Tier,
			"level":         node.Level,
			"category":      strings.Join(node.Tech.Category, ", "),
			"prerequisites": deps,
			"weight":        node.Tech.Weight,
			"sourceFile":    node.Tech.SourceFile,
			"icon":          node.Tech.Icon,
			"isStartTech":   node.Tech.IsStartTech,
			"isDangerous":   node.Tech.IsDangerous,
			"isRare":        node.Tech.IsRare,
			"isEvent":       node.Tech.IsEvent,
			"isReverse":     node.Tech.IsReverse,
			"isRepeatable":  node.Tech.IsRepeatable,
			"levels":        node.Tech.Levels,
			"isGestalt":     node.Tech.IsGestalt,
			"isMegacorp":    node.Tech.IsMegacorp,
		}

		// Group by area
		area := node.Tech.Area
		if area == "" {
			area = "unknown"
		}
		techsByArea[area] = append(techsByArea[area], techData)
	}

	// Sort technologies within each area
	for area := range techsByArea {
		sort.Slice(techsByArea[area], func(i, j int) bool {
			if techsByArea[area][i]["level"].(int) == techsByArea[area][j]["level"].(int) {
				return techsByArea[area][i]["key"].(string) < techsByArea[area][j]["key"].(string)
			}
			return techsByArea[area][i]["level"].(int) < techsByArea[area][j]["level"].(int)
		})
	}

	// Write separate technology files for each area
	for area, techs := range techsByArea {
		techPath := filepath.Join(outputDir, fmt.Sprintf("research-%s.json", strings.ToLower(area)))
		if err := g.writeJSONFile(techPath, map[string]interface{}{
			"area":         area,
			"technologies": techs,
		}); err != nil {
			return fmt.Errorf("failed to write technologies for area %s: %w", area, err)
		}
	}

	// Write metadata file with areas, tiers, categories, and max level
	metaPath := filepath.Join(outputDir, "metadata.json")
	if err := g.writeJSONFile(metaPath, map[string]interface{}{
		"areas":      g.tree.GetAreas(),
		"tiers":      g.tree.GetTiers(),
		"categories": g.tree.GetCategories(),
		"maxLevel":   g.tree.GetMaxLevel(),
	}); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// writeJSONFile is a helper function to write JSON data to a file
func (g *JSONGenerator) writeJSONFile(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatTechName converts tech key to readable name
func formatTechName(key string) string {
	// Remove prefixes like "tech_"
	name := strings.TrimPrefix(key, "tech_")

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Capitalize words
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// ConvertIcons converts all technology icons from DDS to PNG
func (g *JSONGenerator) ConvertIcons(outputDir string) error {
	if g.gameDir == "" {
		return fmt.Errorf("game directory not set")
	}

	// Create icon converter
	converter := NewIconConverter(g.gameDir, outputDir)

	// Collect all unique icon names
	allNodes := g.tree.GetAllNodes()
	iconNames := make([]string, 0, len(allNodes))
	for _, node := range allNodes {
		iconNames = append(iconNames, node.Tech.Icon)
	}

	// Convert icons
	fmt.Printf("🎨 Converting technology icons...\n")
	converted, err := converter.ConvertIcons(iconNames)
	if err != nil {
		fmt.Printf("⚠ Some icons could not be converted: %v\n", err)
	}

	if converted > 0 {
		fmt.Printf("✓ Converted %d technology icons\n", converted)
	} else {
		fmt.Printf("⚠ No icons were converted (icon files may not exist in game directory)\n")
	}

	return nil
}
