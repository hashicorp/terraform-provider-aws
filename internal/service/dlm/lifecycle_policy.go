package dlm

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceLifecyclePolicyCreate,
		Read:   resourceLifecyclePolicyRead,
		Update: resourceLifecyclePolicyUpdate,
		Delete: resourceLifecyclePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile("^[0-9A-Za-z _-]+$"), "see https://docs.aws.amazon.com/cli/latest/reference/dlm/create-lifecycle-policy.html"),
					validation.StringLenBetween(1, 500),
				),
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy_details": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cross_region_copy": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 3,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_configuration": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"cmk_arn": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: verify.ValidARN,
															},
															"encrypted": {
																Type:     schema.TypeBool,
																Optional: true,
																Default:  false,
															},
														},
													},
												},
												"retain_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"interval": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice(
																	dlm.RetentionIntervalUnitValues_Values(),
																	false,
																),
															},
														},
													},
												},
												"target": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w:\-\/\*]+$`), ""),
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 120),
											validation.StringMatch(regexp.MustCompile("^[0-9A-Za-z _-]+$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_Action.html"),
										),
									},
								},
							},
						},
						"event_source": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"parameters": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description_regex": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 1000),
												},
												"event_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(dlm.EventTypeValues_Values(), false),
												},
												"snapshot_owner": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 50,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidAccountID,
													},
												},
											},
										},
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(dlm.EventSourceValues_Values(), false),
									},
								},
							},
						},
						"resource_types": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(dlm.ResourceTypeValues_Values(), false),
							},
						},
						"resource_locations": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(dlm.ResourceLocationValues_Values(), false),
							},
						},
						"parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"exclude_boot_volume": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"no_reboot": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"policy_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      dlm.PolicyTypeValuesEbsSnapshotManagement,
							ValidateFunc: validation.StringInSlice(dlm.PolicyTypeValues_Values(), false),
						},
						"schedule": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"copy_tags": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"create_rule": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cron_expression": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile("^cron\\([^\n]{11,100}\\)$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_CreateRule.html"),
												},
												"interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntInSlice([]int{1, 2, 3, 4, 6, 8, 12, 24}),
												},
												"interval_unit": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(dlm.IntervalUnitValues_Values(), false),
												},
												"location": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(dlm.LocationValues_Values(), false),
												},
												"times": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringMatch(regexp.MustCompile("^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$"), "see https://docs.aws.amazon.com/dlm/latest/APIReference/API_CreateRule.html#dlm-Type-CreateRule-Times"),
													},
												},
											},
										},
									},
									"cross_region_copy_rule": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 3,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cmk_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"copy_tags": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"deprecate_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"interval": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice(
																	dlm.RetentionIntervalUnitValues_Values(),
																	false,
																),
															},
														},
													},
												},
												"encrypted": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"retain_rule": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"interval": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"interval_unit": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice(
																	dlm.RetentionIntervalUnitValues_Values(),
																	false,
																),
															},
														},
													},
												},
												"target": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w:\-\/\*]+$`), ""),
												},
											},
										},
									},
									"deprecate_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												"interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice(
														dlm.RetentionIntervalUnitValues_Values(),
														false,
													),
												},
											},
										},
									},
									"fast_restore_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"availability_zones": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												"interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice(
														dlm.RetentionIntervalUnitValues_Values(),
														false,
													),
												},
											},
										},
									},
									"name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 120),
									},
									"retain_rule": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 1000),
												},
												"interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"interval_unit": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice(
														dlm.RetentionIntervalUnitValues_Values(),
														false,
													),
												},
											},
										},
									},
									"share_rule": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_accounts": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: verify.ValidAccountID,
													},
												},
												"unshare_interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"unshare_interval_unit": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice(
														dlm.RetentionIntervalUnitValues_Values(),
														false,
													),
												},
											},
										},
									},
									"tags_to_add": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"variable_tags": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"target_tags": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      dlm.SettablePolicyStateValuesEnabled,
				ValidateFunc: validation.StringInSlice(dlm.SettablePolicyStateValues_Values(), false),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLifecyclePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DLMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := dlm.CreateLifecyclePolicyInput{
		Description:      aws.String(d.Get("description").(string)),
		ExecutionRoleArn: aws.String(d.Get("execution_role_arn").(string)),
		PolicyDetails:    expandPolicyDetails(d.Get("policy_details").([]interface{})),
		State:            aws.String(d.Get("state").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[INFO] Creating DLM lifecycle policy: %s", input)
	out, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.CreateLifecyclePolicy(&input)
	}, dlm.ErrCodeInvalidRequestException)

	if err != nil {
		return fmt.Errorf("error creating DLM Lifecycle Policy: %s", err)
	}

	d.SetId(aws.StringValue(out.(*dlm.CreateLifecyclePolicyOutput).PolicyId))

	return resourceLifecyclePolicyRead(d, meta)
}

func resourceLifecyclePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DLMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading DLM lifecycle policy: %s", d.Id())
	out, err := conn.GetLifecyclePolicy(&dlm.GetLifecyclePolicyInput{
		PolicyId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, dlm.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] DLM Lifecycle Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DLM Lifecycle Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.Policy.PolicyArn)
	d.Set("description", out.Policy.Description)
	d.Set("execution_role_arn", out.Policy.ExecutionRoleArn)
	d.Set("state", out.Policy.State)
	if err := d.Set("policy_details", flattenPolicyDetails(out.Policy.PolicyDetails)); err != nil {
		return fmt.Errorf("error setting policy details %s", err)
	}

	tags := KeyValueTags(out.Policy.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLifecyclePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DLMConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := dlm.UpdateLifecyclePolicyInput{
			PolicyId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}
		if d.HasChange("execution_role_arn") {
			input.ExecutionRoleArn = aws.String(d.Get("execution_role_arn").(string))
		}
		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}
		if d.HasChange("policy_details") {
			input.PolicyDetails = expandPolicyDetails(d.Get("policy_details").([]interface{}))
		}

		log.Printf("[INFO] Updating lifecycle policy %s", d.Id())
		_, err := conn.UpdateLifecyclePolicy(&input)
		if err != nil {
			return fmt.Errorf("error updating DLM Lifecycle Policy (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceLifecyclePolicyRead(d, meta)
}

func resourceLifecyclePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DLMConn

	log.Printf("[INFO] Deleting DLM lifecycle policy: %s", d.Id())
	_, err := conn.DeleteLifecyclePolicy(&dlm.DeleteLifecyclePolicyInput{
		PolicyId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dlm.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting DLM Lifecycle Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func expandPolicyDetails(cfg []interface{}) *dlm.PolicyDetails {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]interface{})
	policyType := m["policy_type"].(string)

	policyDetails := &dlm.PolicyDetails{
		PolicyType: aws.String(policyType),
	}
	if v, ok := m["resource_types"].([]interface{}); ok && len(v) > 0 {
		policyDetails.ResourceTypes = flex.ExpandStringList(v)
	}
	if v, ok := m["resource_locations"].([]interface{}); ok && len(v) > 0 {
		policyDetails.ResourceLocations = flex.ExpandStringList(v)
	}
	if v, ok := m["schedule"].([]interface{}); ok && len(v) > 0 {
		policyDetails.Schedules = expandSchedules(v)
	}
	if v, ok := m["action"].([]interface{}); ok && len(v) > 0 {
		policyDetails.Actions = expandActions(v)
	}
	if v, ok := m["event_source"].([]interface{}); ok && len(v) > 0 {
		policyDetails.EventSource = expandEventSource(v)
	}
	if v, ok := m["target_tags"].(map[string]interface{}); ok && len(v) > 0 {
		policyDetails.TargetTags = expandTags(v)
	}
	if v, ok := m["parameters"].([]interface{}); ok && len(v) > 0 {
		policyDetails.Parameters = expandParameters(v, policyType)
	}

	return policyDetails
}

func flattenPolicyDetails(policyDetails *dlm.PolicyDetails) []map[string]interface{} {
	result := make(map[string]interface{})
	result["resource_types"] = flex.FlattenStringList(policyDetails.ResourceTypes)
	result["resource_locations"] = flex.FlattenStringList(policyDetails.ResourceLocations)
	result["action"] = flattenActions(policyDetails.Actions)
	result["event_source"] = flattenEventSource(policyDetails.EventSource)
	result["schedule"] = flattenSchedules(policyDetails.Schedules)
	result["target_tags"] = flattenTags(policyDetails.TargetTags)
	result["policy_type"] = aws.StringValue(policyDetails.PolicyType)

	if policyDetails.Parameters != nil {
		result["parameters"] = flattenParameters(policyDetails.Parameters)
	}

	return []map[string]interface{}{result}
}

func expandSchedules(cfg []interface{}) []*dlm.Schedule {
	schedules := make([]*dlm.Schedule, len(cfg))
	for i, c := range cfg {
		schedule := &dlm.Schedule{}
		m := c.(map[string]interface{})
		if v, ok := m["copy_tags"]; ok {
			schedule.CopyTags = aws.Bool(v.(bool))
		}
		if v, ok := m["create_rule"]; ok {
			schedule.CreateRule = expandCreateRule(v.([]interface{}))
		}
		if v, ok := m["cross_region_copy_rule"].(*schema.Set); ok && v.Len() > 0 {
			schedule.CrossRegionCopyRules = expandCrossRegionCopyRules(v.List())
		}
		if v, ok := m["name"]; ok {
			schedule.Name = aws.String(v.(string))
		}
		if v, ok := m["deprecate_rule"]; ok {
			schedule.DeprecateRule = expandDeprecateRule(v.([]interface{}))
		}
		if v, ok := m["fast_restore_rule"]; ok {
			schedule.FastRestoreRule = expandFastRestoreRule(v.([]interface{}))
		}
		if v, ok := m["share_rule"]; ok {
			schedule.ShareRules = expandShareRule(v.([]interface{}))
		}
		if v, ok := m["retain_rule"]; ok {
			schedule.RetainRule = expandRetainRule(v.([]interface{}))
		}
		if v, ok := m["tags_to_add"]; ok {
			schedule.TagsToAdd = expandTags(v.(map[string]interface{}))
		}
		if v, ok := m["variable_tags"]; ok {
			schedule.VariableTags = expandTags(v.(map[string]interface{}))
		}

		schedules[i] = schedule
	}

	return schedules
}

func flattenSchedules(schedules []*dlm.Schedule) []map[string]interface{} {
	result := make([]map[string]interface{}, len(schedules))
	for i, s := range schedules {
		m := make(map[string]interface{})
		m["copy_tags"] = aws.BoolValue(s.CopyTags)
		m["create_rule"] = flattenCreateRule(s.CreateRule)
		m["cross_region_copy_rule"] = flattenCrossRegionCopyRules(s.CrossRegionCopyRules)
		m["name"] = aws.StringValue(s.Name)
		m["retain_rule"] = flattenRetainRule(s.RetainRule)
		m["tags_to_add"] = flattenTags(s.TagsToAdd)
		m["variable_tags"] = flattenTags(s.VariableTags)

		if s.DeprecateRule != nil {
			m["deprecate_rule"] = flattenDeprecateRule(s.DeprecateRule)
		}

		if s.FastRestoreRule != nil {
			m["fast_restore_rule"] = flattenFastRestoreRule(s.FastRestoreRule)
		}

		if s.ShareRules != nil {
			m["share_rule"] = flattenShareRule(s.ShareRules)
		}

		result[i] = m
	}

	return result
}

func expandActions(cfg []interface{}) []*dlm.Action {
	actions := make([]*dlm.Action, len(cfg))
	for i, c := range cfg {
		action := &dlm.Action{}
		m := c.(map[string]interface{})
		if v, ok := m["cross_region_copy"].(*schema.Set); ok {
			action.CrossRegionCopy = expandActionCrossRegionCopyRules(v.List())
		}
		if v, ok := m["name"]; ok {
			action.Name = aws.String(v.(string))
		}

		actions[i] = action
	}

	return actions
}

func flattenActions(actions []*dlm.Action) []map[string]interface{} {
	result := make([]map[string]interface{}, len(actions))
	for i, s := range actions {
		m := make(map[string]interface{})

		m["name"] = aws.StringValue(s.Name)

		if s.CrossRegionCopy != nil {
			m["cross_region_copy"] = flattenActionCrossRegionCopyRules(s.CrossRegionCopy)
		}

		result[i] = m
	}

	return result
}

func expandActionCrossRegionCopyRules(l []interface{}) []*dlm.CrossRegionCopyAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var rules []*dlm.CrossRegionCopyAction

	for _, tfMapRaw := range l {
		m, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		rule := &dlm.CrossRegionCopyAction{}
		if v, ok := m["encryption_configuration"].([]interface{}); ok {
			rule.EncryptionConfiguration = expandActionCrossRegionCopyRuleEncryptionConfiguration(v)
		}
		if v, ok := m["retain_rule"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := m["target"].(string); ok && v != "" {
			rule.Target = aws.String(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenActionCrossRegionCopyRules(rules []*dlm.CrossRegionCopyAction) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var result []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := map[string]interface{}{
			"encryption_configuration": flattenActionCrossRegionCopyRuleEncryptionConfiguration(rule.EncryptionConfiguration),
			"retain_rule":              flattenCrossRegionCopyRuleRetainRule(rule.RetainRule),
			"target":                   aws.StringValue(rule.Target),
		}

		result = append(result, m)
	}

	return result
}

func expandActionCrossRegionCopyRuleEncryptionConfiguration(l []interface{}) *dlm.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := &dlm.EncryptionConfiguration{
		Encrypted: aws.Bool(m["encrypted"].(bool)),
	}

	if v, ok := m["cmk_arn"].(string); ok && v != "" {
		config.CmkArn = aws.String(v)
	}
	return config
}

func flattenActionCrossRegionCopyRuleEncryptionConfiguration(rule *dlm.EncryptionConfiguration) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encrypted": aws.BoolValue(rule.Encrypted),
		"cmk_arn":   aws.StringValue(rule.CmkArn),
	}

	return []interface{}{m}
}

func expandEventSource(l []interface{}) *dlm.EventSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := &dlm.EventSource{
		Type: aws.String(m["type"].(string)),
	}

	if v, ok := m["parameters"].([]interface{}); ok && len(v) > 0 {
		config.Parameters = expandEventSourceParameters(v)
	}

	return config
}

func flattenEventSource(rule *dlm.EventSource) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"parameters": flattenEventSourceParameters(rule.Parameters),
		"type":       aws.StringValue(rule.Type),
	}

	return []interface{}{m}
}

func expandEventSourceParameters(l []interface{}) *dlm.EventParameters {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := &dlm.EventParameters{
		DescriptionRegex: aws.String(m["description_regex"].(string)),
		EventType:        aws.String(m["event_type"].(string)),
		SnapshotOwner:    flex.ExpandStringSet(m["snapshot_owner"].(*schema.Set)),
	}

	return config
}

func flattenEventSourceParameters(rule *dlm.EventParameters) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"description_regex": aws.StringValue(rule.DescriptionRegex),
		"event_type":        aws.StringValue(rule.EventType),
		"snapshot_owner":    flex.FlattenStringSet(rule.SnapshotOwner),
	}

	return []interface{}{m}
}

func expandCrossRegionCopyRules(l []interface{}) []*dlm.CrossRegionCopyRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var rules []*dlm.CrossRegionCopyRule

	for _, tfMapRaw := range l {
		m, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		rule := &dlm.CrossRegionCopyRule{}

		if v, ok := m["cmk_arn"].(string); ok && v != "" {
			rule.CmkArn = aws.String(v)
		}
		if v, ok := m["copy_tags"].(bool); ok {
			rule.CopyTags = aws.Bool(v)
		}
		if v, ok := m["deprecate_rule"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.DeprecateRule = expandCrossRegionCopyRuleDeprecateRule(v)
		}
		if v, ok := m["encrypted"].(bool); ok {
			rule.Encrypted = aws.Bool(v)
		}
		if v, ok := m["retain_rule"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.RetainRule = expandCrossRegionCopyRuleRetainRule(v)
		}
		if v, ok := m["target"].(string); ok && v != "" {
			rule.Target = aws.String(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenCrossRegionCopyRules(rules []*dlm.CrossRegionCopyRule) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var result []interface{}

	for _, rule := range rules {
		if rule == nil {
			continue
		}

		m := map[string]interface{}{
			"cmk_arn":        aws.StringValue(rule.CmkArn),
			"copy_tags":      aws.BoolValue(rule.CopyTags),
			"deprecate_rule": flattenCrossRegionCopyRuleDeprecateRule(rule.DeprecateRule),
			"encrypted":      aws.BoolValue(rule.Encrypted),
			"retain_rule":    flattenCrossRegionCopyRuleRetainRule(rule.RetainRule),
			"target":         aws.StringValue(rule.Target),
		}

		result = append(result, m)
	}

	return result
}

func expandCrossRegionCopyRuleDeprecateRule(l []interface{}) *dlm.CrossRegionCopyDeprecateRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &dlm.CrossRegionCopyDeprecateRule{
		Interval:     aws.Int64(int64(m["interval"].(int))),
		IntervalUnit: aws.String(m["interval_unit"].(string)),
	}
}

func expandCrossRegionCopyRuleRetainRule(l []interface{}) *dlm.CrossRegionCopyRetainRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &dlm.CrossRegionCopyRetainRule{
		Interval:     aws.Int64(int64(m["interval"].(int))),
		IntervalUnit: aws.String(m["interval_unit"].(string)),
	}
}

func flattenCrossRegionCopyRuleDeprecateRule(rule *dlm.CrossRegionCopyDeprecateRule) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"interval":      int(aws.Int64Value(rule.Interval)),
		"interval_unit": aws.StringValue(rule.IntervalUnit),
	}

	return []interface{}{m}
}

func flattenCrossRegionCopyRuleRetainRule(rule *dlm.CrossRegionCopyRetainRule) []interface{} {
	if rule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"interval":      int(aws.Int64Value(rule.Interval)),
		"interval_unit": aws.StringValue(rule.IntervalUnit),
	}

	return []interface{}{m}
}

func expandCreateRule(cfg []interface{}) *dlm.CreateRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	c := cfg[0].(map[string]interface{})
	createRule := &dlm.CreateRule{}

	if v, ok := c["times"].([]interface{}); ok && len(v) > 0 {
		createRule.Times = flex.ExpandStringList(v)
	}

	if v, ok := c["interval"].(int); ok && v > 0 {
		createRule.Interval = aws.Int64(int64(v))
	}

	if v, ok := c["location"].(string); ok && v != "" {
		createRule.Location = aws.String(v)
	}

	if v, ok := c["interval_unit"].(string); ok && v != "" {
		createRule.IntervalUnit = aws.String(v)
	} else {
		createRule.IntervalUnit = aws.String(dlm.IntervalUnitValuesHours)
	}

	if v, ok := c["cron_expression"].(string); ok && v != "" {
		createRule.CronExpression = aws.String(v)
		createRule.IntervalUnit = nil
	}

	return createRule
}

func flattenCreateRule(createRule *dlm.CreateRule) []map[string]interface{} {
	if createRule == nil {
		return []map[string]interface{}{}
	}

	result := make(map[string]interface{})
	result["times"] = flex.FlattenStringList(createRule.Times)

	if createRule.Interval != nil {
		result["interval"] = aws.Int64Value(createRule.Interval)
	}

	if createRule.IntervalUnit != nil {
		result["interval_unit"] = aws.StringValue(createRule.IntervalUnit)
	}

	if createRule.Location != nil {
		result["location"] = aws.StringValue(createRule.Location)
	}

	if createRule.CronExpression != nil {
		result["cron_expression"] = aws.StringValue(createRule.CronExpression)
	}

	return []map[string]interface{}{result}
}

func expandRetainRule(cfg []interface{}) *dlm.RetainRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]interface{})
	rule := &dlm.RetainRule{}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int64(int64(v))
	}

	if v, ok := m["interval"].(int); ok && v > 0 {
		rule.Interval = aws.Int64(int64(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = aws.String(v)
	}

	return rule
}

func flattenRetainRule(retainRule *dlm.RetainRule) []map[string]interface{} {
	result := make(map[string]interface{})
	result["count"] = aws.Int64Value(retainRule.Count)
	result["interval_unit"] = aws.StringValue(retainRule.IntervalUnit)
	result["interval"] = aws.Int64Value(retainRule.Interval)

	return []map[string]interface{}{result}
}

func expandDeprecateRule(cfg []interface{}) *dlm.DeprecateRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]interface{})
	rule := &dlm.DeprecateRule{}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int64(int64(v))
	}

	if v, ok := m["interval"].(int); ok && v > 0 {
		rule.Interval = aws.Int64(int64(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = aws.String(v)
	}

	return rule
}

func flattenDeprecateRule(rule *dlm.DeprecateRule) []map[string]interface{} {
	result := make(map[string]interface{})
	result["count"] = aws.Int64Value(rule.Count)
	result["interval_unit"] = aws.StringValue(rule.IntervalUnit)
	result["interval"] = aws.Int64Value(rule.Interval)

	return []map[string]interface{}{result}
}

func expandFastRestoreRule(cfg []interface{}) *dlm.FastRestoreRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]interface{})
	rule := &dlm.FastRestoreRule{
		AvailabilityZones: flex.ExpandStringSet(m["availability_zones"].(*schema.Set)),
	}

	if v, ok := m["count"].(int); ok && v > 0 {
		rule.Count = aws.Int64(int64(v))
	}

	if v, ok := m["interval"].(int); ok && v > 0 {
		rule.Interval = aws.Int64(int64(v))
	}

	if v, ok := m["interval_unit"].(string); ok && v != "" {
		rule.IntervalUnit = aws.String(v)
	}

	return rule
}

func flattenFastRestoreRule(rule *dlm.FastRestoreRule) []map[string]interface{} {
	result := make(map[string]interface{})
	result["count"] = aws.Int64Value(rule.Count)
	result["interval_unit"] = aws.StringValue(rule.IntervalUnit)
	result["interval"] = aws.Int64Value(rule.Interval)
	result["availability_zones"] = flex.FlattenStringSet(rule.AvailabilityZones)

	return []map[string]interface{}{result}
}

func expandShareRule(cfg []interface{}) []*dlm.ShareRule {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	rules := make([]*dlm.ShareRule, 0)

	for _, shareRule := range cfg {
		m := shareRule.(map[string]interface{})

		rule := &dlm.ShareRule{
			TargetAccounts: flex.ExpandStringSet(m["target_accounts"].(*schema.Set)),
		}

		if v, ok := m["unshare_interval"].(int); ok && v > 0 {
			rule.UnshareInterval = aws.Int64(int64(v))
		}

		if v, ok := m["unshare_interval_unit"].(string); ok && v != "" {
			rule.UnshareIntervalUnit = aws.String(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenShareRule(rules []*dlm.ShareRule) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, v := range rules {
		rule := make(map[string]interface{})

		if v == nil {
			return nil
		}

		if v.TargetAccounts != nil {
			rule["target_accounts"] = flex.FlattenStringSet(v.TargetAccounts)
		}

		if v.UnshareIntervalUnit != nil {
			rule["unshare_interval_unit"] = aws.StringValue(v.UnshareIntervalUnit)
		}

		if v.UnshareInterval != nil {
			rule["unshare_interval"] = aws.Int64Value(v.UnshareInterval)
		}

		values = append(values, rule)
	}

	return values
}

func expandTags(m map[string]interface{}) []*dlm.Tag {
	var result []*dlm.Tag
	for k, v := range m {
		result = append(result, &dlm.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

func flattenTags(tags []*dlm.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	return result
}

func expandParameters(cfg []interface{}, policyType string) *dlm.Parameters {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}
	m := cfg[0].(map[string]interface{})
	parameters := &dlm.Parameters{}

	if v, ok := m["exclude_boot_volume"].(bool); ok && policyType == dlm.PolicyTypeValuesEbsSnapshotManagement {
		parameters.ExcludeBootVolume = aws.Bool(v)
	}

	if v, ok := m["no_reboot"].(bool); ok && policyType == dlm.PolicyTypeValuesImageManagement {
		parameters.NoReboot = aws.Bool(v)
	}

	return parameters
}

func flattenParameters(parameters *dlm.Parameters) []map[string]interface{} {
	result := make(map[string]interface{})
	if parameters.ExcludeBootVolume != nil {
		result["exclude_boot_volume"] = aws.BoolValue(parameters.ExcludeBootVolume)
	}

	if parameters.NoReboot != nil {
		result["no_reboot"] = aws.BoolValue(parameters.NoReboot)
	}

	return []map[string]interface{}{result}
}
