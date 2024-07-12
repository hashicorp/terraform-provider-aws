// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opsworks_instance")
func ResourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstanceUpdate,
		DeleteWithoutTimeout: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceInstanceImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"agent_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "INHERIT",
			},

			"ami_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"architecture": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "x86_64",
				ValidateFunc: validation.StringInSlice(opsworks.Architecture_Values(), false),
			},

			"auto_scaling_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(opsworks.AutoScalingType_Values(), false),
			},

			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"delete_ebs": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"delete_eip": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"ec2_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"elastic_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"infrastructure_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"install_updates_on_boot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"instance_profile_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Optional: true,
			},

			"last_service_error_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"layer_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"os": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"platform": {
				Type:     schema.TypeString,
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

			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"registered_by": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reported_agent_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reported_os_family": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reported_os_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reported_os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"root_device_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(opsworks.RootDeviceType_Values(), false),
			},

			"root_device_volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ssh_host_dsa_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ssh_host_rsa_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ssh_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"stack_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"running",
					"stopped",
				}, false),
			},

			names.AttrStatus: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"tenancy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"dedicated",
					"default",
					"host",
				}, false),
			},

			"virtualization_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(opsworks.VirtualizationType_Values(), false),
			},

			"ebs_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

						names.AttrIOPS: {
							Type:     schema.TypeInt,
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

						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
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

						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrVirtualName].(string)))
					return create.StringHashcode(buf.String())
				},
			},

			"root_block_device": {
				// TODO: This is a set because we don't support singleton
				//       sub-resources today. We'll enforce that the set only ever has
				//       length zero or one below. When TF gains support for
				//       sub-resources this can be converted.
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					// "You can only modify the volume size, volume type, and Delete on
					// Termination flag on the block device mapping entry for the root
					// device volume." - bit.ly/ec2bdmap
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},

						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
				Set: func(v interface{}) int {
					// there can be only one root device; no need to hash anything
					return 0
				},
			},
		},
	}
}

func resourceInstanceValidate(d *schema.ResourceData) error {
	if d.HasChange("ami_id") {
		if v, ok := d.GetOk("os"); ok {
			if v.(string) != "Custom" {
				return fmt.Errorf("OS must be \"Custom\" when using using a custom ami_id")
			}
		}

		if _, ok := d.GetOk("root_block_device"); ok {
			return fmt.Errorf("Cannot specify root_block_device when using a custom ami_id.")
		}

		if _, ok := d.GetOk("ebs_block_device"); ok {
			return fmt.Errorf("Cannot specify ebs_block_device when using a custom ami_id.")
		}

		if _, ok := d.GetOk("ephemeral_block_device"); ok {
			return fmt.Errorf("Cannot specify ephemeral_block_device when using a custom ami_id.")
		}
	}
	return nil
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	req := &opsworks.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks instance: %s", d.Id())

	resp, err := conn.DescribeInstancesWithContext(ctx, req)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] OpsWorks instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): %s", d.Id(), err)
	}

	// If nothing was found, then return no state
	if len(resp.Instances) == 0 || resp.Instances[0] == nil || resp.Instances[0].InstanceId == nil {
		log.Printf("[WARN] OpsWorks instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	instance := resp.Instances[0]

	d.SetId(aws.StringValue(instance.InstanceId))
	d.Set("agent_version", instance.AgentVersion)
	d.Set("ami_id", instance.AmiId)
	d.Set("architecture", instance.Architecture)
	d.Set("auto_scaling_type", instance.AutoScalingType)
	d.Set(names.AttrAvailabilityZone, instance.AvailabilityZone)
	d.Set(names.AttrCreatedAt, instance.CreatedAt)
	d.Set("ebs_optimized", instance.EbsOptimized)
	d.Set("ec2_instance_id", instance.Ec2InstanceId)
	d.Set("ecs_cluster_arn", instance.EcsClusterArn)
	d.Set("elastic_ip", instance.ElasticIp)
	d.Set("hostname", instance.Hostname)
	d.Set("infrastructure_class", instance.InfrastructureClass)
	d.Set("install_updates_on_boot", instance.InstallUpdatesOnBoot)
	d.Set("instance_profile_arn", instance.InstanceProfileArn)
	d.Set(names.AttrInstanceType, instance.InstanceType)
	d.Set("last_service_error_id", instance.LastServiceErrorId)
	var layerIds []string
	for _, v := range instance.LayerIds {
		layerIds = append(layerIds, aws.StringValue(v))
	}
	layerIds, err = sortListBasedonTFFile(layerIds, d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "sorting layer_ids attribute: %#v", err)
	}
	if err := d.Set("layer_ids", layerIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting layer_ids attribute: %#v, error: %#v", layerIds, err)
	}
	d.Set("os", instance.Os)
	d.Set("platform", instance.Platform)
	d.Set("private_dns", instance.PrivateDns)
	d.Set("private_ip", instance.PrivateIp)
	d.Set("public_dns", instance.PublicDns)
	d.Set("public_ip", instance.PublicIp)
	d.Set("registered_by", instance.RegisteredBy)
	d.Set("reported_agent_version", instance.ReportedAgentVersion)
	d.Set("reported_os_family", instance.ReportedOs.Family)
	d.Set("reported_os_name", instance.ReportedOs.Name)
	d.Set("reported_os_version", instance.ReportedOs.Version)
	d.Set("root_device_type", instance.RootDeviceType)
	d.Set("root_device_volume_id", instance.RootDeviceVolumeId)
	d.Set("ssh_host_dsa_key_fingerprint", instance.SshHostDsaKeyFingerprint)
	d.Set("ssh_host_rsa_key_fingerprint", instance.SshHostRsaKeyFingerprint)
	d.Set("ssh_key_name", instance.SshKeyName)
	d.Set("stack_id", instance.StackId)
	d.Set(names.AttrStatus, instance.Status)
	d.Set(names.AttrSubnetID, instance.SubnetId)
	d.Set("tenancy", instance.Tenancy)
	d.Set("virtualization_type", instance.VirtualizationType)

	// Read BlockDeviceMapping
	ibds := readBlockDevices(instance)

	if err := d.Set("ebs_block_device", ibds["ebs"]); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting ebs_block_device: %s", d.Id(), err)
	}
	if err := d.Set("ephemeral_block_device", ibds["ephemeral"]); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting ephemeral_block_device: %s", d.Id(), err)
	}
	if ibds["root"] != nil {
		if err := d.Set("root_block_device", []interface{}{ibds["root"]}); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting root_block_device: %s", d.Id(), err)
		}
	} else {
		d.Set("root_block_device", []interface{}{})
	}

	// Read Security Groups
	sgs := make([]string, 0, len(instance.SecurityGroupIds))
	for _, sg := range instance.SecurityGroupIds {
		sgs = append(sgs, *sg)
	}
	if err := d.Set(names.AttrSecurityGroupIDs, sgs); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting security_group_ids: %s", d.Id(), err)
	}
	return diags
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	err := resourceInstanceValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance: %s", err)
	}

	req := &opsworks.CreateInstanceInput{
		AgentVersion:         aws.String(d.Get("agent_version").(string)),
		Architecture:         aws.String(d.Get("architecture").(string)),
		EbsOptimized:         aws.Bool(d.Get("ebs_optimized").(bool)),
		InstallUpdatesOnBoot: aws.Bool(d.Get("install_updates_on_boot").(bool)),
		InstanceType:         aws.String(d.Get(names.AttrInstanceType).(string)),
		LayerIds:             flex.ExpandStringList(d.Get("layer_ids").([]interface{})),
		StackId:              aws.String(d.Get("stack_id").(string)),
	}

	if v, ok := d.GetOk("ami_id"); ok {
		req.AmiId = aws.String(v.(string))
		req.Os = aws.String("Custom")
	}

	if v, ok := d.GetOk("auto_scaling_type"); ok {
		req.AutoScalingType = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		req.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hostname"); ok {
		req.Hostname = aws.String(v.(string))
	}

	if v, ok := d.GetOk("os"); ok {
		req.Os = aws.String(v.(string))
	}

	if v, ok := d.GetOk("root_device_type"); ok {
		req.RootDeviceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssh_key_name"); ok {
		req.SshKeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		req.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		req.Tenancy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("virtualization_type"); ok {
		req.VirtualizationType = aws.String(v.(string))
	}

	var blockDevices []*opsworks.BlockDeviceMapping

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &opsworks.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrSnapshotID].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			blockDevices = append(blockDevices, &opsworks.BlockDeviceMapping{
				DeviceName: aws.String(bd[names.AttrDeviceName].(string)),
				Ebs:        ebs,
			})
		}
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			blockDevices = append(blockDevices, &opsworks.BlockDeviceMapping{
				DeviceName:  aws.String(bd[names.AttrDeviceName].(string)),
				VirtualName: aws.String(bd[names.AttrVirtualName].(string)),
			})
		}
	}

	if v, ok := d.GetOk("root_block_device"); ok {
		vL := v.(*schema.Set).List()
		if len(vL) > 1 {
			return sdkdiag.AppendErrorf(diags, "Cannot specify more than one root_block_device.")
		}
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &opsworks.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			blockDevices = append(blockDevices, &opsworks.BlockDeviceMapping{
				DeviceName: aws.String("ROOT_DEVICE"),
				Ebs:        ebs,
			})
		}
	}

	if len(blockDevices) > 0 {
		req.BlockDeviceMappings = blockDevices
	}

	log.Printf("[DEBUG] Creating OpsWorks instance")

	var resp *opsworks.CreateInstanceOutput

	resp, err = conn.CreateInstanceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: %s", err)
	}

	if resp.InstanceId == nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: no instance returned")
	}

	instanceId := aws.StringValue(resp.InstanceId)
	d.SetId(instanceId)

	if v, ok := d.GetOk(names.AttrState); ok && v.(string) == instanceStatusRunning {
		err := startInstance(ctx, d, meta, true, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: %s", err)
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	err := resourceInstanceValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Instance (%s): %s", d.Id(), err)
	}

	req := &opsworks.UpdateInstanceInput{
		InstanceId:           aws.String(d.Id()),
		AgentVersion:         aws.String(d.Get("agent_version").(string)),
		Architecture:         aws.String(d.Get("architecture").(string)),
		InstallUpdatesOnBoot: aws.Bool(d.Get("install_updates_on_boot").(bool)),
	}

	if v, ok := d.GetOk("ami_id"); ok {
		req.AmiId = aws.String(v.(string))
		req.Os = aws.String("Custom")
	}

	if v, ok := d.GetOk("auto_scaling_type"); ok {
		req.AutoScalingType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hostname"); ok {
		req.Hostname = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		req.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layer_ids"); ok {
		req.LayerIds = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("os"); ok {
		req.Os = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssh_key_name"); ok {
		req.SshKeyName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating OpsWorks instance: %s", d.Id())

	_, err = conn.UpdateInstanceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Instance (%s): %s", d.Id(), err)
	}

	var status string

	if v, ok := d.GetOk(names.AttrStatus); ok {
		status = v.(string)
	} else {
		status = "stopped"
	}

	if v, ok := d.GetOk(names.AttrState); ok {
		state := v.(string)
		if state == instanceStatusRunning {
			if status == instanceStatusStopped || status == instanceStatusStopping || status == instanceStatusShuttingDown {
				err := startInstance(ctx, d, meta, false, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating OpsWorks Instance (%s): %s", d.Id(), err)
				}
			}
		} else {
			if status != instanceStatusStopped && status != instanceStatusStopping && status != instanceStatusShuttingDown {
				err := stopInstance(ctx, d, meta, d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating OpsWorks Instance (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	if v, ok := d.GetOk(names.AttrStatus); ok && v.(string) != instanceStatusStopped {
		err := stopInstance(ctx, d, meta, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting OpsWorks instance (%s): %s", d.Id(), err)
		}
	}

	req := &opsworks.DeleteInstanceInput{
		InstanceId:      aws.String(d.Id()),
		DeleteElasticIp: aws.Bool(d.Get("delete_eip").(bool)),
		DeleteVolumes:   aws.Bool(d.Get("delete_ebs").(bool)),
	}

	log.Printf("[DEBUG] Deleting OpsWorks instance: %s", d.Id())

	_, err := conn.DeleteInstanceWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks instance (%s): %s", d.Id(), err)
	}

	if err := waitInstanceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWorks instance (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func resourceInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither delete_eip nor delete_ebs can be fetched
	// from any API call, so we need to default to the values
	// we set in the schema by default
	d.Set("delete_ebs", true)
	d.Set("delete_eip", true)
	return []*schema.ResourceData{d}, nil
}

func startInstance(ctx context.Context, d *schema.ResourceData, meta interface{}, wait bool, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	req := &opsworks.StartInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Starting OpsWorks instance: %s", d.Id())

	_, err := conn.StartInstanceWithContext(ctx, req)

	if err != nil {
		return fmt.Errorf("starting instance: %w", err)
	}

	if wait {
		log.Printf("[DEBUG] Waiting for OpsWorks instance (%s) to start", d.Id())

		if err := waitInstanceStarted(ctx, conn, d.Id(), timeout); err != nil {
			return fmt.Errorf("starting instance: waiting for completion: %w", err)
		}
	}

	return nil
}

func stopInstance(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	req := &opsworks.StopInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Stopping OpsWorks instance: %s", d.Id())

	_, err := conn.StopInstanceWithContext(ctx, req)

	if err != nil {
		return fmt.Errorf("stopping instance: %w", err)
	}

	log.Printf("[DEBUG] Waiting for OpsWorks instance (%s) to become stopped", d.Id())

	if err := waitInstanceStopped(ctx, conn, d.Id(), timeout); err != nil {
		return fmt.Errorf("stopping instance: waiting for completion: %w", err)
	}

	return nil
}

func waitInstanceDeleted(ctx context.Context, conn *opsworks.OpsWorks, instanceId string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{instanceStatusStopped, instanceStatusTerminating, instanceStatusTerminated},
		Target:     []string{},
		Refresh:    instanceStatus(ctx, conn, instanceId),
		Timeout:    2 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitInstanceStarted(ctx context.Context, conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{instanceStatusRequested, instanceStatusPending, instanceStatusBooting, instanceStatusRunningSetup},
		Target:     []string{instanceStatusOnline},
		Refresh:    instanceStatus(ctx, conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitInstanceStopped(ctx context.Context, conn *opsworks.OpsWorks, instanceId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{instanceStatusStopping, instanceStatusTerminating, instanceStatusShuttingDown, instanceStatusTerminated},
		Target:     []string{instanceStatusStopped},
		Refresh:    instanceStatus(ctx, conn, instanceId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func instanceStatus(ctx context.Context, conn *opsworks.OpsWorks, instanceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInstancesWithContext(ctx, &opsworks.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(instanceID)},
		})

		if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.Instances) == 0 || resp.Instances[0] == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Instances[0]
		return i, aws.StringValue(i.Status), nil
	}
}

func readBlockDevices(instance *opsworks.Instance) map[string]interface{} {
	blockDevices := make(map[string]interface{})
	blockDevices["ebs"] = make([]map[string]interface{}, 0)
	blockDevices["ephemeral"] = make([]map[string]interface{}, 0)
	blockDevices["root"] = nil

	if len(instance.BlockDeviceMappings) == 0 {
		return nil
	}

	for _, bdm := range instance.BlockDeviceMappings {
		bd := make(map[string]interface{})
		if bdm.Ebs != nil && bdm.Ebs.DeleteOnTermination != nil {
			bd[names.AttrDeleteOnTermination] = aws.BoolValue(bdm.Ebs.DeleteOnTermination)
		}
		if bdm.Ebs != nil && bdm.Ebs.VolumeSize != nil {
			bd[names.AttrVolumeSize] = aws.Int64Value(bdm.Ebs.VolumeSize)
		}
		if bdm.Ebs != nil && bdm.Ebs.VolumeType != nil {
			bd[names.AttrVolumeType] = aws.StringValue(bdm.Ebs.VolumeType)
		}
		if bdm.Ebs != nil && bdm.Ebs.Iops != nil {
			bd[names.AttrIOPS] = aws.Int64Value(bdm.Ebs.Iops)
		}
		if aws.StringValue(bdm.DeviceName) == "ROOT_DEVICE" {
			blockDevices["root"] = bd
		} else {
			if bdm.DeviceName != nil {
				bd[names.AttrDeviceName] = aws.StringValue(bdm.DeviceName)
			}
			if bdm.VirtualName != nil {
				bd[names.AttrVirtualName] = aws.StringValue(bdm.VirtualName)
				blockDevices["ephemeral"] = append(blockDevices["ephemeral"].([]map[string]interface{}), bd)
			} else {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					bd[names.AttrSnapshotID] = aws.StringValue(bdm.Ebs.SnapshotId)
				}
				blockDevices["ebs"] = append(blockDevices["ebs"].([]map[string]interface{}), bd)
			}
		}
	}
	return blockDevices
}
