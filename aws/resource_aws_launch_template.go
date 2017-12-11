package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const awsSpotInstanceTimeLayout = "2006-01-02T15:04:05Z"

func resourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLaunchTemplateCreate,
		Read:   resourceAwsLaunchTemplateRead,
		Update: resourceAwsLaunchTemplateUpdate,
		Delete: resourceAwsLaunchTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateLaunchTemplateName,
			},

			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateLaunchTemplateName,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if len(value) > 255 {
						errors = append(errors, fmt.Errorf(
							"%q cannot be longer than 255 characters", k))
					}
					return
				},
			},

			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"default_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"block_device_mappings": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"ebs": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},

			"credit_specification": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"elastic_gpu_specifications": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"iam_instance_profile": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Optional: true,
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
				Type:     schema.TypeString,
				Optional: true,
			},

			"instance_market_options": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"spot_options": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"instance_interruption_behavior": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"spot_instance_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"valid_until": {
										Type:     schema.TypeString,
										Optional: true,
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

			"monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"network_interfaces": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_public_ip_address": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"ipv6_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv6_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"placement": {
				Type:     schema.TypeSet,
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
						"spread_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tenancy": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"ram_disk_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tag_specifications": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tags": tagsSchema(),
					},
				},
			},

			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsLaunchTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	var ltName string
	if v, ok := d.GetOk("name"); ok {
		ltName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		ltName = resource.PrefixedUniqueId(v.(string))
	} else {
		ltName = resource.UniqueId()
	}

	launchTemplateData, err := buildLaunchTemplateData(d, meta)
	if err != nil {
		return err
	}

	launchTemplateDataOpts := &ec2.RequestLaunchTemplateData{
		BlockDeviceMappings:      launchTemplateData.BlockDeviceMappings,
		CreditSpecification:      launchTemplateData.CreditSpecification,
		DisableApiTermination:    launchTemplateData.DisableApiTermination,
		EbsOptimized:             launchTemplateData.EbsOptimized,
		ElasticGpuSpecifications: launchTemplateData.ElasticGpuSpecifications,
		IamInstanceProfile:       launchTemplateData.IamInstanceProfile,
		ImageId:                  launchTemplateData.ImageId,
		InstanceInitiatedShutdownBehavior: launchTemplateData.InstanceInitiatedShutdownBehavior,
		InstanceMarketOptions:             launchTemplateData.InstanceMarketOptions,
		InstanceType:                      launchTemplateData.InstanceType,
		KernelId:                          launchTemplateData.KernelId,
		KeyName:                           launchTemplateData.KeyName,
		Monitoring:                        launchTemplateData.Monitoring,
		NetworkInterfaces:                 launchTemplateData.NetworkInterfaces,
		Placement:                         launchTemplateData.Placement,
		RamDiskId:                         launchTemplateData.RamDiskId,
		SecurityGroups:                    launchTemplateData.SecurityGroups,
		SecurityGroupIds:                  launchTemplateData.SecurityGroupIds,
		TagSpecifications:                 launchTemplateData.TagSpecifications,
		UserData:                          launchTemplateData.UserData,
	}

	launchTemplateOpts := &ec2.CreateLaunchTemplateInput{
		ClientToken:        aws.String(resource.UniqueId()),
		LaunchTemplateName: aws.String(ltName),
		LaunchTemplateData: launchTemplateDataOpts,
	}

	resp, err := conn.CreateLaunchTemplate(launchTemplateOpts)
	if err != nil {
		return err
	}

	launchTemplate := resp.LaunchTemplate
	d.SetId(*launchTemplate.LaunchTemplateId)

	log.Printf("[DEBUG] Launch Template created: %q (version %d)",
		*launchTemplate.LaunchTemplateId, *launchTemplate.LatestVersionNumber)

	return resourceAwsLaunchTemplateUpdate(d, meta)
}

func resourceAwsLaunchTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading launch template %s", d.Id())

	dlt, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		return fmt.Errorf("Error getting launch template: %s", err)
	}
	if len(dlt.LaunchTemplates) == 0 {
		d.SetId("")
		return nil
	}
	if *dlt.LaunchTemplates[0].LaunchTemplateId != d.Id() {
		return fmt.Errorf("Unable to find launch template: %#v", dlt.LaunchTemplates)
	}

	log.Printf("[DEBUG] Found launch template %s", d.Id())

	lt := dlt.LaunchTemplates[0]
	d.Set("name", lt.LaunchTemplateName)
	d.Set("latest_version", lt.LatestVersionNumber)
	d.Set("default_version", lt.DefaultVersionNumber)
	d.Set("tags", tagsToMap(lt.Tags))

	version := strconv.Itoa(int(*lt.LatestVersionNumber))
	dltv, err := conn.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(d.Id()),
		Versions:         []*string{aws.String(version)},
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received launch template version %q (version %d)", d.Id(), *lt.LatestVersionNumber)

	ltData := dltv.LaunchTemplateVersions[0].LaunchTemplateData

	d.Set("disable_api_termination", ltData.DisableApiTermination)
	d.Set("ebs_optimized", ltData.EbsOptimized)
	d.Set("image_id", ltData.ImageId)
	d.Set("instance_initiated_shutdown_behavior", ltData.InstanceInitiatedShutdownBehavior)
	d.Set("instance_type", ltData.InstanceType)
	d.Set("kernel_id", ltData.KernelId)
	d.Set("key_name", ltData.KeyName)
	d.Set("monitoring", ltData.Monitoring)
	d.Set("ram_dist_id", ltData.RamDiskId)
	d.Set("user_data", ltData.UserData)

	if err := d.Set("block_device_mappings", getBlockDeviceMappings(ltData.BlockDeviceMappings)); err != nil {
		return err
	}

	if err := d.Set("credit_specification", getCreditSpecification(ltData.CreditSpecification)); err != nil {
		return err
	}

	if err := d.Set("elastic_gpu_specifications", getElasticGpuSpecifications(ltData.ElasticGpuSpecifications)); err != nil {
		return err
	}

	if err := d.Set("iam_instance_profile", getIamInstanceProfile(ltData.IamInstanceProfile)); err != nil {
		return err
	}

	if err := d.Set("instance_market_options", getInstanceMarketOptions(ltData.InstanceMarketOptions)); err != nil {
		return err
	}

	if err := d.Set("network_interfaces", getNetworkInterfaces(ltData.NetworkInterfaces)); err != nil {
		return err
	}

	if err := d.Set("placement", getPlacement(ltData.Placement)); err != nil {
		return err
	}

	if err := d.Set("tag_specifications", getTagSpecifications(ltData.TagSpecifications)); err != nil {
		return err
	}

	return nil
}

func resourceAwsLaunchTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsLaunchTemplateRead(d, meta)
}

func resourceAwsLaunchTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Launch Template destroy: %v", d.Id())
	_, err := conn.DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Launch Template deleted: %v", d.Id())
	return nil
}

func getBlockDeviceMappings(m []*ec2.LaunchTemplateBlockDeviceMapping) []interface{} {
	s := []interface{}{}
	for _, v := range m {
		mapping := map[string]interface{}{
			"device_name":  *v.DeviceName,
			"virtual_name": *v.VirtualName,
		}
		if v.NoDevice != nil {
			mapping["no_device"] = *v.NoDevice
		}
		if v.Ebs != nil {
			ebs := map[string]interface{}{
				"delete_on_termination": *v.Ebs.DeleteOnTermination,
				"encrypted":             *v.Ebs.Encrypted,
				"volume_size":           *v.Ebs.VolumeSize,
				"volume_type":           *v.Ebs.VolumeType,
			}
			if v.Ebs.Iops != nil {
				ebs["iops"] = *v.Ebs.Iops
			}
			if v.Ebs.KmsKeyId != nil {
				ebs["kms_key_id"] = *v.Ebs.KmsKeyId
			}
			if v.Ebs.SnapshotId != nil {
				ebs["snapshot_id"] = *v.Ebs.SnapshotId
			}

			mapping["ebs"] = ebs
		}
		s = append(s, mapping)
	}
	return s
}

func getCreditSpecification(cs *ec2.CreditSpecification) []interface{} {
	s := []interface{}{}
	if cs != nil {
		s = append(s, map[string]interface{}{
			"cpu_credits": *cs.CpuCredits,
		})
	}
	return s
}

func getElasticGpuSpecifications(e []*ec2.ElasticGpuSpecificationResponse) []interface{} {
	s := []interface{}{}
	for _, v := range e {
		s = append(s, map[string]interface{}{
			"type": *v.Type,
		})
	}
	return s
}

func getIamInstanceProfile(i *ec2.LaunchTemplateIamInstanceProfileSpecification) []interface{} {
	s := []interface{}{}
	if i != nil {
		s = append(s, map[string]interface{}{
			"arn":  *i.Arn,
			"name": *i.Name,
		})
	}
	return s
}

func getInstanceMarketOptions(m *ec2.LaunchTemplateInstanceMarketOptions) []interface{} {
	s := []interface{}{}
	if m != nil {
		spot := []interface{}{}
		so := m.SpotOptions
		if so != nil {
			spot = append(spot, map[string]interface{}{
				"block_duration_minutes":         *so.BlockDurationMinutes,
				"instance_interruption_behavior": *so.InstanceInterruptionBehavior,
				"max_price":                      *so.MaxPrice,
				"spot_instance_type":             *so.SpotInstanceType,
				"valid_until":                    *so.ValidUntil,
			})
		}
		s = append(s, map[string]interface{}{
			"market_type":  *m.MarketType,
			"spot_options": spot,
		})
	}
	return s
}

func getNetworkInterfaces(n []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecification) []interface{} {
	s := []interface{}{}
	for _, v := range n {
		var ipv6Addresses []string
		var ipv4Addresses []string

		networkInterface := map[string]interface{}{
			"associate_public_ip_address": *v.AssociatePublicIpAddress,
			"delete_on_termination":       *v.DeleteOnTermination,
			"description":                 *v.Description,
			"device_index":                int(*v.DeviceIndex),
			"ipv6_address_count":          int(*v.Ipv6AddressCount),
			"network_interface_id":        *v.NetworkInterfaceId,
			"private_ip_address":          *v.PrivateIpAddress,
			"ipv4_address_count":          int(*v.SecondaryPrivateIpAddressCount),
			"subnet_id":                   *v.SubnetId,
		}

		for _, address := range v.Ipv6Addresses {
			ipv6Addresses = append(ipv6Addresses, *address.Ipv6Address)
		}
		networkInterface["ipv6_addresses"] = ipv6Addresses

		for _, address := range v.PrivateIpAddresses {
			ipv4Addresses = append(ipv4Addresses, *address.PrivateIpAddress)
		}
		networkInterface["ipv4_addresses"] = ipv4Addresses

		s = append(s, networkInterface)
	}
	return s
}

func getPlacement(p *ec2.LaunchTemplatePlacement) []interface{} {
	s := []interface{}{}
	if p != nil {
		s = append(s, map[string]interface{}{
			"affinity":          *p.Affinity,
			"availability_zone": *p.AvailabilityZone,
			"group_name":        *p.GroupName,
			"host_id":           *p.HostId,
			"spread_domain":     *p.SpreadDomain,
			"tenancy":           *p.Tenancy,
		})
	}
	return s
}

func getTagSpecifications(t []*ec2.LaunchTemplateTagSpecification) []interface{} {
	s := []interface{}{}
	for _, v := range t {
		s = append(s, map[string]interface{}{
			"resource_type": v.ResourceType,
			"tags":          tagsToMap(v.Tags),
		})
	}
	return s
}

type launchTemplateOpts struct {
	BlockDeviceMappings               []*ec2.LaunchTemplateBlockDeviceMappingRequest
	CreditSpecification               *ec2.CreditSpecificationRequest
	DisableApiTermination             *bool
	EbsOptimized                      *bool
	ElasticGpuSpecifications          []*ec2.ElasticGpuSpecification
	IamInstanceProfile                *ec2.LaunchTemplateIamInstanceProfileSpecificationRequest
	ImageId                           *string
	InstanceInitiatedShutdownBehavior *string
	InstanceMarketOptions             *ec2.LaunchTemplateInstanceMarketOptionsRequest
	InstanceType                      *string
	KernelId                          *string
	KeyName                           *string
	Monitoring                        *ec2.LaunchTemplatesMonitoringRequest
	NetworkInterfaces                 []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest
	Placement                         *ec2.LaunchTemplatePlacementRequest
	RamDiskId                         *string
	SecurityGroupIds                  []*string
	SecurityGroups                    []*string
	TagSpecifications                 []*ec2.LaunchTemplateTagSpecificationRequest
	UserData                          *string
}

func buildLaunchTemplateData(d *schema.ResourceData, meta interface{}) (*launchTemplateOpts, error) {
	opts := &launchTemplateOpts{
		DisableApiTermination: aws.Bool(d.Get("disable_api_termination").(bool)),
		EbsOptimized:          aws.Bool(d.Get("ebs_optimized").(bool)),
		ImageId:               aws.String(d.Get("image_id").(string)),
		InstanceInitiatedShutdownBehavior: aws.String(d.Get("instance_initiated_shutdown_behavior").(string)),
		InstanceType:                      aws.String(d.Get("instance_type").(string)),
		KernelId:                          aws.String(d.Get("kernel_id").(string)),
		KeyName:                           aws.String(d.Get("key_name").(string)),
		RamDiskId:                         aws.String(d.Get("ram_disk_id").(string)),
		UserData:                          aws.String(d.Get("user_data").(string)),
	}

	if v, ok := d.GetOk("block_device_mappings"); ok {
		var blockDeviceMappings []*ec2.LaunchTemplateBlockDeviceMappingRequest
		bdms := v.(*schema.Set).List()

		for _, bdm := range bdms {
			blockDeviceMap := bdm.(map[string]interface{})
			blockDeviceMappings = append(blockDeviceMappings, readBlockDeviceMappingFromConfig(blockDeviceMap))
		}
		opts.BlockDeviceMappings = blockDeviceMappings
	}

	if v, ok := d.GetOk("credit_specification"); ok {
		cs := v.(*schema.Set).List()

		if len(cs) > 0 {
			csData := cs[0].(map[string]interface{})
			csr := &ec2.CreditSpecificationRequest{
				CpuCredits: aws.String(csData["cpu_credits"].(string)),
			}
			opts.CreditSpecification = csr
		}
	}

	if v, ok := d.GetOk("elastic_gpu_specifications"); ok {
		var elasticGpuSpecifications []*ec2.ElasticGpuSpecification
		egsList := v.(*schema.Set).List()

		for _, egs := range egsList {
			elasticGpuSpecification := egs.(map[string]interface{})
			elasticGpuSpecifications = append(elasticGpuSpecifications, &ec2.ElasticGpuSpecification{
				Type: aws.String(elasticGpuSpecification["type"].(string)),
			})
		}
		opts.ElasticGpuSpecifications = elasticGpuSpecifications
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok {
		iip := v.(*schema.Set).List()

		if len(iip) > 0 {
			iipData := iip[0].(map[string]interface{})
			iamInstanceProfile := &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{
				Arn:  aws.String(iipData["arn"].(string)),
				Name: aws.String(iipData["name"].(string)),
			}
			opts.IamInstanceProfile = iamInstanceProfile
		}
	}

	if v, ok := d.GetOk("instance_market_options"); ok {
		imo := v.(*schema.Set).List()

		if len(imo) > 0 {
			imoData := imo[0].(map[string]interface{})
			spotOptions := &ec2.LaunchTemplateSpotMarketOptionsRequest{}

			if v := imoData["spot_options"]; v != nil {
				so := v.(map[string]interface{})
				spotOptions.BlockDurationMinutes = aws.Int64(int64(so["block_duration_minutes"].(int)))
				spotOptions.InstanceInterruptionBehavior = aws.String(so["instance_interruption_behavior"].(string))
				spotOptions.MaxPrice = aws.String(so["max_price"].(string))
				spotOptions.SpotInstanceType = aws.String(so["spot_instance_type"].(string))

				t, err := time.Parse(awsSpotInstanceTimeLayout, so["valid_until"].(string))
				if err != nil {
					return nil, fmt.Errorf("Error Parsing Launch Template Spot Options valid until: %s", err.Error())
				}
				spotOptions.ValidUntil = aws.Time(t)
			}

			instanceMarketOptions := &ec2.LaunchTemplateInstanceMarketOptionsRequest{
				MarketType:  aws.String(imoData["market_type"].(string)),
				SpotOptions: spotOptions,
			}

			opts.InstanceMarketOptions = instanceMarketOptions
		}
	}

	if v, ok := d.GetOk("monitoring"); ok {
		monitoring := &ec2.LaunchTemplatesMonitoringRequest{
			Enabled: aws.Bool(v.(bool)),
		}
		opts.Monitoring = monitoring
	}

	if v, ok := d.GetOk("network_interfaces"); ok {
		var networkInterfaces []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest
		niList := v.(*schema.Set).List()

		for _, ni := range niList {
			var ipv4Addresses []*ec2.PrivateIpAddressSpecification
			var ipv6Addresses []*ec2.InstanceIpv6AddressRequest
			ni := ni.(map[string]interface{})

			privateIpAddress := ni["private_ip_address"].(string)
			networkInterface := &ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
				AssociatePublicIpAddress: aws.Bool(ni["associate_public_ip_address"].(bool)),
				DeleteOnTermination:      aws.Bool(ni["delete_on_termination"].(bool)),
				Description:              aws.String(ni["description"].(string)),
				DeviceIndex:              aws.Int64(int64(ni["device_index"].(int))),
				NetworkInterfaceId:       aws.String(ni["network_interface_id"].(string)),
				PrivateIpAddress:         aws.String(privateIpAddress),
				SubnetId:                 aws.String(ni["subnet_id"].(string)),
			}

			ipv6AddressList := ni["ipv6_addresses"].(*schema.Set).List()
			for _, address := range ipv6AddressList {
				ipv6Addresses = append(ipv6Addresses, &ec2.InstanceIpv6AddressRequest{
					Ipv6Address: aws.String(address.(string)),
				})
			}
			networkInterface.Ipv6AddressCount = aws.Int64(int64(len(ipv6AddressList)))
			networkInterface.Ipv6Addresses = ipv6Addresses

			ipv4AddressList := ni["ipv4_addresses"].(*schema.Set).List()
			for _, address := range ipv4AddressList {
				privateIp := &ec2.PrivateIpAddressSpecification{
					Primary:          aws.Bool(address.(string) == privateIpAddress),
					PrivateIpAddress: aws.String(address.(string)),
				}
				ipv4Addresses = append(ipv4Addresses, privateIp)
			}
			networkInterface.SecondaryPrivateIpAddressCount = aws.Int64(int64(len(ipv4AddressList)))
			networkInterface.PrivateIpAddresses = ipv4Addresses

			networkInterfaces = append(networkInterfaces, networkInterface)
		}
		opts.NetworkInterfaces = networkInterfaces
	}

	if v, ok := d.GetOk("placement"); ok {
		p := v.(*schema.Set).List()

		if len(p) > 0 {
			pData := p[0].(map[string]interface{})
			placement := &ec2.LaunchTemplatePlacementRequest{
				Affinity:         aws.String(pData["affinity"].(string)),
				AvailabilityZone: aws.String(pData["availability_zone"].(string)),
				GroupName:        aws.String(pData["group_name"].(string)),
				HostId:           aws.String(pData["host_id"].(string)),
				SpreadDomain:     aws.String(pData["spread_domain"].(string)),
				Tenancy:          aws.String(pData["tenancy"].(string)),
			}
			opts.Placement = placement
		}
	}

	if v, ok := d.GetOk("tag_specifications"); ok {
		var tagSpecifications []*ec2.LaunchTemplateTagSpecificationRequest
		t := v.(*schema.Set).List()

		for _, ts := range t {
			tsData := ts.(map[string]interface{})
			tags := tagsFromMap(tsData)
			tagSpecification := &ec2.LaunchTemplateTagSpecificationRequest{
				ResourceType: aws.String(tsData["resource_type"].(string)),
				Tags:         tags,
			}
			tagSpecifications = append(tagSpecifications, tagSpecification)
		}
		opts.TagSpecifications = tagSpecifications
	}

	return opts, nil
}

func readBlockDeviceMappingFromConfig(bdm map[string]interface{}) *ec2.LaunchTemplateBlockDeviceMappingRequest {
	blockDeviceMapping := &ec2.LaunchTemplateBlockDeviceMappingRequest{}

	if v := bdm["device_name"]; v != nil {
		blockDeviceMapping.DeviceName = aws.String(v.(string))
	}

	if v := bdm["no_device"]; v != nil {
		blockDeviceMapping.NoDevice = aws.String(v.(string))
	}

	if v := bdm["virtual_name"]; v != nil {
		blockDeviceMapping.VirtualName = aws.String(v.(string))
	}

	if v := bdm["ebs"]; v.(*schema.Set).Len() > 0 {
		ebs := v.(*schema.Set).List()
		if len(ebs) > 0 {
			ebsData := ebs[0]
			//log.Printf("ebsData: %+v\n", ebsData)
			blockDeviceMapping.Ebs = readEbsBlockDeviceFromConfig(ebsData.(map[string]interface{}))
		}
	}

	//log.Printf("block device mapping: %+v\n", *blockDeviceMapping)
	return blockDeviceMapping
}

func readEbsBlockDeviceFromConfig(ebs map[string]interface{}) *ec2.LaunchTemplateEbsBlockDeviceRequest {
	ebsDevice := &ec2.LaunchTemplateEbsBlockDeviceRequest{}

	if v := ebs["delete_on_termination"]; v != nil {
		ebsDevice.DeleteOnTermination = aws.Bool(v.(bool))
	}

	if v := ebs["encrypted"]; v != nil {
		ebsDevice.Encrypted = aws.Bool(v.(bool))
	}

	if v := ebs["iops"]; v != nil {
		ebsDevice.Iops = aws.Int64(int64(v.(int)))
	}

	if v := ebs["kms_key_id"]; v != nil {
		ebsDevice.KmsKeyId = aws.String(v.(string))
	}

	if v := ebs["snapshot_id"]; v != nil {
		ebsDevice.SnapshotId = aws.String(v.(string))
	}

	if v := ebs["volume_size"]; v != nil {
		ebsDevice.VolumeSize = aws.Int64(int64(v.(int)))
	}

	if v := ebs["volume_type"]; v != nil {
		ebsDevice.VolumeType = aws.String(v.(string))
	}

	return ebsDevice
}
