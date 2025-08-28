// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_location_tracker_association", name="Tracker Association")
func DataSourceTrackerAssociation() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrackerAssociationRead,

		Schema: map[string]*schema.Schema{
			"consumer_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tracker_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
		},
	}
}

const (
	DSNameTrackerAssociation = "Tracker Association Data Source"
)

func dataSourceTrackerAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	consumerArn := d.Get("consumer_arn").(string)
	trackerName := d.Get("tracker_name").(string)
	id := fmt.Sprintf("%s|%s", trackerName, consumerArn)

	err := FindTrackerAssociationByTrackerNameAndConsumerARN(ctx, conn, trackerName, consumerArn)
	if err != nil {
		return create.AppendDiagError(diags, names.Location, create.ErrActionReading, DSNameTrackerAssociation, id, err)
	}

	d.SetId(id)

	return diags
}
