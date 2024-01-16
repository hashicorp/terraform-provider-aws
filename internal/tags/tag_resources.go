// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"fmt"
	"strings"
)

const (
	// Separator used in resource identifiers
	resourceIDSeparator = `,`
)

// GetResourceID parses a given resource identifier for tag identifier and tag key.
func GetResourceID(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, resourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid resource identifier (%[1]s), expected ID%[2]sKEY", resourceID, resourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

// SetResourceID creates a resource identifier given a tag identifier and a tag key.
func SetResourceID(identifier string, key string) string {
	parts := []string{identifier, key}
	resourceID := strings.Join(parts, resourceIDSeparator)

	return resourceID
}
