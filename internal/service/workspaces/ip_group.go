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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_workspaces_ip_group", name="IP Group")
// @Tags(identifierAttribute="id")
func ResourceIPGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPGroupCreate,
		ReadWithoutTimeout:   resourceIPGroupRead,
		UpdateWithoutTimeout: resourceIPGroupUpdate,
		DeleteWithoutTimeout: resourceIPGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSource: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	rules := d.Get("rules").(*schema.Set).List()
	resp, err := conn.CreateIpGroup(ctx, &workspaces.CreateIpGroupInput{
		GroupName: aws.String(d.Get(names.AttrName).(string)),
		GroupDesc: aws.String(d.Get(names.AttrDescription).(string)),
		UserRules: expandIPGroupRules(rules),
		Tags:      getTagsIn(ctx),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WorkSpaces IP Group: %s", err)
	}

	d.SetId(aws.ToString(resp.GroupId))

	return append(diags, resourceIPGroupRead(ctx, d, meta)...)
}

func resourceIPGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	resp, err := conn.DescribeIpGroups(ctx, &workspaces.DescribeIpGroupsInput{
		GroupIds: []string{d.Id()},
	})
	if err != nil {
		if len(resp.Result) == 0 {
			log.Printf("[WARN] WorkSpaces IP Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WorkSpaces IP Group (%s): %s", d.Id(), err)
	}

	ipGroups := resp.Result

	if len(ipGroups) == 0 {
		log.Printf("[WARN] WorkSpaces Ip Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	ipGroup := ipGroups[0]

	d.Set(names.AttrName, ipGroup.GroupName)
	d.Set(names.AttrDescription, ipGroup.GroupDesc)
	d.Set("rules", flattenIPGroupRules(ipGroup.UserRules))

	return diags
}

func resourceIPGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	if d.HasChange("rules") {
		rules := d.Get("rules").(*schema.Set).List()

		_, err := conn.UpdateRulesOfIpGroup(ctx, &workspaces.UpdateRulesOfIpGroupInput{
			GroupId:   aws.String(d.Id()),
			UserRules: expandIPGroupRules(rules),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WorkSpaces IP Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPGroupRead(ctx, d, meta)...)
}

func resourceIPGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	var found bool
	log.Printf("[DEBUG] Finding directories associated with WorkSpaces IP Group (%s)", d.Id())
	paginator := workspaces.NewDescribeWorkspaceDirectoriesPaginator(conn, &workspaces.DescribeWorkspaceDirectoriesInput{}, func(out *workspaces.DescribeWorkspaceDirectoriesPaginatorOptions) {})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			diags = sdkdiag.AppendErrorf(diags, "describing WorkSpaces Directories: %s", err)
		}
		for _, dir := range out.Directories {
			for _, ipg := range dir.IpGroupIds {
				groupID := ipg
				if groupID == d.Id() {
					found = true
					log.Printf("[DEBUG] WorkSpaces IP Group (%s) associated with WorkSpaces Directory (%s), disassociating", groupID, aws.ToString(dir.DirectoryId))
					_, err := conn.DisassociateIpGroups(ctx, &workspaces.DisassociateIpGroupsInput{
						DirectoryId: dir.DirectoryId,
						GroupIds:    []string{d.Id()},
					})
					if err != nil {
						diags = sdkdiag.AppendErrorf(diags, "disassociating WorkSpaces IP Group (%s) from WorkSpaces Directory (%s): %s", d.Id(), aws.ToString(dir.DirectoryId), err)
						continue
					}
					log.Printf("[INFO] WorkSpaces IP Group (%s) disassociated from WorkSpaces Directory (%s)", d.Id(), aws.ToString(dir.DirectoryId))
				}
			}
		}
	}

	if diags.HasError() {
		return diags
	}

	if !found {
		log.Printf("[DEBUG] WorkSpaces IP Group (%s) not associated with any WorkSpaces Directories", d.Id())
	}

	log.Printf("[DEBUG] Deleting WorkSpaces IP Group (%s)", d.Id())
	_, err := conn.DeleteIpGroup(ctx, &workspaces.DeleteIpGroupInput{
		GroupId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WorkSpaces IP Group (%s): %s", d.Id(), err)
	}
	log.Printf("[INFO] WorkSpaces IP Group (%s) deleted", d.Id())

	return diags
}

func expandIPGroupRules(rules []interface{}) []types.IpRuleItem {
	var result []types.IpRuleItem
	for _, rule := range rules {
		r := rule.(map[string]interface{})

		result = append(result, types.IpRuleItem{
			IpRule:   aws.String(r[names.AttrSource].(string)),
			RuleDesc: aws.String(r[names.AttrDescription].(string)),
		})
	}
	return result
}

func flattenIPGroupRules(rules []types.IpRuleItem) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		r := map[string]interface{}{}

		if v := rule.IpRule; v != nil {
			r[names.AttrSource] = aws.ToString(v)
		}

		if v := rule.RuleDesc; v != nil {
			r[names.AttrDescription] = aws.ToString(rule.RuleDesc)
		}

		result = append(result, r)
	}
	return result
}
