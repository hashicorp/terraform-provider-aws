// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_device_pool", name="Device Pool")
// @Tags(identifierAttribute="arn")
func resourceDevicePool() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDevicePoolCreate,
		ReadWithoutTimeout:   resourceDevicePoolRead,
		UpdateWithoutTimeout: resourceDevicePoolUpdate,
		DeleteWithoutTimeout: resourceDevicePoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"max_devices": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"project_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRule: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DeviceAttribute](),
						},
						"operator": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RuleOperator](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDevicePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &devicefarm.CreateDevicePoolInput{
		Name:       aws.String(name),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
		Rules:      expandDevicePoolRules(d.Get(names.AttrRule).(*schema.Set)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_devices"); ok {
		input.MaxDevices = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateDevicePool(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Device Pool (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DevicePool.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Device Pool (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceDevicePoolRead(ctx, d, meta)...)
}

func resourceDevicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	devicePool, err := findDevicePoolByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Device Pool (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Device Pool (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(devicePool.Arn)
	d.Set(names.AttrName, devicePool.Name)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, devicePool.Description)
	d.Set("max_devices", devicePool.MaxDevices)

	projectArn, err := decodeProjectARN(arn, "devicepool", meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding project_arn (%s): %s", arn, err)
	}

	d.Set("project_arn", projectArn)

	if err := d.Set(names.AttrRule, flattenDevicePoolRules(devicePool.Rules)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting rule: %s", err)
	}

	return diags
}

func resourceDevicePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &devicefarm.UpdateDevicePoolInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrRule) {
			input.Rules = expandDevicePoolRules(d.Get(names.AttrRule).(*schema.Set))
		}

		if d.HasChange("max_devices") {
			if v, ok := d.GetOk("max_devices"); ok {
				input.MaxDevices = aws.Int32(int32(v.(int)))
			} else {
				input.ClearMaxDevices = aws.Bool(true)
			}
		}

		_, err := conn.UpdateDevicePool(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Device Pool (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDevicePoolRead(ctx, d, meta)...)
}

func resourceDevicePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Device Pool: %s", d.Id())
	_, err := conn.DeleteDevicePool(ctx, &devicefarm.DeleteDevicePoolInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Device Pool (%s): %s", d.Id(), err)
	}

	return diags
}

func findDevicePoolByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.DevicePool, error) {
	input := &devicefarm.GetDevicePoolInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetDevicePool(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DevicePool == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DevicePool, nil
}

func expandDevicePoolRules(s *schema.Set) []awstypes.Rule {
	rules := make([]awstypes.Rule, 0)

	for _, r := range s.List() {
		rule := awstypes.Rule{}
		tfMap := r.(map[string]interface{})

		if v, ok := tfMap["attribute"].(string); ok && v != "" {
			rule.Attribute = awstypes.DeviceAttribute(v)
		}

		if v, ok := tfMap["operator"].(string); ok && v != "" {
			rule.Operator = awstypes.RuleOperator(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			rule.Value = aws.String(v)
		}

		rules = append(rules, rule)
	}
	return rules
}

func flattenDevicePoolRules(list []awstypes.Rule) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{}

		l["attribute"] = string(setting.Attribute)
		l["operator"] = string(setting.Operator)

		if setting.Value != nil {
			l[names.AttrValue] = aws.ToString(setting.Value)
		}

		result = append(result, l)
	}
	return result
}

func decodeProjectARN(id, typ string, meta interface{}) (string, error) {
	poolArn, err := arn.Parse(id)
	if err != nil {
		return "", fmt.Errorf("parsing '%s': %w", id, err)
	}

	poolArnResouce := poolArn.Resource
	parts := strings.Split(strings.TrimPrefix(poolArnResouce, typ+":"), "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("Unexpected format of ID (%q), expected project-id/%q-id", poolArnResouce, typ)
	}

	projectId := parts[0]
	projectArn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  "project:" + projectId,
		Service:   names.DeviceFarmEndpointID,
	}.String()

	return projectArn, nil
}
