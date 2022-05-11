package autoscaling

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"adjustment_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cooldown": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"estimated_instance_warmup": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"metric_aggregation_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"min_adjustment_magnitude": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "SimpleScaling", // preserve AWS's default to make validation easier.
				ValidateFunc: validation.StringInSlice([]string{
					"SimpleScaling",
					"StepScaling",
					"TargetTrackingScaling",
					"PredictiveScaling",
				}, false),
			},
			"predictive_scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_specification": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"customized_capacity_metric_specification": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metric_data_queries": func() *schema.Schema {
													schema := customizedMetricDataQuerySchema()
													return schema
												}(),
											},
										},
									},
									"customized_load_metric_specification": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_load_metric_specification"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metric_data_queries": func() *schema.Schema {
													schema := customizedMetricDataQuerySchema()
													return schema
												}(),
											},
										},
									},
									"customized_scaling_metric_specification": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.predefined_scaling_metric_specification"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metric_data_queries": func() *schema.Schema {
													schema := customizedMetricDataQuerySchema()
													return schema
												}(),
											},
										},
									},
									"predefined_metric_pair_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"predefined_metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(autoscaling.PredefinedMetricPairType_Values(), false),
												},
												"resource_label": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"predefined_scaling_metric_specification": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.customized_scaling_metric_specification"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"predefined_metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(autoscaling.PredefinedScalingMetricType_Values(), false),
												},
												"resource_label": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"predefined_load_metric_specification": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"predictive_scaling_configuration.0.metric_specification.0.customized_load_metric_specification"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"predefined_metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(autoscaling.PredefinedLoadMetricType_Values(), false),
												},
												"resource_label": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"target_value": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"max_capacity_breach_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "HonorMaxCapacity",
							ValidateFunc: validation.StringInSlice(autoscaling.PredictiveScalingMaxCapacityBreachBehavior_Values(), false),
						},
						"max_capacity_buffer": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 100),
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "ForecastOnly",
							ValidateFunc: validation.StringInSlice(autoscaling.PredictiveScalingMode_Values(), false),
						},
						"scheduling_buffer_time": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
						},
					},
				},
			},
			"scaling_adjustment": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"step_adjustment"},
			},
			"step_adjustment": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"scaling_adjustment"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_interval_lower_bound": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"metric_interval_upper_bound": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"scaling_adjustment": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				Set: resourceScalingAdjustmentHash,
			},
			"target_tracking_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"predefined_metric_specification": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"target_tracking_configuration.0.customized_metric_specification"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"predefined_metric_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"resource_label": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"customized_metric_specification": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"target_tracking_configuration.0.predefined_metric_specification"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_dimension": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"metric_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Required: true,
									},
									"statistic": {
										Type:     schema.TypeString,
										Required: true,
									},
									"unit": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"target_value": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"disable_scale_in": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

// All predictive scaling customized metrics shares same metric data query schema
func customizedMetricDataQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 10,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"expression": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 1023),
				},
				"id": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"label": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 2047),
				},
				"metric_stat": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"metric": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"dimensions": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"name": {
														Type:     schema.TypeString,
														Required: true,
													},
													"value": {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
										"metric_name": {
											Type:     schema.TypeString,
											Required: true,
										},
										"namespace": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
							"stat": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 100),
							},
							"unit": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"return_data": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
			},
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	params, err := getPutScalingPolicyInput(d)
	log.Printf("[DEBUG] AutoScaling PutScalingPolicy on Create: %#v", params)
	if err != nil {
		return err
	}

	resp, err := conn.PutScalingPolicy(&params)
	if err != nil {
		return fmt.Errorf("Error putting scaling policy: %s", err)
	}

	d.Set("arn", resp.PolicyARN)
	d.SetId(d.Get("name").(string))
	log.Printf("[INFO] AutoScaling Scaling PolicyARN: %s", d.Get("arn").(string))

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	p, err := getPolicy(d, meta)
	if err != nil {
		return err
	}
	if p == nil && !d.IsNewResource() {
		log.Printf("[WARN] Autoscaling Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Read Scaling Policy: ASG: %s, SP: %s, Obj: %s", d.Get("autoscaling_group_name"), d.Get("name"), p)

	d.Set("adjustment_type", p.AdjustmentType)
	d.Set("autoscaling_group_name", p.AutoScalingGroupName)
	d.Set("cooldown", p.Cooldown)
	d.Set("estimated_instance_warmup", p.EstimatedInstanceWarmup)
	d.Set("metric_aggregation_type", p.MetricAggregationType)
	d.Set("policy_type", p.PolicyType)
	if p.MinAdjustmentMagnitude != nil {
		d.Set("min_adjustment_magnitude", p.MinAdjustmentMagnitude)
	}
	d.Set("arn", p.PolicyARN)
	d.Set("name", p.PolicyName)
	d.Set("scaling_adjustment", p.ScalingAdjustment)
	if err := d.Set("predictive_scaling_configuration", flattenPredictiveScalingConfig(p.PredictiveScalingConfiguration)); err != nil {
		return fmt.Errorf("error setting predictive_scaling_configuration: %s", err)
	}
	if err := d.Set("step_adjustment", FlattenStepAdjustments(p.StepAdjustments)); err != nil {
		return fmt.Errorf("error setting step_adjustment: %s", err)
	}
	if err := d.Set("target_tracking_configuration", flattenTargetTrackingConfiguration(p.TargetTrackingConfiguration)); err != nil {
		return fmt.Errorf("error setting target_tracking_configuration: %s", err)
	}

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	params, inputErr := getPutScalingPolicyInput(d)
	log.Printf("[DEBUG] AutoScaling PutScalingPolicy on Update: %#v", params)
	if inputErr != nil {
		return inputErr
	}

	_, err := conn.PutScalingPolicy(&params)
	if err != nil {
		return err
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	p, err := getPolicy(d, meta)
	if err != nil {
		return err
	}
	if p == nil {
		return nil
	}

	params := autoscaling.DeletePolicyInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		PolicyName:           aws.String(d.Get("name").(string)),
	}
	log.Printf("[DEBUG] Deleting Autoscaling Policy opts: %s", params)
	if _, err := conn.DeletePolicy(&params); err != nil {
		return fmt.Errorf("Autoscaling Scaling Policy: %s ", err)
	}

	return nil
}

func resourcePolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <asg-name>/<policy-name>", d.Id())
	}

	asgName := idParts[0]
	policyName := idParts[1]

	d.Set("name", policyName)
	d.Set("autoscaling_group_name", asgName)
	d.SetId(policyName)

	return []*schema.ResourceData{d}, nil
}

// PutScalingPolicy can safely resend all parameters without destroying the
// resource, so create and update can share this common function. It will error
// if certain mutually exclusive values are set.
func getPutScalingPolicyInput(d *schema.ResourceData) (autoscaling.PutScalingPolicyInput, error) {
	var params = autoscaling.PutScalingPolicyInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		PolicyName:           aws.String(d.Get("name").(string)),
	}

	// get policy_type first as parameter support depends on policy type
	policyType := d.Get("policy_type")
	params.PolicyType = aws.String(policyType.(string))

	// This parameter is supported if the policy type is SimpleScaling or StepScaling.
	if v, ok := d.GetOk("adjustment_type"); ok && (policyType == "SimpleScaling" || policyType == "StepScaling") {
		params.AdjustmentType = aws.String(v.(string))
	}

	if predictiveScalingConfigFlat := d.Get("predictive_scaling_configuration").([]interface{}); len(predictiveScalingConfigFlat) > 0 {
		params.PredictiveScalingConfiguration = expandPredictiveScalingConfig(predictiveScalingConfigFlat)
	}

	// This parameter is supported if the policy type is SimpleScaling.
	if v, ok := d.GetOkExists("cooldown"); ok {
		// 0 is allowed as placeholder even if policyType is not supported
		params.Cooldown = aws.Int64(int64(v.(int)))
		if v.(int) != 0 && policyType != "SimpleScaling" {
			return params, fmt.Errorf("cooldown is only supported for policy type SimpleScaling")
		}
	}

	// This parameter is supported if the policy type is StepScaling or TargetTrackingScaling.
	if v, ok := d.GetOkExists("estimated_instance_warmup"); ok {
		// 0 is NOT allowed as placeholder if policyType is not supported
		if policyType == "StepScaling" || policyType == "TargetTrackingScaling" {
			params.EstimatedInstanceWarmup = aws.Int64(int64(v.(int)))
		}
		if v.(int) != 0 && policyType != "StepScaling" && policyType != "TargetTrackingScaling" {
			return params, fmt.Errorf("estimated_instance_warmup is only supported for policy type StepScaling and TargetTrackingScaling")
		}
	}

	// This parameter is supported if the policy type is StepScaling.
	if v, ok := d.GetOk("metric_aggregation_type"); ok && policyType == "StepScaling" {
		params.MetricAggregationType = aws.String(v.(string))
	}

	// MinAdjustmentMagnitude is supported if the policy type is SimpleScaling or StepScaling.
	if v, ok := d.GetOkExists("min_adjustment_magnitude"); ok && v.(int) != 0 && (policyType == "SimpleScaling" || policyType == "StepScaling") {
		params.MinAdjustmentMagnitude = aws.Int64(int64(v.(int)))
	}

	// This parameter is required if the policy type is SimpleScaling and not supported otherwise.
	//if policy_type=="SimpleScaling" then scaling_adjustment is required and 0 is allowed
	if v, ok := d.GetOkExists("scaling_adjustment"); ok {
		// 0 is NOT allowed as placeholder if policyType is not supported
		if policyType == "SimpleScaling" {
			params.ScalingAdjustment = aws.Int64(int64(v.(int)))
		}
		if v.(int) != 0 && policyType != "SimpleScaling" {
			return params, fmt.Errorf("scaling_adjustment is only supported for policy type SimpleScaling")
		}
	} else if !ok && policyType == "SimpleScaling" {
		return params, fmt.Errorf("scaling_adjustment is required for policy type SimpleScaling")
	}

	// This parameter is required if the policy type is StepScaling and not supported otherwise.
	if v, ok := d.GetOk("step_adjustment"); ok {
		steps, err := ExpandStepAdjustments(v.(*schema.Set).List())
		if err != nil {
			return params, fmt.Errorf("metric_interval_lower_bound and metric_interval_upper_bound must be strings!")
		}
		params.StepAdjustments = steps
		if len(steps) != 0 && policyType != "StepScaling" {
			return params, fmt.Errorf("step_adjustment is only supported for policy type StepScaling")
		}
	} else if !ok && policyType == "StepScaling" {
		return params, fmt.Errorf("step_adjustment is required for policy type StepScaling")
	}

	// This parameter is required if the policy type is TargetTrackingScaling and not supported otherwise.
	if v, ok := d.GetOk("target_tracking_configuration"); ok {
		params.TargetTrackingConfiguration = expandTargetTrackingConfiguration(v.([]interface{}))
		if policyType != "TargetTrackingScaling" {
			return params, fmt.Errorf("target_tracking_configuration is only supported for policy type TargetTrackingScaling")
		}
	} else if !ok && policyType == "TargetTrackingScaling" {
		return params, fmt.Errorf("target_tracking_configuration is required for policy type TargetTrackingScaling")
	}

	return params, nil
}

func getPolicy(d *schema.ResourceData, meta interface{}) (*autoscaling.ScalingPolicy, error) {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	params := autoscaling.DescribePoliciesInput{
		AutoScalingGroupName: aws.String(d.Get("autoscaling_group_name").(string)),
		PolicyNames:          []*string{aws.String(d.Get("name").(string))},
	}

	log.Printf("[DEBUG] AutoScaling Scaling Policy Describe Params: %#v", params)
	resp, err := conn.DescribePolicies(&params)
	if err != nil {
		//A ValidationError here can mean that either the Policy is missing OR the Autoscaling Group is missing
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "ValidationError" {
			log.Printf("[WARN] Autoscaling Policy (%s) not found, removing from state", d.Id())
			d.SetId("")

			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving scaling policies: %s", err)
	}

	// find scaling policy
	name := d.Get("name")
	for idx, sp := range resp.ScalingPolicies {
		if sp == nil {
			continue
		}

		if aws.StringValue(sp.PolicyName) == name {
			return resp.ScalingPolicies[idx], nil
		}
	}
	// policy not found
	return nil, nil
}

func resourceScalingAdjustmentHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["metric_interval_lower_bound"]; ok {
		buf.WriteString(fmt.Sprintf("%f-", v))
	}
	if v, ok := m["metric_interval_upper_bound"]; ok {
		buf.WriteString(fmt.Sprintf("%f-", v))
	}
	buf.WriteString(fmt.Sprintf("%d-", m["scaling_adjustment"].(int)))

	return create.StringHashcode(buf.String())
}

func expandTargetTrackingConfiguration(configs []interface{}) *autoscaling.TargetTrackingConfiguration {
	if len(configs) < 1 {
		return nil
	}

	config := configs[0].(map[string]interface{})

	result := &autoscaling.TargetTrackingConfiguration{}

	result.TargetValue = aws.Float64(config["target_value"].(float64))
	if v, ok := config["disable_scale_in"]; ok {
		result.DisableScaleIn = aws.Bool(v.(bool))
	}
	if v, ok := config["predefined_metric_specification"]; ok && len(v.([]interface{})) > 0 {
		spec := v.([]interface{})[0].(map[string]interface{})
		predSpec := &autoscaling.PredefinedMetricSpecification{
			PredefinedMetricType: aws.String(spec["predefined_metric_type"].(string)),
		}
		if val, ok := spec["resource_label"]; ok && val.(string) != "" {
			predSpec.ResourceLabel = aws.String(val.(string))
		}
		result.PredefinedMetricSpecification = predSpec
	}
	if v, ok := config["customized_metric_specification"]; ok && len(v.([]interface{})) > 0 {
		spec := v.([]interface{})[0].(map[string]interface{})
		customSpec := &autoscaling.CustomizedMetricSpecification{
			Namespace:  aws.String(spec["namespace"].(string)),
			MetricName: aws.String(spec["metric_name"].(string)),
			Statistic:  aws.String(spec["statistic"].(string)),
		}
		if val, ok := spec["unit"]; ok && len(val.(string)) > 0 {
			customSpec.Unit = aws.String(val.(string))
		}
		if val, ok := spec["metric_dimension"]; ok {
			dims := val.([]interface{})
			metDimList := make([]*autoscaling.MetricDimension, len(dims))
			for i := range metDimList {
				dim := dims[i].(map[string]interface{})
				md := &autoscaling.MetricDimension{
					Name:  aws.String(dim["name"].(string)),
					Value: aws.String(dim["value"].(string)),
				}
				metDimList[i] = md
			}
			customSpec.Dimensions = metDimList
		}
		result.CustomizedMetricSpecification = customSpec
	}
	return result
}

func expandPredictiveScalingConfig(predictiveScalingConfigSlice []interface{}) *autoscaling.PredictiveScalingConfiguration {
	if predictiveScalingConfigSlice == nil || len(predictiveScalingConfigSlice) < 1 {
		return nil
	}
	predictiveScalingConfigFlat := predictiveScalingConfigSlice[0].(map[string]interface{})
	predictiveScalingConfig := &autoscaling.PredictiveScalingConfiguration{
		MetricSpecifications:      expandPredictiveScalingMetricSpecifications(predictiveScalingConfigFlat["metric_specification"].([]interface{})),
		MaxCapacityBreachBehavior: aws.String(predictiveScalingConfigFlat["max_capacity_breach_behavior"].(string)),
		Mode:                      aws.String(predictiveScalingConfigFlat["mode"].(string)),
	}
	if v, null, _ := nullable.Int(predictiveScalingConfigFlat["max_capacity_buffer"].(string)).Value(); !null {
		predictiveScalingConfig.MaxCapacityBuffer = aws.Int64(v)
	}
	if v, null, _ := nullable.Int(predictiveScalingConfigFlat["scheduling_buffer_time"].(string)).Value(); !null {
		predictiveScalingConfig.SchedulingBufferTime = aws.Int64(v)
	}
	return predictiveScalingConfig
}

func expandPredictiveScalingMetricSpecifications(metricSpecificationsSlice []interface{}) []*autoscaling.PredictiveScalingMetricSpecification {
	if metricSpecificationsSlice == nil || len(metricSpecificationsSlice) < 1 {
		return nil
	}
	metricSpecificationsFlat := metricSpecificationsSlice[0].(map[string]interface{})
	metricSpecification := &autoscaling.PredictiveScalingMetricSpecification{
		CustomizedCapacityMetricSpecification: expandCustomizedCapacityMetricSpecification(metricSpecificationsFlat["customized_capacity_metric_specification"].([]interface{})),
		CustomizedLoadMetricSpecification:     expandCustomizedLoadMetricSpecification(metricSpecificationsFlat["customized_load_metric_specification"].([]interface{})),
		CustomizedScalingMetricSpecification:  expandCustomizedScalingMetricSpecification(metricSpecificationsFlat["customized_scaling_metric_specification"].([]interface{})),
		PredefinedLoadMetricSpecification:     expandPredefinedLoadMetricSpecification(metricSpecificationsFlat["predefined_load_metric_specification"].([]interface{})),
		PredefinedMetricPairSpecification:     expandPredefinedMetricPairSpecification(metricSpecificationsFlat["predefined_metric_pair_specification"].([]interface{})),
		PredefinedScalingMetricSpecification:  expandPredefinedScalingMetricSpecification(metricSpecificationsFlat["predefined_scaling_metric_specification"].([]interface{})),
		TargetValue:                           aws.Float64(float64(metricSpecificationsFlat["target_value"].(int))),
	}
	return []*autoscaling.PredictiveScalingMetricSpecification{metricSpecification}
}

func expandPredefinedLoadMetricSpecification(predefinedLoadMetricSpecificationSlice []interface{}) *autoscaling.PredictiveScalingPredefinedLoadMetric {
	if predefinedLoadMetricSpecificationSlice == nil || len(predefinedLoadMetricSpecificationSlice) < 1 {
		return nil
	}
	predefinedLoadMetricSpecificationFlat := predefinedLoadMetricSpecificationSlice[0].(map[string]interface{})
	predefinedLoadMetricSpecification := &autoscaling.PredictiveScalingPredefinedLoadMetric{
		PredefinedMetricType: aws.String(predefinedLoadMetricSpecificationFlat["predefined_metric_type"].(string)),
		ResourceLabel:        aws.String(predefinedLoadMetricSpecificationFlat["resource_label"].(string)),
	}
	return predefinedLoadMetricSpecification
}

func expandPredefinedMetricPairSpecification(predefinedMetricPairSpecificationSlice []interface{}) *autoscaling.PredictiveScalingPredefinedMetricPair {
	if predefinedMetricPairSpecificationSlice == nil || len(predefinedMetricPairSpecificationSlice) < 1 {
		return nil
	}
	predefinedMetricPairSpecificationFlat := predefinedMetricPairSpecificationSlice[0].(map[string]interface{})
	predefinedMetricPairSpecification := &autoscaling.PredictiveScalingPredefinedMetricPair{
		PredefinedMetricType: aws.String(predefinedMetricPairSpecificationFlat["predefined_metric_type"].(string)),
		ResourceLabel:        aws.String(predefinedMetricPairSpecificationFlat["resource_label"].(string)),
	}
	return predefinedMetricPairSpecification
}

func expandPredefinedScalingMetricSpecification(predefinedScalingMetricSpecificationSlice []interface{}) *autoscaling.PredictiveScalingPredefinedScalingMetric {
	if predefinedScalingMetricSpecificationSlice == nil || len(predefinedScalingMetricSpecificationSlice) < 1 {
		return nil
	}
	predefinedScalingMetricSpecificationFlat := predefinedScalingMetricSpecificationSlice[0].(map[string]interface{})
	predefinedScalingMetricSpecification := &autoscaling.PredictiveScalingPredefinedScalingMetric{
		PredefinedMetricType: aws.String(predefinedScalingMetricSpecificationFlat["predefined_metric_type"].(string)),
		ResourceLabel:        aws.String(predefinedScalingMetricSpecificationFlat["resource_label"].(string)),
	}
	return predefinedScalingMetricSpecification
}

func expandCustomizedScalingMetricSpecification(customizedScalingMetricSpecificationSlice []interface{}) *autoscaling.PredictiveScalingCustomizedScalingMetric {
	if customizedScalingMetricSpecificationSlice == nil || len(customizedScalingMetricSpecificationSlice) < 1 {
		return nil
	}
	customizedScalingMetricSpecificationFlat := customizedScalingMetricSpecificationSlice[0].(map[string]interface{})
	customizedScalingMetricSpecification := &autoscaling.PredictiveScalingCustomizedScalingMetric{
		MetricDataQueries: expandMetricDataQueries(customizedScalingMetricSpecificationFlat["metric_data_queries"].([]interface{})),
	}
	return customizedScalingMetricSpecification
}

func expandCustomizedLoadMetricSpecification(customizedLoadMetricSpecificationSlice []interface{}) *autoscaling.PredictiveScalingCustomizedLoadMetric {
	if customizedLoadMetricSpecificationSlice == nil || len(customizedLoadMetricSpecificationSlice) < 1 {
		return nil
	}
	customizedLoadMetricSpecificationSliceFlat := customizedLoadMetricSpecificationSlice[0].(map[string]interface{})
	customizedLoadMetricSpecification := &autoscaling.PredictiveScalingCustomizedLoadMetric{
		MetricDataQueries: expandMetricDataQueries(customizedLoadMetricSpecificationSliceFlat["metric_data_queries"].([]interface{})),
	}
	return customizedLoadMetricSpecification
}

func expandCustomizedCapacityMetricSpecification(customizedCapacityMetricSlice []interface{}) *autoscaling.PredictiveScalingCustomizedCapacityMetric {
	if customizedCapacityMetricSlice == nil || len(customizedCapacityMetricSlice) < 1 {
		return nil
	}
	customizedCapacityMetricSliceFlat := customizedCapacityMetricSlice[0].(map[string]interface{})
	customizedCapacityMetricSpecification := &autoscaling.PredictiveScalingCustomizedCapacityMetric{
		MetricDataQueries: expandMetricDataQueries(customizedCapacityMetricSliceFlat["metric_data_queries"].([]interface{})),
	}
	return customizedCapacityMetricSpecification
}

func expandMetricDataQueries(metricDataQuerySlices []interface{}) []*autoscaling.MetricDataQuery {
	if metricDataQuerySlices == nil || len(metricDataQuerySlices) < 1 {
		return nil
	}
	metricDataQueries := make([]*autoscaling.MetricDataQuery, len(metricDataQuerySlices))

	for i := range metricDataQueries {

		metricDataQueryFlat := metricDataQuerySlices[i].(map[string]interface{})
		metricDataQuery := &autoscaling.MetricDataQuery{
			Id: aws.String(metricDataQueryFlat["id"].(string)),
		}
		if val, ok := metricDataQueryFlat["metric_stat"]; ok && len(val.([]interface{})) > 0 {
			metricStatSpec := val.([]interface{})[0].(map[string]interface{})
			metricSpec := metricStatSpec["metric"].([]interface{})[0].(map[string]interface{})
			metric := &autoscaling.Metric{
				MetricName: aws.String(metricSpec["metric_name"].(string)),
				Namespace:  aws.String(metricSpec["namespace"].(string)),
			}
			if v, ok := metricSpec["dimensions"]; ok {
				dims := v.(*schema.Set).List()
				dimList := make([]*autoscaling.MetricDimension, len(dims))
				for i := range dimList {
					dim := dims[i].(map[string]interface{})
					md := &autoscaling.MetricDimension{
						Name:  aws.String(dim["name"].(string)),
						Value: aws.String(dim["value"].(string)),
					}
					dimList[i] = md
				}
				metric.Dimensions = dimList
			}
			metricStat := &autoscaling.MetricStat{
				Metric: metric,
				Stat:   aws.String(metricStatSpec["stat"].(string)),
			}
			if v, ok := metricStatSpec["unit"]; ok && len(v.(string)) > 0 {
				metricStat.Unit = aws.String(v.(string))
			}
			metricDataQuery.MetricStat = metricStat
		}
		if val, ok := metricDataQueryFlat["expression"]; ok && val.(string) != "" {
			metricDataQuery.Expression = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["label"]; ok && val.(string) != "" {
			metricDataQuery.Label = aws.String(val.(string))
		}
		if val, ok := metricDataQueryFlat["return_data"]; ok {
			metricDataQuery.ReturnData = aws.Bool(val.(bool))
		}
		metricDataQueries[i] = metricDataQuery
	}
	return metricDataQueries
}

func flattenTargetTrackingConfiguration(config *autoscaling.TargetTrackingConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{}
	result["disable_scale_in"] = aws.BoolValue(config.DisableScaleIn)
	result["target_value"] = aws.Float64Value(config.TargetValue)
	if config.PredefinedMetricSpecification != nil {
		spec := map[string]interface{}{}
		spec["predefined_metric_type"] = aws.StringValue(config.PredefinedMetricSpecification.PredefinedMetricType)
		if config.PredefinedMetricSpecification.ResourceLabel != nil {
			spec["resource_label"] = aws.StringValue(config.PredefinedMetricSpecification.ResourceLabel)
		}
		result["predefined_metric_specification"] = []map[string]interface{}{spec}
	}
	if config.CustomizedMetricSpecification != nil {
		spec := map[string]interface{}{}
		spec["metric_name"] = aws.StringValue(config.CustomizedMetricSpecification.MetricName)
		spec["namespace"] = aws.StringValue(config.CustomizedMetricSpecification.Namespace)
		spec["statistic"] = aws.StringValue(config.CustomizedMetricSpecification.Statistic)
		if config.CustomizedMetricSpecification.Unit != nil {
			spec["unit"] = aws.StringValue(config.CustomizedMetricSpecification.Unit)
		}
		if config.CustomizedMetricSpecification.Dimensions != nil {
			dimSpec := make([]interface{}, len(config.CustomizedMetricSpecification.Dimensions))
			for i := range dimSpec {
				dim := map[string]interface{}{}
				rawDim := config.CustomizedMetricSpecification.Dimensions[i]
				dim["name"] = aws.StringValue(rawDim.Name)
				dim["value"] = aws.StringValue(rawDim.Value)
				dimSpec[i] = dim
			}
			spec["metric_dimension"] = dimSpec
		}
		result["customized_metric_specification"] = []map[string]interface{}{spec}
	}
	return []interface{}{result}
}

func flattenPredictiveScalingConfig(predictiveScalingConfig *autoscaling.PredictiveScalingConfiguration) []map[string]interface{} {
	predictiveScalingConfigFlat := map[string]interface{}{}
	if predictiveScalingConfig == nil {
		return nil
	}
	if predictiveScalingConfig.MetricSpecifications != nil && len(predictiveScalingConfig.MetricSpecifications) > 0 {
		predictiveScalingConfigFlat["metric_specification"] = flattenPredictiveScalingMetricSpecifications(predictiveScalingConfig.MetricSpecifications)
	}
	if predictiveScalingConfig.Mode != nil {
		predictiveScalingConfigFlat["mode"] = aws.StringValue(predictiveScalingConfig.Mode)
	}
	if predictiveScalingConfig.SchedulingBufferTime != nil {
		predictiveScalingConfigFlat["scheduling_buffer_time"] = strconv.FormatInt(aws.Int64Value(predictiveScalingConfig.SchedulingBufferTime), 10)
	}
	if predictiveScalingConfig.MaxCapacityBreachBehavior != nil {
		predictiveScalingConfigFlat["max_capacity_breach_behavior"] = aws.StringValue(predictiveScalingConfig.MaxCapacityBreachBehavior)
	}
	if predictiveScalingConfig.MaxCapacityBuffer != nil {
		predictiveScalingConfigFlat["max_capacity_buffer"] = strconv.FormatInt(aws.Int64Value(predictiveScalingConfig.MaxCapacityBuffer), 10)
	}
	return []map[string]interface{}{predictiveScalingConfigFlat}
}

func flattenPredictiveScalingMetricSpecifications(metricSpecification []*autoscaling.PredictiveScalingMetricSpecification) []map[string]interface{} {
	metricSpecificationFlat := map[string]interface{}{}
	if metricSpecification == nil || len(metricSpecification) < 1 {
		return []map[string]interface{}{metricSpecificationFlat}
	}
	if metricSpecification[0].TargetValue != nil {
		metricSpecificationFlat["target_value"] = aws.Float64Value(metricSpecification[0].TargetValue)
	}
	if metricSpecification[0].CustomizedCapacityMetricSpecification != nil {
		metricSpecificationFlat["customized_capacity_metric_specification"] = flattenCustomizedCapacityMetricSpecification(metricSpecification[0].CustomizedCapacityMetricSpecification)
	}
	if metricSpecification[0].CustomizedLoadMetricSpecification != nil {
		metricSpecificationFlat["customized_load_metric_specification"] = flattenCustomizedLoadMetricSpecification(metricSpecification[0].CustomizedLoadMetricSpecification)
	}
	if metricSpecification[0].CustomizedScalingMetricSpecification != nil {
		metricSpecificationFlat["customized_scaling_metric_specification"] = flattenCustomizedScalingMetricSpecification(metricSpecification[0].CustomizedScalingMetricSpecification)
	}
	if metricSpecification[0].PredefinedLoadMetricSpecification != nil {
		metricSpecificationFlat["predefined_load_metric_specification"] = flattenPredefinedLoadMetricSpecification(metricSpecification[0].PredefinedLoadMetricSpecification)
	}
	if metricSpecification[0].PredefinedMetricPairSpecification != nil {
		metricSpecificationFlat["predefined_metric_pair_specification"] = flattenPredefinedMetricPairSpecification(metricSpecification[0].PredefinedMetricPairSpecification)
	}
	if metricSpecification[0].PredefinedScalingMetricSpecification != nil {
		metricSpecificationFlat["predefined_scaling_metric_specification"] = flattenPredefinedScalingMetricSpecification(metricSpecification[0].PredefinedScalingMetricSpecification)
	}
	return []map[string]interface{}{metricSpecificationFlat}
}

func flattenPredefinedScalingMetricSpecification(predefinedScalingMetricSpecification *autoscaling.PredictiveScalingPredefinedScalingMetric) []map[string]interface{} {
	predefinedScalingMetricSpecificationFlat := map[string]interface{}{}
	if predefinedScalingMetricSpecification == nil {
		return []map[string]interface{}{predefinedScalingMetricSpecificationFlat}
	}
	predefinedScalingMetricSpecificationFlat["predefined_metric_type"] = aws.StringValue(predefinedScalingMetricSpecification.PredefinedMetricType)
	predefinedScalingMetricSpecificationFlat["resource_label"] = aws.StringValue(predefinedScalingMetricSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedScalingMetricSpecificationFlat}
}

func flattenPredefinedLoadMetricSpecification(predefinedLoadMetricSpecification *autoscaling.PredictiveScalingPredefinedLoadMetric) []map[string]interface{} {
	predefinedLoadMetricSpecificationFlat := map[string]interface{}{}
	if predefinedLoadMetricSpecification == nil {
		return []map[string]interface{}{predefinedLoadMetricSpecificationFlat}
	}
	predefinedLoadMetricSpecificationFlat["predefined_metric_type"] = aws.StringValue(predefinedLoadMetricSpecification.PredefinedMetricType)
	predefinedLoadMetricSpecificationFlat["resource_label"] = aws.StringValue(predefinedLoadMetricSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedLoadMetricSpecificationFlat}
}

func flattenPredefinedMetricPairSpecification(predefinedMetricPairSpecification *autoscaling.PredictiveScalingPredefinedMetricPair) []map[string]interface{} {
	predefinedMetricPairSpecificationFlat := map[string]interface{}{}
	if predefinedMetricPairSpecification == nil {
		return []map[string]interface{}{predefinedMetricPairSpecificationFlat}
	}
	predefinedMetricPairSpecificationFlat["predefined_metric_type"] = aws.StringValue(predefinedMetricPairSpecification.PredefinedMetricType)
	predefinedMetricPairSpecificationFlat["resource_label"] = aws.StringValue(predefinedMetricPairSpecification.ResourceLabel)
	return []map[string]interface{}{predefinedMetricPairSpecificationFlat}
}

func flattenCustomizedScalingMetricSpecification(customizedScalingMetricSpecification *autoscaling.PredictiveScalingCustomizedScalingMetric) []map[string]interface{} {
	customizedScalingMetricSpecificationFlat := map[string]interface{}{}
	if customizedScalingMetricSpecification == nil {
		return []map[string]interface{}{customizedScalingMetricSpecificationFlat}
	}
	customizedScalingMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedScalingMetricSpecification.MetricDataQueries)
	return []map[string]interface{}{customizedScalingMetricSpecificationFlat}
}

func flattenCustomizedLoadMetricSpecification(customizedLoadMetricSpecification *autoscaling.PredictiveScalingCustomizedLoadMetric) []map[string]interface{} {
	customizedLoadMetricSpecificationFlat := map[string]interface{}{}
	if customizedLoadMetricSpecification == nil {
		return []map[string]interface{}{customizedLoadMetricSpecificationFlat}
	}
	customizedLoadMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedLoadMetricSpecification.MetricDataQueries)
	return []map[string]interface{}{customizedLoadMetricSpecificationFlat}
}

func flattenCustomizedCapacityMetricSpecification(customizedCapacityMetricSpecification *autoscaling.PredictiveScalingCustomizedCapacityMetric) []map[string]interface{} {
	customizedCapacityMetricSpecificationFlat := map[string]interface{}{}
	if customizedCapacityMetricSpecification == nil {
		return []map[string]interface{}{customizedCapacityMetricSpecificationFlat}
	}
	customizedCapacityMetricSpecificationFlat["metric_data_queries"] = flattenMetricDataQueries(customizedCapacityMetricSpecification.MetricDataQueries)

	return []map[string]interface{}{customizedCapacityMetricSpecificationFlat}
}

func flattenMetricDataQueries(metricDataQueries []*autoscaling.MetricDataQuery) []interface{} {
	metricDataQueriesSpec := make([]interface{}, len(metricDataQueries))
	for i := range metricDataQueriesSpec {
		metricDataQuery := map[string]interface{}{}
		rawMetricDataQuery := metricDataQueries[i]
		metricDataQuery["id"] = aws.StringValue(rawMetricDataQuery.Id)
		if rawMetricDataQuery.Expression != nil {
			metricDataQuery["expression"] = aws.StringValue(rawMetricDataQuery.Expression)
		}
		if rawMetricDataQuery.Label != nil {
			metricDataQuery["label"] = aws.StringValue(rawMetricDataQuery.Label)
		}
		if rawMetricDataQuery.MetricStat != nil {
			metricStatSpec := map[string]interface{}{}
			rawMetricStat := rawMetricDataQuery.MetricStat
			rawMetric := rawMetricStat.Metric
			metricSpec := map[string]interface{}{}
			if rawMetric.Dimensions != nil {
				dimSpec := make([]interface{}, len(rawMetric.Dimensions))
				for i := range dimSpec {
					dim := map[string]interface{}{}
					rawDim := rawMetric.Dimensions[i]
					dim["name"] = aws.StringValue(rawDim.Name)
					dim["value"] = aws.StringValue(rawDim.Value)
					dimSpec[i] = dim
				}
				metricSpec["dimensions"] = dimSpec
			}
			metricSpec["metric_name"] = aws.StringValue(rawMetric.MetricName)
			metricSpec["namespace"] = aws.StringValue(rawMetric.Namespace)
			metricStatSpec["metric"] = []map[string]interface{}{metricSpec}
			metricStatSpec["stat"] = aws.StringValue(rawMetricStat.Stat)
			if rawMetricStat.Unit != nil {
				metricStatSpec["unit"] = aws.StringValue(rawMetricStat.Unit)
			}
			metricDataQuery["metric_stat"] = []map[string]interface{}{metricStatSpec}
		}
		if rawMetricDataQuery.ReturnData != nil {
			metricDataQuery["return_data"] = aws.BoolValue(rawMetricDataQuery.ReturnData)
		}
		metricDataQueriesSpec[i] = metricDataQuery
	}
	return metricDataQueriesSpec
}
