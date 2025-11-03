package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"stellaris-data-parser/lib/generator"
	"stellaris-data-parser/lib/localization"
	"stellaris-data-parser/lib/parser"
	"stellaris-data-parser/lib/tree"
)

const (
	version = "1.0.0"
)

func main() {
	// Define command-line flags
	gameDir := flag.String("input", "", "Path to Stellaris game directory (required)")
	outputDir := flag.String("output", "output", "Output directory for JSON files and icons")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Stellaris Data Parser v%s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Validate input directory
	if *gameDir == "" {
		fmt.Println("Error: game directory is required")
		fmt.Println()
		printHelp()
		os.Exit(1)
	}

	// Check if input directory exists
	if _, err := os.Stat(*gameDir); os.IsNotExist(err) {
		fmt.Printf("Error: game directory does not exist: %s\n", *gameDir)
		os.Exit(1)
	}

	// Detect technology and localization directories
	techDir := filepath.Join(*gameDir, "common", "technology")
	localizationDir := filepath.Join(*gameDir, "localisation")

	// Validate technology directory
	if _, err := os.Stat(techDir); os.IsNotExist(err) {
		fmt.Printf("Error: Technology directory not found: %s\n", techDir)
		fmt.Println("       Make sure you're pointing to the Stellaris game directory")
		fmt.Println("       Expected structure: <game_dir>/common/technology/")
		os.Exit(1)
	}

	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║      Stellaris Data Parser v1.0.0              ║")
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("🎮 Stellaris game directory: %s\n", *gameDir)
	fmt.Println()

	// Parse technology files
	fmt.Printf("📂 Reading technology files from: %s\n", techDir)
	techParser := parser.NewTechParser()

	if err := techParser.ParseDirectory(techDir); err != nil {
		fmt.Printf("❌ Error parsing technology files: %v\n", err)
		os.Exit(1)
	}

	technologies := techParser.GetTechnologies()
	fmt.Printf("✓ Parsed %d technologies\n", len(technologies))

	if len(technologies) == 0 {
		fmt.Println("⚠ Warning: No technologies found in the input directory")
		fmt.Println("   Make sure the directory contains Stellaris technology .txt files")
		os.Exit(1)
	}

	// Parse localization files (English only)
	fmt.Println("\n🌍 Loading English localization data...")
	locParser := localization.NewLocalizationParser()

	if _, err := os.Stat(localizationDir); err == nil {
		fmt.Printf("📂 Reading localization files from: %s\n", localizationDir)
		if err := locParser.ParseDirectory(localizationDir); err != nil {
			fmt.Printf("⚠ Warning: Failed to parse localization files: %v\n", err)
			fmt.Println("   Continuing without localization data...")
		} else {
			// Add English localization data directly to technologies
			for key, tech := range technologies {
				name := locParser.GetLocalizedName(key, "english")
				desc := locParser.GetLocalizedDescription(key, "english")
				if name != "" {
					tech.Name = name
				}
				if desc != "" {
					tech.Description = desc
				}
			}
			fmt.Printf("✓ Added English localization to technologies\n")
		}
	} else {
		fmt.Printf("⚠ Warning: Localization directory not found: %s\n", localizationDir)
		fmt.Println("   Continuing without localization data...")
	}

	// Build technology tree
	fmt.Println("\n🌳 Building technology tree...")
	techTree := tree.NewTechTree(technologies)

	fmt.Printf("✓ Built tree with %d levels\n", techTree.GetMaxLevel()+1)
	fmt.Printf("✓ Found %d root technologies (no prerequisites)\n", len(techTree.GetRootNodes()))

	// Print statistics
	areas := techTree.GetAreas()
	if len(areas) > 0 {
		fmt.Printf("✓ Research areas: %v\n", areas)
	}

	tiers := techTree.GetTiers()
	if len(tiers) > 0 {
		fmt.Printf("✓ Technology tiers: %v\n", tiers)
	}

	// Generate JSON output
	fmt.Printf("\n📊 Generating JSON data files...\n")
	jsonGenerator := generator.NewJSONGenerator(techTree)
	jsonGenerator.SetGameDir(*gameDir) // Set game directory for icon extraction

	// Resolve output path
	absOutputPath, err := filepath.Abs(*outputDir)
	if err != nil {
		absOutputPath = *outputDir
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(absOutputPath, 0755); err != nil {
		fmt.Printf("❌ Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := jsonGenerator.Generate(absOutputPath); err != nil {
		fmt.Printf("❌ Error generating JSON files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ JSON data files created in: %s\n", absOutputPath)
	fmt.Println("  - metadata.json (areas, tiers, categories)")

	// List technology files by area
	if len(areas) > 0 {
		for _, area := range areas {
			fmt.Printf("  - research-%s.json\n", strings.ToLower(area))
		}
	}

	fmt.Println("\n✨ Success! JSON files ready for use with Docusaurus.")
}

func printHelp() {
	fmt.Println("Stellaris Data Parser")
	fmt.Println("Parses Stellaris technology and localization files to generate JSON data and icons for Docusaurus.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  stellaris-data-parser -input <game_directory> [-output <directory>]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -input string")
	fmt.Println("        Path to Stellaris game directory (required)")
	fmt.Println("        Example: C:\\Steam\\steamapps\\common\\Stellaris")
	fmt.Println()
	fmt.Println("  -output string")
	fmt.Println("        Output directory for JSON files and icons (default: output)")
	fmt.Println()
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println()
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate data from Stellaris installation")
	fmt.Println("  stellaris-data-parser -input \"C:\\Steam\\steamapps\\common\\Stellaris\"")
	fmt.Println()
	fmt.Println("  # Specify custom output directory")
	fmt.Println("  stellaris-data-parser -input \"C:\\Steam\\steamapps\\common\\Stellaris\" -output data")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Point -input to the Stellaris game root directory")
	fmt.Println("  - The tool will automatically find common/technology/ and localisation/ subdirectories")
	fmt.Println("  - Default Stellaris path: <Steam>\\steamapps\\common\\Stellaris")
	fmt.Println("  - Generates JSON files for each research area (Physics, Engineering, Society)")
	fmt.Println("  - Each technology includes English name and description")
	fmt.Println("  - Generates metadata.json with areas, tiers, and categories")
	fmt.Println("  - Converts technology icons from DDS to PNG format")
}
