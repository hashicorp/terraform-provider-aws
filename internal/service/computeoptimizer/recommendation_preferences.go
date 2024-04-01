// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"log"

	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"

	// tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_computeoptimizer_recommendation_preferences", name="Recommendation Preferences")
func ResourceRecommendationPreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecommendationPreferencesCreate,
		ReadWithoutTimeout:   resourceRecommendationPreferencesRead,
		// UpdateWithoutTimeout: resourceRecommendationPreferencesUpdate,
		DeleteWithoutTimeout: resourceRecommendationPreferencesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"enhanced_infrastructure_metrics": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Description:      "The status of the enhanced infrastructure metrics recommendation preference to create or update.",
				ValidateDiagFunc: enum.Validate[types.EnhancedInfrastructureMetrics](),
			},
			"external_metrics_preference": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.ExternalMetricsSource](),
						},
					},
				},
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inferred_workload_types": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.InferredWorkloadType](),
			},
			"look_back_period": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.LookBackPeriodPreference](),
			},
			"preferred_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude_list": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"include_list": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"resource_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ResourceType](),
			},
			"savings_estimation_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.SavingsEstimationMode](),
			},
			"scope": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.ScopeName](),
							// Update Validation: https://docs.aws.amazon.com/compute-optimizer/latest/APIReference/API_Scope.html
							// DMS Endpoint Case?
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							// Update Validation: https://docs.aws.amazon.com/compute-optimizer/latest/APIReference/API_Scope.html
							// DMS Endpoint Case?
						},
					},
				},
			},
			// names.AttrTags:    tftags.TagsSchema(),
			// names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"utilization_preferences": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"metric_parameters": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"headroom": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[types.CustomizableMetricHeadroom](),
									},
									"threshold": {
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[types.CustomizableMetricThreshold](),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	ResNameRecommendationPreferences = "Recommendation Preferences"
)

func resourceRecommendationPreferencesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)

	// resourceType := d.Get("resource_type").(string)
	resourceType := "Ec2Instance"
	in := &computeoptimizer.PutRecommendationPreferencesInput{
		// ResourceType: types.ResourceType(*aws.String(d.Get("resource_type").(string))),
		ResourceType: types.ResourceType(resourceType),
	}

	recommendationPreferences := []string{}
	recommendationPreferences = append(recommendationPreferences, resourceType)

	if v, ok := d.GetOk("enhanced_infrastructure_metrics"); ok {
		in.EnhancedInfrastructureMetrics = types.EnhancedInfrastructureMetrics(*aws.String(v.(string)))
		recommendationPreferences = append(recommendationPreferences, "EnhancedInfrastructureMetrics")
	}

	if v, ok := d.GetOk("external_metrics_preference"); ok && len(v.([]interface{})) > 0 {
		in.ExternalMetricsPreference = expandExternalMetricsPreference(v.([]interface{}))
		recommendationPreferences = append(recommendationPreferences, "ExternalMetricsPreference")
	}

	if v, ok := d.GetOk("inferred_workload_types"); ok {
		in.InferredWorkloadTypes = types.InferredWorkloadTypesPreference(*aws.String(v.(string)))
		recommendationPreferences = append(recommendationPreferences, "InferredWorkloadTypes")
	}

	if v, ok := d.GetOk("look_back_period"); ok {
		in.LookBackPeriod = types.LookBackPeriodPreference(*aws.String(v.(string)))
		recommendationPreferences = append(recommendationPreferences, "LookBackPeriodPreference")
	}

	if v, ok := d.GetOk("preferred_resources"); ok && len(v.([]interface{})) > 0 {
		in.PreferredResources = expandPreferredResources(v.([]interface{}))
		recommendationPreferences = append(recommendationPreferences, "PreferredResources")
	}

	if v, ok := d.GetOk("savings_estimation_mode"); ok {
		in.SavingsEstimationMode = types.SavingsEstimationMode(*aws.String(v.(string)))
	}

	if v, ok := d.GetOk("scope"); ok && len(v.([]interface{})) > 0 {
		in.Scope = expandScope(v.([]interface{}))
	}

	if v, ok := d.GetOk("utilization_preferences"); ok && len(v.([]interface{})) > 0 {
		in.UtilizationPreferences = expandUtilizationPreferences(v.([]interface{}))
		recommendationPreferences = append(recommendationPreferences, "UtilizationPreferences")
	}

	out, err := conn.PutRecommendationPreferences(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.ComputeOptimizer, create.ErrActionCreating, ResNameRecommendationPreferences, d.Get("name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.ComputeOptimizer, create.ErrActionCreating, ResNameRecommendationPreferences, d.Get("name").(string), errors.New("empty output"))
	}

	id := preferencesCreateResourceID(recommendationPreferences)

	d.SetId(id)

	return append(diags, resourceRecommendationPreferencesRead(ctx, d, meta)...)
}

func resourceRecommendationPreferencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)

	out, err := findRecommendationPreferencesByResource(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ComputeOptimizer RecommendationPreferences (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ComputeOptimizer, create.ErrActionReading, ResNameRecommendationPreferences, d.Id(), err)
	}

	// TODO - How to set a list?
	outputData := out.RecommendationPreferencesDetails
	d.Set("resource_type", outputData[0].ResourceType)

	return diags
}

func resourceRecommendationPreferencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)

	log.Printf("[INFO] Deleting ComputeOptimizer RecommendationPreferences %s", d.Id())

	var resourceType string
	var recommendationPreferencesNames []types.RecommendationPreferenceName
	var err error

	resourceType, recommendationPreferencesNames, err = preferencesParseResourceID(d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.ComputeOptimizer, create.ErrActionDeleting, ResNameRecommendationPreferences, d.Id(), err)
	}

	_, err = conn.DeleteRecommendationPreferences(ctx, &computeoptimizer.DeleteRecommendationPreferencesInput{
		ResourceType:                  types.ResourceType(resourceType),
		RecommendationPreferenceNames: recommendationPreferencesNames,
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.ComputeOptimizer, create.ErrActionDeleting, ResNameRecommendationPreferences, d.Id(), err)
	}

	return diags
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func statusRecommendationPreferences(ctx context.Context, conn *computeoptimizer.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findRecommendationPreferencesByResource(ctx, conn, id)
		if err != nil {
			return nil, "", err
		}

		var recommendationStatus string

		if tfresource.NotFound(err) {
			return nil, "", nil
		} else {
			recommendationStatus = "ACTIVE"
		}

		return out, recommendationStatus, nil
	}
}

func findRecommendationPreferencesByResource(ctx context.Context, conn *computeoptimizer.Client, resource string) (*computeoptimizer.GetRecommendationPreferencesOutput, error) { // doesn't returns subtype types.RecommendationPreferencesDetails with details
	in := &computeoptimizer.GetRecommendationPreferencesInput{
		ResourceType: types.ResourceType(resource),
	}
	out, err := conn.GetRecommendationPreferences(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.RecommendationPreferencesDetails == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return something you can use in
// a d.Set() call. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
// func flattenComplexArgument(apiObject *computeoptimizer.ComplexArgument) map[string]interface{} {
// 	if apiObject == nil {
// 		return nil
// 	}

// 	m := map[string]interface{}{}

// 	if v := apiObject.SubFieldOne; v != nil {
// 		m["sub_field_one"] = aws.ToString(v)
// 	}

// 	if v := apiObject.SubFieldTwo; v != nil {
// 		m["sub_field_two"] = aws.ToString(v)
// 	}

// 	return m
// }

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
// func flattenComplexArguments(apiObjects []*computeoptimizer.ComplexArgument) []interface{} {
// 	if len(apiObjects) == 0 {
// 		return nil
// 	}

// 	var l []interface{}

// 	for _, apiObject := range apiObjects {
// 		if apiObject == nil {
// 			continue
// 		}

// 		l = append(l, flattenComplexArgument(apiObject))
// 	}

// 	return l
// }

func expandExternalMetricsPreference(tfList []interface{}) *types.ExternalMetricsPreference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	externalMetricsPreference := &types.ExternalMetricsPreference{
		Source: types.ExternalMetricsSource(p["source"].(string)),
	}

	return externalMetricsPreference
}

func expandPreferredResources(tfList []interface{}) []types.PreferredResource {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var preferredResourceArray []types.PreferredResource

	preferredResourceConfig := &types.PreferredResource{
		ExcludeList: []string{p["exclude_list"].(string)},
		IncludeList: []string{p["include_list"].(string)},
		Name:        types.PreferredResourceName(p["name"].(string)),
	}

	preferredResourceArray = append(preferredResourceArray, *preferredResourceConfig)

	return preferredResourceArray
}

func expandScope(tfList []interface{}) *types.Scope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	scope := &types.Scope{
		Name:  types.ScopeName(p["name"].(string)),
		Value: aws.String(p["value"].(string)),
	}

	return scope
}

func expandUtilizationPreferences(tfList []interface{}) []types.UtilizationPreference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var utilizationPreferenceArray []types.UtilizationPreference

	utilizationPreference := &types.UtilizationPreference{
		MetricName: types.CustomizableMetricName(p["metric_name"].(string)),
		MetricParameters: &types.CustomizableMetricParameters{
			Headroom:  types.CustomizableMetricHeadroom(p["headroom"].(string)),
			Threshold: types.CustomizableMetricThreshold(p["threshold"].(string)),
		},
	}

	utilizationPreferenceArray = append(utilizationPreferenceArray, *utilizationPreference)

	return utilizationPreferenceArray
}

const preferencesIDSeparator = ":"

func preferencesCreateResourceID(recommendationPreferencesName []string) string {

	id := strings.Join(recommendationPreferencesName, preferencesIDSeparator)

	return id
}

func preferencesParseResourceID(id string) (string, []types.RecommendationPreferenceName, error) {
	parts := strings.Split(id, preferencesIDSeparator)

	if len(parts) < 2 || parts[0] == "" || len(parts[1:]) == 0 {
		return "", []types.RecommendationPreferenceName{}, fmt.Errorf("unexpected format of ID (%[1]s), expected resource_type:%[2]s<...recommendation_preference_names>", id, preferencesIDSeparator)
	}

	var recommendationPreferenceNames []types.RecommendationPreferenceName

	for _, part := range parts[1:] {
		recommendationPreferenceNames = append(recommendationPreferenceNames, types.RecommendationPreferenceName(part))
	}

	return parts[0], recommendationPreferenceNames, nil
}
