// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mpa

import (
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_mpa_identity_source", sweepIdentitySources)
}
