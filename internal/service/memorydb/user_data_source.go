// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_memorydb_user", name="User")
// @Tags(identifierAttribute="arn")
func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_mode": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"minimum_engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrUserName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	userName := d.Get(names.AttrUserName).(string)
	user, err := findUserByName(ctx, conn, userName)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MemoryDB User", err))
	}

	d.SetId(aws.ToString(user.Name))
	d.Set("access_string", user.AccessString)
	d.Set(names.AttrARN, user.ARN)
	if v := user.Authentication; v != nil {
		tfMap := map[string]any{
			"password_count": aws.ToInt32(v.PasswordCount),
			names.AttrType:   v.Type,
		}

		if err := d.Set("authentication_mode", []any{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting authentication_mode: %s", err)
		}
	}
	d.Set("minimum_engine_version", user.MinimumEngineVersion)
	d.Set(names.AttrUserName, user.Name)

	return diags
}
