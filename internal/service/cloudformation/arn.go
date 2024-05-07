// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	arnSeparator = "/"
	arnService   = "cloudformation"
)

// typeVersionARNToTypeARNAndVersionID converts Type Version Amazon Resource Name (ARN) to Type ARN and Version ID.
//
// Given input: arn:aws:cloudformation:us-west-2:123456789012:type/resource/HashiCorp-TerraformAwsProvider-TfAccTestzwv6r2i7/00000001,
// returns arn:aws:cloudformation:us-west-2:123456789012:type/resource/HashiCorp-TerraformAwsProvider-TfAccTestzwv6r2i7 and 00000001.
func typeVersionARNToTypeARNAndVersionID(inputARN string) (string, string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", "", fmt.Errorf("parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, arnSeparator)

	if actual, expected := len(resourceParts), 4; actual != expected {
		return "", "", fmt.Errorf("expected %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	if actual, expected := resourceParts[0], names.AttrType; actual != expected {
		return "", "", fmt.Errorf("expected resource prefix %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	outputTypeARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(resourceParts[:3], arnSeparator),
	}.String()

	return outputTypeARN, resourceParts[len(resourceParts)-1], nil
}
