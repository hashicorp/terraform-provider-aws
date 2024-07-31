// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_xray_group", name="Group")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray/types;types.Group")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"insights_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insights_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"notifications_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	name := d.Get(names.AttrGroupName).(string)
	input := &xray.CreateGroupInput{
		GroupName:        aws.String(name),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("insights_configuration"); ok {
		input.InsightsConfiguration = expandInsightsConfig(v.([]interface{}))
	}

	output, err := conn.CreateGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Group.GroupARN))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	group, err := findGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] XRay Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.GroupARN)
	d.Set("filter_expression", group.FilterExpression)
	d.Set(names.AttrGroupName, group.GroupName)
	if err := d.Set("insights_configuration", flattenInsightsConfig(group.InsightsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting insights_configuration: %s", err)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &xray.UpdateGroupInput{GroupARN: aws.String(d.Id())}

		if v, ok := d.GetOk("filter_expression"); ok {
			input.FilterExpression = aws.String(v.(string))
		}

		if v, ok := d.GetOk("insights_configuration"); ok {
			input.InsightsConfiguration = expandInsightsConfig(v.([]interface{}))
		}

		_, err := conn.UpdateGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating XRay Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	log.Printf("[INFO] Deleting XRay Group: %s", d.Id())
	_, err := conn.DeleteGroup(ctx, &xray.DeleteGroupInput{
		GroupARN: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "Group not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting XRay Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findGroupByARN(ctx context.Context, conn *xray.Client, arn string) (*types.Group, error) {
	input := &xray.GetGroupInput{
		GroupARN: aws.String(arn),
	}

	output, err := conn.GetGroup(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "Group not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Group, nil
}

func expandInsightsConfig(l []interface{}) *types.InsightsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := types.InsightsConfiguration{}

	if v, ok := m["insights_enabled"]; ok {
		config.InsightsEnabled = aws.Bool(v.(bool))
	}
	if v, ok := m["notifications_enabled"]; ok {
		config.NotificationsEnabled = aws.Bool(v.(bool))
	}

	return &config
}

func flattenInsightsConfig(config *types.InsightsConfiguration) []interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{}

	if config.InsightsEnabled != nil {
		m["insights_enabled"] = config.InsightsEnabled
	}
	if config.NotificationsEnabled != nil {
		m["notifications_enabled"] = config.NotificationsEnabled
	}

	return []interface{}{m}
}
