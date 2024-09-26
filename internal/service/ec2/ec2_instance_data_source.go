// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_instance", name="Instance")
// @Tags
func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"disable_api_stop": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSnapshotID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
						names.AttrThroughput: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				// This should not be necessary, but currently is (see #7198)
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
				Computed: true,
			},
			"enclave_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"ephemeral_block_device": {
				Type:     schema.TypeList,
				Computed: true,
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
			},
			names.AttrFilter: customFiltersSchema(),
			"get_password_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"get_user_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"host_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_resource_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_tags": tftags.TagsSchemaComputed(),
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintenance_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_protocol_ipv6": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_metadata_tags": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_partition_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_resource_name_dns_aaaa_record": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_resource_name_dns_a_record": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"hostname_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
						names.AttrThroughput: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"secondary_private_ips": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"source_dest_check": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_data_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// dataSourceInstanceRead performs the instanceID lookup
func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Build up search parameters
	input := &ec2.DescribeInstancesInput{}

	if tags, tagsOk := d.GetOk("instance_tags"); tagsOk {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	if v, ok := d.GetOk(names.AttrInstanceID); ok {
		input.InstanceIds = []string{v.(string)}
	}

	instance, err := findInstance(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Instance", err))
	}

	log.Printf("[DEBUG] aws_instance - Single Instance ID found: %s", aws.ToString(instance.InstanceId))
	if err := instanceDescriptionAttributes(ctx, d, meta, instance); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", aws.ToString(instance.InstanceId), err)
	}

	if d.Get("get_password_data").(bool) {
		passwordData, err := getInstancePasswordData(ctx, aws.ToString(instance.InstanceId), conn, d.Timeout(schema.TimeoutRead))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Instance (%s): %s", aws.ToString(instance.InstanceId), err)
		}
		d.Set("password_data", passwordData)
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

	return diags
}

// Populate instance attribute fields with the returned instance
func instanceDescriptionAttributes(ctx context.Context, d *schema.ResourceData, meta interface{}, instance *awstypes.Instance) error {
	d.SetId(aws.ToString(instance.InstanceId))
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instanceType := string(instance.InstanceType)
	instanceTypeInfo, err := findInstanceTypeByName(ctx, conn, instanceType)

	if err != nil {
		return fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
	}

	// Set the easy attributes
	d.Set("instance_state", instance.State.Name)
	d.Set(names.AttrAvailabilityZone, instance.Placement.AvailabilityZone)
	d.Set("placement_group", instance.Placement.GroupName)
	d.Set("placement_partition_number", instance.Placement.PartitionNumber)
	d.Set("tenancy", instance.Placement.Tenancy)
	d.Set("host_id", instance.Placement.HostId)
	d.Set("host_resource_group_arn", instance.Placement.HostResourceGroupArn)

	d.Set("ami", instance.ImageId)
	d.Set(names.AttrInstanceType, instanceType)
	d.Set("key_name", instance.KeyName)
	d.Set("launch_time", instance.LaunchTime.Format(time.RFC3339))
	d.Set("outpost_arn", instance.OutpostArn)
	d.Set("private_dns", instance.PrivateDnsName)
	d.Set("private_ip", instance.PrivateIpAddress)
	d.Set("public_dns", instance.PublicDnsName)
	d.Set("public_ip", instance.PublicIpAddress)

	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		name, err := instanceProfileARNToName(aws.ToString(instance.IamInstanceProfile.Arn))

		if err != nil {
			return fmt.Errorf("setting iam_instance_profile: %w", err)
		}

		d.Set("iam_instance_profile", name)
	} else {
		d.Set("iam_instance_profile", nil)
	}

	// iterate through network interfaces, and set subnet, network_interface, public_addr
	if len(instance.NetworkInterfaces) > 0 {
		for _, ni := range instance.NetworkInterfaces {
			if aws.ToInt32(ni.Attachment.DeviceIndex) == 0 {
				d.Set(names.AttrSubnetID, ni.SubnetId)
				d.Set(names.AttrNetworkInterfaceID, ni.NetworkInterfaceId)
				d.Set("associate_public_ip_address", ni.Association != nil)

				secondaryIPs := make([]string, 0, len(ni.PrivateIpAddresses))
				for _, ip := range ni.PrivateIpAddresses {
					if !aws.ToBool(ip.Primary) {
						secondaryIPs = append(secondaryIPs, aws.ToString(ip.PrivateIpAddress))
					}
				}
				if err := d.Set("secondary_private_ips", secondaryIPs); err != nil {
					return fmt.Errorf("setting secondary_private_ips: %w", err)
				}

				ipV6Addresses := make([]string, 0, len(ni.Ipv6Addresses))
				for _, ip := range ni.Ipv6Addresses {
					ipV6Addresses = append(ipV6Addresses, aws.ToString(ip.Ipv6Address))
				}
				if err := d.Set("ipv6_addresses", ipV6Addresses); err != nil {
					return fmt.Errorf("setting ipv6_addresses: %w", err)
				}
			}
		}
	} else {
		d.Set(names.AttrSubnetID, instance.SubnetId)
		d.Set(names.AttrNetworkInterfaceID, "")
	}

	d.Set("ebs_optimized", instance.EbsOptimized)
	if aws.ToString(instance.SubnetId) != "" {
		d.Set("source_dest_check", instance.SourceDestCheck)
	}

	if instance.Monitoring != nil {
		monitoringState := string(instance.Monitoring.State)
		d.Set("monitoring", monitoringState == names.AttrEnabled || monitoringState == "pending")
	}

	setTagsOut(ctx, instance.Tags)

	// Security Groups
	if err := readSecurityGroups(ctx, d, instance, conn); err != nil {
		return fmt.Errorf("reading EC2 Instance (%s): %w", aws.ToString(instance.InstanceId), err)
	}

	// Block devices
	if err := readBlockDevices(ctx, d, meta, instance, true); err != nil {
		return fmt.Errorf("reading EC2 Instance (%s): %w", aws.ToString(instance.InstanceId), err)
	}
	if _, ok := d.GetOk("ephemeral_block_device"); !ok {
		d.Set("ephemeral_block_device", []interface{}{})
	}

	// Lookup and Set Instance Attributes
	{
		attr, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
			Attribute:  awstypes.InstanceAttributeNameDisableApiStop,
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return fmt.Errorf("getting attribute (%s): %w", awstypes.InstanceAttributeNameDisableApiStop, err)
		}
		d.Set("disable_api_stop", attr.DisableApiStop.Value)
	}
	{
		attr, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
			Attribute:  awstypes.InstanceAttributeNameDisableApiTermination,
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return fmt.Errorf("getting attribute (%s): %w", awstypes.InstanceAttributeNameDisableApiTermination, err)
		}
		d.Set("disable_api_termination", attr.DisableApiTermination.Value)
	}
	{
		attr, err := conn.DescribeInstanceAttribute(ctx, &ec2.DescribeInstanceAttributeInput{
			Attribute:  awstypes.InstanceAttributeNameUserData,
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return fmt.Errorf("getting attribute (%s): %w", awstypes.InstanceAttributeNameUserData, err)
		}
		if attr != nil && attr.UserData != nil && attr.UserData.Value != nil {
			d.Set("user_data", userDataHashSum(aws.ToString(attr.UserData.Value)))
			if d.Get("get_user_data").(bool) {
				d.Set("user_data_base64", attr.UserData.Value)
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
			return fmt.Errorf("reading EC2 Instance (%s) credit specification: %w", d.Id(), err)
		}

		if instanceCreditSpecification != nil {
			if err := d.Set("credit_specification", []interface{}{flattenInstanceCreditSpecification(instanceCreditSpecification)}); err != nil {
				return fmt.Errorf("setting credit_specification: %w", err)
			}
		} else {
			d.Set("credit_specification", nil)
		}
	} else {
		d.Set("credit_specification", nil)
	}

	if err := d.Set("enclave_options", flattenEnclaveOptions(instance.EnclaveOptions)); err != nil {
		return fmt.Errorf("setting enclave_options: %w", err)
	}

	if instance.MaintenanceOptions != nil {
		if err := d.Set("maintenance_options", []interface{}{flattenInstanceMaintenanceOptions(instance.MaintenanceOptions)}); err != nil {
			return fmt.Errorf("setting maintenance_options: %w", err)
		}
	} else {
		d.Set("maintenance_options", nil)
	}

	if err := d.Set("metadata_options", flattenInstanceMetadataOptions(instance.MetadataOptions)); err != nil {
		return fmt.Errorf("setting metadata_options: %w", err)
	}

	if instance.PrivateDnsNameOptions != nil {
		if err := d.Set("private_dns_name_options", []interface{}{flattenPrivateDNSNameOptionsResponse(instance.PrivateDnsNameOptions)}); err != nil {
			return fmt.Errorf("setting private_dns_name_options: %w", err)
		}
	} else {
		d.Set("private_dns_name_options", nil)
	}

	return nil
}
