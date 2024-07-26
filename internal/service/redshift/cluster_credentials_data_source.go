// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_redshift_cluster_credentials", name="Cluster Credentials")
func dataSourceClusterCredentials() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterCredentialsRead,

		Schema: map[string]*schema.Schema{
			"auto_create": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
				Required: true,
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

func dataSourceClusterCredentialsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterID := d.Get(names.AttrClusterIdentifier).(string)
	input := &redshift.GetClusterCredentialsInput{
		AutoCreate:        aws.Bool(d.Get("auto_create").(bool)),
		ClusterIdentifier: aws.String(clusterID),
		DbUser:            aws.String(d.Get("db_user").(string)),
		DurationSeconds:   aws.Int64(int64(d.Get("duration_seconds").(int))),
	}

	if v, ok := d.GetOk("db_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.DbGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("db_name"); ok {
		input.DbName = aws.String(v.(string))
	}

	creds, err := conn.GetClusterCredentialsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster Credentials for Cluster (%s): %s", clusterID, err)
	}

	d.SetId(clusterID)

	d.Set("db_password", creds.DbPassword)
	d.Set("db_user", creds.DbUser)
	d.Set("expiration", aws.TimeValue(creds.Expiration).Format(time.RFC3339))

	return diags
}
