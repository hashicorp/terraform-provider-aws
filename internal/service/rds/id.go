// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"fmt"
	"strings"
)

const clusterRoleAssociationResourceIDSeparator = ","

func ClusterRoleAssociationCreateResourceID(dbClusterID, roleARN string) string {
	parts := []string{dbClusterID, roleARN}
	id := strings.Join(parts, clusterRoleAssociationResourceIDSeparator)

	return id
}

func ClusterRoleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, clusterRoleAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBCLUSTERID%[2]sROLEARN", id, clusterRoleAssociationResourceIDSeparator)
}
