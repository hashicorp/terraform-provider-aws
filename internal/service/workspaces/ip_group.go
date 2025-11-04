// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_ip_group", name="IP Group")
// @Tags(identifierAttribute="id")
func resourceIPGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPGroupCreate,
		ReadWithoutTimeout:   resourceIPGroupRead,
		UpdateWithoutTimeout: resourceIPGroupUpdate,
		DeleteWithoutTimeout: resourceIPGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrSource: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIPGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &workspaces.CreateIpGroupInput{
		GroupDesc: aws.String(d.Get(names.AttrDescription).(string)),
		GroupName: aws.String(name),
		Tags:      getTagsIn(ctx),
		UserRules: expandIPRuleItems(d.Get("rules").(*schema.Set).List()),
	}

	output, err := conn.CreateIpGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkSpaces IP Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GroupId))

	return append(diags, resourceIPGroupRead(ctx, d, meta)...)
}

func resourceIPGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	ipGroup, err := findIPGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkSpaces IP Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces IP Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrDescription, ipGroup.GroupDesc)
	d.Set(names.AttrName, ipGroup.GroupName)
	if err := d.Set("rules", flattenIPRuleItems(ipGroup.UserRules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rules: %s", err)
	}

	return diags
}

func resourceIPGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	if d.HasChange("rules") {
		input := &workspaces.UpdateRulesOfIpGroupInput{
			GroupId:   aws.String(d.Id()),
			UserRules: expandIPRuleItems(d.Get("rules").(*schema.Set).List()),
		}

		_, err := conn.UpdateRulesOfIpGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces IP Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPGroupRead(ctx, d, meta)...)
}

func resourceIPGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	describeInput := &workspaces.DescribeWorkspaceDirectoriesInput{}
	directories, err := findDirectories(ctx, conn, describeInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces Directories: %s", err)
	}

	for _, v := range directories {
		directoryID := aws.ToString(v.DirectoryId)
		for _, v := range v.IpGroupIds {
			if v == d.Id() {
				input := &workspaces.DisassociateIpGroupsInput{
					DirectoryId: aws.String(directoryID),
					GroupIds:    []string{d.Id()},
				}

				_, err := conn.DisassociateIpGroups(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "disassociating WorkSpaces Directory (%s) IP Group (%s): %s", directoryID, d.Id(), err)
				}
			}
		}
	}

	log.Printf("[DEBUG] Deleting WorkSpaces IP Group (%s)", d.Id())
	deleteInput := workspaces.DeleteIpGroupInput{
		GroupId: aws.String(d.Id()),
	}
	_, err = conn.DeleteIpGroup(ctx, &deleteInput)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkSpaces IP Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findIPGroupByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspacesIpGroup, error) {
	input := &workspaces.DescribeIpGroupsInput{
		GroupIds: []string{id},
	}

	return findIPGroup(ctx, conn, input)
}

func findIPGroup(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeIpGroupsInput) (*types.WorkspacesIpGroup, error) {
	output, err := findIPGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPGroups(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeIpGroupsInput) ([]types.WorkspacesIpGroup, error) {
	var output []types.WorkspacesIpGroup

	err := describeIPGroupsPages(ctx, conn, input, func(page *workspaces.DescribeIpGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Result...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandIPRuleItems(tfList []any) []types.IpRuleItem {
	var apiObjects []types.IpRuleItem

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObjects = append(apiObjects, types.IpRuleItem{
			IpRule:   aws.String(tfMap[names.AttrSource].(string)),
			RuleDesc: aws.String(tfMap[names.AttrDescription].(string)),
		})
	}

	return apiObjects
}

func flattenIPRuleItems(apiObjects []types.IpRuleItem) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.IpRule; v != nil {
			tfMap[names.AttrSource] = aws.ToString(v)
		}

		if v := apiObject.RuleDesc; v != nil {
			tfMap[names.AttrDescription] = aws.ToString(apiObject.RuleDesc)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
