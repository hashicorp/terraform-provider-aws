// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	aclNameMaxLength            = 40
	clusterNameMaxLength        = 40
	parameterGroupNameMaxLength = 255
	snapshotNameMaxLength       = 255
	subnetGroupNameMaxLength    = 255
	userNameMaxLength           = 40
)

// validateResourceName returns a validation function applicable to all MemoryDB
// resource names.
//
// MemoryDB, similar to ElastiCache, allows upper-case names when creating
// resources, but then normalises them to lowercase on any subsequent read.
// This complicates Terraform state management, so disallow uppercase characters
// entirely.
func validateResourceName(maxLength int) schema.SchemaValidateFunc {
	return validation.All(
		validateResourceNamePrefix(maxLength),
		validation.StringDoesNotMatch(
			regexache.MustCompile(`[-]$`),
			"The name may not end with a hyphen."),
	)
}

func validateResourceNamePrefix(maxLength int) schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, maxLength),
		validation.StringDoesNotMatch(
			regexache.MustCompile(`[-][-]`),
			"The name may not contain two consecutive hyphens."),
		validation.StringMatch(
			regexache.MustCompile(`^[0-9a-z-]+$`),
			"Only lowercase alphanumeric characters and hyphens are allowed."),
	)
}
