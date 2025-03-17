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
				ConflictsWith: []string{"ecs_properties", "eks_properties", "node_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentContainerPropertiesJSON(old, new)
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
			"ecs_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"container_properties", "eks_properties", "node_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentECSPropertiesJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobECSProperties,
			},
			"eks_properties": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"ecs_properties", "container_properties", "node_properties"},
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
										MaxItems: 10,
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
									"init_containers": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 10,
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
									"share_process_namespace": {
										Type:     schema.TypeBool,
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
				ConflictsWith: []string{"container_properties", "ecs_properties", "eks_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentNodePropertiesJSON(old, new)
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

		CustomizeDiff: jobDefinitionCustomizeDiff,
	}
}

func jobDefinitionCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
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

		equivalent, err := equivalentContainerPropertiesJSON(o.(string), n.(string))
		if err != nil {
			return false
		}

		if !equivalent {
			return true
		}
	}

	if d.HasChange("ecs_properties") {
		o, n := d.GetChange("ecs_properties")

		equivalent, err := equivalentECSPropertiesJSON(o.(string), n.(string))
		if err != nil {
			return false
		}

		if !equivalent {
			return true
		}
	}

	if d.HasChange("node_properties") {
		o, n := d.GetChange("node_properties")

		equivalent, err := equivalentNodePropertiesJSON(o.(string), n.(string))
		if err != nil {
			return false
		}

		if !equivalent {
			return true
		}
	}

	if d.HasChange("eks_properties") {
		o, n := d.GetChange("eks_properties")
		if len(o.([]any)) == 0 && len(n.([]any)) == 0 {
			return false
		}

		if awstypes.JobDefinitionType(d.Get(names.AttrType).(string)) != awstypes.JobDefinitionTypeContainer {
			return false
		}

		var oeks, neks *awstypes.EksPodProperties
		if len(o.([]any)) > 0 && o.([]any)[0] != nil {
			oProps := o.([]any)[0].(map[string]any)
			if opodProps, ok := oProps["pod_properties"].([]any); ok && len(opodProps) > 0 {
				oeks = expandEKSPodProperties(opodProps[0].(map[string]any))
			}
		}

		if len(n.([]any)) > 0 && n.([]any)[0] != nil {
			nProps := n.([]any)[0].(map[string]any)
			if npodProps, ok := nProps["pod_properties"].([]any); ok && len(npodProps) > 0 {
				neks = expandEKSPodProperties(npodProps[0].(map[string]any))
			}
		}

		return !reflect.DeepEqual(oeks, neks)
	}

	if d.HasChange("retry_strategy") {
		o, n := d.GetChange("retry_strategy")
		if len(o.([]any)) == 0 && len(n.([]any)) == 0 {
			return false
		}

		var ors, nrs *awstypes.RetryStrategy
		if len(o.([]any)) > 0 && o.([]any)[0] != nil {
			oProps := o.([]any)[0].(map[string]any)
			ors = expandRetryStrategy(oProps)
		}

		if len(n.([]any)) > 0 && n.([]any)[0] != nil {
			nProps := n.([]any)[0].(map[string]any)
			nrs = expandRetryStrategy(nProps)
		}

		return !reflect.DeepEqual(ors, nrs)
	}

	if d.HasChange(names.AttrTimeout) {
		o, n := d.GetChange(names.AttrTimeout)
		if len(o.([]any)) == 0 && len(n.([]any)) == 0 {
			return false
		}

		var ors, nrs *awstypes.JobTimeout
		if len(o.([]any)) > 0 && o.([]any)[0] != nil {
			oProps := o.([]any)[0].(map[string]any)
			ors = expandJobTimeout(oProps)
		}

		if len(n.([]any)) > 0 && n.([]any)[0] != nil {
			nProps := n.([]any)[0].(map[string]any)
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

func resourceJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	switch jobDefinitionType {
	case awstypes.JobDefinitionTypeContainer:
		if v, ok := d.GetOk("node_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `node_properties` can be specified when `type` is %q", jobDefinitionType)
		}

		if v, ok := d.GetOk("container_properties"); ok {
			props, err := expandContainerProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			diags = append(diags, removeEmptyEnvironmentVariables(props.Environment, cty.GetAttrPath("container_properties"))...)
			input.ContainerProperties = props
		}

		if v, ok := d.GetOk("ecs_properties"); ok {
			props, err := expandECSProperties(v.(string))
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			for _, taskProps := range props.TaskProperties {
				for _, container := range taskProps.Containers {
					diags = append(diags, removeEmptyEnvironmentVariables(container.Environment, cty.GetAttrPath("ecs_properties"))...)
				}
			}
			input.EcsProperties = props
		}

		if v, ok := d.GetOk("eks_properties"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			eksProps := v.([]any)[0].(map[string]any)
			if podProps, ok := eksProps["pod_properties"].([]any); ok && len(podProps) > 0 {
				props := expandEKSPodProperties(podProps[0].(map[string]any))
				input.EksProperties = &awstypes.EksProperties{
					PodProperties: props,
				}
			}
		}

	case awstypes.JobDefinitionTypeMultinode:
		if v, ok := d.GetOk("container_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `container_properties` can be specified when `type` is %q", jobDefinitionType)
		}
		if v, ok := d.GetOk("ecs_properties"); ok && v != nil {
			return sdkdiag.AppendErrorf(diags, "No `ecs_properties` can be specified when `type` is %q", jobDefinitionType)
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
				if node.Container != nil {
					diags = append(diags, removeEmptyEnvironmentVariables(node.Container.Environment, cty.GetAttrPath("node_properties"))...)
				}
			}
			input.NodeProperties = props
		}
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = flex.ExpandStringyValueSet[awstypes.PlatformCapability](v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.RetryStrategy = expandRetryStrategy(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("scheduling_priority"); ok {
		input.SchedulingPriority = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Timeout = expandJobTimeout(v.([]any)[0].(map[string]any))
	}

	output, err := conn.RegisterJobDefinition(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Batch Job Definition (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.JobDefinitionArn))

	return append(diags, resourceJobDefinitionRead(ctx, d, meta)...)
}

func resourceJobDefinitionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	d.Set("container_properties", containerProperties)
	ecsProperties, err := flattenECSProperties(jobDefinition.EcsProperties)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("ecs_properties", ecsProperties)
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
		if err := d.Set("retry_strategy", []any{flattenRetryStrategy(jobDefinition.RetryStrategy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting retry_strategy: %s", err)
		}
	} else {
		d.Set("retry_strategy", nil)
	}
	d.Set("revision", revision)
	d.Set("scheduling_priority", jobDefinition.SchedulingPriority)
	if jobDefinition.Timeout != nil {
		if err := d.Set(names.AttrTimeout, []any{flattenJobTimeout(jobDefinition.Timeout)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting timeout: %s", err)
		}
	} else {
		d.Set(names.AttrTimeout, nil)
	}
	d.Set(names.AttrType, jobDefinition.Type)

	setTagsOut(ctx, jobDefinition.Tags)

	return diags
}

func resourceJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		jobDefinitionType := awstypes.JobDefinitionType(d.Get(names.AttrType).(string))
		input := &batch.RegisterJobDefinitionInput{
			JobDefinitionName: aws.String(name),
			Tags:              getTagsIn(ctx),
			Type:              jobDefinitionType,
		}

		switch jobDefinitionType {
		case awstypes.JobDefinitionTypeContainer:
			if v, ok := d.GetOk("container_properties"); ok {
				props, err := expandContainerProperties(v.(string))
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				diags = append(diags, removeEmptyEnvironmentVariables(props.Environment, cty.GetAttrPath("container_properties"))...)
				input.ContainerProperties = props
			}

			if v, ok := d.GetOk("ecs_properties"); ok {
				props, err := expandECSProperties(v.(string))
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				for _, taskProps := range props.TaskProperties {
					for _, container := range taskProps.Containers {
						diags = append(diags, removeEmptyEnvironmentVariables(container.Environment, cty.GetAttrPath("ecs_properties"))...)
					}
				}
				input.EcsProperties = props
			}

			if v, ok := d.GetOk("eks_properties"); ok {
				eksProps := v.([]any)[0].(map[string]any)
				if podProps, ok := eksProps["pod_properties"].([]any); ok && len(podProps) > 0 {
					props := expandEKSPodProperties(podProps[0].(map[string]any))
					input.EksProperties = &awstypes.EksProperties{
						PodProperties: props,
					}
				}
			}

		case awstypes.JobDefinitionTypeMultinode:
			if v, ok := d.GetOk("node_properties"); ok {
				props, err := expandJobNodeProperties(v.(string))
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				for _, node := range props.NodeRangeProperties {
					diags = append(diags, removeEmptyEnvironmentVariables(node.Container.Environment, cty.GetAttrPath("node_properties"))...)
				}
				input.NodeProperties = props
			}
		}

		if v, ok := d.GetOk(names.AttrParameters); ok {
			input.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
		}

		if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
			input.PlatformCapabilities = flex.ExpandStringyValueSet[awstypes.PlatformCapability](v.(*schema.Set))
		}

		if v, ok := d.GetOk(names.AttrPropagateTags); ok {
			input.PropagateTags = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.RetryStrategy = expandRetryStrategy(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("scheduling_priority"); ok {
			input.SchedulingPriority = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrTimeout); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Timeout = expandJobTimeout(v.([]any)[0].(map[string]any))
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
			input := batch.DeregisterJobDefinitionInput{
				JobDefinition: aws.String(currentARN),
			}
			_, err := conn.DeregisterJobDefinition(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Batch Job Definition (%s): %s", currentARN, err)
			}
		}
	}

	return append(diags, resourceJobDefinitionRead(ctx, d, meta)...)
}

func resourceJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
		input := batch.DeregisterJobDefinitionInput{
			JobDefinition: aws.String(arn),
		}
		_, err := conn.DeregisterJobDefinition(ctx, &input)

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

func validJobContainerProperties(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func validJobECSProperties(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandECSProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job ecs_properties is invalid: %s", err))
	}
	return
}

func validJobNodeProperties(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobNodeProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job node_properties is invalid: %s", err))
	}
	return
}

func expandRetryStrategy(tfMap map[string]any) *awstypes.RetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok && v != 0 {
		apiObject.Attempts = aws.Int32(int32(v))
	}

	if v, ok := tfMap["evaluate_on_exit"].([]any); ok && len(v) > 0 {
		apiObject.EvaluateOnExit = expandEvaluateOnExits(v)
	}

	return apiObject
}

func expandEvaluateOnExit(tfMap map[string]any) *awstypes.EvaluateOnExit {
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

func expandEvaluateOnExits(tfList []any) []awstypes.EvaluateOnExit {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EvaluateOnExit

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

func flattenRetryStrategy(apiObject *awstypes.RetryStrategy) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Attempts; v != nil {
		tfMap["attempts"] = aws.ToInt32(v)
	}

	if v := apiObject.EvaluateOnExit; v != nil {
		tfMap["evaluate_on_exit"] = flattenEvaluateOnExits(v)
	}

	return tfMap
}

func flattenEvaluateOnExit(apiObject *awstypes.EvaluateOnExit) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

func flattenEvaluateOnExits(apiObjects []awstypes.EvaluateOnExit) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenEvaluateOnExit(&apiObject))
	}

	return tfList
}

func expandJobTimeout(tfMap map[string]any) *awstypes.JobTimeout {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.JobTimeout{}

	if v, ok := tfMap["attempt_duration_seconds"].(int); ok && v != 0 {
		apiObject.AttemptDurationSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenJobTimeout(apiObject *awstypes.JobTimeout) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AttemptDurationSeconds; v != nil {
		tfMap["attempt_duration_seconds"] = aws.ToInt32(v)
	}

	return tfMap
}

func removeEmptyEnvironmentVariables(environment []awstypes.KeyValuePair, attributePath cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, env := range environment {
		if aws.ToString(env.Value) == "" {
			diags = append(diags, errs.NewAttributeWarningDiagnostic(
				attributePath,
				"Ignoring environment variable",
				fmt.Sprintf("The environment variable %q has an empty value, which is ignored by the Batch service", aws.ToString(env.Name))),
			)
		}
	}

	return diags
}

func expandEKSPodProperties(tfMap map[string]any) *awstypes.EksPodProperties {
	apiObject := &awstypes.EksPodProperties{}

	if v, ok := tfMap["containers"]; ok {
		apiObject.Containers = expandContainers(v.([]any))
	}

	if v, ok := tfMap["dns_policy"].(string); ok && v != "" {
		apiObject.DnsPolicy = aws.String(v)
	}

	if v, ok := tfMap["host_network"]; ok {
		apiObject.HostNetwork = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["image_pull_secret"]; ok {
		apiObject.ImagePullSecrets = expandImagePullSecrets(v.([]any))
	}

	if v, ok := tfMap["init_containers"]; ok {
		apiObject.InitContainers = expandContainers(v.([]any))
	}

	if v, ok := tfMap["metadata"].([]any); ok && len(v) > 0 {
		if v, ok := v[0].(map[string]any)["labels"]; ok {
			apiObject.Metadata = &awstypes.EksMetadata{
				Labels: flex.ExpandStringValueMap(v.(map[string]any)),
			}
		}
	}

	if v, ok := tfMap["service_account_name"].(string); ok && v != "" {
		apiObject.ServiceAccountName = aws.String(v)
	}

	if v, ok := tfMap["share_process_namespace"]; ok {
		apiObject.ShareProcessNamespace = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["volumes"]; ok {
		apiObject.Volumes = expandVolumes(v.([]any))
	}

	return apiObject
}

func expandContainers(tfList []any) []awstypes.EksContainer {
	var apiObjects []awstypes.EksContainer

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.EksContainer{}

		if v, ok := tfMap["args"]; ok {
			apiObject.Args = flex.ExpandStringValueList(v.([]any))
		}

		if v, ok := tfMap["command"]; ok {
			apiObject.Command = flex.ExpandStringValueList(v.([]any))
		}

		if v, ok := tfMap["env"].(*schema.Set); ok && v.Len() > 0 {
			apiObjects := []awstypes.EksContainerEnvironmentVariable{}

			for _, tfMapRaw := range v.List() {
				apiObject := awstypes.EksContainerEnvironmentVariable{}
				tfMap := tfMapRaw.(map[string]any)

				if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
					apiObject.Name = aws.String(v)
				}

				if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
					apiObject.Value = aws.String(v)
				}

				apiObjects = append(apiObjects, apiObject)
			}

			apiObject.Env = apiObjects
		}

		if v, ok := tfMap["image"]; ok {
			apiObject.Image = aws.String(v.(string))
		}

		if v, ok := tfMap["image_pull_policy"].(string); ok && v != "" {
			apiObject.ImagePullPolicy = aws.String(v)
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap[names.AttrResources].([]any); ok && len(v) > 0 {
			resources := &awstypes.EksContainerResourceRequirements{}
			tfMap := v[0].(map[string]any)

			if v, ok := tfMap["limits"]; ok {
				resources.Limits = flex.ExpandStringValueMap(v.(map[string]any))
			}

			if v, ok := tfMap["requests"]; ok {
				resources.Requests = flex.ExpandStringValueMap(v.(map[string]any))
			}

			apiObject.Resources = resources
		}

		if v, ok := tfMap["security_context"].([]any); ok && len(v) > 0 {
			securityContext := &awstypes.EksContainerSecurityContext{}
			tfMap := v[0].(map[string]any)

			if v, ok := tfMap["privileged"]; ok {
				securityContext.Privileged = aws.Bool(v.(bool))
			}

			if v, ok := tfMap["read_only_root_file_system"]; ok {
				securityContext.ReadOnlyRootFilesystem = aws.Bool(v.(bool))
			}

			if v, ok := tfMap["run_as_group"]; ok {
				securityContext.RunAsGroup = aws.Int64(int64(v.(int)))
			}

			if v, ok := tfMap["run_as_non_root"]; ok {
				securityContext.RunAsNonRoot = aws.Bool(v.(bool))
			}

			if v, ok := tfMap["run_as_user"]; ok {
				securityContext.RunAsUser = aws.Int64(int64(v.(int)))
			}

			apiObject.SecurityContext = securityContext
		}

		if v, ok := tfMap["volume_mounts"]; ok {
			apiObject.VolumeMounts = expandVolumeMounts(v.([]any))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandImagePullSecrets(tfList []any) []awstypes.ImagePullSecret {
	var apiObjects []awstypes.ImagePullSecret

	for _, tfMapRaw := range tfList {
		apiObject := awstypes.ImagePullSecret{}
		tfMap := tfMapRaw.(map[string]any)

		if v, ok := tfMap[names.AttrName].(string); ok {
			apiObject.Name = aws.String(v)
			apiObjects = append(apiObjects, apiObject) // move out of "if" when more fields are added
		}
	}

	return apiObjects
}

func expandVolumes(tfList []any) []awstypes.EksVolume {
	var apiObjects []awstypes.EksVolume

	for _, tfMapRaw := range tfList {
		apiObject := awstypes.EksVolume{}
		tfMap := tfMapRaw.(map[string]any)

		if v, ok := tfMap["empty_dir"].([]any); ok && len(v) > 0 {
			if v, ok := v[0].(map[string]any); ok {
				apiObject.EmptyDir = &awstypes.EksEmptyDir{
					Medium:    aws.String(v["medium"].(string)),
					SizeLimit: aws.String(v["size_limit"].(string)),
				}
			}
		}

		if v, ok := tfMap[names.AttrName].(string); ok {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap["host_path"].([]any); ok && len(v) > 0 {
			apiObject.HostPath = &awstypes.EksHostPath{}

			if v, ok := v[0].(map[string]any); ok {
				if v, ok := v[names.AttrPath]; ok {
					apiObject.HostPath.Path = aws.String(v.(string))
				}
			}
		}

		if v, ok := tfMap["secret"].([]any); ok && len(v) > 0 {
			apiObject.Secret = &awstypes.EksSecret{}

			if v := v[0].(map[string]any); ok {
				if v, ok := v["optional"]; ok {
					apiObject.Secret.Optional = aws.Bool(v.(bool))
				}

				if v, ok := v["secret_name"]; ok {
					apiObject.Secret.SecretName = aws.String(v.(string))
				}
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandVolumeMounts(tfList []any) []awstypes.EksContainerVolumeMount {
	var apiObjects []awstypes.EksContainerVolumeMount

	for _, tfMapRaw := range tfList {
		apiObject := awstypes.EksContainerVolumeMount{}
		tfMap := tfMapRaw.(map[string]any)

		if v, ok := tfMap["mount_path"]; ok {
			apiObject.MountPath = aws.String(v.(string))
		}

		if v, ok := tfMap[names.AttrName]; ok {
			apiObject.Name = aws.String(v.(string))
		}

		if v, ok := tfMap["read_only"]; ok {
			apiObject.ReadOnly = aws.Bool(v.(bool))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEKSProperties(apiObject *awstypes.EksProperties) []any {
	var tfList []any

	if apiObject == nil {
		return tfList
	}

	if v := apiObject.PodProperties; v != nil {
		tfList = append(tfList, map[string]any{
			"pod_properties": flattenEKSPodProperties(apiObject.PodProperties),
		})
	}

	return tfList
}

func flattenEKSPodProperties(apiObject *awstypes.EksPodProperties) []any {
	var tfList []any
	tfMap := make(map[string]any, 0)

	if v := apiObject.Containers; v != nil {
		tfMap["containers"] = flattenEKSContainers(v)
	}

	if v := apiObject.DnsPolicy; v != nil {
		tfMap["dns_policy"] = aws.ToString(v)
	}

	if v := apiObject.HostNetwork; v != nil {
		tfMap["host_network"] = aws.ToBool(v)
	}

	if v := apiObject.ImagePullSecrets; v != nil {
		tfMap["image_pull_secret"] = flattenImagePullSecrets(v)
	}

	if v := apiObject.InitContainers; v != nil {
		tfMap["init_containers"] = flattenEKSContainers(v)
	}

	if v := apiObject.Metadata; v != nil {
		metadata := make([]map[string]any, 0)

		if v := v.Labels; v != nil {
			metadata = append(metadata, map[string]any{
				"labels": v,
			})
		}

		tfMap["metadata"] = metadata
	}

	if v := apiObject.ServiceAccountName; v != nil {
		tfMap["service_account_name"] = aws.ToString(v)
	}

	if v := apiObject.ShareProcessNamespace; v != nil {
		tfMap["share_process_namespace"] = aws.ToBool(v)
	}

	if v := apiObject.Volumes; v != nil {
		tfMap["volumes"] = flattenEKSVolumes(v)
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenImagePullSecrets(apiObjects []awstypes.ImagePullSecret) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
			tfList = append(tfList, tfMap) // move out of "if" when more fields are added
		}
	}

	return tfList
}

func flattenEKSContainers(apiObjects []awstypes.EksContainer) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.Args; v != nil {
			tfMap["args"] = v
		}

		if v := apiObject.Command; v != nil {
			tfMap["command"] = v
		}

		if v := apiObject.Env; v != nil {
			tfMap["env"] = flattenEKSContainerEnvironmentVariables(v)
		}

		if v := apiObject.Image; v != nil {
			tfMap["image"] = aws.ToString(v)
		}

		if v := apiObject.ImagePullPolicy; v != nil {
			tfMap["image_pull_policy"] = aws.ToString(v)
		}

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		if v := apiObject.Resources; v != nil {
			tfMap[names.AttrResources] = []map[string]any{{
				"limits":   v.Limits,
				"requests": v.Requests,
			}}
		}

		if v := apiObject.SecurityContext; v != nil {
			tfMap["security_context"] = []map[string]any{{
				"privileged":                 aws.ToBool(v.Privileged),
				"read_only_root_file_system": aws.ToBool(v.ReadOnlyRootFilesystem),
				"run_as_group":               aws.ToInt64(v.RunAsGroup),
				"run_as_non_root":            aws.ToBool(v.RunAsNonRoot),
				"run_as_user":                aws.ToInt64(v.RunAsUser),
			}}
		}

		if v := apiObject.VolumeMounts; v != nil {
			tfMap["volume_mounts"] = flattenEKSContainerVolumeMounts(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSContainerEnvironmentVariables(apiObjects []awstypes.EksContainerEnvironmentVariable) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		if v := apiObject.Value; v != nil {
			tfMap[names.AttrValue] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSContainerVolumeMounts(apiObjects []awstypes.EksContainerVolumeMount) []any {
	var tfList []any

	for _, v := range apiObjects {
		tfMap := map[string]any{}

		if v := v.MountPath; v != nil {
			tfMap["mount_path"] = aws.ToString(v)
		}

		if v := v.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		if v := v.ReadOnly; v != nil {
			tfMap["read_only"] = aws.ToBool(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSVolumes(apiObjects []awstypes.EksVolume) []any {
	var tfList []any

	for _, v := range apiObjects {
		tfMap := map[string]any{}

		if v := v.EmptyDir; v != nil {
			tfMap["empty_dir"] = []map[string]any{{
				"medium":     aws.ToString(v.Medium),
				"size_limit": aws.ToString(v.SizeLimit),
			}}
		}

		if v := v.HostPath; v != nil {
			tfMap["host_path"] = []map[string]any{{
				names.AttrPath: aws.ToString(v.Path),
			}}
		}

		if v := v.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}

		if v := v.Secret; v != nil {
			tfMap["secret"] = []map[string]any{{
				"optional":    aws.ToBool(v.Optional),
				"secret_name": aws.ToString(v.SecretName),
			}}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
