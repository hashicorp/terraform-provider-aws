// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_event_connection", name="Connection)
func dataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := findConnectionByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Connection (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, output.ConnectionArn)
	d.Set("authorization_type", output.AuthorizationType)
	d.Set(names.AttrName, output.Name)
	d.Set("secret_arn", output.SecretArn)

	return diags
}
