// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func (p *servicePackage) Actions(ctx context.Context) []*inttypes.ServicePackageAction {
	return []*inttypes.ServicePackageAction{
		{
			Factory:  newUpdateServiceAction,
			TypeName: "aws_ecs_update_service",
			Name:     "Update Service",
		},
	}
}
