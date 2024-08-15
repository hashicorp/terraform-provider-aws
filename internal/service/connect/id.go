// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"fmt"
	"strings"
)

const lambdaFunctionAssociationIDSeparator = ","

func LambdaFunctionAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, lambdaFunctionAssociationIDSeparator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "",
			fmt.Errorf("unexpected format for ID (%q), expected instanceID"+lambdaFunctionAssociationIDSeparator+
				"functionARN", id)
	}

	return parts[0], parts[1], nil
}

func LambdaFunctionAssociationCreateResourceID(instanceID string, functionArn string) string {
	parts := []string{instanceID, functionArn}
	id := strings.Join(parts, lambdaFunctionAssociationIDSeparator)

	return id
}
