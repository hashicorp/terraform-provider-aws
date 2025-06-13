package types

import "testing"

func TestIsAWSDataCatalogID(t *testing.T) {
	// Test cases for IsAWSDataCatalogID function
	// The function checks for a valid AWS Data Catalog ID format, which is either:
	// 1. A 12-digit account ID (e.g., "123456789012")
	// 2. A string in the format "s3tablescatalog/my-catalog"

	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		// Valid 12-digit account ID
		{"123456789012", true},
		// Valid s3tablescatalog format, min and max length
		{"s3tablescatalog/abc", true},
		{"s3tablescatalog/abc-def-ghi", true},
		{"s3tablescatalog/a23", true},
		{"s3tablescatalog/" + string(make([]byte, 63)), false}, // too long
		// Valid: max length (63 chars)
		{"s3tablescatalog/abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabcdeabc", true}, // 63 chars

		// Invalid: less than 12 digits
		{"12345678901", false},
		// Invalid: more than 12 digits
		{"1234567890123", false},
		// Invalid: contains letters in account ID
		{"12345678901a", false},
		// Invalid: wrong prefix after colon
		{"somethingelse/my-catalog", false},
		// Invalid: missing catalog name
		{"s3tablescatalog/", false},
		// Invalid: empty string
		{"", false},
		// Invalid: only s3tablescatalog part
		{"s3tablescatalog/my-catalog", false},

		// Invalid: underscores
		{"s3tablescatalog/my_catalog", false},
		// Invalid: periods
		{"s3tablescatalog/my.catalog", false},
		// Invalid: starts with hyphen
		{"s3tablescatalog/-abc", false},
		// Invalid: ends with hyphen
		{"s3tablescatalog/abc-", false},
		// Invalid: starts with reserved prefix
		{"s3tablescatalog/xn--bucket", false},
		{"s3tablescatalog/sthree-bucket", false},
		{"s3tablescatalog/amzn-s3-demo-bucket", false},
		// Invalid: ends with reserved suffix
		{"s3tablescatalog/bucket-s3alias", false},
		{"s3tablescatalog/bucket--ol-s3", false},
		{"s3tablescatalog/bucket--x-s3", false},
		{"s3tablescatalog/bucket--table-s3", false},
		// Invalid: too short
		{"s3tablescatalog/ab", false},
		// Invalid: too long (64 chars)
		{"s3tablescatalog/abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl", false},
		// Invalid: uppercase letters
		{"s3tablescatalog/Mycatalog", false},
		// Invalid: spaces
		{"s3tablescatalog/my catalog", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsAWSDataCatalogID(tt.input)
			if result != tt.expected {
				t.Errorf("IsAWSDataCatalogID(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
