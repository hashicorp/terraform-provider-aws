// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func instanceMigrateState(
	v int, is *terraform.InstanceState, meta any) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Instance State v0; migrating to v1")
		return migrateInstanceStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateInstanceStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// Delete old count
	delete(is.Attributes, "block_device.#")

	oldBds, err := readV0BlockDevices(is)
	if err != nil {
		return is, err
	}
	// seed count fields for new types
	is.Attributes["ebs_block_device.#"] = "0"
	is.Attributes["ephemeral_block_device.#"] = "0"
	// depending on if state was v0.3.7 or an earlier version, it might have
	// root_block_device defined already
	if _, ok := is.Attributes["root_block_device.#"]; !ok {
		is.Attributes["root_block_device.#"] = "0"
	}
	for _, oldBd := range oldBds {
		writeV1BlockDevice(is, oldBd)
	}
	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}

func readV0BlockDevices(is *terraform.InstanceState) (map[string]map[string]string, error) {
	oldBds := make(map[string]map[string]string)
	for k, v := range is.Attributes {
		if !strings.HasPrefix(k, "block_device.") {
			continue
		}
		path := strings.Split(k, ".")
		if len(path) != 3 {
			return oldBds, fmt.Errorf("Found unexpected block_device field: %#v", k)
		}
		hashcode, attribute := path[1], path[2]
		oldBd, ok := oldBds[hashcode]
		if !ok {
			oldBd = make(map[string]string)
			oldBds[hashcode] = oldBd
		}
		oldBd[attribute] = v
		delete(is.Attributes, k)
	}
	return oldBds, nil
}

func writeV1BlockDevice(is *terraform.InstanceState, oldBd map[string]string) {
	code := create.StringHashcode(oldBd[names.AttrDeviceName])
	bdType := "ebs_block_device"
	if vn, ok := oldBd[names.AttrVirtualName]; ok && strings.HasPrefix(vn, "ephemeral") {
		bdType = "ephemeral_block_device"
	} else if dn, ok := oldBd[names.AttrDeviceName]; ok && dn == "/dev/sda1" {
		bdType = "root_block_device"
	}

	switch bdType {
	case "ebs_block_device":
		delete(oldBd, names.AttrVirtualName)
	case "root_block_device":
		delete(oldBd, names.AttrVirtualName)
		delete(oldBd, names.AttrEncrypted)
		delete(oldBd, names.AttrSnapshotID)
	case "ephemeral_block_device":
		delete(oldBd, names.AttrDeleteOnTermination)
		delete(oldBd, names.AttrEncrypted)
		delete(oldBd, names.AttrIOPS)
		delete(oldBd, names.AttrVolumeSize)
		delete(oldBd, names.AttrVolumeType)
	}
	for attr, val := range oldBd {
		attrKey := fmt.Sprintf("%s.%d.%s", bdType, code, attr)
		is.Attributes[attrKey] = val
	}

	countAttr := fmt.Sprintf("%s.#", bdType)
	count, _ := strconv.Atoi(is.Attributes[countAttr])
	is.Attributes[countAttr] = strconv.Itoa(count + 1)
}

func resourceInstanceV1() *schema.Resource {
	return &schema.Resource{
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
				Deprecated:    "cpu_core_count is deprecated. Use cpu_options instead.",
				ConflictsWith: []string{"cpu_options.0.core_count"},
			},
			"cpu_threads_per_core": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Deprecated:    "cpu_threads_per_core is deprecated. Use cpu_options instead.",
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
				Set: func(v any) int {
					var buf bytes.Buffer
					m := v.(map[string]any)
					fmt.Fprintf(&buf, "%s-", m[names.AttrDeviceName].(string))
					fmt.Fprintf(&buf, "%s-", m[names.AttrSnapshotID].(string))
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
			"enable_primary_ipv6": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
				Set: func(v any) int {
					var buf bytes.Buffer
					m := v.(map[string]any)
					fmt.Fprintf(&buf, "%s-", m[names.AttrDeviceName].(string))
					fmt.Fprintf(&buf, "%s-", m[names.AttrVirtualName].(string))
					if v, ok := m["no_device"].(bool); ok && v {
						fmt.Fprintf(&buf, "%t-", v)
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
				ConflictsWith: []string{"associate_public_ip_address", "enable_primary_ipv6", names.AttrSubnetID, "private_ip", "secondary_private_ips", names.AttrVPCSecurityGroupIDs, names.AttrSecurityGroups, "ipv6_addresses", "ipv6_address_count", "source_dest_check"},
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
				StateFunc: func(v any) string {
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
	}
}

func instanceStateUpgradeV1(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// Initialize cpu_options if it doesn't exist
	cpuOptions, cpuOptionsExisted := rawState["cpu_options"].([]any)
	if !cpuOptionsExisted || len(cpuOptions) == 0 || cpuOptions[0] == nil {
		cpuOptions = []any{map[string]any{}}
	}

	// Get the cpu_options map
	cpuOptionsMap := cpuOptions[0].(map[string]any)

	// Move cpu_core_count to cpu_options if not already set
	if _, exists := cpuOptionsMap["core_count"]; !exists {
		if v, ok := rawState["cpu_core_count"]; ok {
			cpuOptionsMap["core_count"] = v
		}
	}

	// Move cpu_threads_per_core to cpu_options if not already set
	if _, exists := cpuOptionsMap["threads_per_core"]; !exists {
		if v, ok := rawState["cpu_threads_per_core"]; ok {
			cpuOptionsMap["threads_per_core"] = v
		}
	}

	delete(rawState, "cpu_core_count")
	delete(rawState, "cpu_threads_per_core")

	// If cpu_options didn't exist initially and remains empty, don't add it to rawState
	if !cpuOptionsExisted && len(cpuOptionsMap) == 0 {
		return rawState, nil
	}

	// Assign the updated cpu_options map back to rawState
	rawState["cpu_options"] = []any{cpuOptionsMap}

	return rawState, nil
}
