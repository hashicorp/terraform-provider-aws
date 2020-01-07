package check

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	FileExtensionHtmlMarkdown = `.html.markdown`
	FileExtensionHtmlMd       = `.html.md`
	FileExtensionMarkdown     = `.markdown`
	FileExtensionMd           = `.md`
)

var ValidLegacyFileExtensions = []string{
	FileExtensionHtmlMarkdown,
	FileExtensionHtmlMd,
	FileExtensionMarkdown,
	FileExtensionMd,
}

var ValidRegistryFileExtensions = []string{
	FileExtensionMd,
}

func LegacyFileExtensionCheck(path string) error {
	fileExtension := GetFileExtension(path)

	if !IsValidLegacyFileExtension(fileExtension) {
		return fmt.Errorf("invalid file extension (%s), valid values: %v", fileExtension, ValidLegacyFileExtensions)
	}

	return nil
}

func RegistryFileExtensionCheck(path string) error {
	fileExtension := GetFileExtension(path)

	if !IsValidRegistryFileExtension(fileExtension) {
		return fmt.Errorf("invalid file extension (%s), valid values: %v", fileExtension, ValidRegistryFileExtensions)
	}

	return nil
}

// GetFileExtension fetches file extensions including those with multiple periods.
// This is a replacement for filepath.Ext(), which only returns the final period and extension.
func GetFileExtension(path string) string {
	filename := filepath.Base(path)

	if filename == "." {
		return ""
	}

	dotIndex := strings.IndexByte(filename, '.')

	if dotIndex > 0 {
		return filename[dotIndex:]
	}

	return filename
}

func IsValidLegacyFileExtension(fileExtension string) bool {
	for _, validLegacyFileExtension := range ValidLegacyFileExtensions {
		if fileExtension == validLegacyFileExtension {
			return true
		}
	}

	return false
}

func IsValidRegistryFileExtension(fileExtension string) bool {
	for _, validRegistryFileExtension := range ValidRegistryFileExtensions {
		if fileExtension == validRegistryFileExtension {
			return true
		}
	}

	return false
}

// TrimFileExtension removes file extensions including those with multiple periods.
func TrimFileExtension(path string) string {
	filename := filepath.Base(path)

	if filename == "." {
		return ""
	}

	dotIndex := strings.IndexByte(filename, '.')

	if dotIndex > 0 {
		return filename[:dotIndex]
	}

	return filename
}
