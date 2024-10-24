// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_ecs_service", name="Service")
// @Tags(identifierAttribute="id")
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
							ForceNew: true,
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
			"deployment_controller": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							ForceNew:         true,
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
							StateFunc: func(v interface{}) string {
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
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
											},
										},
									},
									"discovery_name": {
										Type:     schema.TypeString,
										Optional: true,
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
									names.AttrThroughput: {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 1000),
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag_specifications": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrResourceType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.EBSResourceType](),
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
												names.AttrTags: tftags.TagsSchema(),
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

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type: resourceV0.CoreConfigSchema().ImpliedType(),
				Upgrade: func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
					// Convert volume_configuration.managed_ebs_volume.throughput from string to int.
					if v, ok := rawState["volume_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
						tfMap := v[0].(map[string]interface{})
						if v, ok := tfMap["managed_ebs_volume"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
							tfMap := v[0].(map[string]interface{})
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
			verify.SetTagsDiff,
			capacityProviderStrategyCustomizeDiff,
			triggersCustomizeDiff,
		),
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	deploymentController := expandDeploymentController(d.Get("deployment_controller").([]interface{}))
	deploymentMinimumHealthyPercent := d.Get("deployment_minimum_healthy_percent").(int)
	name := d.Get(names.AttrName).(string)
	schedulingStrategy := awstypes.SchedulingStrategy(d.Get("scheduling_strategy").(string))
	input := &ecs.CreateServiceInput{
		CapacityProviderStrategy: expandCapacityProviderStrategyItems(d.Get(names.AttrCapacityProviderStrategy).(*schema.Set)),
		ClientToken:              aws.String(id.UniqueId()),
		DeploymentConfiguration:  &awstypes.DeploymentConfiguration{},
		DeploymentController:     deploymentController,
		EnableECSManagedTags:     d.Get("enable_ecs_managed_tags").(bool),
		EnableExecuteCommand:     d.Get("enable_execute_command").(bool),
		NetworkConfiguration:     expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]interface{})),
		SchedulingStrategy:       schedulingStrategy,
		ServiceName:              aws.String(name),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alarms"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.Alarms = expandAlarms(v.([]interface{})[0].(map[string]interface{}))
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

	if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
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

	if v := expandLoadBalancers(d.Get("load_balancer").(*schema.Set).List()); len(v) > 0 {
		input.LoadBalancers = v
	}

	if v, ok := d.GetOk("ordered_placement_strategy"); ok {
		apiObject, err := expandPlacementStrategy(v.([]interface{}))
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

	if v, ok := d.GetOk("service_connect_configuration"); ok && len(v.([]interface{})) > 0 {
		input.ServiceConnectConfiguration = expandServiceConnectConfiguration(v.([]interface{}))
	}

	if v := d.Get("service_registries").([]interface{}); len(v) > 0 {
		input.ServiceRegistries = expandServiceRegistries(v)
	}

	if v, ok := d.GetOk("task_definition"); ok {
		input.TaskDefinition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume_configuration"); ok && len(v.([]interface{})) > 0 {
		input.VolumeConfigurations = expandVolumeConfigurations(ctx, v.([]interface{}))
	}

	output, err := retryServiceCreate(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryServiceCreate(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Service.ServiceArn))

	fn := waitServiceActive
	if d.Get("wait_for_steady_state").(bool) {
		fn = waitServiceStable
	}
	if _, err := fn(ctx, conn, d.Id(), d.Get("cluster").(string), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceServiceRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Service (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster := d.Get("cluster").(string)
	service, err := findServiceByTwoPartKeyWaitForActive(ctx, conn, d.Id(), cluster)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Service (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(service.ServiceArn))
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
			if err := d.Set("alarms", []interface{}{flattenAlarms(service.DeploymentConfiguration.Alarms)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting alarms: %s", err)
			}
		} else {
			d.Set("alarms", nil)
		}

		if service.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			if err := d.Set("deployment_circuit_breaker", []interface{}{flattenDeploymentCircuitBreaker(service.DeploymentConfiguration.DeploymentCircuitBreaker)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting deployment_circuit_breaker: %s", err)
			}
		} else {
			d.Set("deployment_circuit_breaker", nil)
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
		if err := d.Set("load_balancer", flattenLoadBalancers(service.LoadBalancers)); err != nil {
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
	// if err := d.Set("service_connect_configuration", flattenServiceConnectConfiguration(service.ServiceConnectConfiguration)); err != nil {
	// 	return sdkdiag.AppendErrorf(diags, "setting service_connect_configuration: %s", err)
	// }
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

	setTagsOut(ctx, service.Tags)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChangesExcept(names.AttrForceDelete, names.AttrTags, names.AttrTagsAll) {
		cluster := d.Get("cluster").(string)
		input := &ecs.UpdateServiceInput{
			Cluster:            aws.String(cluster),
			ForceNewDeployment: d.Get("force_new_deployment").(bool),
			Service:            aws.String(d.Id()),
		}

		if d.HasChange("alarms") {
			if input.DeploymentConfiguration == nil {
				input.DeploymentConfiguration = &awstypes.DeploymentConfiguration{}
			}

			if v, ok := d.GetOk("alarms"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.DeploymentConfiguration.Alarms = expandAlarms(v.([]interface{})[0].(map[string]interface{}))
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

			if v, ok := d.GetOk("deployment_circuit_breaker"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.DeploymentConfiguration.DeploymentCircuitBreaker = expandDeploymentCircuitBreaker(v.([]interface{})[0].(map[string]interface{}))
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
				input.LoadBalancers = expandLoadBalancers(v.List())
			}
		}

		if d.HasChange(names.AttrNetworkConfiguration) {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]interface{}))
		}

		if d.HasChange("ordered_placement_strategy") {
			// Reference: https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_UpdateService.html#ECS-UpdateService-request-placementStrategy
			// To remove an existing placement strategy, specify an empty object.
			input.PlacementStrategy = []awstypes.PlacementStrategy{}

			if v, ok := d.GetOk("ordered_placement_strategy"); ok && len(v.([]interface{})) > 0 {
				apiObject, err := expandPlacementStrategy(v.([]interface{}))
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
			input.ServiceConnectConfiguration = expandServiceConnectConfiguration(d.Get("service_connect_configuration").([]interface{}))
		}

		if d.HasChange("service_registries") {
			input.ServiceRegistries = expandServiceRegistries(d.Get("service_registries").([]interface{}))
		}

		if d.HasChange("task_definition") {
			input.TaskDefinition = aws.String(d.Get("task_definition").(string))
		}

		if d.HasChange("volume_configuration") {
			input.VolumeConfigurations = expandVolumeConfigurations(ctx, d.Get("volume_configuration").([]interface{}))
		}

		// Retry due to IAM eventual consistency.
		const (
			serviceUpdateTimeout = 2 * time.Minute
			timeout              = propagationTimeout + serviceUpdateTimeout
		)
		_, err := tfresource.RetryWhen(ctx, timeout,
			func() (interface{}, error) {
				return conn.UpdateService(ctx, input)
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

		fn := waitServiceActive
		if d.Get("wait_for_steady_state").(bool) {
			fn = waitServiceStable
		}
		if _, err := fn(ctx, conn, d.Id(), cluster, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Service (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	cluster := d.Get("cluster").(string)
	service, err := findServiceNoTagsByTwoPartKey(ctx, conn, d.Id(), cluster)

	if tfresource.NotFound(err) {
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
		func() (interface{}, error) {
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

func resourceServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of resource: %s, expecting 'cluster-name/service-name'", d.Id())
	}
	cluster := strings.Split(d.Id(), "/")[0]
	name := strings.Split(d.Id(), "/")[1]
	log.Printf("[DEBUG] Importing ECS service %s from cluster %s", name, cluster)

	d.SetId(name)
	clusterArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "ecs",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("cluster/%s", cluster),
	}.String()
	d.Set("cluster", clusterArn)
	return []*schema.ResourceData{d}, nil
}

func retryServiceCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error) {
	const (
		serviceCreateTimeout = 2 * time.Minute
		timeout              = propagationTimeout + serviceCreateTimeout
	)
	outputRaw, err := tfresource.RetryWhen(ctx, timeout,
		func() (interface{}, error) {
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
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	// When an ECS Service is not found by DescribeServices(), it will return a Failure struct with Reason = "MISSING"
	for _, v := range output.Failures {
		if aws.ToString(v.Reason) == failureReasonMissing {
			return nil, &retry.NotFoundError{
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

	// Use the retry.RetryContext function instead of WaitForState() because we don't want the timeout error, if any.
	const (
		timeout = 2 * time.Minute
	)
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error

		service, err = findServiceByTwoPartKey(ctx, conn, serviceName, clusterNameOrARN)

		if tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if status := aws.ToString(service.Status); status != serviceStatusActive {
			return retry.RetryableError(newExpectServiceActiveError(status))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		service, err = findServiceByTwoPartKey(ctx, conn, serviceName, clusterNameOrARN)
	}

	if errs.IsA[*expectServiceActiveError](err) {
		return nil, &retry.NotFoundError{
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

func statusService(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServiceNoTagsByTwoPartKey(ctx, conn, serviceName, clusterNameOrARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), err
	}
}

func statusServiceWaitForStable(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		outputRaw, status, err := statusService(ctx, conn, serviceName, clusterNameOrARN)()

		if err != nil {
			return nil, "", err
		}

		if status != serviceStatusActive {
			return outputRaw, status, nil
		}

		output := outputRaw.(*awstypes.Service)

		if n, dc, rc := len(output.Deployments), output.DesiredCount, output.RunningCount; n == 1 && dc == rc {
			status = serviceStatusStable
		} else {
			status = serviceStatusPending
		}

		return output, status, nil
	}
}

// waitServiceStable waits for an ECS Service to reach the status "ACTIVE" and have all desired tasks running.
// Does not return tags.
func waitServiceStable(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, timeout time.Duration) (*awstypes.Service, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining, serviceStatusPending},
		Target:  []string{serviceStatusStable},
		Refresh: statusServiceWaitForStable(ctx, conn, serviceName, clusterNameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Service); ok {
		return output, err
	}

	return nil, err
}

// Does not return tags.
func waitServiceActive(ctx context.Context, conn *ecs.Client, serviceName, clusterNameOrARN string, timeout time.Duration) (*awstypes.Service, error) {
	stateConf := &retry.StateChangeConf{
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
	stateConf := &retry.StateChangeConf{
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

func triggersCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
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
		if len(o.(map[string]interface{})) > 0 && len(n.(map[string]interface{})) == 0 {
			return d.Clear(names.AttrTriggers)
		}

		return nil
	}

	return nil
}

func capacityProviderStrategyCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// to be backward compatible, should ForceNew almost always (previous behavior), unless force_new_deployment is true
	if v := d.Get("force_new_deployment").(bool); !v {
		return capacityProviderStrategyForceNew(d)
	}

	return nil
}

func capacityProviderStrategyForceNew(d *schema.ResourceDiff) error {
	for _, key := range d.GetChangedKeysPrefix(names.AttrCapacityProviderStrategy) {
		if d.HasChange(key) {
			if err := d.ForceNew(key); err != nil {
				return fmt.Errorf("while attempting to force a new ECS service for capacity_provider_strategy: %w", err)
			}
		}
	}
	return nil
}

func expandAlarms(tfMap map[string]interface{}) *awstypes.DeploymentAlarms {
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

func flattenAlarms(apiObject *awstypes.DeploymentAlarms) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AlarmNames; v != nil {
		tfMap["alarm_names"] = v
	}

	tfMap["enable"] = apiObject.Enable

	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func expandDeploymentController(l []interface{}) *awstypes.DeploymentController {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	deploymentController := &awstypes.DeploymentController{
		Type: awstypes.DeploymentControllerType(m[names.AttrType].(string)),
	}

	return deploymentController
}

func flattenDeploymentController(deploymentController *awstypes.DeploymentController) []interface{} {
	m := map[string]interface{}{
		names.AttrType: awstypes.DeploymentControllerTypeEcs,
	}

	if deploymentController == nil {
		return []interface{}{m}
	}

	m[names.AttrType] = string(deploymentController.Type)

	return []interface{}{m}
}

func expandDeploymentCircuitBreaker(tfMap map[string]interface{}) *awstypes.DeploymentCircuitBreaker {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DeploymentCircuitBreaker{}

	apiObject.Enable = tfMap["enable"].(bool)
	apiObject.Rollback = tfMap["rollback"].(bool)

	return apiObject
}

func flattenDeploymentCircuitBreaker(apiObject *awstypes.DeploymentCircuitBreaker) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["enable"] = apiObject.Enable
	tfMap["rollback"] = apiObject.Rollback

	return tfMap
}

func flattenNetworkConfiguration(nc *awstypes.NetworkConfiguration) []interface{} {
	if nc == nil {
		return nil
	}

	result := make(map[string]interface{})
	result[names.AttrSecurityGroups] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.SecurityGroups)
	result[names.AttrSubnets] = flex.FlattenStringValueSet(nc.AwsvpcConfiguration.Subnets)

	result["assign_public_ip"] = nc.AwsvpcConfiguration.AssignPublicIp == awstypes.AssignPublicIpEnabled

	return []interface{}{result}
}

func expandNetworkConfiguration(nc []interface{}) *awstypes.NetworkConfiguration {
	if len(nc) == 0 {
		return nil
	}
	awsVpcConfig := &awstypes.AwsVpcConfiguration{}
	raw := nc[0].(map[string]interface{})
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

func expandPlacementConstraints(tfList []interface{}) ([]awstypes.PlacementConstraint, error) {
	if len(tfList) == 0 {
		return nil, nil
	}

	var result []awstypes.PlacementConstraint

	for _, tfMapRaw := range tfList {
		if tfMapRaw == nil {
			continue
		}

		tfMap := tfMapRaw.(map[string]interface{})

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

func flattenServicePlacementConstraints(pcs []awstypes.PlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c[names.AttrType] = string(pc.Type)
		if pc.Expression != nil {
			c[names.AttrExpression] = aws.ToString(pc.Expression)
		}

		results = append(results, c)
	}
	return results
}

func expandPlacementStrategy(s []interface{}) ([]awstypes.PlacementStrategy, error) {
	if len(s) == 0 {
		return nil, nil
	}
	pss := make([]awstypes.PlacementStrategy, 0)
	for _, raw := range s {
		p, ok := raw.(map[string]interface{})

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

func flattenPlacementStrategy(pss []awstypes.PlacementStrategy) []interface{} {
	if len(pss) == 0 {
		return nil
	}
	results := make([]interface{}, 0, len(pss))
	for _, ps := range pss {
		c := make(map[string]interface{})
		c[names.AttrType] = string(ps.Type)

		if ps.Field != nil {
			c[names.AttrField] = aws.ToString(ps.Field)

			// for some fields the API requires lowercase for creation but will return uppercase on query
			if aws.ToString(ps.Field) == "MEMORY" || aws.ToString(ps.Field) == "CPU" {
				c[names.AttrField] = strings.ToLower(aws.ToString(ps.Field))
			}
		}

		results = append(results, c)
	}
	return results
}

func expandServiceConnectConfiguration(sc []interface{}) *awstypes.ServiceConnectConfiguration {
	if len(sc) == 0 {
		return &awstypes.ServiceConnectConfiguration{
			Enabled: false,
		}
	}
	raw := sc[0].(map[string]interface{})

	config := &awstypes.ServiceConnectConfiguration{}
	if v, ok := raw[names.AttrEnabled].(bool); ok {
		config.Enabled = v
	}

	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandLogConfiguration(v)
	}

	if v, ok := raw[names.AttrNamespace].(string); ok && v != "" {
		config.Namespace = aws.String(v)
	}

	if v, ok := raw["service"].([]interface{}); ok && len(v) > 0 {
		config.Services = expandServices(v)
	}

	return config
}

func expandLogConfiguration(lc []interface{}) *awstypes.LogConfiguration {
	if len(lc) == 0 {
		return &awstypes.LogConfiguration{}
	}
	raw := lc[0].(map[string]interface{})

	config := &awstypes.LogConfiguration{}
	if v, ok := raw["log_driver"].(string); ok && v != "" {
		config.LogDriver = awstypes.LogDriver(v)
	}
	if v, ok := raw["options"].(map[string]interface{}); ok && len(v) > 0 {
		config.Options = flex.ExpandStringValueMap(v)
	}
	if v, ok := raw["secret_option"].([]interface{}); ok && len(v) > 0 {
		config.SecretOptions = expandSecretOptions(v)
	}

	return config
}

func expandSecretOptions(sop []interface{}) []awstypes.Secret {
	if len(sop) == 0 {
		return nil
	}

	var out []awstypes.Secret
	for _, item := range sop {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config awstypes.Secret
		if v, ok := raw[names.AttrName].(string); ok && v != "" {
			config.Name = aws.String(v)
		}
		if v, ok := raw["value_from"].(string); ok && v != "" {
			config.ValueFrom = aws.String(v)
		}

		out = append(out, config)
	}

	return out
}

func expandVolumeConfigurations(ctx context.Context, vc []interface{}) []awstypes.ServiceVolumeConfiguration {
	if len(vc) == 0 {
		return nil
	}

	vcs := make([]awstypes.ServiceVolumeConfiguration, 0)

	for _, raw := range vc {
		p := raw.(map[string]interface{})

		config := awstypes.ServiceVolumeConfiguration{
			Name: aws.String(p[names.AttrName].(string)),
		}

		if v, ok := p["managed_ebs_volume"].([]interface{}); ok && len(v) > 0 {
			config.ManagedEBSVolume = expandManagedEBSVolume(ctx, v)
		}
		vcs = append(vcs, config)
	}

	return vcs
}

func expandManagedEBSVolume(ctx context.Context, ebs []interface{}) *awstypes.ServiceManagedEBSVolumeConfiguration {
	if len(ebs) == 0 {
		return &awstypes.ServiceManagedEBSVolumeConfiguration{}
	}
	raw := ebs[0].(map[string]interface{})

	config := &awstypes.ServiceManagedEBSVolumeConfiguration{}
	if v, ok := raw[names.AttrRoleARN].(string); ok && v != "" {
		config.RoleArn = aws.String(v)
	}
	if v, ok := raw[names.AttrEncrypted].(bool); ok {
		config.Encrypted = aws.Bool(v)
	}
	if v, ok := raw["file_system_type"].(string); ok && v != "" {
		config.FilesystemType = awstypes.TaskFilesystemType(v)
	}
	if v, ok := raw[names.AttrIOPS].(int); ok && v != 0 {
		config.Iops = aws.Int32(int32(v))
	}
	if v, ok := raw[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}
	if v, ok := raw["size_in_gb"].(int); ok && v != 0 {
		config.SizeInGiB = aws.Int32(int32(v))
	}
	if v, ok := raw[names.AttrSnapshotID].(string); ok && v != "" {
		config.SnapshotId = aws.String(v)
	}
	if v, ok := raw[names.AttrThroughput].(int); ok && v != 0 {
		config.Throughput = aws.Int32(int32(v))
	}
	if v, ok := raw[names.AttrVolumeType].(string); ok && v != "" {
		config.VolumeType = aws.String(v)
	}
	if v, ok := raw["tag_specifications"].([]interface{}); ok && len(v) > 0 {
		config.TagSpecifications = expandTagSpecifications(ctx, v)
	}

	return config
}

func expandTagSpecifications(ctx context.Context, ts []interface{}) []awstypes.EBSTagSpecification {
	if len(ts) == 0 {
		return []awstypes.EBSTagSpecification{}
	}

	var s []awstypes.EBSTagSpecification
	for _, item := range ts {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config awstypes.EBSTagSpecification
		if v, ok := raw[names.AttrResourceType].(string); ok && v != "" {
			config.ResourceType = awstypes.EBSResourceType(v)
		}
		if v, ok := raw[names.AttrPropagateTags].(string); ok && v != "" {
			config.PropagateTags = awstypes.PropagateTags(v)
		}
		if v, ok := raw[names.AttrTags].(map[string]any); ok && len(v) > 0 {
			if v := tftags.New(ctx, v).IgnoreAWS(); len(v) > 0 {
				config.Tags = Tags(v)
			}
		}
		s = append(s, config)
	}

	return s
}

func expandServices(srv []interface{}) []awstypes.ServiceConnectService {
	if len(srv) == 0 {
		return nil
	}

	var out []awstypes.ServiceConnectService
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var config awstypes.ServiceConnectService
		if v, ok := raw["client_alias"].([]interface{}); ok && len(v) > 0 {
			config.ClientAliases = expandClientAliases(v)
		}
		if v, ok := raw["discovery_name"].(string); ok && v != "" {
			config.DiscoveryName = aws.String(v)
		}
		if v, ok := raw["ingress_port_override"].(int); ok && v != 0 {
			config.IngressPortOverride = aws.Int32(int32(v))
		}
		if v, ok := raw["port_name"].(string); ok && v != "" {
			config.PortName = aws.String(v)
		}

		if v, ok := raw[names.AttrTimeout].([]interface{}); ok && len(v) > 0 {
			config.Timeout = expandTimeout(v)
		}

		if v, ok := raw["tls"].([]interface{}); ok && len(v) > 0 {
			config.Tls = expandTLS(v)
		}

		out = append(out, config)
	}

	return out
}

func expandTimeout(timeout []interface{}) *awstypes.TimeoutConfiguration {
	if len(timeout) == 0 {
		return nil
	}

	raw, ok := timeout[0].(map[string]interface{})
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

func expandTLS(tls []interface{}) *awstypes.ServiceConnectTlsConfiguration {
	if len(tls) == 0 {
		return nil
	}

	raw, ok := tls[0].(map[string]interface{})
	if !ok {
		return nil
	}
	tlsConfig := &awstypes.ServiceConnectTlsConfiguration{}
	if v, ok := raw["issuer_cert_authority"].([]interface{}); ok && len(v) > 0 {
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

func expandIssuerCertAuthority(pca []interface{}) *awstypes.ServiceConnectTlsCertificateAuthority {
	if len(pca) == 0 {
		return nil
	}

	raw, ok := pca[0].(map[string]interface{})
	if !ok {
		return nil
	}
	config := &awstypes.ServiceConnectTlsCertificateAuthority{}

	if v, ok := raw["aws_pca_authority_arn"].(string); ok && v != "" {
		config.AwsPcaAuthorityArn = aws.String(v)
	}
	return config
}

func expandClientAliases(srv []interface{}) []awstypes.ServiceConnectClientAlias {
	if len(srv) == 0 {
		return nil
	}

	var out []awstypes.ServiceConnectClientAlias
	for _, item := range srv {
		raw, ok := item.(map[string]interface{})
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

		out = append(out, config)
	}

	return out
}

func flattenServiceRegistries(srs []awstypes.ServiceRegistry) []map[string]interface{} {
	if len(srs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, sr := range srs {
		c := map[string]interface{}{
			"registry_arn": aws.ToString(sr.RegistryArn),
		}
		if sr.Port != nil {
			c[names.AttrPort] = int(aws.ToInt32(sr.Port))
		}
		if sr.ContainerPort != nil {
			c["container_port"] = int(aws.ToInt32(sr.ContainerPort))
		}
		if sr.ContainerName != nil {
			c["container_name"] = aws.ToString(sr.ContainerName)
		}
		results = append(results, c)
	}
	return results
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
