// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_job_definition", name="Job Definition")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/batch;batch.JobDefinition", importIgnore="deregister_on_new_revision")
func ResourceJobDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobDefinitionCreate,
		ReadWithoutTimeout:   resourceJobDefinitionRead,
		UpdateWithoutTimeout: resourceJobDefinitionUpdate,
		DeleteWithoutTimeout: resourceJobDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"container_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"eks_properties", "node_properties"},
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := EquivalentContainerPropertiesJSON(old, new)

					return equal
				},
				ValidateFunc: validJobContainerProperties,
			},

			"deregister_on_new_revision": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},

			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},

			"node_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"container_properties", "eks_properties"},
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := EquivalentNodePropertiesJSON(old, new)
					return equal
				},
				ValidateFunc: validJobNodeProperties,
			},

			"eks_properties": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"container_properties", "node_properties"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pod_properties": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"containers": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"args": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"command": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"env": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrName: {
																Type:     schema.TypeString,
																Required: true,
															},
															names.AttrValue: {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"image": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_pull_policy": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ImagePullPolicy_Values(), false),
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrResources: {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"limits": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"requests": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
														},
													},
												},
												"security_context": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"privileged": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"read_only_root_file_system": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"run_as_group": {
																Type:     schema.TypeInt,
																Optional: true,
															},
															"run_as_non_root": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"run_as_user": {
																Type:     schema.TypeInt,
																Optional: true,
															},
														},
													},
												},
												"volume_mounts": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"mount_path": {
																Type:     schema.TypeString,
																Required: true,
															},
															names.AttrName: {
																Type:     schema.TypeString,
																Required: true,
															},
															"read_only": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
									"dns_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(DNSPolicy_Values(), false),
									},
									"host_network": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"image_pull_secret": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"metadata": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"labels": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"service_account_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"volumes": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"empty_dir": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"medium": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      "",
																ValidateFunc: validation.StringInSlice([]string{"", "Memory"}, true),
															},
															"size_limit": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"host_path": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrPath: {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "Default",
												},
												"secret": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"secret_name": {
																Type:     schema.TypeString,
																Required: true,
															},
															"optional": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"platform_capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(batch.PlatformCapability_Values(), false),
				},
			},

			names.AttrPropagateTags: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"retry_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
						"evaluate_on_exit": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 0,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAction: {
										Type:     schema.TypeString,
										Required: true,
										StateFunc: func(v interface{}) string {
											return strings.ToLower(v.(string))
										},
										ValidateFunc: validation.StringInSlice(batch.RetryAction_Values(), true),
									},
									"on_exit_code": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9]*\*?$`), "must contain only numbers, and can optionally end with an asterisk"),
										),
									},
									"on_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
									"on_status_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
								},
							},
						},
					},
				},
			},

			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"scheduling_priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),

			names.AttrTimeout: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempt_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},

			names.AttrType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.JobDefinitionTypeContainer, batch.JobDefinitionTypeMultinode}, true),
			},
		},

		CustomizeDiff: customdiff.Sequence(
			jobDefinitionCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func jobDefinitionCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if d.Id() != "" && needsJobDefUpdate(d) && d.Get(names.AttrARN).(string) != "" {
		d.SetNewComputed(names.AttrARN)
		d.SetNewComputed("revision")
		d.SetNewComputed(names.AttrID)
	}

	return nil
}

// needsJobDefUpdate determines if the Job Definition needs to be updated. This is the
// cost of not forcing new when updates to one argument (eg, container_properties)
// simultaneously impact a computed attribute (eg, arn). The real challenge here is that
// we have to figure out if a change is **GOING** to cause a new revision to be created,
// without the benefit of AWS just telling us. This is necessary because waiting until
// after AWS tells us is too late, and practitioners will need to refresh or worse, get
// an inconsistent plan. BUT, if we SetNewComputed **without** a change, we'll get a
// testing error: "the non-refresh plan was not empty".
func needsJobDefUpdate(d *schema.ResourceDiff) bool {
	if d.HasChange("container_properties") {
		o, n := d.GetChange("container_properties")

		equivalent, err := EquivalentContainerPropertiesJSON(o.(string), n.(string))
		if err != nil {
			return false
		}

		if !equivalent {
			return true
		}
	}

	if d.HasChange("node_properties") {
		o, n := d.GetChange("node_properties")

		equivalent, err := EquivalentNodePropertiesJSON(o.(string), n.(string))
		if err != nil {
			return false
		}

		if !equivalent {
			return true
		}
	}

	if d.HasChange("eks_properties") {
		o, n := d.GetChange("eks_properties")
		if len(o.([]interface{})) == 0 && len(n.([]interface{})) == 0 {
			return false
		}

		if d.Get(names.AttrType).(string) != batch.JobDefinitionTypeContainer {
			return false
		}

		var oeks, neks *batch.EksPodProperties
		if len(o.([]interface{})) > 0 {
			oProps := o.([]interface{})[0].(map[string]interface{})
			if opodProps, ok := oProps["pod_properties"].([]interface{}); ok && len(opodProps) > 0 {
				oeks = expandEKSPodProperties(opodProps[0].(map[string]interface{}))
			}
		}

		if len(n.([]interface{})) > 0 {
			nProps := n.([]interface{})[0].(map[string]interface{})
			if npodProps, ok := nProps["pod_properties"].([]interface{}); ok && len(npodProps) > 0 {
				neks = expandEKSPodProperties(npodProps[0].(map[string]interface{}))
			}
		}

		return !reflect.DeepEqual(oeks, neks)
	}

	if d.HasChange("retry_strategy") {
		o, n := d.GetChange("retry_strategy")
		if len(o.([]interface{})) == 0 && len(n.([]interface{})) == 0 {
			return false
		}

		var ors, nrs *batch.RetryStrategy
		if len(o.([]interface{})) > 0 {
			oProps := o.([]interface{})[0].(map[string]interface{})
			ors = expandRetryStrategy(oProps)
		}

		if len(n.([]interface{})) > 0 {
			nProps := n.([]interface{})[0].(map[string]interface{})
			nrs = expandRetryStrategy(nProps)
		}

		return !reflect.DeepEqual(ors, nrs)
	}

	if d.HasChange(names.AttrTimeout) {
		o, n := d.GetChange(names.AttrTimeout)
		if len(o.([]interface{})) == 0 && len(n.([]interface{})) == 0 {
			return false
		}

		var ors, nrs *batch.JobTimeout
		if len(o.([]interface{})) > 0 {
			oProps := o.([]interface{})[0].(map[string]interface{})
			ors = expandJobTimeout(oProps)
		}

		if len(n.([]interface{})) > 0 {
			nProps := n.([]interface{})[0].(map[string]interface{})
			nrs = expandJobTimeout(nProps)
		}

		return !reflect.DeepEqual(ors, nrs)
	}

	if d.HasChanges(
		names.AttrPropagateTags,
		names.AttrParameters,
		"platform_capabilities",
		"scheduling_priority",
		names.AttrType,
	) {
		return true
	}

	return false
}

func resourceJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	name := d.Get(names.AttrName).(string)
	jobDefinitionType := d.Get(names.AttrType).(string)
	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(name),
		PropagateTags:     aws.Bool(d.Get(names.AttrPropagateTags).(bool)),
		Tags:              getTagsIn(ctx),
		Type:              aws.String(jobDefinitionType),
	}

	if jobDefinitionType == batch.JobDefinitionTypeContainer {
		if v, ok := d.GetOk("node_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `node_properties` can be specified when `type` is %q", jobDefinitionType)
		}

		if v, ok := d.GetOk("container_properties"); ok {
			props, err := expandJobContainerProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Batch Job Definition (%s): %s", name, err)
			}

			if aws.StringValue(input.Type) == batch.JobDefinitionTypeContainer {
				removeEmptyEnvironmentVariables(&diags, props.Environment, cty.GetAttrPath("container_properties"))
				input.ContainerProperties = props
			}
		}

		if v, ok := d.GetOk("eks_properties"); ok && len(v.([]interface{})) > 0 {
			eksProps := v.([]interface{})[0].(map[string]interface{})
			if podProps, ok := eksProps["pod_properties"].([]interface{}); ok && len(podProps) > 0 {
				if aws.StringValue(input.Type) == batch.JobDefinitionTypeContainer {
					props := expandEKSPodProperties(podProps[0].(map[string]interface{}))
					input.EksProperties = &batch.EksProperties{
						PodProperties: props,
					}
				}
			}
		}
	}

	if jobDefinitionType == batch.JobDefinitionTypeMultinode {
		if v, ok := d.GetOk("container_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `container_properties` can be specified when `type` is %q", jobDefinitionType)
		}
		if v, ok := d.GetOk("eks_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `eks_properties` can be specified when `type` is %q", jobDefinitionType)
		}

		if v, ok := d.GetOk("node_properties"); ok {
			props, err := expandJobNodeProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Batch Job Definition (%s): %s", name, err)
			}

			for _, node := range props.NodeRangeProperties {
				removeEmptyEnvironmentVariables(&diags, node.Container.Environment, cty.GetAttrPath("node_properties"))
			}
			input.NodeProperties = props
		}
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetryStrategy = expandRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("scheduling_priority"); ok {
		input.SchedulingPriority = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Timeout = expandJobTimeout(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.RegisterJobDefinitionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Job Definition (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.JobDefinitionArn))

	return append(diags, resourceJobDefinitionRead(ctx, d, meta)...)
}

func resourceJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	jobDefinition, err := FindJobDefinitionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Definition (%s): %s", d.Id(), err)
	}

	arn, revision := aws.StringValue(jobDefinition.JobDefinitionArn), aws.Int64Value(jobDefinition.Revision)
	d.Set(names.AttrARN, arn)
	d.Set("arn_prefix", strings.TrimSuffix(arn, fmt.Sprintf(":%d", revision)))

	containerProperties, err := flattenContainerProperties(jobDefinition.ContainerProperties)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "converting Batch Container Properties to JSON: %s", err)
	}

	if err := d.Set("container_properties", containerProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting container_properties: %s", err)
	}

	if err := d.Set("eks_properties", flattenEKSProperties(jobDefinition.EksProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting eks_properties: %s", err)
	}

	d.Set(names.AttrName, jobDefinition.JobDefinitionName)

	nodeProperties, err := flattenNodeProperties(jobDefinition.NodeProperties)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "converting Batch Node Properties to JSON: %s", err)
	}

	if err := d.Set("node_properties", nodeProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_properties: %s", err)
	}

	d.Set(names.AttrParameters, aws.StringValueMap(jobDefinition.Parameters))
	d.Set("platform_capabilities", aws.StringValueSlice(jobDefinition.PlatformCapabilities))
	d.Set(names.AttrPropagateTags, jobDefinition.PropagateTags)

	if jobDefinition.RetryStrategy != nil {
		if err := d.Set("retry_strategy", []interface{}{flattenRetryStrategy(jobDefinition.RetryStrategy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting retry_strategy: %s", err)
		}
	} else {
		d.Set("retry_strategy", nil)
	}

	d.Set("revision", revision)
	d.Set("scheduling_priority", jobDefinition.SchedulingPriority)

	if jobDefinition.Timeout != nil {
		if err := d.Set(names.AttrTimeout, []interface{}{flattenJobTimeout(jobDefinition.Timeout)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting timeout: %s", err)
		}
	} else {
		d.Set(names.AttrTimeout, nil)
	}

	d.Set(names.AttrType, jobDefinition.Type)

	setTagsOut(ctx, jobDefinition.Tags)

	return diags
}

func resourceJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		input := &batch.RegisterJobDefinitionInput{
			JobDefinitionName: aws.String(name),
			Type:              aws.String(d.Get(names.AttrType).(string)),
		}

		if v, ok := d.GetOk("container_properties"); ok {
			props, err := expandJobContainerProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Batch Job Definition (%s): %s", name, err)
			}

			if aws.StringValue(input.Type) == batch.JobDefinitionTypeContainer {
				removeEmptyEnvironmentVariables(&diags, props.Environment, cty.GetAttrPath("container_properties"))
				input.ContainerProperties = props
			}
		}

		if v, ok := d.GetOk("eks_properties"); ok {
			eksProps := v.([]interface{})[0].(map[string]interface{})
			if podProps, ok := eksProps["pod_properties"].([]interface{}); ok && len(podProps) > 0 {
				props := expandEKSPodProperties(podProps[0].(map[string]interface{}))
				input.EksProperties = &batch.EksProperties{
					PodProperties: props,
				}
			}
		}

		if v, ok := d.GetOk("node_properties"); ok {
			props, err := expandJobNodeProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Batch Job Definition (%s): %s", name, err)
			}

			for _, node := range props.NodeRangeProperties {
				removeEmptyEnvironmentVariables(&diags, node.Container.Environment, cty.GetAttrPath("node_properties"))
			}
			input.NodeProperties = props
		}

		if v, ok := d.GetOk(names.AttrPropagateTags); ok {
			input.PropagateTags = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok {
			input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
			input.PlatformCapabilities = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("scheduling_priority"); ok {
			input.SchedulingPriority = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.RetryStrategy = expandRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Timeout = expandJobTimeout(v.([]interface{})[0].(map[string]interface{}))
		}

		jd, err := conn.RegisterJobDefinitionWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Job Definition (%s): %s", name, err)
		}

		// arn contains revision which is used in the Read call
		currentARN := d.Get(names.AttrARN).(string)
		d.SetId(aws.StringValue(jd.JobDefinitionArn))
		d.Set("revision", jd.Revision)
		d.Set(names.AttrARN, jd.JobDefinitionArn)

		if v := d.Get("deregister_on_new_revision"); v == true {
			log.Printf("[DEBUG] Deleting Previous Batch Job Definition: %s", currentARN)
			_, err := conn.DeregisterJobDefinitionWithContext(ctx, &batch.DeregisterJobDefinitionInput{
				JobDefinition: aws.String(currentARN),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Batch Job Definition (%s): %s", currentARN, err)
			}
		}
	}

	return append(diags, resourceJobDefinitionRead(ctx, d, meta)...)
}

func resourceJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	name := d.Get(names.AttrName).(string)
	jds, err := ListActiveJobDefinitionByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Batch Job Definitions (%s): %s", name, err)
	}

	for i := range jds {
		arn := aws.StringValue(jds[i].JobDefinitionArn)
		log.Printf("[DEBUG] Deleting Batch Job Definition: %s", arn)
		_, err := conn.DeregisterJobDefinitionWithContext(ctx, &batch.DeregisterJobDefinitionInput{
			JobDefinition: aws.String(arn),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Batch Job Definition (%s): %s", arn, err)
		}
	}

	return diags
}

func FindJobDefinitionByARN(ctx context.Context, conn *batch.Batch, arn string) (*batch.JobDefinition, error) {
	const (
		jobDefinitionStatusInactive = "INACTIVE"
	)
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: aws.StringSlice([]string{arn}),
	}

	output, err := findJobDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == jobDefinitionStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func ListActiveJobDefinitionByName(ctx context.Context, conn *batch.Batch, name string) ([]*batch.JobDefinition, error) {
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitionName: aws.String(name),
		Status:            aws.String(jobDefinitionStatusActive),
	}

	output, err := conn.DescribeJobDefinitionsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	return output.JobDefinitions, nil
}

func findJobDefinition(ctx context.Context, conn *batch.Batch, input *batch.DescribeJobDefinitionsInput) (*batch.JobDefinition, error) {
	output, err := conn.DescribeJobDefinitionsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.JobDefinitions) == 0 || output.JobDefinitions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.JobDefinitions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.JobDefinitions[0], nil
}

func validJobContainerProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func expandJobContainerProperties(rawProps string) (*batch.ContainerProperties, error) {
	var props *batch.ContainerProperties

	err := json.Unmarshal([]byte(rawProps), &props)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %s", err)
	}

	return props, nil
}

// Convert batch.ContainerProperties object into its JSON representation
func flattenContainerProperties(containerProperties *batch.ContainerProperties) (string, error) {
	b, err := jsonutil.BuildJSON(containerProperties)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func validJobNodeProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobNodeProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job node_properties is invalid: %s", err))
	}
	return
}

func expandJobNodeProperties(rawProps string) (*batch.NodeProperties, error) {
	var props *batch.NodeProperties

	err := json.Unmarshal([]byte(rawProps), &props)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %s", err)
	}

	return props, nil
}

// Convert batch.NodeProperties object into its JSON representation
func flattenNodeProperties(nodeProperties *batch.NodeProperties) (string, error) {
	b, err := jsonutil.BuildJSON(nodeProperties)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func expandJobDefinitionParameters(params map[string]interface{}) map[string]*string {
	var jobParams = make(map[string]*string)
	for k, v := range params {
		jobParams[k] = aws.String(v.(string))
	}

	return jobParams
}

func expandRetryStrategy(tfMap map[string]interface{}) *batch.RetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.RetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok && v != 0 {
		apiObject.Attempts = aws.Int64(int64(v))
	}

	if v, ok := tfMap["evaluate_on_exit"].([]interface{}); ok && len(v) > 0 {
		apiObject.EvaluateOnExit = expandEvaluateOnExits(v)
	}

	return apiObject
}

func expandEvaluateOnExit(tfMap map[string]interface{}) *batch.EvaluateOnExit {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.EvaluateOnExit{}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		apiObject.Action = aws.String(strings.ToLower(v))
	}

	if v, ok := tfMap["on_exit_code"].(string); ok && v != "" {
		apiObject.OnExitCode = aws.String(v)
	}

	if v, ok := tfMap["on_reason"].(string); ok && v != "" {
		apiObject.OnReason = aws.String(v)
	}

	if v, ok := tfMap["on_status_reason"].(string); ok && v != "" {
		apiObject.OnStatusReason = aws.String(v)
	}

	return apiObject
}

func expandEvaluateOnExits(tfList []interface{}) []*batch.EvaluateOnExit {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*batch.EvaluateOnExit

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEvaluateOnExit(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenRetryStrategy(apiObject *batch.RetryStrategy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attempts; v != nil {
		tfMap["attempts"] = aws.Int64Value(v)
	}

	if v := apiObject.EvaluateOnExit; v != nil {
		tfMap["evaluate_on_exit"] = flattenEvaluateOnExits(v)
	}

	return tfMap
}

func flattenEvaluateOnExit(apiObject *batch.EvaluateOnExit) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Action; v != nil {
		tfMap[names.AttrAction] = aws.StringValue(v)
	}

	if v := apiObject.OnExitCode; v != nil {
		tfMap["on_exit_code"] = aws.StringValue(v)
	}

	if v := apiObject.OnReason; v != nil {
		tfMap["on_reason"] = aws.StringValue(v)
	}

	if v := apiObject.OnStatusReason; v != nil {
		tfMap["on_status_reason"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEvaluateOnExits(apiObjects []*batch.EvaluateOnExit) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEvaluateOnExit(apiObject))
	}

	return tfList
}

func expandJobTimeout(tfMap map[string]interface{}) *batch.JobTimeout {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.JobTimeout{}

	if v, ok := tfMap["attempt_duration_seconds"].(int); ok && v != 0 {
		apiObject.AttemptDurationSeconds = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenJobTimeout(apiObject *batch.JobTimeout) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttemptDurationSeconds; v != nil {
		tfMap["attempt_duration_seconds"] = aws.Int64Value(v)
	}

	return tfMap
}

func removeEmptyEnvironmentVariables(diags *diag.Diagnostics, environment []*batch.KeyValuePair, attributePath cty.Path) {
	for _, env := range environment {
		if aws.StringValue(env.Value) == "" {
			*diags = append(*diags, errs.NewAttributeWarningDiagnostic(
				attributePath,
				"Ignoring environment variable",
				fmt.Sprintf("The environment variable %q has an empty value, which is ignored by the Batch service", aws.StringValue(env.Name))),
			)
		}
	}
}
