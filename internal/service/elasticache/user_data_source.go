// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_elasticache_user")
func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authentication_mode": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password_count": {
							Optional: true,
							Type:     schema.TypeInt,
						},
						"type": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"no_password_required": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"passwords": {
				Type:      schema.TypeSet,
				Optional:  true,
				Elem:      &schema.Schema{Type: schema.TypeString},
				Set:       schema.HashString,
				Sensitive: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	user, err := FindUserByID(ctx, conn, d.Get("user_id").(string))
	if tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): Not found. Please change your search criteria and try again: %s", d.Get("user_id").(string), err)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Cluster (%s): %s", d.Get("user_id").(string), err)
	}

	d.SetId(aws.StringValue(user.UserId))

	d.Set("access_string", user.AccessString)

	if v := user.Authentication; v != nil {
		authenticationMode := map[string]interface{}{
			"password_count": aws.Int64Value(v.PasswordCount),
			"type":           aws.StringValue(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{authenticationMode}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	}

	d.Set("engine", user.Engine)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return diags
}
