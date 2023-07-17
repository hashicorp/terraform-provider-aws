// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_memorydb_user")
func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"access_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
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
						"type": {
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
			"tags": tftags.TagsSchemaComputed(),
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	userName := d.Get("user_name").(string)

	user, err := FindUserByName(ctx, conn, userName)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("MemoryDB User", err))
	}

	d.SetId(aws.StringValue(user.Name))

	d.Set("access_string", user.AccessString)
	d.Set("arn", user.ARN)

	if v := user.Authentication; v != nil {
		authenticationMode := map[string]interface{}{
			"password_count": aws.Int64Value(v.PasswordCount),
			"type":           aws.StringValue(v.Type),
		}

		if err := d.Set("authentication_mode", []interface{}{authenticationMode}); err != nil {
			return diag.Errorf("failed to set authentication_mode of MemoryDB User (%s): %s", d.Id(), err)
		}
	}

	d.Set("minimum_engine_version", user.MinimumEngineVersion)
	d.Set("user_name", user.Name)

	tags, err := listTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for MemoryDB User (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
