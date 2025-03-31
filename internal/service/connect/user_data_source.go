// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_user", name="User")
// @Tags
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
						"secondary_email": {
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

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeUserInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("user_id"); ok {
		input.UserId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		userSummary, err := findUserSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect User (%s) summary: %s", name, err)
		}

		input.UserId = userSummary.Id
	}

	user, err := findUser(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User: %s", err)
	}

	userID := aws.ToString(user.Id)
	id := userCreateResourceID(instanceID, userID)
	d.SetId(id)
	d.Set(names.AttrARN, user.Arn)
	d.Set("directory_user_id", user.DirectoryUserId)
	d.Set("hierarchy_group_id", user.HierarchyGroupId)
	if err := d.Set("identity_info", flattenUserIdentityInfo(user.IdentityInfo)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting identity_info: %s", err)
	}
	d.Set(names.AttrInstanceID, instanceID)
	d.Set(names.AttrName, user.Username)
	if err := d.Set("phone_config", flattenUserPhoneConfig(user.PhoneConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting phone_config: %s", err)
	}
	d.Set("routing_profile_id", user.RoutingProfileId)
	d.Set("security_profile_ids", user.SecurityProfileIds)
	d.Set("user_id", userID)

	setTagsOut(ctx, user.Tags)

	return diags
}

func findUserSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.UserSummary, error) {
	const maxResults = 60
	input := &connect.ListUsersInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findUserSummary(ctx, conn, input, func(v *awstypes.UserSummary) bool {
		return aws.ToString(v.Username) == name
	})
}

func findUserSummary(ctx context.Context, conn *connect.Client, input *connect.ListUsersInput, filter tfslices.Predicate[*awstypes.UserSummary]) (*awstypes.UserSummary, error) {
	output, err := findUserSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUserSummaries(ctx context.Context, conn *connect.Client, input *connect.ListUsersInput, filter tfslices.Predicate[*awstypes.UserSummary]) ([]awstypes.UserSummary, error) {
	var output []awstypes.UserSummary

	pages := connect.NewListUsersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
