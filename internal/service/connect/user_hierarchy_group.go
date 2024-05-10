// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_connect_user_hierarchy_group", name="User Hierarchy Group")
// @Tags(identifierAttribute="arn")
func ResourceUserHierarchyGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserHierarchyGroupCreate,
		ReadWithoutTimeout:   resourceUserHierarchyGroupRead,
		UpdateWithoutTimeout: resourceUserHierarchyGroupUpdate,
		DeleteWithoutTimeout: resourceUserHierarchyGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

// Each level shares the same schema
func userHierarchyPathLevelSchema() *schema.Schema {
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

func resourceUserHierarchyGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

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

	log.Printf("[DEBUG] Creating Connect User Hierarchy Group %s", input)
	output, err := conn.CreateUserHierarchyGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User Hierarchy Group (%s): %s", userHierarchyGroupName, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect User Hierarchy Group (%s): empty output", userHierarchyGroupName)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(output.HierarchyGroupId)))

	return append(diags, resourceUserHierarchyGroupRead(ctx, d, meta)...)
}

func resourceUserHierarchyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userHierarchyGroupID, err := UserHierarchyGroupParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := conn.DescribeUserHierarchyGroupWithContext(ctx, &connect.DescribeUserHierarchyGroupInput{
		HierarchyGroupId: aws.String(userHierarchyGroupID),
		InstanceId:       aws.String(instanceID),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Connect User Hierarchy Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User Hierarchy Group (%s): %s", d.Id(), err)
	}

	if resp == nil || resp.HierarchyGroup == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect User Hierarchy Group (%s): empty response", d.Id())
	}

	d.Set(names.AttrARN, resp.HierarchyGroup.Arn)
	d.Set("hierarchy_group_id", resp.HierarchyGroup.Id)
	d.Set(names.AttrInstanceID, instanceID)
	d.Set("level_id", resp.HierarchyGroup.LevelId)
	d.Set(names.AttrName, resp.HierarchyGroup.Name)

	if err := d.Set("hierarchy_path", flattenUserHierarchyPath(resp.HierarchyGroup.HierarchyPath)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Connect User Hierarchy Group hierarchy_path (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, resp.HierarchyGroup.Tags)

	return diags
}

func resourceUserHierarchyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userHierarchyGroupID, err := UserHierarchyGroupParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChange(names.AttrName) {
		_, err = conn.UpdateUserHierarchyGroupNameWithContext(ctx, &connect.UpdateUserHierarchyGroupNameInput{
			HierarchyGroupId: aws.String(userHierarchyGroupID),
			InstanceId:       aws.String(instanceID),
			Name:             aws.String(d.Get(names.AttrName).(string)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating User Hierarchy Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserHierarchyGroupRead(ctx, d, meta)...)
}

func resourceUserHierarchyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)

	instanceID, userHierarchyGroupID, err := UserHierarchyGroupParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteUserHierarchyGroupWithContext(ctx, &connect.DeleteUserHierarchyGroupInput{
		HierarchyGroupId: aws.String(userHierarchyGroupID),
		InstanceId:       aws.String(instanceID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting User Hierarchy Group (%s): %s", d.Id(), err)
	}

	return diags
}

func UserHierarchyGroupParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:userHierarchyGroupID", id)
	}

	return parts[0], parts[1], nil
}

func flattenUserHierarchyPath(userHierarchyPath *connect.HierarchyPath) []interface{} {
	if userHierarchyPath == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if userHierarchyPath.LevelOne != nil {
		values["level_one"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelOne)
	}

	if userHierarchyPath.LevelTwo != nil {
		values["level_two"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelTwo)
	}

	if userHierarchyPath.LevelThree != nil {
		values["level_three"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelThree)
	}

	if userHierarchyPath.LevelFour != nil {
		values["level_four"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelFour)
	}

	if userHierarchyPath.LevelFive != nil {
		values["level_five"] = flattenUserHierarchyPathLevel(userHierarchyPath.LevelFive)
	}

	return []interface{}{values}
}

func flattenUserHierarchyPathLevel(userHierarchyPathLevel *connect.HierarchyGroupSummary) []interface{} {
	if userHierarchyPathLevel == nil {
		return []interface{}{}
	}

	level := map[string]interface{}{
		names.AttrARN:  aws.StringValue(userHierarchyPathLevel.Arn),
		names.AttrID:   aws.StringValue(userHierarchyPathLevel.Id),
		names.AttrName: aws.StringValue(userHierarchyPathLevel.Name),
	}

	return []interface{}{level}
}
