// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_redshiftserverless_credentials", name="Credentials")
func dataSourceCredentials() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCredentialsRead,

		Schema: map[string]*schema.Schema{
			"workgroup_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"duration_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      900,
				ValidateFunc: validation.IntBetween(900, 3600),
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCredentialsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	workgroupName := d.Get("workgroup_name").(string)
	input := &redshiftserverless.GetCredentialsInput{
		WorkgroupName:   aws.String(workgroupName),
		DurationSeconds: aws.Int32(int32(d.Get("duration_seconds").(int))),
	}

	if v, ok := d.GetOk("db_name"); ok {
		input.DbName = aws.String(v.(string))
	}

	creds, err := conn.GetCredentials(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Credentials for Workgroup (%s): %s", workgroupName, err)
	}

	d.SetId(workgroupName)

	d.Set("db_password", creds.DbPassword)
	d.Set("db_user", creds.DbUser)
	d.Set("expiration", aws.ToTime(creds.Expiration).Format(time.RFC3339))

	return diags
}
