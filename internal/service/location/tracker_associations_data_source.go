// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_location_tracker_associations", name="Tracker Associations")
func DataSourceTrackerAssociations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTrackerAssociationsRead,

		Schema: map[string]*schema.Schema{
			"consumer_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
	DSNameTrackerAssociations = "Tracker Associations Data Source"
)

func dataSourceTrackerAssociationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	name := d.Get("tracker_name").(string)

	in := &location.ListTrackerConsumersInput{
		TrackerName: aws.String(name),
	}

	var arns []string

	pages := location.NewListTrackerConsumersPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return create.AppendDiagError(diags, names.Location, create.ErrActionReading, DSNameTrackerAssociations, name, err)
		}

		arns = append(arns, page.ConsumerArns...)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))
	d.Set("consumer_arns", arns)

	return diags
}
