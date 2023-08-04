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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_connect_user_hierarchy_group")
func DataSourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserHierarchyGroupRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hierarchy_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hierarchy_group_id", "name"},
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
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"level_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "hierarchy_group_id"},
			},
			// parent_group_id is not returned by DescribeUserHierarchyGroup
			// "parent_group_id": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeUserHierarchyGroupInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("hierarchy_group_id"); ok {
		input.HierarchyGroupId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		hierarchyGroupSummary, err := userHierarchyGroupSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.Errorf("finding Connect Hierarchy Group Summary by name (%s): %s", name, err)
		}

		if hierarchyGroupSummary == nil {
			return diag.Errorf("finding Connect Hierarchy Group Summary by name (%s): not found", name)
		}

		input.HierarchyGroupId = hierarchyGroupSummary.Id
	}

	resp, err := conn.DescribeUserHierarchyGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("getting Connect Hierarchy Group: %s", err)
	}

	if resp == nil || resp.HierarchyGroup == nil {
		return diag.Errorf("getting Connect Hierarchy Group: empty response")
	}

	hierarchyGroup := resp.HierarchyGroup

	d.Set("arn", hierarchyGroup.Arn)
	d.Set("hierarchy_group_id", hierarchyGroup.Id)
	d.Set("instance_id", instanceID)
	d.Set("level_id", hierarchyGroup.LevelId)
	d.Set("name", hierarchyGroup.Name)

	if err := d.Set("hierarchy_path", flattenUserHierarchyPath(hierarchyGroup.HierarchyPath)); err != nil {
		return diag.Errorf("setting Connect User Hierarchy Group hierarchy_path (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, hierarchyGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(hierarchyGroup.Id)))

	return nil
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
