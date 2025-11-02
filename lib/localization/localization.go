package localization

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LocalizationData stores translations for all languages
type LocalizationData struct {
	Languages map[string]*LanguageData // key: language code (e.g., "english", "german")
}

// LanguageData stores translations for a specific language
type LanguageData struct {
	Translations map[string]string // key: translation key, value: localized text
}

// LocalizationParser parses Stellaris localization files
type LocalizationParser struct {
	data *LocalizationData
}

// NewLocalizationParser creates a new localization parser
func NewLocalizationParser() *LocalizationParser {
	return &LocalizationParser{
		data: &LocalizationData{
			Languages: make(map[string]*LanguageData),
		},
	}
}

// ParseDirectory parses all localization files in the given directory and subdirectories
func (p *LocalizationParser) ParseDirectory(localizationDir string) error {
	// Check if directory exists
	if _, err := os.Stat(localizationDir); os.IsNotExist(err) {
		return fmt.Errorf("localization directory does not exist: %s", localizationDir)
	}

	// Walk through all subdirectories
	err := filepath.Walk(localizationDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a file or not a YAML file
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".yml") {
			return nil
		}

		// Extract language code from file name
		// Format: *_l_<language>.yml
		fileName := filepath.Base(path)
		languagePattern := regexp.MustCompile(`_l_(\w+)\.yml$`)
		matches := languagePattern.FindStringSubmatch(fileName)

		if len(matches) < 2 {
			// Skip files that don't match the pattern
			return nil
		}

		language := matches[1]

		// Parse the file
		if err := p.parseFile(path, language); err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: failed to parse localization file %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk localization directory: %w", err)
	}

	return nil
}

// parseFile parses a single localization YAML file
func (p *LocalizationParser) parseFile(filePath string, language string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Ensure language data exists
	if p.data.Languages[language] == nil {
		p.data.Languages[language] = &LanguageData{
			Translations: make(map[string]string),
		}
	}

	langData := p.data.Languages[language]
	scanner := bufio.NewScanner(file)

	// Pattern to match localization entries with optional version number:
	// Format 1: key:version "value" (e.g., tech_basic_science_lab_1:0 "Scientific Method")
	// Format 2: key: "value" (e.g., tech_basic_science_lab_1: "Scientific Method")
	entryPattern1 := regexp.MustCompile(`^\s*([a-zA-Z0-9_]+):\d+\s+"(.+)"`)
	entryPattern2 := regexp.MustCompile(`^\s*([a-zA-Z0-9_]+):\s*"(.+)"`)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines, comments, and language header
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), "l_") {
			continue
		}

		// Try pattern 1 first (with version number)
		matches := entryPattern1.FindStringSubmatch(line)
		if len(matches) < 3 {
			// Try pattern 2 (without version number)
			matches = entryPattern2.FindStringSubmatch(line)
		}

		if len(matches) >= 3 {
			key := matches[1]
			value := matches[2]

			// Unescape quotes and other special characters
			value = strings.ReplaceAll(value, `\"`, `"`)
			value = strings.ReplaceAll(value, `\n`, "\n")

			langData.Translations[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// GetLocalizedName returns the localized name for a technology key
func (p *LocalizationParser) GetLocalizedName(techKey string, language string) string {
	if langData, ok := p.data.Languages[language]; ok {
		if name, ok := langData.Translations[techKey]; ok {
			return name
		}
	}
	return ""
}

// GetLocalizedDescription returns the localized description for a technology key
func (p *LocalizationParser) GetLocalizedDescription(techKey string, language string) string {
	descKey := techKey + "_desc"
	if langData, ok := p.data.Languages[language]; ok {
		if desc, ok := langData.Translations[descKey]; ok {
			return desc
		}
	}
	return ""
}

// GetAvailableLanguages returns a list of all parsed languages
func (p *LocalizationParser) GetAvailableLanguages() []string {
	languages := make([]string, 0, len(p.data.Languages))
	for lang := range p.data.Languages {
		languages = append(languages, lang)
	}
	return languages
}

// GetAllTranslations returns all translations for all languages
// Returns a map of language -> technology key -> translations (name and desc)
func (p *LocalizationParser) GetAllTranslations() map[string]map[string]map[string]string {
	result := make(map[string]map[string]map[string]string)

	for lang, langData := range p.data.Languages {
		result[lang] = make(map[string]map[string]string)

		// Group translations by tech key
		for key, value := range langData.Translations {
			// Check if this is a description key
			if strings.HasSuffix(key, "_desc") {
				techKey := strings.TrimSuffix(key, "_desc")
				if result[lang][techKey] == nil {
					result[lang][techKey] = make(map[string]string)
				}
				result[lang][techKey]["desc"] = value
			} else {
				// It's a name key
				if result[lang][key] == nil {
					result[lang][key] = make(map[string]string)
				}
				result[lang][key]["name"] = value
			}
		}
	}

	return result
}

// GetData returns the raw localization data
func (p *LocalizationParser) GetData() *LocalizationData {
	return p.data
}
