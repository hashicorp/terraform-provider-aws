// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elastic_beanstalk_application")
func DataSourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceApplicationRead,

		Schema: map[string]*schema.Schema{
			"appversion_lifecycle": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_source_from_s3": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_age_in_days": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrServiceRole: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	app, err := FindApplicationByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application (%s): %s", name, err)
	}

	d.SetId(name)
	if err := d.Set("appversion_lifecycle", flattenApplicationResourceLifecycleConfig(app.ResourceLifecycleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting appversion_lifecycle: %s", err)
	}
	d.Set(names.AttrARN, app.ApplicationArn)
	d.Set(names.AttrDescription, app.Description)
	d.Set(names.AttrName, app.ApplicationName)

	return diags
}
