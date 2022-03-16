package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceLaunchTemplateCreate,
		Read:   resourceLaunchTemplateRead,
		Update: resourceLaunchTemplateUpdate,
		Delete: resourceLaunchTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
										ValidateFunc:     verify.ValidTypeStringNullableBoolean,
									},
									"encrypted": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
										ValidateFunc:     verify.ValidTypeStringNullableBoolean,
									},
									"iops": {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"throughput": {
										Type:         schema.TypeInt,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(125, 1000),
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"volume_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
									},
								},
							},
						},
						"device_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"no_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"capacity_reservation_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_reservation_preference": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.CapacityReservationPreference_Values(), false),
						},
						"capacity_reservation_target": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_reservation_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"cpu_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"core_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"threads_per_core": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(CPUCredits_Values(), false),
						},
					},
				},
			},
			"default_version": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"update_default_version"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ebs_optimized": {
				// Use TypeString to allow an "unspecified" value,
				// since TypeBool only has true/false with false default.
				// The conversion from bare true/false values in
				// configurations to TypeString value is currently safe.
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
				ValidateFunc:     verify.ValidTypeStringNullableBoolean,
			},
			"elastic_gpu_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"elastic_inference_accelerator": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"enclave_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"hibernation_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configured": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"iam_instance_profile.0.name"},
							ValidateFunc:  verify.ValidARN,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_initiated_shutdown_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.ShutdownBehavior_Values(), false),
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.MarketType_Values(), false),
						},
						"spot_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntDivisibleBy(60),
									},
									"instance_interruption_behavior": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ec2.InstanceInterruptionBehavior_Values(), false),
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"spot_instance_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ec2.SpotInstanceType_Values(), false),
									},
									"valid_until": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},
								},
							},
						},
					},
				},
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"license_specification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"license_configuration_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataEndpointState_Values(), false),
						},
						"http_protocol_ipv6": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ec2.LaunchTemplateInstanceMetadataProtocolIpv6Disabled,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataProtocolIpv6_Values(), false),
						},
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateHttpTokensState_Values(), false),
						},
						"instance_metadata_tags": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ec2.LaunchTemplateInstanceMetadataTagsStateDisabled,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataTagsState_Values(), false),
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"associate_public_ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"delete_on_termination": {
							// Use TypeString to allow an "unspecified" value,
							// since TypeBool only has true/false with false default.
							// The conversion from bare true/false values in
							// configurations to TypeString value is currently safe.
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"interface_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"efa", "interface"}, false),
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPv4Address,
							},
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv6_address_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv6_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPv6Address,
							},
						},
						"network_card_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"private_ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"placement": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"affinity": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_resource_group_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"placement.0.host_id"},
							ValidateFunc:  verify.ValidARN,
						},
						"partition_number": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"spread_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tenancy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.Tenancy_Values(), false),
						},
					},
				},
			},
			"ram_disk_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_names": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vpc_security_group_ids"},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.ResourceType_Values(), false),
						},
						"tags": tftags.TagsSchema(),
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"update_default_version": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"default_version"},
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_security_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"security_group_names"},
			},
		},

		// Enable downstream updates for resources referencing schema attributes
		// to prevent non-empty plans after "terraform apply"
		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("default_version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				for _, changedKey := range diff.GetChangedKeysPrefix("") {
					switch changedKey {
					case "name", "name_prefix", "description":
						continue
					default:
						return diff.Get("update_default_version").(bool)
					}
				}
				return false
			}),
			customdiff.ComputedIf("latest_version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				for _, changedKey := range diff.GetChangedKeysPrefix("") {
					switch changedKey {
					case "name", "name_prefix", "description", "default_version", "update_default_version":
						continue
					default:
						return true
					}
				}
				return false
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceLaunchTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &ec2.CreateLaunchTemplateInput{
		ClientToken:        aws.String(resource.UniqueId()),
		LaunchTemplateName: aws.String(name),
		LaunchTemplateData: expandRequestLaunchTemplateData(d),
		TagSpecifications:  ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeLaunchTemplate),
	}

	if v, ok := d.GetOk("description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Launch Template: %s", input)
	output, err := conn.CreateLaunchTemplate(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Launch Template: %w", err)
	}

	d.SetId(aws.StringValue(output.LaunchTemplate.LaunchTemplateId))

	return resourceLaunchTemplateRead(d, meta)
}

func resourceLaunchTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lt, err := FindLaunchTemplateByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Launch Template %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Launch Template (%s): %w", d.Id(), err)
	}

	version := strconv.FormatInt(aws.Int64Value(lt.LatestVersionNumber), 10)
	ltv, err := FindLaunchTemplateVersionByTwoPartKey(conn, d.Id(), version)

	if err != nil {
		return fmt.Errorf("error reading EC2 Launch Template (%s) Version (%s): %w", d.Id(), version, err)
	}

	ltd := ltv.LaunchTemplateData

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("launch-template/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if err := d.Set("block_device_mappings", getBlockDeviceMappings(ltd.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting block_device_mappings: %w", err)
	}
	if err := d.Set("capacity_reservation_specification", getCapacityReservationSpecification(ltd.CapacityReservationSpecification)); err != nil {
		return fmt.Errorf("error setting capacity_reservation_specification: %w", err)
	}
	if err := d.Set("cpu_options", getCpuOptions(ltd.CpuOptions)); err != nil {
		return fmt.Errorf("error setting cpu_options: %w", err)
	}
	if strings.HasPrefix(aws.StringValue(ltd.InstanceType), "t2") || strings.HasPrefix(aws.StringValue(ltd.InstanceType), "t3") {
		if err := d.Set("credit_specification", getCreditSpecification(ltd.CreditSpecification)); err != nil {
			return fmt.Errorf("error setting credit_specification: %w", err)
		}
	}
	d.Set("default_version", lt.DefaultVersionNumber)
	d.Set("description", ltv.VersionDescription)
	d.Set("disable_api_termination", ltd.DisableApiTermination)
	if ltd.EbsOptimized != nil {
		d.Set("ebs_optimized", strconv.FormatBool(aws.BoolValue(ltd.EbsOptimized)))
	} else {
		d.Set("ebs_optimized", "")
	}
	if err := d.Set("elastic_gpu_specifications", getElasticGpuSpecifications(ltd.ElasticGpuSpecifications)); err != nil {
		return fmt.Errorf("error setting elastic_gpu_specifications: %w", err)
	}
	if err := d.Set("elastic_inference_accelerator", flattenEc2LaunchTemplateElasticInferenceAcceleratorResponse(ltd.ElasticInferenceAccelerators)); err != nil {
		return fmt.Errorf("error setting elastic_inference_accelerator: %w", err)
	}
	if err := d.Set("enclave_options", getEnclaveOptions(ltd.EnclaveOptions)); err != nil {
		return fmt.Errorf("error setting enclave_options: %w", err)
	}
	if err := d.Set("hibernation_options", flattenLaunchTemplateHibernationOptions(ltd.HibernationOptions)); err != nil {
		return fmt.Errorf("error setting hibernation_options: %w", err)
	}
	if err := d.Set("iam_instance_profile", getIamInstanceProfile(ltd.IamInstanceProfile)); err != nil {
		return fmt.Errorf("error setting iam_instance_profile: %w", err)
	}
	d.Set("image_id", ltd.ImageId)
	d.Set("instance_initiated_shutdown_behavior", ltd.InstanceInitiatedShutdownBehavior)
	if err := d.Set("instance_market_options", getInstanceMarketOptions(ltd.InstanceMarketOptions)); err != nil {
		return fmt.Errorf("error setting instance_market_options: %w", err)
	}
	d.Set("instance_type", ltd.InstanceType)
	d.Set("kernel_id", ltd.KernelId)
	d.Set("key_name", ltd.KeyName)
	d.Set("latest_version", lt.LatestVersionNumber)
	if err := d.Set("license_specification", getLicenseSpecifications(ltd.LicenseSpecifications)); err != nil {
		return fmt.Errorf("error setting license_specification: %w", err)
	}
	if err := d.Set("metadata_options", flattenLaunchTemplateInstanceMetadataOptions(ltd.MetadataOptions)); err != nil {
		return fmt.Errorf("error setting metadata_options: %w", err)
	}
	if err := d.Set("monitoring", getMonitoring(ltd.Monitoring)); err != nil {
		return fmt.Errorf("error setting monitoring: %w", err)
	}
	d.Set("name", lt.LaunchTemplateName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lt.LaunchTemplateName)))
	if err := d.Set("network_interfaces", getNetworkInterfaces(ltd.NetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting network_interfaces: %w", err)
	}
	if err := d.Set("placement", getPlacement(ltd.Placement)); err != nil {
		return fmt.Errorf("error setting placement: %w", err)
	}
	d.Set("ram_disk_id", ltd.RamDiskId)
	d.Set("security_group_names", aws.StringValueSlice(ltd.SecurityGroups))
	if err := d.Set("tag_specifications", getTagSpecifications(ltd.TagSpecifications)); err != nil {
		return fmt.Errorf("error setting tag_specifications: %w", err)
	}
	d.Set("user_data", ltd.UserData)
	d.Set("vpc_security_group_ids", aws.StringValueSlice(ltd.SecurityGroupIds))

	tags := KeyValueTags(lt.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLaunchTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	updateKeys := []string{
		"block_device_mappings",
		"capacity_reservation_specification",
		"cpu_options",
		"credit_specification",
		"description",
		"disable_api_termination",
		"ebs_optimized",
		"elastic_gpu_specifications",
		"elastic_inference_accelerator",
		"enclave_options",
		"hibernation_options",
		"iam_instance_profile",
		"image_id",
		"instance_initiated_shutdown_behavior",
		"instance_market_options",
		"instance_type",
		"kernel_id",
		"key_name",
		"license_specification",
		"metadata_options",
		"monitoring",
		"network_interfaces",
		"placement",
		"ram_disk_id",
		"security_group_names",
		"tag_specifications",
		"user_data",
		"vpc_security_group_ids",
	}
	latestVersion := int64(d.Get("latest_version").(int))

	if d.HasChanges(updateKeys...) {
		input := &ec2.CreateLaunchTemplateVersionInput{
			ClientToken:        aws.String(resource.UniqueId()),
			LaunchTemplateData: expandRequestLaunchTemplateData(d),
			LaunchTemplateId:   aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.VersionDescription = aws.String(v.(string))
		}

		output, err := conn.CreateLaunchTemplateVersion(input)

		if err != nil {
			return fmt.Errorf("error creating EC2 Launch Template (%s) Version: %w", d.Id(), err)
		}

		latestVersion = aws.Int64Value(output.LaunchTemplateVersion.VersionNumber)

	}

	if d.Get("update_default_version").(bool) || d.HasChange("default_version") {
		input := &ec2.ModifyLaunchTemplateInput{
			LaunchTemplateId: aws.String(d.Id()),
		}

		if d.Get("update_default_version").(bool) {
			input.DefaultVersion = aws.String(strconv.FormatInt(latestVersion, 10))
		} else if d.HasChange("default_version") {
			input.DefaultVersion = aws.String(strconv.Itoa(d.Get("default_version").(int)))
		}

		_, err := conn.ModifyLaunchTemplate(input)

		if err != nil {
			return fmt.Errorf("error updating EC2 Launch Template (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Launch Template (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLaunchTemplateRead(d, meta)
}

func resourceLaunchTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Launch Template: %s", d.Id())
	_, err := conn.DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidLaunchTemplateIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Launch Template (%s): %w", d.Id(), err)
	}

	return nil
}

func getBlockDeviceMappings(m []*ec2.LaunchTemplateBlockDeviceMapping) []interface{} {
	s := []interface{}{}
	for _, v := range m {
		mapping := map[string]interface{}{
			"device_name":  aws.StringValue(v.DeviceName),
			"virtual_name": aws.StringValue(v.VirtualName),
		}
		if v.NoDevice != nil {
			mapping["no_device"] = aws.StringValue(v.NoDevice)
		}
		if v.Ebs != nil {
			ebs := map[string]interface{}{
				"volume_size": int(aws.Int64Value(v.Ebs.VolumeSize)),
				"volume_type": aws.StringValue(v.Ebs.VolumeType),
			}
			if v.Ebs.DeleteOnTermination != nil {
				ebs["delete_on_termination"] = strconv.FormatBool(aws.BoolValue(v.Ebs.DeleteOnTermination))
			}
			if v.Ebs.Encrypted != nil {
				ebs["encrypted"] = strconv.FormatBool(aws.BoolValue(v.Ebs.Encrypted))
			}
			if v.Ebs.Iops != nil {
				ebs["iops"] = aws.Int64Value(v.Ebs.Iops)
			}
			if v.Ebs.KmsKeyId != nil {
				ebs["kms_key_id"] = aws.StringValue(v.Ebs.KmsKeyId)
			}
			if v.Ebs.SnapshotId != nil {
				ebs["snapshot_id"] = aws.StringValue(v.Ebs.SnapshotId)
			}
			if v.Ebs.Throughput != nil {
				ebs["throughput"] = aws.Int64Value(v.Ebs.Throughput)
			}

			mapping["ebs"] = []interface{}{ebs}
		}
		s = append(s, mapping)
	}
	return s
}

func getCapacityReservationSpecification(crs *ec2.LaunchTemplateCapacityReservationSpecificationResponse) []interface{} {
	s := []interface{}{}
	if crs != nil {
		s = append(s, map[string]interface{}{
			"capacity_reservation_preference": aws.StringValue(crs.CapacityReservationPreference),
			"capacity_reservation_target":     getCapacityReservationTarget(crs.CapacityReservationTarget),
		})
	}
	return s
}

func getCapacityReservationTarget(crt *ec2.CapacityReservationTargetResponse) []interface{} {
	s := []interface{}{}
	if crt != nil {
		s = append(s, map[string]interface{}{
			"capacity_reservation_id": aws.StringValue(crt.CapacityReservationId),
		})
	}
	return s
}

func getCpuOptions(cs *ec2.LaunchTemplateCpuOptions) []interface{} {
	s := []interface{}{}
	if cs != nil {
		s = append(s, map[string]interface{}{
			"core_count":       aws.Int64Value(cs.CoreCount),
			"threads_per_core": aws.Int64Value(cs.ThreadsPerCore),
		})
	}
	return s
}

func getCreditSpecification(cs *ec2.CreditSpecification) []interface{} {
	s := []interface{}{}
	if cs != nil {
		s = append(s, map[string]interface{}{
			"cpu_credits": aws.StringValue(cs.CpuCredits),
		})
	}
	return s
}

func getElasticGpuSpecifications(e []*ec2.ElasticGpuSpecificationResponse) []interface{} {
	s := []interface{}{}
	for _, v := range e {
		s = append(s, map[string]interface{}{
			"type": aws.StringValue(v.Type),
		})
	}
	return s
}

func flattenEc2LaunchTemplateElasticInferenceAcceleratorResponse(accelerators []*ec2.LaunchTemplateElasticInferenceAcceleratorResponse) []interface{} {
	l := []interface{}{}

	for _, accelerator := range accelerators {
		if accelerator == nil {
			continue
		}

		m := map[string]interface{}{
			"type": aws.StringValue(accelerator.Type),
		}

		l = append(l, m)
	}

	return l
}

func getIamInstanceProfile(i *ec2.LaunchTemplateIamInstanceProfileSpecification) []interface{} {
	s := []interface{}{}
	if i != nil {
		s = append(s, map[string]interface{}{
			"arn":  aws.StringValue(i.Arn),
			"name": aws.StringValue(i.Name),
		})
	}
	return s
}

func getInstanceMarketOptions(m *ec2.LaunchTemplateInstanceMarketOptions) []interface{} {
	s := []interface{}{}
	if m != nil {
		mo := map[string]interface{}{
			"market_type": aws.StringValue(m.MarketType),
		}
		so := m.SpotOptions
		if so != nil {
			spotOptions := map[string]interface{}{}

			if so.BlockDurationMinutes != nil {
				spotOptions["block_duration_minutes"] = aws.Int64Value(so.BlockDurationMinutes)
			}

			if so.InstanceInterruptionBehavior != nil {
				spotOptions["instance_interruption_behavior"] = aws.StringValue(so.InstanceInterruptionBehavior)
			}

			if so.MaxPrice != nil {
				spotOptions["max_price"] = aws.StringValue(so.MaxPrice)
			}

			if so.SpotInstanceType != nil {
				spotOptions["spot_instance_type"] = aws.StringValue(so.SpotInstanceType)
			}

			if so.ValidUntil != nil {
				spotOptions["valid_until"] = aws.TimeValue(so.ValidUntil).Format(time.RFC3339)
			}

			mo["spot_options"] = []interface{}{spotOptions}
		}
		s = append(s, mo)
	}
	return s
}

func getLicenseSpecifications(licenseSpecifications []*ec2.LaunchTemplateLicenseConfiguration) []map[string]interface{} {
	var s []map[string]interface{}
	for _, v := range licenseSpecifications {
		s = append(s, map[string]interface{}{
			"license_configuration_arn": aws.StringValue(v.LicenseConfigurationArn),
		})
	}
	return s
}

func flattenLaunchTemplateInstanceMetadataOptions(opts *ec2.LaunchTemplateInstanceMetadataOptions) []interface{} {
	if opts == nil {
		return nil
	}

	m := map[string]interface{}{
		"http_endpoint":               aws.StringValue(opts.HttpEndpoint),
		"http_protocol_ipv6":          aws.StringValue(opts.HttpProtocolIpv6),
		"http_put_response_hop_limit": aws.Int64Value(opts.HttpPutResponseHopLimit),
		"http_tokens":                 aws.StringValue(opts.HttpTokens),
		"instance_metadata_tags":      aws.StringValue(opts.InstanceMetadataTags),
	}

	return []interface{}{m}
}

func getEnclaveOptions(m *ec2.LaunchTemplateEnclaveOptions) []interface{} {
	s := []interface{}{}
	if m != nil {
		mo := map[string]interface{}{
			"enabled": aws.BoolValue(m.Enabled),
		}
		s = append(s, mo)
	}
	return s
}

func getMonitoring(m *ec2.LaunchTemplatesMonitoring) []interface{} {
	s := []interface{}{}
	if m != nil {
		mo := map[string]interface{}{
			"enabled": aws.BoolValue(m.Enabled),
		}
		s = append(s, mo)
	}
	return s
}

func getNetworkInterfaces(n []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecification) []interface{} {
	s := []interface{}{}
	for _, v := range n {
		var ipv4Addresses []string

		networkInterface := map[string]interface{}{
			"description":          aws.StringValue(v.Description),
			"device_index":         aws.Int64Value(v.DeviceIndex),
			"interface_type":       aws.StringValue(v.InterfaceType),
			"ipv4_address_count":   aws.Int64Value(v.SecondaryPrivateIpAddressCount),
			"ipv6_address_count":   aws.Int64Value(v.Ipv6AddressCount),
			"network_card_index":   aws.Int64Value(v.NetworkCardIndex),
			"network_interface_id": aws.StringValue(v.NetworkInterfaceId),
			"private_ip_address":   aws.StringValue(v.PrivateIpAddress),
			"subnet_id":            aws.StringValue(v.SubnetId),
		}

		if v.AssociateCarrierIpAddress != nil {
			networkInterface["associate_carrier_ip_address"] = strconv.FormatBool(aws.BoolValue(v.AssociateCarrierIpAddress))
		}

		if v.AssociatePublicIpAddress != nil {
			networkInterface["associate_public_ip_address"] = strconv.FormatBool(aws.BoolValue(v.AssociatePublicIpAddress))
		}

		if v.DeleteOnTermination != nil {
			networkInterface["delete_on_termination"] = strconv.FormatBool(aws.BoolValue(v.DeleteOnTermination))
		}

		if len(v.Ipv6Addresses) > 0 {
			raw, ok := networkInterface["ipv6_addresses"]
			if !ok {
				raw = schema.NewSet(schema.HashString, nil)
			}

			list := raw.(*schema.Set)

			for _, address := range v.Ipv6Addresses {
				list.Add(aws.StringValue(address.Ipv6Address))
			}

			networkInterface["ipv6_addresses"] = list
		}

		for _, address := range v.PrivateIpAddresses {
			ipv4Addresses = append(ipv4Addresses, aws.StringValue(address.PrivateIpAddress))
		}
		if len(ipv4Addresses) > 0 {
			networkInterface["ipv4_addresses"] = ipv4Addresses
		}

		if len(v.Groups) > 0 {
			raw, ok := networkInterface["security_groups"]
			if !ok {
				raw = schema.NewSet(schema.HashString, nil)
			}
			list := raw.(*schema.Set)

			for _, group := range v.Groups {
				list.Add(aws.StringValue(group))
			}

			networkInterface["security_groups"] = list
		}

		s = append(s, networkInterface)
	}
	return s
}

func getPlacement(p *ec2.LaunchTemplatePlacement) []interface{} {
	var s []interface{}
	if p != nil {
		s = append(s, map[string]interface{}{
			"affinity":                aws.StringValue(p.Affinity),
			"availability_zone":       aws.StringValue(p.AvailabilityZone),
			"group_name":              aws.StringValue(p.GroupName),
			"host_id":                 aws.StringValue(p.HostId),
			"host_resource_group_arn": aws.StringValue(p.HostResourceGroupArn),
			"spread_domain":           aws.StringValue(p.SpreadDomain),
			"tenancy":                 aws.StringValue(p.Tenancy),
			"partition_number":        aws.Int64Value(p.PartitionNumber),
		})
	}
	return s
}

func flattenLaunchTemplateHibernationOptions(m *ec2.LaunchTemplateHibernationOptions) []interface{} {
	s := []interface{}{}
	if m != nil {
		mo := map[string]interface{}{
			"configured": aws.BoolValue(m.Configured),
		}
		s = append(s, mo)
	}
	return s
}

func getTagSpecifications(t []*ec2.LaunchTemplateTagSpecification) []interface{} {
	var s []interface{}
	for _, v := range t {
		s = append(s, map[string]interface{}{
			"resource_type": aws.StringValue(v.ResourceType),
			"tags":          KeyValueTags(v.Tags).IgnoreAWS().Map(),
		})
	}
	return s
}

func expandRequestLaunchTemplateData(d *schema.ResourceData) *ec2.RequestLaunchTemplateData {
	apiObject := &ec2.RequestLaunchTemplateData{
		// Always set at least one field.
		UserData: aws.String(d.Get("user_data").(string)),
	}

	var instanceType string
	if v, ok := d.GetOk("instance_type"); ok {
		v := v.(string)

		instanceType = v
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := d.GetOk("block_device_mappings"); ok && len(v.([]interface{})) > 0 {
		apiObject.BlockDeviceMappings = expandLaunchTemplateBlockDeviceMappingRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("capacity_reservation_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CapacityReservationSpecification = expandLaunchTemplateCapacityReservationSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cpu_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CpuOptions = expandLaunchTemplateCpuOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("credit_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil && (strings.HasPrefix(instanceType, "t2") || strings.HasPrefix(instanceType, "t3")) {
		apiObject.CreditSpecification = expandCreditSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("disable_api_termination"); ok {
		apiObject.DisableApiTermination = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		v, _ := strconv.ParseBool(v.(string))

		apiObject.EbsOptimized = aws.Bool(v)
	}

	if v, ok := d.GetOk("elastic_gpu_specifications"); ok && len(v.([]interface{})) > 0 {
		apiObject.ElasticGpuSpecifications = expandElasticGpuSpecifications(v.([]interface{}))
	}

	if v, ok := d.GetOk("elastic_inference_accelerator"); ok && len(v.([]interface{})) > 0 {
		apiObject.ElasticInferenceAccelerators = expandLaunchTemplateElasticInferenceAccelerators(v.([]interface{}))
	}

	if v, ok := d.GetOk("enclave_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.EnclaveOptions = &ec2.LaunchTemplateEnclaveOptionsRequest{
			Enabled: aws.Bool(tfMap["enabled"].(bool)),
		}
	}

	if v, ok := d.GetOk("hibernation_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.HibernationOptions = &ec2.LaunchTemplateHibernationOptionsRequest{
			Configured: aws.Bool(tfMap["configured"].(bool)),
		}
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.IamInstanceProfile = expandLaunchTemplateIamInstanceProfileSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("image_id"); ok {
		apiObject.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_initiated_shutdown_behavior"); ok {
		apiObject.InstanceInitiatedShutdownBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_market_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.InstanceMarketOptions = expandLaunchTemplateInstanceMarketOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("kernel_id"); ok {
		apiObject.KernelId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key_name"); ok {
		apiObject.KeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_specification"); ok && len(v.([]interface{})) > 0 {
		apiObject.LicenseSpecifications = expandLaunchTemplateLicenseConfigurationRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("metadata_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.MetadataOptions = expandLaunchTemplateInstanceMetadataOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{
			Enabled: aws.Bool(tfMap["enabled"].(bool)),
		}
	}

	if v, ok := d.GetOk("network_interfaces"); ok && len(v.([]interface{})) > 0 {
		apiObject.NetworkInterfaces = expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("placement"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Placement = expandLaunchTemplatePlacementRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ram_disk_id"); ok {
		apiObject.RamDiskId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_names"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_specifications"); ok && len(v.([]interface{})) > 0 {
		apiObject.TagSpecifications = expandLaunchTemplateTagSpecificationRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject
}

func expandLaunchTemplateBlockDeviceMappingRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateBlockDeviceMappingRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateBlockDeviceMappingRequest{}

	if v, ok := tfMap["ebs"].([]interface{}); ok && len(v) > 0 {
		apiObject.Ebs = expandLaunchTemplateEbsBlockDeviceRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(string); ok && v != "" {
		apiObject.NoDevice = aws.String(v)
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateBlockDeviceMappingRequests(tfList []interface{}) []*ec2.LaunchTemplateBlockDeviceMappingRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateBlockDeviceMappingRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateBlockDeviceMappingRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateEbsBlockDeviceRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateEbsBlockDeviceRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateEbsBlockDeviceRequest{}

	if v, ok := tfMap["delete_on_termination"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["encrypted"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateCapacityReservationSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateCapacityReservationSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateCapacityReservationSpecificationRequest{}

	if v, ok := tfMap["capacity_reservation_preference"].(string); ok && v != "" {
		apiObject.CapacityReservationPreference = aws.String(v)
	}

	if v, ok := tfMap["capacity_reservation_target"].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityReservationTarget = expandCapacityReservationTarget(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCapacityReservationTarget(tfMap map[string]interface{}) *ec2.CapacityReservationTarget {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.CapacityReservationTarget{}

	if v, ok := tfMap["capacity_reservation_id"].(string); ok && v != "" {
		apiObject.CapacityReservationId = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateCpuOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateCpuOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateCpuOptionsRequest{}

	if v, ok := tfMap["core_count"].(int); ok && v != 0 {
		apiObject.CoreCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["threads_per_core"].(int); ok && v != 0 {
		apiObject.ThreadsPerCore = aws.Int64(int64(v))
	}

	return apiObject
}

func expandCreditSpecificationRequest(tfMap map[string]interface{}) *ec2.CreditSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.CreditSpecificationRequest{}

	if v, ok := tfMap["cpu_credits"].(string); ok && v != "" {
		apiObject.CpuCredits = aws.String(v)
	}

	return apiObject
}

func expandElasticGpuSpecification(tfMap map[string]interface{}) *ec2.ElasticGpuSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.ElasticGpuSpecification{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandElasticGpuSpecifications(tfList []interface{}) []*ec2.ElasticGpuSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.ElasticGpuSpecification

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandElasticGpuSpecification(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateElasticInferenceAccelerator(tfMap map[string]interface{}) *ec2.LaunchTemplateElasticInferenceAccelerator {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateElasticInferenceAccelerator{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateElasticInferenceAccelerators(tfList []interface{}) []*ec2.LaunchTemplateElasticInferenceAccelerator {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateElasticInferenceAccelerator

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateElasticInferenceAccelerator(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateIamInstanceProfileSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateIamInstanceProfileSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceMarketOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceMarketOptionsRequest{}

	if v, ok := tfMap["market_type"].(string); ok && v != "" {
		apiObject.MarketType = aws.String(v)
	}

	if v, ok := tfMap["spot_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.SpotOptions = expandLaunchTemplateSpotMarketOptionsRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLaunchTemplateSpotMarketOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateSpotMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateSpotMarketOptionsRequest{}

	if v, ok := tfMap["block_duration_minutes"].(int); ok && v != 0 {
		apiObject.BlockDurationMinutes = aws.Int64(int64(v))
	}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = aws.String(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_type"].(string); ok && v != "" {
		apiObject.SpotInstanceType = aws.String(v)
	}

	if v, ok := tfMap["valid_until"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.ValidUntil = aws.Time(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateLicenseConfigurationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateLicenseConfigurationRequest{}

	if v, ok := tfMap["license_configuration_arn"].(string); ok && v != "" {
		apiObject.LicenseConfigurationArn = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequests(tfList []interface{}) []*ec2.LaunchTemplateLicenseConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateLicenseConfigurationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateLicenseConfigurationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateInstanceMetadataOptions(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceMetadataOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceMetadataOptionsRequest{}

	if v, ok := tfMap["http_endpoint"].(string); ok && v != "" {
		apiObject.HttpEndpoint = aws.String(v)

		if v == ec2.LaunchTemplateInstanceMetadataEndpointStateEnabled {
			// These parameters are not allowed unless HttpEndpoint is enabled.
			if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
				apiObject.HttpTokens = aws.String(v)
			}

			if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
				apiObject.HttpPutResponseHopLimit = aws.Int64(int64(v))
			}

			if v, ok := tfMap["instance_metadata_tags"].(string); ok && v != "" {
				apiObject.InstanceMetadataTags = aws.String(v)
			}
		}
	}

	if v, ok := tfMap["http_protocol_ipv6"].(string); ok && v != "" {
		apiObject.HttpProtocolIpv6 = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}

	if v, ok := tfMap["associate_carrier_ip_address"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.AssociateCarrierIpAddress = aws.Bool(v)
	}

	if v, ok := tfMap["associate_public_ip_address"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.AssociatePublicIpAddress = aws.Bool(v)
	}

	if v, ok := tfMap["delete_on_termination"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["device_index"].(int); ok && v != 0 {
		apiObject.DeviceIndex = aws.Int64(int64(v))
	}

	if v, ok := tfMap["interface_type"].(string); ok && v != "" {
		apiObject.InterfaceType = aws.String(v)
	}

	var privateIPAddress string

	if v, ok := tfMap["private_ip_address"].(string); ok && v != "" {
		privateIPAddress = v
		apiObject.PrivateIpAddress = aws.String(v)
	}

	if v, ok := tfMap["ipv4_address_count"].(int); ok && v != 0 {
		apiObject.SecondaryPrivateIpAddressCount = aws.Int64(int64(v))
	} else if v, ok := tfMap["ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			v := v.(string)

			apiObject.PrivateIpAddresses = append(apiObject.PrivateIpAddresses, &ec2.PrivateIpAddressSpecification{
				Primary:          aws.Bool(v == privateIPAddress),
				PrivateIpAddress: aws.String(v),
			})
		}
	}

	if v, ok := tfMap["ipv6_address_count"].(int); ok && v != 0 {
		apiObject.Ipv6AddressCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Ipv6Addresses = append(apiObject.Ipv6Addresses, &ec2.InstanceIpv6AddressRequest{
				Ipv6Address: aws.String(v.(string)),
			})
		}
	}

	if v, ok := tfMap["network_card_index"].(int); ok && v != 0 {
		apiObject.NetworkCardIndex = aws.Int64(int64(v))
	}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Groups = append(apiObject.Groups, aws.String(v.(string)))
		}
	}

	if v, ok := tfMap["subnet_id"].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequests(tfList []interface{}) []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplatePlacementRequest(tfMap map[string]interface{}) *ec2.LaunchTemplatePlacementRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplatePlacementRequest{}

	if v, ok := tfMap["affinity"].(string); ok && v != "" {
		apiObject.Affinity = aws.String(v)
	}

	if v, ok := tfMap["availability_zone"].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["group_name"].(string); ok && v != "" {
		apiObject.GroupName = aws.String(v)
	}

	if v, ok := tfMap["host_id"].(string); ok && v != "" {
		apiObject.HostId = aws.String(v)
	}

	if v, ok := tfMap["host_resource_group_arn"].(string); ok && v != "" {
		apiObject.HostResourceGroupArn = aws.String(v)
	}

	if v, ok := tfMap["partition_number"].(int); ok && v != 0 {
		apiObject.PartitionNumber = aws.Int64(int64(v))
	}

	if v, ok := tfMap["spread_domain"].(string); ok && v != "" {
		apiObject.SpreadDomain = aws.String(v)
	}

	if v, ok := tfMap["tenancy"].(string); ok && v != "" {
		apiObject.Tenancy = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateTagSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateTagSpecificationRequest{}

	if v, ok := tfMap["resource_type"].(string); ok && v != "" {
		apiObject.ResourceType = aws.String(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		if v := tftags.New(v).IgnoreAWS(); len(v) > 0 {
			apiObject.Tags = Tags(v)
		}
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequests(tfList []interface{}) []*ec2.LaunchTemplateTagSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateTagSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateTagSpecificationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
