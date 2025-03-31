// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emr_cluster", name="Cluster")
// @Tags(identifierAttribute="id")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			instanceFleetConfigSchema := func() *schema.Resource {
				return &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_type_configs": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bid_price": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"bid_price_as_percentage_of_on_demand_price": {
										Type:     schema.TypeFloat,
										Optional: true,
										ForceNew: true,
										Default:  100,
									},
									"configurations": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"classification": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												names.AttrProperties: {
													Type:     schema.TypeMap,
													Optional: true,
													ForceNew: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"ebs_config": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrIOPS: {
													Type:     schema.TypeInt,
													Optional: true,
													ForceNew: true,
												},
												names.AttrSize: {
													Type:     schema.TypeInt,
													Required: true,
													ForceNew: true,
												},
												names.AttrType: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validEBSVolumeType(),
												},
												"volumes_per_instance": {
													Type:     schema.TypeInt,
													Optional: true,
													ForceNew: true,
													Default:  1,
												},
											},
										},
										Set: resourceClusterEBSHashConfig,
									},
									names.AttrInstanceType: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"weighted_capacity": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  1,
									},
								},
							},
							Set: resourceInstanceTypeHashConfig,
						},
						"launch_specifications": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"on_demand_specification": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allocation_strategy": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.OnDemandProvisioningAllocationStrategy](),
												},
											},
										},
									},
									"spot_specification": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allocation_strategy": {
													Type:             schema.TypeString,
													ForceNew:         true,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.SpotProvisioningAllocationStrategy](),
												},
												"block_duration_minutes": {
													Type:     schema.TypeInt,
													Optional: true,
													ForceNew: true,
													Default:  0,
												},
												"timeout_action": {
													Type:             schema.TypeString,
													Required:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.SpotProvisioningTimeoutAction](),
												},
												"timeout_duration_minutes": {
													Type:     schema.TypeInt,
													ForceNew: true,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"provisioned_on_demand_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"provisioned_spot_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"target_on_demand_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  0,
						},
						"target_spot_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  0,
						},
					},
				}
			}

			return map[string]*schema.Schema{
				"additional_info": {
					Type:                  schema.TypeString,
					Optional:              true,
					ForceNew:              true,
					ValidateFunc:          validation.StringIsJSON,
					DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
					DiffSuppressOnRefresh: true,
					StateFunc: func(v any) string {
						json, _ := structure.NormalizeJsonString(v)
						return json
					},
				},
				"applications": {
					Type:     schema.TypeSet,
					Optional: true,
					ForceNew: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"auto_termination_policy": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"idle_timeout": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(60, 604800),
							},
						},
					},
				},
				"autoscaling_role": {
					Type:     schema.TypeString,
					ForceNew: true,
					Optional: true,
				},
				"bootstrap_action": {
					Type:     schema.TypeList,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"args": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrPath: {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"cluster_state": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"configurations": {
					Type:          schema.TypeString,
					ForceNew:      true,
					Optional:      true,
					ConflictsWith: []string{"configurations_json"},
				},
				"configurations_json": {
					Type:                  schema.TypeString,
					Optional:              true,
					ForceNew:              true,
					ValidateFunc:          validation.StringIsJSON,
					DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
					DiffSuppressOnRefresh: true,
					StateFunc: func(v any) string {
						json, _ := structure.NormalizeJsonString(v)
						return json
					},
					ConflictsWith: []string{"configurations"},
				},
				"core_instance_fleet": {
					Type:          schema.TypeList,
					Optional:      true,
					ForceNew:      true,
					Computed:      true,
					MaxItems:      1,
					Elem:          instanceFleetConfigSchema(),
					ConflictsWith: []string{"core_instance_group", "master_instance_group"},
				},
				"core_instance_group": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"autoscaling_policy": {
								Type:                  schema.TypeString,
								Optional:              true,
								ValidateFunc:          validation.StringIsJSON,
								DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
								DiffSuppressOnRefresh: true,
							},
							"bid_price": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"ebs_config": {
								Type:     schema.TypeSet,
								Optional: true,
								Computed: true,
								ForceNew: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrIOPS: {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
										},
										names.AttrSize: {
											Type:     schema.TypeInt,
											Required: true,
											ForceNew: true,
										},
										names.AttrThroughput: {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
										},
										names.AttrType: {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: validEBSVolumeType(),
										},
										"volumes_per_instance": {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
											Default:  1,
										},
									},
								},
								Set: resourceClusterEBSHashConfig,
							},
							names.AttrID: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrInstanceCount: {
								Type:         schema.TypeInt,
								Optional:     true,
								Default:      1,
								ValidateFunc: validation.IntAtLeast(1),
							},
							names.AttrInstanceType: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
						},
					},
				},
				"custom_ami_id": {
					Type:         schema.TypeString,
					ForceNew:     true,
					Optional:     true,
					ValidateFunc: validCustomAMIID,
				},
				"ebs_root_volume_size": {
					Type:     schema.TypeInt,
					ForceNew: true,
					Optional: true,
				},
				"ec2_attributes": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"additional_master_security_groups": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"additional_slave_security_groups": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"emr_managed_master_security_group": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},
							"emr_managed_slave_security_group": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},
							"instance_profile": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							"key_name": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"service_access_security_group": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								Computed: true,
							},
							names.AttrSubnetID: {
								Type:          schema.TypeString,
								Optional:      true,
								Computed:      true,
								ForceNew:      true,
								ConflictsWith: []string{"ec2_attributes.0.subnet_ids"},
							},
							names.AttrSubnetIDs: {
								Type:          schema.TypeSet,
								Optional:      true,
								Computed:      true,
								ForceNew:      true,
								Elem:          &schema.Schema{Type: schema.TypeString},
								ConflictsWith: []string{"ec2_attributes.0.subnet_id"},
							},
						},
					},
				},
				"keep_job_flow_alive_when_no_steps": {
					Type:     schema.TypeBool,
					ForceNew: true,
					Optional: true,
					Computed: true,
				},
				"kerberos_attributes": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					ForceNew: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ad_domain_join_password": {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
								ForceNew:  true,
							},
							"ad_domain_join_user": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"cross_realm_trust_principal_password": {
								Type:      schema.TypeString,
								Optional:  true,
								Sensitive: true,
								ForceNew:  true,
							},
							"kdc_admin_password": {
								Type:      schema.TypeString,
								Required:  true,
								Sensitive: true,
								ForceNew:  true,
							},
							"realm": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
						},
					},
				},
				"list_steps_states": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: enum.Validate[awstypes.StepState](),
					},
				},
				"log_encryption_kms_key_id": {
					Type:     schema.TypeString,
					ForceNew: true,
					Optional: true,
				},
				"log_uri": {
					Type:     schema.TypeString,
					ForceNew: true,
					Optional: true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// EMR uses a proprietary filesystem called EMRFS
						// and both s3n & s3 protocols are mapped to that FS
						// so they're equvivalent in this context (confirmed by AWS support)
						old = strings.Replace(old, "s3n://", "s3://", -1)
						return old == new
					},
				},
				"master_instance_fleet": {
					Type:          schema.TypeList,
					Optional:      true,
					ForceNew:      true,
					Computed:      true,
					MaxItems:      1,
					Elem:          instanceFleetConfigSchema(),
					ConflictsWith: []string{"core_instance_group", "master_instance_group"},
				},
				"master_instance_group": {
					Type:     schema.TypeList,
					Optional: true,
					Computed: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bid_price": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
							"ebs_config": {
								Type:     schema.TypeSet,
								Optional: true,
								Computed: true,
								ForceNew: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrIOPS: {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
										},
										names.AttrSize: {
											Type:     schema.TypeInt,
											Required: true,
											ForceNew: true,
										},
										names.AttrThroughput: {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
										},
										names.AttrType: {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: validEBSVolumeType(),
										},
										"volumes_per_instance": {
											Type:     schema.TypeInt,
											Optional: true,
											ForceNew: true,
											Default:  1,
										},
									},
								},
								Set: resourceClusterEBSHashConfig,
							},
							names.AttrID: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrInstanceCount: {
								Type:         schema.TypeInt,
								Optional:     true,
								ForceNew:     true,
								Default:      1,
								ValidateFunc: validation.IntInSlice([]int{1, 3}),
							},
							names.AttrInstanceType: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
							},
						},
					},
				},
				"master_public_dns": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					ForceNew: true,
					Required: true,
				},
				"placement_group_config": {
					Type:       schema.TypeList,
					ForceNew:   true,
					Optional:   true,
					ConfigMode: schema.SchemaConfigModeAttr,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"instance_role": {
								Type:             schema.TypeString,
								ForceNew:         true,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.InstanceRoleType](),
							},
							"placement_strategy": {
								Type:             schema.TypeString,
								ForceNew:         true,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[awstypes.PlacementGroupStrategy](),
							},
						},
					},
				},
				"release_label": {
					Type:     schema.TypeString,
					ForceNew: true,
					Required: true,
				},
				"scale_down_behavior": {
					Type:             schema.TypeString,
					ForceNew:         true,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ScaleDownBehavior](),
				},
				"security_configuration": {
					Type:     schema.TypeString,
					ForceNew: true,
					Optional: true,
				},
				names.AttrServiceRole: {
					Type:     schema.TypeString,
					ForceNew: true,
					Required: true,
				},
				"step": {
					Type:       schema.TypeList,
					Optional:   true,
					Computed:   true,
					ForceNew:   true,
					ConfigMode: schema.SchemaConfigModeAttr,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"action_on_failure": {
								Type:             schema.TypeString,
								Required:         true,
								ForceNew:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ActionOnFailure](),
							},
							"hadoop_jar_step": {
								Type:       schema.TypeList,
								MaxItems:   1,
								Required:   true,
								ForceNew:   true,
								ConfigMode: schema.SchemaConfigModeAttr,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"args": {
											Type:     schema.TypeList,
											Optional: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"jar": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"main_class": {
											Type:     schema.TypeString,
											Optional: true,
											ForceNew: true,
										},
										names.AttrProperties: {
											Type:     schema.TypeMap,
											Optional: true,
											ForceNew: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
									},
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
						},
					},
				},
				"step_concurrency_level": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      1,
					ValidateFunc: validation.IntBetween(1, 256),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"termination_protection": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
				"unhealthy_node_replacement": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"visible_to_all_users": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
			}
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	applications := d.Get("applications").(*schema.Set).List()
	keepJobFlowAliveWhenNoSteps := true
	if v, ok := d.GetOkExists("keep_job_flow_alive_when_no_steps"); ok {
		keepJobFlowAliveWhenNoSteps = v.(bool)
	}

	// For multiple master nodes, EMR automatically enables
	// termination protection and ignores this configuration at launch.
	// There is additional handling after the job flow is running
	// to potentially disable termination protection to match the
	// desired Terraform configuration.
	terminationProtection := false
	if v, ok := d.GetOk("termination_protection"); ok {
		terminationProtection = v.(bool)
	}

	unhealthyNodeReplacement := false
	if v, ok := d.GetOk("unhealthy_node_replacement"); ok {
		unhealthyNodeReplacement = v.(bool)
	}

	instanceConfig := &awstypes.JobFlowInstancesConfig{
		KeepJobFlowAliveWhenNoSteps: aws.Bool(keepJobFlowAliveWhenNoSteps),
		TerminationProtected:        aws.Bool(terminationProtection),
		UnhealthyNodeReplacement:    aws.Bool(unhealthyNodeReplacement),
	}

	if l := d.Get("master_instance_group").([]any); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]any)

		instanceGroup := awstypes.InstanceGroupConfig{
			InstanceCount: aws.Int32(int32(m[names.AttrInstanceCount].(int))),
			InstanceRole:  awstypes.InstanceRoleTypeMaster,
			InstanceType:  aws.String(m[names.AttrInstanceType].(string)),
			Market:        awstypes.MarketTypeOnDemand,
			Name:          aws.String(m[names.AttrName].(string)),
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = awstypes.MarketTypeSpot
		}

		expandEBSConfig(m, &instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("core_instance_group").([]any); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]any)

		instanceGroup := awstypes.InstanceGroupConfig{
			InstanceCount: aws.Int32(int32(m[names.AttrInstanceCount].(int))),
			InstanceRole:  awstypes.InstanceRoleTypeCore,
			InstanceType:  aws.String(m[names.AttrInstanceType].(string)),
			Market:        awstypes.MarketTypeOnDemand,
			Name:          aws.String(m[names.AttrName].(string)),
		}

		if v, ok := m["autoscaling_policy"]; ok && v.(string) != "" {
			var autoScalingPolicy awstypes.AutoScalingPolicy

			if err := tfjson.DecodeFromString(v.(string), &autoScalingPolicy); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			instanceGroup.AutoScalingPolicy = &autoScalingPolicy
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = awstypes.MarketTypeSpot
		}

		expandEBSConfig(m, &instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("master_instance_fleet").([]any); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := expandInstanceFleetConfig(l[0].(map[string]any), awstypes.InstanceFleetTypeMaster)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, *instanceFleetConfig)
	}

	if l := d.Get("core_instance_fleet").([]any); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := expandInstanceFleetConfig(l[0].(map[string]any), awstypes.InstanceFleetTypeCore)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, *instanceFleetConfig)
	}

	var instanceProfile string
	if a, ok := d.GetOk("ec2_attributes"); ok {
		ec2Attributes := a.([]any)
		attributes := ec2Attributes[0].(map[string]any)

		if v, ok := attributes["key_name"]; ok {
			instanceConfig.Ec2KeyName = aws.String(v.(string))
		}
		if v, ok := attributes[names.AttrSubnetID]; ok {
			instanceConfig.Ec2SubnetId = aws.String(v.(string))
		}
		if v, ok := attributes[names.AttrSubnetIDs]; ok {
			instanceConfig.Ec2SubnetIds = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := attributes["additional_master_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalMasterSecurityGroups = strSlice
		}

		if v, ok := attributes["additional_slave_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalSlaveSecurityGroups = strSlice
		}

		if v, ok := attributes["emr_managed_master_security_group"]; ok {
			instanceConfig.EmrManagedMasterSecurityGroup = aws.String(v.(string))
		}
		if v, ok := attributes["emr_managed_slave_security_group"]; ok {
			instanceConfig.EmrManagedSlaveSecurityGroup = aws.String(v.(string))
		}

		if len(strings.TrimSpace(attributes["instance_profile"].(string))) != 0 {
			instanceProfile = strings.TrimSpace(attributes["instance_profile"].(string))
		}

		if v, ok := attributes["service_access_security_group"]; ok {
			instanceConfig.ServiceAccessSecurityGroup = aws.String(v.(string))
		}
	}

	name := d.Get(names.AttrName).(string)
	input := &emr.RunJobFlowInput{
		Instances:    instanceConfig,
		Name:         aws.String(name),
		Applications: expandApplications(applications),

		ReleaseLabel:      aws.String(d.Get("release_label").(string)),
		ServiceRole:       aws.String(d.Get(names.AttrServiceRole).(string)),
		VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("additional_info"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.AdditionalInfo = aws.String(v)
	}

	if v, ok := d.GetOk("log_encryption_kms_key_id"); ok {
		input.LogEncryptionKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_uri"); ok {
		input.LogUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("autoscaling_role"); ok {
		input.AutoScalingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scale_down_behavior"); ok {
		input.ScaleDownBehavior = awstypes.ScaleDownBehavior(v.(string))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ebs_root_volume_size"); ok {
		input.EbsRootVolumeSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("custom_ami_id"); ok {
		input.CustomAmiId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("step_concurrency_level"); ok {
		input.StepConcurrencyLevel = aws.Int32(int32(v.(int)))
	}

	if instanceProfile != "" {
		input.JobFlowRole = aws.String(instanceProfile)
	}

	if v, ok := d.GetOk("bootstrap_action"); ok {
		input.BootstrapActions = expandBootstrapActions(v.([]any))
	}
	if v, ok := d.GetOk("step"); ok {
		input.Steps = expandStepConfigs(v.([]any))
	}
	if v, ok := d.GetOk("configurations"); ok {
		input.Configurations = expandConfigures(v.(string))
	}

	if v, ok := d.GetOk("configurations_json"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.Configurations, err = expandConfigurationJSON(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v, ok := d.GetOk("kerberos_attributes"); ok {
		input.KerberosAttributes = expandKerberosAttributes(v.([]any)[0].(map[string]any))
	}
	if v, ok := d.GetOk("auto_termination_policy"); ok && len(v.([]any)) > 0 {
		input.AutoTerminationPolicy = expandAutoTerminationPolicy(v.([]any))
	}

	if v, ok := d.GetOk("placement_group_config"); ok {
		input.PlacementGroupConfigs = expandPlacementGroupConfigs(v.([]any))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.RunJobFlow(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, errCodeValidationException, "Invalid InstanceProfile:") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, errCodeAccessDeniedException, "Failed to authorize instance profile") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "running EMR Job Flow (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*emr.RunJobFlowOutput).JobFlowId))
	// This value can only be obtained through a deprecated function
	d.Set("keep_job_flow_alive_when_no_steps", input.Instances.KeepJobFlowAliveWhenNoSteps)

	cluster, err := waitClusterCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) create: %s", d.Id(), err)
	}

	// For multiple master nodes, EMR automatically enables
	// termination protection and ignores the configuration at launch.
	// This additional handling is to potentially disable termination
	// protection to match the desired Terraform configuration.
	if aws.ToBool(cluster.TerminationProtected) != terminationProtection {
		input := &emr.SetTerminationProtectionInput{
			JobFlowIds:           []string{d.Id()},
			TerminationProtected: aws.Bool(terminationProtection),
		}

		_, err := conn.SetTerminationProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EMR Cluster (%s) termination protection to match configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	cluster, err := findClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Cluster (%s): %s", d.Id(), err)
	}

	d.Set("cluster_state", cluster.Status.State)
	d.Set(names.AttrARN, cluster.ClusterArn)

	instanceGroups, err := findInstanceGroupsByClusterID(ctx, conn, d.Id())

	if err == nil { // find instance group
		coreGroup, _ := coreInstanceGroup(instanceGroups)
		masterGroup, _ := masterInstanceGroup(instanceGroups)

		flattenedCoreInstanceGroup, err := flattenCoreInstanceGroup(coreGroup)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if err := d.Set("core_instance_group", flattenedCoreInstanceGroup); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting core_instance_group: %s", err)
		}

		if err := d.Set("master_instance_group", flattenMasterInstanceGroup(masterGroup)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_instance_group: %s", err)
		}
	}

	instanceFleets, err := findInstanceFleetsByClusterID(ctx, conn, d.Id())

	if err == nil { // find instance fleets
		coreFleet, _ := instanceFleetForRole(instanceFleets, awstypes.InstanceFleetTypeCore)
		masterFleet, _ := instanceFleetForRole(instanceFleets, awstypes.InstanceFleetTypeMaster)

		flattenedCoreInstanceFleet := flattenInstanceFleet(coreFleet)
		if err := d.Set("core_instance_fleet", flattenedCoreInstanceFleet); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting core_instance_fleet: %s", err)
		}

		flattenedMasterInstanceFleet := flattenInstanceFleet(masterFleet)
		if err := d.Set("master_instance_fleet", flattenedMasterInstanceFleet); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_instance_fleet: %s", err)
		}
	}

	setTagsOut(ctx, cluster.Tags)

	d.Set(names.AttrName, cluster.Name)

	d.Set(names.AttrServiceRole, cluster.ServiceRole)
	d.Set("security_configuration", cluster.SecurityConfiguration)
	d.Set("autoscaling_role", cluster.AutoScalingRole)
	d.Set("release_label", cluster.ReleaseLabel)
	d.Set("log_encryption_kms_key_id", cluster.LogEncryptionKmsKeyId)
	d.Set("log_uri", cluster.LogUri)
	d.Set("master_public_dns", cluster.MasterPublicDnsName)
	d.Set("visible_to_all_users", cluster.VisibleToAllUsers)
	d.Set("ebs_root_volume_size", cluster.EbsRootVolumeSize)
	d.Set("scale_down_behavior", cluster.ScaleDownBehavior)
	d.Set("termination_protection", cluster.TerminationProtected)
	d.Set("unhealthy_node_replacement", cluster.UnhealthyNodeReplacement)
	d.Set("step_concurrency_level", cluster.StepConcurrencyLevel)

	d.Set("custom_ami_id", cluster.CustomAmiId)

	if err := d.Set("applications", flattenApplications(cluster.Applications)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting applications: %s", err)
	}

	if _, ok := d.GetOk("configurations_json"); ok {
		configOut, err := flattenConfigurationJSON(cluster.Configurations)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		if err := d.Set("configurations_json", configOut); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configurations_json: %s", err)
		}
	}

	if err := d.Set("ec2_attributes", flattenEC2InstanceAttributes(cluster.Ec2InstanceAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ec2_attributes: %s", err)
	}

	if err := d.Set("kerberos_attributes", flattenKerberosAttributes(d, cluster.KerberosAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kerberos_attributes: %s", err)
	}

	bootstrapActions, err := findBootstrapActionsByClusterID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Cluster (%s) bootstrap actions: %s", d.Id(), err)
	}

	if err := d.Set("bootstrap_action", flattenBootstrapArguments(bootstrapActions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting bootstrap_action: %s", err)
	}

	input := &emr.ListStepsInput{
		ClusterId: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("list_steps_states"); ok && v.(*schema.Set).Len() > 0 {
		input.StepStates = flex.ExpandStringyValueSet[awstypes.StepState](v.(*schema.Set))
	}

	stepSummaries, err := findStepSummaries(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EMR Cluster (%s) step summaries: %s", d.Id(), err)
	}

	if err := d.Set("step", flattenStepSummaries(stepSummaries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step: %s", err)
	}

	// AWS provides no other way to read back the additional_info
	if v, ok := d.GetOk("additional_info"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("additional_info", v)
	}

	autoTerminationPolicy, err := findAutoTerminationPolicyByClusterID(ctx, conn, d.Id())
	switch {
	case tfresource.NotFound(err):
		d.Set("auto_termination_policy", nil)
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading EMR Cluster (%s) auto-termination policy: %s", d.Id(), err)
	default:
		if err := d.Set("auto_termination_policy", flattenAutoTerminationPolicy(autoTerminationPolicy)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_termination_policy: %s", err)
		}
	}

	if err := d.Set("placement_group_config", flattenPlacementGroupConfigs(cluster.PlacementGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_group_config: %s", err)
	}

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	if d.HasChange("visible_to_all_users") {
		input := &emr.SetVisibleToAllUsersInput{
			JobFlowIds:        []string{d.Id()},
			VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		}

		_, err := conn.SetVisibleToAllUsers(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting visibility: %s", d.Id(), err)
		}
	}

	if d.HasChange("auto_termination_policy") {
		_, n := d.GetChange("auto_termination_policy")
		if len(n.([]any)) > 0 {
			input := &emr.PutAutoTerminationPolicyInput{
				AutoTerminationPolicy: expandAutoTerminationPolicy(n.([]any)),
				ClusterId:             aws.String(d.Id()),
			}

			_, err := conn.PutAutoTerminationPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting auto termination policy: %s", d.Id(), err)
			}
		} else {
			input := &emr.RemoveAutoTerminationPolicyInput{
				ClusterId: aws.String(d.Id()),
			}

			_, err := conn.RemoveAutoTerminationPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing auto termination policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("termination_protection") {
		input := &emr.SetTerminationProtectionInput{
			JobFlowIds:           []string{d.Id()},
			TerminationProtected: aws.Bool(d.Get("termination_protection").(bool)),
		}

		_, err := conn.SetTerminationProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting termination protection: %s", d.Id(), err)
		}
	}

	if d.HasChange("unhealthy_node_replacement") {
		input := &emr.SetUnhealthyNodeReplacementInput{
			JobFlowIds:               []string{d.Id()},
			UnhealthyNodeReplacement: aws.Bool(d.Get("unhealthy_node_replacement").(bool)),
		}

		_, err := conn.SetUnhealthyNodeReplacement(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting unhealthy node replacement: %s", d.Id(), err)
		}
	}

	if d.HasChange("core_instance_group.0.autoscaling_policy") {
		autoscalingPolicyStr := d.Get("core_instance_group.0.autoscaling_policy").(string)
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		if autoscalingPolicyStr != "" {
			var autoScalingPolicy awstypes.AutoScalingPolicy

			if err := tfjson.DecodeFromString(autoscalingPolicyStr, &autoScalingPolicy); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input := &emr.PutAutoScalingPolicyInput{
				ClusterId:         aws.String(d.Id()),
				AutoScalingPolicy: &autoScalingPolicy,
				InstanceGroupId:   aws.String(instanceGroupID),
			}

			_, err := conn.PutAutoScalingPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting autoscaling policy: %s", d.Id(), err)
			}
		} else {
			input := &emr.RemoveAutoScalingPolicyInput{
				ClusterId:       aws.String(d.Id()),
				InstanceGroupId: aws.String(instanceGroupID),
			}

			_, err := conn.RemoveAutoScalingPolicy(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing autoscaling policy: %s", d.Id(), err)
			}

			// RemoveAutoScalingPolicy seems to have eventual consistency.
			// Retry reading Instance Group configuration until the policy is removed.
			const (
				timeout = 1 * time.Minute
			)
			_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (any, error) {
				return findCoreInstanceGroupAutoScalingPolicy(ctx, conn, d.Id())
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing autoscaling policy: waiting for completion: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("core_instance_group.0.instance_count") {
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		input := &emr.ModifyInstanceGroupsInput{
			InstanceGroups: []awstypes.InstanceGroupModifyConfig{
				{
					InstanceGroupId: aws.String(instanceGroupID),
					InstanceCount:   aws.Int32(int32(d.Get("core_instance_group.0.instance_count").(int))),
				},
			},
		}

		_, err := conn.ModifyInstanceGroups(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EMR Cluster (%s) Instance Group (%s): %s", d.Id(), instanceGroupID, err)
		}

		const (
			timeout = 20 * time.Minute
		)
		if _, err := waitInstanceGroupRunning(ctx, conn, d.Id(), instanceGroupID, timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) Instance Group (%s) modification: %s", d.Id(), instanceGroupID, err)
		}
	}

	if d.HasChange("instance_group") {
		o, n := d.GetChange("instance_group")
		oSet := o.(*schema.Set).List()
		nSet := n.(*schema.Set).List()
		for _, currInstanceGroup := range oSet {
			for _, nextInstanceGroup := range nSet {
				oInstanceGroup := currInstanceGroup.(map[string]any)
				nInstanceGroup := nextInstanceGroup.(map[string]any)

				if oInstanceGroup["instance_role"].(string) != nInstanceGroup["instance_role"].(string) || oInstanceGroup[names.AttrName].(string) != nInstanceGroup[names.AttrName].(string) {
					continue
				}

				// Prevent duplicate PutAutoScalingPolicy from earlier update logic
				if nInstanceGroup[names.AttrID] == d.Get("core_instance_group.0.id").(string) && d.HasChange("core_instance_group.0.autoscaling_policy") {
					continue
				}

				if v, ok := nInstanceGroup["autoscaling_policy"]; ok && v.(string) != "" {
					var autoScalingPolicy awstypes.AutoScalingPolicy

					if err := tfjson.DecodeFromString(v.(string), &autoScalingPolicy); err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}

					input := &emr.PutAutoScalingPolicyInput{
						ClusterId:         aws.String(d.Id()),
						AutoScalingPolicy: &autoScalingPolicy,
						InstanceGroupId:   aws.String(oInstanceGroup[names.AttrID].(string)),
					}

					_, err := conn.PutAutoScalingPolicy(ctx, input)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "updating autoscaling policy for instance group %q: %s", oInstanceGroup[names.AttrID].(string), err)
					}

					break
				}
			}
		}
	}

	if d.HasChange("step_concurrency_level") {
		input := &emr.ModifyClusterInput{
			ClusterId:            aws.String(d.Id()),
			StepConcurrencyLevel: aws.Int32(int32(d.Get("step_concurrency_level").(int))),
		}

		_, err := conn.ModifyCluster(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): updating step concurrency level: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Printf("[DEBUG] Deleting EMR Cluster: %s", d.Id())
	_, err := conn.TerminateJobFlows(ctx, &emr.TerminateJobFlowsInput{
		JobFlowIds: []string{d.Id()},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "terminating EMR Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findClusterByID(ctx context.Context, conn *emr.Client, id string) (*awstypes.Cluster, error) {
	input := &emr.DescribeClusterInput{
		ClusterId: aws.String(id),
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.Id) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	if output.Status.State == awstypes.ClusterStateTerminated || output.Status.State == awstypes.ClusterStateTerminatedWithErrors {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status.State),
			LastRequest: input,
		}
	}

	return output, nil
}

func findCluster(ctx context.Context, conn *emr.Client, input *emr.DescribeClusterInput) (*awstypes.Cluster, error) {
	output, err := conn.DescribeCluster(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeClusterNotFound) || errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil || output.Cluster.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Cluster, nil
}

func statusCluster(ctx context.Context, conn *emr.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &emr.DescribeClusterInput{
			ClusterId: aws.String(id),
		}
		output, err := findCluster(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}

func waitClusterCreated(ctx context.Context, conn *emr.Client, id string) (*awstypes.Cluster, error) {
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ClusterStateBootstrapping, awstypes.ClusterStateStarting),
		Target:     enum.Slice(awstypes.ClusterStateRunning, awstypes.ClusterStateWaiting),
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", stateChangeReason.Code, aws.ToString(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *emr.Client, id string) (*awstypes.Cluster, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ClusterStateTerminating),
		Target:     enum.Slice(awstypes.ClusterStateTerminated, awstypes.ClusterStateTerminatedWithErrors),
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", stateChangeReason.Code, aws.ToString(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func findBootstrapActionsByClusterID(ctx context.Context, conn *emr.Client, id string) ([]awstypes.Command, error) {
	input := &emr.ListBootstrapActionsInput{
		ClusterId: aws.String(id),
	}

	return findBootstrapActions(ctx, conn, input)
}

func findBootstrapActions(ctx context.Context, conn *emr.Client, input *emr.ListBootstrapActionsInput) ([]awstypes.Command, error) {
	var output []awstypes.Command

	pages := emr.NewListBootstrapActionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.BootstrapActions...)
	}

	return output, nil
}

func findStepSummaries(ctx context.Context, conn *emr.Client, input *emr.ListStepsInput) ([]awstypes.StepSummary, error) {
	var output []awstypes.StepSummary

	pages := emr.NewListStepsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Steps...)
	}

	// ListSteps returns steps in reverse order (newest first).
	slices.Reverse(output)

	return output, nil
}

func findAutoTerminationPolicyByClusterID(ctx context.Context, conn *emr.Client, id string) (*awstypes.AutoTerminationPolicy, error) {
	input := &emr.GetAutoTerminationPolicyInput{
		ClusterId: aws.String(id),
	}

	return findAutoTerminationPolicy(ctx, conn, input)
}

func findAutoTerminationPolicy(ctx context.Context, conn *emr.Client, input *emr.GetAutoTerminationPolicyInput) (*awstypes.AutoTerminationPolicy, error) {
	output, err := conn.GetAutoTerminationPolicy(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") ||
		tfawserr.ErrMessageContains(err, errCodeUnknownOperationException, "Could not find operation GetAutoTerminationPolicy") ||
		tfawserr.ErrMessageContains(err, errCodeValidationException, "Auto-termination is not available for this account when using this release of EMR") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AutoTerminationPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AutoTerminationPolicy, nil
}

func expandApplications(tfList []any) []awstypes.Application {
	apiObjects := make([]awstypes.Application, 0, len(tfList))

	for _, v := range flex.ExpandStringList(tfList) {
		apiObjects = append(apiObjects, awstypes.Application{
			Name: v,
		})
	}

	return apiObjects
}

func flattenApplications(apiObjects []awstypes.Application) []any {
	return tfslices.ApplyToAll(apiObjects, func(app awstypes.Application) any {
		return aws.ToString(app.Name)
	})
}

func flattenEC2InstanceAttributes(apiObject *awstypes.Ec2InstanceAttributes) []any {
	tfList := make([]any, 0)
	tfMap := map[string]any{}

	if apiObject.Ec2KeyName != nil {
		tfMap["key_name"] = aws.ToString(apiObject.Ec2KeyName)
	}
	if apiObject.Ec2SubnetId != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(apiObject.Ec2SubnetId)
	}
	if len(apiObject.RequestedEc2SubnetIds) > 0 {
		tfMap[names.AttrSubnetIDs] = flex.FlattenStringValueSet(apiObject.RequestedEc2SubnetIds)
	}
	if apiObject.IamInstanceProfile != nil {
		tfMap["instance_profile"] = aws.ToString(apiObject.IamInstanceProfile)
	}
	if apiObject.EmrManagedMasterSecurityGroup != nil {
		tfMap["emr_managed_master_security_group"] = aws.ToString(apiObject.EmrManagedMasterSecurityGroup)
	}
	if apiObject.EmrManagedSlaveSecurityGroup != nil {
		tfMap["emr_managed_slave_security_group"] = aws.ToString(apiObject.EmrManagedSlaveSecurityGroup)
	}

	if len(apiObject.AdditionalMasterSecurityGroups) > 0 {
		tfMap["additional_master_security_groups"] = strings.Join(apiObject.AdditionalMasterSecurityGroups, ",")
	}
	if len(apiObject.AdditionalSlaveSecurityGroups) > 0 {
		tfMap["additional_slave_security_groups"] = strings.Join(apiObject.AdditionalSlaveSecurityGroups, ",")
	}

	if apiObject.ServiceAccessSecurityGroup != nil {
		tfMap["service_access_security_group"] = aws.ToString(apiObject.ServiceAccessSecurityGroup)
	}

	tfList = append(tfList, tfMap)

	return tfList
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeAutoScalingPolicy github.com/aws/aws-sdk-go-v2/service/emr.awsAwsjson11_serializeDocumentAutoScalingPolicy
func serializeAutoScalingPolicy(v *awstypes.AutoScalingPolicy, value smithyjson.Value) error

func flattenAutoScalingPolicyDescription(apiObject *awstypes.AutoScalingPolicyDescription) (string, error) {
	if apiObject == nil {
		return "", nil
	}

	for i, rule := range apiObject.Rules {
		dimensions := rule.Trigger.CloudWatchAlarmDefinition.Dimensions
		dimensions = slices.DeleteFunc(dimensions, func(v awstypes.MetricDimension) bool {
			return aws.ToString(v.Key) == "JobFlowId"
		})
		if len(dimensions) == 0 {
			dimensions = nil
		}
		apiObject.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = dimensions
	}

	autoScalingPolicy := &awstypes.AutoScalingPolicy{
		Constraints: apiObject.Constraints,
		Rules:       apiObject.Rules,
	}
	jsonEncoder := smithyjson.NewEncoder()
	err := serializeAutoScalingPolicy(autoScalingPolicy, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}

func flattenCoreInstanceGroup(apiObject *awstypes.InstanceGroup) ([]any, error) {
	if apiObject == nil {
		return []any{}, nil
	}

	autoscalingPolicy, err := flattenAutoScalingPolicyDescription(apiObject.AutoScalingPolicy)
	if err != nil {
		return nil, err
	}

	tfMap := map[string]any{
		"autoscaling_policy":    autoscalingPolicy,
		"bid_price":             aws.ToString(apiObject.BidPrice),
		"ebs_config":            flattenEBSConfig(apiObject.EbsBlockDevices),
		names.AttrID:            aws.ToString(apiObject.Id),
		names.AttrInstanceCount: aws.ToInt32(apiObject.RequestedInstanceCount),
		names.AttrInstanceType:  aws.ToString(apiObject.InstanceType),
		names.AttrName:          aws.ToString(apiObject.Name),
	}

	return []any{tfMap}, nil
}

func flattenMasterInstanceGroup(apiObject *awstypes.InstanceGroup) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"bid_price":             aws.ToString(apiObject.BidPrice),
		"ebs_config":            flattenEBSConfig(apiObject.EbsBlockDevices),
		names.AttrID:            aws.ToString(apiObject.Id),
		names.AttrInstanceCount: aws.ToInt32(apiObject.RequestedInstanceCount),
		names.AttrInstanceType:  aws.ToString(apiObject.InstanceType),
		names.AttrName:          aws.ToString(apiObject.Name),
	}

	return []any{tfMap}
}

func flattenKerberosAttributes(d *schema.ResourceData, apiObject *awstypes.KerberosAttributes) []any {
	tfList := make([]any, 0)

	if apiObject == nil || apiObject.Realm == nil {
		return tfList
	}

	// Do not set from API:
	// * ad_domain_join_password
	// * ad_domain_join_user
	// * cross_realm_trust_principal_password
	// * kdc_admin_password

	tfMap := map[string]any{
		"kdc_admin_password": d.Get("kerberos_attributes.0.kdc_admin_password").(string),
		"realm":              aws.ToString(apiObject.Realm),
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_password"); ok {
		tfMap["ad_domain_join_password"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_user"); ok {
		tfMap["ad_domain_join_user"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.cross_realm_trust_principal_password"); ok {
		tfMap["cross_realm_trust_principal_password"] = v.(string)
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenHadoopStepConfig(apiObject *awstypes.HadoopStepConfig) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"args":               apiObject.Args,
		"jar":                aws.ToString(apiObject.Jar),
		"main_class":         aws.ToString(apiObject.MainClass),
		names.AttrProperties: apiObject.Properties,
	}

	return tfMap
}

func flattenStepSummaries(apiObjects []awstypes.StepSummary) []any {
	tfList := make([]any, 0)

	if len(apiObjects) == 0 {
		return tfList
	}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenStepSummary(&apiObject))
	}

	return tfList
}

func flattenStepSummary(apiObject *awstypes.StepSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"action_on_failure": apiObject.ActionOnFailure,
		"hadoop_jar_step":   []map[string]any{flattenHadoopStepConfig(apiObject.Config)},
		names.AttrName:      aws.ToString(apiObject.Name),
	}

	return tfMap
}

func flattenEBSConfig(apiObjects []awstypes.EbsBlockDevice) *schema.Set {
	uniqueEBS := make(map[int]int)
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.VolumeSpecification.Iops != nil {
			tfMap[names.AttrIOPS] = int(aws.ToInt32(apiObject.VolumeSpecification.Iops))
		}
		if apiObject.VolumeSpecification.SizeInGB != nil {
			tfMap[names.AttrSize] = int(aws.ToInt32(apiObject.VolumeSpecification.SizeInGB))
		}
		if apiObject.VolumeSpecification.Throughput != nil {
			tfMap[names.AttrThroughput] = aws.ToInt32(apiObject.VolumeSpecification.Throughput)
		}
		if apiObject.VolumeSpecification.VolumeType != nil {
			tfMap[names.AttrType] = aws.ToString(apiObject.VolumeSpecification.VolumeType)
		}
		tfMap["volumes_per_instance"] = 1

		uniqueEBS[resourceClusterEBSHashConfig(tfMap)] += 1

		tfList = append(tfList, tfMap)
	}

	for _, tfMapRaw := range tfList {
		tfMapRaw.(map[string]any)["volumes_per_instance"] = uniqueEBS[resourceClusterEBSHashConfig(tfMapRaw)]
	}

	return schema.NewSet(resourceClusterEBSHashConfig, tfList)
}

func flattenBootstrapArguments(apiObjects []awstypes.Command) []any {
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		tfMap[names.AttrPath] = aws.ToString(apiObject.ScriptPath)
		tfMap["args"] = apiObject.Args

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func coreInstanceGroup(grps []awstypes.InstanceGroup) (*awstypes.InstanceGroup, error) {
	return tfresource.AssertSingleValueResult(tfslices.Filter(grps, func(v awstypes.InstanceGroup) bool {
		return v.InstanceGroupType == awstypes.InstanceGroupTypeCore
	}))
}

func expandBootstrapActions(tfList []any) []awstypes.BootstrapActionConfig {
	apiObjects := []awstypes.BootstrapActionConfig{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.BootstrapActionConfig{
			Name: aws.String(tfMap[names.AttrName].(string)),
			ScriptBootstrapAction: &awstypes.ScriptBootstrapActionConfig{
				Path: aws.String(tfMap[names.AttrPath].(string)),
				Args: flex.ExpandStringValueListEmpty(tfMap["args"].([]any)),
			},
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandHadoopJarStepConfig(tfMap map[string]any) *awstypes.HadoopJarStepConfig {
	apiObject := &awstypes.HadoopJarStepConfig{
		Jar: aws.String(tfMap["jar"].(string)),
	}

	if v, ok := tfMap["args"]; ok {
		apiObject.Args = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap["main_class"]; ok {
		apiObject.MainClass = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrProperties]; ok {
		apiObject.Properties = expandKeyValues(v.(map[string]any))
	}

	return apiObject
}

func expandKeyValues(tfMap map[string]any) []awstypes.KeyValue {
	apiObjects := make([]awstypes.KeyValue, 0)

	for k, v := range tfMap {
		apiObject := awstypes.KeyValue{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKerberosAttributes(tfMap map[string]any) *awstypes.KerberosAttributes {
	apiObject := &awstypes.KerberosAttributes{
		KdcAdminPassword: aws.String(tfMap["kdc_admin_password"].(string)),
		Realm:            aws.String(tfMap["realm"].(string)),
	}

	if v, ok := tfMap["ad_domain_join_password"]; ok && v.(string) != "" {
		apiObject.ADDomainJoinPassword = aws.String(v.(string))
	}
	if v, ok := tfMap["ad_domain_join_user"]; ok && v.(string) != "" {
		apiObject.ADDomainJoinUser = aws.String(v.(string))
	}
	if v, ok := tfMap["cross_realm_trust_principal_password"]; ok && v.(string) != "" {
		apiObject.CrossRealmTrustPrincipalPassword = aws.String(v.(string))
	}

	return apiObject
}

func expandStepConfig(tfMap map[string]any) awstypes.StepConfig {
	apiObject := awstypes.StepConfig{
		ActionOnFailure: awstypes.ActionOnFailure(tfMap["action_on_failure"].(string)),
		HadoopJarStep:   expandHadoopJarStepConfig(tfMap["hadoop_jar_step"].([]any)[0].(map[string]any)),
		Name:            aws.String(tfMap[names.AttrName].(string)),
	}

	return apiObject
}

func expandStepConfigs(tfList []any) []awstypes.StepConfig {
	apiObjects := []awstypes.StepConfig{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, expandStepConfig(tfMap))
	}

	return apiObjects
}

func expandEBSConfig(tfMap map[string]any, apiObject *awstypes.InstanceGroupConfig) {
	if v, ok := tfMap["ebs_config"]; ok {
		ebsConfig := &awstypes.EbsConfiguration{}
		ebsBlockDeviceConfigs := make([]awstypes.EbsBlockDeviceConfig, 0)

		for _, v := range v.(*schema.Set).List() {
			tfMap := v.(map[string]any)
			ebsBlockDeviceConfig := awstypes.EbsBlockDeviceConfig{
				VolumesPerInstance: aws.Int32(int32(tfMap["volumes_per_instance"].(int))),
				VolumeSpecification: &awstypes.VolumeSpecification{
					SizeInGB:   aws.Int32(int32(tfMap[names.AttrSize].(int))),
					VolumeType: aws.String(tfMap[names.AttrType].(string)),
				},
			}

			if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
				ebsBlockDeviceConfig.VolumeSpecification.Throughput = aws.Int32(int32(v))
			}
			if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
				ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int32(int32(v))
			}

			ebsBlockDeviceConfigs = append(ebsBlockDeviceConfigs, ebsBlockDeviceConfig)
		}

		ebsConfig.EbsBlockDeviceConfigs = ebsBlockDeviceConfigs

		apiObject.EbsConfiguration = ebsConfig
	}
}

func expandConfigurationJSON(tfString string) ([]awstypes.Configuration, error) {
	apiObjects := []awstypes.Configuration{}

	if err := tfjson.DecodeFromString(tfString, &apiObjects); err != nil {
		return nil, err
	}

	return apiObjects, nil
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeConfigurations github.com/aws/aws-sdk-go-v2/service/emr.awsAwsjson11_serializeDocumentConfigurationList
func serializeConfigurations(v []awstypes.Configuration, value smithyjson.Value) error

func flattenConfigurationJSON(apiObjects []awstypes.Configuration) (string, error) {
	jsonEncoder := smithyjson.NewEncoder()
	err := serializeConfigurations(apiObjects, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}

func expandConfigures(tfString string) []awstypes.Configuration {
	apiObjects := []awstypes.Configuration{}

	if strings.HasPrefix(tfString, "http") {
		if err := readHTTPJSON(tfString, &apiObjects); err != nil {
			log.Printf("[ERR] Error reading HTTP JSON: %s", err)
		}
	} else if strings.HasSuffix(tfString, ".json") {
		if err := readLocalJSON(tfString, &apiObjects); err != nil {
			log.Printf("[ERR] Error reading local JSON: %s", err)
		}
	} else {
		if err := readBodyJSON(tfString, &apiObjects); err != nil {
			log.Printf("[ERR] Error reading body JSON: %s", err)
		}
	}

	return apiObjects
}

func readHTTPJSON(url string, target any) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return tfjson.DecodeFromReader(r.Body, target)
}

func readLocalJSON(localFile string, target any) error {
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return tfjson.DecodeFromReader(file, target)
}

func readBodyJSON(body string, target any) error {
	return tfjson.DecodeFromString(body, target)
}

func masterInstanceGroup(instanceGroups []awstypes.InstanceGroup) (*awstypes.InstanceGroup, error) {
	return tfresource.AssertSingleValueResult(tfslices.Filter(instanceGroups, func(v awstypes.InstanceGroup) bool {
		return v.InstanceGroupType == awstypes.InstanceGroupTypeMaster
	}))
}

var resourceClusterEBSHashConfig = sdkv2.SimpleSchemaSetFunc(
	names.AttrSize,
	names.AttrType,
	"volumes_per_instance",
	names.AttrThroughput,
	names.AttrIOPS,
)

func findCoreInstanceGroupAutoScalingPolicy(ctx context.Context, conn *emr.Client, clusterID string) (*awstypes.AutoScalingPolicyDescription, error) {
	instanceGroups, err := findInstanceGroupsByClusterID(ctx, conn, clusterID)

	if err != nil {
		return nil, err
	}

	instanceGroup, err := coreInstanceGroup(instanceGroups)

	if err != nil {
		return nil, err
	}

	if instanceGroup.AutoScalingPolicy == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return instanceGroup.AutoScalingPolicy, nil
}

func findInstanceGroupsByClusterID(ctx context.Context, conn *emr.Client, clusterID string) ([]awstypes.InstanceGroup, error) {
	input := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterID),
	}

	return findInstanceGroups(ctx, conn, input, tfslices.PredicateTrue[*awstypes.InstanceGroup]())
}

func expandInstanceFleetConfig(tfMap map[string]any, instanceFleetType awstypes.InstanceFleetType) *awstypes.InstanceFleetConfig {
	apiObject := &awstypes.InstanceFleetConfig{
		InstanceFleetType:      instanceFleetType,
		Name:                   aws.String(tfMap[names.AttrName].(string)),
		TargetOnDemandCapacity: aws.Int32(int32(tfMap["target_on_demand_capacity"].(int))),
		TargetSpotCapacity:     aws.Int32(int32(tfMap["target_spot_capacity"].(int))),
	}

	if v, ok := tfMap["instance_type_configs"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := tfMap["launch_specifications"].([]any); ok && len(v) == 1 && v[0] != nil {
		apiObject.LaunchSpecifications = expandLaunchSpecification(v[0].(map[string]any))
	}

	return apiObject
}

func findInstanceFleetsByClusterID(ctx context.Context, conn *emr.Client, clusterID string) ([]awstypes.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}

	return findInstanceFleets(ctx, conn, input, tfslices.PredicateTrue[*awstypes.InstanceFleet]())
}

func instanceFleetForRole(instanceFleets []awstypes.InstanceFleet, instanceRoleType awstypes.InstanceFleetType) (*awstypes.InstanceFleet, error) {
	return tfresource.AssertSingleValueResult(tfslices.Filter(instanceFleets, func(v awstypes.InstanceFleet) bool {
		return v.InstanceFleetType == instanceRoleType
	}))
}

func flattenInstanceFleet(apiObject *awstypes.InstanceFleet) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrID:                     aws.ToString(apiObject.Id),
		names.AttrName:                   aws.ToString(apiObject.Name),
		"target_on_demand_capacity":      aws.ToInt32(apiObject.TargetOnDemandCapacity),
		"target_spot_capacity":           aws.ToInt32(apiObject.TargetSpotCapacity),
		"provisioned_on_demand_capacity": aws.ToInt32(apiObject.ProvisionedOnDemandCapacity),
		"provisioned_spot_capacity":      aws.ToInt32(apiObject.ProvisionedSpotCapacity),
		"instance_type_configs":          flattenInstanceTypeSpecifications(apiObject.InstanceTypeSpecifications),
		"launch_specifications":          flattenInstanceFleetProvisioningSpecifications(apiObject.LaunchSpecifications),
	}

	return []any{tfMap}
}

func flattenInstanceTypeSpecifications(apiObjects []awstypes.InstanceTypeSpecification) []any {
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.BidPrice != nil {
			tfMap["bid_price"] = aws.ToString(apiObject.BidPrice)
		}

		if apiObject.BidPriceAsPercentageOfOnDemandPrice != nil {
			tfMap["bid_price_as_percentage_of_on_demand_price"] = aws.ToFloat64(apiObject.BidPriceAsPercentageOfOnDemandPrice)
		}

		tfMap["ebs_config"] = flattenEBSConfig(apiObject.EbsBlockDevices)
		tfMap[names.AttrInstanceType] = aws.ToString(apiObject.InstanceType)
		tfMap["weighted_capacity"] = int(aws.ToInt32(apiObject.WeightedCapacity))

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenInstanceFleetProvisioningSpecifications(apiObject *awstypes.InstanceFleetProvisioningSpecifications) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"on_demand_specification": flattenOnDemandProvisioningSpecification(apiObject.OnDemandSpecification),
		"spot_specification":      flattenSpotProvisioningSpecification(apiObject.SpotSpecification),
	}

	return []any{tfMap}
}

func flattenOnDemandProvisioningSpecification(apiObject *awstypes.OnDemandProvisioningSpecification) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		// The return value from api is wrong. it return the value with uppercase letters and '_' vs. '-'
		// The value needs to be normalized to avoid perpetual difference in the Terraform plan
		"allocation_strategy": strings.Replace(strings.ToLower(string(apiObject.AllocationStrategy)), "_", "-", -1),
	}

	return []any{tfMap}
}

func flattenSpotProvisioningSpecification(apiObject *awstypes.SpotProvisioningSpecification) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"timeout_action":           apiObject.TimeoutAction,
		"timeout_duration_minutes": aws.ToInt32(apiObject.TimeoutDurationMinutes),
	}

	if apiObject.BlockDurationMinutes != nil {
		tfMap["block_duration_minutes"] = aws.ToInt32(apiObject.BlockDurationMinutes)
	}

	// The return value from api is wrong. it return the value with uppercase letters and '_' vs. '-'
	// The value needs to be normalized to avoid perpetual difference in the Terraform plan
	tfMap["allocation_strategy"] = strings.Replace(strings.ToLower(string(apiObject.AllocationStrategy)), "_", "-", -1)

	return []any{tfMap}
}

// TODO
func expandEBSConfiguration(ebsConfigurations []any) *awstypes.EbsConfiguration {
	ebsConfig := &awstypes.EbsConfiguration{}
	ebsConfigs := make([]awstypes.EbsBlockDeviceConfig, 0)
	for _, ebsConfiguration := range ebsConfigurations {
		cfg := ebsConfiguration.(map[string]any)
		ebsBlockDeviceConfig := awstypes.EbsBlockDeviceConfig{
			VolumesPerInstance: aws.Int32(int32(cfg["volumes_per_instance"].(int))),
			VolumeSpecification: &awstypes.VolumeSpecification{
				SizeInGB:   aws.Int32(int32(cfg[names.AttrSize].(int))),
				VolumeType: aws.String(cfg[names.AttrType].(string)),
			},
		}
		if v, ok := cfg[names.AttrThroughput].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Throughput = aws.Int32(int32(v))
		}
		if v, ok := cfg[names.AttrIOPS].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int32(int32(v))
		}
		ebsConfigs = append(ebsConfigs, ebsBlockDeviceConfig)
	}
	ebsConfig.EbsBlockDeviceConfigs = ebsConfigs
	return ebsConfig
}

func expandInstanceTypeConfigs(tfList []any) []awstypes.InstanceTypeConfig {
	apiObjects := []awstypes.InstanceTypeConfig{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.InstanceTypeConfig{
			InstanceType: aws.String(tfMap[names.AttrInstanceType].(string)),
		}

		if v, ok := tfMap["bid_price"]; ok && v != "" {
			apiObject.BidPrice = aws.String(v.(string))
		}

		if v, ok := tfMap["bid_price_as_percentage_of_on_demand_price"].(float64); ok && v != 0 {
			apiObject.BidPriceAsPercentageOfOnDemandPrice = aws.Float64(v)
		}

		if v, ok := tfMap["weighted_capacity"].(int); ok {
			apiObject.WeightedCapacity = aws.Int32(int32(v))
		}

		if v, ok := tfMap["configurations"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.Configurations = expandConfigurations(v.List())
		}

		if v, ok := tfMap["ebs_config"].(*schema.Set); ok && v.Len() == 1 {
			apiObject.EbsConfiguration = expandEBSConfiguration(v.List())
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchSpecification(tfMap map[string]any) *awstypes.InstanceFleetProvisioningSpecifications {
	apiObject := &awstypes.InstanceFleetProvisioningSpecifications{}

	if v := tfMap["on_demand_specification"].([]any); len(v) > 0 {
		apiObject.OnDemandSpecification = &awstypes.OnDemandProvisioningSpecification{
			AllocationStrategy: awstypes.OnDemandProvisioningAllocationStrategy(v[0].(map[string]any)["allocation_strategy"].(string)),
		}
	}

	if v := tfMap["spot_specification"].([]any); len(v) > 0 {
		tfMap := v[0].(map[string]any)
		spotProvisioning := &awstypes.SpotProvisioningSpecification{
			TimeoutAction:          awstypes.SpotProvisioningTimeoutAction(tfMap["timeout_action"].(string)),
			TimeoutDurationMinutes: aws.Int32(int32(tfMap["timeout_duration_minutes"].(int))),
		}
		if v, ok := tfMap["block_duration_minutes"]; ok && v != 0 {
			spotProvisioning.BlockDurationMinutes = aws.Int32(int32(v.(int)))
		}
		if v, ok := tfMap["allocation_strategy"]; ok {
			spotProvisioning.AllocationStrategy = awstypes.SpotProvisioningAllocationStrategy(v.(string))
		}

		apiObject.SpotSpecification = spotProvisioning
	}

	return apiObject
}

func expandConfigurations(tfList []any) []awstypes.Configuration {
	apiObjects := []awstypes.Configuration{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.Configuration{}

		if v, ok := tfMap["classification"].(string); ok {
			apiObject.Classification = aws.String(v)
		}

		if v, ok := tfMap["configurations"].([]any); ok {
			apiObject.Configurations = expandConfigurations(v)
		}

		if v, ok := tfMap[names.AttrProperties].(map[string]any); ok && len(v) > 0 {
			apiObject.Properties = flex.ExpandStringValueMap(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func resourceInstanceTypeHashConfig(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrInstanceType].(string)))
	if v, ok := m["bid_price"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["weighted_capacity"]; ok && v.(int) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["bid_price_as_percentage_of_on_demand_price"]; ok && v.(float64) != 0 {
		buf.WriteString(fmt.Sprintf("%f-", v.(float64)))
	}
	return create.StringHashcode(buf.String())
}

func expandAutoTerminationPolicy(tfList []any) *awstypes.AutoTerminationPolicy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.AutoTerminationPolicy{}

	if v, ok := tfMap["idle_timeout"].(int); ok && v > 0 {
		apiObject.IdleTimeout = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenAutoTerminationPolicy(apiObject *awstypes.AutoTerminationPolicy) []any {
	tfList := make([]any, 0)

	if apiObject == nil {
		return tfList
	}

	tfMap := map[string]any{}

	if apiObject.IdleTimeout != nil {
		tfMap["idle_timeout"] = aws.ToInt64(apiObject.IdleTimeout)
	}

	tfList = append(tfList, tfMap)

	return tfList
}

func expandPlacementGroupConfigs(tfList []any) []awstypes.PlacementGroupConfig {
	apiObjects := []awstypes.PlacementGroupConfig{}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.PlacementGroupConfig{
			InstanceRole: awstypes.InstanceRoleType(tfMap["instance_role"].(string)),
		}

		if v, ok := tfMap["placement_strategy"]; ok && v.(string) != "" {
			apiObject.PlacementStrategy = awstypes.PlacementGroupStrategy(v.(string))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPlacementGroupConfigs(apiObjects []awstypes.PlacementGroupConfig) []any {
	if apiObjects == nil {
		return []any{}
	}

	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		tfMap["instance_role"] = apiObject.InstanceRole
		tfMap["placement_strategy"] = apiObject.PlacementStrategy

		tfList = append(tfList, tfMap)
	}

	return tfList
}
