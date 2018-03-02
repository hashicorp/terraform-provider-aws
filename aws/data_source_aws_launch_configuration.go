package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsLaunchConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLaunchConfigurationRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"vpc_classic_link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_classic_link_security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"spot_price": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ebs_optimized": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"placement_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"enable_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"ebs_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"root_block_device": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsLaunchConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*AWSClient).autoscalingconn
	ec2conn := meta.(*AWSClient).ec2conn

	if v, ok := d.GetOk("name"); ok {
		d.SetId(v.(string))
	}

	describeOpts := autoscaling.DescribeLaunchConfigurationsInput{
		LaunchConfigurationNames: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] launch configuration describe configuration: %s", describeOpts)
	describConfs, err := autoscalingconn.DescribeLaunchConfigurations(&describeOpts)
	if err != nil {
		return fmt.Errorf("Error retrieving launch configuration: %s", err)
	}

	lc := describConfs.LaunchConfigurations[0]

	d.Set("key_name", lc.KeyName)
	d.Set("image_id", lc.ImageId)
	d.Set("instance_type", lc.InstanceType)
	d.Set("name", lc.LaunchConfigurationName)
	d.Set("user_data", lc.UserData)

	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("ebs_optimized", lc.EbsOptimized)
	d.Set("spot_price", lc.SpotPrice)
	d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	d.Set("security_groups", lc.SecurityGroups)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)

	d.Set("vpc_classic_link_id", lc.ClassicLinkVPCId)
	d.Set("vpc_classic_link_security_groups", lc.ClassicLinkVPCSecurityGroups)

	if err := readLCBlockDevices(d, lc, ec2conn); err != nil {
		return err
	}

	return nil
}
