// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workmail

import "github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"

func RegisterSweepers() {
	awsv2.Register("aws_workmail_organization", sweepOrganizations)
}
