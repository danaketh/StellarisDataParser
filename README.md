# Stellaris Data Parser

A Go console application that parses Stellaris technology and localization files to generate JSON data files and icons for use with Docusaurus or other web applications.

## Features

- **Automated Parsing**: Reads Stellaris technology files in the game's native format
- **English Localization**: Extracts English names and descriptions for all technologies
- **Dependency Resolution**: Automatically builds the complete dependency tree
- **JSON Export**: Generates structured JSON files organized by research area
- **Icon Conversion**: Converts technology icons from DDS to PNG format
- **Metadata Generation**: Exports research areas, tiers, categories, and tree depth

## Installation

### Prerequisites

- Go 1.16 or higher
- Access to Stellaris game files

### Building from Source

```bash
# Clone or download this repository
cd StellarisDataParser

# Build the application
go build

# This will create stellaris-data-parser.exe (Windows) or stellaris-data-parser (Linux/Mac)
```

## Usage

### Basic Usage

```bash
stellaris-data-parser -input "C:\Steam\steamapps\common\Stellaris"
```

The tool will:
1. Automatically detect the `common/technology/` subdirectory
2. Parse all technology files
3. Load English localization from `localisation/` subdirectory
4. Generate JSON files in the `output/` directory
5. Convert technology icons to PNG format

### Custom Output Directory

```bash
stellaris-data-parser -input "C:\Steam\steamapps\common\Stellaris" -output data
```

### Command-Line Flags

- `-input` (required): Path to the Stellaris game root directory
- `-output` (optional): Output directory for JSON files and icons (default: `output`)
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

## Generated Output

The tool generates the following files in the output directory:

### JSON Data Files

- **`research-physics.json`** - All physics research technologies
- **`research-engineering.json`** - All engineering research technologies
- **`research-society.json`** - All society research technologies
- **`metadata.json`** - Research areas, tiers, categories, and max tree level

### Icons Directory

- **`icons/`** - Contains PNG versions of all technology icons

### JSON Structure

Each research JSON file contains:

```json
{
  "area": "physics",
  "technologies": [
    {
      "key": "tech_lasers_1",
      "name": "Red Lasers",
      "description": "Basic laser technology...",
      "cost": 1000,
      "area": "physics",
      "tier": 1,
      "level": 0,
      "category": "particles",
      "prerequisites": [],
      "weight": 100,
      "sourceFile": "00_phys_weapon_tech.txt",
      "icon": "tech_lasers_1",
      "isStartTech": false,
      "isDangerous": false,
      "isRare": false,
      "isEvent": false,
      "isReverse": false,
      "isRepeatable": false,
      "levels": 0,
      "isGestalt": false,
      "isMegacorp": false
    }
  ]
}
```

The `metadata.json` file contains:

```json
{
  "areas": ["physics", "engineering", "society"],
  "tiers": [0, 1, 2, 3, 4, 5],
  "categories": ["particles", "computing", "field_manipulation", ...],
  "maxLevel": 8
}
```

## How It Works

1. **Localization Parser** (`lib/localization`):
   - Reads Stellaris YAML localization files
   - Extracts English technology names and descriptions
   - Handles special characters and escape sequences

2. **Technology Parser** (`lib/parser`):
   - Reads Stellaris technology files (.txt)
   - Extracts technology metadata (cost, tier, area, etc.)
   - Parses prerequisites and dependencies
   - Identifies special flags (starting, rare, dangerous, etc.)
   - Handles empire type restrictions
   - Supports nested structures using recursive parsing

3. **Tree Builder** (`lib/tree`):
   - Constructs a dependency graph
   - Calculates technology levels based on prerequisites
   - Organizes technologies by area, tier, and category
   - Identifies root nodes and dependency chains

4. **JSON Generator** (`lib/generator`):
   - Exports separate JSON files for each research area
   - Generates metadata file with areas, tiers, and categories
   - Embeds English names and descriptions directly in technology objects

5. **Icon Converter** (`lib/generator/icons.go`):
   - Locates technology icons in the game files
   - Converts DDS format to PNG
   - Organizes icons in the output directory

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
    is_rare = yes
    is_dangerous = yes
}
```

## Development

### Project Structure

```
StellarisDataParser/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── lib/                         # Core packages
│   ├── models/                  # Data structures
│   │   └── technology.go        # Technology, Modifier, Condition models
│   ├── localization/            # Localization parsing
│   │   └── localization.go      # YAML localization parser
│   ├── parser/                  # Parsing logic
│   │   └── parser.go            # Stellaris file parser
│   ├── tree/                    # Dependency tree
│   │   └── tree.go              # Tech tree building and analysis
│   └── generator/               # JSON and icon generation
│       ├── generator.go         # JSON export
│       └── icons.go             # Icon conversion (DDS to PNG)
├── testdata/                    # Test fixtures
└── README.md                    # This file
```

### Running Tests

```bash
go test ./...
```

### Using with Docusaurus

The generated JSON files are ready to be imported into a Docusaurus application:

1. Copy the generated JSON files to your Docusaurus `static/data/` directory
2. Copy the `icons/` directory to your Docusaurus `static/` directory
3. Import and use the JSON data in your React components:

```javascript
import physicsData from '@site/static/data/research-physics.json';
import metadata from '@site/static/data/metadata.json';

// Access technology data
const technologies = physicsData.technologies;
const areas = metadata.areas;
```

## Troubleshooting

### "No technologies found"

- Ensure the input path points to the Stellaris game root directory
- Check that `common/technology/` subdirectory exists
- Verify you have read permissions for the directory

### "Unknown prerequisite" warnings

- This is normal for modded games or incomplete tech trees
- The tool will still generate output, skipping invalid prerequisites

### "No icons were converted"

- This warning appears when the game's `gfx/interface/icons/technologies/` directory is not found
- Icons are optional - the JSON data will still be generated correctly
- Make sure you're pointing to the game root directory, not a subdirectory

### Missing localization data

- If localization files aren't found, technology names will be auto-generated from keys
- For example: `tech_lasers_1` becomes "Lasers 1"
- Point to the game root directory to include localization

## Dependencies

- [github.com/lukegb/dds](https://github.com/lukegb/dds) - DDS image format decoder

## Version History

### v1.0.0
- Initial release
- Full technology parsing with nested structures
- English localization support
- JSON export organized by research area
- Metadata generation (areas, tiers, categories)
- DDS to PNG icon conversion
- Modular package architecture
- Auto-detection of technology and localization directories
- Optimized for Docusaurus integration

## License

This project is open source and available for personal use and educational purposes.

### Copyright Notice

**Stellaris Game Data**: All Stellaris game data, including technology definitions, localization text, icons, and other game assets, are the intellectual property of Paradox Development Studio and Paradox Interactive AB. This parser tool is designed to work with legally obtained copies of Stellaris and does not distribute any game content.

**This Tool**: The Stellaris Data Parser source code is provided as-is for the community. Users must own a legitimate copy of Stellaris to use this tool, as it requires access to the game's data files.

Stellaris © 2016-2025 Paradox Development Studio. Developed by Paradox Development Studio. Published by Paradox Interactive AB. STELLARIS and PARADOX INTERACTIVE are trademarks and/or registered trademarks of Paradox Interactive AB in Europe, the U.S., and other countries.

## Acknowledgments

- Built for fans of Stellaris by Paradox Interactive
- Inspired by the need for better tech tree visualization tools
- Java parser reference: [BloodStainedCrow/stellaris-technology](https://github.com/BloodStainedCrow/stellaris-technology)
