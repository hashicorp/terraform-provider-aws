// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

const (
	arnSeparator = "/"
	arnService   = "globalaccelerator"
)

// endpointGroupARNToListenerARN converts an endpoint group ARN to a listener ARN.
// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsglobalaccelerator.html#awsglobalaccelerator-resources-for-iam-policies.
func endpointGroupARNToListenerARN(inputARN string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, arnSeparator)

	if actual, expected := len(resourceParts), 6; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(resourceParts[0:4], arnSeparator),
	}.String()

	return outputARN, nil
}

// listenerOrEndpointGroupARNToAcceleratorARN converts a listener or endpoint group ARN to an accelerator ARN.
// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsglobalaccelerator.html#awsglobalaccelerator-resources-for-iam-policies.
func listenerOrEndpointGroupARNToAcceleratorARN(inputARN string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, arnSeparator)

	if actual, expected := len(resourceParts), 4; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(resourceParts[0:2], arnSeparator),
	}.String()

	return outputARN, nil
}
