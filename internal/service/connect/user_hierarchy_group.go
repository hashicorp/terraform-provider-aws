// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user_hierarchy_group", name="User Hierarchy Group")
// @Tags(identifierAttribute="arn")
func resourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserHierarchyGroupCreate,
		ReadWithoutTimeout:   resourceUserHierarchyGroupRead,
		UpdateWithoutTimeout: resourceUserHierarchyGroupUpdate,
		DeleteWithoutTimeout: resourceUserHierarchyGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"hierarchy_group_id": {
					Type:     schema.TypeString,
					Computed: true,
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
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
				"parent_group_id": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

// Each level shares the same schema.
func hierarchyPathLevelSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}

func resourceUserHierarchyGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	userHierarchyGroupName := d.Get(names.AttrName).(string)
	input := &connect.CreateUserHierarchyGroupInput{
		InstanceId: aws.String(instanceID),
		Name:       aws.String(userHierarchyGroupName),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("parent_group_id"); ok {
		input.ParentGroupId = aws.String(v.(string))
	}

	output, err := conn.CreateUserHierarchyGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User Hierarchy Group (%s): %s", userHierarchyGroupName, err)
	}

	id := userHierarchyGroupCreateResourceID(instanceID, aws.ToString(output.HierarchyGroupId))
	d.SetId(id)

	return append(diags, resourceUserHierarchyGroupRead(ctx, d, meta)...)
}

func resourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userHierarchyGroupID, err := userHierarchyGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	hierarchyGroup, err := findUserHierarchyGroupByTwoPartKey(ctx, conn, instanceID, userHierarchyGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect User Hierarchy Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User Hierarchy Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, hierarchyGroup.Arn)
	d.Set("hierarchy_group_id", hierarchyGroup.Id)
	if err := d.Set("hierarchy_path", flattenHierarchyPath(hierarchyGroup.HierarchyPath)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hierarchy_path: %s", err)
	}
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("level_id", hierarchyGroup.LevelId)
	d.Set(names.AttrName, hierarchyGroup.Name)

	setTagsOut(ctx, hierarchyGroup.Tags)

	return diags
}

func resourceUserHierarchyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userHierarchyGroupID, err := userHierarchyGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrName) {
		input := &connect.UpdateUserHierarchyGroupNameInput{
			HierarchyGroupId: aws.String(userHierarchyGroupID),
			InstanceId:       aws.String(instanceID),
			Name:             aws.String(d.Get(names.AttrName).(string)),
		}

		_, err = conn.UpdateUserHierarchyGroupName(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User Hierarchy Group (%s) Name: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserHierarchyGroupRead(ctx, d, meta)...)
}

func resourceUserHierarchyGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID, userHierarchyGroupID, err := userHierarchyGroupParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Connect User Hierarchy Group: %s", d.Id())
	input := connect.DeleteUserHierarchyGroupInput{
		HierarchyGroupId: aws.String(userHierarchyGroupID),
		InstanceId:       aws.String(instanceID),
	}
	_, err = conn.DeleteUserHierarchyGroup(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect User Hierarchy Group (%s): %s", d.Id(), err)
	}

	return diags
}

const userHierarchyGroupResourceIDSeparator = ":"

func userHierarchyGroupCreateResourceID(instanceID, userHierarchyGroupID string) string {
	parts := []string{instanceID, userHierarchyGroupID}
	id := strings.Join(parts, userHierarchyGroupResourceIDSeparator)

	return id
}

func userHierarchyGroupParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, userHierarchyGroupResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected instanceID%[2]suserHierarchyGroupID", id, userHierarchyGroupResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findUserHierarchyGroupByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, userHierarchyGroupID string) (*awstypes.HierarchyGroup, error) {
	input := &connect.DescribeUserHierarchyGroupInput{
		HierarchyGroupId: aws.String(userHierarchyGroupID),
		InstanceId:       aws.String(instanceID),
	}

	return findUserHierarchyGroup(ctx, conn, input)
}

func findUserHierarchyGroup(ctx context.Context, conn *connect.Client, input *connect.DescribeUserHierarchyGroupInput) (*awstypes.HierarchyGroup, error) {
	output, err := conn.DescribeUserHierarchyGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HierarchyGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.HierarchyGroup, nil
}

func flattenHierarchyPath(apiObject *awstypes.HierarchyPath) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.LevelOne != nil {
		tfMap["level_one"] = flattenHierarchyGroupSummary(apiObject.LevelOne)
	}

	if apiObject.LevelTwo != nil {
		tfMap["level_two"] = flattenHierarchyGroupSummary(apiObject.LevelTwo)
	}

	if apiObject.LevelThree != nil {
		tfMap["level_three"] = flattenHierarchyGroupSummary(apiObject.LevelThree)
	}

	if apiObject.LevelFour != nil {
		tfMap["level_four"] = flattenHierarchyGroupSummary(apiObject.LevelFour)
	}

	if apiObject.LevelFive != nil {
		tfMap["level_five"] = flattenHierarchyGroupSummary(apiObject.LevelFive)
	}

	return []any{tfMap}
}

func flattenHierarchyGroupSummary(apiObject *awstypes.HierarchyGroupSummary) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrARN:  aws.ToString(apiObject.Arn),
		names.AttrID:   aws.ToString(apiObject.Id),
		names.AttrName: aws.ToString(apiObject.Name),
	}

	return []any{tfMap}
}
