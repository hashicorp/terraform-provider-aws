// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"fmt"
	"strings"
)

const domainAssociationResourceIDSeparator = "/"

func DomainAssociationCreateResourceID(appID, domainName string) string {
	parts := []string{appID, domainName}
	id := strings.Join(parts, domainAssociationResourceIDSeparator)

	return id
}

func DomainAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, domainAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sDOMAINNAME", id, domainAssociationResourceIDSeparator)
}
