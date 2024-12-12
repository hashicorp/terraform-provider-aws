// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elasticache_user", name="User")
func dataSourceUser() *schema.Resource {
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
						names.AttrType: {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			names.AttrEngine: {
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
			names.AttrUserName: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	user, err := findUserByID(ctx, conn, d.Get("user_id").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ElastiCache User", err))
	}

	d.SetId(aws.ToString(user.UserId))
	d.Set("access_string", user.AccessString)
	if v := user.Authentication; v != nil {
		tfMap := map[string]interface{}{
			"password_count": aws.ToInt32(v.PasswordCount),
			names.AttrType:   string(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	}
	d.Set(names.AttrEngine, user.Engine)
	d.Set("user_id", user.UserId)
	d.Set(names.AttrUserName, user.UserName)

	return diags
}
