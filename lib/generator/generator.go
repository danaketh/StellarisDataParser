package generator

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"stellaris-research-tree/lib/tree"
)

// HTMLGenerator generates HTML visualization of the tech tree
type HTMLGenerator struct {
	tree     *tree.TechTree
	template *template.Template
}

// NewHTMLGenerator creates a new HTML generator
func NewHTMLGenerator(techTree *tree.TechTree) *HTMLGenerator {
	return &HTMLGenerator{
		tree:     techTree,
		template: template.Must(template.New("tech-tree").Parse(htmlTemplate)),
	}
}

// Generate creates the HTML file and JSON data files
func (g *HTMLGenerator) Generate(outputPath string) error {
	// Get base path for JSON files
	basePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath))
	outputDir := filepath.Dir(outputPath)

	// Generate separate JSON files
	if err := g.GenerateJSONFiles(outputDir, filepath.Base(basePath)); err != nil {
		return fmt.Errorf("failed to generate JSON files: %w", err)
	}

	// Generate HTML file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Pass base filename to template
	data := map[string]interface{}{
		"BaseFilename": filepath.Base(basePath),
	}

	if err := g.template.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateJSONFiles creates separate JSON files for localizations and technologies by area
func (g *HTMLGenerator) GenerateJSONFiles(outputDir, baseFilename string) error {
	// Prepare all data
	allNodes := g.tree.GetAllNodes()
	languagesMap := make(map[string]bool)
	localizationsData := make(map[string]map[string]map[string]string) // techKey -> lang -> {name, description}
	techsByArea := make(map[string][]map[string]interface{})            // area -> technologies

	// Process all technologies
	for key, node := range allNodes {
		// Collect localizations
		for lang := range node.Tech.Localizations {
			languagesMap[lang] = true
			if localizationsData[key] == nil {
				localizationsData[key] = make(map[string]map[string]string)
			}
			localizationsData[key][lang] = map[string]string{
				"name":        node.Tech.Localizations[lang].Name,
				"description": node.Tech.Localizations[lang].Description,
			}
		}

		// Prepare tech data without localizations
		deps := make([]string, len(node.Dependencies))
		for i, dep := range node.Dependencies {
			deps[i] = dep.Tech.Key
		}

		techData := map[string]interface{}{
			"key":           key,
			"name":          formatTechName(key),
			"cost":          node.Tech.Cost,
			"area":          node.Tech.Area,
			"tier":          node.Tech.Tier,
			"level":         node.Level,
			"category":      strings.Join(node.Tech.Category, ", "),
			"prerequisites": deps,
			"weight":        node.Tech.Weight,
			"sourceFile":    node.Tech.SourceFile,
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

	// Convert languages map to sorted slice
	languages := make([]string, 0, len(languagesMap))
	for lang := range languagesMap {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	// Write localizations.json
	locPath := filepath.Join(outputDir, "localizations.json")
	if err := g.writeJSONFile(locPath, map[string]interface{}{
		"languages":     languages,
		"localizations": localizationsData,
	}); err != nil {
		return fmt.Errorf("failed to write localizations: %w", err)
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
		techPath := filepath.Join(outputDir, fmt.Sprintf("technologies-%s.json", strings.ToLower(area)))
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
func (g *HTMLGenerator) writeJSONFile(path string, data interface{}) error {
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

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stellaris Technology Tree</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            theme: {
                extend: {
                    colors: {
                        'stellaris-blue': '#6eb4ff',
                        'stellaris-physics': '#6eb4ff',
                        'stellaris-society': '#5fca5f',
                        'stellaris-engineering': '#ff8d3a',
                        'stellaris-dark': '#0d1117',
                        'stellaris-dark-secondary': '#161b22',
                        'stellaris-panel': '#1c2128',
                    }
                }
            }
        }
    </script>
    <style>
        body {
            background: linear-gradient(135deg, #0d1117 0%, #161b22 50%, #0d1117 100%);
            font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
        }
        .tech-card {
            background: linear-gradient(135deg, rgba(28, 33, 40, 0.95) 0%, rgba(22, 27, 34, 0.95) 100%);
            border: 1px solid rgba(110, 180, 255, 0.3);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
        }
        .tech-card:hover {
            border-color: rgba(110, 180, 255, 0.6);
            box-shadow: 0 4px 16px rgba(110, 180, 255, 0.2);
        }
        .area-physics { border-left: 3px solid #6eb4ff; }
        .area-society { border-left: 3px solid #5fca5f; }
        .area-engineering { border-left: 3px solid #ff8d3a; }
    </style>
</head>
<body class="text-gray-200 min-h-screen p-4 md:p-6">
    <div class="container mx-auto max-w-7xl">
        <header class="mb-6">
            <h1 class="text-center text-stellaris-blue text-3xl md:text-4xl font-bold mb-2">
                Stellaris Technology Tree
            </h1>
            <p class="text-center text-gray-400 text-sm">Interactive Research Dependency Viewer</p>
        </header>

        <div class="bg-stellaris-panel/80 backdrop-blur-sm p-4 md:p-6 rounded-lg mb-6 border border-gray-700/50 shadow-xl">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
                <div>
                    <label for="language-select" class="block mb-2 text-sm font-semibold text-gray-300">🌍 Language</label>
                    <select id="language-select" onchange="changeLanguage()"
                        class="w-full px-3 py-2 bg-stellaris-dark border border-gray-600 rounded text-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-stellaris-blue capitalize">
                        <option value="">Loading...</option>
                    </select>
                </div>

                <div>
                    <label for="search" class="block mb-2 text-sm font-semibold text-gray-300">Search</label>
                    <input type="text" id="search" placeholder="Search technologies..."
                        class="w-full px-3 py-2 bg-stellaris-dark border border-gray-600 rounded text-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-stellaris-blue focus:border-transparent placeholder-gray-500"
                        onkeyup="filterTechs()">
                </div>

                <div>
                    <label for="area-filter" class="block mb-2 text-sm font-semibold text-gray-300">Research Area</label>
                    <select id="area-filter" onchange="filterTechs()"
                        class="w-full px-3 py-2 bg-stellaris-dark border border-gray-600 rounded text-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-stellaris-blue">
                        <option value="">All Areas</option>
                    </select>
                </div>

                <div>
                    <label for="tier-filter" class="block mb-2 text-sm font-semibold text-gray-300">Tier</label>
                    <select id="tier-filter" onchange="filterTechs()"
                        class="w-full px-3 py-2 bg-stellaris-dark border border-gray-600 rounded text-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-stellaris-blue">
                        <option value="">All Tiers</option>
                    </select>
                </div>
            </div>

            <div class="flex flex-wrap gap-4 text-sm">
                <label class="inline-flex items-center cursor-pointer hover:text-white transition-colors">
                    <input type="checkbox" id="show-start" checked onchange="filterTechs()"
                        class="w-4 h-4 mr-2 rounded bg-stellaris-dark border-gray-600 text-green-500 focus:ring-2 focus:ring-green-500">
                    <span class="text-gray-300">Starting</span>
                </label>
                <label class="inline-flex items-center cursor-pointer hover:text-white transition-colors">
                    <input type="checkbox" id="show-rare" checked onchange="filterTechs()"
                        class="w-4 h-4 mr-2 rounded bg-stellaris-dark border-gray-600 text-purple-500 focus:ring-2 focus:ring-purple-500">
                    <span class="text-gray-300">Rare</span>
                </label>
                <label class="inline-flex items-center cursor-pointer hover:text-white transition-colors">
                    <input type="checkbox" id="show-dangerous" checked onchange="filterTechs()"
                        class="w-4 h-4 mr-2 rounded bg-stellaris-dark border-gray-600 text-red-500 focus:ring-2 focus:ring-red-500">
                    <span class="text-gray-300">Dangerous</span>
                </label>
            </div>
        </div>

        <div class="bg-stellaris-panel/60 p-4 rounded-lg mb-6 border border-gray-700/50">
            <div class="flex flex-wrap items-center justify-between gap-4">
                <div class="flex flex-wrap items-center gap-6 text-sm">
                    <div class="flex items-center gap-2">
                        <span class="inline-block w-4 h-4 rounded border-2 border-yellow-400 bg-yellow-400/20"></span>
                        <span class="text-gray-300">Selected</span>
                    </div>
                    <div class="flex items-center gap-2">
                        <span class="inline-block w-4 h-4 rounded border-2 border-green-400"></span>
                        <span class="text-gray-300">Prerequisites</span>
                    </div>
                    <div class="flex items-center gap-2 text-gray-400">
                        <span>Total: <span id="total-count" class="font-semibold text-stellaris-blue">0</span></span>
                        <span class="mx-1">|</span>
                        <span>Visible: <span id="visible-count" class="font-semibold text-stellaris-blue">0</span></span>
                    </div>
                </div>
                <button id="clear-selection" onclick="clearSelection()"
                    class="hidden px-4 py-2 bg-stellaris-blue/20 hover:bg-stellaris-blue/30 border border-stellaris-blue/50 hover:border-stellaris-blue text-stellaris-blue font-semibold rounded transition-all duration-200 text-sm">
                    Clear Selection
                </button>
            </div>
        </div>

        <div id="tree-container" class="relative overflow-x-auto p-6 bg-stellaris-panel/40 rounded-lg border border-gray-700/50 min-h-[600px]"></div>
    </div>

    <script>
        let technologies = [];
        let localizations = {};
        let languages = [];
        let metadata = null;
        let selectedTech = null;
        let currentLanguage = 'english';

        // Load data from multiple JSON files
        async function loadData() {
            try {
                // Load metadata first
                const metaResponse = await fetch('metadata.json');
                if (!metaResponse.ok) {
                    throw new Error('Failed to load metadata');
                }
                metadata = await metaResponse.json();

                // Load localizations
                const locResponse = await fetch('localizations.json');
                if (!locResponse.ok) {
                    throw new Error('Failed to load localizations');
                }
                const locData = await locResponse.json();
                localizations = locData.localizations;
                languages = locData.languages;

                // Load technologies from each area
                technologies = [];
                const areaPromises = metadata.areas.map(async (area) => {
                    const response = await fetch('technologies-' + area.toLowerCase() + '.json');
                    if (!response.ok) {
                        throw new Error('Failed to load technologies for area: ' + area);
                    }
                    const data = await response.json();
                    technologies.push(...data.technologies);
                });

                await Promise.all(areaPromises);

                // Update total count
                document.getElementById('total-count').textContent = technologies.length;

                // Populate filters
                populateFilters();

                // Render the tree
                renderTree();
            } catch (error) {
                console.error('Error loading data:', error);
                document.getElementById('tree-container').innerHTML =
                    '<div class="text-center text-red-400 p-8">Error loading technology data: ' + error.message + '<br>Please make sure all JSON files are in the same directory as this HTML file.</div>';
            }
        }

        // Get localized name for a technology
        function getLocalizedName(tech) {
            if (localizations[tech.key] && localizations[tech.key][currentLanguage] && localizations[tech.key][currentLanguage].name) {
                return localizations[tech.key][currentLanguage].name;
            }
            return tech.name; // Fallback to formatted key
        }

        // Get localized description for a technology
        function getLocalizedDescription(tech) {
            if (localizations[tech.key] && localizations[tech.key][currentLanguage] && localizations[tech.key][currentLanguage].description) {
                return localizations[tech.key][currentLanguage].description;
            }
            return ''; // No description available
        }

        // Change the display language
        function changeLanguage() {
            const languageSelect = document.getElementById('language-select');
            currentLanguage = languageSelect.value;
            renderTree(); // Re-render with new language
        }

        function populateFilters() {
            // Populate language selector
            const languageSelect = document.getElementById('language-select');
            languageSelect.innerHTML = ''; // Clear loading option

            if (languages && languages.length > 0) {
                languages.forEach((lang, index) => {
                    const option = document.createElement('option');
                    option.value = lang;
                    // Capitalize and format language name
                    option.textContent = lang.charAt(0).toUpperCase() + lang.slice(1).replace('_', ' ');
                    languageSelect.appendChild(option);
                });
                // Default to english if available, otherwise first language
                if (languages.includes('english')) {
                    languageSelect.value = 'english';
                    currentLanguage = 'english';
                } else {
                    currentLanguage = languages[0];
                }
            } else {
                const option = document.createElement('option');
                option.value = '';
                option.textContent = 'No localizations available';
                languageSelect.appendChild(option);
            }

            // Populate area filter
            const areaFilter = document.getElementById('area-filter');
            metadata.areas.forEach(area => {
                const option = document.createElement('option');
                option.value = area;
                option.textContent = area;
                areaFilter.appendChild(option);
            });

            // Populate tier filter
            const tierFilter = document.getElementById('tier-filter');
            metadata.tiers.forEach(tier => {
                const option = document.createElement('option');
                option.value = tier;
                option.textContent = 'Tier ' + tier;
                tierFilter.appendChild(option);
            });
        }

        function renderTree() {
            if (!metadata) return;

            const container = document.getElementById('tree-container');
            const maxLevel = metadata.maxLevel;

            // Group technologies by level
            const levels = {};
            for (let i = 0; i <= maxLevel; i++) {
                levels[i] = [];
            }

            technologies.forEach(tech => {
                levels[tech.level].push(tech);
            });

            // Render columns for each level
            container.innerHTML = '';
            for (let i = 0; i <= maxLevel; i++) {
                if (levels[i].length === 0) continue;

                const column = document.createElement('div');
                column.className = 'inline-block align-top mr-8 min-w-[220px]';

                const header = document.createElement('h3');
                header.className = 'text-stellaris-blue text-base font-bold mb-3 pb-2 border-b border-stellaris-blue/30 sticky top-0 bg-stellaris-panel/80 backdrop-blur-sm z-10';
                header.textContent = 'Level ' + i;
                column.appendChild(header);

                levels[i].forEach(tech => {
                    const node = createTechNode(tech);
                    column.appendChild(node);
                });

                container.appendChild(column);
            }

            updateVisibleCount();
        }

        function createTechNode(tech) {
            const node = document.createElement('div');
            node.className = 'tech-card rounded-lg p-3 mb-3 cursor-pointer transition-all duration-200 relative';
            node.id = 'tech-' + tech.key;
            node.setAttribute('data-key', tech.key);
            node.setAttribute('data-area', tech.area);
            node.setAttribute('data-tier', tech.tier);
            node.setAttribute('data-start', tech.isStartTech);
            node.setAttribute('data-rare', tech.isRare);
            node.setAttribute('data-dangerous', tech.isDangerous);

            // Add area-specific border
            if (tech.area) {
                node.classList.add('area-' + tech.area.toLowerCase());
            }

            // Get localized name and description
            const localizedName = getLocalizedName(tech);
            const localizedDesc = getLocalizedDescription(tech);

            // Add tooltip with description if available
            if (localizedDesc) {
                node.title = localizedDesc;
            }

            let badges = '';
            if (tech.isStartTech) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-green-600/80 text-white border border-green-500/50">Start</span>';
            if (tech.isRare) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-purple-600/80 text-white border border-purple-500/50">Rare</span>';
            if (tech.isDangerous) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-red-600/80 text-white border border-red-500/50">Dangerous</span>';
            if (tech.isEvent) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-yellow-600/80 text-white border border-yellow-500/50">Event</span>';
            if (tech.isReverse) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-cyan-600/80 text-white border border-cyan-500/50">Reverse Eng.</span>';
            if (tech.isRepeatable) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-blue-600/80 text-white border border-blue-500/50">Repeatable' + (tech.levels > 0 ? ' (' + tech.levels + ')' : '') + '</span>';
            if (tech.isGestalt) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-indigo-600/80 text-white border border-indigo-500/50">Gestalt</span>';
            if (tech.isMegacorp) badges += '<span class="inline-block px-1.5 py-0.5 rounded text-xs mr-1 bg-amber-600/80 text-white border border-amber-500/50">Megacorp</span>';

            const areaColors = {
                'physics': { bg: 'bg-stellaris-physics/20', text: 'text-stellaris-physics', border: 'border-stellaris-physics/50' },
                'society': { bg: 'bg-stellaris-society/20', text: 'text-stellaris-society', border: 'border-stellaris-society/50' },
                'engineering': { bg: 'bg-stellaris-engineering/20', text: 'text-stellaris-engineering', border: 'border-stellaris-engineering/50' }
            };

            const areaColor = areaColors[tech.area?.toLowerCase()] || { bg: 'bg-gray-600/20', text: 'text-gray-400', border: 'border-gray-600/50' };

            node.innerHTML =
                '<div class="font-semibold text-sm mb-2 text-white leading-tight">' + localizedName + '</div>' +
                '<div class="text-xs text-gray-400 space-y-1">' +
                    '<div class="flex items-center justify-between">' +
                        '<span class="text-gray-500">Cost:</span>' +
                        '<span class="font-medium text-gray-300">' + tech.cost + '</span>' +
                    '</div>' +
                    '<div class="flex items-center justify-between">' +
                        '<span class="text-gray-500">Tier:</span>' +
                        '<span class="font-medium text-gray-300">' + tech.tier + '</span>' +
                    '</div>' +
                    (tech.weight ? '<div class="flex items-center justify-between">' +
                        '<span class="text-gray-500">Weight:</span>' +
                        '<span class="font-medium text-gray-300">' + tech.weight + '</span>' +
                    '</div>' : '') +
                    (tech.sourceFile ? '<div class="flex items-center justify-between">' +
                        '<span class="text-gray-500">File:</span>' +
                        '<span class="font-medium text-gray-300 text-[10px] truncate" title="' + tech.sourceFile + '">' + tech.sourceFile + '</span>' +
                    '</div>' : '') +
                    (tech.area ? '<div class="mt-2"><span class="inline-block px-2 py-0.5 rounded text-xs ' + areaColor.bg + ' ' + areaColor.text + ' border ' + areaColor.border + '">' + tech.area + '</span></div>' : '') +
                    (badges ? '<div class="mt-2 flex flex-wrap gap-1">' + badges + '</div>' : '') +
                '</div>';

            node.onclick = () => selectTech(tech.key);
            return node;
        }

        function getAllPrerequisites(techKey, visited = new Set()) {
            if (visited.has(techKey)) return visited;
            visited.add(techKey);

            const tech = technologies.find(t => t.key === techKey);
            if (!tech) return visited;

            tech.prerequisites.forEach(prereqKey => {
                getAllPrerequisites(prereqKey, visited);
            });

            return visited;
        }

        function selectTech(key) {
            const tech = technologies.find(t => t.key === key);
            if (!tech) return;

            selectedTech = key;

            // Get all prerequisites recursively
            const prerequisitesToShow = getAllPrerequisites(key);

            // Hide all techs first, then show only selected and prerequisites
            document.querySelectorAll('[id^="tech-"]').forEach(node => {
                const nodeKey = node.getAttribute('data-key');

                // Remove all highlight classes
                node.classList.remove('!border-yellow-400', '!border-2', 'shadow-[0_0_20px_rgba(255,215,0,0.6)]');
                node.classList.remove('!border-green-400', 'shadow-[0_0_15px_rgba(95,202,95,0.4)]');
                node.style.borderColor = '';
                node.style.borderWidth = '';

                // Hide nodes that are not in the prerequisite chain
                if (!prerequisitesToShow.has(nodeKey)) {
                    node.classList.add('!hidden');
                } else {
                    node.classList.remove('!hidden');
                }
            });

            // Highlight selected tech
            const selectedNode = document.getElementById('tech-' + key);
            if (selectedNode) {
                selectedNode.classList.add('!border-yellow-400', '!border-2', 'shadow-[0_0_20px_rgba(255,215,0,0.6)]');
                selectedNode.style.borderColor = '#fbbf24';
                selectedNode.style.borderWidth = '2px';
            }

            // Highlight prerequisites (all except the selected one)
            prerequisitesToShow.forEach(prereqKey => {
                if (prereqKey !== key) {
                    const prereqNode = document.getElementById('tech-' + prereqKey);
                    if (prereqNode) {
                        prereqNode.classList.add('!border-green-400', '!border-2', 'shadow-[0_0_15px_rgba(95,202,95,0.4)]');
                        prereqNode.style.borderColor = '#5fca5f';
                        prereqNode.style.borderWidth = '2px';
                    }
                }
            });

            // Show the clear selection button
            document.getElementById('clear-selection').classList.remove('hidden');

            updateVisibleCount();
        }

        function clearSelection() {
            selectedTech = null;

            // Remove all highlight classes and show all techs
            document.querySelectorAll('[id^="tech-"]').forEach(node => {
                node.classList.remove('!border-yellow-400', '!border-2', 'shadow-[0_0_20px_rgba(255,215,0,0.6)]');
                node.classList.remove('!border-green-400', 'shadow-[0_0_15px_rgba(95,202,95,0.4)]');
                node.classList.remove('!hidden');
                node.style.borderColor = '';
                node.style.borderWidth = '';
            });

            // Hide the clear selection button
            document.getElementById('clear-selection').classList.add('hidden');

            // Reapply filters
            filterTechs();
        }

        function filterTechs() {
            // Don't apply filters if a tech is selected
            if (selectedTech) {
                return;
            }

            const searchTerm = document.getElementById('search').value.toLowerCase();
            const areaFilter = document.getElementById('area-filter').value;
            const tierFilter = document.getElementById('tier-filter').value;
            const showStart = document.getElementById('show-start').checked;
            const showRare = document.getElementById('show-rare').checked;
            const showDangerous = document.getElementById('show-dangerous').checked;

            document.querySelectorAll('[id^="tech-"]').forEach(node => {
                const key = node.getAttribute('data-key');
                const tech = technologies.find(t => t.key === key);
                if (!tech) return;

                let visible = true;

                // Search filter
                if (searchTerm && !tech.name.toLowerCase().includes(searchTerm) && !tech.key.toLowerCase().includes(searchTerm)) {
                    visible = false;
                }

                // Area filter
                if (areaFilter && tech.area !== areaFilter) {
                    visible = false;
                }

                // Tier filter
                if (tierFilter && tech.tier.toString() !== tierFilter) {
                    visible = false;
                }

                // Special tech filters
                if (!showStart && tech.isStartTech) visible = false;
                if (!showRare && tech.isRare) visible = false;
                if (!showDangerous && tech.isDangerous) visible = false;

                node.classList.toggle('hidden', !visible);
            });

            updateVisibleCount();
        }

        function updateVisibleCount() {
            const visibleNodes = document.querySelectorAll('[id^="tech-"]:not(.hidden):not(.\\!hidden)');
            document.getElementById('visible-count').textContent = visibleNodes.length;
        }

        // Initialize - Load data when page loads
        document.addEventListener('DOMContentLoaded', loadData);
    </script>
</body>
</html>`
