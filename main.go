package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"stellaris-research-tree/lib/generator"
	"stellaris-research-tree/lib/localization"
	"stellaris-research-tree/lib/models"
	"stellaris-research-tree/lib/parser"
	"stellaris-research-tree/lib/tree"
)

const (
	version = "1.0.0"
)

func main() {
	// Define command-line flags
	gameDir := flag.String("input", "", "Path to Stellaris game directory (required)")
	outputFile := flag.String("output", "tech-tree.html", "Output HTML file path")
	servePort := flag.String("serve", "", "Start HTTP server on specified port (e.g., :8080)")
	showVersion := flag.Bool("version", false, "Show version information")
	showHelp := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Stellaris Research Tree Generator v%s\n", version)
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
	fmt.Println("║   Stellaris Research Tree Generator v1.0.0    ║")
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

	// Parse localization files
	fmt.Println("\n🌍 Loading localization data...")
	locParser := localization.NewLocalizationParser()

	if _, err := os.Stat(localizationDir); err == nil {
		fmt.Printf("📂 Reading localization files from: %s\n", localizationDir)
		if err := locParser.ParseDirectory(localizationDir); err != nil {
			fmt.Printf("⚠ Warning: Failed to parse localization files: %v\n", err)
			fmt.Println("   Continuing without localization data...")
		} else {
			languages := locParser.GetAvailableLanguages()
			fmt.Printf("✓ Loaded %d languages: %v\n", len(languages), languages)

			// Add localization data to technologies
			for key, tech := range technologies {
				tech.Localizations = make(map[string]models.TechLocalization)
				for _, lang := range languages {
					name := locParser.GetLocalizedName(key, lang)
					desc := locParser.GetLocalizedDescription(key, lang)
					if name != "" || desc != "" {
						tech.Localizations[lang] = models.TechLocalization{
							Name:        name,
							Description: desc,
						}
					}
				}
			}
			fmt.Printf("✓ Added localization data to technologies\n")
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

	// Generate HTML output
	fmt.Printf("\n🎨 Generating HTML visualization...\n")
	htmlGenerator := generator.NewHTMLGenerator(techTree)

	// Resolve output path
	absOutputPath, err := filepath.Abs(*outputFile)
	if err != nil {
		absOutputPath = *outputFile
	}

	if err := htmlGenerator.Generate(absOutputPath); err != nil {
		fmt.Printf("❌ Error generating HTML: %v\n", err)
		os.Exit(1)
	}

	outputDir := filepath.Dir(absOutputPath)

	fmt.Printf("✓ HTML file created: %s\n", absOutputPath)
	fmt.Printf("✓ JSON data files created in: %s\n", outputDir)
	fmt.Println("  - localizations.json (language data)")
	fmt.Println("  - metadata.json (areas, tiers, categories)")

	// List technology files by area
	if len(areas) > 0 {
		for _, area := range areas {
			fmt.Printf("  - technologies-%s.json\n", strings.ToLower(area))
		}
	}

	fmt.Println("\n✨ Success! Open the HTML file in your browser to view the tech tree.")
	fmt.Println("   Note: Keep all JSON files in the same directory as the HTML file.")

	// Start HTTP server if requested
	if *servePort != "" {
		startServer(*servePort, outputDir)
	}
}

func printHelp() {
	fmt.Println("Stellaris Research Tree Generator")
	fmt.Println("Parses Stellaris technology and localization files to generate an interactive HTML tech tree.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  stellaris-research-tree -input <game_directory> [-output <path>]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -input string")
	fmt.Println("        Path to Stellaris game directory (required)")
	fmt.Println("        Example: C:\\Steam\\steamapps\\common\\Stellaris")
	fmt.Println()
	fmt.Println("  -output string")
	fmt.Println("        Output HTML file path (default: tech-tree.html)")
	fmt.Println()
	fmt.Println("  -serve string")
	fmt.Println("        Start HTTP server on specified port after generation")
	fmt.Println("        Example: -serve :8080 or -serve 8080")
	fmt.Println()
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println()
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Generate tech tree from Stellaris installation")
	fmt.Println("  stellaris-research-tree -input \"C:\\Steam\\steamapps\\common\\Stellaris\"")
	fmt.Println()
	fmt.Println("  # Specify custom output file")
	fmt.Println("  stellaris-research-tree -input \"C:\\Steam\\steamapps\\common\\Stellaris\" -output stellaris-tree.html")
	fmt.Println()
	fmt.Println("  # Generate and serve on HTTP server")
	fmt.Println("  stellaris-research-tree -input \"C:\\Steam\\steamapps\\common\\Stellaris\" -serve :8080")
	fmt.Println()
	fmt.Println("Notes:")
	fmt.Println("  - Point -input to the Stellaris game root directory")
	fmt.Println("  - The tool will automatically find common/technology/ and localisation/ subdirectories")
	fmt.Println("  - Default Stellaris path: <Steam>\\steamapps\\common\\Stellaris")
	fmt.Println("  - Generates two files: HTML and JSON data file")
	fmt.Println("  - Supports 10 languages with a language switcher in the web interface")
	fmt.Println("  - Keep both output files in the same directory for the visualization to work")
}

func startServer(port, dir string) {
	// Ensure port starts with ":"
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// Create file server for the output directory
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	fmt.Printf("\n🌐 Starting HTTP server on http://localhost%s\n", port)
	fmt.Printf("   Serving files from: %s\n", dir)
	fmt.Println("   Press Ctrl+C to stop the server")
	fmt.Println()

	// Start the server
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
