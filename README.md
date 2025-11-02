# Stellaris Research Tree Generator

A Go console application that parses Stellaris technology and localization files to generate an interactive, multilingual HTML visualization of the complete tech tree, showing dependencies and prerequisites.

## Features

- **Automated Parsing**: Reads Stellaris technology files in the game's native format
- **Multilingual Support**: Includes all 10 official Stellaris languages with a language switcher
- **Dependency Resolution**: Automatically builds the complete dependency tree
- **Interactive Visualization**:
  - **Language Switcher**: Change between English, French, German, Spanish, Japanese, Korean, Polish, Russian, Simplified Chinese, and Brazilian Portuguese
  - Click on any technology to see its prerequisites and what it unlocks
  - Hover over technologies to see full descriptions
  - Filter by research area (Physics, Society, Engineering)
  - Filter by tier
  - Search technologies by name
  - Toggle display of starting, rare, and dangerous technologies
- **Beautiful UI**: Modern, space-themed interface with smooth animations
- **Optimized Performance**: Separate JSON data files for faster loading

## Installation

### Prerequisites

- Go 1.16 or higher
- Access to Stellaris game files

### Building from Source

```bash
# Clone or download this repository
cd StellarisResearchTree

# Build the application
go build -o stellaris-research-tree.exe

# Or use the build script
go build
```

## Usage

### Basic Usage

```bash
stellaris-research-tree -input "C:\Steam\steamapps\common\Stellaris"
```

The tool will automatically detect the `common/technology/` and `localisation/` subdirectories.

### Custom Output File

```bash
stellaris-research-tree -input "C:\Steam\steamapps\common\Stellaris" -output my-tech-tree.html
```

### Command-Line Flags

- `-input` (required): Path to the Stellaris game root directory
- `-output` (optional): Output HTML file path (default: `tech-tree.html`)
- `-version`: Display version information
- `-help`: Show help message

### Finding Your Stellaris Installation

The Stellaris game directory is typically located at:

**Windows (Steam)**:
```
C:\Program Files (x86)\Steam\steamapps\common\Stellaris
```

**Windows (GOG)**:
```
C:\Program Files\GOG Games\Stellaris
```

**Windows (Paradox Launcher)**:
```
C:\Program Files\Paradox Interactive\Stellaris
```

**Linux (Steam)**:
```
~/.steam/steam/steamapps/common/Stellaris
```

**macOS (Steam)**:
```
~/Library/Application Support/Steam/steamapps/common/Stellaris
```

## Example Output

The generated HTML file displays:

1. **Technology Cards** organized by dependency level
2. **Color-coded badges** for:
   - Starting technologies (green)
   - Rare technologies (purple)
   - Dangerous technologies (red)
   - Research areas (blue/orange/teal)
3. **Interactive features**:
   - Click any tech to highlight its dependencies and dependents
   - Golden border: Selected technology
   - Green border: Prerequisites (what you need first)
   - Pink border: Unlocks (what this enables)

## How It Works

1. **Localization Parser** (`lib/localization`): Reads all Stellaris language files:
   - Parses YAML localization files for all 10 languages
   - Extracts technology names and descriptions
   - Supports both versioned and non-versioned formats
   - Handles special characters and escape sequences

2. **Technology Parser** (`lib/parser`): Reads Stellaris technology files (.txt) and extracts:
   - Technology keys and metadata
   - Research costs, tiers, and weights
   - Prerequisites and dependencies
   - Special flags (starting, rare, dangerous, event, reverse-engineerable)
   - Empire type restrictions (gestalt, megacorp, machine empire, etc.)
   - Weight modifiers and potential conditions
   - Nested structures using recursive parsing

3. **Tree Builder** (`lib/tree`):
   - Constructs a dependency graph
   - Calculates technology levels based on prerequisites
   - Organizes technologies by area, tier, and category
   - Identifies root nodes and dependency chains

4. **HTML Generator** (`lib/generator`):
   - Creates an interactive visualization with Tailwind CSS
   - Exports technology data to separate JSON file
   - Includes language switcher, filtering, search, and selection
   - Generates lightweight HTML that loads data dynamically
   - Stellaris-themed dark design with colored tech cards

## Technology File Format

The parser understands the Stellaris technology file format:

```
tech_example = {
    cost = 2000
    area = physics
    tier = 2
    category = { particles }
    prerequisites = { "tech_prerequisite_1" "tech_prerequisite_2" }
    weight = 100
}
```

## Development

### Project Structure

```
StellarisResearchTree/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── lib/                         # Core packages
│   ├── models/                  # Data structures
│   │   └── technology.go        # Technology, Modifier, Condition, Localization models
│   ├── localization/            # Localization parsing
│   │   └── localization.go      # YAML localization parser for all languages
│   ├── parser/                  # Parsing logic
│   │   └── parser.go            # Stellaris file parser with nested structure support
│   ├── tree/                    # Dependency tree
│   │   └── tree.go              # Tech tree building and analysis
│   └── generator/               # HTML generation
│       └── generator.go         # HTML template and JSON export
├── testdata/                    # Test fixtures
├── sample_tech_files/           # Example technology files
└── README.md                    # This file
```

### Running Tests

```bash
go test ./...
```

### Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Troubleshooting

### "No technologies found"

- Ensure the input path points to the correct technology directory
- Check that the directory contains `.txt` files
- Verify you have read permissions for the directory

### "Unknown prerequisite" warnings

- This is normal for modded games or incomplete tech trees
- The tool will still generate output, skipping invalid prerequisites

### HTML file doesn't display correctly

- Ensure you're using a modern browser (Chrome, Firefox, Edge, Safari)
- Check that JavaScript is enabled
- Open the HTML file directly (don't try to open it through a web server initially)

## Improvements Based on Java Parser

This implementation incorporates features from the [stellaris-technology](https://github.com/BloodStainedCrow/stellaris-technology) Java parser:

1. **Enhanced Data Model**:
   - Added `WeightModifier` and `Condition` structures
   - Support for empire type restrictions
   - Event and reverse-engineerable tech flags
   - Feature unlocks tracking

2. **Advanced Parsing**:
   - Nested structure parsing for modifiers and conditions
   - Recursive block parsing
   - Support for arrays and complex data types
   - Handles quoted strings, numbers, and booleans
   - AND/OR/NOT logical operators in conditions

3. **Modular Architecture**:
   - Separated concerns into packages (models, parser, tree, generator)
   - Clean interfaces between components
   - Easier to test and extend

## Version History

### v1.0.0
- Initial release
- Full technology parsing with nested structures
- Multilingual support with 10 languages
- Interactive HTML visualization with Tailwind CSS
- Dynamic language switcher in web interface
- Separate JSON data files for optimized loading
- Filter, search, and prerequisite chain selection
- Modular package architecture
- Simplified command-line interface (game directory instead of subdirectories)
- Auto-detection of technology and localization directories
- Enhanced parser based on Java implementation

## License

This project is for personal use and educational purposes. Stellaris and all related content are property of Paradox Interactive.

## Acknowledgments

- Built for fans of Stellaris by Paradox Interactive
- Inspired by the need for better tech tree planning tools
- Java parser reference: [BloodStainedCrow/stellaris-technology](https://github.com/BloodStainedCrow/stellaris-technology)
- Original web visualizer: [turanar/stellaris-tech-tree](https://github.com/turanar/stellaris-tech-tree)
