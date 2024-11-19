// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// SetDefaultTagsConfig is only intended for use in tests
func SetDefaultTagsConfig(client *AWSClient, d *tftags.DefaultConfig) {
	client.defaultTagsConfig = d
}

// SetIgnoreTagsConfig is only intended for use in tests
func SetIgnoreTagsConfig(client *AWSClient, i *tftags.IgnoreConfig) {
	client.ignoreTagsConfig = i
}
