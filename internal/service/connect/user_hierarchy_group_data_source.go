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

// @SDKDataSource("aws_connect_user_hierarchy_group", name="User Hierarchy Group")
// @Tags
func dataSourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserHierarchyGroupRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"hierarchy_group_id": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ExactlyOneOf: []string{"hierarchy_group_id", names.AttrName},
				},
				"hierarchy_path": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"level_one":   hierarchyPathLevelSchema(),
							"level_two":   hierarchyPathLevelSchema(),
							"level_three": hierarchyPathLevelSchema(),
							"level_four":  hierarchyPathLevelSchema(),
							"level_five":  hierarchyPathLevelSchema(),
						},
					},
				},
				names.AttrInstanceID: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
				"level_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ExactlyOneOf: []string{names.AttrName, "hierarchy_group_id"},
				},
				// parent_group_id is not returned by DescribeUserHierarchyGroup
				// "parent_group_id": {
				// 	Type:     schema.TypeString,
				// 	Computed: true,
				// },
				names.AttrTags: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func dataSourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeUserHierarchyGroupInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hierarchy_group_id"); ok {
		input.HierarchyGroupId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		hierarchyGroupSummary, err := findUserHierarchyGroupSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect User Hierarchy Group (%s) summary: %s", name, err)
		}

		input.HierarchyGroupId = hierarchyGroupSummary.Id
	}

	hierarchyGroup, err := findUserHierarchyGroup(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User Hierarchy Group: %s", err)
	}

	userHierarchyGroupID := aws.ToString(hierarchyGroup.Id)
	id := userHierarchyGroupCreateResourceID(instanceID, userHierarchyGroupID)
	d.SetId(id)
	d.Set(names.AttrARN, hierarchyGroup.Arn)
	d.Set("hierarchy_group_id", userHierarchyGroupID)
	if err := d.Set("hierarchy_path", flattenHierarchyPath(hierarchyGroup.HierarchyPath)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hierarchy_path: %s", err)
	}
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("level_id", hierarchyGroup.LevelId)
	d.Set(names.AttrName, hierarchyGroup.Name)

	setTagsOut(ctx, hierarchyGroup.Tags)

	return diags
}

func findUserHierarchyGroupSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.HierarchyGroupSummary, error) {
	const maxResults = 60
	input := &connect.ListUserHierarchyGroupsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findUserHierarchyGroupSummary(ctx, conn, input, func(v *awstypes.HierarchyGroupSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findUserHierarchyGroupSummary(ctx context.Context, conn *connect.Client, input *connect.ListUserHierarchyGroupsInput, filter tfslices.Predicate[*awstypes.HierarchyGroupSummary]) (*awstypes.HierarchyGroupSummary, error) {
	output, err := findUserHierarchyGroupSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findUserHierarchyGroupSummaries(ctx context.Context, conn *connect.Client, input *connect.ListUserHierarchyGroupsInput, filter tfslices.Predicate[*awstypes.HierarchyGroupSummary]) ([]awstypes.HierarchyGroupSummary, error) {
	var output []awstypes.HierarchyGroupSummary

	pages := connect.NewListUserHierarchyGroupsPaginator(conn, input)
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

		for _, v := range page.UserHierarchyGroupSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
