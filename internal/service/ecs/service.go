// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_service", name="Service")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("cluster")
// @IdentityAttribute("name)
// @ImportIDHandler("serviceImportID")
// @CustomImport
// @Testing(preIdentityVersion="v6.33.0")
// @Testing(idAttrDuplicates="arn")
// @Testing(importStateIdFunc=testAccServiceImportStateIdFunc)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ecs/types;types.Service")
// @Testing(importIgnore="wait_for_steady_state;task_definition")
func resourceService() *schema.Resource {
	// Resource with v0 schema (provider v5.58.0).
	resourceV0 := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"alarms": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			names.AttrCapacityProviderStrategy: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrWeight: {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"deployment_circuit_breaker": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"deployment_controller": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.DeploymentControllerTypeEcs,
						},
					},
				},
			},
			"deployment_maximum_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  200,
			},
			"deployment_minimum_healthy_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"force_new_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"health_check_grace_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"launch_type": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"load_balancer": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ordered_placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPropagateTags: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scheduling_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  awstypes.SchedulingStrategyReplica,
			},
			"service_connect_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"log_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_driver": {
										Type:     schema.TypeString,
										Required: true,
									},
									"options": {
										Type:     schema.TypeMap,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"secret_option": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												"value_from": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrNamespace: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"service": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_alias": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDNSName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrPort: {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"discovery_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ingress_port_override": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"port_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrTimeout: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_timeout_seconds": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"per_request_timeout_seconds": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"issuer_cert_authority": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aws_pca_authority_arn": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												names.AttrKMSKey: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrRoleARN: {
													Type:     schema.TypeString,
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
			"service_registries": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"registry_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_for_steady_state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"volume_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"managed_ebs_volume": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEncrypted: {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"file_system_type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  awstypes.TaskFilesystemTypeXfs,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrRoleARN: {
										Type:     schema.TypeString,
										Required: true,
									},
									"size_in_gb": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"alarms": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_names": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_rebalancing": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AvailabilityZoneRebalancing](),
			},
			names.AttrCapacityProviderStrategy: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrWeight: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"cluster": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"deployment_circuit_breaker": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"rollback": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"deployment_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bake_time_in_minutes": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 1440),
						},
						"linear_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"step_bake_time_in_minutes": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 1440),
									},
									"step_percent": {
										Type:         schema.TypeFloat,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.FloatBetween(3.0, 100.0),
									},
								},
							},
						},
						"canary_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"canary_bake_time_in_minutes": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 1440),
									},
									"canary_percent": {
										Type:         schema.TypeFloat,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.FloatBetween(0.1, 100.0),
									},
								},
							},
						},
						"lifecycle_hook": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hook_target_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"lifecycle_stages": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[awstypes.DeploymentLifecycleHookStage](),
										},
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"hook_details": {
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
										ValidateFunc:     verify.ValidStringIsJSONOrYAML,
									},
								},
							},
						},
						"strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DeploymentStrategy](),
						},
					},
				},
			},
			"deployment_controller": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DeploymentControllerTypeEcs,
							ValidateDiagFunc: enum.Validate[awstypes.DeploymentControllerType](),
						},
					},
				},
			},
			"deployment_maximum_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  200,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string)) == awstypes.SchedulingStrategyDaemon && new == "200" {
						return true
					}
					return false
				},
			},
			"deployment_minimum_healthy_percent": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string)) == awstypes.SchedulingStrategyDaemon && new == "100" {
						return true
					}
					return false
				},
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string)) == awstypes.SchedulingStrategyDaemon
				},
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"force_new_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"health_check_grace_period_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxInt32),
			},
			"iam_role": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"launch_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LaunchType](),
			},
			"load_balancer": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"advanced_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alternate_target_group_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"production_listener_rule": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"test_listener_rule": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ordered_placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeString,
							Optional: true,
							StateFunc: func(v any) string {
								value := v.(string)
								if value == "host" {
									return "instanceId"
								}
								return value
							},
							DiffSuppressFunc: sdkv2.SuppressEquivalentStringCaseInsensitive,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PlacementStrategyType](),
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.PlacementConstraintType](),
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrPropagateTags: {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if awstypes.PropagateTags(old) == awstypes.PropagateTagsNone && new == "" {
						return true
					}
					return false
				},
				ValidateDiagFunc: enum.Validate[awstypes.PropagateTags](),
			},
			"scheduling_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.SchedulingStrategyReplica,
				ValidateDiagFunc: enum.Validate[awstypes.SchedulingStrategy](),
			},
			"service_connect_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_log_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFormat: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ServiceConnectAccessLoggingFormat](),
									},
									"include_query_parameters": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ServiceConnectIncludeQueryParameters](),
									},
								},
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"log_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_driver": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.LogDriver](),
									},
									"options": {
										Type:     schema.TypeMap,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"secret_option": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												"value_from": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrNamespace: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_alias": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDNSName: {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												names.AttrPort: {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
												"test_traffic_rules": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrHeader: {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrName: {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		names.AttrValue: {
																			Type:     schema.TypeList,
																			Required: true,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact": {
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
									"discovery_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"ingress_port_override": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 65535),
									},
									"port_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrTimeout: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle_timeout_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 2147483647),
												},
												"per_request_timeout_seconds": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 2147483647),
												},
											},
										},
									},
									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"issuer_cert_authority": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aws_pca_authority_arn": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
														},
													},
												},
												names.AttrKMSKey: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrRoleARN: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
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
			"service_registries": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"registry_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"sigint_rollback": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// modeled after null_resource & aws_api_gateway_deployment
			// only for _updates in-place_ rather than replacements
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_for_steady_state": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"volume_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"managed_ebs_volume": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEncrypted: {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},
									"file_system_type": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.TaskFilesystemTypeXfs,
										ValidateDiagFunc: enum.Validate[awstypes.TaskFilesystemType](),
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"size_in_gb": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag_specifications": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPropagateTags: {
													Type:     schema.TypeString,
													Optional: true,
													DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
														if awstypes.PropagateTags(old) == awstypes.PropagateTagsNone && new == "" {
															return true
														}
														return false
													},
													ValidateDiagFunc: enum.Validate[awstypes.PropagateTags](),
												},
												names.AttrResourceType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EBSResourceType](),
												},
												names.AttrTags: tftags.TagsSchema(),
											},
										},
									},
									names.AttrThroughput: {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 1000),
									},
									"volume_initialization_rate": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"vpc_lattice_configurations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"port_name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexache.MustCompile(`^[0-9a-z_-]+$`), "must contain only lowercase letters, numbers, underscores and hyphens"),
								validation.StringDoesNotMatch(regexache.MustCompile(`^-`), "cannot begin with a hyphen"),
							),
						},
					},
				},
			},
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: resourceV0.CoreConfigSchema().ImpliedType(),
				Upgrade: func(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
					// Convert volume_configuration.managed_ebs_volume.throughput from string to int.
					if v, ok := rawState["volume_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
						tfMap := v[0].(map[string]any)
						if v, ok := tfMap["managed_ebs_volume"].([]any); ok && len(v) > 0 && v[0] != nil {
							tfMap := v[0].(map[string]any)
							if v, ok := tfMap[names.AttrThroughput]; ok {
								if v, ok := v.(string); ok {
									if v == "" {
										tfMap[names.AttrThroughput] = 0
									} else {
										if v, err := strconv.Atoi(v); err == nil {
											tfMap[names.AttrThroughput] = v
										} else {
											return nil, err
										}
									}
								}
							}
						}
					}

					return rawState, nil
				},
				Version: 0,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			capacityProviderStrategyCustomizeDiff,
			triggersCustomizeDiff,
		),
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition(ctx)

	deploymentController := expandDeploymentController(d.Get("deployment_controller").([]any))
	deploymentMinimumHealthyPercent := d.Get("deployment_minimum_healthy_percent").(int)
	name := d.Get(names.AttrName).(string)
	schedulingStrategy := awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string))
	input := ecs.CreateServiceInput{
		CapacityProviderStrategy: expandCapacityProviderStrategyItems(d.Get(names.AttrCapacityProviderStrategy).(*schema.Set)),
		ClientToken:              aws.String(sdkid.UniqueId()),
		DeploymentConfiguration:  &awstypes.DeploymentConfiguration{},
		DeploymentController:     deploymentController,
		EnableECSManagedTags:     d.Get("enable_ecs_managed_tags").(bool),
		EnableExecuteCommand:     d.Get("enable_execute_command").(bool),
		NetworkConfiguration:     expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]any)),
		SchedulingStrategy:       schedulingStrategy,
		ServiceName:              aws.String(name),
		Tags:                     getTagsIn(ctx),
		VpcLatticeConfigurations: expandVPCLatticeConfiguration(d.Get("vpc_lattice_configurations").(*schema.Set)),
	}

	if v, ok := d.GetOk("alarms"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeploymentConfiguration.Alarms = expandAlarms(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("availability_zone_rebalancing"); ok {
		input.AvailabilityZoneRebalancing = awstypes.AvailabilityZoneRebalancing(v.(string))
	}

	if v, ok := d.GetOk("cluster"); ok {
		input.Cluster = aws.String(v.(string))
	}

	if schedulingStrategy == awstypes.SchedulingStrategyDaemon && deploymentMinimumHealthyPercent != 100 {
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(deploymentMinimumHealthyPercent))
	} else if schedulingStrategy == awstypes.SchedulingStrategyReplica {
		input.DeploymentConfiguration.MaximumPercent = aws.Int32(int32(d.Get("deployment_maximum_percent").(int)))
		input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(deploymentMinimumHealthyPercent))
		input.DesiredCount = aws.Int32(int32(d.Get("desired_count").(int)))
	}

	if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("deployment_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		config := v.([]any)[0].(map[string]any)

		if strategy, ok := config["strategy"].(string); ok && strategy != "" {
			input.DeploymentConfiguration.Strategy = awstypes.DeploymentStrategy(strategy)

			if v, ok := config["bake_time_in_minutes"].(string); ok {
				bt, err := expandBakeTimeInMinutes(v)
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
				input.DeploymentConfiguration.BakeTimeInMinutes = bt
			}

			if awstypes.DeploymentStrategy(strategy) == awstypes.DeploymentStrategyLinear {
				if v, ok := config["linear_configuration"].([]any); ok && len(v) > 0 {
					if linearConfig, ok := v[0].(map[string]any); ok {
						sp, sbtm, err := expandLinearConfiguration(linearConfig)
						if err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}
						input.DeploymentConfiguration.LinearConfiguration = &awstypes.LinearConfiguration{
							StepPercent:           sp,
							StepBakeTimeInMinutes: sbtm,
						}
					}
				}
			}

			if awstypes.DeploymentStrategy(strategy) == awstypes.DeploymentStrategyCanary {
				if v, ok := config["canary_configuration"].([]any); ok && len(v) > 0 {
					if canaryConfig, ok := v[0].(map[string]any); ok {
						cp, cbtm, err := expandCanaryConfiguration(canaryConfig)
						if err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}
						input.DeploymentConfiguration.CanaryConfiguration = &awstypes.CanaryConfiguration{
							CanaryPercent:           cp,
							CanaryBakeTimeInMinutes: cbtm,
						}
					}
				}
			}

			if hooks := config["lifecycle_hook"].(*schema.Set).List(); len(hooks) > 0 {
				input.DeploymentConfiguration.LifecycleHooks = expandLifecycleHooks(hooks)
			}
		}
	}

	if v, ok := d.GetOk("health_check_grace_period_seconds"); ok {
		input.HealthCheckGracePeriodSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("iam_role"); ok {
		input.Role = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = awstypes.LaunchType(v.(string))
		// When creating a service that uses the EXTERNAL deployment controller,
		// you can specify only parameters that aren't controlled at the task set level
		// hence you cannot set LaunchType, not changing the default launch_type from EC2 to empty
		// string to have backward compatibility
		if deploymentController != nil && deploymentController.Type == awstypes.DeploymentControllerTypeExternal {
			input.LaunchType = awstypes.LaunchType("")
		}
	}

	if v := expandServiceLoadBalancers(d.Get("load_balancer").(*schema.Set).List()); len(v) > 0 {
		input.LoadBalancers = v
	}

	if v, ok := d.GetOk("ordered_placement_strategy"); ok {
		apiObject, err := expandPlacementStrategy(v.([]any))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementStrategy = apiObject
	}

	if v, ok := d.Get("placement_constraints").(*schema.Set); ok {
		apiObject, err := expandPlacementConstraints(v.List())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.PlacementConstraints = apiObject
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPropagateTags); ok {
		input.PropagateTags = awstypes.PropagateTags(v.(string))
	}

	if v, ok := d.GetOk("service_connect_configuration"); ok && len(v.([]any)) > 0 {
		input.ServiceConnectConfiguration = expandServiceConnectConfiguration(v.([]any))
	}

	if v := d.Get("service_registries").([]any); len(v) > 0 {
		input.ServiceRegistries = expandServiceRegistries(v)
	}

	if v, ok := d.GetOk("task_definition"); ok {
		input.TaskDefinition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume_configuration"); ok && len(v.([]any)) > 0 {
		input.VolumeConfigurations = expandServiceVolumeConfigurations(ctx, v.([]any))
	}

	operationTime := time.Now().UTC()
	output, err := retryServiceCreate(ctx, conn, &input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryServiceCreate(ctx, conn, &input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Service.ServiceArn))
	d.Set(names.AttrARN, output.Service.ServiceArn)

	if d.Get("wait_for_steady_state").(bool) {
		if _, err := waitServiceStable(ctx, conn, d.Id(), d.Get("cluster").(string), operationTime, d.Get("sigint_rollback").(bool), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) create: %s", d.Id(), err)
		}
	} else if _, err := waitServiceActive(ctx, conn, d.Id(), d.Get("cluster").(string), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceServiceRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Service (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster := d.Get("cluster").(string)
	service, err := findServiceByTwoPartKeyWaitForActive(ctx, conn, d.Id(), cluster)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ECS Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Service (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, service.ServiceArn)
	d.Set("availability_zone_rebalancing", service.AvailabilityZoneRebalancing)
	if err := d.Set(names.AttrCapacityProviderStrategy, flattenCapacityProviderStrategyItems(service.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}
	// Save cluster in the same format.
	if arn.IsARN(cluster) {
		d.Set("cluster", service.ClusterArn)
	} else {
		d.Set("cluster", clusterNameFromARN(aws.ToString(service.ClusterArn)))
	}
	if service.DeploymentConfiguration != nil {
		d.Set("deployment_maximum_percent", service.DeploymentConfiguration.MaximumPercent)
		d.Set("deployment_minimum_healthy_percent", service.DeploymentConfiguration.MinimumHealthyPercent)

		if service.DeploymentConfiguration.Alarms != nil {
			if err := d.Set("alarms", []any{flattenAlarms(service.DeploymentConfiguration.Alarms)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting alarms: %s", err)
			}
		} else {
			d.Set("alarms", nil)
		}

		if service.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			if err := d.Set("deployment_circuit_breaker", []any{flattenDeploymentCircuitBreaker(service.DeploymentConfiguration.DeploymentCircuitBreaker)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting deployment_circuit_breaker: %s", err)
			}
		} else {
			d.Set("deployment_circuit_breaker", nil)
		}

		if err := d.Set("deployment_configuration", flattenDeploymentConfiguration(service.DeploymentConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting deployment_configuration: %s", err)
		}
	}
	if err := d.Set("deployment_controller", flattenDeploymentController(service.DeploymentController)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_controller: %s", err)
	}
	d.Set("desired_count", service.DesiredCount)
	d.Set("enable_execute_command", service.EnableExecuteCommand)
	d.Set("enable_ecs_managed_tags", service.EnableECSManagedTags)
	d.Set("health_check_grace_period_seconds", service.HealthCheckGracePeriodSeconds)
	// Save IAM role in the same format.
	if service.RoleArn != nil {
		if arn.IsARN(d.Get("iam_role").(string)) {
			d.Set("iam_role", service.RoleArn)
		} else {
			d.Set("iam_role", roleNameFromARN(aws.ToString(service.RoleArn)))
		}
	}
	d.Set("launch_type", service.LaunchType)
	if service.LoadBalancers != nil {
		if err := d.Set("load_balancer", flattenServiceLoadBalancers(service.LoadBalancers)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
		}
	}
	d.Set(names.AttrName, service.ServiceName)
	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}
	if err := d.Set("ordered_placement_strategy", flattenPlacementStrategy(service.PlacementStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ordered_placement_strategy: %s", err)
	}
	if err := d.Set("placement_constraints", flattenServicePlacementConstraints(service.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}
	d.Set("platform_version", service.PlatformVersion)
	d.Set(names.AttrPropagateTags, service.PropagateTags)
	d.Set("scheduling_strategy", service.SchedulingStrategy)
	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}
	// When creating a service that uses the EXTERNAL deployment controller,
	// you can specify only parameters that aren't controlled at the task set level
	// hence TaskDefinition will not be set by aws sdk
	if service.TaskDefinition != nil {
		// Save task definition in the same format.
		if arn.IsARN(d.Get("task_definition").(string)) {
			d.Set("task_definition", service.TaskDefinition)
		} else {
			d.Set("task_definition", familyAndRevisionFromTaskDefinitionARN(aws.ToString(service.TaskDefinition)))
		}
	}
	d.Set(names.AttrTriggers, d.Get(names.AttrTriggers))
	for _, deployment := range service.Deployments {
		if aws.ToString(deployment.Status) == "PRIMARY" {
			if v := deployment.ServiceConnectConfiguration; v != nil {
				if err := d.Set("service_connect_configuration", flattenServiceConnectConfiguration(v)); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting service_connect_configuration: %s", err)
				}
			} else {
				d.Set("service_connect_configuration", nil)
			}
			if v := deployment.VolumeConfigurations; len(v) > 0 {
				if err := d.Set("volume_configuration", flattenServiceVolumeConfigurations(ctx, v)); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting volume_configurations: %s", err)
				}
			}
			if err := d.Set("vpc_lattice_configurations", flattenVPCLatticeConfigurations(deployment.VpcLatticeConfigurations)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting vpc_lattice_configurations: %s", err)
			}
		}
	}

	setTagsOut(ctx, service.Tags)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChangesExcept(names.AttrForceDelete, names.AttrTags, names.AttrTagsAll) {
		cluster := d.Get("cluster").(string)
		input := ecs.UpdateServiceInput{
			Cluster:            aws.String(cluster),
			ForceNewDeployment: d.Get("force_new_deployment").(bool),
			Service:            aws.String(d.Id()),
		}

		if d.HasChange("alarms") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
			}

			if v, ok := d.GetOk("alarms"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.DeploymentConfiguration.Alarms = expandAlarms(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("availability_zone_rebalancing") {
			if v, ok := d.GetOk("availability_zone_rebalancing"); ok {
				input.AvailabilityZoneRebalancing = awstypes.AvailabilityZoneRebalancing(v.(string))
			}
		}

		if d.HasChange("deployment_controller") {
			if input.DeploymentController == nil {
				input.DeploymentController = &awstypes.DeploymentController{}
			}

			if v, ok := d.GetOk("deployment_controller"); ok {
				input.DeploymentController = expandDeploymentController(v.([]any))
			}
		}

		if d.HasChange("deployment_configuration") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
			}

			if v, ok := d.GetOk("deployment_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				config := v.([]any)[0].(map[string]any)

				if strategy, ok := config["strategy"].(string); ok && strategy != "" {
					input.DeploymentConfiguration.Strategy = awstypes.DeploymentStrategy(strategy)

					if v, ok := config["bake_time_in_minutes"].(string); ok {
						bt, err := expandBakeTimeInMinutes(v)
						if err != nil {
							return sdkdiag.AppendFromErr(diags, err)
						}
						input.DeploymentConfiguration.BakeTimeInMinutes = bt
					}

					if awstypes.DeploymentStrategy(strategy) == awstypes.DeploymentStrategyLinear {
						if v, ok := config["linear_configuration"].([]any); ok && len(v) > 0 {
							if linearConfig, ok := v[0].(map[string]any); ok {
								sp, sbtm, err := expandLinearConfiguration(linearConfig)
								if err != nil {
									return sdkdiag.AppendFromErr(diags, err)
								}
								input.DeploymentConfiguration.LinearConfiguration = &awstypes.LinearConfiguration{
									StepPercent:           sp,
									StepBakeTimeInMinutes: sbtm,
								}
							}
						}
					}

					if awstypes.DeploymentStrategy(strategy) == awstypes.DeploymentStrategyCanary {
						if v, ok := config["canary_configuration"].([]any); ok && len(v) > 0 {
							if canaryConfig, ok := v[0].(map[string]any); ok {
								cp, cbtm, err := expandCanaryConfiguration(canaryConfig)
								if err != nil {
									return sdkdiag.AppendFromErr(diags, err)
								}
								input.DeploymentConfiguration.CanaryConfiguration = &awstypes.CanaryConfiguration{
									CanaryPercent:           cp,
									CanaryBakeTimeInMinutes: cbtm,
								}
							}
						}
					}
				}
				if hooks := config["lifecycle_hook"].(*schema.Set).List(); len(hooks) > 0 {
					input.DeploymentConfiguration.LifecycleHooks = expandLifecycleHooks(hooks)
				}
			}
		}

		if d.HasChange(names.AttrCapacityProviderStrategy) {
			input.CapacityProviderStrategy = expandCapacityProviderStrategyItems(d.Get(names.AttrCapacityProviderStrategy).(*schema.Set))
		}

		if d.HasChange("deployment_circuit_breaker") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
			}

			// To remove an existing deployment circuit breaker, specify an empty object.
			input.DeploymentConfiguration.DeploymentCircuitBreaker = &awstypes.DeploymentCircuitBreaker{}

			if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]any)[0].(map[string]any))
			}
		}

		switch schedulingStrategy := awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string)); schedulingStrategy {
		case awstypes.SchedulingStrategyDaemon:
			if d.HasChange("deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(d.Get("deployment_minimum_healthy_percent").(int)))
			}
		case awstypes.SchedulingStrategyReplica:
			if d.HasChanges("deployment_maximum_percent", "deployment_minimum_healthy_percent") {
				if input.DeploymentConfiguration == nil {
					input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
				}

				input.DeploymentConfiguration.MaximumPercent = aws.Int32(int32(d.Get("deployment_maximum_percent").(int)))
				input.DeploymentConfiguration.MinimumHealthyPercent = aws.Int32(int32(d.Get("deployment_minimum_healthy_percent").(int)))
			}

			if d.HasChange("desired_count") {
				input.DesiredCount = aws.Int32(int32(d.Get("desired_count").(int)))
			}
		}

		if d.HasChange("enable_ecs_managed_tags") {
			input.EnableECSManagedTags = aws.Bool(d.Get("enable_ecs_managed_tags").(bool))
		}

		if d.HasChange("enable_execute_command") {
			input.EnableExecuteCommand = aws.Bool(d.Get("enable_execute_command").(bool))
		}

		if d.HasChange("health_check_grace_period_seconds") {
			input.HealthCheckGracePeriodSeconds = aws.Int32(int32(d.Get("health_check_grace_period_seconds").(int)))
		}

		if d.HasChange("load_balancer") {
			if v, ok := d.Get("load_balancer").(*schema.Set); ok && v != nil {
				input.LoadBalancers = expandServiceLoadBalancers(v.List())
			}
		}

		if d.HasChange(names.AttrNetworkConfiguration) {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]any))
		}

		if d.HasChange("ordered_placement_strategy") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementStrategy
			// To remove an existing placement strategy, specify an empty object.
			input.PlacementStrategy = []awstypes.PlacementStrategy{}

			if v, ok := d.GetOk("ordered_placement_strategy"); ok && len(v.([]any)) > 0 {
				apiObject, err := expandPlacementStrategy(v.([]any))
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				input.PlacementStrategy = apiObject
			}
		}

		if d.HasChange("placement_constraints") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementConstraints
			// To remove all existing placement constraints, specify an empty array.
			input.PlacementConstraints = []awstypes.PlacementConstraint{}

			if v, ok := d.Get("placement_constraints").(*schema.Set); ok && v.Len() > 0 {
				apiObject, err := expandPlacementConstraints(v.List())
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				input.PlacementConstraints = apiObject
			}
		}

		if d.HasChange("platform_version") {
			input.PlatformVersion = aws.String(d.Get("platform_version").(string))
		}

		if d.HasChange(names.AttrPropagateTags) {
			input.PropagateTags = awstypes.PropagateTags(d.Get(names.AttrPropagateTags).(string))
		}

		if d.HasChange("service_connect_configuration") {
			input.ServiceConnectConfiguration = expandServiceConnectConfiguration(d.Get("service_connect_configuration").([]any))
		}

		if d.HasChange("service_registries") {
			input.ServiceRegistries = expandServiceRegistries(d.Get("service_registries").([]any))
		}

		if d.HasChange("task_definition") {
			input.TaskDefinition = aws.String(d.Get("task_definition").(string))
		}

		if d.HasChange("volume_configuration") {
			input.VolumeConfigurations = expandServiceVolumeConfigurations(ctx, d.Get("volume_configuration").([]any))
		}

		if d.HasChange("vpc_lattice_configurations") {
			input.VpcLatticeConfigurations = expandVPCLatticeConfiguration(d.Get("vpc_lattice_configurations").(*schema.Set))
		}

		// Retry due to IAM eventual consistency.
		const (
			serviceUpdateTimeout = 2 * time.Minute
			timeout              = propagationTimeout + serviceUpdateTimeout
		)
		operationTime := time.Now().UTC()
		_, err := tfresource.RetryWhen(ctx, timeout,
			func(ctx context.Context) (any, error) {
				return conn.UpdateService(ctx, &input)
			},
			func(err error) (bool, error) {
				if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "verify that the ECS service role being passed has the proper permissions") {
					return true, err
				}

				if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "does not have an associated load balancer") {
					return true, err
				}

				return false, err
			},
		)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Service (%s): %s", d.Id(), err)
		}

		if d.Get("wait_for_steady_state").(bool) {
			if _, err := waitServiceStable(ctx, conn, d.Id(), cluster, operationTime, d.Get("sigint_rollback").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) update: %s", d.Id(), err)
			}
		} else if _, err := waitServiceActive(ctx, conn, d.Id(), cluster, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster := d.Get("cluster").(string)
	service, err := findServiceNoTagsByTwoPartKey(ctx, conn, d.Id(), cluster)

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Service (%s): %s", d.Id(), err)
	}

	status := aws.ToString(service.Status)
	if status == serviceStatusInactive {
		return diags
	}

	forceDelete := d.Get(names.AttrForceDelete).(bool)

	// Drain the ECS service.
	if status != serviceStatusDraining && service.SchedulingStrategy != awstypes.SchedulingStrategyDaemon && !forceDelete {
		input := &ecs.UpdateServiceInput{
			Cluster:      aws.String(cluster),
			DesiredCount: aws.Int32(0),
			Service:      aws.String(d.Id()),
		}

		_, err := conn.UpdateService(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "draining ECS Service (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting ECS Service: %s", d.Id())
	_, err = tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutDelete),
		func(ctx context.Context) (any, error) {
			return conn.DeleteService(ctx, &ecs.DeleteServiceInput{
				Cluster: aws.String(cluster),
				Force:   aws.Bool(forceDelete),
				Service: aws.String(d.Id()),
			})
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "The service cannot be stopped while deployments are active.") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, errCodeDependencyViolation, "has a dependent object") {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Service (%s): %s", d.Id(), err)
	}

	if _, err := waitServiceInactive(ctx, conn, d.Id(), cluster, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceServiceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	if err := importer.Import(ctx, d, meta); err != nil {
		return nil, err
	}

	clusterName, serviceName, err := parseServiceImportID(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	log.Printf("[DEBUG] Importing ECS service %s from cluster %s", serviceName, clusterName)

	region := d.Get(names.AttrRegion).(string)

	clusterARN := arn.ARN{
		Partition: names.PartitionForRegion(region).ID(),
		Region:    region,
		Service:   "ecs",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("cluster/%s", clusterName),
	}.String()
	d.Set("cluster", clusterARN)

	serviceARN := arn.ARN{
		Partition: names.PartitionForRegion(region).ID(),
		Region:    region,
		Service:   "ecs",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("service/%s/%s", clusterName, serviceName),
	}.String()
	d.SetId(serviceARN)
	d.Set(names.AttrARN, serviceARN)

	return []*schema.ResourceData{d}, nil
}

func retryServiceCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	const (
		serviceCreateTimeout = 2 * time.Minute
		timeout              = propagationTimeout + serviceCreateTimeout
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateService(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.ClusterNotFoundException](err) {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "verify that the ECS service role being passed has the proper permissions") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "does not have an associated load balancer") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Unable to assume the service linked role") {
				return true, err
			}

			return false, err
		},
	)
	if err != nil {
		return nil, err
	}

	return outputRaw.(*ecs.CreateServiceOutput), err
}

func findService(ctx context.Context, conn *ecs.Client, input *ecs.DescribeServicesInput) (*awstypes.Service, error) {
	output, err := findServices(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServices(ctx context.Context, conn *ecs.Client, input *ecs.DescribeServicesInput) ([]awstypes.Service, error) {
	output, err := conn.DescribeServices(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	// When an ECS Service is not found by DescribeServices(), it will return a Failure struct with Reason = "MISSING"
	for _, v := range output.Failures {
		if aws.ToString(v.Reason) == failureReasonMissing {
			return nil, &sdkretry.NotFoundError{
				LastError:   failureError(&v),
				LastRequest: input,
			}
		}
	}

	return output.Services, nil
}

func findServiceByTwoPartKey(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(clusterNameOrARN),
		Include:  []awstypes.ServiceField{awstypes.ServiceFieldTags},
		Services: []string{serviceName},
	}

	output, err := findService(ctx, conn, input)

	// Some partitions (i.e., ISO) may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(partitionFromConn(conn), err) {
		input.Include = nil

		output, err = findService(ctx, conn, input)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findServiceNoTagsByTwoPartKey(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Services: []string{serviceName},
	}
	if clusterNameOrARN != "" {
		input.Cluster = aws.String(clusterNameOrARN)
	}

	return findService(ctx, conn, input)
}

type expectServiceActiveError struct {
	status string
}

func newExpectServiceActiveError(status string) *expectServiceActiveError {
	return &expectServiceActiveError{
		status: status,
	}
}

func (e *expectServiceActiveError) Error() string {
	return fmt.Sprintf("expected status %[1]q, was %[2]q", serviceStatusActive, e.status)
}

func findServiceByTwoPartKeyWaitForActive(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) (*awstypes.Service, error) {
	var service *awstypes.Service

	// Use the tfresource.Retry function instead of WaitForState() because we don't want the timeout error, if any.
	const (
		timeout = 2 * time.Minute
	)
	err := tfresource.Retry(ctx, timeout, func(ctx context.Context) *tfresource.RetryError {
		var err error

		service, err = findServiceByTwoPartKey(ctx, conn, serviceName, clusterNameOrARN)

		if retry.NotFound(err) {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		if status := aws.ToString(service.Status); status != serviceStatusActive {
			return tfresource.RetryableError(newExpectServiceActiveError(status))
		}

		return nil
	})

	if errs.IsA[*expectServiceActiveError](err) {
		return nil, &sdkretry.NotFoundError{
			LastError: err,
		}
	}

	return service, err
}

const (
	serviceStatusInactive = "INACTIVE"
	serviceStatusActive   = "ACTIVE"
	serviceStatusDraining = "DRAINING"

	// Non-standard statuses for statusServiceWaitForStable().
	serviceStatusPending = "tfPENDING"
	serviceStatusStable  = "tfSTABLE"
)

var deploymentTerminalStates = enum.Slice(
	awstypes.ServiceDeploymentStatusSuccessful,
	awstypes.ServiceDeploymentStatusStopped,
	awstypes.ServiceDeploymentStatusRollbackFailed,
	awstypes.ServiceDeploymentStatusRollbackSuccessful,
)

func statusService(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findServiceNoTagsByTwoPartKey(ctx, conn, serviceName, clusterNameOrARN)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), err
	}
}

func statusServiceWaitForStable(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, sigintConfig *rollbackState, operationTime time.Time) sdkretry.StateRefreshFunc {
	var primaryTaskSet *awstypes.Deployment
	var primaryDeploymentArn *string
	var isNewPrimaryDeployment bool

	return func() (any, string, error) {
		outputRaw, serviceStatus, err := statusService(ctx, conn, serviceName, clusterNameOrARN)()
		if err != nil {
			return nil, "", err
		}

		if serviceStatus != serviceStatusActive {
			return outputRaw, serviceStatus, nil
		}

		output := outputRaw.(*awstypes.Service)

		if primaryTaskSet == nil {
			primaryTaskSet = findPrimaryTaskSet(output.Deployments)

			if primaryTaskSet != nil && primaryTaskSet.CreatedAt != nil {
				createdAtUTC := primaryTaskSet.CreatedAt.UTC()
				isNewPrimaryDeployment = createdAtUTC.After(operationTime)
			}
		}

		isNewECSDeployment := output.DeploymentController != nil &&
			output.DeploymentController.Type == awstypes.DeploymentControllerTypeEcs &&
			isNewPrimaryDeployment

		// For new deployments with ECS deployment controller, check the deployment status
		if isNewECSDeployment {
			if primaryDeploymentArn == nil {
				serviceArn := aws.ToString(output.ServiceArn)

				var err error
				primaryDeploymentArn, err = findPrimaryDeploymentARN(ctx, conn, primaryTaskSet, serviceArn, clusterNameOrARN, operationTime)
				if err != nil {
					return nil, "", err
				}
				if primaryDeploymentArn == nil {
					return output, serviceStatusPending, nil
				}
			}

			if sigintConfig.rollbackConfigured && !sigintConfig.rollbackRoutineStarted {
				sigintConfig.waitGroup.Add(1)
				go rollbackRoutine(ctx, conn, sigintConfig, primaryDeploymentArn)
				sigintConfig.rollbackRoutineStarted = true
			}

			deploymentStatus, err := findDeploymentStatus(ctx, conn, *primaryDeploymentArn)
			if err != nil {
				return nil, "", err
			}
			return output, deploymentStatus, nil
		}

		// For other deployment controllers or in-place updates, check based on desired count
		if n, dc, rc := len(output.Deployments), output.DesiredCount, output.RunningCount; n == 1 && dc == rc {
			serviceStatus = serviceStatusStable
		} else {
			serviceStatus = serviceStatusPending
		}

		return output, serviceStatus, nil
	}
}

func findPrimaryTaskSet(deployments []awstypes.Deployment) *awstypes.Deployment {
	for _, deployment := range deployments {
		if aws.ToString(deployment.Status) == taskSetStatusPrimary {
			return &deployment
		}
	}
	return nil
}

func findPrimaryDeploymentARN(ctx context.Context, conn *ecs.Client, primaryTaskSet *awstypes.Deployment, serviceNameOrARN, clusterNameOrARN string, operationTime time.Time) (*string, error) {
	parts := strings.Split(aws.ToString(primaryTaskSet.Id), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid primary task set ID format: %s", aws.ToString(primaryTaskSet.Id))
	}
	taskSetID := parts[1]

	input := ecs.ListServiceDeploymentsInput{
		Cluster: aws.String(clusterNameOrARN),
		Service: aws.String(serviceNameFromARN(serviceNameOrARN)),
		CreatedAt: &awstypes.CreatedAt{
			After: &operationTime,
		},
	}

	deployments, err := findServiceDeploymentBriefs(ctx, conn, &input)
	if err != nil {
		return nil, err
	}

	// Find deployment matching task set
	for _, deployment := range deployments {
		if strings.Contains(aws.ToString(deployment.TargetServiceRevisionArn), taskSetID) {
			return deployment.ServiceDeploymentArn, nil
		}
	}

	return nil, nil
}

func findDeploymentStatus(ctx context.Context, conn *ecs.Client, deploymentARN string) (string, error) {
	input := ecs.DescribeServiceDeploymentsInput{
		ServiceDeploymentArns: []string{deploymentARN},
	}

	output, err := findServiceDeployments(ctx, conn, &input)

	if err != nil {
		return "", err
	}

	if len(output) == 0 {
		return serviceStatusPending, nil
	}

	switch deployment := output[0]; deployment.Status {
	case awstypes.ServiceDeploymentStatusSuccessful:
		return serviceStatusStable, nil
	case awstypes.ServiceDeploymentStatusInProgress:
		return serviceStatusPending, nil
	case awstypes.ServiceDeploymentStatusStopped,
		awstypes.ServiceDeploymentStatusRollbackFailed,
		awstypes.ServiceDeploymentStatusRollbackSuccessful:
		message := "Deployment failed"
		if deployment.StatusReason != nil {
			message = aws.ToString(deployment.StatusReason)
		}
		return "", errors.New(message)
	default:
		return serviceStatusPending, nil
	}
}

func findServiceDeployments(ctx context.Context, conn *ecs.Client, input *ecs.DescribeServiceDeploymentsInput) ([]awstypes.ServiceDeployment, error) {
	output, err := conn.DescribeServiceDeployments(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ServiceDeployments, nil
}

func findServiceDeploymentBriefs(ctx context.Context, conn *ecs.Client, input *ecs.ListServiceDeploymentsInput) ([]awstypes.ServiceDeploymentBrief, error) {
	var output []awstypes.ServiceDeploymentBrief

	err := listServiceDeploymentsPages(ctx, conn, input, func(page *ecs.ListServiceDeploymentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ServiceDeployments...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ClusterNotFoundException](err) || errs.IsA[*awstypes.ServiceNotFoundException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type rollbackState struct {
	rollbackConfigured     bool
	rollbackRoutineStarted bool
	rollbackRoutineStopped chan struct{}
	waitGroup              sync.WaitGroup
}

func rollbackRoutine(ctx context.Context, conn *ecs.Client, rollbackState *rollbackState, primaryDeploymentArn *string) {
	defer rollbackState.waitGroup.Done()

	select {
	case <-ctx.Done():
		log.Printf("[INFO] SIGINT detected. Initiating rollback for deployment: %s", *primaryDeploymentArn)
		ctx, cancel := context.WithTimeout(context.Background(), (1 * time.Hour)) // Maximum time before SIGKILL
		defer cancel()

		if err := rollbackDeployment(ctx, conn, primaryDeploymentArn); err != nil { //nolint:contextcheck // Original Context has been cancelled
			log.Printf("[ERROR] Failed to rollback deployment: %s. Err: %s", *primaryDeploymentArn, err)
		} else {
			log.Printf("[INFO] Deployment: %s rolled back successfully.", *primaryDeploymentArn)
		}

	case <-rollbackState.rollbackRoutineStopped:
		return
	}
}

func rollbackDeployment(ctx context.Context, conn *ecs.Client, primaryDeploymentArn *string) error {
	// Check if deployment is already in terminal state, meaning rollback is not needed
	deploymentStatus, err := findDeploymentStatus(ctx, conn, *primaryDeploymentArn)
	if err != nil {
		return err
	}
	if slices.Contains(deploymentTerminalStates, deploymentStatus) {
		return nil
	}

	log.Printf("[INFO] Rolling back deployment %s. This may take a few minutes...", *primaryDeploymentArn)

	input := &ecs.StopServiceDeploymentInput{
		ServiceDeploymentArn: primaryDeploymentArn,
		StopType:             awstypes.StopServiceDeploymentStopTypeRollback,
	}

	_, err = conn.StopServiceDeployment(ctx, input)
	if err != nil {
		return err
	}

	return waitForDeploymentTerminalStatus(ctx, conn, *primaryDeploymentArn)
}

func waitForDeploymentTerminalStatus(ctx context.Context, conn *ecs.Client, primaryDeploymentArn string) error {
	stateConf := &sdkretry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ServiceDeploymentStatusPending,
			awstypes.ServiceDeploymentStatusInProgress,
			awstypes.ServiceDeploymentStatusRollbackRequested,
			awstypes.ServiceDeploymentStatusRollbackInProgress,
		),
		Target: deploymentTerminalStates,
		Refresh: func() (any, string, error) {
			status, err := findDeploymentStatus(ctx, conn, primaryDeploymentArn)
			return nil, status, err
		},
		Timeout: 1 * time.Hour, // Maximum time before SIGKILL
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// waitServiceStable waits for an ECS Service to reach the status "ACTIVE" and have all desired tasks running.
// Does not return tags.
func waitServiceStable(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, operationTime time.Time, sigintCancellation bool, timeout time.Duration) (*awstypes.Service, error) { //nolint:unparam
	sigintConfig := &rollbackState{
		rollbackConfigured:     sigintCancellation,
		rollbackRoutineStarted: false,
		rollbackRoutineStopped: make(chan struct{}),
		waitGroup:              sync.WaitGroup{},
	}

	stateConf := &sdkretry.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining, serviceStatusPending},
		Target:  []string{serviceStatusStable},
		Refresh: statusServiceWaitForStable(ctx, conn, serviceName, clusterNameOrARN, sigintConfig, operationTime),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if sigintConfig.rollbackRoutineStarted {
		close(sigintConfig.rollbackRoutineStopped)
		sigintConfig.waitGroup.Wait()
	}

	if output, ok := outputRaw.(*awstypes.Service); ok {
		return output, err
	}

	return nil, err
}

// Does not return tags.
func waitServiceActive(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, timeout time.Duration) (*awstypes.Service, error) { //nolint:unparam
	stateConf := &sdkretry.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining},
		Target:  []string{serviceStatusActive},
		Refresh: statusService(ctx, conn, serviceName, clusterNameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Service); ok {
		return output, err
	}

	return nil, err
}

// Does not return tags.
func waitServiceInactive(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, timeout time.Duration) (*awstypes.Service, error) {
	stateConf := &sdkretry.StateChangeConf{
		Pending:    []string{serviceStatusActive, serviceStatusDraining},
		Target:     []string{serviceStatusInactive},
		Refresh:    statusService(ctx, conn, serviceName, clusterNameOrARN),
		Timeout:    timeout,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Service); ok {
		return output, err
	}

	return nil, err
}

func triggersCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	// clears diff to avoid extraneous diffs but lets it pass for triggering update
	fnd := false
	if v, ok := d.GetOk("force_new_deployment"); ok {
		fnd = v.(bool)
	}

	if d.HasChange(names.AttrTriggers) && !fnd {
		return d.Clear(names.AttrTriggers)
	}

	if d.HasChange(names.AttrTriggers) && fnd {
		o, n := d.GetChange(names.AttrTriggers)
		if len(o.(map[string]any)) > 0 && len(n.(map[string]any)) == 0 {
			return d.Clear(names.AttrTriggers)
		}

		return nil
	}

	return nil
}

func capacityProviderStrategyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	// This if-statement is true only when the resource is being updated.
	// d.Id() != "" means the resource (ecs service) already exists.
	if d.Id() != "" && d.HasChange(names.AttrCapacityProviderStrategy) {
		if v := d.Get("force_new_deployment").(bool); !v {
			return fmt.Errorf("force_new_deployment should be true when capacity_provider_strategy is being updated")
		}
	}
	return nil
}

func expandAlarms(tfMap map[string]any) *awstypes.DeploymentAlarms {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DeploymentAlarms{}

	if v, ok := tfMap["enable"].(bool); ok {
		apiObject.Enable = v
	}

	if v, ok := tfMap["rollback"].(bool); ok {
		apiObject.Rollback = v
	}

	if v, ok := tfMap["alarm_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AlarmNames = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenAlarms(apiObject *awstypes.DeploymentAlarms) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AlarmNames; v != nil {
		tfMap["alarm_names"] = v
	}

	tfMap["enable"] = apiObject.Enable

	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func expandDeploymentController(l []any) *awstypes.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	deploymentController := &awstypes.DeploymentController{
		Type: awstypes.DeploymentControllerType(m[names.AttrType].(string)),
	}

	return deploymentController
}

func flattenDeploymentController(apiObject *awstypes.DeploymentController) []any {
	tfMap := map[string]any{
		names.AttrType: awstypes.DeploymentControllerTypeEcs,
	}

	if apiObject == nil {
		return []any{tfMap}
	}

	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}

func expandDeploymentCircuitBreaker(tfMap map[string]any) *awstypes.DeploymentCircuitBreaker {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DeploymentCircuitBreaker{}

	apiObject.Enable = tfMap["enable"].(bool)
	apiObject.Rollback = tfMap["rollback"].(bool)

	return apiObject
}

func flattenDeploymentCircuitBreaker(apiObject *awstypes.DeploymentCircuitBreaker) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["enable"] = apiObject.Enable
	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func flattenDeploymentConfiguration(apiObject *awstypes.DeploymentConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.BakeTimeInMinutes; v != nil {
		tfMap["bake_time_in_minutes"] = flex.Int32ToStringValue(v)
	}

	if v := apiObject.CanaryConfiguration; v != nil {
		tfMap["canary_configuration"] = flattenCanaryConfiguration(v)
	}

	if v := apiObject.LinearConfiguration; v != nil {
		tfMap["linear_configuration"] = flattenLinearConfiguration(v)
	}

	if v := apiObject.LifecycleHooks; len(v) > 0 {
		tfMap["lifecycle_hook"] = flattenLifecycleHooks(v)
	}

	if v := apiObject.Strategy; v != "" {
		tfMap["strategy"] = v
	}

	if len(tfMap) == 0 {
		return nil
	}

	return []any{tfMap}
}

func flattenLifecycleHooks(apiObjects []awstypes.DeploymentLifecycleHook) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.HookDetails; v != nil {
			if jsonString, err := smithy.DocumentToJSONString(v); err == nil {
				tfMap["hook_details"] = jsonString
			}
		}

		if v := apiObject.HookTargetArn; v != nil {
			tfMap["hook_target_arn"] = aws.ToString(v)
		}

		if v := apiObject.RoleArn; v != nil {
			tfMap[names.AttrRoleARN] = aws.ToString(v)
		}

		if v := apiObject.LifecycleStages; len(v) > 0 {
			tfMap["lifecycle_stages"] = v
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCanaryConfiguration(apiObject *awstypes.CanaryConfiguration) []map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.CanaryBakeTimeInMinutes; v != nil {
		tfMap["canary_bake_time_in_minutes"] = flex.Int32ToStringValue(v)
	}

	tfMap["canary_percent"] = aws.ToFloat64(apiObject.CanaryPercent)

	return []map[string]any{tfMap}
}

func flattenLinearConfiguration(apiObject *awstypes.LinearConfiguration) []map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.StepBakeTimeInMinutes; v != nil {
		tfMap["step_bake_time_in_minutes"] = flex.Int32ToStringValue(v)
	}

	tfMap["step_percent"] = aws.ToFloat64(apiObject.StepPercent)

	return []map[string]any{tfMap}
}

func expandLifecycleHooks(tfList []any) []awstypes.DeploymentLifecycleHook {
	apiObject := make([]awstypes.DeploymentLifecycleHook, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]any)

		hook := awstypes.DeploymentLifecycleHook{}

		if v, ok := tfMap["hook_target_arn"].(string); ok && v != "" {
			hook.HookTargetArn = aws.String(v)
		}

		if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
			hook.RoleArn = aws.String(v)
		}

		if v, ok := tfMap["lifecycle_stages"].([]any); ok && len(v) > 0 {
			stages := make([]awstypes.DeploymentLifecycleHookStage, 0, len(v))
			for _, stage := range v {
				if stageStr, ok := stage.(string); ok && stageStr != "" {
					stages = append(stages, awstypes.DeploymentLifecycleHookStage(stageStr))
				}
			}
			hook.LifecycleStages = stages
		}

		if v, ok := tfMap["hook_details"].(string); ok && v != "" {
			var jsonValue any
			if err := json.Unmarshal([]byte(v), &jsonValue); err == nil {
				hook.HookDetails = document.NewLazyDocument(jsonValue)
			}
		}

		apiObject = append(apiObject, hook)
	}

	return apiObject
}

func expandBakeTimeInMinutes(bakeTimeStr string) (*int32, error) {
	var ptrBakeTimeRet *int32

	bakeTime := nullable.Int(bakeTimeStr)
	if !bakeTime.IsNull() {
		value, _, err := bakeTime.ValueInt32()
		if err != nil {
			return nil, err
		}
		ptrBakeTimeRet = aws.Int32(value)
	}

	return ptrBakeTimeRet, nil
}

func expandCanaryConfiguration(canaryConfig map[string]any) (*float64, *int32, error) {
	var canaryPercentRet *float64
	var ptrCanaryBakeTimeRet *int32

	if cp, ok := canaryConfig["canary_percent"].(float64); ok {
		canaryPercentRet = aws.Float64(cp)
	} else {
		return nil, nil, fmt.Errorf("canary_percent is required for canary deployment configuration")
	}
	if cbtm, ok := canaryConfig["canary_bake_time_in_minutes"].(string); ok {
		canaryBakeTimeInMinutes := nullable.Int(cbtm)
		value, _, err := canaryBakeTimeInMinutes.ValueInt32()
		if err != nil {
			return nil, nil, err
		}
		ptrCanaryBakeTimeRet = aws.Int32(value)
	}

	return canaryPercentRet, ptrCanaryBakeTimeRet, nil
}

func expandLinearConfiguration(linearConfig map[string]any) (*float64, *int32, error) {
	var stepPercentRet *float64
	var ptrStepBakeTimeRet *int32

	if sp, ok := linearConfig["step_percent"].(float64); ok {
		stepPercentRet = aws.Float64(sp)
	} else {
		return nil, nil, fmt.Errorf("step_percent is required for linear deployment configuration")
	}
	if sbtm, ok := linearConfig["step_bake_time_in_minutes"].(string); ok {
		stepBakeTimeInMinutes := nullable.Int(sbtm)
		value, _, err := stepBakeTimeInMinutes.ValueInt32()
		if err != nil {
			return nil, nil, err
		}
		ptrStepBakeTimeRet = aws.Int32(value)
	}

	return stepPercentRet, ptrStepBakeTimeRet, nil
}

func flattenNetworkConfiguration(nc *awstypes.NetworkConfiguration) []any {
	if nc == nil {
		return nil
	}

	result := make(map[string]any)
	result[names.AttrSecurityGroups] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.SecurityGroups)
	result[names.AttrSubnets] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.Subnets)

	result["assign_public_ip"] = nc.AwsvpcConfiguration.AssignPublicIp == awstypes.AssignPublicIpEnabled

	return []any{result}
}

func expandNetworkConfiguration(nc []any) *awstypes.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &awstypes.AwsVpcConfiguration{}
	raw := nc[0].(map[string]any)
	if val, ok := raw[names.AttrSecurityGroups]; ok {
		awsVpcConfig.SecurityGroups = flex.ExpandStringValueSet(val.(*schema.Set))
	}
	awsVpcConfig.Subnets = flex.ExpandStringValueSet(raw[names.AttrSubnets].(*schema.Set))
	if val, ok := raw["assign_public_ip"].(bool); ok {
		awsVpcConfig.AssignPublicIp = awstypes.AssignPublicIpDisabled
		if val {
			awsVpcConfig.AssignPublicIp = awstypes.AssignPublicIpEnabled
		}
	}

	return &awstypes.NetworkConfiguration{AwsvpcConfiguration: awsVpcConfig}
}

func expandVPCLatticeConfiguration(tfSet *schema.Set) []awstypes.VpcLatticeConfiguration {
	apiObjects := make([]awstypes.VpcLatticeConfiguration, 0)

	for _, tfMapRaw := range tfSet.List() {
		config := tfMapRaw.(map[string]any)

		apiObject := awstypes.VpcLatticeConfiguration{
			RoleArn:        aws.String(config[names.AttrRoleARN].(string)),
			TargetGroupArn: aws.String(config["target_group_arn"].(string)),
			PortName:       aws.String(config["port_name"].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenVPCLatticeConfigurations(apiObjects []awstypes.VpcLatticeConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrRoleARN:  aws.ToString(apiObject.RoleArn),
			"target_group_arn": aws.ToString(apiObject.TargetGroupArn),
			"port_name":        aws.ToString(apiObject.PortName),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandPlacementConstraints(tfList []any) ([]awstypes.PlacementConstraint, error) {
	if len(tfList) == 0 {
		return nil, nil
	}

	var result []awstypes.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.PlacementConstraint{}

		if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
			apiObject.Expression = aws.String(v)
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type = awstypes.PlacementConstraintType(v)
		}

		if err := validPlacementConstraint(string(apiObject.Type), aws.ToString(apiObject.Expression)); err != nil {
			return result, err
		}

		result = append(result, apiObject)
	}

	return result, nil
}

func flattenServicePlacementConstraints(apiObjects []awstypes.PlacementConstraint) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.Expression != nil {
			tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
		}
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandPlacementStrategy(s []any) ([]awstypes.PlacementStrategy, error) {
	if len(s) == 0 {
		return nil, nil
	}
	pss := make([]awstypes.PlacementStrategy, 0)
	for _, raw := range s {
		p, ok := raw.(map[string]any)

		if !ok {
			continue
		}

		t, ok := p[names.AttrType].(string)

		if !ok {
			return nil, fmt.Errorf("missing type attribute in placement strategy configuration block")
		}

		f, ok := p[names.AttrField].(string)

		if !ok {
			return nil, fmt.Errorf("missing field attribute in placement strategy configuration block")
		}

		if err := validPlacementStrategy(t, f); err != nil {
			return nil, err
		}
		ps := awstypes.PlacementStrategy{
			Type: awstypes.PlacementStrategyType(t),
		}
		if f != "" {
			// Field must be omitted (i.e. not empty string) for random strategy
			ps.Field = aws.String(f)
		}
		pss = append(pss, ps)
	}
	return pss, nil
}

func flattenPlacementStrategy(apiObjects []awstypes.PlacementStrategy) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]any)

		if apiObject.Field != nil {
			field := aws.ToString(apiObject.Field)
			tfMap[names.AttrField] = field
			// for some fields the API requires lowercase for creation but will return uppercase on query
			if field == "MEMORY" || field == "CPU" {
				tfMap[names.AttrField] = strings.ToLower(field)
			}
		}
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceConnectConfiguration(tfList []any) *awstypes.ServiceConnectConfiguration {
	if len(tfList) == 0 {
		return &awstypes.ServiceConnectConfiguration{
			Enabled: false,
		}
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ServiceConnectConfiguration{}

	if v, ok := tfMap["access_log_configuration"].([]any); ok && len(v) > 0 {
		apiObject.AccessLogConfiguration = expandServiceConnectAccessLogConfiguration(v)
	}
	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}
	if v, ok := tfMap["log_configuration"].([]any); ok && len(v) > 0 {
		apiObject.LogConfiguration = expandLogConfiguration(v)
	}
	if v, ok := tfMap[names.AttrNamespace].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}
	if v, ok := tfMap["service"].([]any); ok && len(v) > 0 {
		apiObject.Services = expandServiceConnectServices(v)
	}

	return apiObject
}

func flattenServiceConnectConfiguration(apiObject *awstypes.ServiceConnectConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.AccessLogConfiguration; v != nil {
		tfMap["access_log_configuration"] = flattenServiceConnectAccessLogConfiguration(v)
	}
	if v := apiObject.LogConfiguration; v != nil {
		tfMap["log_configuration"] = []any{flattenLogConfiguration(*v)}
	}
	if v := apiObject.Namespace; v != nil {
		tfMap[names.AttrNamespace] = aws.ToString(v)
	}
	if v := apiObject.Services; v != nil {
		tfMap["service"] = flattenServiceConnectServices(v)
	}

	return []any{tfMap}
}

func expandServiceConnectAccessLogConfiguration(tfList []any) *awstypes.ServiceConnectAccessLogConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	config := &awstypes.ServiceConnectAccessLogConfiguration{}

	if v, ok := tfMap[names.AttrFormat].(string); ok && v != "" {
		config.Format = awstypes.ServiceConnectAccessLoggingFormat(v)
	}

	if v, ok := tfMap["include_query_parameters"].(string); ok && v != "" {
		config.IncludeQueryParameters = awstypes.ServiceConnectIncludeQueryParameters(v)
	}

	return config
}

func flattenServiceConnectAccessLogConfiguration(config *awstypes.ServiceConnectAccessLogConfiguration) []any {
	if config == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrFormat:           config.Format,
		"include_query_parameters": config.IncludeQueryParameters,
	}

	return []any{tfMap}
}

func expandLogConfiguration(tfList []any) *awstypes.LogConfiguration {
	if len(tfList) == 0 {
		return &awstypes.LogConfiguration{}
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.LogConfiguration{}

	if v, ok := tfMap["log_driver"].(string); ok && v != "" {
		apiObject.LogDriver = awstypes.LogDriver(v)
	}
	if v, ok := tfMap["options"].(map[string]any); ok && len(v) > 0 {
		apiObject.Options = flex.ExpandStringValueMap(v)
	}
	if v, ok := tfMap["secret_option"].([]any); ok && len(v) > 0 {
		apiObject.SecretOptions = expandSecrets(v)
	}

	return apiObject
}

func flattenLogConfiguration(apiObject awstypes.LogConfiguration) map[string]any {
	tfMap := map[string]any{
		"log_driver": apiObject.LogDriver,
	}

	if v := apiObject.Options; v != nil {
		tfMap["options"] = v
	}
	if v := apiObject.SecretOptions; v != nil {
		tfMap["secret_option"] = flattenSecrets(v)
	}

	return tfMap
}

func expandSecrets(tfList []any) []awstypes.Secret {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Secret

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		var apiObject awstypes.Secret

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}
		if v, ok := tfMap["value_from"].(string); ok && v != "" {
			apiObject.ValueFrom = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenSecrets(apiObjects []awstypes.Secret) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.Name; v != nil {
			tfMap[names.AttrName] = aws.ToString(v)
		}
		if v := apiObject.ValueFrom; v != nil {
			tfMap["value_from"] = aws.ToString(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceLoadBalancers(tfList []any) []awstypes.LoadBalancer {
	apiObjects := make([]awstypes.LoadBalancer, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.LoadBalancer{
			ContainerName: aws.String(tfMap["container_name"].(string)),
			ContainerPort: aws.Int32(int32(tfMap["container_port"].(int))),
		}

		if v, ok := tfMap["elb_name"]; ok && v.(string) != "" {
			apiObject.LoadBalancerName = aws.String(v.(string))
		}

		if v, ok := tfMap["target_group_arn"]; ok && v.(string) != "" {
			apiObject.TargetGroupArn = aws.String(v.(string))
		}

		if advConfig, ok := tfMap["advanced_configuration"].([]any); ok && len(advConfig) > 0 && advConfig[0] != nil {
			config := advConfig[0].(map[string]any)
			apiObject.AdvancedConfiguration = &awstypes.AdvancedConfiguration{
				AlternateTargetGroupArn: aws.String(config["alternate_target_group_arn"].(string)),
				ProductionListenerRule:  aws.String(config["production_listener_rule"].(string)),
				RoleArn:                 aws.String(config[names.AttrRoleARN].(string)),
			}
			if v, ok := config["test_listener_rule"].(string); ok && v != "" {
				apiObject.AdvancedConfiguration.TestListenerRule = aws.String(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenServiceLoadBalancers(apiObjects []awstypes.LoadBalancer) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"container_name": aws.ToString(apiObject.ContainerName),
			"container_port": aws.ToInt32(apiObject.ContainerPort),
		}

		if apiObject.LoadBalancerName != nil {
			tfMap["elb_name"] = aws.ToString(apiObject.LoadBalancerName)
		}

		if apiObject.TargetGroupArn != nil {
			tfMap["target_group_arn"] = aws.ToString(apiObject.TargetGroupArn)
		}

		if apiObject.AdvancedConfiguration != nil {
			tfMap["advanced_configuration"] = []any{
				map[string]any{
					"alternate_target_group_arn": aws.ToString(apiObject.AdvancedConfiguration.AlternateTargetGroupArn),
					"production_listener_rule":   aws.ToString(apiObject.AdvancedConfiguration.ProductionListenerRule),
					names.AttrRoleARN:            aws.ToString(apiObject.AdvancedConfiguration.RoleArn),
				},
			}
			if apiObject.AdvancedConfiguration.TestListenerRule != nil {
				tfMap["advanced_configuration"].([]any)[0].(map[string]any)["test_listener_rule"] = aws.ToString(apiObject.AdvancedConfiguration.TestListenerRule)
			}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceVolumeConfigurations(ctx context.Context, tfList []any) []awstypes.ServiceVolumeConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.ServiceVolumeConfiguration, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.ServiceVolumeConfiguration{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}

		if v, ok := tfMap["managed_ebs_volume"].([]any); ok && len(v) > 0 {
			apiObject.ManagedEBSVolume = expandServiceManagedEBSVolumeConfiguration(ctx, v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandServiceManagedEBSVolumeConfiguration(ctx context.Context, tfList []any) *awstypes.ServiceManagedEBSVolumeConfiguration {
	if len(tfList) == 0 {
		return &awstypes.ServiceManagedEBSVolumeConfiguration{}
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ServiceManagedEBSVolumeConfiguration{}

	if v, ok := tfMap[names.AttrEncrypted].(bool); ok {
		apiObject.Encrypted = aws.Bool(v)
	}
	if v, ok := tfMap["file_system_type"].(string); ok && v != "" {
		apiObject.FilesystemType = awstypes.TaskFilesystemType(v)
	}
	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Iops = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}
	if v, ok := tfMap["size_in_gb"].(int); ok && v != 0 {
		apiObject.SizeInGiB = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}
	if v, ok := tfMap["tag_specifications"].([]any); ok && len(v) > 0 {
		apiObject.TagSpecifications = expandEBSTagSpecifications(ctx, v)
	}
	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int32(int32(v))
	}
	if v, ok := tfMap["volume_initialization_rate"].(int); ok && v != 0 {
		apiObject.VolumeInitializationRate = aws.Int32(int32(v))
	}
	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandEBSTagSpecifications(ctx context.Context, tfList []any) []awstypes.EBSTagSpecification {
	if len(tfList) == 0 {
		return []awstypes.EBSTagSpecification{}
	}

	var apiObjects []awstypes.EBSTagSpecification

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		var apiObject awstypes.EBSTagSpecification

		if v, ok := tfMap[names.AttrPropagateTags].(string); ok && v != "" {
			apiObject.PropagateTags = awstypes.PropagateTags(v)
		}
		if v, ok := tfMap[names.AttrResourceType].(string); ok && v != "" {
			apiObject.ResourceType = awstypes.EBSResourceType(v)
		}
		if v, ok := tfMap[names.AttrTags].(map[string]any); ok && len(v) > 0 {
			if v := tftags.New(ctx, v).IgnoreAWS(); len(v) > 0 {
				apiObject.Tags = svcTags(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenServiceVolumeConfigurations(ctx context.Context, apiObjects []awstypes.ServiceVolumeConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrName: aws.ToString(apiObject.Name),
		}

		if v := apiObject.ManagedEBSVolume; v != nil {
			tfMap["managed_ebs_volume"] = []any{flattenServiceManagedEBSVolumeConfiguration(ctx, v)}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenServiceManagedEBSVolumeConfiguration(ctx context.Context, apiObject *awstypes.ServiceManagedEBSVolumeConfiguration) map[string]any {
	tfMap := map[string]any{
		names.AttrRoleARN: aws.ToString(apiObject.RoleArn),
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap[names.AttrEncrypted] = aws.ToBool(v)
	}
	if v := apiObject.FilesystemType; v != "" {
		tfMap["file_system_type"] = v
	}
	if v := apiObject.Iops; v != nil {
		tfMap[names.AttrIOPS] = aws.ToInt32(v)
	}
	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}
	if v := apiObject.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}
	if v := apiObject.SizeInGiB; v != nil {
		tfMap["size_in_gb"] = aws.ToInt32(v)
	}
	if v := apiObject.SnapshotId; v != nil {
		tfMap[names.AttrSnapshotID] = aws.ToString(v)
	}
	if v := apiObject.TagSpecifications; v != nil {
		tfMap["tag_specifications"] = flattenEBSTagSpecifications(ctx, apiObject.TagSpecifications)
	}
	if v := apiObject.Throughput; v != nil {
		tfMap[names.AttrThroughput] = aws.ToInt32(v)
	}
	if v := apiObject.VolumeInitializationRate; v != nil {
		tfMap["volume_initialization_rate"] = aws.ToInt32(v)
	}
	if v := apiObject.VolumeType; v != nil {
		tfMap[names.AttrVolumeType] = aws.ToString(v)
	}

	return tfMap
}

func flattenEBSTagSpecifications(ctx context.Context, apiObjects []awstypes.EBSTagSpecification) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.PropagateTags; v != "" {
			tfMap[names.AttrPropagateTags] = v
		}
		if v := apiObject.ResourceType; v != "" {
			tfMap[names.AttrResourceType] = v
		}
		if v := apiObject.Tags; v != nil {
			tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceConnectServices(tfList []any) []awstypes.ServiceConnectService {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ServiceConnectService

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		var apiObject awstypes.ServiceConnectService

		if v, ok := tfMap["client_alias"].([]any); ok && len(v) > 0 {
			apiObject.ClientAliases = expandClientAliases(v)
		}
		if v, ok := tfMap["discovery_name"].(string); ok && v != "" {
			apiObject.DiscoveryName = aws.String(v)
		}
		if v, ok := tfMap["ingress_port_override"].(int); ok && v != 0 {
			apiObject.IngressPortOverride = aws.Int32(int32(v))
		}
		if v, ok := tfMap["port_name"].(string); ok && v != "" {
			apiObject.PortName = aws.String(v)
		}
		if v, ok := tfMap[names.AttrTimeout].([]any); ok && len(v) > 0 {
			apiObject.Timeout = expandTimeout(v)
		}
		if v, ok := tfMap["tls"].([]any); ok && len(v) > 0 {
			apiObject.Tls = expandTLS(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenServiceConnectServices(apiObjects []awstypes.ServiceConnectService) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.ClientAliases; v != nil {
			tfMap["client_alias"] = flattenServiceConnectClientAliases(v)
		}
		if v := apiObject.DiscoveryName; v != nil {
			tfMap["discovery_name"] = aws.ToString(v)
		}
		if v := apiObject.IngressPortOverride; v != nil {
			tfMap["ingress_port_override"] = aws.ToInt32(v)
		}
		if v := apiObject.PortName; v != nil {
			tfMap["port_name"] = aws.ToString(v)
		}
		if v := apiObject.Timeout; v != nil {
			tfMap[names.AttrTimeout] = []any{flattenTimeoutConfiguration(v)}
		}
		if v := apiObject.Tls; v != nil {
			tfMap["tls"] = []any{flattenServiceConnectTLSConfiguration(v)}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceConnectTestTrafficRules(tfList []any) *awstypes.ServiceConnectTestTrafficRules {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ServiceConnectTestTrafficRules{}

	if v, ok := tfMap[names.AttrHeader].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Header = expandServiceConnectHeader(v)
	}

	return apiObject
}

func flattenServiceConnectTestTrafficRules(apiObject *awstypes.ServiceConnectTestTrafficRules) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Header != nil {
		tfMap[names.AttrHeader] = flattenServiceConnectHeader(apiObject.Header)
	}

	return []any{tfMap}
}

func expandTimeout(timeout []any) *awstypes.TimeoutConfiguration {
	if len(timeout) == 0 {
		return nil
	}

	raw, ok := timeout[0].(map[string]any)
	if !ok {
		return nil
	}
	timeoutConfig := &awstypes.TimeoutConfiguration{}
	if v, ok := raw["idle_timeout_seconds"].(int); ok {
		timeoutConfig.IdleTimeoutSeconds = aws.Int32(int32(v))
	}
	if v, ok := raw["per_request_timeout_seconds"].(int); ok {
		timeoutConfig.PerRequestTimeoutSeconds = aws.Int32(int32(v))
	}
	return timeoutConfig
}

func flattenTimeoutConfiguration(apiObject *awstypes.TimeoutConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.IdleTimeoutSeconds; v != nil {
		tfMap["idle_timeout_seconds"] = aws.ToInt32(v)
	}
	if v := apiObject.PerRequestTimeoutSeconds; v != nil {
		tfMap["per_request_timeout_seconds"] = aws.ToInt32(v)
	}

	return tfMap
}

func expandTLS(tls []any) *awstypes.ServiceConnectTlsConfiguration {
	if len(tls) == 0 {
		return nil
	}

	raw, ok := tls[0].(map[string]any)
	if !ok {
		return nil
	}
	tlsConfig := &awstypes.ServiceConnectTlsConfiguration{}
	if v, ok := raw["issuer_cert_authority"].([]any); ok && len(v) > 0 {
		tlsConfig.IssuerCertificateAuthority = expandIssuerCertAuthority(v)
	}
	if v, ok := raw[names.AttrKMSKey].(string); ok && v != "" {
		tlsConfig.KmsKey = aws.String(v)
	}
	if v, ok := raw[names.AttrRoleARN].(string); ok && v != "" {
		tlsConfig.RoleArn = aws.String(v)
	}
	return tlsConfig
}

func flattenServiceConnectTLSConfiguration(c *awstypes.ServiceConnectTlsConfiguration) map[string]any {
	tfMap := map[string]any{}

	if v := c.IssuerCertificateAuthority; v != nil {
		tfMap["issuer_cert_authority"] = []any{flattenServiceConnectTLSCertificateAuthority(v)}
	}
	if v := c.KmsKey; v != nil {
		tfMap[names.AttrKMSKey] = aws.ToString(v)
	}
	if v := c.RoleArn; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	return tfMap
}

func expandIssuerCertAuthority(pca []any) *awstypes.ServiceConnectTlsCertificateAuthority {
	if len(pca) == 0 {
		return nil
	}

	raw, ok := pca[0].(map[string]any)
	if !ok {
		return nil
	}
	config := &awstypes.ServiceConnectTlsCertificateAuthority{}

	if v, ok := raw["aws_pca_authority_arn"].(string); ok && v != "" {
		config.AwsPcaAuthorityArn = aws.String(v)
	}
	return config
}

func flattenServiceConnectTLSCertificateAuthority(apiObject *awstypes.ServiceConnectTlsCertificateAuthority) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AwsPcaAuthorityArn; v != nil {
		tfMap["aws_pca_authority_arn"] = aws.ToString(v)
	}

	return tfMap
}

func expandClientAliases(srv []any) []awstypes.ServiceConnectClientAlias {
	if len(srv) == 0 {
		return nil
	}

	var out []awstypes.ServiceConnectClientAlias
	for _, item := range srv {
		raw, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var config awstypes.ServiceConnectClientAlias
		if v, ok := raw[names.AttrPort].(int); ok {
			config.Port = aws.Int32(int32(v))
		}
		if v, ok := raw[names.AttrDNSName].(string); ok && v != "" {
			config.DnsName = aws.String(v)
		}
		if v, ok := raw["test_traffic_rules"].([]any); ok && len(v) > 0 && v[0] != nil {
			config.TestTrafficRules = expandServiceConnectTestTrafficRules(v)
		}

		out = append(out, config)
	}

	return out
}

func flattenServiceConnectClientAliases(apiObjects []awstypes.ServiceConnectClientAlias) []any {
	tfList := []any{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if v := apiObject.DnsName; v != nil {
			tfMap[names.AttrDNSName] = aws.ToString(v)
		}
		if v := apiObject.Port; v != nil {
			tfMap[names.AttrPort] = aws.ToInt32(v)
		}
		if v := apiObject.TestTrafficRules; v != nil {
			tfMap["test_traffic_rules"] = flattenServiceConnectTestTrafficRules(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandServiceConnectHeader(tfList []any) *awstypes.ServiceConnectTestTrafficHeaderRules {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ServiceConnectTestTrafficHeaderRules{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandServiceConnectHeaderValue(v)
	}

	return apiObject
}

func flattenServiceConnectHeader(apiObject *awstypes.ServiceConnectTestTrafficHeaderRules) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}
	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = flattenServiceConnectHeaderValue(v)
	}

	return []any{tfMap}
}

func expandServiceConnectHeaderValue(tfList []any) *awstypes.ServiceConnectTestTrafficHeaderMatchRules {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ServiceConnectTestTrafficHeaderMatchRules{}

	if v, ok := tfMap["exact"].(string); ok && v != "" {
		apiObject.Exact = aws.String(v)
	}

	return apiObject
}

func flattenServiceConnectHeaderValue(apiObject *awstypes.ServiceConnectTestTrafficHeaderMatchRules) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Exact != nil {
		tfMap["exact"] = aws.ToString(apiObject.Exact)
	}

	return []any{tfMap}
}

func flattenServiceRegistries(apiObjects []awstypes.ServiceRegistry) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"registry_arn": aws.ToString(apiObject.RegistryArn),
		}

		if apiObject.ContainerName != nil {
			tfMap["container_name"] = aws.ToString(apiObject.ContainerName)
		}
		if apiObject.ContainerPort != nil {
			tfMap["container_port"] = aws.ToInt32(apiObject.ContainerPort)
		}
		if apiObject.Port != nil {
			tfMap[names.AttrPort] = aws.ToInt32(apiObject.Port)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func familyAndRevisionFromTaskDefinitionARN(arn string) string {
	return strings.Split(arn, "/")[1]
}

// roleNameFromARN parses a role name from a fully qualified ARN
//
// When providing a role name with a path, it must be prefixed with the full path
// including a leading `/`.
// See: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_CreateService.html#ECS-CreateService-request-role
//
// Expects an IAM role ARN:
//
//	arn:aws:iam::0123456789:role/EcsService
//	arn:aws:iam::0123456789:role/group/my-role
func roleNameFromARN(arn string) string {
	if parts := strings.Split(arn, "/"); len(parts) == 2 {
		return parts[1]
	} else if len(parts) > 2 {
		return fmt.Sprintf("/%s", strings.Join(parts[1:], "/"))
	}
	return ""
}

// clusterNameFromARN parses a cluster name from a fully qualified ARN
//
// Expects an ECS cluster ARN:
//
//	arn:aws:ecs:us-west-2:0123456789:cluster/my-cluster
func clusterNameFromARN(arn string) string {
	if parts := strings.Split(arn, "/"); len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func serviceNameFromARN(s string) string {
	a, err := arn.Parse(s)
	if err != nil {
		return ""
	}

	resParts := strings.Split(a.Resource, "/")
	switch len(resParts) {
	case 3:
		// long arn format arn:aws:ecs:us-west-2:123456789:service/cluster_name/service_name
		return resParts[2]
	case 2:
		// short arn format arn:aws:ecs:us-west-2:123456789:service/service_name
		return resParts[1]
	default:
		return ""
	}
}

func parseServiceImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("wrong format of resource: %s, expecting 'cluster-name/service-name'", id)
	}

	return parts[0], parts[1], nil
}

type serviceImportID struct{}

func (serviceImportID) Create(d *schema.ResourceData) string {
	cluster := d.Get("cluster").(string)
	if arn.IsARN(cluster) {
		cluster = clusterNameFromARN(cluster)
	}

	return fmt.Sprintf("%s/%s", cluster, d.Get(names.AttrName).(string))
}

func (serviceImportID) Parse(id string) (string, map[string]any, error) {
	cluster, service, err := parseServiceImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"cluster":      cluster,
		names.AttrName: service,
	}
	return id, result, nil
}
