// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

// instanceProfileARNToName converts Amazon Resource Name (ARN) to Name.
func instanceProfileARNToName(inputARN string) (string, error) {
	const (
		arnSeparator                  = "/"
		arnService                    = "iam"
		instanceProfileResourcePrefix = "instance-profile"
	)

	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, arnSeparator)

	if actual, expected := len(resourceParts), 2; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	if actual, expected := resourceParts[0], instanceProfileResourcePrefix; actual != expected {
		return "", fmt.Errorf("expected resource prefix %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	return resourceParts[len(resourceParts)-1], nil
}
