// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegionalSingleton(ctx context.Context, rd *schema.ResourceData) error {
	if region, ok := rd.GetOk(names.AttrRegion); ok {
		if region != rd.Id() {
			return fmt.Errorf("the region passed for import %q does not match the region %q in the ID", region, rd.Id())
		}
	} else {
		rd.Set(names.AttrRegion, rd.Id())
	}

	return nil
}
