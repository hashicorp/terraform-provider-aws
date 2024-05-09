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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_user_hierarchy_group")
func DataSourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserHierarchyGroupRead,
		Schema: map[string]*schema.Schema{
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
						"level_one": func() *schema.Schema {
							schema := userHierarchyPathLevelSchema()
							return schema
						}(),
						"level_two": func() *schema.Schema {
							schema := userHierarchyPathLevelSchema()
							return schema
						}(),
						"level_three": func() *schema.Schema {
							schema := userHierarchyPathLevelSchema()
							return schema
						}(),
						"level_four": func() *schema.Schema {
							schema := userHierarchyPathLevelSchema()
							return schema
						}(),
						"level_five": func() *schema.Schema {
							schema := userHierarchyPathLevelSchema()
							return schema
						}(),
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
		},
	}
}

func dataSourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeUserHierarchyGroupInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hierarchy_group_id"); ok {
		input.HierarchyGroupId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		hierarchyGroupSummary, err := userHierarchyGroupSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Hierarchy Group Summary by name (%s): %s", name, err)
		}

		if hierarchyGroupSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Hierarchy Group Summary by name (%s): not found", name)
		}

		input.HierarchyGroupId = hierarchyGroupSummary.Id
	}

	resp, err := conn.DescribeUserHierarchyGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hierarchy Group: %s", err)
	}

	if resp == nil || resp.HierarchyGroup == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Hierarchy Group: empty response")
	}

	hierarchyGroup := resp.HierarchyGroup

	d.Set(names.AttrARN, hierarchyGroup.Arn)
	d.Set("hierarchy_group_id", hierarchyGroup.Id)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("level_id", hierarchyGroup.LevelId)
	d.Set(names.AttrName, hierarchyGroup.Name)

	if err := d.Set("hierarchy_path", flattenUserHierarchyPath(hierarchyGroup.HierarchyPath)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Connect User Hierarchy Group hierarchy_path (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, hierarchyGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(hierarchyGroup.Id)))

	return diags
}

func userHierarchyGroupSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.HierarchyGroupSummary, error) {
	var result *connect.HierarchyGroupSummary

	input := &connect.ListUserHierarchyGroupsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListUserHierarchyGroupsMaxResults),
	}

	err := conn.ListUserHierarchyGroupsPagesWithContext(ctx, input, func(page *connect.ListUserHierarchyGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.UserHierarchyGroupSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
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
