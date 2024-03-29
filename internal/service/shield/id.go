// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"fmt"
	"strings"
)

const protectionHealthCheckAssociationResourceIDSeparator = "+"

func ProtectionHealthCheckAssociationCreateResourceID(protectionId, healthCheckArn string) string {
	parts := []string{protectionId, healthCheckArn}
	id := strings.Join(parts, protectionHealthCheckAssociationResourceIDSeparator)

	return id
}

func ProtectionHealthCheckAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, protectionHealthCheckAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected PROTECTIONID%[2]sHEALTHCHECKARN", id, protectionHealthCheckAssociationResourceIDSeparator)
}
