// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user_hierarchy_structure")
func ResourceUserHierarchyStructure() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserHierarchyStructureCreate,
		ReadWithoutTimeout:   resourceUserHierarchyStructureRead,
		UpdateWithoutTimeout: resourceUserHierarchyStructureUpdate,
		DeleteWithoutTimeout: resourceUserHierarchyStructureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"hierarchy_structure": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"level_one": func() *schema.Schema {
							schema := userHierarchyLevelSchema()
							return schema
						}(),
						"level_two": func() *schema.Schema {
							schema := userHierarchyLevelSchema()
							return schema
						}(),
						"level_three": func() *schema.Schema {
							schema := userHierarchyLevelSchema()
							return schema
						}(),
						"level_four": func() *schema.Schema {
							schema := userHierarchyLevelSchema()
							return schema
						}(),
						"level_five": func() *schema.Schema {
							schema := userHierarchyLevelSchema()
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
		},
	}
}

// Each level shares the same schema
func userHierarchyLevelSchema() *schema.Schema {
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

func resourceUserHierarchyStructureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.UpdateUserHierarchyStructureInput{
		HierarchyStructure: expandUserHierarchyStructure(d.Get("hierarchy_structure").([]interface{})),
		InstanceId:         aws.String(instanceID),
	}

	log.Printf("[DEBUG] Creating Connect User Hierarchy Structure %s", input)
	_, err := conn.UpdateUserHierarchyStructureWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User Hierarchy Structure for Connect Instance (%s): %s", instanceID, err)
	}

	d.SetId(instanceID)

	return append(diags, resourceUserHierarchyStructureRead(ctx, d, meta)...)
}

func resourceUserHierarchyStructureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Id()

	resp, err := conn.DescribeUserHierarchyStructureWithContext(ctx, &connect.DescribeUserHierarchyStructureInput{
		InstanceId: aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect User Hierarchy Structure (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User Hierarchy Structure (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.HierarchyStructure == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User Hierarchy Structure (%s): empty response", d.Id())
	}

	if err := d.Set("hierarchy_structure", flattenUserHierarchyStructure(resp.HierarchyStructure)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Connect User Hierarchy Structure hierarchy_structure for Connect instance: (%s)", d.Id())
	}

	d.Set(names.AttrInstanceID, instanceID)

	return diags
}

func resourceUserHierarchyStructureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Id()

	if d.HasChange("hierarchy_structure") {
		_, err := conn.UpdateUserHierarchyStructureWithContext(ctx, &connect.UpdateUserHierarchyStructureInput{
			HierarchyStructure: expandUserHierarchyStructure(d.Get("hierarchy_structure").([]interface{})),
			InstanceId:         aws.String(instanceID),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating UserHierarchyStructure Name (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserHierarchyStructureRead(ctx, d, meta)...)
}

func resourceUserHierarchyStructureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID := d.Id()

	_, err := conn.UpdateUserHierarchyStructureWithContext(ctx, &connect.UpdateUserHierarchyStructureInput{
		HierarchyStructure: &connect.HierarchyStructureUpdate{},
		InstanceId:         aws.String(instanceID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting UserHierarchyStructure (%s): %s", d.Id(), err)
	}

	return diags
}

func expandUserHierarchyStructure(userHierarchyStructure []interface{}) *connect.HierarchyStructureUpdate {
	if len(userHierarchyStructure) == 0 {
		return &connect.HierarchyStructureUpdate{}
	}

	tfMap, ok := userHierarchyStructure[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &connect.HierarchyStructureUpdate{
		LevelOne:   expandUserHierarchyStructureLevel(tfMap["level_one"].([]interface{})),
		LevelTwo:   expandUserHierarchyStructureLevel(tfMap["level_two"].([]interface{})),
		LevelThree: expandUserHierarchyStructureLevel(tfMap["level_three"].([]interface{})),
		LevelFour:  expandUserHierarchyStructureLevel(tfMap["level_four"].([]interface{})),
		LevelFive:  expandUserHierarchyStructureLevel(tfMap["level_five"].([]interface{})),
	}

	return result
}

func expandUserHierarchyStructureLevel(userHierarchyStructureLevel []interface{}) *connect.HierarchyLevelUpdate {
	if len(userHierarchyStructureLevel) == 0 {
		return nil
	}

	tfMap, ok := userHierarchyStructureLevel[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &connect.HierarchyLevelUpdate{
		Name: aws.String(tfMap[names.AttrName].(string)),
	}

	return result
}

func flattenUserHierarchyStructure(userHierarchyStructure *connect.HierarchyStructure) []interface{} {
	if userHierarchyStructure == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if userHierarchyStructure.LevelOne != nil {
		values["level_one"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelOne)
	}

	if userHierarchyStructure.LevelTwo != nil {
		values["level_two"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelTwo)
	}

	if userHierarchyStructure.LevelThree != nil {
		values["level_three"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelThree)
	}

	if userHierarchyStructure.LevelFour != nil {
		values["level_four"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelFour)
	}

	if userHierarchyStructure.LevelFive != nil {
		values["level_five"] = flattenUserHierarchyStructureLevel(userHierarchyStructure.LevelFive)
	}

	return []interface{}{values}
}

func flattenUserHierarchyStructureLevel(userHierarchyStructureLevel *connect.HierarchyLevel) []interface{} {
	if userHierarchyStructureLevel == nil {
		return []interface{}{}
	}

	level := map[string]interface{}{
		names.AttrARN:  aws.StringValue(userHierarchyStructureLevel.Arn),
		names.AttrID:   aws.StringValue(userHierarchyStructureLevel.Id),
		names.AttrName: aws.StringValue(userHierarchyStructureLevel.Name),
	}

	return []interface{}{level}
}
