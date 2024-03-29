// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_usage_plan", name="Usage Plan")
// @Tags(identifierAttribute="arn")
func ResourceUsagePlan() *schema.Resource {
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
						"stage": {
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
									"path": {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
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
							Type:         schema.TypeString,
							Required:     true, // Required as not removable
							ValidateFunc: validation.StringInSlice(apigateway.QuotaPeriodType_Values(), false),
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
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	name := d.Get("name").(string)
	input := &apigateway.CreateUsagePlanInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("api_stages"); ok && v.(*schema.Set).Len() > 0 {
		input.ApiStages = expandAPIStages(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
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

	output, err := conn.CreateUsagePlanWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Usage Plan (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	// Handle case of adding the product code since not addable when
	// creating the Usage Plan initially.
	if v, ok := d.GetOk("product_code"); ok {
		input := &apigateway.UpdateUsagePlanInput{
			PatchOperations: []*apigateway.PatchOperation{
				{
					Op:    aws.String(apigateway.OpAdd),
					Path:  aws.String("/productCode"),
					Value: aws.String(v.(string)),
				},
			},
			UsagePlanId: aws.String(d.Id()),
		}

		_, err = conn.UpdateUsagePlanWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding API Gateway Usage Plan (%s) product code: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsagePlanRead(ctx, d, meta)...)
}

func resourceUsagePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	up, err := FindUsagePlanByID(ctx, conn, d.Id())

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
	d.Set("arn", arn)
	d.Set("description", up.Description)
	d.Set("name", up.Name)
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
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		operations := make([]*apigateway.PatchOperation, 0)

		if d.HasChange("name") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/name"),
				Value: aws.String(d.Get("name").(string)),
			})
		}

		if d.HasChange("description") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			})
		}

		if d.HasChange("product_code") {
			if v, ok := d.GetOk("product_code"); ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpReplace),
					Path:  aws.String("/productCode"),
					Value: aws.String(v.(string)),
				})
			} else {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
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
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpRemove),
					Path:  aws.String("/apiStages"),
					Value: aws.String(fmt.Sprintf("%s:%s", m["api_id"].(string), m["stage"].(string))),
				})
			}

			// Handle additions
			if len(ns) > 0 {
				for _, v := range ns {
					m := v.(map[string]interface{})
					id := fmt.Sprintf("%s:%s", m["api_id"].(string), m["stage"].(string))
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
						Path:  aws.String("/apiStages"),
						Value: aws.String(id),
					})
					if t, ok := m["throttle"].(*schema.Set); ok && t.Len() > 0 {
						for _, throttle := range t.List() {
							th := throttle.(map[string]interface{})
							operations = append(operations, &apigateway.PatchOperation{
								Op:    aws.String(apigateway.OpReplace),
								Path:  aws.String(fmt.Sprintf("/apiStages/%s/throttle/%s/rateLimit", id, th["path"].(string))),
								Value: aws.String(strconv.FormatFloat(th["rate_limit"].(float64), 'f', -1, 64)),
							})
							operations = append(operations, &apigateway.PatchOperation{
								Op:    aws.String(apigateway.OpReplace),
								Path:  aws.String(fmt.Sprintf("/apiStages/%s/throttle/%s/burstLimit", id, th["path"].(string))),
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
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String("/throttle"),
				})
			}

			if len(diff) > 0 {
				d := diff[0].(map[string]interface{})

				// Handle Replaces
				if o != nil && n != nil {
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/throttle/rateLimit"),
						Value: aws.String(strconv.FormatFloat(d["rate_limit"].(float64), 'f', -1, 64)),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/throttle/burstLimit"),
						Value: aws.String(strconv.Itoa(d["burst_limit"].(int))),
					})
				}

				// Handle Additions
				if o == nil && n != nil {
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
						Path:  aws.String("/throttle/rateLimit"),
						Value: aws.String(strconv.FormatFloat(d["rate_limit"].(float64), 'f', -1, 64)),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
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
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
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
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/quota/limit"),
						Value: aws.String(strconv.Itoa(d["limit"].(int))),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/quota/offset"),
						Value: aws.String(strconv.Itoa(d["offset"].(int))),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/quota/period"),
						Value: aws.String(d["period"].(string)),
					})
				}

				// Handle Additions
				if o == nil && n != nil {
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
						Path:  aws.String("/quota/limit"),
						Value: aws.String(strconv.Itoa(d["limit"].(int))),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
						Path:  aws.String("/quota/offset"),
						Value: aws.String(strconv.Itoa(d["offset"].(int))),
					})
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpAdd),
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

		_, err := conn.UpdateUsagePlanWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Usage Plan (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsagePlanRead(ctx, d, meta)...)
}

func resourceUsagePlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	// Removing existing api stages associated
	if apistages, ok := d.GetOk("api_stages"); ok {
		stages := apistages.(*schema.Set)
		operations := []*apigateway.PatchOperation{}

		for _, v := range stages.List() {
			sv := v.(map[string]interface{})

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpRemove),
				Path:  aws.String("/apiStages"),
				Value: aws.String(fmt.Sprintf("%s:%s", sv["api_id"].(string), sv["stage"].(string))),
			})
		}

		_, err := conn.UpdateUsagePlanWithContext(ctx, &apigateway.UpdateUsagePlanInput{
			PatchOperations: operations,
			UsagePlanId:     aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "removing API Gateway Usage Plan (%s) API stages: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan: %s", d.Id())
	_, err := conn.DeleteUsagePlanWithContext(ctx, &apigateway.DeleteUsagePlanInput{
		UsagePlanId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Usage Plan (%s): %s", d.Id(), err)
	}

	return diags
}

func FindUsagePlanByID(ctx context.Context, conn *apigateway.APIGateway, id string) (*apigateway.UsagePlan, error) {
	input := &apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(id),
	}

	output, err := conn.GetUsagePlanWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
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

func expandAPIStages(s *schema.Set) []*apigateway.ApiStage {
	stages := []*apigateway.ApiStage{}

	for _, stageRaw := range s.List() {
		stage := &apigateway.ApiStage{}
		mStage := stageRaw.(map[string]interface{})

		if v, ok := mStage["api_id"].(string); ok && v != "" {
			stage.ApiId = aws.String(v)
		}

		if v, ok := mStage["stage"].(string); ok && v != "" {
			stage.Stage = aws.String(v)
		}

		if v, ok := mStage["throttle"].(*schema.Set); ok && v.Len() > 0 {
			stage.Throttle = expandThrottleSettingsList(v.List())
		}

		stages = append(stages, stage)
	}

	return stages
}

func expandQuotaSettings(l []interface{}) *apigateway.QuotaSettings {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	qs := &apigateway.QuotaSettings{}

	if v, ok := m["limit"].(int); ok {
		qs.Limit = aws.Int64(int64(v))
	}

	if v, ok := m["offset"].(int); ok {
		qs.Offset = aws.Int64(int64(v))
	}

	if v, ok := m["period"].(string); ok && v != "" {
		qs.Period = aws.String(v)
	}

	return qs
}

func expandThrottleSettings(l []interface{}) *apigateway.ThrottleSettings {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	ts := &apigateway.ThrottleSettings{}

	if sv, ok := m["burst_limit"].(int); ok {
		ts.BurstLimit = aws.Int64(int64(sv))
	}

	if sv, ok := m["rate_limit"].(float64); ok {
		ts.RateLimit = aws.Float64(sv)
	}

	return ts
}

func flattenAPIStages(s []*apigateway.ApiStage) []map[string]interface{} {
	stages := make([]map[string]interface{}, 0)

	for _, bd := range s {
		if bd.ApiId != nil && bd.Stage != nil {
			stage := make(map[string]interface{})
			stage["api_id"] = aws.StringValue(bd.ApiId)
			stage["stage"] = aws.StringValue(bd.Stage)
			stage["throttle"] = flattenThrottleSettingsMap(bd.Throttle)

			stages = append(stages, stage)
		}
	}

	if len(stages) > 0 {
		return stages
	}

	return nil
}

func flattenThrottleSettings(s *apigateway.ThrottleSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.BurstLimit != nil {
		settings["burst_limit"] = aws.Int64Value(s.BurstLimit)
	}

	if s.RateLimit != nil {
		settings["rate_limit"] = aws.Float64Value(s.RateLimit)
	}

	return []map[string]interface{}{settings}
}

func flattenQuotaSettings(s *apigateway.QuotaSettings) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	if s.Limit != nil {
		settings["limit"] = aws.Int64Value(s.Limit)
	}

	if s.Offset != nil {
		settings["offset"] = aws.Int64Value(s.Offset)
	}

	if s.Period != nil {
		settings["period"] = aws.StringValue(s.Period)
	}

	return []map[string]interface{}{settings}
}

func expandThrottleSettingsList(tfList []interface{}) map[string]*apigateway.ThrottleSettings {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := map[string]*apigateway.ThrottleSettings{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &apigateway.ThrottleSettings{}

		if v, ok := tfMap["burst_limit"].(int); ok {
			apiObject.BurstLimit = aws.Int64(int64(v))
		}

		if v, ok := tfMap["rate_limit"].(float64); ok {
			apiObject.RateLimit = aws.Float64(v)
		}

		if v, ok := tfMap["path"].(string); ok && v != "" {
			apiObjects[v] = apiObject
		}
	}

	return apiObjects
}

func flattenThrottleSettingsMap(apiObjects map[string]*apigateway.ThrottleSettings) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, map[string]interface{}{
			"path":        k,
			"rate_limit":  aws.Float64Value(apiObject.RateLimit),
			"burst_limit": aws.Int64Value(apiObject.BurstLimit),
		})
	}

	return tfList
}
