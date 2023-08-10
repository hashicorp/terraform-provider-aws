// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
						"service_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn(ctx)

	name := d.Get("name").(string)
	app, err := FindApplicationByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application (%s): %s", name, err)
	}

	d.SetId(name)
	if err := d.Set("appversion_lifecycle", flattenApplicationResourceLifecycleConfig(app.ResourceLifecycleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting appversion_lifecycle: %s", err)
	}
	d.Set("arn", app.ApplicationArn)
	d.Set("description", app.Description)
	d.Set("name", app.ApplicationName)

	return diags
}
