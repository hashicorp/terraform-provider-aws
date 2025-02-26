// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package interceptors

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func InfoFromContext(ctx context.Context, c *conns.AWSClient) (conns.ServicePackage, string, string, *tftags.InContext, bool) {
	if inContext, ok := conns.FromContext(ctx); ok {
		if sp := c.ServicePackage(ctx, inContext.ServicePackageName()); sp != nil {
			serviceName, err := names.HumanFriendly(sp.ServicePackageName())
			if err != nil {
				serviceName = "<service>"
			}

			resourceName := inContext.ResourceName()
			if resourceName == "" {
				resourceName = "<thing>"
			}

			if tagsInContext, ok := tftags.FromContext(ctx); ok {
				return sp, serviceName, resourceName, tagsInContext, true
			}
		}
	}

	return nil, "", "", nil, false
}
