// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interceptors

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type infoAWSClient interface {
	ServicePackage(_ context.Context, name string) conns.ServicePackage
}

func InfoFromContext(ctx context.Context, c infoAWSClient) (conns.ServicePackage, string, string, string, *tftags.InContext, bool) {
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

			typeName := inContext.TypeName()
			if typeName == "" {
				resourceName = "aws_<service>_<thing>"
			}

			if tagsInContext, ok := tftags.FromContext(ctx); ok {
				return sp, serviceName, resourceName, typeName, tagsInContext, true
			}
		}
	}

	return nil, "", "", "", nil, false
}
