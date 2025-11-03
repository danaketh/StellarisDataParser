package generator

import (
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG format
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lukegb/dds" // Register DDS format
)

// IconConverter handles conversion of DDS icons to PNG format
type IconConverter struct {
	gameDir   string
	outputDir string
}

// NewIconConverter creates a new icon converter
func NewIconConverter(gameDir, outputDir string) *IconConverter {
	return &IconConverter{
		gameDir:   gameDir,
		outputDir: outputDir,
	}
}

// ConvertIcon converts a single icon from DDS to PNG
// iconName is the base name without extension (e.g., "tech_lasers")
func (ic *IconConverter) ConvertIcon(iconName string) error {
	// Look for the icon in multiple locations
	possiblePaths := []string{
		filepath.Join(ic.gameDir, "gfx", "interface", "icons", "technologies", iconName+".dds"),
		filepath.Join(ic.gameDir, "gfx", "interface", "icons", "technologies", iconName+".png"),
		filepath.Join(ic.gameDir, "gfx", "interface", "icons", "technologies", iconName+".jpg"),
	}

	var sourcePath string
	var sourceExt string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			sourcePath = path
			sourceExt = filepath.Ext(path)
			break
		}
	}

	if sourcePath == "" {
		// Icon file not found - this is not necessarily an error
		// as some mods or DLCs might be missing
		return nil
	}

	// If already PNG or JPG, just copy it
	outputPath := filepath.Join(ic.outputDir, "icons", iconName+".png")
	if sourceExt == ".png" || sourceExt == ".jpg" {
		return ic.copyFile(sourcePath, outputPath)
	}

	// Convert DDS to PNG
	return ic.convertDDSToPNG(sourcePath, outputPath)
}

// convertDDSToPNG converts a DDS file to PNG format
func (ic *IconConverter) convertDDSToPNG(sourcePath, outputPath string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Decode image (DDS decoder is registered)
	img, format, err := image.Decode(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Encode as PNG
	if err := png.Encode(outputFile, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func (ic *IconConverter) copyFile(src, dst string) error {
	// Create output directory if needed
	outputDir := filepath.Dir(dst)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ConvertIcons converts all icons for the given technology keys
func (ic *IconConverter) ConvertIcons(iconNames []string) (int, error) {
	converted := 0
	errors := []string{}

	for _, iconName := range iconNames {
		if err := ic.ConvertIcon(iconName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", iconName, err))
		} else {
			// Check if file was actually created
			outputPath := filepath.Join(ic.outputDir, "icons", iconName+".png")
			if _, err := os.Stat(outputPath); err == nil {
				converted++
			}
		}
	}

	if len(errors) > 0 {
		return converted, fmt.Errorf("failed to convert some icons:\n%s", strings.Join(errors, "\n"))
	}

	return converted, nil
}
