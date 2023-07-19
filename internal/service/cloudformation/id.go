// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"fmt"
	"strings"
)

const stackSetInstanceResourceIDSeparator = ","

func StackSetInstanceCreateResourceID(stackSetName, accountID, region string) string {
	parts := []string{stackSetName, accountID, region}
	id := strings.Join(parts, stackSetInstanceResourceIDSeparator)

	return id
}

func StackSetInstanceParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, stackSetInstanceResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected STACKSETNAME%[2]sACCOUNTID%[2]sREGION", id, stackSetInstanceResourceIDSeparator)
}
