// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_batch_job_definition", name="Job Definition")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/batch/types;types.JobDefinition", importIgnore="deregister_on_new_revision")
func resourceJobDefinition() *schema.Resource {
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
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobContainerProperties,
			},
			"deregister_on_new_revision": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
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
													ValidateFunc: validation.StringInSlice(imagePullPolicy_Values(), false),
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
										ValidateFunc: validation.StringInSlice(dnsPolicy_Values(), false),
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
															"optional": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"secret_name": {
																Type:     schema.TypeString,
																Required: true,
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
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobNodeProperties,
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.PlatformCapability](),
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
										Type:             schema.TypeString,
										Required:         true,
										StateFunc:        sdkv2.ToLowerSchemaStateFunc,
										ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.RetryAction](),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.JobDefinitionType](),
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

		if awstypes.JobDefinitionType(d.Get(names.AttrType).(string)) != awstypes.JobDefinitionTypeContainer {
			return false
		}

		var oeks, neks *awstypes.EksPodProperties
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			oProps := o.([]interface{})[0].(map[string]interface{})
			if opodProps, ok := oProps["pod_properties"].([]interface{}); ok && len(opodProps) > 0 {
				oeks = expandEKSPodProperties(opodProps[0].(map[string]interface{}))
			}
		}

		if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
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

		var ors, nrs *awstypes.RetryStrategy
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			oProps := o.([]interface{})[0].(map[string]interface{})
			ors = expandRetryStrategy(oProps)
		}

		if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
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

		var ors, nrs *awstypes.JobTimeout
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			oProps := o.([]interface{})[0].(map[string]interface{})
			ors = expandJobTimeout(oProps)
		}

		if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
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
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	name := d.Get(names.AttrName).(string)
	jobDefinitionType := awstypes.JobDefinitionType(d.Get(names.AttrType).(string))
	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(name),
		PropagateTags:     aws.Bool(d.Get(names.AttrPropagateTags).(bool)),
		Tags:              getTagsIn(ctx),
		Type:              jobDefinitionType,
	}

	if jobDefinitionType == awstypes.JobDefinitionTypeContainer {
		if v, ok := d.GetOk("node_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `node_properties` can be specified when `type` is %q", jobDefinitionType)
		}

		if v, ok := d.GetOk("container_properties"); ok {
			props, err := expandJobContainerProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			removeEmptyEnvironmentVariables(&diags, props.Environment, cty.GetAttrPath("container_properties"))
			input.ContainerProperties = props
		}

		if v, ok := d.GetOk("eks_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			eksProps := v.([]interface{})[0].(map[string]interface{})
			if podProps, ok := eksProps["pod_properties"].([]interface{}); ok && len(podProps) > 0 {
				props := expandEKSPodProperties(podProps[0].(map[string]interface{}))
				input.EksProperties = &awstypes.EksProperties{
					PodProperties: props,
				}
			}
		}
	}

	if jobDefinitionType == awstypes.JobDefinitionTypeMultinode {
		if v, ok := d.GetOk("container_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `container_properties` can be specified when `type` is %q", jobDefinitionType)
		}
		if v, ok := d.GetOk("eks_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `eks_properties` can be specified when `type` is %q", jobDefinitionType)
		}

		if v, ok := d.GetOk("node_properties"); ok {
			props, err := expandJobNodeProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			for _, node := range props.NodeRangeProperties {
				removeEmptyEnvironmentVariables(&diags, node.Container.Environment, cty.GetAttrPath("node_properties"))
			}
			input.NodeProperties = props
		}
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = flex.ExpandStringyValueSet[awstypes.PlatformCapability](v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetryStrategy = expandRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("scheduling_priority"); ok {
		input.SchedulingPriority = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Timeout = expandJobTimeout(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.RegisterJobDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Job Definition (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.JobDefinitionArn))

	return append(diags, resourceJobDefinitionRead(ctx, d, meta)...)
}

func resourceJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	jobDefinition, err := findJobDefinitionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Definition (%s): %s", d.Id(), err)
	}

	arn, revision := aws.ToString(jobDefinition.JobDefinitionArn), aws.ToInt32(jobDefinition.Revision)
	d.Set(names.AttrARN, arn)
	d.Set("arn_prefix", strings.TrimSuffix(arn, fmt.Sprintf(":%d", revision)))

	containerProperties, err := flattenContainerProperties(jobDefinition.ContainerProperties)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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
		return sdkdiag.AppendFromErr(diags, err)
	}
	if err := d.Set("node_properties", nodeProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_properties: %s", err)
	}
	d.Set(names.AttrParameters, jobDefinition.Parameters)
	d.Set("platform_capabilities", jobDefinition.PlatformCapabilities)
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
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		jobDefinitionType := awstypes.JobDefinitionType(d.Get(names.AttrType).(string))
		input := &batch.RegisterJobDefinitionInput{
			JobDefinitionName: aws.String(name),
			Type:              jobDefinitionType,
		}

		if v, ok := d.GetOk("container_properties"); ok {
			props, err := expandJobContainerProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			if jobDefinitionType == awstypes.JobDefinitionTypeContainer {
				removeEmptyEnvironmentVariables(&diags, props.Environment, cty.GetAttrPath("container_properties"))
				input.ContainerProperties = props
			}
		}

		if v, ok := d.GetOk("eks_properties"); ok {
			eksProps := v.([]interface{})[0].(map[string]interface{})
			if podProps, ok := eksProps["pod_properties"].([]interface{}); ok && len(podProps) > 0 {
				props := expandEKSPodProperties(podProps[0].(map[string]interface{}))
				input.EksProperties = &awstypes.EksProperties{
					PodProperties: props,
				}
			}
		}

		if v, ok := d.GetOk("node_properties"); ok {
			props, err := expandJobNodeProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
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
			input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}

		if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
			input.PlatformCapabilities = flex.ExpandStringyValueSet[awstypes.PlatformCapability](v.(*schema.Set))
		}

		if v, ok := d.GetOk("scheduling_priority"); ok {
			input.SchedulingPriority = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.RetryStrategy = expandRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Timeout = expandJobTimeout(v.([]interface{})[0].(map[string]interface{}))
		}

		jd, err := conn.RegisterJobDefinition(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Batch Job Definition (%s): %s", name, err)
		}

		// arn contains revision which is used in the Read call
		currentARN := d.Get(names.AttrARN).(string)
		newARN := aws.ToString(jd.JobDefinitionArn)
		d.SetId(newARN)
		d.Set(names.AttrARN, newARN)
		d.Set("revision", jd.Revision)

		if v := d.Get("deregister_on_new_revision"); v == true {
			log.Printf("[DEBUG] Deleting previous Batch Job Definition: %s", currentARN)
			_, err := conn.DeregisterJobDefinition(ctx, &batch.DeregisterJobDefinitionInput{
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
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitionName: aws.String(name),
		Status:            aws.String(jobDefinitionStatusActive),
	}

	jds, err := findJobDefinitions(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Definitions (%s): %s", name, err)
	}

	for i := range jds {
		arn := aws.ToString(jds[i].JobDefinitionArn)

		log.Printf("[DEBUG] Deregistering Batch Job Definition: %s", arn)
		_, err := conn.DeregisterJobDefinition(ctx, &batch.DeregisterJobDefinitionInput{
			JobDefinition: aws.String(arn),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deregistering Batch Job Definition (%s): %s", arn, err)
		}
	}

	return diags
}

func findJobDefinitionByARN(ctx context.Context, conn *batch.Client, arn string) (*awstypes.JobDefinition, error) {
	const (
		jobDefinitionStatusInactive = "INACTIVE"
	)
	input := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: []string{arn},
	}

	output, err := findJobDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == jobDefinitionStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findJobDefinition(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) (*awstypes.JobDefinition, error) {
	output, err := findJobDefinitions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findJobDefinitions(ctx context.Context, conn *batch.Client, input *batch.DescribeJobDefinitionsInput) ([]awstypes.JobDefinition, error) {
	var output []awstypes.JobDefinition

	pages := batch.NewDescribeJobDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.JobDefinitions...)
	}

	return output, nil
}

func validJobContainerProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func validJobNodeProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobNodeProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job node_properties is invalid: %s", err))
	}
	return
}

func expandRetryStrategy(tfMap map[string]interface{}) *awstypes.RetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok && v != 0 {
		apiObject.Attempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["evaluate_on_exit"].([]interface{}); ok && len(v) > 0 {
		apiObject.EvaluateOnExit = expandEvaluateOnExits(v)
	}

	return apiObject
}

func expandEvaluateOnExit(tfMap map[string]interface{}) *awstypes.EvaluateOnExit {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EvaluateOnExit{}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		apiObject.Action = awstypes.RetryAction(strings.ToLower(v))
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

func expandEvaluateOnExits(tfList []interface{}) []awstypes.EvaluateOnExit {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EvaluateOnExit

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandEvaluateOnExit(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenRetryStrategy(apiObject *awstypes.RetryStrategy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attempts; v != nil {
		tfMap["attempts"] = aws.ToInt32(v)
	}

	if v := apiObject.EvaluateOnExit; v != nil {
		tfMap["evaluate_on_exit"] = flattenEvaluateOnExits(v)
	}

	return tfMap
}

func flattenEvaluateOnExit(apiObject *awstypes.EvaluateOnExit) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrAction: apiObject.Action,
	}

	if v := apiObject.OnExitCode; v != nil {
		tfMap["on_exit_code"] = aws.ToString(v)
	}

	if v := apiObject.OnReason; v != nil {
		tfMap["on_reason"] = aws.ToString(v)
	}

	if v := apiObject.OnStatusReason; v != nil {
		tfMap["on_status_reason"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEvaluateOnExits(apiObjects []awstypes.EvaluateOnExit) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEvaluateOnExit(&apiObject))
	}

	return tfList
}

func expandJobTimeout(tfMap map[string]interface{}) *awstypes.JobTimeout {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JobTimeout{}

	if v, ok := tfMap["attempt_duration_seconds"].(int); ok && v != 0 {
		apiObject.AttemptDurationSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenJobTimeout(apiObject *awstypes.JobTimeout) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttemptDurationSeconds; v != nil {
		tfMap["attempt_duration_seconds"] = aws.ToInt32(v)
	}

	return tfMap
}

func removeEmptyEnvironmentVariables(diags *diag.Diagnostics, environment []awstypes.KeyValuePair, attributePath cty.Path) {
	for _, env := range environment {
		if aws.ToString(env.Value) == "" {
			*diags = append(*diags, errs.NewAttributeWarningDiagnostic(
				attributePath,
				"Ignoring environment variable",
				fmt.Sprintf("The environment variable %q has an empty value, which is ignored by the Batch service", aws.ToString(env.Name))),
			)
		}
	}
}
