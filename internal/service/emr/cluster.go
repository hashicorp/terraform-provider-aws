// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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

		CustomizeDiff: verify.SetTagsDiff,

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
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(emr.OnDemandProvisioningAllocationStrategy_Values(), false),
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
													Type:         schema.TypeString,
													ForceNew:     true,
													Required:     true,
													ValidateFunc: validation.StringInSlice(emr.SpotProvisioningAllocationStrategy_Values(), false),
												},
												"block_duration_minutes": {
													Type:     schema.TypeInt,
													Optional: true,
													ForceNew: true,
													Default:  0,
												},
												"timeout_action": {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(emr.SpotProvisioningTimeoutAction_Values(), false),
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
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         true,
					ValidateFunc:     validation.StringIsJSON,
					DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
					StateFunc: func(v interface{}) string {
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
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         true,
					ValidateFunc:     validation.StringIsJSON,
					DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
					StateFunc: func(v interface{}) string {
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
								Type:             schema.TypeString,
								Optional:         true,
								DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
								ValidateFunc:     validation.StringIsJSON,
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
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice(emr.StepState_Values(), false),
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
								Type:         schema.TypeString,
								ForceNew:     true,
								Required:     true,
								ValidateFunc: validation.StringInSlice(emr.InstanceRoleType_Values(), false),
							},
							"placement_strategy": {
								Type:         schema.TypeString,
								ForceNew:     true,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringInSlice(emr.PlacementGroupStrategy_Values(), false),
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
					Type:         schema.TypeString,
					ForceNew:     true,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.StringInSlice(emr.ScaleDownBehavior_Values(), false),
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
								Type:         schema.TypeString,
								Required:     true,
								ForceNew:     true,
								ValidateFunc: validation.StringInSlice(emr.ActionOnFailure_Values(), false),
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

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

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

	instanceConfig := &emr.JobFlowInstancesConfig{
		KeepJobFlowAliveWhenNoSteps: aws.Bool(keepJobFlowAliveWhenNoSteps),
		TerminationProtected:        aws.Bool(terminationProtection),
		UnhealthyNodeReplacement:    aws.Bool(unhealthyNodeReplacement),
	}

	if l := d.Get("master_instance_group").([]interface{}); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]interface{})

		instanceGroup := &emr.InstanceGroupConfig{
			InstanceCount: aws.Int64(int64(m[names.AttrInstanceCount].(int))),
			InstanceRole:  aws.String(emr.InstanceRoleTypeMaster),
			InstanceType:  aws.String(m[names.AttrInstanceType].(string)),
			Market:        aws.String(emr.MarketTypeOnDemand),
			Name:          aws.String(m[names.AttrName].(string)),
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = aws.String(emr.MarketTypeSpot)
		}

		expandEBSConfig(m, instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("core_instance_group").([]interface{}); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]interface{})

		instanceGroup := &emr.InstanceGroupConfig{
			InstanceCount: aws.Int64(int64(m[names.AttrInstanceCount].(int))),
			InstanceRole:  aws.String(emr.InstanceRoleTypeCore),
			InstanceType:  aws.String(m[names.AttrInstanceType].(string)),
			Market:        aws.String(emr.MarketTypeOnDemand),
			Name:          aws.String(m[names.AttrName].(string)),
		}

		if v, ok := m["autoscaling_policy"]; ok && v.(string) != "" {
			var autoScalingPolicy *emr.AutoScalingPolicy

			if err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy); err != nil {
				return sdkdiag.AppendErrorf(diags, "parsing core_instance_group Auto Scaling Policy JSON: %s", err)
			}

			instanceGroup.AutoScalingPolicy = autoScalingPolicy
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = aws.String(emr.MarketTypeSpot)
		}

		expandEBSConfig(m, instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("master_instance_fleet").([]interface{}); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := readInstanceFleetConfig(l[0].(map[string]interface{}), emr.InstanceFleetTypeMaster)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, instanceFleetConfig)
	}

	if l := d.Get("core_instance_fleet").([]interface{}); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := readInstanceFleetConfig(l[0].(map[string]interface{}), emr.InstanceFleetTypeCore)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, instanceFleetConfig)
	}

	var instanceProfile string
	if a, ok := d.GetOk("ec2_attributes"); ok {
		ec2Attributes := a.([]interface{})
		attributes := ec2Attributes[0].(map[string]interface{})

		if v, ok := attributes["key_name"]; ok {
			instanceConfig.Ec2KeyName = aws.String(v.(string))
		}
		if v, ok := attributes[names.AttrSubnetID]; ok {
			instanceConfig.Ec2SubnetId = aws.String(v.(string))
		}
		if v, ok := attributes[names.AttrSubnetIDs]; ok {
			instanceConfig.Ec2SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := attributes["additional_master_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalMasterSecurityGroups = aws.StringSlice(strSlice)
		}

		if v, ok := attributes["additional_slave_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalSlaveSecurityGroups = aws.StringSlice(strSlice)
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

	emrApps := expandApplications(applications)

	params := &emr.RunJobFlowInput{
		Instances:    instanceConfig,
		Name:         aws.String(d.Get(names.AttrName).(string)),
		Applications: emrApps,

		ReleaseLabel:      aws.String(d.Get("release_label").(string)),
		ServiceRole:       aws.String(d.Get(names.AttrServiceRole).(string)),
		VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("additional_info"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Additional Info contains an invalid JSON: %v", err)
		}
		params.AdditionalInfo = aws.String(info)
	}

	if v, ok := d.GetOk("log_encryption_kms_key_id"); ok {
		params.LogEncryptionKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_uri"); ok {
		params.LogUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("autoscaling_role"); ok {
		params.AutoScalingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scale_down_behavior"); ok {
		params.ScaleDownBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		params.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ebs_root_volume_size"); ok {
		params.EbsRootVolumeSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("custom_ami_id"); ok {
		params.CustomAmiId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("step_concurrency_level"); ok {
		params.StepConcurrencyLevel = aws.Int64(int64(v.(int)))
	}

	if instanceProfile != "" {
		params.JobFlowRole = aws.String(instanceProfile)
	}

	if v, ok := d.GetOk("bootstrap_action"); ok {
		bootstrapActions := v.([]interface{})
		params.BootstrapActions = expandBootstrapActions(bootstrapActions)
	}
	if v, ok := d.GetOk("step"); ok {
		steps := v.([]interface{})
		params.Steps = expandStepConfigs(steps)
	}
	if v, ok := d.GetOk("configurations"); ok {
		confUrl := v.(string)
		params.Configurations = expandConfigures(confUrl)
	}

	if v, ok := d.GetOk("configurations_json"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "configurations_json contains an invalid JSON: %v", err)
		}
		params.Configurations, err = expandConfigurationJSON(info)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EMR configurations_json: %s", err)
		}
	}

	if v, ok := d.GetOk("kerberos_attributes"); ok {
		kerberosAttributesList := v.([]interface{})
		kerberosAttributesMap := kerberosAttributesList[0].(map[string]interface{})
		params.KerberosAttributes = expandKerberosAttributes(kerberosAttributesMap)
	}
	if v, ok := d.GetOk("auto_termination_policy"); ok && len(v.([]interface{})) > 0 {
		params.AutoTerminationPolicy = expandAutoTerminationPolicy(v.([]interface{}))
	}

	if v, ok := d.GetOk("placement_group_config"); ok {
		placementGroupConfigs := v.([]interface{})
		params.PlacementGroupConfigs = expandPlacementGroupConfigs(placementGroupConfigs)
	}

	var resp *emr.RunJobFlowOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		resp, err = conn.RunJobFlowWithContext(ctx, params)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "ValidationException", "Invalid InstanceProfile:") {
				return retry.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Failed to authorize instance profile") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.RunJobFlowWithContext(ctx, params)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "running EMR Job Flow: %s", err)
	}

	d.SetId(aws.StringValue(resp.JobFlowId))
	// This value can only be obtained through a deprecated function
	d.Set("keep_job_flow_alive_when_no_steps", params.Instances.KeepJobFlowAliveWhenNoSteps)

	log.Println("[INFO] Waiting for EMR Cluster to be available")
	cluster, err := waitClusterCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) to create: %s", d.Id(), err)
	}

	// For multiple master nodes, EMR automatically enables
	// termination protection and ignores the configuration at launch.
	// This additional handling is to potentially disable termination
	// protection to match the desired Terraform configuration.
	if aws.BoolValue(cluster.TerminationProtected) != terminationProtection {
		input := &emr.SetTerminationProtectionInput{
			JobFlowIds:           []*string{aws.String(d.Id())},
			TerminationProtected: aws.Bool(terminationProtection),
		}

		if _, err := conn.SetTerminationProtectionWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EMR Cluster (%s) termination protection to match configuration: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

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

	instanceGroups, err := fetchAllInstanceGroups(ctx, conn, d.Id())

	if err == nil { // find instance group
		coreGroup := coreInstanceGroup(instanceGroups)
		masterGroup := findMasterGroup(instanceGroups)

		flattenedCoreInstanceGroup, err := flattenCoreInstanceGroup(coreGroup)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "flattening core_instance_group: %s", err)
		}

		if err := d.Set("core_instance_group", flattenedCoreInstanceGroup); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting core_instance_group: %s", err)
		}

		if err := d.Set("master_instance_group", flattenMasterInstanceGroup(masterGroup)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting master_instance_group: %s", err)
		}
	}

	instanceFleets, err := FetchAllInstanceFleets(ctx, conn, d.Id())

	if err == nil { // find instance fleets
		coreFleet := findInstanceFleet(instanceFleets, emr.InstanceFleetTypeCore)
		masterFleet := findInstanceFleet(instanceFleets, emr.InstanceFleetTypeMaster)

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
		return sdkdiag.AppendErrorf(diags, "setting EMR Applications for cluster (%s): %s", d.Id(), err)
	}

	if _, ok := d.GetOk("configurations_json"); ok {
		configOut, err := flattenConfigurationJSON(cluster.Configurations)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EMR cluster configurations: %s", err)
		}
		if err := d.Set("configurations_json", configOut); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EMR configurations_json for cluster (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set("ec2_attributes", flattenEC2InstanceAttributes(cluster.Ec2InstanceAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EMR Ec2 Attributes: %s", err)
	}

	if err := d.Set("kerberos_attributes", flattenKerberosAttributes(d, cluster.KerberosAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kerberos_attributes: %s", err)
	}

	respBootstraps, err := conn.ListBootstrapActionsWithContext(ctx, &emr.ListBootstrapActionsInput{
		ClusterId: cluster.Id,
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EMR Cluster (%s) bootstrap actions: %s", d.Id(), err)
	}

	if err := d.Set("bootstrap_action", flattenBootstrapArguments(respBootstraps.BootstrapActions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Bootstrap Actions: %s", err)
	}

	var stepSummaries []*emr.StepSummary
	input := &emr.ListStepsInput{
		ClusterId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("list_steps_states"); ok && v.(*schema.Set).Len() > 0 {
		input.StepStates = flex.ExpandStringSet(v.(*schema.Set))
	}

	err = conn.ListStepsPagesWithContext(ctx, input, func(page *emr.ListStepsOutput, lastPage bool) bool {
		// ListSteps returns steps in reverse order (newest first).
		for _, step := range page.Steps {
			stepSummaries = append([]*emr.StepSummary{step}, stepSummaries...)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EMR Cluster (%s) steps: %s", d.Id(), err)
	}

	if err := d.Set("step", flattenStepSummaries(stepSummaries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting step: %s", err)
	}

	// AWS provides no other way to read back the additional_info
	if v, ok := d.GetOk("additional_info"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Additional Info contains an invalid JSON: %v", err)
		}
		d.Set("additional_info", info)
	}

	atpOut, err := conn.GetAutoTerminationPolicyWithContext(ctx, &emr.GetAutoTerminationPolicyInput{
		ClusterId: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Auto-termination is not available for this account when using this release of EMR") ||
			tfawserr.ErrMessageContains(err, ErrCodeUnknownOperationException, "Could not find operation GetAutoTerminationPolicy") {
			err = nil
		}
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Cluster (%s) auto-termination policy: %s", d.Id(), err)
	}

	if err := d.Set("auto_termination_policy", flattenAutoTerminationPolicy(atpOut.AutoTerminationPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting auto_termination_policy: %s", err)
	}

	if err := d.Set("placement_group_config", flattenPlacementGroupConfigs(cluster.PlacementGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_group_config: %s", err)
	}

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	if d.HasChange("visible_to_all_users") {
		_, err := conn.SetVisibleToAllUsersWithContext(ctx, &emr.SetVisibleToAllUsersInput{
			JobFlowIds:        []*string{aws.String(d.Id())},
			VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting visibility: %s", d.Id(), err)
		}
	}

	if d.HasChange("auto_termination_policy") {
		_, n := d.GetChange("auto_termination_policy")
		if len(n.([]interface{})) > 0 {
			log.Printf("[DEBUG] Putting EMR cluster Auto Termination Policy")

			_, err := conn.PutAutoTerminationPolicyWithContext(ctx, &emr.PutAutoTerminationPolicyInput{
				AutoTerminationPolicy: expandAutoTerminationPolicy(n.([]interface{})),
				ClusterId:             aws.String(d.Id()),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting auto termination policy: %s", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Removing EMR cluster Auto Termination Policy")

			_, err := conn.RemoveAutoTerminationPolicyWithContext(ctx, &emr.RemoveAutoTerminationPolicyInput{
				ClusterId: aws.String(d.Id()),
			})
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing auto termination policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("termination_protection") {
		_, err := conn.SetTerminationProtectionWithContext(ctx, &emr.SetTerminationProtectionInput{
			JobFlowIds:           []*string{aws.String(d.Id())},
			TerminationProtected: aws.Bool(d.Get("termination_protection").(bool)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting termination protection: %s", d.Id(), err)
		}
	}

	if d.HasChange("unhealthy_node_replacement") {
		_, err := conn.SetUnhealthyNodeReplacementWithContext(ctx, &emr.SetUnhealthyNodeReplacementInput{
			JobFlowIds:               []*string{aws.String(d.Id())},
			UnhealthyNodeReplacement: aws.Bool(d.Get("unhealthy_node_replacement").(bool)),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting unhealthy node replacement: %s", d.Id(), err)
		}
	}

	if d.HasChange("core_instance_group.0.autoscaling_policy") {
		autoscalingPolicyStr := d.Get("core_instance_group.0.autoscaling_policy").(string)
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		if autoscalingPolicyStr != "" {
			var autoScalingPolicy *emr.AutoScalingPolicy

			if err := json.Unmarshal([]byte(autoscalingPolicyStr), &autoScalingPolicy); err != nil {
				return sdkdiag.AppendErrorf(diags, "parsing core_instance_group Auto Scaling Policy JSON: %s", err)
			}

			input := &emr.PutAutoScalingPolicyInput{
				ClusterId:         aws.String(d.Id()),
				AutoScalingPolicy: autoScalingPolicy,
				InstanceGroupId:   aws.String(instanceGroupID),
			}

			if _, err := conn.PutAutoScalingPolicyWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): setting autoscaling policy: %s", d.Id(), err)
			}
		} else {
			input := &emr.RemoveAutoScalingPolicyInput{
				ClusterId:       aws.String(d.Id()),
				InstanceGroupId: aws.String(instanceGroupID),
			}

			if _, err := conn.RemoveAutoScalingPolicyWithContext(ctx, input); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing autoscaling policy: %s", d.Id(), err)
			}

			// RemoveAutoScalingPolicy seems to have eventual consistency.
			// Retry reading Instance Group configuration until the policy is removed.
			err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
				autoscalingPolicy, err := getCoreInstanceGroupAutoScalingPolicy(ctx, conn, d.Id())

				if err != nil {
					return retry.NonRetryableError(err)
				}

				if autoscalingPolicy != nil {
					return retry.RetryableError(errors.New("still exists"))
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				var autoscalingPolicy *emr.AutoScalingPolicyDescription

				autoscalingPolicy, err = getCoreInstanceGroupAutoScalingPolicy(ctx, conn, d.Id())

				if autoscalingPolicy != nil {
					err = errors.New("still exists")
				}
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): removing autoscaling policy: waiting for completion: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("core_instance_group.0.instance_count") {
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		input := &emr.ModifyInstanceGroupsInput{
			InstanceGroups: []*emr.InstanceGroupModifyConfig{
				{
					InstanceGroupId: aws.String(instanceGroupID),
					InstanceCount:   aws.Int64(int64(d.Get("core_instance_group.0.instance_count").(int))),
				},
			},
		}

		if _, err := conn.ModifyInstanceGroupsWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EMR Cluster (%s) Instance Group (%s): %s", d.Id(), instanceGroupID, err)
		}

		stateConf := &retry.StateChangeConf{
			Pending: []string{
				emr.InstanceGroupStateBootstrapping,
				emr.InstanceGroupStateProvisioning,
				emr.InstanceGroupStateReconfiguring,
				emr.InstanceGroupStateResizing,
			},
			Target:  []string{emr.InstanceGroupStateRunning},
			Refresh: instanceGroupStateRefresh(ctx, conn, d.Id(), instanceGroupID),
			Timeout: 20 * time.Minute,
			Delay:   10 * time.Second,
		}

		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) Instance Group (%s) modification: %s", d.Id(), instanceGroupID, err)
		}
	}

	if d.HasChange("instance_group") {
		o, n := d.GetChange("instance_group")
		oSet := o.(*schema.Set).List()
		nSet := n.(*schema.Set).List()
		for _, currInstanceGroup := range oSet {
			for _, nextInstanceGroup := range nSet {
				oInstanceGroup := currInstanceGroup.(map[string]interface{})
				nInstanceGroup := nextInstanceGroup.(map[string]interface{})

				if oInstanceGroup["instance_role"].(string) != nInstanceGroup["instance_role"].(string) || oInstanceGroup[names.AttrName].(string) != nInstanceGroup[names.AttrName].(string) {
					continue
				}

				// Prevent duplicate PutAutoScalingPolicy from earlier update logic
				if nInstanceGroup[names.AttrID] == d.Get("core_instance_group.0.id").(string) && d.HasChange("core_instance_group.0.autoscaling_policy") {
					continue
				}

				if v, ok := nInstanceGroup["autoscaling_policy"]; ok && v.(string) != "" {
					var autoScalingPolicy *emr.AutoScalingPolicy

					err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy)
					if err != nil {
						return sdkdiag.AppendErrorf(diags, "parsing EMR Auto Scaling Policy JSON for update: \n\n%s\n\n%s", v.(string), err)
					}

					putAutoScalingPolicy := &emr.PutAutoScalingPolicyInput{
						ClusterId:         aws.String(d.Id()),
						AutoScalingPolicy: autoScalingPolicy,
						InstanceGroupId:   aws.String(oInstanceGroup[names.AttrID].(string)),
					}

					_, errModify := conn.PutAutoScalingPolicyWithContext(ctx, putAutoScalingPolicy)
					if errModify != nil {
						return sdkdiag.AppendErrorf(diags, "updating autoscaling policy for instance group %q: %s", oInstanceGroup[names.AttrID].(string), errModify)
					}

					break
				}
			}
		}
	}

	if d.HasChange("step_concurrency_level") {
		_, err := conn.ModifyClusterWithContext(ctx, &emr.ModifyClusterInput{
			ClusterId:            aws.String(d.Id()),
			StepConcurrencyLevel: aws.Int64(int64(d.Get("step_concurrency_level").(int))),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Cluster (%s): updating step concurrency level: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	log.Printf("[DEBUG] Deleting EMR Cluster: %s", d.Id())
	_, err := conn.TerminateJobFlowsWithContext(ctx, &emr.TerminateJobFlowsInput{
		JobFlowIds: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "terminating EMR Cluster (%s): %s", d.Id(), err)
	}

	log.Println("[INFO] Waiting for EMR Cluster to be terminated")
	_, err = waitClusterDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Cluster (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findCluster(ctx context.Context, conn *emr.EMR, input *emr.DescribeClusterInput) (*emr.Cluster, error) {
	output, err := conn.DescribeClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeClusterNotFound) || tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "is not valid") {
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

func findClusterByID(ctx context.Context, conn *emr.EMR, id string) (*emr.Cluster, error) {
	input := &emr.DescribeClusterInput{
		ClusterId: aws.String(id),
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.Id) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	if state := aws.StringValue(output.Status.State); state == emr.ClusterStateTerminated || state == emr.ClusterStateTerminatedWithErrors {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusCluster(ctx context.Context, conn *emr.EMR, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
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

		return output, aws.StringValue(output.Status.State), nil
	}
}

func waitClusterCreated(ctx context.Context, conn *emr.EMR, id string) (*emr.Cluster, error) {
	const (
		timeout = 75 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{emr.ClusterStateBootstrapping, emr.ClusterStateStarting},
		Target:     []string{emr.ClusterStateRunning, emr.ClusterStateWaiting},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*emr.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateChangeReason.Code), aws.StringValue(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *emr.EMR, id string) (*emr.Cluster, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{emr.ClusterStateTerminating},
		Target:     []string{emr.ClusterStateTerminated, emr.ClusterStateTerminatedWithErrors},
		Refresh:    statusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*emr.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateChangeReason.Code), aws.StringValue(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func expandApplications(apps []interface{}) []*emr.Application {
	appOut := make([]*emr.Application, 0, len(apps))

	for _, appName := range flex.ExpandStringList(apps) {
		app := &emr.Application{
			Name: appName,
		}
		appOut = append(appOut, app)
	}
	return appOut
}

func flattenApplications(apps []*emr.Application) []interface{} {
	appOut := make([]interface{}, 0, len(apps))

	for _, app := range apps {
		appOut = append(appOut, aws.StringValue(app.Name))
	}
	return appOut
}

func flattenEC2InstanceAttributes(ia *emr.Ec2InstanceAttributes) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0)

	if ia.Ec2KeyName != nil {
		attrs["key_name"] = aws.StringValue(ia.Ec2KeyName)
	}
	if ia.Ec2SubnetId != nil {
		attrs[names.AttrSubnetID] = aws.StringValue(ia.Ec2SubnetId)
	}
	if ia.RequestedEc2SubnetIds != nil && len(ia.RequestedEc2SubnetIds) > 0 {
		attrs[names.AttrSubnetIDs] = flex.FlattenStringSet(ia.RequestedEc2SubnetIds)
	}
	if ia.IamInstanceProfile != nil {
		attrs["instance_profile"] = aws.StringValue(ia.IamInstanceProfile)
	}
	if ia.EmrManagedMasterSecurityGroup != nil {
		attrs["emr_managed_master_security_group"] = aws.StringValue(ia.EmrManagedMasterSecurityGroup)
	}
	if ia.EmrManagedSlaveSecurityGroup != nil {
		attrs["emr_managed_slave_security_group"] = aws.StringValue(ia.EmrManagedSlaveSecurityGroup)
	}

	if len(ia.AdditionalMasterSecurityGroups) > 0 {
		strs := aws.StringValueSlice(ia.AdditionalMasterSecurityGroups)
		attrs["additional_master_security_groups"] = strings.Join(strs, ",")
	}
	if len(ia.AdditionalSlaveSecurityGroups) > 0 {
		strs := aws.StringValueSlice(ia.AdditionalSlaveSecurityGroups)
		attrs["additional_slave_security_groups"] = strings.Join(strs, ",")
	}

	if ia.ServiceAccessSecurityGroup != nil {
		attrs["service_access_security_group"] = aws.StringValue(ia.ServiceAccessSecurityGroup)
	}

	result = append(result, attrs)

	return result
}

func flattenAutoScalingPolicyDescription(policy *emr.AutoScalingPolicyDescription) (string, error) {
	if policy == nil {
		return "", nil
	}

	// AutoScalingPolicy has an additional Status field and null values that are causing a new hashcode to be generated
	// for `instance_group`.
	// We are purposefully omitting that field and the null values here when we flatten the autoscaling policy string
	// for the statefile.
	for i, rule := range policy.Rules {
		for j, dimension := range rule.Trigger.CloudWatchAlarmDefinition.Dimensions {
			if aws.StringValue(dimension.Key) == "JobFlowId" {
				policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = slices.Delete(policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions, j, j+1)
			}
		}
		if len(policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions) == 0 {
			policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = nil
		}
	}

	tmpAutoScalingPolicy := emr.AutoScalingPolicy{
		Constraints: policy.Constraints,
		Rules:       policy.Rules,
	}
	autoscalingPolicyConstraintsBytes, err := json.Marshal(tmpAutoScalingPolicy.Constraints)
	if err != nil {
		return "", fmt.Errorf("parsing EMR Cluster Instance Groups AutoScalingPolicy Constraints: %w", err)
	}
	autoscalingPolicyConstraintsString := string(autoscalingPolicyConstraintsBytes)

	autoscalingPolicyRulesBytes, err := json.Marshal(tmpAutoScalingPolicy.Rules)
	if err != nil {
		return "", fmt.Errorf("parsing EMR Cluster Instance Groups AutoScalingPolicy Rules: %w", err)
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(autoscalingPolicyRulesBytes, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNil(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}
	autoscalingPolicyRulesString := string(withoutNulls)

	autoscalingPolicyString := fmt.Sprintf("{\"Constraints\":%s,\"Rules\":%s}", autoscalingPolicyConstraintsString, autoscalingPolicyRulesString)

	return autoscalingPolicyString, nil
}

func flattenCoreInstanceGroup(instanceGroup *emr.InstanceGroup) ([]interface{}, error) {
	if instanceGroup == nil {
		return []interface{}{}, nil
	}

	autoscalingPolicy, err := flattenAutoScalingPolicyDescription(instanceGroup.AutoScalingPolicy)

	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"autoscaling_policy":    autoscalingPolicy,
		"bid_price":             aws.StringValue(instanceGroup.BidPrice),
		"ebs_config":            flattenEBSConfig(instanceGroup.EbsBlockDevices),
		names.AttrID:            aws.StringValue(instanceGroup.Id),
		names.AttrInstanceCount: aws.Int64Value(instanceGroup.RequestedInstanceCount),
		names.AttrInstanceType:  aws.StringValue(instanceGroup.InstanceType),
		names.AttrName:          aws.StringValue(instanceGroup.Name),
	}

	return []interface{}{m}, nil
}

func flattenMasterInstanceGroup(instanceGroup *emr.InstanceGroup) []interface{} {
	if instanceGroup == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bid_price":             aws.StringValue(instanceGroup.BidPrice),
		"ebs_config":            flattenEBSConfig(instanceGroup.EbsBlockDevices),
		names.AttrID:            aws.StringValue(instanceGroup.Id),
		names.AttrInstanceCount: aws.Int64Value(instanceGroup.RequestedInstanceCount),
		names.AttrInstanceType:  aws.StringValue(instanceGroup.InstanceType),
		names.AttrName:          aws.StringValue(instanceGroup.Name),
	}

	return []interface{}{m}
}

func flattenKerberosAttributes(d *schema.ResourceData, kerberosAttributes *emr.KerberosAttributes) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)

	if kerberosAttributes == nil || kerberosAttributes.Realm == nil {
		return l
	}

	// Do not set from API:
	// * ad_domain_join_password
	// * ad_domain_join_user
	// * cross_realm_trust_principal_password
	// * kdc_admin_password

	m := map[string]interface{}{
		"kdc_admin_password": d.Get("kerberos_attributes.0.kdc_admin_password").(string),
		"realm":              aws.StringValue(kerberosAttributes.Realm),
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_password"); ok {
		m["ad_domain_join_password"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_user"); ok {
		m["ad_domain_join_user"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.cross_realm_trust_principal_password"); ok {
		m["cross_realm_trust_principal_password"] = v.(string)
	}

	l = append(l, m)

	return l
}

func flattenHadoopStepConfig(config *emr.HadoopStepConfig) map[string]interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{
		"args":               aws.StringValueSlice(config.Args),
		"jar":                aws.StringValue(config.Jar),
		"main_class":         aws.StringValue(config.MainClass),
		names.AttrProperties: aws.StringValueMap(config.Properties),
	}

	return m
}

func flattenStepSummaries(stepSummaries []*emr.StepSummary) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)

	if len(stepSummaries) == 0 {
		return l
	}

	for _, stepSummary := range stepSummaries {
		l = append(l, flattenStepSummary(stepSummary))
	}

	return l
}

func flattenStepSummary(stepSummary *emr.StepSummary) map[string]interface{} {
	if stepSummary == nil {
		return nil
	}

	m := map[string]interface{}{
		"action_on_failure": aws.StringValue(stepSummary.ActionOnFailure),
		"hadoop_jar_step":   []map[string]interface{}{flattenHadoopStepConfig(stepSummary.Config)},
		names.AttrName:      aws.StringValue(stepSummary.Name),
	}

	return m
}

func flattenEBSConfig(ebsBlockDevices []*emr.EbsBlockDevice) *schema.Set {
	uniqueEBS := make(map[int]int)
	ebsConfig := make([]interface{}, 0)
	for _, ebs := range ebsBlockDevices {
		ebsAttrs := make(map[string]interface{})
		if ebs.VolumeSpecification.Iops != nil {
			ebsAttrs[names.AttrIOPS] = int(aws.Int64Value(ebs.VolumeSpecification.Iops))
		}
		if ebs.VolumeSpecification.SizeInGB != nil {
			ebsAttrs[names.AttrSize] = int(aws.Int64Value(ebs.VolumeSpecification.SizeInGB))
		}
		if ebs.VolumeSpecification.Throughput != nil {
			ebsAttrs[names.AttrThroughput] = aws.Int64Value(ebs.VolumeSpecification.Throughput)
		}
		if ebs.VolumeSpecification.VolumeType != nil {
			ebsAttrs[names.AttrType] = aws.StringValue(ebs.VolumeSpecification.VolumeType)
		}
		ebsAttrs["volumes_per_instance"] = 1
		uniqueEBS[resourceClusterEBSHashConfig(ebsAttrs)] += 1
		ebsConfig = append(ebsConfig, ebsAttrs)
	}

	for _, ebs := range ebsConfig {
		ebs.(map[string]interface{})["volumes_per_instance"] = uniqueEBS[resourceClusterEBSHashConfig(ebs)]
	}
	return schema.NewSet(resourceClusterEBSHashConfig, ebsConfig)
}

func flattenBootstrapArguments(actions []*emr.Command) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, b := range actions {
		attrs := make(map[string]interface{})
		attrs[names.AttrName] = aws.StringValue(b.Name)
		attrs[names.AttrPath] = aws.StringValue(b.ScriptPath)
		attrs["args"] = flex.FlattenStringList(b.Args)
		result = append(result, attrs)
	}

	return result
}

func coreInstanceGroup(grps []*emr.InstanceGroup) *emr.InstanceGroup {
	for _, grp := range grps {
		if aws.StringValue(grp.InstanceGroupType) == emr.InstanceGroupTypeCore {
			return grp
		}
	}
	return nil
}

func expandBootstrapActions(bootstrapActions []interface{}) []*emr.BootstrapActionConfig {
	actionsOut := []*emr.BootstrapActionConfig{}

	for _, raw := range bootstrapActions {
		actionAttributes := raw.(map[string]interface{})
		actionName := actionAttributes[names.AttrName].(string)
		actionPath := actionAttributes[names.AttrPath].(string)
		actionArgs := actionAttributes["args"].([]interface{})

		action := &emr.BootstrapActionConfig{
			Name: aws.String(actionName),
			ScriptBootstrapAction: &emr.ScriptBootstrapActionConfig{
				Path: aws.String(actionPath),
				Args: flex.ExpandStringListEmpty(actionArgs),
			},
		}
		actionsOut = append(actionsOut, action)
	}

	return actionsOut
}

func expandHadoopJarStepConfig(m map[string]interface{}) *emr.HadoopJarStepConfig {
	hadoopJarStepConfig := &emr.HadoopJarStepConfig{
		Jar: aws.String(m["jar"].(string)),
	}

	if v, ok := m["args"]; ok {
		hadoopJarStepConfig.Args = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := m["main_class"]; ok {
		hadoopJarStepConfig.MainClass = aws.String(v.(string))
	}

	if v, ok := m[names.AttrProperties]; ok {
		hadoopJarStepConfig.Properties = expandKeyValues(v.(map[string]interface{}))
	}

	return hadoopJarStepConfig
}

func expandKeyValues(m map[string]interface{}) []*emr.KeyValue {
	keyValues := make([]*emr.KeyValue, 0)

	for k, v := range m {
		keyValue := &emr.KeyValue{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		keyValues = append(keyValues, keyValue)
	}

	return keyValues
}

func expandKerberosAttributes(m map[string]interface{}) *emr.KerberosAttributes {
	kerberosAttributes := &emr.KerberosAttributes{
		KdcAdminPassword: aws.String(m["kdc_admin_password"].(string)),
		Realm:            aws.String(m["realm"].(string)),
	}
	if v, ok := m["ad_domain_join_password"]; ok && v.(string) != "" {
		kerberosAttributes.ADDomainJoinPassword = aws.String(v.(string))
	}
	if v, ok := m["ad_domain_join_user"]; ok && v.(string) != "" {
		kerberosAttributes.ADDomainJoinUser = aws.String(v.(string))
	}
	if v, ok := m["cross_realm_trust_principal_password"]; ok && v.(string) != "" {
		kerberosAttributes.CrossRealmTrustPrincipalPassword = aws.String(v.(string))
	}
	return kerberosAttributes
}

func expandStepConfig(m map[string]interface{}) *emr.StepConfig {
	hadoopJarStepList := m["hadoop_jar_step"].([]interface{})
	hadoopJarStepMap := hadoopJarStepList[0].(map[string]interface{})

	stepConfig := &emr.StepConfig{
		ActionOnFailure: aws.String(m["action_on_failure"].(string)),
		HadoopJarStep:   expandHadoopJarStepConfig(hadoopJarStepMap),
		Name:            aws.String(m[names.AttrName].(string)),
	}

	return stepConfig
}

func expandStepConfigs(l []interface{}) []*emr.StepConfig {
	stepConfigs := []*emr.StepConfig{}

	for _, raw := range l {
		m := raw.(map[string]interface{})
		stepConfigs = append(stepConfigs, expandStepConfig(m))
	}

	return stepConfigs
}

func expandEBSConfig(configAttributes map[string]interface{}, config *emr.InstanceGroupConfig) {
	if rawEbsConfigs, ok := configAttributes["ebs_config"]; ok {
		ebsConfig := &emr.EbsConfiguration{}

		ebsBlockDeviceConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
		for _, rawEbsConfig := range rawEbsConfigs.(*schema.Set).List() {
			rawEbsConfig := rawEbsConfig.(map[string]interface{})
			ebsBlockDeviceConfig := &emr.EbsBlockDeviceConfig{
				VolumesPerInstance: aws.Int64(int64(rawEbsConfig["volumes_per_instance"].(int))),
				VolumeSpecification: &emr.VolumeSpecification{
					SizeInGB:   aws.Int64(int64(rawEbsConfig[names.AttrSize].(int))),
					VolumeType: aws.String(rawEbsConfig[names.AttrType].(string)),
				},
			}
			if v, ok := rawEbsConfig[names.AttrThroughput].(int); ok && v != 0 {
				ebsBlockDeviceConfig.VolumeSpecification.Throughput = aws.Int64(int64(v))
			}
			if v, ok := rawEbsConfig[names.AttrIOPS].(int); ok && v != 0 {
				ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int64(int64(v))
			}
			ebsBlockDeviceConfigs = append(ebsBlockDeviceConfigs, ebsBlockDeviceConfig)
		}
		ebsConfig.EbsBlockDeviceConfigs = ebsBlockDeviceConfigs

		config.EbsConfiguration = ebsConfig
	}
}

func expandConfigurationJSON(input string) ([]*emr.Configuration, error) {
	configsOut := []*emr.Configuration{}
	err := json.Unmarshal([]byte(input), &configsOut)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Expanded EMR Configurations %s", configsOut)

	return configsOut, nil
}

func flattenConfigurationJSON(config []*emr.Configuration) (string, error) {
	out, err := jsonutil.BuildJSON(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func expandConfigures(input string) []*emr.Configuration {
	configsOut := []*emr.Configuration{}
	if strings.HasPrefix(input, "http") {
		if err := readHTTPJSON(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading HTTP JSON: %s", err)
		}
	} else if strings.HasSuffix(input, ".json") {
		if err := readLocalJSON(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading local JSON: %s", err)
		}
	} else {
		if err := readBodyJSON(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading body JSON: %s", err)
		}
	}
	log.Printf("[DEBUG] Expanded EMR Configurations %s", configsOut)

	return configsOut
}

func readHTTPJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func readLocalJSON(localFile string, target interface{}) error {
	file, e := os.ReadFile(localFile)
	if e != nil {
		log.Printf("[ERROR] %s", e)
		return e
	}

	return json.Unmarshal(file, target)
}

func readBodyJSON(body string, target interface{}) error {
	log.Printf("[DEBUG] Raw Body %s\n", body)
	err := json.Unmarshal([]byte(body), target)
	if err != nil {
		log.Printf("[ERROR] parsing JSON %s", err)
		return err
	}
	return nil
}

func findMasterGroup(instanceGroups []*emr.InstanceGroup) *emr.InstanceGroup {
	for _, group := range instanceGroups {
		if aws.StringValue(group.InstanceGroupType) == emr.InstanceRoleTypeMaster {
			return group
		}
	}
	return nil
}

func resourceClusterEBSHashConfig(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m[names.AttrSize].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrType].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["volumes_per_instance"].(int)))
	if v, ok := m[names.AttrThroughput].(int); ok && v != 0 {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	if v, ok := m[names.AttrIOPS].(int); ok && v != 0 {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	return create.StringHashcode(buf.String())
}

func getCoreInstanceGroupAutoScalingPolicy(ctx context.Context, conn *emr.EMR, clusterID string) (*emr.AutoScalingPolicyDescription, error) {
	instanceGroups, err := fetchAllInstanceGroups(ctx, conn, clusterID)

	if err != nil {
		return nil, err
	}

	coreGroup := coreInstanceGroup(instanceGroups)

	if coreGroup == nil {
		return nil, fmt.Errorf("EMR Cluster (%s) Core Instance Group not found", clusterID)
	}

	return coreGroup.AutoScalingPolicy, nil
}

func fetchAllInstanceGroups(ctx context.Context, conn *emr.EMR, clusterID string) ([]*emr.InstanceGroup, error) {
	input := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterID),
	}
	var groups []*emr.InstanceGroup

	err := conn.ListInstanceGroupsPagesWithContext(ctx, input, func(page *emr.ListInstanceGroupsOutput, lastPage bool) bool {
		groups = append(groups, page.InstanceGroups...)

		return !lastPage
	})

	return groups, err
}

func readInstanceFleetConfig(data map[string]interface{}, InstanceFleetType string) *emr.InstanceFleetConfig {
	config := &emr.InstanceFleetConfig{
		InstanceFleetType:      &InstanceFleetType,
		Name:                   aws.String(data[names.AttrName].(string)),
		TargetOnDemandCapacity: aws.Int64(int64(data["target_on_demand_capacity"].(int))),
		TargetSpotCapacity:     aws.Int64(int64(data["target_spot_capacity"].(int))),
	}

	if v, ok := data["instance_type_configs"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := data["launch_specifications"].([]interface{}); ok && len(v) == 1 {
		config.LaunchSpecifications = expandLaunchSpecification(v[0].(map[string]interface{}))
	}

	return config
}

func FetchAllInstanceFleets(ctx context.Context, conn *emr.EMR, clusterID string) ([]*emr.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}
	var fleets []*emr.InstanceFleet

	err := conn.ListInstanceFleetsPagesWithContext(ctx, input, func(page *emr.ListInstanceFleetsOutput, lastPage bool) bool {
		fleets = append(fleets, page.InstanceFleets...)

		return !lastPage
	})

	return fleets, err
}

func findInstanceFleet(instanceFleets []*emr.InstanceFleet, instanceRoleType string) *emr.InstanceFleet {
	for _, instanceFleet := range instanceFleets {
		if instanceFleet.InstanceFleetType != nil {
			if aws.StringValue(instanceFleet.InstanceFleetType) == instanceRoleType {
				return instanceFleet
			}
		}
	}
	return nil
}

func flattenInstanceFleet(instanceFleet *emr.InstanceFleet) []interface{} {
	if instanceFleet == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrID:                     aws.StringValue(instanceFleet.Id),
		names.AttrName:                   aws.StringValue(instanceFleet.Name),
		"target_on_demand_capacity":      aws.Int64Value(instanceFleet.TargetOnDemandCapacity),
		"target_spot_capacity":           aws.Int64Value(instanceFleet.TargetSpotCapacity),
		"provisioned_on_demand_capacity": aws.Int64Value(instanceFleet.ProvisionedOnDemandCapacity),
		"provisioned_spot_capacity":      aws.Int64Value(instanceFleet.ProvisionedSpotCapacity),
		"instance_type_configs":          flatteninstanceTypeConfigs(instanceFleet.InstanceTypeSpecifications),
		"launch_specifications":          flattenLaunchSpecifications(instanceFleet.LaunchSpecifications),
	}

	return []interface{}{m}
}

func flatteninstanceTypeConfigs(instanceTypeSpecifications []*emr.InstanceTypeSpecification) *schema.Set {
	instanceTypeConfigs := make([]interface{}, 0)

	for _, itc := range instanceTypeSpecifications {
		flattenTypeConfig := make(map[string]interface{})

		if itc.BidPrice != nil {
			flattenTypeConfig["bid_price"] = aws.StringValue(itc.BidPrice)
		}

		if itc.BidPriceAsPercentageOfOnDemandPrice != nil {
			flattenTypeConfig["bid_price_as_percentage_of_on_demand_price"] = aws.Float64Value(itc.BidPriceAsPercentageOfOnDemandPrice)
		}

		flattenTypeConfig[names.AttrInstanceType] = aws.StringValue(itc.InstanceType)
		flattenTypeConfig["weighted_capacity"] = int(aws.Int64Value(itc.WeightedCapacity))

		flattenTypeConfig["ebs_config"] = flattenEBSConfig(itc.EbsBlockDevices)

		instanceTypeConfigs = append(instanceTypeConfigs, flattenTypeConfig)
	}

	return schema.NewSet(resourceInstanceTypeHashConfig, instanceTypeConfigs)
}

func flattenLaunchSpecifications(launchSpecifications *emr.InstanceFleetProvisioningSpecifications) []interface{} {
	if launchSpecifications == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"on_demand_specification": flattenOnDemandSpecification(launchSpecifications.OnDemandSpecification),
		"spot_specification":      flattenSpotSpecification(launchSpecifications.SpotSpecification),
	}
	return []interface{}{m}
}

func flattenOnDemandSpecification(onDemandSpecification *emr.OnDemandProvisioningSpecification) []interface{} {
	if onDemandSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		// The return value from api is wrong. it return the value with uppercase letters and '_' vs. '-'
		// The value needs to be normalized to avoid perpetual difference in the Terraform plan
		"allocation_strategy": strings.Replace(strings.ToLower(aws.StringValue(onDemandSpecification.AllocationStrategy)), "_", "-", -1),
	}
	return []interface{}{m}
}

func flattenSpotSpecification(spotSpecification *emr.SpotProvisioningSpecification) []interface{} {
	if spotSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"timeout_action":           aws.StringValue(spotSpecification.TimeoutAction),
		"timeout_duration_minutes": aws.Int64Value(spotSpecification.TimeoutDurationMinutes),
	}
	if spotSpecification.BlockDurationMinutes != nil {
		m["block_duration_minutes"] = aws.Int64Value(spotSpecification.BlockDurationMinutes)
	}
	if spotSpecification.AllocationStrategy != nil {
		// The return value from api is wrong. it return the value with uppercase letters and '_' vs. '-'
		// The value needs to be normalized to avoid perpetual difference in the Terraform plan
		m["allocation_strategy"] = strings.Replace(strings.ToLower(aws.StringValue(spotSpecification.AllocationStrategy)), "_", "-", -1)
	}

	return []interface{}{m}
}

func expandEBSConfiguration(ebsConfigurations []interface{}) *emr.EbsConfiguration {
	ebsConfig := &emr.EbsConfiguration{}
	ebsConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
	for _, ebsConfiguration := range ebsConfigurations {
		cfg := ebsConfiguration.(map[string]interface{})
		ebsBlockDeviceConfig := &emr.EbsBlockDeviceConfig{
			VolumesPerInstance: aws.Int64(int64(cfg["volumes_per_instance"].(int))),
			VolumeSpecification: &emr.VolumeSpecification{
				SizeInGB:   aws.Int64(int64(cfg[names.AttrSize].(int))),
				VolumeType: aws.String(cfg[names.AttrType].(string)),
			},
		}
		if v, ok := cfg[names.AttrThroughput].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Throughput = aws.Int64(int64(v))
		}
		if v, ok := cfg[names.AttrIOPS].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int64(int64(v))
		}
		ebsConfigs = append(ebsConfigs, ebsBlockDeviceConfig)
	}
	ebsConfig.EbsBlockDeviceConfigs = ebsConfigs
	return ebsConfig
}

func expandInstanceTypeConfigs(instanceTypeConfigs []interface{}) []*emr.InstanceTypeConfig {
	configsOut := []*emr.InstanceTypeConfig{}

	for _, raw := range instanceTypeConfigs {
		configAttributes := raw.(map[string]interface{})

		config := &emr.InstanceTypeConfig{
			InstanceType: aws.String(configAttributes[names.AttrInstanceType].(string)),
		}

		if bidPrice, ok := configAttributes["bid_price"]; ok {
			if bidPrice != "" {
				config.BidPrice = aws.String(bidPrice.(string))
			}
		}

		if v, ok := configAttributes["bid_price_as_percentage_of_on_demand_price"].(float64); ok && v != 0 {
			config.BidPriceAsPercentageOfOnDemandPrice = aws.Float64(v)
		}

		if v, ok := configAttributes["weighted_capacity"].(int); ok {
			config.WeightedCapacity = aws.Int64(int64(v))
		}

		if v, ok := configAttributes["configurations"].(*schema.Set); ok && v.Len() > 0 {
			config.Configurations = expandConfigurations(v.List())
		}

		if v, ok := configAttributes["ebs_config"].(*schema.Set); ok && v.Len() == 1 {
			config.EbsConfiguration = expandEBSConfiguration(v.List())
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func expandLaunchSpecification(launchSpecification map[string]interface{}) *emr.InstanceFleetProvisioningSpecifications {
	onDemandSpecification := launchSpecification["on_demand_specification"].([]interface{})
	spotSpecification := launchSpecification["spot_specification"].([]interface{})

	fleetSpecification := &emr.InstanceFleetProvisioningSpecifications{}

	if len(onDemandSpecification) > 0 {
		fleetSpecification.OnDemandSpecification = &emr.OnDemandProvisioningSpecification{
			AllocationStrategy: aws.String(onDemandSpecification[0].(map[string]interface{})["allocation_strategy"].(string)),
		}
	}

	if len(spotSpecification) > 0 {
		configAttributes := spotSpecification[0].(map[string]interface{})
		spotProvisioning := &emr.SpotProvisioningSpecification{
			TimeoutAction:          aws.String(configAttributes["timeout_action"].(string)),
			TimeoutDurationMinutes: aws.Int64(int64(configAttributes["timeout_duration_minutes"].(int))),
		}
		if v, ok := configAttributes["block_duration_minutes"]; ok && v != 0 {
			spotProvisioning.BlockDurationMinutes = aws.Int64(int64(v.(int)))
		}
		if v, ok := configAttributes["allocation_strategy"]; ok {
			spotProvisioning.AllocationStrategy = aws.String(v.(string))
		}

		fleetSpecification.SpotSpecification = spotProvisioning
	}

	return fleetSpecification
}

func expandConfigurations(configurations []interface{}) []*emr.Configuration {
	configsOut := []*emr.Configuration{}

	for _, raw := range configurations {
		configAttributes := raw.(map[string]interface{})

		config := &emr.Configuration{}

		if v, ok := configAttributes["classification"].(string); ok {
			config.Classification = aws.String(v)
		}

		if v, ok := configAttributes["configurations"].([]interface{}); ok {
			config.Configurations = expandConfigurations(v)
		}

		if v, ok := configAttributes[names.AttrProperties].(map[string]interface{}); ok {
			properties := make(map[string]string)
			for k, pv := range v {
				properties[k] = pv.(string)
			}
			config.Properties = aws.StringMap(properties)
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func resourceInstanceTypeHashConfig(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
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

func removeNil(data map[string]interface{}) map[string]interface{} {
	withoutNil := make(map[string]interface{})

	for k, v := range data {
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			withoutNil[k] = removeNil(v)
		default:
			withoutNil[k] = v
		}
	}

	return withoutNil
}

func expandAutoTerminationPolicy(policy []interface{}) *emr.AutoTerminationPolicy {
	if len(policy) == 0 || policy[0] == nil {
		return nil
	}

	m := policy[0].(map[string]interface{})
	app := &emr.AutoTerminationPolicy{}

	if v, ok := m["idle_timeout"].(int); ok && v > 0 {
		app.IdleTimeout = aws.Int64(int64(v))
	}

	return app
}

func flattenAutoTerminationPolicy(atp *emr.AutoTerminationPolicy) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0)

	if atp == nil {
		return result
	}

	if atp.IdleTimeout != nil {
		attrs["idle_timeout"] = aws.Int64Value(atp.IdleTimeout)
	}

	result = append(result, attrs)

	return result
}

func expandPlacementGroupConfigs(placementGroupConfigs []interface{}) []*emr.PlacementGroupConfig {
	placementGroupConfigsOut := []*emr.PlacementGroupConfig{}

	for _, raw := range placementGroupConfigs {
		placementGroupAttributes := raw.(map[string]interface{})
		instanceRole := placementGroupAttributes["instance_role"].(string)

		placementGroupConfig := &emr.PlacementGroupConfig{
			InstanceRole: aws.String(instanceRole),
		}
		if v, ok := placementGroupAttributes["placement_strategy"]; ok && v.(string) != "" {
			placementGroupConfig.PlacementStrategy = aws.String(v.(string))
		}
		placementGroupConfigsOut = append(placementGroupConfigsOut, placementGroupConfig)
	}

	return placementGroupConfigsOut
}

func flattenPlacementGroupConfigs(placementGroupSpecifications []*emr.PlacementGroupConfig) []interface{} {
	if placementGroupSpecifications == nil {
		return []interface{}{}
	}

	placementGroupConfigs := make([]interface{}, 0)

	for _, pgc := range placementGroupSpecifications {
		placementGroupConfig := make(map[string]interface{})

		placementGroupConfig["instance_role"] = aws.StringValue(pgc.InstanceRole)

		if pgc.PlacementStrategy != nil {
			placementGroupConfig["placement_strategy"] = aws.StringValue(pgc.PlacementStrategy)
		}
		placementGroupConfigs = append(placementGroupConfigs, placementGroupConfig)
	}

	return placementGroupConfigs
}
