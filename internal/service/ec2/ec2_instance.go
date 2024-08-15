// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_instance", name="Instance")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.Instance")
// @Testing(importIgnore="user_data_replace_on_change")
func resourceInstance() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstanceUpdate,
		DeleteWithoutTimeout: resourceInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		MigrateState:  instanceMigrateState,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"ami", names.AttrLaunchTemplate},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Computed: true,
				Optional: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"capacity_reservation_specification": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_reservation_preference": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationPreference](),
							ExactlyOneOf:     []string{"capacity_reservation_specification.0.capacity_reservation_preference", "capacity_reservation_specification.0.capacity_reservation_target"},
						},
						"capacity_reservation_target": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_reservation_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn"},
									},
									"capacity_reservation_resource_group_arn": {
										Type:          schema.TypeString,
										Optional:      true,
										ValidateFunc:  verify.ValidARN,
										ConflictsWith: []string{"capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"},
									},
								},
							},
							ExactlyOneOf: []string{"capacity_reservation_specification.0.capacity_reservation_preference", "capacity_reservation_specification.0.capacity_reservation_target"},
						},
					},
				},
			},
			"cpu_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amd_sev_snp": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AmdSevSnpSpecification](),
							// prevents ForceNew for the case where users launch EC2 instances without cpu_options
							// then in a second apply set cpu_options.0.amd_sev_snp to "disabled"
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if d.Id() != "" && old == "" && new == string(awstypes.AmdSevSnpSpecificationDisabled) {
									return true
								}
								return false
							},
						},
						"core_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"cpu_core_count"},
						},
						"threads_per_core": {
							Type:          schema.TypeInt,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"cpu_threads_per_core"},
						},
					},
				},
			},
			"cpu_core_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Deprecated:    "use 'cpu_options' argument instead",
				ConflictsWith: []string{"cpu_options.0.core_count"},
			},
			"cpu_threads_per_core": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Deprecated:    "use 'cpu_options' argument instead",
				ConflictsWith: []string{"cpu_options.0.threads_per_core"},
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cpuCredits_Values(), false),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Only work with existing instances
								if d.Id() == "" {
									return false
								}
								// Only work with missing configurations
								if new != "" {
									return false
								}
								// Only work when already set in Terraform state
								if old == "" {
									return false
								}
								return true
							},
						},
					},
				},
			},
			"disable_api_stop": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrIOPS: {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: iopsDiffSuppressFunc,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrSnapshotID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrTags:    tagsSchemaConflictsWith([]string{"volume_tags"}),
						names.AttrTagsAll: tftags.TagsSchemaComputed(),
						names.AttrThroughput: {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: throughputDiffSuppressFunc,
						},
						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrVolumeType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrSnapshotID].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"enclave_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"no_device": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrVirtualName].(string)))
					if v, ok := m["no_device"].(bool); ok && v {
						buf.WriteString(fmt.Sprintf("%t-", v))
					}
					return create.StringHashcode(buf.String())
				},
			},
			"get_password_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"hibernation": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"host_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"host_resource_group_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"placement_group"},
			},
			"iam_instance_profile": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_initiated_shutdown_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_lifecycle": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MarketType](),
						},
						"spot_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_interruption_behavior": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InstanceInterruptionBehavior](),
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
										DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
											if (oldValue != "" && newValue == "") || (strings.TrimRight(oldValue, "0") == strings.TrimRight(newValue, "0")) {
												return true
											}
											return false
										},
									},
									"spot_instance_type": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.SpotInstanceType](),
									},
									"valid_until": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
								},
							},
						},
					},
				},
			},
			"instance_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{names.AttrInstanceType, names.AttrLaunchTemplate},
			},
			"ipv6_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_addresses"},
			},
			"ipv6_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count"},
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			names.AttrLaunchTemplate: {
				Type:         schema.TypeList,
				MaxItems:     1,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"ami", names.AttrInstanceType, names.AttrLaunchTemplate},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"launch_template.0.name", "launch_template.0.id"},
							ValidateFunc: verify.ValidLaunchTemplateID,
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"launch_template.0.name", "launch_template.0.id"},
							ValidateFunc: verify.ValidLaunchTemplateName,
						},
						names.AttrVersion: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
							Default:      launchTemplateVersionDefault,
						},
					},
				},
			},
			"maintenance_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_recovery": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InstanceAutoRecoveryState](),
						},
					},
				},
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.InstanceMetadataEndpointStateEnabled,
							ValidateDiagFunc: enum.Validate[awstypes.InstanceMetadataEndpointState](),
						},
						"http_protocol_ipv6": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.InstanceMetadataProtocolStateDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.InstanceMetadataProtocolState](),
						},
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HttpTokensState](),
						},
						"instance_metadata_tags": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.InstanceMetadataTagsState](),
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"network_interface": {
				ConflictsWith: []string{"associate_public_ip_address", names.AttrSubnetID, "private_ip", "secondary_private_ips", names.AttrVPCSecurityGroupIDs, names.AttrSecurityGroups, "ipv6_addresses", "ipv6_address_count", "source_dest_check"},
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
							ForceNew: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"network_card_index": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  0,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_group": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"host_resource_group_arn"},
			},
			"placement_partition_number": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"primary_network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_resource_name_dns_aaaa_record": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_resource_name_dns_a_record": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"hostname_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HostnameType](),
						},
					},
				},
			},
			"private_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_block_device": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					// "For the root volume, you can only modify the following: volume size, volume type, and the Delete on Termination flag."
					// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrIOPS: {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: iopsDiffSuppressFunc,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrTags:    tagsSchemaConflictsWith([]string{"volume_tags"}),
						names.AttrTagsAll: tftags.TagsSchemaComputed(),
						names.AttrThroughput: {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: throughputDiffSuppressFunc,
						},
						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						names.AttrVolumeType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
						},
					},
				},
			},
			"secondary_private_ips": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Suppress diff if network_interface is set
					_, ok := d.GetOk("network_interface")
					return ok
				},
			},
			"spot_instance_request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tenancy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Tenancy](),
			},
			"user_data": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_data_base64"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Sometimes the EC2 API responds with the equivalent, empty SHA1 sum
					// echo -n "" | shasum
					if (old == "da39a3ee5e6b4b0d3255bfef95601890afd80709" && new == "") ||
						(old == "" && new == "da39a3ee5e6b4b0d3255bfef95601890afd80709") {
						return true
					}
					return false
				},
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return userDataHashSum(v)
					default:
						return ""
					}
				},
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
			"user_data_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"user_data"},
				ValidateFunc:  verify.ValidBase64String,
			},
			"user_data_replace_on_change": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"volume_tags": tftags.TagsSchema(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				_, ok := diff.GetOk(names.AttrLaunchTemplate)

				if diff.Id() != "" && diff.HasChange("launch_template.0.version") && ok {
					conn := meta.(*conns.AWSClient).EC2Client(ctx)

					stateVersion := diff.Get("launch_template.0.version")

					var err error
					var launchTemplateID, instanceVersion, defaultVersion, latestVersion string

					launchTemplateID, err = findInstanceLaunchTemplateID(ctx, conn, diff.Id())

					if err != nil {
						return err
					}

					if launchTemplateID != "" {
						instanceVersion, err = findInstanceLaunchTemplateVersion(ctx, conn, diff.Id())

						if err != nil {
							return err
						}

						_, defaultVersion, latestVersion, err = findLaunchTemplateNameAndVersions(ctx, conn, launchTemplateID)

						if err != nil {
							return err
						}
					}

					switch stateVersion {
					case launchTemplateVersionDefault:
						if instanceVersion != defaultVersion {
							diff.ForceNew("launch_template.0.version")
						}
					case launchTemplateVersionLatest:
						if instanceVersion != latestVersion {
							diff.ForceNew("launch_template.0.version")
						}
					default:
						if stateVersion != instanceVersion {
							diff.ForceNew("launch_template.0.version")
						}
					}
				}

				return nil
			},
			customdiff.ComputedIf("launch_template.0.id", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("launch_template.0.name")
			}),
			customdiff.ComputedIf("launch_template.0.name", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("launch_template.0.id")
			}),
			customdiff.ForceNewIf("user_data", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.Get("user_data_replace_on_change").(bool)
			}),
			customdiff.ForceNewIf("user_data_base64", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.Get("user_data_replace_on_change").(bool)
			}),
			customdiff.ForceNewIf(names.AttrInstanceType, func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				conn := meta.(*conns.AWSClient).EC2Client(ctx)

				_, ok := diff.GetOk(names.AttrInstanceType)

				if diff.Id() == "" || !diff.HasChange(names.AttrInstanceType) || !ok {
					return false
				}

				o, n := diff.GetChange(names.AttrInstanceType)
				it1, err := findInstanceTypeByName(ctx, conn, o.(string))
				if err != nil {
					return false
				}

				it2, err := findInstanceTypeByName(ctx, conn, n.(string))
				if err != nil {
					return false
				}

				if it1 == nil || it2 == nil {
					return false
				}

				if it1.InstanceType == it2.InstanceType {
					return false
				}

				if hasCommonElement(it1.ProcessorInfo.SupportedArchitectures, it2.ProcessorInfo.SupportedArchitectures) {
					return false
				}

				return true
			}),
		),
	}
}

func iopsDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// Suppress diff if volume_type is not io1, io2, or gp3 and iops is unset or configured as 0
	i := strings.LastIndexByte(k, '.')
	vt := k[:i+1] + names.AttrVolumeType
	v := d.Get(vt).(string)
	return (strings.ToLower(v) != string(awstypes.VolumeTypeIo1) && strings.ToLower(v) != string(awstypes.VolumeTypeIo2) && strings.ToLower(v) != string(awstypes.VolumeTypeGp3)) && new == "0"
}

func throughputDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	// Suppress diff if volume_type is not gp3 and throughput is unset or configured as 0
	i := strings.LastIndexByte(k, '.')
	vt := k[:i+1] + names.AttrVolumeType
	v := d.Get(vt).(string)
	return strings.ToLower(v) != string(awstypes.VolumeTypeGp3) && new == "0"
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instanceOpts, err := buildInstanceOpts(ctx, d, meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "collecting instance settings: %s", err)
	}

	// instance itself
	tagSpecifications := getTagSpecificationsIn(ctx, awstypes.ResourceTypeInstance)

	// block devices
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tagSpecifications = append(tagSpecifications,
		tagSpecificationsFromKeyValue(
			defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("volume_tags").(map[string]interface{}))),
			string(awstypes.ResourceTypeVolume))...)

	input := &ec2.RunInstancesInput{
		BlockDeviceMappings:               instanceOpts.BlockDeviceMappings,
		CapacityReservationSpecification:  instanceOpts.CapacityReservationSpecification,
		ClientToken:                       aws.String(id.UniqueId()),
		CpuOptions:                        instanceOpts.CpuOptions,
		CreditSpecification:               instanceOpts.CreditSpecification,
		DisableApiTermination:             instanceOpts.DisableAPITermination,
		EbsOptimized:                      instanceOpts.EBSOptimized,
		EnclaveOptions:                    instanceOpts.EnclaveOptions,
		HibernationOptions:                instanceOpts.HibernationOptions,
		IamInstanceProfile:                instanceOpts.IAMInstanceProfile,
		ImageId:                           instanceOpts.ImageID,
		InstanceInitiatedShutdownBehavior: instanceOpts.InstanceInitiatedShutdownBehavior,
		InstanceMarketOptions:             instanceOpts.InstanceMarketOptions,
		InstanceType:                      instanceOpts.InstanceType,
		Ipv6AddressCount:                  instanceOpts.Ipv6AddressCount,
		Ipv6Addresses:                     instanceOpts.Ipv6Addresses,
		KeyName:                           instanceOpts.KeyName,
		LaunchTemplate:                    instanceOpts.LaunchTemplate,
		MaintenanceOptions:                instanceOpts.MaintenanceOptions,
		MaxCount:                          aws.Int32(1),
		MetadataOptions:                   instanceOpts.MetadataOptions,
		MinCount:                          aws.Int32(1),
		Monitoring:                        instanceOpts.Monitoring,
		NetworkInterfaces:                 instanceOpts.NetworkInterfaces,
		Placement:                         instanceOpts.Placement,
		PrivateDnsNameOptions:             instanceOpts.PrivateDNSNameOptions,
		PrivateIpAddress:                  instanceOpts.PrivateIPAddress,
		SecurityGroupIds:                  instanceOpts.SecurityGroupIDs,
		SecurityGroups:                    instanceOpts.SecurityGroups,
		SubnetId:                          instanceOpts.SubnetID,
		TagSpecifications:                 tagSpecifications,
		UserData:                          instanceOpts.UserData64,
	}

	if instanceOpts.DisableAPIStop != nil {
		input.DisableApiStop = instanceOpts.DisableAPIStop
	}

	log.Printf("[DEBUG] Creating EC2 Instance: %s", d.Id())
	outputRaw, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.RunInstances(ctx, input)
		},
		func(err error) (bool, error) {
			// IAM instance profiles can take ~10 seconds to propagate in AWS:
			// http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "Invalid IAM Instance Profile") {
				return true, err
			}

			// IAM roles can also take time to propagate in AWS:
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, " has no associated IAM Roles") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Instance: %s", err)
	}

	instanceId := outputRaw.(*ec2.RunInstancesOutput).Instances[0].InstanceId

	d.SetId(aws.ToString(instanceId))

	instance, err := waitInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) create: %s", d.Id(), err)
	}

	// Initialize the connection info
	if instance.PublicIpAddress != nil {
		d.SetConnInfo(map[string]string{
			names.AttrType: "ssh",
			"host":         aws.ToString(instance.PublicIpAddress),
		})
	} else if instance.PrivateIpAddress != nil {
		d.SetConnInfo(map[string]string{
			names.AttrType: "ssh",
			"host":         aws.ToString(instance.PrivateIpAddress),
		})
	}

	// tags in root_block_device and ebs_block_device
	blockDeviceTagsToCreate := map[string]map[string]interface{}{}
	if v, ok := d.GetOk("root_block_device"); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			bd := v.(map[string]interface{})

			blockDeviceTags, ok := bd[names.AttrTags].(map[string]interface{})
			if !ok || len(blockDeviceTags) == 0 {
				continue
			}

			volID := getRootVolID(instance)
			if volID == "" {
				continue
			}

			blockDeviceTagsToCreate[volID] = blockDeviceTags
		}
	}

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})

			blockDeviceTags, ok := bd[names.AttrTags].(map[string]interface{})
			if !ok || len(blockDeviceTags) == 0 {
				continue
			}

			volID := getVolIDByDeviceName(instance, bd[names.AttrDeviceName].(string))
			if volID == "" {
				continue
			}

			blockDeviceTagsToCreate[volID] = blockDeviceTags
		}
	}

	for vol, blockDeviceTags := range blockDeviceTagsToCreate {
		if err := createTags(ctx, conn, vol, Tags(tftags.New(ctx, blockDeviceTags))); err != nil {
			log.Printf("[ERR] Error creating tags for EBS volume %s: %s", vol, err)
		}
	}

	// Update if we need to
	return append(diags, resourceInstanceUpdate(ctx, d, meta)...)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instance, err := findInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
	}

	instanceType := string(instance.InstanceType)
	instanceTypeInfo, err := findInstanceTypeByName(ctx, conn, instanceType)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance Type (%s): %s", instanceType, err)
	}

	d.Set("instance_state", instance.State.Name)

	if v := instance.Placement; v != nil {
		d.Set(names.AttrAvailabilityZone, v.AvailabilityZone)

		d.Set("placement_group", v.GroupName)

		d.Set("host_id", v.HostId)

		if v := v.HostResourceGroupArn; v != nil {
			d.Set("host_resource_group_arn", instance.Placement.HostResourceGroupArn)
		}

		d.Set("placement_partition_number", v.PartitionNumber)

		d.Set("tenancy", v.Tenancy)
	}

	// preserved to maintain backward compatibility
	if v := instance.CpuOptions; v != nil {
		d.Set("cpu_core_count", v.CoreCount)
		d.Set("cpu_threads_per_core", v.ThreadsPerCore)
	}

	if err := d.Set("cpu_options", flattenCPUOptions(instance.CpuOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cpu_options: %s", err)
	}

	if v := instance.HibernationOptions; v != nil {
		d.Set("hibernation", v.Configured)
	}

	if err := d.Set("enclave_options", flattenEnclaveOptions(instance.EnclaveOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting enclave_options: %s", err)
	}

	if instance.MaintenanceOptions != nil {
		if err := d.Set("maintenance_options", []interface{}{flattenInstanceMaintenanceOptions(instance.MaintenanceOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting maintenance_options: %s", err)
		}
	} else {
		d.Set("maintenance_options", nil)
	}

	if err := d.Set("metadata_options", flattenInstanceMetadataOptions(instance.MetadataOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metadata_options: %s", err)
	}

	if instance.PrivateDnsNameOptions != nil {
		if err := d.Set("private_dns_name_options", []interface{}{flattenPrivateDNSNameOptionsResponse(instance.PrivateDnsNameOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting private_dns_name_options: %s", err)
		}
	} else {
		d.Set("private_dns_name_options", nil)
	}

	d.Set("ami", instance.ImageId)
	d.Set(names.AttrInstanceType, instanceType)
	d.Set("key_name", instance.KeyName)
	d.Set("public_dns", instance.PublicDnsName)
	d.Set("public_ip", instance.PublicIpAddress)
	d.Set("private_dns", instance.PrivateDnsName)
	d.Set("private_ip", instance.PrivateIpAddress)
	d.Set("outpost_arn", instance.OutpostArn)

	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		name, err := instanceProfileARNToName(aws.ToString(instance.IamInstanceProfile.Arn))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting iam_instance_profile: %s", err)
		}

		d.Set("iam_instance_profile", name)
	} else {
		d.Set("iam_instance_profile", nil)
	}

	{
		launchTemplate, err := flattenInstanceLaunchTemplate(ctx, conn, d.Id(), d.Get("launch_template.0.version").(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s) launch template: %s", d.Id(), err)
		}

		if err := d.Set(names.AttrLaunchTemplate, launchTemplate); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
		}
	}

	// Set configured Network Interface Device Index Slice
	// We only want to read, and populate state for the configured network_interface attachments. Otherwise, other
	// resources have the potential to attach network interfaces to the instance, and cause a perpetual create/destroy
	// diff. We should only read on changes configured for this specific resource because of this.
	var configuredDeviceIndexes []int
	if v, ok := d.GetOk("network_interface"); ok {
		vL := v.(*schema.Set).List()
		for _, vi := range vL {
			mVi := vi.(map[string]interface{})
			configuredDeviceIndexes = append(configuredDeviceIndexes, mVi["device_index"].(int))
		}
	}

	var secondaryPrivateIPs []string
	var ipv6Addresses []string
	if len(instance.NetworkInterfaces) > 0 {
		var primaryNetworkInterface awstypes.InstanceNetworkInterface
		var networkInterfaces []map[string]interface{}
		for _, iNi := range instance.NetworkInterfaces {
			ni := make(map[string]interface{})
			if aws.ToInt32(iNi.Attachment.DeviceIndex) == 0 {
				primaryNetworkInterface = iNi
			}
			// If the attached network device is inside our configuration, refresh state with values found.
			// Otherwise, assume the network device was attached via an outside resource.
			for _, index := range configuredDeviceIndexes {
				if index == int(aws.ToInt32(iNi.Attachment.DeviceIndex)) {
					ni["device_index"] = aws.ToInt32(iNi.Attachment.DeviceIndex)
					ni["network_card_index"] = aws.ToInt32(iNi.Attachment.NetworkCardIndex)
					ni[names.AttrNetworkInterfaceID] = aws.ToString(iNi.NetworkInterfaceId)
					ni[names.AttrDeleteOnTermination] = aws.ToBool(iNi.Attachment.DeleteOnTermination)
				}
			}
			// Don't add empty network interfaces to schema
			if len(ni) == 0 {
				continue
			}
			networkInterfaces = append(networkInterfaces, ni)
		}
		if err := d.Set("network_interface", networkInterfaces); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting network_interfaces: %v", err)
		}

		// Set primary network interface details
		// If an instance is shutting down, network interfaces are detached, and attributes may be nil,
		// need to protect against nil pointer dereferences
		if primaryNetworkInterface.NetworkInterfaceId != nil {
			if primaryNetworkInterface.SubnetId != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
				d.Set(names.AttrSubnetID, primaryNetworkInterface.SubnetId)
			}
			if primaryNetworkInterface.NetworkInterfaceId != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
				d.Set("primary_network_interface_id", primaryNetworkInterface.NetworkInterfaceId)
			}
			d.Set("ipv6_address_count", len(primaryNetworkInterface.Ipv6Addresses))
			if primaryNetworkInterface.SourceDestCheck != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
				d.Set("source_dest_check", primaryNetworkInterface.SourceDestCheck)
			}

			d.Set("associate_public_ip_address", primaryNetworkInterface.Association != nil)

			for _, address := range primaryNetworkInterface.PrivateIpAddresses {
				if !aws.ToBool(address.Primary) {
					secondaryPrivateIPs = append(secondaryPrivateIPs, aws.ToString(address.PrivateIpAddress))
				}
			}

			for _, address := range primaryNetworkInterface.Ipv6Addresses {
				ipv6Addresses = append(ipv6Addresses, aws.ToString(address.Ipv6Address))
			}
		}
	} else {
		d.Set("associate_public_ip_address", instance.PublicIpAddress != nil)
		d.Set("ipv6_address_count", 0)
		d.Set("primary_network_interface_id", "")
		d.Set(names.AttrSubnetID, instance.SubnetId)
	}

	if err := d.Set("secondary_private_ips", secondaryPrivateIPs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_ips for AWS Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("ipv6_addresses", ipv6Addresses); err != nil {
		log.Printf("[WARN] Error setting ipv6_addresses for AWS Instance (%s): %s", d.Id(), err)
	}

	d.Set("ebs_optimized", instance.EbsOptimized)
	if aws.ToString(instance.SubnetId) != "" {
		d.Set("source_dest_check", instance.SourceDestCheck)
	}

	if instance.Monitoring != nil && instance.Monitoring.State != "" {
		monitoringState := instance.Monitoring.State
		d.Set("monitoring", monitoringState == awstypes.MonitoringStateEnabled || monitoringState == awstypes.MonitoringStatePending)
	}

	setTagsOut(ctx, instance.Tags)

	if _, ok := d.GetOk("volume_tags"); ok && !blockDeviceTagsDefined(d) {
		volumeTags, err := readVolumeTags(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
		}

		defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
		ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
		tags := keyValueTags(ctx, volumeTags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		if err := d.Set("volume_tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting volume_tags: %s", err)
		}
	}

	if err := readSecurityGroups(ctx, d, instance, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
	}

	// Retrieve instance shutdown behavior
	if err := readInstanceShutdownBehavior(ctx, d, conn); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
	}

	if err := readBlockDevices(ctx, d, meta, instance, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
	}

	if _, ok := d.GetOk("ephemeral_block_device"); !ok {
		d.Set("ephemeral_block_device", []interface{}{})
	}

	// ARN

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   names.EC2,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("instance/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	// Instance attributes
	{
		attr, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
			Attribute:  awstypes.InstanceAttributeNameDisableApiStop,
			InstanceId: aws.String(d.Id()),
		})
		if err != nil && !errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			return sdkdiag.AppendErrorf(diags, "getting attribute (%s): %s", awstypes.InstanceAttributeNameDisableApiStop, err)
		}
		if !errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			d.Set("disable_api_stop", attr.DisableApiStop.Value)
		}
	}
	{
		if isSnowballEdgeInstance(d.Id()) {
			log.Printf("[INFO] Determined deploying to Snowball Edge based off Instance ID %s. Skip setting the 'disable_api_termination' attribute.", d.Id())
		} else {
			output, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
				Attribute:  awstypes.InstanceAttributeNameDisableApiTermination,
				InstanceId: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "getting attribute (%s): %s", awstypes.InstanceAttributeNameDisableApiTermination, err)
			}

			d.Set("disable_api_termination", output.DisableApiTermination.Value)
		}
	}
	{
		attr, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
			Attribute:  awstypes.InstanceAttributeNameUserData,
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting attribute (%s): %s", awstypes.InstanceAttributeNameUserData, err)
		}
		if attr.UserData != nil && attr.UserData.Value != nil {
			// Since user_data and user_data_base64 conflict with each other,
			// we'll only set one or the other here to avoid a perma-diff.
			// Since user_data_base64 was added later, we'll prefer to set
			// user_data.
			_, b64 := d.GetOk("user_data_base64")
			if b64 {
				d.Set("user_data_base64", attr.UserData.Value)
			} else {
				d.Set("user_data", userDataHashSum(aws.ToString(attr.UserData.Value)))
			}
		}
	}

	// AWS Standard will return InstanceCreditSpecification.NotSupported errors for EC2 Instance IDs outside T2 and T3 instance types
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8055
	if aws.ToBool(instanceTypeInfo.BurstablePerformanceSupported) {
		instanceCreditSpecification, err := findInstanceCreditSpecificationByID(ctx, conn, d.Id())

		// Ignore UnsupportedOperation errors for AWS China and GovCloud (US).
		// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/4362.
		if tfawserr.ErrCodeEquals(err, errCodeUnsupportedOperation) {
			err = nil
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s) credit specification: %s", d.Id(), err)
		}

		if instanceCreditSpecification != nil {
			if err := d.Set("credit_specification", []interface{}{flattenInstanceCreditSpecification(instanceCreditSpecification)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting credit_specification: %s", err)
			}
		} else {
			d.Set("credit_specification", nil)
		}
	}

	if d.Get("get_password_data").(bool) {
		passwordData, err := getInstancePasswordData(ctx, aws.ToString(instance.InstanceId), conn, d.Timeout(schema.TimeoutRead))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
		}
		d.Set("password_data", passwordData)
	} else {
		d.Set("get_password_data", false)
		d.Set("password_data", nil)
	}

	if instance.CapacityReservationSpecification != nil {
		if err := d.Set("capacity_reservation_specification", []interface{}{flattenCapacityReservationSpecificationResponse(instance.CapacityReservationSpecification)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting capacity_reservation_specification: %s", err)
		}
	} else {
		d.Set("capacity_reservation_specification", nil)
	}

	if spotInstanceRequestID := aws.ToString(instance.SpotInstanceRequestId); spotInstanceRequestID != "" && instance.InstanceLifecycle != "" {
		d.Set("instance_lifecycle", instance.InstanceLifecycle)
		d.Set("spot_instance_request_id", spotInstanceRequestID)

		input := &ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []string{spotInstanceRequestID},
		}

		apiObject, err := findSpotInstanceRequest(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Spot Instance Request (%s): %s", spotInstanceRequestID, err)
		}

		tfMap := map[string]interface{}{
			"instance_interruption_behavior": apiObject.InstanceInterruptionBehavior,
			"spot_instance_type":             apiObject.Type,
		}

		if v := apiObject.SpotPrice; v != nil {
			tfMap["max_price"] = aws.ToString(v)
		}

		if v := apiObject.ValidUntil; v != nil {
			tfMap["valid_until"] = aws.ToTime(v).Format(time.RFC3339)
		}

		if err := d.Set("instance_market_options", []interface{}{map[string]interface{}{
			"market_type":  awstypes.MarketTypeSpot,
			"spot_options": []interface{}{tfMap},
		}}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting instance_market_options: %s", err)
		}
	} else {
		d.Set("instance_lifecycle", nil)
		d.Set("instance_market_options", nil)
		d.Set("spot_instance_request_id", nil)
	}

	return diags
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("volume_tags") && !d.IsNewResource() {
		volIDs, err := getInstanceVolIDs(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
		}

		o, n := d.GetChange("volume_tags")

		for _, volID := range volIDs {
			if err := updateTags(ctx, conn, volID, o, n); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating volume_tags (%s): %s", volID, err)
			}
		}
	}

	if d.HasChange("iam_instance_profile") && !d.IsNewResource() {
		request := &ec2.DescribeIamInstanceProfileAssociationsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("instance-id"),
					Values: []string{d.Id()},
				},
			},
		}

		resp, err := conn.DescribeIamInstanceProfileAssociations(ctx, request)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
		}

		// An Iam Instance Profile has been provided and is pending a change
		// This means it is an association or a replacement to an association
		if _, ok := d.GetOk("iam_instance_profile"); ok {
			// Does not have an Iam Instance Profile associated with it, need to associate
			if len(resp.IamInstanceProfileAssociations) == 0 {
				if err := associateInstanceProfile(ctx, d, conn); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
				}
			} else {
				// Has an Iam Instance Profile associated with it, need to replace the association
				associationId := resp.IamInstanceProfileAssociations[0].AssociationId
				input := &ec2.ReplaceIamInstanceProfileAssociationInput{
					AssociationId: associationId,
					IamInstanceProfile: &awstypes.IamInstanceProfileSpecification{
						Name: aws.String(d.Get("iam_instance_profile").(string)),
					},
				}

				// If the instance is running, we can replace the instance profile association.
				// If it is stopped, the association must be removed and the new one attached separately. (GH-8262)
				instanceState := awstypes.InstanceStateName(d.Get("instance_state").(string))

				if instanceState != "" {
					if instanceState == awstypes.InstanceStateNameStopped || instanceState == awstypes.InstanceStateNameStopping || instanceState == awstypes.InstanceStateNameShuttingDown {
						if err := disassociateInstanceProfile(ctx, associationId, conn); err != nil {
							return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
						}
						if err := associateInstanceProfile(ctx, d, conn); err != nil {
							return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
						}
					} else {
						err := retry.RetryContext(ctx, iamPropagationTimeout, func() *retry.RetryError {
							_, err := conn.ReplaceIamInstanceProfileAssociation(ctx, input)
							if err != nil {
								if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Invalid IAM Instance Profile") {
									return retry.RetryableError(err)
								}
								return retry.NonRetryableError(err)
							}
							return nil
						})
						if tfresource.TimedOut(err) {
							_, err = conn.ReplaceIamInstanceProfileAssociation(ctx, input)
						}
						if err != nil {
							return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): replacing instance profile: %s", d.Id(), err)
						}
					}
				}
			}
			// An Iam Instance Profile has _not_ been provided but is pending a change. This means there is a pending removal
		} else {
			if len(resp.IamInstanceProfileAssociations) > 0 {
				// Has an Iam Instance Profile associated with it, need to remove the association
				associationId := resp.IamInstanceProfileAssociations[0].AssociationId
				if err := disassociateInstanceProfile(ctx, associationId, conn); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
				}
			}
		}

		if _, err := waitInstanceIAMInstanceProfileUpdated(ctx, conn, d.Id(), d.Get("iam_instance_profile").(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) IAM Instance Profile update: %s", d.Id(), err)
		}
	}

	// SourceDestCheck can only be modified on an instance without manually specified network interfaces.
	// SourceDestCheck, in that case, is configured at the network interface level
	if _, ok := d.GetOk("network_interface"); !ok {
		// If we have a new resource and source_dest_check is still true, don't modify
		sourceDestCheck := d.Get("source_dest_check").(bool)

		// Because we're calling Update prior to Read, and the default value of `source_dest_check` is `true`,
		// HasChange() thinks there is a diff between what is set on the instance and what is set in state. We need to ensure that
		// if a diff has occurred, it's not because it's a new instance.
		if d.HasChange("source_dest_check") && !d.IsNewResource() || d.IsNewResource() && !sourceDestCheck {
			input := &ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(d.Id()),
				SourceDestCheck: &awstypes.AttributeBooleanValue{
					Value: aws.Bool(sourceDestCheck),
				},
			}

			_, err := conn.ModifyInstanceAttribute(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying EC2 Instance (%s) SourceDestCheck attribute: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_address_count") && !d.IsNewResource() {
		instance, err := findInstanceByID(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
		}

		var primaryInterface awstypes.InstanceNetworkInterface
		for _, ni := range instance.NetworkInterfaces {
			if aws.ToInt32(ni.Attachment.DeviceIndex) == 0 {
				primaryInterface = ni
			}
		}

		if primaryInterface.NetworkInterfaceId == nil {
			return sdkdiag.AppendErrorf(diags, "Failed to update ipv6_address_count on %q, which does not contain a primary network interface", d.Id())
		}

		o, n := d.GetChange("ipv6_address_count")
		os, ns := o.(int), n.(int)

		if ns > os {
			// Add more to the primary NIC.
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: primaryInterface.NetworkInterfaceId,
				Ipv6AddressCount:   aws.Int32(int32(ns - os)),
			}

			_, err := conn.AssignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Instance (%s) IPv6 addresses: %s", d.Id(), err)
			}
		} else if os > ns {
			// Remove IP addresses.
			if len(primaryInterface.Ipv6Addresses) != os {
				return sdkdiag.AppendErrorf(diags, "IPv6 address count (%d) on the instance does not match state's count (%d), we're in a race with something else", len(primaryInterface.Ipv6Addresses), os)
			}

			toRemove := make([]string, 0)
			for _, addr := range primaryInterface.Ipv6Addresses[ns:] { // Can I assume this is strongly ordered?
				toRemove = append(toRemove, *addr.Ipv6Address)
			}

			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: primaryInterface.NetworkInterfaceId,
				Ipv6Addresses:      toRemove,
			}

			_, err := conn.UnassignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Instance (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("secondary_private_ips", names.AttrVPCSecurityGroupIDs) && !d.IsNewResource() {
		instance, err := findInstanceByID(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", d.Id(), err)
		}

		var primaryInterface awstypes.InstanceNetworkInterface
		for _, ni := range instance.NetworkInterfaces {
			if aws.ToInt32(ni.Attachment.DeviceIndex) == 0 {
				primaryInterface = ni
			}
		}

		if d.HasChange("secondary_private_ips") {
			if primaryInterface.NetworkInterfaceId == nil {
				return sdkdiag.AppendErrorf(diags, "Failed to update secondary_private_ips on %q, which does not contain a primary network interface",
					d.Id())
			}
			o, n := d.GetChange("secondary_private_ips")
			if o == nil {
				o = new(schema.Set)
			}
			if n == nil {
				n = new(schema.Set)
			}

			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			// Unassign old IP addresses
			unassignIps := os.Difference(ns)
			if unassignIps.Len() != 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: primaryInterface.NetworkInterfaceId,
					PrivateIpAddresses: flex.ExpandStringValueSet(unassignIps),
				}
				log.Printf("[INFO] Unassigning secondary_private_ips on Instance %q", d.Id())
				_, err := conn.UnassignPrivateIpAddresses(ctx, input)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "Failure to unassign Secondary Private IPs: %s", err)
				}
			}

			// Assign new IP addresses
			assignIps := ns.Difference(os)
			if assignIps.Len() != 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId: primaryInterface.NetworkInterfaceId,
					PrivateIpAddresses: flex.ExpandStringValueSet(assignIps),
				}
				log.Printf("[INFO] Assigning secondary_private_ips on Instance %q", d.Id())
				_, err := conn.AssignPrivateIpAddresses(ctx, input)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "Failure to assign Secondary Private IPs: %s", err)
				}
			}
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			if primaryInterface.NetworkInterfaceId == nil {
				return sdkdiag.AppendErrorf(diags, "Failed to update vpc_security_group_ids on %q, which does not contain a primary network interface",
					d.Id())
			}
			var groups []string
			if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
				for _, v := range v.List() {
					groups = append(groups, v.(string))
				}
			}

			if len(groups) < 1 {
				return sdkdiag.AppendErrorf(diags, "VPC-based instances require at least one security group to be attached.")
			}
			// If a user has multiple network interface attachments on the target EC2 instance, simply modifying the
			// instance attributes via a `ModifyInstanceAttributes()` request would fail with the following error message:
			// "There are multiple interfaces attached to instance 'i-XX'. Please specify an interface ID for the operation instead."
			// Thus, we need to actually modify the primary network interface for the new security groups, as the primary
			// network interface is where we modify/create security group assignments during Create.
			log.Printf("[INFO] Modifying `vpc_security_group_ids` on Instance %q", d.Id())
			if _, err := conn.ModifyNetworkInterfaceAttribute(ctx, &ec2.ModifyNetworkInterfaceAttributeInput{
				NetworkInterfaceId: primaryInterface.NetworkInterfaceId,
				Groups:             groups,
			}); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): modifying network interface: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges(names.AttrInstanceType, "user_data", "user_data_base64") && !d.IsNewResource() {
		// For each argument change, we start and stop the instance
		// to account for behaviors occurring outside terraform.
		// Only one attribute can be modified at a time, else we get
		// "InvalidParameterCombination: Fields for multiple attribute types specified"
		if d.HasChange(names.AttrInstanceType) {
			if !d.HasChange("capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id") {
				instanceType := d.Get(names.AttrInstanceType).(string)
				input := &ec2.ModifyInstanceAttributeInput{
					InstanceId: aws.String(d.Id()),
					InstanceType: &awstypes.AttributeValue{
						Value: aws.String(instanceType),
					},
				}

				if err := modifyInstanceAttributeWithStopStart(ctx, conn, input, fmt.Sprintf("InstanceType (%s)", instanceType)); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) type: %s", d.Id(), err)
				}
			}
		}

		// From the API reference:
		// "If you are using an AWS SDK or command line tool,
		// base64-encoding is performed for you, and you can load the text from a file.
		// Otherwise, you must provide base64-encoded text".

		if d.HasChange("user_data") {
			// Decode so the AWS SDK doesn't double encode.
			v, err := itypes.Base64Decode(d.Get("user_data").(string))
			if err != nil {
				v = []byte(d.Get("user_data").(string))
			}

			input := &ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(d.Id()),
				UserData: &awstypes.BlobAttributeValue{
					Value: v,
				},
			}

			if err := modifyInstanceAttributeWithStopStart(ctx, conn, input, "UserData"); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) user data: %s", d.Id(), err)
			}
		}

		if d.HasChange("user_data_base64") {
			// Schema validation technically ensures the data is Base64 encoded.
			// Decode so the AWS SDK doesn't double encode.
			v, err := itypes.Base64Decode(d.Get("user_data_base64").(string))
			if err != nil {
				v = []byte(d.Get("user_data_base64").(string))
			}

			input := &ec2.ModifyInstanceAttributeInput{
				InstanceId: aws.String(d.Id()),
				UserData: &awstypes.BlobAttributeValue{
					Value: v,
				},
			}

			if err := modifyInstanceAttributeWithStopStart(ctx, conn, input, "UserData (base64)"); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) user data base64: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("disable_api_stop") && !d.IsNewResource() {
		if err := disableInstanceAPIStop(ctx, conn, d.Id(), d.Get("disable_api_stop").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("disable_api_termination") && !d.IsNewResource() {
		if err := disableInstanceAPITermination(ctx, conn, d.Id(), d.Get("disable_api_termination").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("instance_initiated_shutdown_behavior") {
		input := &ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(d.Id()),
			InstanceInitiatedShutdownBehavior: &awstypes.AttributeValue{
				Value: aws.String(d.Get("instance_initiated_shutdown_behavior").(string)),
			},
		}

		_, err := conn.ModifyInstanceAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Instance (%s) InstanceInitiatedShutdownBehavior attribute: %s", d.Id(), err)
		}
	}

	if d.HasChange("maintenance_options") && !d.IsNewResource() {
		autoRecovery := d.Get("maintenance_options.0.auto_recovery").(string)

		log.Printf("[INFO] Modifying instance automatic recovery settings %s", d.Id())
		_, err := conn.ModifyInstanceMaintenanceOptions(ctx, &ec2.ModifyInstanceMaintenanceOptionsInput{
			AutoRecovery: awstypes.InstanceAutoRecoveryState(autoRecovery),
			InstanceId:   aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): modifying maintenance options: %s", d.Id(), err)
		}

		if _, err := waitInstanceMaintenanceOptionsAutoRecoveryUpdated(ctx, conn, d.Id(), autoRecovery, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): modifying maintenance options: waiting for completion: %s", d.Id(), err)
		}
	}

	if d.HasChange("monitoring") {
		var mErr error
		if d.Get("monitoring").(bool) {
			log.Printf("[DEBUG] Enabling monitoring for Instance (%s)", d.Id())
			_, mErr = conn.MonitorInstances(ctx, &ec2.MonitorInstancesInput{
				InstanceIds: []string{d.Id()},
			})
		} else {
			log.Printf("[DEBUG] Disabling monitoring for Instance (%s)", d.Id())
			_, mErr = conn.UnmonitorInstances(ctx, &ec2.UnmonitorInstancesInput{
				InstanceIds: []string{d.Id()},
			})
		}
		if mErr != nil {
			return sdkdiag.AppendErrorf(diags, "updating Instance monitoring: %s", mErr)
		}
	}

	if d.HasChange("credit_specification") && !d.IsNewResource() {
		if v, ok := d.GetOk("credit_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			instanceCreditSpecification := expandInstanceCreditSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
			instanceCreditSpecification.InstanceId = aws.String(d.Id())
			input := &ec2.ModifyInstanceCreditSpecificationInput{
				ClientToken:                  aws.String(id.UniqueId()),
				InstanceCreditSpecifications: []awstypes.InstanceCreditSpecificationRequest{instanceCreditSpecification},
			}

			log.Printf("[DEBUG] Modifying EC2 Instance credit specification: %s", d.Id())
			_, err := conn.ModifyInstanceCreditSpecification(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) credit specification: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("metadata_options") && !d.IsNewResource() {
		if v, ok := d.GetOk("metadata_options"); ok {
			if tfMap, ok := v.([]interface{})[0].(map[string]interface{}); ok {
				input := &ec2.ModifyInstanceMetadataOptionsInput{
					HttpEndpoint: awstypes.InstanceMetadataEndpointState(tfMap["http_endpoint"].(string)),
					InstanceId:   aws.String(d.Id()),
				}

				if tfMap["http_endpoint"].(string) == string(awstypes.InstanceMetadataEndpointStateEnabled) {
					// These parameters are not allowed unless HttpEndpoint is enabled.
					input.HttpProtocolIpv6 = awstypes.InstanceMetadataProtocolState(tfMap["http_protocol_ipv6"].(string))
					input.HttpPutResponseHopLimit = aws.Int32(int32(tfMap["http_put_response_hop_limit"].(int)))
					input.HttpTokens = awstypes.HttpTokensState(tfMap["http_tokens"].(string))
					input.InstanceMetadataTags = awstypes.InstanceMetadataTagsState(tfMap["instance_metadata_tags"].(string))
				}

				_, err := conn.ModifyInstanceMetadataOptions(ctx, input)
				if tfawserr.ErrMessageContains(err, errCodeUnsupportedOperation, "InstanceMetadataTags") {
					log.Printf("[WARN] updating EC2 Instance (%s) metadata options: %s. Retrying without instance metadata tags.", d.Id(), err)

					_, err = conn.ModifyInstanceMetadataOptions(ctx, input)
				}
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) metadata options: %s", d.Id(), err)
				}

				if _, err := waitInstanceMetadataOptionsApplied(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) metadata options update: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("root_block_device.0") && !d.IsNewResource() {
		volID := d.Get("root_block_device.0.volume_id").(string)

		input := &ec2.ModifyVolumeInput{
			VolumeId: aws.String(volID),
		}
		modifyVolume := false

		if d.HasChange("root_block_device.0.volume_size") {
			if v, ok := d.Get("root_block_device.0.volume_size").(int); ok && v != 0 {
				modifyVolume = true
				input.Size = aws.Int32(int32(v))
			}
		}
		if d.HasChange("root_block_device.0.volume_type") {
			if v, ok := d.Get("root_block_device.0.volume_type").(string); ok && v != "" {
				modifyVolume = true
				input.VolumeType = awstypes.VolumeType(v)
			}
		}
		if d.HasChange("root_block_device.0.iops") {
			if v, ok := d.Get("root_block_device.0.iops").(int); ok && v != 0 {
				// Enforce IOPs usage with a valid volume type
				// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
				if t, ok := d.Get("root_block_device.0.volume_type").(string); ok && t != string(awstypes.VolumeTypeIo1) && t != string(awstypes.VolumeTypeIo2) && t != string(awstypes.VolumeTypeGp3) {
					if t == "" {
						// Volume defaults to gp2
						t = string(awstypes.VolumeTypeGp2)
					}
					return sdkdiag.AppendErrorf(diags, "updating instance: iops attribute not supported for type %s", t)
				}
				modifyVolume = true
				input.Iops = aws.Int32(int32(v))
			}
		}
		if d.HasChange("root_block_device.0.throughput") {
			if v, ok := d.Get("root_block_device.0.throughput").(int); ok && v != 0 {
				// Enforce throughput usage with a valid volume type
				if t, ok := d.Get("root_block_device.0.volume_type").(string); ok && t != string(awstypes.VolumeTypeGp3) {
					return sdkdiag.AppendErrorf(diags, "updating instance: throughput attribute not supported for type %s", t)
				}
				modifyVolume = true
				input.Throughput = aws.Int32(int32(v))
			}
		}
		if modifyVolume {
			log.Printf("[DEBUG] Modifying volume: %s", d.Id())
			_, err := conn.ModifyVolume(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) volume (%s): %s", d.Id(), volID, err)
			}

			if _, err := waitVolumeModificationComplete(ctx, conn, volID, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) volume (%s) update: %s", d.Id(), volID, err)
			}
		}

		if d.HasChange("root_block_device.0.delete_on_termination") {
			deviceName := d.Get("root_block_device.0.device_name").(string)
			if v, ok := d.Get("root_block_device.0.delete_on_termination").(bool); ok {
				input := &ec2.ModifyInstanceAttributeInput{
					BlockDeviceMappings: []awstypes.InstanceBlockDeviceMappingSpecification{
						{
							DeviceName: aws.String(deviceName),
							Ebs: &awstypes.EbsInstanceBlockDeviceSpecification{
								DeleteOnTermination: aws.Bool(v),
							},
						},
					},
					InstanceId: aws.String(d.Id()),
				}

				_, err := conn.ModifyInstanceAttribute(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying EC2 Instance (%s) BlockDeviceMappings (%s) attribute: %s", d.Id(), deviceName, err)
				}

				if _, err := waitInstanceRootBlockDeviceDeleteOnTerminationUpdated(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) root block device DeleteOnTermination update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("root_block_device.0.tags") {
			o, n := d.GetChange("root_block_device.0.tags")

			if err := updateTags(ctx, conn, volID, o, n); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating tags for volume (%s): %s", volID, err)
			}
		}

		if d.HasChange("root_block_device.0.tags_all") && !d.HasChange("root_block_device.0.tags") {
			o, n := d.GetChange("root_block_device.0.tags_all")

			if err := updateTags(ctx, conn, volID, o, n); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating tags for volume (%s): %s", volID, err)
			}
		}
	}

	// To modify capacity reservation attributes of an instance, instance state needs to be in ec2.InstanceStateNameStopped,
	// otherwise the modification will return an IncorrectInstanceState error
	if d.HasChange("capacity_reservation_specification") && !d.IsNewResource() {
		if v, ok := d.GetOk("capacity_reservation_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			if v := expandCapacityReservationSpecification(v.([]interface{})[0].(map[string]interface{})); v != nil && (v.CapacityReservationPreference != "" || v.CapacityReservationTarget != nil) {
				if err := stopInstance(ctx, conn, d.Id(), false, instanceStopTimeout); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				if d.HasChange("capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id") && d.HasChange(names.AttrInstanceType) {
					instanceType := d.Get(names.AttrInstanceType).(string)
					input := &ec2.ModifyInstanceAttributeInput{
						InstanceId: aws.String(d.Id()),
						InstanceType: &awstypes.AttributeValue{
							Value: aws.String(instanceType),
						},
					}

					if _, err := conn.ModifyInstanceAttribute(ctx, input); err != nil {
						return sdkdiag.AppendErrorf(diags, "modifying EC2 Instance (%s) InstanceType (%s) attribute: %s", d.Id(), instanceType, err)
					}
				}

				input := &ec2.ModifyInstanceCapacityReservationAttributesInput{
					CapacityReservationSpecification: v,
					InstanceId:                       aws.String(d.Id()),
				}

				log.Printf("[DEBUG] Modifying EC2 Instance capacity reservation attributes: %s", d.Id())

				_, err := conn.ModifyInstanceCapacityReservationAttributes(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s) capacity reservation attributes: %s", d.Id(), err)
				}

				if _, err := waitInstanceCapacityReservationSpecificationUpdated(ctx, conn, d.Id(), v); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for EC2 Instance (%s) capacity reservation attributes update: %s", d.Id(), err)
				}

				if err := startInstance(ctx, conn, d.Id(), true, instanceStartTimeout); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	if d.HasChange("private_dns_name_options") && !d.IsNewResource() {
		if v, ok := d.GetOk("private_dns_name_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			input := &ec2.ModifyPrivateDnsNameOptionsInput{
				InstanceId: aws.String(d.Id()),
			}

			if d.HasChange("private_dns_name_options.0.enable_resource_name_dns_aaaa_record") {
				input.EnableResourceNameDnsAAAARecord = aws.Bool(tfMap["enable_resource_name_dns_aaaa_record"].(bool))
			}

			if d.HasChange("private_dns_name_options.0.enable_resource_name_dns_a_record") {
				input.EnableResourceNameDnsARecord = aws.Bool(tfMap["enable_resource_name_dns_a_record"].(bool))
			}

			if d.HasChange("private_dns_name_options.0.hostname_type") {
				input.PrivateDnsHostnameType = awstypes.HostnameType(tfMap["hostname_type"].(string))
			}

			_, err := conn.ModifyPrivateDnsNameOptions(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): modifying private DNS name options: %s", d.Id(), err)
			}
		}
	}

	// TODO(mitchellh): wait for the attributes we modified to
	// persist the change...

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := disableInstanceAPITermination(ctx, conn, d.Id(), false); err != nil {
		log.Printf("[WARN] attempting to terminate EC2 Instance (%s) despite error disabling API termination: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("disable_api_stop"); ok {
		if err := disableInstanceAPIStop(ctx, conn, d.Id(), v.(bool)); err != nil {
			log.Printf("[WARN] attempting to terminate EC2 Instance (%s) despite error disabling API stop: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("instance_lifecycle"); ok && v == awstypes.InstanceLifecycleSpot {
		spotInstanceRequestID := d.Get("spot_instance_request_id").(string)
		_, err := conn.CancelSpotInstanceRequests(ctx, &ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []string{spotInstanceRequestID},
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "cancelling EC2 Spot Fleet Request (%s): %s", spotInstanceRequestID, err)
		}
	}

	if err := terminateInstance(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func disableInstanceAPIStop(ctx context.Context, conn *ec2.Client, id string, disableAPIStop bool) error {
	input := &ec2.ModifyInstanceAttributeInput{
		DisableApiStop: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(disableAPIStop),
		},
		InstanceId: aws.String(id),
	}

	_, err := conn.ModifyInstanceAttribute(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeUnsupportedOperation, "not supported for spot instances") {
		log.Printf("[WARN] failed to modify EC2 Instance (%s) attribute: %s", id, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("modifying EC2 Instance (%s) DisableApiStop attribute: %s", id, err)
	}

	return nil
}

func disableInstanceAPITermination(ctx context.Context, conn *ec2.Client, id string, disableAPITermination bool) error {
	input := &ec2.ModifyInstanceAttributeInput{
		DisableApiTermination: &awstypes.AttributeBooleanValue{
			Value: aws.Bool(disableAPITermination),
		},
		InstanceId: aws.String(id),
	}

	_, err := conn.ModifyInstanceAttribute(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeUnsupportedOperation, "not supported for spot instances") {
		log.Printf("[WARN] failed to modify EC2 Instance (%s) attribute: %s", id, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("modifying EC2 Instance (%s) DisableApiTermination attribute: %s", id, err)
	}

	return nil
}

// modifyInstanceAttributeWithStopStart modifies a specific attribute provided
// as input by first stopping the EC2 instance before the modification
// and then starting up the EC2 instance after modification.
// Reference: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Stop_Start.html
func modifyInstanceAttributeWithStopStart(ctx context.Context, conn *ec2.Client, input *ec2.ModifyInstanceAttributeInput, attrName string) error {
	id := aws.ToString(input.InstanceId)

	if err := stopInstance(ctx, conn, id, false, instanceStopTimeout); err != nil {
		return err
	}

	if _, err := conn.ModifyInstanceAttribute(ctx, input); err != nil {
		return fmt.Errorf("modifying EC2 Instance (%s) %s attribute: %w", id, attrName, err)
	}

	if err := startInstance(ctx, conn, id, true, instanceStartTimeout); err != nil {
		return err
	}

	return nil
}

func readBlockDevices(ctx context.Context, d *schema.ResourceData, meta interface{}, instance *awstypes.Instance, ds bool) error {
	ibds, err := readBlockDevicesFromInstance(ctx, d, meta, instance, ds)
	if err != nil {
		return fmt.Errorf("reading block devices: %w", err)
	}

	// Special handling for instances where the only block device is the root device:
	// The call to readBlockDevicesFromInstance above will return the block device
	// in ibds["root"] not ibds["ebs"], thus to set the state correctly,
	// the root block device must be copied over to ibds["ebs"]
	if ibds != nil {
		if _, ok := d.GetOk("ebs_block_device"); ok {
			if v, ok := ibds["ebs"].([]map[string]interface{}); ok && len(v) == 0 {
				if root, ok := ibds["root"].(map[string]interface{}); ok {
					// Make deep copy of data
					m := make(map[string]interface{})

					for k, v := range root {
						m[k] = v
					}

					if snapshotID, ok := ibds[names.AttrSnapshotID].(string); ok {
						m[names.AttrSnapshotID] = snapshotID
					}

					ibds["ebs"] = []interface{}{m}
				}
			}
		}
	}

	if err := d.Set("ebs_block_device", ibds["ebs"]); err != nil {
		return err // nosemgrep:ci.bare-error-returns
	}

	// This handles the import case which needs to be defaulted to empty
	if _, ok := d.GetOk("root_block_device"); !ok {
		if err := d.Set("root_block_device", []interface{}{}); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}

	if ibds["root"] != nil {
		roots := []interface{}{ibds["root"]}
		if err := d.Set("root_block_device", roots); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}

	return nil
}

func readBlockDevicesFromInstance(ctx context.Context, d *schema.ResourceData, meta interface{}, instance *awstypes.Instance, ds bool) (map[string]interface{}, error) {
	blockDevices := make(map[string]interface{})
	blockDevices["ebs"] = make([]map[string]interface{}, 0)
	blockDevices["root"] = nil
	// Ephemeral devices don't show up in BlockDeviceMappings or DescribeVolumes so we can't actually set them

	instanceBlockDevices := make(map[string]awstypes.InstanceBlockDeviceMapping)
	for _, bd := range instance.BlockDeviceMappings {
		if bd.Ebs != nil {
			instanceBlockDevices[aws.ToString(bd.Ebs.VolumeId)] = bd
		}
	}

	if len(instanceBlockDevices) == 0 {
		return nil, nil
	}

	volIDs := make([]string, 0, len(instanceBlockDevices))
	for volID := range instanceBlockDevices {
		volIDs = append(volIDs, volID)
	}

	// Need to call DescribeVolumes to get volume_size and volume_type for each
	// EBS block device
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	volResp, err := conn.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: volIDs,
	})
	if err != nil {
		return nil, err
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	for _, vol := range volResp.Volumes {
		instanceBd := instanceBlockDevices[aws.ToString(vol.VolumeId)]
		bd := make(map[string]interface{})

		bd["volume_id"] = aws.ToString(vol.VolumeId)

		if instanceBd.Ebs != nil && instanceBd.Ebs.DeleteOnTermination != nil {
			bd[names.AttrDeleteOnTermination] = aws.ToBool(instanceBd.Ebs.DeleteOnTermination)
		}
		if vol.Size != nil {
			bd[names.AttrVolumeSize] = aws.ToInt32(vol.Size)
		}
		if vol.VolumeType != "" {
			bd[names.AttrVolumeType] = vol.VolumeType
		}
		if vol.Iops != nil {
			bd[names.AttrIOPS] = aws.ToInt32(vol.Iops)
		}
		if vol.Encrypted != nil {
			bd[names.AttrEncrypted] = aws.ToBool(vol.Encrypted)
		}
		if vol.KmsKeyId != nil {
			bd[names.AttrKMSKeyID] = aws.ToString(vol.KmsKeyId)
		}
		if vol.Throughput != nil {
			bd[names.AttrThroughput] = aws.ToInt32(vol.Throughput)
		}
		if instanceBd.DeviceName != nil {
			bd[names.AttrDeviceName] = aws.ToString(instanceBd.DeviceName)
		}
		if v, ok := d.GetOk("volume_tags"); !ok || v == nil || len(v.(map[string]interface{})) == 0 {
			if ds {
				bd[names.AttrTags] = keyValueTags(ctx, vol.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()
			} else {
				tags := keyValueTags(ctx, vol.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
				bd[names.AttrTags] = tags.RemoveDefaultConfig(defaultTagsConfig).Map()
				bd[names.AttrTagsAll] = tags.Map()
			}
		}

		if blockDeviceIsRoot(instanceBd, instance) {
			blockDevices["root"] = bd
		} else {
			if vol.SnapshotId != nil {
				bd[names.AttrSnapshotID] = aws.ToString(vol.SnapshotId)
			}

			blockDevices["ebs"] = append(blockDevices["ebs"].([]map[string]interface{}), bd)
		}
	}
	// If we determine the root device is the only block device mapping
	// in the instance (including ephemerals) after returning from this function,
	// we'll need to set the ebs_block_device as a clone of the root device
	// with the snapshot_id populated; thus, we store the ID for safe-keeping
	if blockDevices["root"] != nil && len(blockDevices["ebs"].([]map[string]interface{})) == 0 {
		blockDevices[names.AttrSnapshotID] = volResp.Volumes[0].SnapshotId
	}

	return blockDevices, nil
}

func blockDeviceIsRoot(bd awstypes.InstanceBlockDeviceMapping, instance *awstypes.Instance) bool {
	return bd.DeviceName != nil &&
		instance.RootDeviceName != nil &&
		aws.ToString(bd.DeviceName) == aws.ToString(instance.RootDeviceName)
}

func associateInstanceProfile(ctx context.Context, d *schema.ResourceData, conn *ec2.Client) error {
	input := &ec2.AssociateIamInstanceProfileInput{
		InstanceId: aws.String(d.Id()),
		IamInstanceProfile: &awstypes.IamInstanceProfileSpecification{
			Name: aws.String(d.Get("iam_instance_profile").(string)),
		},
	}
	err := retry.RetryContext(ctx, iamPropagationTimeout, func() *retry.RetryError {
		_, err := conn.AssociateIamInstanceProfile(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Invalid IAM Instance Profile") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AssociateIamInstanceProfile(ctx, input)
	}
	if err != nil {
		return fmt.Errorf("associating instance profile: %s", err)
	}
	return nil
}

func disassociateInstanceProfile(ctx context.Context, associationId *string, conn *ec2.Client) error {
	_, err := conn.DisassociateIamInstanceProfile(ctx, &ec2.DisassociateIamInstanceProfileInput{
		AssociationId: associationId,
	})
	if err != nil {
		return fmt.Errorf("disassociating instance profile: %w", err)
	}
	return nil
}

func findRootDeviceName(ctx context.Context, conn *ec2.Client, amiID string) (*string, error) {
	if amiID == "" {
		return nil, errors.New("Cannot fetch root device name for blank AMI ID.")
	}

	image, err := findImageByID(ctx, conn, amiID)

	if err != nil {
		return nil, err
	}

	rootDeviceName := image.RootDeviceName

	// Instance store backed AMIs do not provide a root device name.
	if image.RootDeviceType == awstypes.DeviceTypeInstanceStore {
		return nil, nil
	}

	// Some AMIs have a RootDeviceName like "/dev/sda1" that does not appear as a
	// DeviceName in the BlockDeviceMapping list (which will instead have
	// something like "/dev/sda")
	//
	// While this seems like it breaks an invariant of AMIs, it ends up working
	// on the AWS side, and AMIs like this are common enough that we need to
	// special case it so Terraform does the right thing.
	//
	// Our heuristic is: if the RootDeviceName does not appear in the
	// BlockDeviceMapping, assume that the DeviceName of the first
	// BlockDeviceMapping entry serves as the root device.
	rootDeviceNameInMapping := false
	for _, bdm := range image.BlockDeviceMappings {
		if aws.ToString(bdm.DeviceName) == aws.ToString(image.RootDeviceName) {
			rootDeviceNameInMapping = true
		}
	}

	if !rootDeviceNameInMapping && len(image.BlockDeviceMappings) > 0 {
		rootDeviceName = image.BlockDeviceMappings[0].DeviceName
	}

	if rootDeviceName == nil {
		return nil, fmt.Errorf("finding Root Device Name for AMI (%s)", amiID)
	}

	return rootDeviceName, nil
}

func buildNetworkInterfaceOpts(d *schema.ResourceData, groups []string, nInterfaces interface{}) []awstypes.InstanceNetworkInterfaceSpecification {
	networkInterfaces := []awstypes.InstanceNetworkInterfaceSpecification{}
	// Get necessary items
	subnet, hasSubnet := d.GetOk(names.AttrSubnetID)

	if hasSubnet {
		// If we have a non-default VPC / Subnet specified, we can flag
		// AssociatePublicIpAddress to get a Public IP assigned. By default these are not provided.
		// You cannot specify both SubnetId and the NetworkInterface.0.* parameters though, otherwise
		// you get: Network interfaces and an instance-level subnet ID may not be specified on the same request
		// You also need to attach Security Groups to the NetworkInterface instead of the instance,
		// to avoid: Network interfaces and an instance-level security groups may not be specified on
		// the same request
		ni := awstypes.InstanceNetworkInterfaceSpecification{
			DeviceIndex: aws.Int32(0),
			SubnetId:    aws.String(subnet.(string)),
			Groups:      groups,
		}

		if v, ok := d.GetOkExists("associate_public_ip_address"); ok {
			ni.AssociatePublicIpAddress = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("private_ip"); ok {
			ni.PrivateIpAddress = aws.String(v.(string))
		}

		if v, ok := d.GetOk("secondary_private_ips"); ok && v.(*schema.Set).Len() > 0 {
			ni.PrivateIpAddresses = expandSecondaryPrivateIPAddresses(v.(*schema.Set).List())
		}

		if v, ok := d.GetOk("ipv6_address_count"); ok {
			ni.Ipv6AddressCount = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("ipv6_addresses"); ok {
			ipv6Addresses := make([]awstypes.InstanceIpv6Address, len(v.([]interface{})))
			for i, address := range v.([]interface{}) {
				ipv6Addresses[i] = awstypes.InstanceIpv6Address{
					Ipv6Address: aws.String(address.(string)),
				}
			}

			ni.Ipv6Addresses = ipv6Addresses
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			for _, v := range v.List() {
				ni.Groups = append(ni.Groups, v.(string))
			}
		}

		networkInterfaces = append(networkInterfaces, ni)
	} else {
		// If we have manually specified network interfaces, build and attach those here.
		vL := nInterfaces.(*schema.Set).List()
		for _, v := range vL {
			ini := v.(map[string]interface{})
			ni := awstypes.InstanceNetworkInterfaceSpecification{
				DeviceIndex:         aws.Int32(int32(ini["device_index"].(int))),
				NetworkCardIndex:    aws.Int32(int32(ini["network_card_index"].(int))),
				NetworkInterfaceId:  aws.String(ini[names.AttrNetworkInterfaceID].(string)),
				DeleteOnTermination: aws.Bool(ini[names.AttrDeleteOnTermination].(bool)),
			}
			networkInterfaces = append(networkInterfaces, ni)
		}
	}

	return networkInterfaces
}

func readBlockDeviceMappingsFromConfig(ctx context.Context, d *schema.ResourceData, conn *ec2.Client) ([]awstypes.BlockDeviceMapping, error) {
	blockDevices := make([]awstypes.BlockDeviceMapping, 0)

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrSnapshotID].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd[names.AttrEncrypted].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd[names.AttrKMSKeyID].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
				if iops, ok := bd[names.AttrIOPS].(int); ok && iops > 0 {
					if awstypes.VolumeTypeIo1 == ebs.VolumeType || awstypes.VolumeTypeIo2 == ebs.VolumeType || awstypes.VolumeTypeGp3 == ebs.VolumeType {
						// Condition: This parameter is required for requests to create io1 or io2
						// volumes and optional for gp3; it is not used in requests to create gp2, st1, sc1, or
						// standard volumes.
						// See: http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html
						ebs.Iops = aws.Int32(int32(iops))
					} else {
						// Enforce IOPs usage with a valid volume type
						// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
						return nil, fmt.Errorf("creating resource: iops attribute not supported for ebs_block_device with volume_type %s", v)
					}
				}
				if throughput, ok := bd[names.AttrThroughput].(int); ok && throughput > 0 {
					// `throughput` is only valid for gp3
					if awstypes.VolumeTypeGp3 == ebs.VolumeType {
						ebs.Throughput = aws.Int32(int32(throughput))
					} else {
						return nil, fmt.Errorf("creating resource: throughput attribute not supported for ebs_block_device with volume_type %s", v)
					}
				}
			}

			blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
				DeviceName: aws.String(bd[names.AttrDeviceName].(string)),
				Ebs:        ebs,
			})
		}
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			bdm := awstypes.BlockDeviceMapping{
				DeviceName:  aws.String(bd[names.AttrDeviceName].(string)),
				VirtualName: aws.String(bd[names.AttrVirtualName].(string)),
			}
			if v, ok := bd["no_device"].(bool); ok && v {
				bdm.NoDevice = aws.String("")
				// When NoDevice is true, just ignore VirtualName since it's not needed
				bdm.VirtualName = nil
			}

			if bdm.NoDevice == nil && aws.ToString(bdm.VirtualName) == "" {
				return nil, errors.New("virtual_name cannot be empty when no_device is false or undefined.")
			}

			blockDevices = append(blockDevices, bdm)
		}
	}

	if v, ok := d.GetOk("root_block_device"); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrEncrypted].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd[names.AttrKMSKeyID].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(bd[names.AttrKMSKeyID].(string))
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
				if iops, ok := bd[names.AttrIOPS].(int); ok && iops > 0 {
					if awstypes.VolumeTypeIo1 == ebs.VolumeType || awstypes.VolumeTypeIo2 == ebs.VolumeType || awstypes.VolumeTypeGp3 == ebs.VolumeType {
						// Only set the iops attribute if the volume type is io1, io2, or gp3. Setting otherwise
						// can trigger a refresh/plan loop based on the computed value that is given
						// from AWS, and prevent us from specifying 0 as a valid iops.
						//   See https://github.com/hashicorp/terraform/pull/4146
						//   See https://github.com/hashicorp/terraform/issues/7765
						ebs.Iops = aws.Int32(int32(iops))
					} else {
						// Enforce IOPs usage with a valid volume type
						// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
						return nil, fmt.Errorf("creating resource: iops attribute not supported for root_block_device with volume_type %s", v)
					}
				}
				if throughput, ok := bd[names.AttrThroughput].(int); ok && throughput > 0 {
					// throughput is only valid for gp3
					if awstypes.VolumeTypeGp3 == ebs.VolumeType {
						ebs.Throughput = aws.Int32(int32(throughput))
					} else {
						// Enforce throughput usage with a valid volume type
						return nil, fmt.Errorf("creating resource: throughput attribute not supported for root_block_device with volume_type %s", v)
					}
				}
			}

			var amiID string

			if v, ok := d.GetOk(names.AttrLaunchTemplate); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				launchTemplateData, err := findLaunchTemplateData(ctx, conn, expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{})))

				if err != nil {
					return nil, err
				}

				amiID = aws.ToString(launchTemplateData.ImageId)
			}

			// AMI from configuration overrides the one from the launch template.
			if v, ok := d.GetOk("ami"); ok {
				amiID = v.(string)
			}

			if amiID == "" {
				return nil, errors.New("`ami` must be set or provided via `launch_template`")
			}

			if dn, err := findRootDeviceName(ctx, conn, amiID); err == nil {
				if dn == nil {
					return nil, fmt.Errorf(
						"Expected 1 AMI for ID: %s, got none",
						amiID)
				}

				blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
					DeviceName: dn,
					Ebs:        ebs,
				})
			} else {
				return nil, err
			}
		}
	}

	return blockDevices, nil
}

func readVolumeTags(ctx context.Context, conn *ec2.Client, instanceId string) ([]awstypes.Tag, error) {
	volIDs, err := getInstanceVolIDs(ctx, conn, instanceId)
	if err != nil {
		return nil, fmt.Errorf("getting tags for volumes (%s): %s", volIDs, err)
	}

	resp, err := conn.DescribeTags(ctx, &ec2.DescribeTagsInput{
		Filters: attributeFiltersFromMultimap(map[string][]string{
			"resource-id": volIDs,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("getting tags for volumes (%s): %s", volIDs, err)
	}

	return tagsFromTagDescriptions(resp.Tags), nil
}

// Determine whether we're referring to security groups with IDs or names. We
// use a heuristic to figure this out. The default VPC can have security groups
// with IDs or names, so store them both here and let the config determine
// which one to use in Plan and Apply.
func readSecurityGroups(ctx context.Context, d *schema.ResourceData, instance *awstypes.Instance, conn *ec2.Client) error {
	// An instance with a subnet is in a VPC, and possibly the default VPC.
	// An instance without a subnet is in the default VPC.
	hasSubnet := aws.ToString(instance.SubnetId) != ""
	useID, useName := hasSubnet, !hasSubnet

	// If the instance is in a VPC, find out if that VPC is Default to determine
	// whether to store names.
	if vpcID := aws.ToString(instance.VpcId); vpcID != "" {
		vpc, err := findVPCByID(ctx, conn, vpcID)

		if err != nil {
			log.Printf("[WARN] error reading EC2 Instance (%s) VPC (%s): %s", d.Id(), vpcID, err)
		} else {
			useName = aws.ToBool(vpc.IsDefault)
		}
	}

	// Build up the security groups
	if useID {
		sgs := make([]string, 0, len(instance.SecurityGroups))
		for _, sg := range instance.SecurityGroups {
			sgs = append(sgs, aws.ToString(sg.GroupId))
		}
		log.Printf("[DEBUG] Setting Security Group IDs: %#v", sgs)
		if err := d.Set(names.AttrVPCSecurityGroupIDs, sgs); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	} else {
		if err := d.Set(names.AttrVPCSecurityGroupIDs, []string{}); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	if useName {
		sgs := make([]string, 0, len(instance.SecurityGroups))
		for _, sg := range instance.SecurityGroups {
			sgs = append(sgs, aws.ToString(sg.GroupName))
		}
		log.Printf("[DEBUG] Setting Security Group Names: %#v", sgs)
		if err := d.Set(names.AttrSecurityGroups, sgs); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	} else {
		if err := d.Set(names.AttrSecurityGroups, []string{}); err != nil {
			return err // nosemgrep:ci.bare-error-returns
		}
	}
	return nil
}

func readInstanceShutdownBehavior(ctx context.Context, d *schema.ResourceData, conn *ec2.Client) error {
	output, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
		InstanceId: aws.String(d.Id()),
		Attribute:  awstypes.InstanceAttributeNameInstanceInitiatedShutdownBehavior,
	})

	if err != nil {
		return fmt.Errorf("getting attribute (%s): %w", awstypes.InstanceAttributeNameInstanceInitiatedShutdownBehavior, err)
	}

	if output != nil && output.InstanceInitiatedShutdownBehavior != nil {
		d.Set("instance_initiated_shutdown_behavior", output.InstanceInitiatedShutdownBehavior.Value)
	}

	return nil
}

func getInstancePasswordData(ctx context.Context, instanceID string, conn *ec2.Client, timeout time.Duration) (string, error) {
	log.Printf("[INFO] Reading password data for instance %s", instanceID)

	var passwordData string
	var resp *ec2.GetPasswordDataOutput
	input := &ec2.GetPasswordDataInput{
		InstanceId: aws.String(instanceID),
	}
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error
		resp, err = conn.GetPasswordData(ctx, input)

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if resp.PasswordData == nil || aws.ToString(resp.PasswordData) == "" {
			return retry.RetryableError(fmt.Errorf("Password data is blank for instance ID: %s", instanceID))
		}

		passwordData = strings.TrimSpace(aws.ToString(resp.PasswordData))

		log.Printf("[INFO] Password data read for instance %s", instanceID)
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.GetPasswordData(ctx, input)
		if err != nil {
			return "", fmt.Errorf("getting password data: %s", err)
		}
		if resp.PasswordData == nil || aws.ToString(resp.PasswordData) == "" {
			return "", errors.New("password data is blank")
		}
		passwordData = strings.TrimSpace(aws.ToString(resp.PasswordData))
	}
	if err != nil {
		return "", err
	}

	return passwordData, nil
}

type instanceOpts struct {
	BlockDeviceMappings               []awstypes.BlockDeviceMapping
	CapacityReservationSpecification  *awstypes.CapacityReservationSpecification
	CpuOptions                        *awstypes.CpuOptionsRequest
	CreditSpecification               *awstypes.CreditSpecificationRequest
	DisableAPIStop                    *bool
	DisableAPITermination             *bool
	EBSOptimized                      *bool
	EnclaveOptions                    *awstypes.EnclaveOptionsRequest
	HibernationOptions                *awstypes.HibernationOptionsRequest
	IAMInstanceProfile                *awstypes.IamInstanceProfileSpecification
	ImageID                           *string
	InstanceInitiatedShutdownBehavior awstypes.ShutdownBehavior
	InstanceMarketOptions             *awstypes.InstanceMarketOptionsRequest
	InstanceType                      awstypes.InstanceType
	Ipv6AddressCount                  *int32
	Ipv6Addresses                     []awstypes.InstanceIpv6Address
	KeyName                           *string
	LaunchTemplate                    *awstypes.LaunchTemplateSpecification
	MaintenanceOptions                *awstypes.InstanceMaintenanceOptionsRequest
	MetadataOptions                   *awstypes.InstanceMetadataOptionsRequest
	Monitoring                        *awstypes.RunInstancesMonitoringEnabled
	NetworkInterfaces                 []awstypes.InstanceNetworkInterfaceSpecification
	Placement                         *awstypes.Placement
	PrivateDNSNameOptions             *awstypes.PrivateDnsNameOptionsRequest
	PrivateIPAddress                  *string
	SecurityGroupIDs                  []string
	SecurityGroups                    []string
	SpotPlacement                     *awstypes.SpotPlacement
	SubnetID                          *string
	UserData64                        *string
}

func buildInstanceOpts(ctx context.Context, d *schema.ResourceData, meta interface{}) (*instanceOpts, error) {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	opts := &instanceOpts{
		DisableAPITermination: aws.Bool(d.Get("disable_api_termination").(bool)),
		EBSOptimized:          aws.Bool(d.Get("ebs_optimized").(bool)),
		EnclaveOptions:        expandEnclaveOptions(d.Get("enclave_options").([]interface{})),
		MetadataOptions:       expandInstanceMetadataOptions(d.Get("metadata_options").([]interface{})),
	}

	if v, ok := d.GetOk("disable_api_stop"); ok {
		opts.DisableAPIStop = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ami"); ok {
		opts.ImageID = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		opts.InstanceType = awstypes.InstanceType(v.(string))
	}

	var instanceInterruptionBehavior string

	if v, ok := d.GetOk(names.AttrLaunchTemplate); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		launchTemplateSpecification := expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}))
		launchTemplateData, err := findLaunchTemplateData(ctx, conn, launchTemplateSpecification)

		if err != nil {
			return nil, err
		}

		opts.LaunchTemplate = launchTemplateSpecification

		if launchTemplateData.InstanceMarketOptions != nil && launchTemplateData.InstanceMarketOptions.SpotOptions != nil {
			instanceInterruptionBehavior = string(launchTemplateData.InstanceMarketOptions.SpotOptions.InstanceInterruptionBehavior)
		}
	}

	instanceType := d.Get(names.AttrInstanceType).(string)

	// Set default cpu_credits as Unlimited for T3/T3a instance type
	if strings.HasPrefix(instanceType, "t3") {
		opts.CreditSpecification = &awstypes.CreditSpecificationRequest{
			CpuCredits: aws.String(cpuCreditsUnlimited),
		}
	}

	if v, ok := d.GetOk("credit_specification"); ok && len(v.([]interface{})) > 0 {
		if instanceType != "" {
			instanceTypeInfo, err := findInstanceTypeByName(ctx, conn, instanceType)

			if err != nil {
				return nil, fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
			}

			if aws.ToBool(instanceTypeInfo.BurstablePerformanceSupported) {
				if v, ok := v.([]interface{})[0].(map[string]interface{}); ok {
					opts.CreditSpecification = expandCreditSpecificationRequest(v)
				} else {
					log.Print("[WARN] credit_specification is defined but the value of cpu_credits is missing, default value will be used.")
				}
			} else {
				log.Print("[WARN] credit_specification is defined but instance type does not support burstable performance. Ignoring...")
			}
		}
	}

	if v := d.Get("instance_initiated_shutdown_behavior").(string); v != "" {
		opts.InstanceInitiatedShutdownBehavior = awstypes.ShutdownBehavior(v)
	}

	opts.Monitoring = &awstypes.RunInstancesMonitoringEnabled{
		Enabled: aws.Bool(d.Get("monitoring").(bool)),
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok {
		opts.IAMInstanceProfile = &awstypes.IamInstanceProfileSpecification{
			Name: aws.String(v.(string)),
		}
	}

	userData := d.Get("user_data").(string)
	userDataBase64 := d.Get("user_data_base64").(string)

	if userData != "" {
		opts.UserData64 = flex.StringValueToBase64String(userData)
	} else if userDataBase64 != "" {
		opts.UserData64 = aws.String(userDataBase64)
	}

	// check for non-default Subnet, and cast it to a String
	subnet, hasSubnet := d.GetOk(names.AttrSubnetID)
	subnetID := subnet.(string)

	// Placement is used for aws_instance; SpotPlacement is used for
	// aws_spot_instance_request. They represent the same data. :-|
	opts.Placement = &awstypes.Placement{
		AvailabilityZone: aws.String(d.Get(names.AttrAvailabilityZone).(string)),
	}

	if v, ok := d.GetOk("placement_partition_number"); ok {
		opts.Placement.PartitionNumber = aws.Int32(int32(v.(int)))
	}

	opts.SpotPlacement = &awstypes.SpotPlacement{
		AvailabilityZone: aws.String(d.Get(names.AttrAvailabilityZone).(string)),
	}

	if v, ok := d.GetOk("placement_group"); ok && (instanceInterruptionBehavior == "" || instanceInterruptionBehavior == string(awstypes.InstanceInterruptionBehaviorTerminate)) {
		opts.Placement.GroupName = aws.String(v.(string))
		opts.SpotPlacement.GroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		opts.Placement.Tenancy = awstypes.Tenancy(v.(string))
	}
	if v, ok := d.GetOk("host_id"); ok {
		opts.Placement.HostId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_resource_group_arn"); ok {
		opts.Placement.HostResourceGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cpu_options"); ok {
		opts.CpuOptions = expandCPUOptions(v.([]interface{}))
	} else if v := d.Get("cpu_core_count").(int); v > 0 {
		// preserved to maintain backward compatibility
		tc := d.Get("cpu_threads_per_core").(int)
		if tc < 0 {
			tc = 2
		}
		opts.CpuOptions = &awstypes.CpuOptionsRequest{
			CoreCount:      aws.Int32(int32(v)),
			ThreadsPerCore: aws.Int32(int32(tc)),
		}
	}

	if v := d.Get("hibernation"); v != "" {
		opts.HibernationOptions = &awstypes.HibernationOptionsRequest{
			Configured: aws.Bool(v.(bool)),
		}
	}

	var groups []string
	if v := d.Get(names.AttrSecurityGroups); v != nil {
		// Security group names.
		// For a nondefault VPC, you must use security group IDs instead.
		// See http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RunInstances.html
		sgs := v.(*schema.Set).List()
		if len(sgs) > 0 && hasSubnet {
			log.Print("[WARN] Deprecated. Attempting to use 'security_groups' within a VPC instance. Use 'vpc_security_group_ids' instead.")
		}
		for _, v := range sgs {
			str := v.(string)
			groups = append(groups, str)
		}
	}

	_, assocPubIPA := d.GetOkExists("associate_public_ip_address")
	_, privIP := d.GetOk("private_ip")
	_, secPrivIP := d.GetOk("secondary_private_ips")
	networkInterfaces, interfacesOk := d.GetOk("network_interface")

	// If setting subnet and public address, OR manual network interfaces, populate those now.
	if (hasSubnet && (assocPubIPA || privIP || secPrivIP)) || interfacesOk {
		// Otherwise we're attaching (a) network interface(s)
		opts.NetworkInterfaces = buildNetworkInterfaceOpts(d, groups, networkInterfaces)
	} else {
		// If simply specifying a subnetID, privateIP, Security Groups, or VPC Security Groups, build these now
		if subnetID != "" {
			opts.SubnetID = aws.String(subnetID)
		}

		if v, ok := d.GetOk("private_ip"); ok {
			opts.PrivateIPAddress = aws.String(v.(string))
		}
		if opts.SubnetID != nil &&
			aws.ToString(opts.SubnetID) != "" {
			opts.SecurityGroupIDs = groups
		} else {
			opts.SecurityGroups = groups
		}

		if v, ok := d.GetOk("ipv6_address_count"); ok {
			opts.Ipv6AddressCount = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("ipv6_addresses"); ok {
			ipv6Addresses := make([]awstypes.InstanceIpv6Address, len(v.([]interface{})))
			for i, address := range v.([]interface{}) {
				ipv6Addresses[i] = awstypes.InstanceIpv6Address{
					Ipv6Address: aws.String(address.(string)),
				}
			}

			opts.Ipv6Addresses = ipv6Addresses
		}

		if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
			for _, v := range v.List() {
				opts.SecurityGroupIDs = append(opts.SecurityGroupIDs, v.(string))
			}
		}
	}

	if v, ok := d.GetOk("key_name"); ok {
		opts.KeyName = aws.String(v.(string))
	}

	blockDevices, err := readBlockDeviceMappingsFromConfig(ctx, d, conn)
	if err != nil {
		return nil, err
	}
	if len(blockDevices) > 0 {
		opts.BlockDeviceMappings = blockDevices
	}

	if v, ok := d.GetOk("capacity_reservation_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		opts.CapacityReservationSpecification = expandCapacityReservationSpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("maintenance_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		opts.MaintenanceOptions = expandInstanceMaintenanceOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("private_dns_name_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		opts.PrivateDNSNameOptions = expandPrivateDNSNameOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("instance_market_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		opts.InstanceMarketOptions = expandInstanceMarketOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	return opts, nil
}

// startInstance starts an EC2 instance and waits for the instance to start.
func startInstance(ctx context.Context, conn *ec2.Client, id string, retry bool, timeout time.Duration) error {
	var err error

	tflog.Info(ctx, "Starting EC2 Instance", map[string]any{
		"ec2_instance_id": id,
	})
	if retry {
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16433.
		_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, ec2PropagationTimeout,
			func() (interface{}, error) {
				return conn.StartInstances(ctx, &ec2.StartInstancesInput{
					InstanceIds: []string{id},
				})
			},
			errCodeInvalidParameterValue, "LaunchPlan instance type does not match attribute value",
		)
	} else {
		_, err = conn.StartInstances(ctx, &ec2.StartInstancesInput{
			InstanceIds: []string{id},
		})
	}

	if err != nil {
		return fmt.Errorf("starting EC2 Instance (%s): %w", id, err)
	}

	if _, err := waitInstanceStarted(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Instance (%s) start: %w", id, err)
	}

	return nil
}

// stopInstance stops an EC2 instance and waits for the instance to stop.
func stopInstance(ctx context.Context, conn *ec2.Client, id string, force bool, timeout time.Duration) error {
	tflog.Info(ctx, "Stopping EC2 Instance", map[string]any{
		"ec2_instance_id": id,
		"force":           force,
	})
	_, err := conn.StopInstances(ctx, &ec2.StopInstancesInput{
		Force:       aws.Bool(force),
		InstanceIds: []string{id},
	})

	if err != nil {
		return fmt.Errorf("stopping EC2 Instance (%s): %w", id, err)
	}

	if _, err := waitInstanceStopped(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Instance (%s) stop: %w", id, err)
	}

	return nil
}

// terminateInstance shuts down an EC2 instance and waits for the instance to be deleted.
func terminateInstance(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Terminating EC2 Instance: %s", id)
	_, err := conn.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{id},
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("terminating EC2 Instance (%s): %w", id, err)
	}

	if _, err := waitInstanceDeleted(ctx, conn, id, timeout); err != nil {
		return fmt.Errorf("waiting for EC2 Instance (%s) delete: %w", id, err)
	}

	return nil
}

func waitInstanceCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceStateNamePending),
		Target:     enum.Slice(awstypes.InstanceStateNameRunning),
		Refresh:    statusInstance(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitInstanceDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.InstanceStateNamePending,
			awstypes.InstanceStateNameRunning,
			awstypes.InstanceStateNameShuttingDown,
			awstypes.InstanceStateNameStopping,
			awstypes.InstanceStateNameStopped,
		),
		Target:     enum.Slice(awstypes.InstanceStateNameTerminated),
		Refresh:    statusInstance(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitInstanceReady(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceStateNamePending, awstypes.InstanceStateNameStopping),
		Target:     enum.Slice(awstypes.InstanceStateNameRunning, awstypes.InstanceStateNameStopped),
		Refresh:    statusInstance(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitInstanceStarted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceStateNamePending, awstypes.InstanceStateNameStopped),
		Target:     enum.Slice(awstypes.InstanceStateNameRunning),
		Refresh:    statusInstance(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitInstanceStopped(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.InstanceStateNamePending,
			awstypes.InstanceStateNameRunning,
			awstypes.InstanceStateNameShuttingDown,
			awstypes.InstanceStateNameStopping,
		),
		Target:     enum.Slice(awstypes.InstanceStateNameStopped),
		Refresh:    statusInstance(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func userDataHashSum(userData string) string {
	// Check whether the user_data is not Base64 encoded.
	// Always calculate hash of base64 decoded value since we
	// check against double-encoding when setting it.
	v, err := itypes.Base64Decode(userData)
	if err != nil {
		v = []byte(userData)
	}

	hash := sha1.Sum(v)
	return hex.EncodeToString(hash[:])
}

func getInstanceVolIDs(ctx context.Context, conn *ec2.Client, instanceId string) ([]string, error) {
	volIDs := []string{}

	resp, err := conn.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.instance-id": instanceId,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("getting volumes: %s", err)
	}

	for _, v := range resp.Volumes {
		volIDs = append(volIDs, aws.ToString(v.VolumeId))
	}

	return volIDs, nil
}

func getRootVolID(instance *awstypes.Instance) string {
	volID := ""
	for _, bd := range instance.BlockDeviceMappings {
		if bd.Ebs != nil && blockDeviceIsRoot(bd, instance) {
			if bd.Ebs.VolumeId != nil {
				volID = aws.ToString(bd.Ebs.VolumeId)
			}
			break
		}
	}

	return volID
}

func getVolIDByDeviceName(instance *awstypes.Instance, deviceName string) string {
	volID := ""
	for _, bd := range instance.BlockDeviceMappings {
		if aws.ToString(bd.DeviceName) == deviceName {
			if bd.Ebs != nil {
				volID = aws.ToString(bd.Ebs.VolumeId)
				break
			}
		}
	}

	return volID
}

func blockDeviceTagsDefined(d *schema.ResourceData) bool {
	if v, ok := d.GetOk("root_block_device"); ok {
		vL := v.([]interface{})
		for _, v := range vL {
			bd := v.(map[string]interface{})
			if blockDeviceTags, ok := bd[names.AttrTags].(map[string]interface{}); ok && len(blockDeviceTags) > 0 {
				return true
			}
		}
	}

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			if blockDeviceTags, ok := bd[names.AttrTags].(map[string]interface{}); ok && len(blockDeviceTags) > 0 {
				return true
			}
		}
	}

	return false
}

func expandInstanceMetadataOptions(l []interface{}) *awstypes.InstanceMetadataOptionsRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	opts := &awstypes.InstanceMetadataOptionsRequest{
		HttpEndpoint: awstypes.InstanceMetadataEndpointState(m["http_endpoint"].(string)),
	}

	if m["http_endpoint"].(string) == string(awstypes.InstanceMetadataEndpointStateEnabled) {
		// These parameters are not allowed unless HttpEndpoint is enabled
		if v, ok := m["http_protocol_ipv6"].(string); ok && v != "" {
			opts.HttpProtocolIpv6 = awstypes.InstanceMetadataProtocolState(v)
		}

		if v, ok := m["http_tokens"].(string); ok && v != "" {
			opts.HttpTokens = awstypes.HttpTokensState(v)
		}

		if v, ok := m["http_put_response_hop_limit"].(int); ok && v != 0 {
			opts.HttpPutResponseHopLimit = aws.Int32(int32(v))
		}

		if v, ok := m["instance_metadata_tags"].(string); ok && v != "" {
			opts.InstanceMetadataTags = awstypes.InstanceMetadataTagsState(v)
		}
	}

	return opts
}

func expandCPUOptions(l []interface{}) *awstypes.CpuOptionsRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	opts := &awstypes.CpuOptionsRequest{}

	if v, ok := m["amd_sev_snp"].(string); ok && v != "" {
		opts.AmdSevSnp = awstypes.AmdSevSnpSpecification(v)
	}

	if v, ok := m["core_count"].(int); ok && v > 0 {
		tc := m["threads_per_core"].(int)
		if tc < 0 {
			tc = 2
		}
		opts.CoreCount = aws.Int32(int32(v))
		opts.ThreadsPerCore = aws.Int32(int32(tc))
	}

	return opts
}

func expandEnclaveOptions(l []interface{}) *awstypes.EnclaveOptionsRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	opts := &awstypes.EnclaveOptionsRequest{
		Enabled: aws.Bool(m[names.AttrEnabled].(bool)),
	}

	return opts
}

// Expands an array of secondary Private IPs into a ec2 Private IP Address Spec
func expandSecondaryPrivateIPAddresses(ips []interface{}) []awstypes.PrivateIpAddressSpecification {
	specs := make([]awstypes.PrivateIpAddressSpecification, 0, len(ips))
	for _, v := range ips {
		spec := awstypes.PrivateIpAddressSpecification{
			PrivateIpAddress: aws.String(v.(string)),
		}

		specs = append(specs, spec)
	}
	return specs
}

func flattenInstanceMetadataOptions(opts *awstypes.InstanceMetadataOptionsResponse) []interface{} {
	if opts == nil {
		return nil
	}

	m := map[string]interface{}{
		"http_endpoint":               opts.HttpEndpoint,
		"http_protocol_ipv6":          opts.HttpProtocolIpv6,
		"http_put_response_hop_limit": opts.HttpPutResponseHopLimit,
		"http_tokens":                 opts.HttpTokens,
		"instance_metadata_tags":      opts.InstanceMetadataTags,
	}

	return []interface{}{m}
}

func flattenCPUOptions(opts *awstypes.CpuOptions) []interface{} {
	if opts == nil {
		return nil
	}

	m := map[string]interface{}{
		"amd_sev_snp": opts.AmdSevSnp,
	}

	if v := opts.CoreCount; v != nil {
		m["core_count"] = aws.ToInt32(v)
	}

	if v := opts.ThreadsPerCore; v != nil {
		m["threads_per_core"] = aws.ToInt32(v)
	}

	return []interface{}{m}
}

func flattenEnclaveOptions(opts *awstypes.EnclaveOptions) []interface{} {
	if opts == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.ToBool(opts.Enabled),
	}

	return []interface{}{m}
}

func expandCapacityReservationSpecification(tfMap map[string]interface{}) *awstypes.CapacityReservationSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CapacityReservationSpecification{}

	if v, ok := tfMap["capacity_reservation_preference"].(string); ok && v != "" {
		apiObject.CapacityReservationPreference = awstypes.CapacityReservationPreference(v)
	}

	if v, ok := tfMap["capacity_reservation_target"].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityReservationTarget = expandCapacityReservationTarget(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCapacityReservationTarget(tfMap map[string]interface{}) *awstypes.CapacityReservationTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CapacityReservationTarget{}

	if v, ok := tfMap["capacity_reservation_id"].(string); ok && v != "" {
		apiObject.CapacityReservationId = aws.String(v)
	}

	if v, ok := tfMap["capacity_reservation_resource_group_arn"].(string); ok && v != "" {
		apiObject.CapacityReservationResourceGroupArn = aws.String(v)
	}

	return apiObject
}

func flattenCapacityReservationSpecificationResponse(apiObject *awstypes.CapacityReservationSpecificationResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"capacity_reservation_preference": apiObject.CapacityReservationPreference,
	}

	if v := apiObject.CapacityReservationTarget; v != nil {
		tfMap["capacity_reservation_target"] = []interface{}{flattenCapacityReservationTargetResponse(v)}
	}

	return tfMap
}

func flattenCapacityReservationTargetResponse(apiObject *awstypes.CapacityReservationTargetResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CapacityReservationId; v != nil {
		tfMap["capacity_reservation_id"] = aws.ToString(v)
	}

	if v := apiObject.CapacityReservationResourceGroupArn; v != nil {
		tfMap["capacity_reservation_resource_group_arn"] = aws.ToString(v)
	}

	return tfMap
}

func capacityReservationSpecificationResponsesEqual(v1 *awstypes.CapacityReservationSpecificationResponse, v2 *awstypes.CapacityReservationSpecification) bool {
	if v1 == nil {
		return v2 == nil
	}

	if v2 == nil {
		return false
	}

	if v1.CapacityReservationPreference != v2.CapacityReservationPreference {
		return false
	}

	if !capacityReservationTargetResponsesEqual(v1.CapacityReservationTarget, v2.CapacityReservationTarget) {
		return false
	}

	return true
}

func capacityReservationTargetResponsesEqual(v1 *awstypes.CapacityReservationTargetResponse, v2 *awstypes.CapacityReservationTarget) bool {
	if v1 == nil {
		return v2 == nil
	}

	if v2 == nil {
		return false
	}

	if aws.ToString(v1.CapacityReservationId) != aws.ToString(v2.CapacityReservationId) {
		return false
	}

	if aws.ToString(v1.CapacityReservationResourceGroupArn) != aws.ToString(v2.CapacityReservationResourceGroupArn) {
		return false
	}

	return true
}

func expandCreditSpecificationRequest(tfMap map[string]interface{}) *awstypes.CreditSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CreditSpecificationRequest{}

	if v, ok := tfMap["cpu_credits"].(string); ok && v != "" {
		apiObject.CpuCredits = aws.String(v)
	}

	return apiObject
}

func expandInstanceCreditSpecificationRequest(tfMap map[string]interface{}) awstypes.InstanceCreditSpecificationRequest {
	apiObject := awstypes.InstanceCreditSpecificationRequest{}

	if v, ok := tfMap["cpu_credits"].(string); ok && v != "" {
		apiObject.CpuCredits = aws.String(v)
	}

	return apiObject
}

func flattenInstanceCreditSpecification(apiObject *awstypes.InstanceCreditSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CpuCredits; v != nil {
		tfMap["cpu_credits"] = aws.ToString(v)
	}

	return tfMap
}

func expandInstanceMaintenanceOptionsRequest(tfMap map[string]interface{}) *awstypes.InstanceMaintenanceOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceMaintenanceOptionsRequest{}

	if v, ok := tfMap["auto_recovery"].(string); ok && v != "" {
		apiObject.AutoRecovery = awstypes.InstanceAutoRecoveryState(v)
	}

	return apiObject
}

func flattenInstanceMaintenanceOptions(apiObject *awstypes.InstanceMaintenanceOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"auto_recovery": apiObject.AutoRecovery,
	}

	return tfMap
}

func expandPrivateDNSNameOptionsRequest(tfMap map[string]interface{}) *awstypes.PrivateDnsNameOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PrivateDnsNameOptionsRequest{}

	if v, ok := tfMap["enable_resource_name_dns_aaaa_record"].(bool); ok {
		apiObject.EnableResourceNameDnsAAAARecord = aws.Bool(v)
	}

	if v, ok := tfMap["enable_resource_name_dns_a_record"].(bool); ok {
		apiObject.EnableResourceNameDnsARecord = aws.Bool(v)
	}

	if v, ok := tfMap["hostname_type"].(string); ok && v != "" {
		apiObject.HostnameType = awstypes.HostnameType(v)
	}

	return apiObject
}

func flattenPrivateDNSNameOptionsResponse(apiObject *awstypes.PrivateDnsNameOptionsResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"hostname_type": apiObject.HostnameType,
	}

	if v := apiObject.EnableResourceNameDnsAAAARecord; v != nil {
		tfMap["enable_resource_name_dns_aaaa_record"] = v
	}

	if v := apiObject.EnableResourceNameDnsARecord; v != nil {
		tfMap["enable_resource_name_dns_a_record"] = v
	}

	return tfMap
}

func expandInstanceMarketOptionsRequest(tfMap map[string]interface{}) *awstypes.InstanceMarketOptionsRequest {
	apiObject := &awstypes.InstanceMarketOptionsRequest{}

	if v, ok := tfMap["market_type"]; ok && v.(string) != "" {
		apiObject.MarketType = awstypes.MarketType(tfMap["market_type"].(string))
	}

	if v, ok := tfMap["spot_options"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.SpotOptions = expandSpotMarketOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	return apiObject
}

func expandSpotMarketOptions(tfMap map[string]interface{}) *awstypes.SpotMarketOptions {
	apiObject := &awstypes.SpotMarketOptions{}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = awstypes.InstanceInterruptionBehavior(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_type"].(string); ok && v != "" {
		apiObject.SpotInstanceType = awstypes.SpotInstanceType(v)
	}

	if v, ok := tfMap["valid_until"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.ValidUntil = aws.Time(v)
	}

	return apiObject
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *awstypes.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpecification{}

	// DescribeLaunchTemplates returns both name and id but LaunchTemplateSpecification
	// allows only one of them to be set.
	if v, ok := tfMap[names.AttrID]; ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v.(string))
	} else if v, ok := tfMap[names.AttrName]; ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func flattenInstanceLaunchTemplate(ctx context.Context, conn *ec2.Client, instanceID, previousLaunchTemplateVersion string) ([]interface{}, error) {
	launchTemplateID, err := findInstanceLaunchTemplateID(ctx, conn, instanceID)

	if err != nil {
		return nil, err
	}

	if launchTemplateID == "" {
		return nil, nil
	}

	name, defaultVersion, latestVersion, err := findLaunchTemplateNameAndVersions(ctx, conn, launchTemplateID)

	if tfresource.NotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading EC2 Launch Template (%s): %w", launchTemplateID, err)
	}

	tfMap := map[string]interface{}{
		names.AttrID:   launchTemplateID,
		names.AttrName: name,
	}

	currentLaunchTemplateVersion, err := findInstanceLaunchTemplateVersion(ctx, conn, instanceID)

	if err != nil {
		return nil, err
	}

	_, err = findLaunchTemplateVersionByTwoPartKey(ctx, conn, launchTemplateID, currentLaunchTemplateVersion)

	if tfresource.NotFound(err) {
		return []interface{}{tfMap}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading EC2 Launch Template (%s) version (%s): %w", launchTemplateID, currentLaunchTemplateVersion, err)
	}

	switch previousLaunchTemplateVersion {
	case launchTemplateVersionDefault:
		if currentLaunchTemplateVersion == defaultVersion {
			tfMap[names.AttrVersion] = launchTemplateVersionDefault
		} else {
			tfMap[names.AttrVersion] = currentLaunchTemplateVersion
		}
	case launchTemplateVersionLatest:
		if currentLaunchTemplateVersion == latestVersion {
			tfMap[names.AttrVersion] = launchTemplateVersionLatest
		} else {
			tfMap[names.AttrVersion] = currentLaunchTemplateVersion
		}
	default:
		tfMap[names.AttrVersion] = currentLaunchTemplateVersion
	}

	return []interface{}{tfMap}, nil
}

func findInstanceLaunchTemplateID(ctx context.Context, conn *ec2.Client, id string) (string, error) {
	launchTemplateID, err := findInstanceTagValue(ctx, conn, id, "aws:ec2launchtemplate:id")

	if err != nil {
		return "", fmt.Errorf("reading EC2 Instance (%s) launch template ID tag: %w", id, err)
	}

	return launchTemplateID, nil
}

func findInstanceLaunchTemplateVersion(ctx context.Context, conn *ec2.Client, id string) (string, error) {
	launchTemplateVersion, err := findInstanceTagValue(ctx, conn, id, "aws:ec2launchtemplate:version")

	if err != nil {
		return "", fmt.Errorf("reading EC2 Instance (%s) launch template version tag: %w", id, err)
	}

	return launchTemplateVersion, nil
}

func findLaunchTemplateData(ctx context.Context, conn *ec2.Client, launchTemplateSpecification *awstypes.LaunchTemplateSpecification) (*awstypes.ResponseLaunchTemplateData, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{}

	if v := aws.ToString(launchTemplateSpecification.LaunchTemplateId); v != "" {
		input.LaunchTemplateId = aws.String(v)
	} else if v := aws.ToString(launchTemplateSpecification.LaunchTemplateName); v != "" {
		input.LaunchTemplateName = aws.String(v)
	}

	var latestVersion bool

	if v := aws.ToString(launchTemplateSpecification.Version); v != "" {
		switch v {
		case launchTemplateVersionDefault:
			input.Filters = newAttributeFilterList(map[string]string{
				"is-default-version": "true",
			})
		case launchTemplateVersionLatest:
			latestVersion = true
		default:
			input.Versions = []string{v}
		}
	}

	output, err := findLaunchTemplateVersions(ctx, conn, input)

	if err != nil {
		return nil, fmt.Errorf("reading EC2 Launch Template versions: %w", err)
	}

	if latestVersion {
		return output[len(output)-1].LaunchTemplateData, nil
	}

	return output[0].LaunchTemplateData, nil
}

// findLaunchTemplateNameAndVersions returns the specified launch template's name, default version and latest version.
func findLaunchTemplateNameAndVersions(ctx context.Context, conn *ec2.Client, id string) (string, string, string, error) {
	lt, err := findLaunchTemplateByID(ctx, conn, id)

	if err != nil {
		return "", "", "", err
	}

	return aws.ToString(lt.LaunchTemplateName), strconv.FormatInt(aws.ToInt64(lt.DefaultVersionNumber), 10), strconv.FormatInt(aws.ToInt64(lt.LatestVersionNumber), 10), nil
}

func findInstanceTagValue(ctx context.Context, conn *ec2.Client, instanceID, tagKey string) (string, error) {
	input := &ec2.DescribeTagsInput{
		Filters: newAttributeFilterList(map[string]string{
			"resource-id": instanceID,
			names.AttrKey: tagKey,
		}),
	}

	output, err := conn.DescribeTags(ctx, input)

	if err != nil {
		return "", err
	}

	switch count := len(output.Tags); count {
	case 0:
		return "", nil
	case 1:
		return aws.ToString(output.Tags[0].Value), nil
	default:
		return "", tfresource.NewTooManyResultsError(count, input)
	}
}

// isSnowballEdgeInstance returns whether or not the specified instance ID indicates an SBE instance.
func isSnowballEdgeInstance(id string) bool {
	return strings.Contains(id, "s.")
}

// instanceType describes an EC2 instance type.
type instanceType struct {
	// e.g. "m6i"
	Type string
	// e.g. "m"
	Family string
	// e.g. 6
	Generation int
	// e.g. "i"
	AdditionalCapabilities string
	// e.g. "9xlarge"
	Size string
}

func parseInstanceType(s string) (*instanceType, error) {
	matches := regexache.MustCompile(`(([[:alpha:]]+)([[:digit:]])+([[:alpha:]]*))\.([[:alnum:]]+)`).FindStringSubmatch(s)

	if matches == nil {
		return nil, fmt.Errorf("invalid EC2 Instance Type name: %s", s)
	}

	generation, err := strconv.Atoi(matches[3])

	if err != nil {
		return nil, err
	}

	return &instanceType{
		Type:                   matches[1],
		Family:                 matches[2],
		Generation:             generation,
		AdditionalCapabilities: matches[4],
		Size:                   matches[5],
	}, nil
}

func hasCommonElement(slice1 []awstypes.ArchitectureType, slice2 []awstypes.ArchitectureType) bool {
	for _, type1 := range slice1 {
		for _, type2 := range slice2 {
			if type1 == type2 {
				return true
			}
		}
	}
	return false
}
