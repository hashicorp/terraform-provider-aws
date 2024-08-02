// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_user")
func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hierarchy_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEmail: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"first_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "user_id"},
			},
			"phone_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"after_contact_work_time_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"auto_accept": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"desk_phone_number": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"phone_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"routing_profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_profile_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"user_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"user_id", names.AttrName},
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeUserInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("user_id"); ok {
		input.UserId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		userSummary, err := dataSourceGetUserSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect User Summary by name (%s): %s", name, err)
		}

		if userSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect User Summary by name (%s): not found", name)
		}

		input.UserId = userSummary.Id
	}

	resp, err := conn.DescribeUserWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User: %s", err)
	}

	if resp == nil || resp.User == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User: empty response")
	}

	user := resp.User

	d.Set(names.AttrARN, user.Arn)
	d.Set("directory_user_id", user.DirectoryUserId)
	d.Set("hierarchy_group_id", user.HierarchyGroupId)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, user.Username)
	d.Set("routing_profile_id", user.RoutingProfileId)
	d.Set("security_profile_ids", flex.FlattenStringSet(user.SecurityProfileIds))
	d.Set("user_id", user.Id)

	if err := d.Set("identity_info", flattenIdentityInfo(user.IdentityInfo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting identity_info: %s", err)
	}

	if err := d.Set("phone_config", flattenPhoneConfig(user.PhoneConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting phone_config: %s", err)
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, user.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(user.Id)))

	return diags
}

func dataSourceGetUserSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.UserSummary, error) {
	var result *connect.UserSummary

	input := &connect.ListUsersInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListUsersMaxResults),
	}

	err := conn.ListUsersPagesWithContext(ctx, input, func(page *connect.ListUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.UserSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Username) == name {
				result = qs
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
