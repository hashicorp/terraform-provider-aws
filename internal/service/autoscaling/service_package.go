// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

func (p *servicePackage) Actions(ctx context.Context) []*types.ServicePackageAction {
	return []*types.ServicePackageAction{
		{
			Factory:  newSetDesiredCapacityAction,
			TypeName: "aws_autoscaling_set_desired_capacity",
			Name:     "Set Desired Capacity",
		},
	}
}
