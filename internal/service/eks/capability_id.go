// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"
	"strings"
)

const capabilityResourceIDSeparator = ":"

func capabilityCreateResourceID(clusterName, capabilityName string) string {
	return fmt.Sprintf("%s%s%s", clusterName, capabilityResourceIDSeparator, capabilityName)
}

func capabilityParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, capabilityResourceIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CLUSTER_NAME%[2]sCAPABILITY_NAME", id, capabilityResourceIDSeparator)
}
