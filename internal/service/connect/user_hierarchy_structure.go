// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"log"

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user_hierarchy_structure", name="User Hierarchy Structure")
func resourceUserHierarchyStructure() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserHierarchyStructureCreate,
		ReadWithoutTimeout:   resourceUserHierarchyStructureRead,
		UpdateWithoutTimeout: resourceUserHierarchyStructureUpdate,
		DeleteWithoutTimeout: resourceUserHierarchyStructureDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"hierarchy_structure": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"level_one":   hierarchyStructureLevelSchema(),
							"level_two":   hierarchyStructureLevelSchema(),
							"level_three": hierarchyStructureLevelSchema(),
							"level_four":  hierarchyStructureLevelSchema(),
							"level_five":  hierarchyStructureLevelSchema(),
						},
					},
				},
				names.AttrInstanceID: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 100),
				},
			}
		},
	}
}

// Each level shares the same schema.
func hierarchyStructureLevelSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MaxItems: 1,
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
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 50),
				},
			},
		},
	}
}

func resourceUserHierarchyStructureCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.UpdateUserHierarchyStructureInput{
		HierarchyStructure: expandHierarchyStructureUpdate(d.Get("hierarchy_structure").([]any)),
		InstanceId:         aws.String(instanceID),
	}

	_, err := conn.UpdateUserHierarchyStructure(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User Hierarchy Structure (%s): %s", instanceID, err)
	}

	d.SetId(instanceID)

	return append(diags, resourceUserHierarchyStructureRead(ctx, d, meta)...)
}

func resourceUserHierarchyStructureRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	hierarchyStructure, err := findUserHierarchyStructureByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect User Hierarchy Structure (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect User Hierarchy Structure (%s): %s", d.Id(), err)
	}

	if err := d.Set("hierarchy_structure", flattenHierarchyStructure(hierarchyStructure)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hierarchy_structure: %s", err)
	}
	d.Set(names.AttrInstanceID, d.Id())

	return diags
}

func resourceUserHierarchyStructureUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	if d.HasChange("hierarchy_structure") {
		input := &connect.UpdateUserHierarchyStructureInput{
			HierarchyStructure: expandHierarchyStructureUpdate(d.Get("hierarchy_structure").([]any)),
			InstanceId:         aws.String(d.Id()),
		}

		_, err := conn.UpdateUserHierarchyStructure(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect User Hierarchy Structure (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserHierarchyStructureRead(ctx, d, meta)...)
}

func resourceUserHierarchyStructureDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	log.Printf("[DEBUG] Deleting Connect User Hierarchy Structure: %s", d.Id())
	input := connect.UpdateUserHierarchyStructureInput{
		HierarchyStructure: &awstypes.HierarchyStructureUpdate{},
		InstanceId:         aws.String(d.Id()),
	}
	_, err := conn.UpdateUserHierarchyStructure(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect User Hierarchy Structure (%s): %s", d.Id(), err)
	}

	return diags
}

func findUserHierarchyStructureByID(ctx context.Context, conn *connect.Client, instanceID string) (*awstypes.HierarchyStructure, error) {
	input := &connect.DescribeUserHierarchyStructureInput{
		InstanceId: aws.String(instanceID),
	}

	return findUserHierarchyStructure(ctx, conn, input)
}

func findUserHierarchyStructure(ctx context.Context, conn *connect.Client, input *connect.DescribeUserHierarchyStructureInput) (*awstypes.HierarchyStructure, error) {
	output, err := conn.DescribeUserHierarchyStructure(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HierarchyStructure == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.HierarchyStructure, nil
}

func expandHierarchyStructureUpdate(tfList []any) *awstypes.HierarchyStructureUpdate {
	if len(tfList) == 0 {
		return &awstypes.HierarchyStructureUpdate{}
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HierarchyStructureUpdate{
		LevelOne:   expandHierarchyLevelUpdate(tfMap["level_one"].([]any)),
		LevelTwo:   expandHierarchyLevelUpdate(tfMap["level_two"].([]any)),
		LevelThree: expandHierarchyLevelUpdate(tfMap["level_three"].([]any)),
		LevelFour:  expandHierarchyLevelUpdate(tfMap["level_four"].([]any)),
		LevelFive:  expandHierarchyLevelUpdate(tfMap["level_five"].([]any)),
	}

	return apiObject
}

func expandHierarchyLevelUpdate(tfList []any) *awstypes.HierarchyLevelUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.HierarchyLevelUpdate{
		Name: aws.String(tfMap[names.AttrName].(string)),
	}

	return apiObject
}

func flattenHierarchyStructure(apiObject *awstypes.HierarchyStructure) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.LevelOne != nil {
		tfMap["level_one"] = flattenHierarchyLevel(apiObject.LevelOne)
	}

	if apiObject.LevelTwo != nil {
		tfMap["level_two"] = flattenHierarchyLevel(apiObject.LevelTwo)
	}

	if apiObject.LevelThree != nil {
		tfMap["level_three"] = flattenHierarchyLevel(apiObject.LevelThree)
	}

	if apiObject.LevelFour != nil {
		tfMap["level_four"] = flattenHierarchyLevel(apiObject.LevelFour)
	}

	if apiObject.LevelFive != nil {
		tfMap["level_five"] = flattenHierarchyLevel(apiObject.LevelFive)
	}

	return []any{tfMap}
}

func flattenHierarchyLevel(apiObject *awstypes.HierarchyLevel) []any {
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
