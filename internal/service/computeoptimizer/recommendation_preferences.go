package computeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_computeoptimizer_recommendation_preferences", name="Recommendation Preferences")
func ResourceRecommendationPreferences() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecommendationPreferencesCreate,
		ReadWithoutTimeout:   resourceRecommendationPreferencesRead,
		UpdateWithoutTimeout: resourceRecommendationPreferencesUpdate,
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
				Type:        schema.TypeString,
				Optional: true,
				Description: "The status of the enhanced infrastructure metrics recommendation preference to create or update.",
				// TODO Verify validation function type
				// Valid choices: "Active", "Inactive",				
				ValidateDiagFunc: enum.Validate[types.EnhancedInfrastructureMetrics](),
			},
			"external_metrics_preference": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The provider of the external metrics recommendation preference to create or update.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:        schema.TypeString,
							// TODO Verify this is required.
							// Required:    true,
							// Option: true,
							Description: "The external metrics source.",
							// TODO Validate function requirements
							// Only used when ResoureceType == Ec2Instances
							// ValidateFunc: validation.All(),
						},
					},
				},
			},
			"id": {
				Type: schema.TypeString,
				Computed: true,
			},
			"inferred_workload_types": {
				Type: schema.TypeString,
				Optional: true,
				Description: "The provider of the external metrics recommendation preference to create or update.",
				// TODO Verify validation function type
				// Valid choices: "Active", "Inactive",
				ValidateDiagFunc: enum.Validate[types.InferredWorkloadTypesPreference](),
			},
			"look_back_period": {
				Type: schema.TypeString,
				Optional: true,
				Description: "The preference to control the number of days the utilization metrics of the AWS resource are analyzed.",
				// TODO Verify validation function type
				// Valid choices: "DAYS_14", "DAYS_32", "DAYS_93",
				// Default: "DAYS_14",
				ValidateDiagFunc: enum.Validate[types.LookBackPeriodPreference](),
			},
			"preferred_resources": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The preferred resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude_list": {
							Type: schema.TypeList,
							Optional: true,
							Description: "The preference to control which resource type values are considered when generating rightsizing recommendations.",
							Elem: &schema.Schema{
								// TODO Validate appropriate attribute for values.
								"name": {
									Type:        schema.TypeString,
									Description: "The name of the resource.",
								},
							},
						},
						"include_list": {
							Type: schema.TypeList,
							Optional: true,
							Description: "The preferred resource type values to include in the recommendation candidates.",
							Elem: &schema.Schema{
								// TODO Validate appropriate attribute for values.
								"name": {
									Type:        schema.TypeString,
									Description: "The name of the resource.",
								},
							},
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The type of preferred resource to customize.",
							Optional: true,
							// TODO Verify validation function type
							ValidateDiagFunc: enum.Validate[types.PreferredResourceName](),
						},
					},
				},
			},
			"savings_estimation_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The status of the savings estimation mode preference to create or update.",
				// TODO Verify validation function type
				ValidateDiagFunc: enum.Validate[types.SavingsEstimationMode](),

			},
			"scope": {
				Type: schema.TypeSet,
				Optional: true,
				Description: "An object that describes the scope of the recommendation preference to create.",
				Elem: &schema.Schema{
					"name": {
						Type: schema.TypeString,
						Optional: true,
						Description: "The name of the scope.",
						// TODO Verify validation function type
						ValidateDiagFunc: enum.Validate[types.ScopeName](),
					},
					"value": {
						Type: schema.TypeString,
						Optional: true,
						Description: "The value of the scope.",
						// TODO Verify validation function type
						ValidateDiagFunc: enum.Validate[types.ScopeValue](),
					},
				},
			},
			"resource_type": {
				Type:        schema.TypeString,
				Required:    true,
				//TODO Validate name type for EC2 Instance Build
				// Valid choices: "Ec2Instances", "AutoScalingGroup"
				ValidateDiagFunc: enum.Validate[types.ResourceType](),
			},
			"utilization_preferences": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The preference to control the resource’s CPU utilization thresholds - threshold and headroom.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_name": {
							Type:        schema.TypeString,
							Optional: true,
							Description: "The name of the metric.",
						},
						"metric_parameters": {
							Type: schema.TypeSet,
							Optional: true,
							Description: "The metric parameters.",
							Elem: &schema.Schema{
								"headroom":{
									Type: schema.TypeString,
									Optional: true,
									Description: "The headroom percentage.",
									// TODO Verify validation function type
									ValidateDiagFunc: enum.EnumValues[types.CustomizableMetricHeadroom](),
								},
								"threshold": {
									Type: schema.TypeString,
									Optional: true,
									Description: "The threshold percentage.",
									// TODO Verify validation function type
									ValidateDiagFunc: enum.EnumValues[types.CustomizableMetricThreshold](),
								},
							},
						},
					},
				},

			},
		},
	},
}

const (
	ResNameRecommendationPreferences = "Recommendation Preferences"
)

func resourceRecommendationPreferencesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Sections with questions	
	// 2. Populate a create input structure
	// 4. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work. At a minimum, set the
	//    resource ID. E.g., d.SetId(<Identifier, such as AWS ID or ARN>)

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)
	
	// TIP: -- 2. Populate a create input structure
	in := &computeoptimizer.PutRecommendationPreferencesInput{
		ResourceType: types.ResourceType(d.Get("resource_type").(string)),
	}

	if v, ok := d.GetOk("enhanced_infrastructure_metrics"); ok {
		in.EnhancedInfrastructureMetrics = types.EnhancedInfrastructureMetrics(v.(string))
	}

	if v, ok := d.GetOk("external_metrics_preference"); ok && len(v.([]interface{})) > 0 {
		in.ExternalMetricsPreference = expandExternalMetricsPreference(v.([]interface{}))
	}

	if v, ok := d.GetOk("inferred_workload_types"); ok {
		in.InferredWorkloadTypes = types.InferredWorkloadTypesPreference(v.(string))
	}

	if v, ok := d.GetOk("look_back_period"); ok {
		in.LookBackPeriod = types.LookBackPeriodPreference(v.(string))
	}

	if v, ok := d.GetOk("preferred_resources"); ok && len(v.([]interface{})) > 0 {
		in.PreferredResources = expandPreferredResources(v.([]interface{}))
	// Issue with expansion
	}

	if v, ok := d.GetOk("savings_estimation_mode"); ok {
		in.SavingsEstimationMode = types.SavingsEstimationMode(v.(string))
	}

	if v, ok := d.GetOk("scope"); ok && len(v.([]interface{})) > 0 {
		in.Scope = expandScope(v.([]interface{}))
	// Check expansion
	}

	if v, ok := d.GetOk("utilization_preferences"); ok && len(v.([]interface{})) > 0 {
		in.UtilizationPreferences = expandUtilizationPreferences(v.([]interface{}))
		// Issue with expansion
	}
	
	// TIP: -- 3. Call the AWS create function
	out, err := conn.PutRecommendationPreferences(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionCreating, ResNameRecommendationPreferences, d.Get("name").(string), err)...)
	}

	// TODO - Is `out.ResultMetadata == nil` logic needed?
	if out == nil || out.ResultMetadata == nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionCreating, ResNameRecommendationPreferences, d.Get("name").(string), errors.New("empty output"))...)
	}
	
	// TODO What ID type should be used?
	// d.SetId("accountID")
	d.SetId(d.Get("resource_type").(string))
	
	if _, err := waitRecommendationPreferencesCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionWaitingForCreation, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	return append(diags, resourceRecommendationPreferencesRead(ctx, d, meta)...)
}

func resourceRecommendationPreferencesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Sections with questions
	// 4. Set the arguments and attributes
	// 6. Return diags

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)
	
	out, err := findRecommendationPreferencesByID(ctx, conn, d.Id())
	
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Compute Optimizer Recommendation Preferences (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionReading, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	// TODO Verify data types
	d.Set("enhanced_infrastructure_metrics", out.EnhancedInfrastructureMetrics)
	d.Set("inferred_workload_types", *out.InferredWorkloadTypes)
	d.Set("look_back_period", out.LookBackPeriod)
	d.Set("resource_type", out.ResourceType)
	d.Set("savings_estimation_mode", out.SavingsEstimationMode)
	
	if err := d.Set("external_metrics_source", flattenExternalMetricsPreference(out.ExternalMetricsPreference)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionSetting, ResNameComputeOptimizerPreferences, d.Id(), err)...)
	}
	
	if err := d.Set("preferred_resources", flattenPreferredResources(out.PreferredResource)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionSetting, ResNameComputeOptimizerPreferences, d.Id(), err)...)
	}

	if err := d.Set("scope", flattenScope(out.Scope)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionSetting, ResNameComputeOptimizerPreferences, d.Id(), err)...)
	}

	if err := d.Set("utilization_preferences", flattenUtilizationPreferences(out.UtilizationPreference)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionSetting, ResNameComputeOptimizerPreferences, d.Id(), err)...)
	}
		
	return diags
}

func resourceRecommendationPreferencesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Sections with questions	
	// 4. Use a waiter to wait for update to complete

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)
	
	update := false

	in := &computeoptimizer.PutRecommendationPreferencesInput{
		ResourceType: types.ResourceType(d.Id()),
	}

	if !update {
		return diags
	}
	
	log.Printf("[DEBUG] Updating Compute Optimizer Recommendation Preferences (%s): %#v", d.Id(), in)
	out, err := conn.PutRecommendationPreferences(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionUpdating, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	// TIP: -- 4. Use a waiter to wait for update to complete
	// TODO PutRecommendationPreferences returns a 200 if set. What out value should be listed?
	if _, err := waitRecommendationPreferencesUpdated(ctx, conn, aws.ToString(out.ResultMetadata), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionWaitingForUpdate, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	return append(diags, resourceRecommendationPreferencesRead(ctx, d, meta)...)
}

func resourceRecommendationPreferencesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// 3. Call the AWS delete function

	conn := meta.(*conns.AWSClient).ComputeOptimizerClient(ctx)
	
	log.Printf("[INFO] Deleting Compute Optimizer Recommendation Preferences %s", d.Id())
	
	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteRecommendationPreferences(ctx, &computeoptimizer.DeleteRecommendationPreferencesInput{
		// TODO Validate Recommendation Preference Name
		RecommendationPreferenceNames: types.RecommendationPreferenceName(d.State().Attributes["enhanced_infrastructure_metrics", "external_metrics_preference". "inferred_workload_types", "look_back_period", "preferred_resources"]),
		ResourceType: types.ResourceType(d.Id()),
	})
	
	if errs.IsA[*types.ResourceNotFoundException](err){
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionDeleting, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	if _, err := waitRecommendationPreferencesDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.ComputeOptimizer, create.ErrActionWaitingForDeletion, ResNameRecommendationPreferences, d.Id(), err)...)
	}
	
	return diags
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitRecommendationPreferencesCreated(ctx context.Context, conn *computeoptimizer.Client, id string, timeout time.Duration) (*computeoptimizer.RecommendationPreferences, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusRecommendationPreferences(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*computeoptimizer.RecommendationPreferences); ok {
		return out, err
	}

	return nil, err
}

func waitRecommendationPreferencesUpdated(ctx context.Context, conn *computeoptimizer.Client, id string, timeout time.Duration) (*computeoptimizer.RecommendationPreferences, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusRecommendationPreferences(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*computeoptimizer.RecommendationPreferences); ok {
		return out, err
	}

	return nil, err
}

func waitRecommendationPreferencesDeleted(ctx context.Context, conn *computeoptimizer.Client, id string, timeout time.Duration) (*computeoptimizer.RecommendationPreferences, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusRecommendationPreferences(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*computeoptimizer.RecommendationPreferences); ok {
		return out, err
	}

	return nil, err
}

func statusRecommendationPreferences(ctx context.Context, conn *computeoptimizer.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findRecommendationPreferencesByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findRecommendationPreferencesByID(ctx context.Context, conn *computeoptimizer.Client, id string) (*computeoptimizer.RecommendationPreferences, error) {
	in := &computeoptimizer.GetRecommendationPreferencesInput{
		ResourceType: types.ResourceType(id),
	}
	out, err := conn.GetRecommendationPreferences(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err){
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.RecommendationPreferences == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.RecommendationPreferences, nil
}

// TODO Validate flex functions
func flattenExternalMetricsPreference(apiObject *types.ExternalMetricsPreference) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Source; v != nil {
		m["Source"] = aws.ToString(v)
	}

	return m
}

func flattenPreferredResources(apiObject *types.PreferredResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.ExcludeList; v != nil {
		m["exclude_list"] = aws.ToString(v)
	}

	if v := apiObject.IncludeList; v != nil {
		m["include_list"] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		m["name"] = aws.ToString(v)
	}

	return m
}

func flattenScope(apiObject *types.Scope) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		m["name"] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		m["value"] = aws.ToString(v)
	}

	return m
}

func flattenUtilizationPreferences(apiObject *types.UtilizationPreference) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MetricName; v != nil {
		m["metric_name"] = aws.ToString(v)
	}

	if v := apiObject.MetricParameters; v != nil {
		m["metric_parameters"] = v
	}

	return m
}

func expandExternalMetricsPreference(tfList []interface{}) *types.ExternalMetricsPreference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	externalMetricsConfig := &types.ExternalMetricsPreference{
		Source: types.ExternalMetricsSource(p["external_metrics_source"].(string)),
	}

	return externalMetricsConfig
}

func expandPreferredResources(tfList []interface{}) *types.PreferredResource {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	preferredResourceConfig := &types.PreferredResource{
		ExcludeList: []string{p["exclude_list"]},
		IncludeList: []string{p["include_list"]},
		Name:        types.PreferredResourceName(p["name"].(string)),
	}

	return preferredResourceConfig
}

func expandScope(tfList []interface{}) *types.Scope {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	scopeConfig := &types.Scope{
		Name:        types.ScopeName(p["name"].(string)),
		Value:       aws.String(p["value"].(string)),
	}

	return scopeConfig
}

func expandUtilizationPreferences(tfList []interface{}) *types.UtilizationPreference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	p, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	utilizationPreferenceConfig := &types.UtilizationPreference{
		MetricName:        *types.CustomizableMetricHeadroom(p["metric_name"].(string)),
		MetricParameters:       &types.CustomizableMetricParameters{
			Headroom: *types.CustomizableMetricHeadroom(p["metric_parameters"]["headroom"].(map[string]interface{})),
			Threshold: *types.CustomizableMetricThreshold(p["metric_parameters"]["threshold"].(map[string]interface{})),
		},
	}

	return utilizationPreferenceConfig
}