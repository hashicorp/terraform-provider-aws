// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_usage_plan", name="Usage Plan")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetUsagePlanOutput")
func resourceUsagePlan() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUsagePlanCreate,
		ReadWithoutTimeout:   resourceUsagePlanRead,
		UpdateWithoutTimeout: resourceUsagePlanUpdate,
		DeleteWithoutTimeout: resourceUsagePlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_stages": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrStage: {
							Type:     schema.TypeString,
							Required: true,
						},
						"throttle": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"burst_limit": {
										Type:     schema.TypeInt,
										Default:  0,
										Optional: true,
									},
									names.AttrPath: {
										Type:     schema.TypeString,
										Required: true,
									},
									"rate_limit": {
										Type:     schema.TypeFloat,
										Default:  0,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true, // Required since not addable nor removable afterwards
			},
			"product_code": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"quota_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"limit": {
							Type:     schema.TypeInt,
							Required: true, // Required as not removable singularly
						},
						"offset": {
							Type:     schema.TypeInt,
							Default:  0,
							Optional: true,
						},
						"period": {
							Type:             schema.TypeString,
							Required:         true, // Required as not removable
							ValidateDiagFunc: enum.Validate[types.QuotaPeriodType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"throttle_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"burst_limit": {
							Type:         schema.TypeInt,
							Default:      0,
							Optional:     true,
							AtLeastOneOf: []string{"throttle_settings.0.burst_limit", "throttle_settings.0.rate_limit"},
						},
						"rate_limit": {
							Type:         schema.TypeFloat,
							Default:      0,
							Optional:     true,
							AtLeastOneOf: []string{"throttle_settings.0.burst_limit", "throttle_settings.0.rate_limit"},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUsagePlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateUsagePlanInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("api_stages"); ok && v.(*schema.Set).Len() > 0 {
		input.ApiStages = expandAPIStages(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("quota_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		settings := v.([]interface{})
		q, ok := settings[0].(map[string]interface{})

		if errs := validUsagePlanQuotaSettings(q); len(errs) > 0 {
			return sdkdiag.AppendErrorf(diags, "validating the quota settings: %v", errs)
		}

		if !ok {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside quota_settings")
		}

		input.Quota = expandQuotaSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("throttle_settings"); ok {
		input.Throttle = expandThrottleSettings(v.([]interface{}))
	}

	output, err := conn.CreateUsagePlan(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Usage Plan (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	// Handle case of adding the product code since not addable when
	// creating the Usage Plan initially.
	if v, ok := d.GetOk("product_code"); ok {
		input := &apigateway.UpdateUsagePlanInput{
			PatchOperations: []types.PatchOperation{
				{
					Op:    types.OpAdd,
					Path:  aws.String("/productCode"),
					Value: aws.String(v.(string)),
				},
			},
			UsagePlanId: aws.String(d.Id()),
		}

		_, err = conn.UpdateUsagePlan(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding API Gateway Usage Plan (%s) product code: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsagePlanRead(ctx, d, meta)...)
}

func resourceUsagePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	up, err := findUsagePlanByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Usage Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Usage Plan (%s): %s", d.Id(), err)
	}

	if up.ApiStages != nil {
		if err := d.Set("api_stages", flattenAPIStages(up.ApiStages)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting api_stages: %s", err)
		}
	}
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/usageplans/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, up.Description)
	d.Set(names.AttrName, up.Name)
	d.Set("product_code", up.ProductCode)
	if up.Quota != nil {
		if err := d.Set("quota_settings", flattenQuotaSettings(up.Quota)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting quota_settings: %s", err)
		}
	}
	if up.Throttle != nil {
		if err := d.Set("throttle_settings", flattenThrottleSettings(up.Throttle)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting throttle_settings: %s", err)
		}
	}

	setTagsOut(ctx, up.Tags)

	return diags
}

func resourceUsagePlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		operations := make([]types.PatchOperation, 0)

		if d.HasChange(names.AttrName) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/name"),
				Value: aws.String(d.Get(names.AttrName).(string)),
			})
		}

		if d.HasChange(names.AttrDescription) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/description"),
				Value: aws.String(d.Get(names.AttrDescription).(string)),
			})
		}

		if d.HasChange("product_code") {
			if v, ok := d.GetOk("product_code"); ok {
				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String("/productCode"),
					Value: aws.String(v.(string)),
				})
			} else {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String("/productCode"),
				})
			}
		}

		if d.HasChange("api_stages") {
			o, n := d.GetChange("api_stages")
			os := o.(*schema.Set).List()
			ns := n.(*schema.Set).List()

			// Remove every stages associated. Simpler to remove and add new ones,
			// since there are no replacings.
			for _, v := range os {
				m := v.(map[string]interface{})
				operations = append(operations, types.PatchOperation{
					Op:    types.OpRemove,
					Path:  aws.String("/apiStages"),
					Value: aws.String(fmt.Sprintf("%s:%s", m["api_id"].(string), m[names.AttrStage].(string))),
				})
			}

			// Handle additions
			if len(ns) > 0 {
				for _, v := range ns {
					m := v.(map[string]interface{})
					id := fmt.Sprintf("%s:%s", m["api_id"].(string), m[names.AttrStage].(string))
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/apiStages"),
						Value: aws.String(id),
					})
					if t, ok := m["throttle"].(*schema.Set); ok && t.Len() > 0 {
						for _, throttle := range t.List() {
							th := throttle.(map[string]interface{})
							operations = append(operations, types.PatchOperation{
								Op:    types.OpReplace,
								Path:  aws.String(fmt.Sprintf("/apiStages/%s/throttle/%s/rateLimit", id, th[names.AttrPath].(string))),
								Value: aws.String(strconv.FormatFloat(th["rate_limit"].(float64), 'f', -1, 64)),
							})
							operations = append(operations, types.PatchOperation{
								Op:    types.OpReplace,
								Path:  aws.String(fmt.Sprintf("/apiStages/%s/throttle/%s/burstLimit", id, th[names.AttrPath].(string))),
								Value: aws.String(strconv.Itoa(th["burst_limit"].(int))),
							})
						}
					}
				}
			}
		}

		if d.HasChange("throttle_settings") {
			o, n := d.GetChange("throttle_settings")
			diff := n.([]interface{})

			// Handle Removal
			if len(diff) == 0 {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String("/throttle"),
				})
			}

			if len(diff) > 0 {
				d := diff[0].(map[string]interface{})

				// Handle Replaces
				if o != nil && n != nil {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/throttle/rateLimit"),
						Value: aws.String(strconv.FormatFloat(d["rate_limit"].(float64), 'f', -1, 64)),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/throttle/burstLimit"),
						Value: aws.String(strconv.Itoa(d["burst_limit"].(int))),
					})
				}

				// Handle Additions
				if o == nil && n != nil {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/throttle/rateLimit"),
						Value: aws.String(strconv.FormatFloat(d["rate_limit"].(float64), 'f', -1, 64)),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/throttle/burstLimit"),
						Value: aws.String(strconv.Itoa(d["burst_limit"].(int))),
					})
				}
			}
		}

		if d.HasChange("quota_settings") {
			o, n := d.GetChange("quota_settings")
			diff := n.([]interface{})

			// Handle Removal
			if len(diff) == 0 {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String("/quota"),
				})
			}

			if len(diff) > 0 {
				d := diff[0].(map[string]interface{})

				if errors := validUsagePlanQuotaSettings(d); len(errors) > 0 {
					return sdkdiag.AppendErrorf(diags, "validating the quota settings: %v", errors)
				}

				// Handle Replaces
				if o != nil && n != nil {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/quota/limit"),
						Value: aws.String(strconv.Itoa(d["limit"].(int))),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/quota/offset"),
						Value: aws.String(strconv.Itoa(d["offset"].(int))),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/quota/period"),
						Value: aws.String(d["period"].(string)),
					})
				}

				// Handle Additions
				if o == nil && n != nil {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/quota/limit"),
						Value: aws.String(strconv.Itoa(d["limit"].(int))),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/quota/offset"),
						Value: aws.String(strconv.Itoa(d["offset"].(int))),
					})
					operations = append(operations, types.PatchOperation{
						Op:    types.OpAdd,
						Path:  aws.String("/quota/period"),
						Value: aws.String(d["period"].(string)),
					})
				}
			}
		}

		input := &apigateway.UpdateUsagePlanInput{
			PatchOperations: operations,
			UsagePlanId:     aws.String(d.Id()),
		}

		_, err := conn.UpdateUsagePlan(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Usage Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsagePlanRead(ctx, d, meta)...)
}

func resourceUsagePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	// Removing existing api stages associated
	if apistages, ok := d.GetOk("api_stages"); ok {
		stages := apistages.(*schema.Set)
		operations := []types.PatchOperation{}

		for _, v := range stages.List() {
			sv := v.(map[string]interface{})

			operations = append(operations, types.PatchOperation{
				Op:    types.OpRemove,
				Path:  aws.String("/apiStages"),
				Value: aws.String(fmt.Sprintf("%s:%s", sv["api_id"].(string), sv[names.AttrStage].(string))),
			})
		}

		_, err := conn.UpdateUsagePlan(ctx, &apigateway.UpdateUsagePlanInput{
			PatchOperations: operations,
			UsagePlanId:     aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "removing API Gateway Usage Plan (%s) API stages: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan: %s", d.Id())
	_, err := conn.DeleteUsagePlan(ctx, &apigateway.DeleteUsagePlanInput{
		UsagePlanId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Usage Plan (%s): %s", d.Id(), err)
	}

	return diags
}

func findUsagePlanByID(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetUsagePlanOutput, error) {
	input := &apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(id),
	}

	output, err := conn.GetUsagePlan(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandAPIStages(s *schema.Set) []types.ApiStage {
	stages := []types.ApiStage{}

	for _, stageRaw := range s.List() {
		stage := types.ApiStage{}
		mStage := stageRaw.(map[string]interface{})

		if v, ok := mStage["api_id"].(string); ok && v != "" {
			stage.ApiId = aws.String(v)
		}

		if v, ok := mStage[names.AttrStage].(string); ok && v != "" {
			stage.Stage = aws.String(v)
		}

		if v, ok := mStage["throttle"].(*schema.Set); ok && v.Len() > 0 {
			stage.Throttle = expandThrottleSettingsList(v.List())
		}

		stages = append(stages, stage)
	}

	return stages
}

func expandQuotaSettings(l []interface{}) *types.QuotaSettings {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	qs := &types.QuotaSettings{}

	if v, ok := m["limit"].(int); ok {
		qs.Limit = int32(v)
	}

	if v, ok := m["offset"].(int); ok {
		qs.Offset = int32(v)
	}

	if v, ok := m["period"].(string); ok && v != "" {
		qs.Period = types.QuotaPeriodType(v)
	}

	return qs
}

func expandThrottleSettings(l []interface{}) *types.ThrottleSettings {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	ts := &types.ThrottleSettings{}

	if sv, ok := m["burst_limit"].(int); ok {
		ts.BurstLimit = int32(sv)
	}

	if sv, ok := m["rate_limit"].(float64); ok {
		ts.RateLimit = sv
	}

	return ts
}

func flattenAPIStages(s []types.ApiStage) []map[string]interface{} {
	stages := make([]map[string]interface{}, 0)

	for _, bd := range s {
		if bd.ApiId != nil && bd.Stage != nil {
			stage := make(map[string]interface{})
			stage["api_id"] = aws.ToString(bd.ApiId)
			stage[names.AttrStage] = aws.ToString(bd.Stage)
			stage["throttle"] = flattenThrottleSettingsMap(bd.Throttle)

			stages = append(stages, stage)
		}
	}

	if len(stages) > 0 {
		return stages
	}

	return nil
}

func flattenThrottleSettings(s *types.ThrottleSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	settings["burst_limit"] = s.BurstLimit
	settings["rate_limit"] = s.RateLimit

	return []map[string]interface{}{settings}
}

func flattenQuotaSettings(s *types.QuotaSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	settings["limit"] = s.Limit
	settings["offset"] = s.Offset
	settings["period"] = s.Period

	return []map[string]interface{}{settings}
}

func expandThrottleSettingsList(tfList []interface{}) map[string]types.ThrottleSettings {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := map[string]types.ThrottleSettings{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := types.ThrottleSettings{}

		if v, ok := tfMap["burst_limit"].(int); ok {
			apiObject.BurstLimit = int32(v)
		}

		if v, ok := tfMap["rate_limit"].(float64); ok {
			apiObject.RateLimit = v
		}

		if v, ok := tfMap[names.AttrPath].(string); ok && v != "" {
			apiObjects[v] = apiObject
		}
	}

	return apiObjects
}

func flattenThrottleSettingsMap(apiObjects map[string]types.ThrottleSettings) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrPath: k,
			"rate_limit":   apiObject.RateLimit,
			"burst_limit":  apiObject.BurstLimit,
		})
	}

	return tfList
}
