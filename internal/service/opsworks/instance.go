// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opsworks"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opsworks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opsworks_instance", name="Instance")
func resourceInstance() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage:   "This resource is deprecated and will be removed in the next major version of the AWS Provider. Consider the AWS Systems Manager service instead.",
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "x86_64",
				ValidateDiagFunc: enum.Validate[awstypes.Architecture](),
			},

			"auto_scaling_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AutoScalingType](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RootDeviceType](),
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VirtualizationType](),
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
				Set: func(v any) int {
					var buf bytes.Buffer
					m := v.(map[string]any)
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
				Set: func(v any) int {
					var buf bytes.Buffer
					m := v.(map[string]any)
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
				Set: func(v any) int {
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

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	output, err := findInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks instance %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): %s", d.Id(), err)
	}

	d.SetId(aws.ToString(output.InstanceId))
	d.Set("agent_version", output.AgentVersion)
	d.Set("ami_id", output.AmiId)
	d.Set("architecture", output.Architecture)
	d.Set("auto_scaling_type", output.AutoScalingType)
	d.Set(names.AttrAvailabilityZone, output.AvailabilityZone)
	d.Set(names.AttrCreatedAt, output.CreatedAt)
	d.Set("ebs_optimized", output.EbsOptimized)
	d.Set("ec2_instance_id", output.Ec2InstanceId)
	d.Set("ecs_cluster_arn", output.EcsClusterArn)
	d.Set("elastic_ip", output.ElasticIp)
	d.Set("hostname", output.Hostname)
	d.Set("infrastructure_class", output.InfrastructureClass)
	d.Set("install_updates_on_boot", output.InstallUpdatesOnBoot)
	d.Set("instance_profile_arn", output.InstanceProfileArn)
	d.Set(names.AttrInstanceType, output.InstanceType)
	d.Set("last_service_error_id", output.LastServiceErrorId)
	layerIds, err := sortListBasedonTFFile(output.LayerIds, d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "sorting layer_ids attribute: %#v", err)
	}
	if err := d.Set("layer_ids", layerIds); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting layer_ids attribute: %#v, error: %#v", layerIds, err)
	}
	d.Set("os", output.Os)
	d.Set("platform", output.Platform)
	d.Set("private_dns", output.PrivateDns)
	d.Set("private_ip", output.PrivateIp)
	d.Set("public_dns", output.PublicDns)
	d.Set("public_ip", output.PublicIp)
	d.Set("registered_by", output.RegisteredBy)
	d.Set("reported_agent_version", output.ReportedAgentVersion)
	d.Set("reported_os_family", output.ReportedOs.Family)
	d.Set("reported_os_name", output.ReportedOs.Name)
	d.Set("reported_os_version", output.ReportedOs.Version)
	d.Set("root_device_type", output.RootDeviceType)
	d.Set("root_device_volume_id", output.RootDeviceVolumeId)
	d.Set("ssh_host_dsa_key_fingerprint", output.SshHostDsaKeyFingerprint)
	d.Set("ssh_host_rsa_key_fingerprint", output.SshHostRsaKeyFingerprint)
	d.Set("ssh_key_name", output.SshKeyName)
	d.Set("stack_id", output.StackId)
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrSubnetID, output.SubnetId)
	d.Set("tenancy", output.Tenancy)
	d.Set("virtualization_type", output.VirtualizationType)

	// Read BlockDeviceMapping
	ibds := readBlockDevices(output)

	if err := d.Set("ebs_block_device", ibds["ebs"]); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting ebs_block_device: %s", d.Id(), err)
	}
	if err := d.Set("ephemeral_block_device", ibds["ephemeral"]); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting ephemeral_block_device: %s", d.Id(), err)
	}
	if ibds["root"] != nil {
		if err := d.Set("root_block_device", []any{ibds["root"]}); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting root_block_device: %s", d.Id(), err)
		}
	} else {
		d.Set("root_block_device", []any{})
	}

	// Read Security Groups
	sgs := make([]string, 0, len(output.SecurityGroupIds))
	sgs = append(sgs, output.SecurityGroupIds...)
	if err := d.Set(names.AttrSecurityGroupIDs, sgs); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance (%s): setting security_group_ids: %s", d.Id(), err)
	}
	return diags
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	err := resourceInstanceValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Instance: %s", err)
	}

	req := &opsworks.CreateInstanceInput{
		AgentVersion:         aws.String(d.Get("agent_version").(string)),
		Architecture:         awstypes.Architecture(d.Get("architecture").(string)),
		EbsOptimized:         aws.Bool(d.Get("ebs_optimized").(bool)),
		InstallUpdatesOnBoot: aws.Bool(d.Get("install_updates_on_boot").(bool)),
		InstanceType:         aws.String(d.Get(names.AttrInstanceType).(string)),
		LayerIds:             flex.ExpandStringValueList(d.Get("layer_ids").([]any)),
		StackId:              aws.String(d.Get("stack_id").(string)),
	}

	if v, ok := d.GetOk("ami_id"); ok {
		req.AmiId = aws.String(v.(string))
		req.Os = aws.String("Custom")
	}

	if v, ok := d.GetOk("auto_scaling_type"); ok {
		req.AutoScalingType = awstypes.AutoScalingType(v.(string))
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
		req.RootDeviceType = awstypes.RootDeviceType(v.(string))
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

	var blockDevices []awstypes.BlockDeviceMapping

	if v, ok := d.GetOk("ebs_block_device"); ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]any)
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrSnapshotID].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int32(int32(v))
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
			bd := v.(map[string]any)
			blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
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
			bd := v.(map[string]any)
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int32(int32(v))
			}

			blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
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

	resp, err = conn.CreateInstance(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: %s", err)
	}

	if resp.InstanceId == nil {
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: no instance returned")
	}

	instanceId := aws.ToString(resp.InstanceId)
	d.SetId(instanceId)

	if v, ok := d.GetOk(names.AttrState); ok && v.(string) == instanceStatusRunning {
		err := startInstance(ctx, d, meta, true, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating OpsWorks Instance: %s", err)
		}
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	err := resourceInstanceValidate(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating OpsWorks Instance (%s): %s", d.Id(), err)
	}

	req := &opsworks.UpdateInstanceInput{
		InstanceId:           aws.String(d.Id()),
		AgentVersion:         aws.String(d.Get("agent_version").(string)),
		Architecture:         awstypes.Architecture(d.Get("architecture").(string)),
		InstallUpdatesOnBoot: aws.Bool(d.Get("install_updates_on_boot").(bool)),
	}

	if v, ok := d.GetOk("ami_id"); ok {
		req.AmiId = aws.String(v.(string))
		req.Os = aws.String("Custom")
	}

	if v, ok := d.GetOk("auto_scaling_type"); ok {
		req.AutoScalingType = awstypes.AutoScalingType(v.(string))
	}

	if v, ok := d.GetOk("hostname"); ok {
		req.Hostname = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		req.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layer_ids"); ok {
		req.LayerIds = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := d.GetOk("os"); ok {
		req.Os = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ssh_key_name"); ok {
		req.SshKeyName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating OpsWorks instance: %s", d.Id())

	_, err = conn.UpdateInstance(ctx, req)
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

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

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

	_, err := conn.DeleteInstance(ctx, req)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findInstanceByID(ctx context.Context, conn *opsworks.Client, id string) (*awstypes.Instance, error) {
	input := &opsworks.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	output, err := conn.DescribeInstances(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.Instances == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.Instances)
}

func resourceInstanceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	// Neither delete_eip nor delete_ebs can be fetched
	// from any API call, so we need to default to the values
	// we set in the schema by default
	d.Set("delete_ebs", true)
	d.Set("delete_eip", true)
	return []*schema.ResourceData{d}, nil
}

func startInstance(ctx context.Context, d *schema.ResourceData, meta any, wait bool, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	req := &opsworks.StartInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Starting OpsWorks instance: %s", d.Id())

	_, err := conn.StartInstance(ctx, req)

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

func stopInstance(ctx context.Context, d *schema.ResourceData, meta any, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).OpsWorksClient(ctx)

	req := &opsworks.StopInstanceInput{
		InstanceId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Stopping OpsWorks instance: %s", d.Id())

	_, err := conn.StopInstance(ctx, req)

	if err != nil {
		return fmt.Errorf("stopping instance: %w", err)
	}

	log.Printf("[DEBUG] Waiting for OpsWorks instance (%s) to become stopped", d.Id())

	if err := waitInstanceStopped(ctx, conn, d.Id(), timeout); err != nil {
		return fmt.Errorf("stopping instance: waiting for completion: %w", err)
	}

	return nil
}

func waitInstanceDeleted(ctx context.Context, conn *opsworks.Client, instanceId string) error {
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

func waitInstanceStarted(ctx context.Context, conn *opsworks.Client, instanceId string, timeout time.Duration) error {
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

func waitInstanceStopped(ctx context.Context, conn *opsworks.Client, instanceId string, timeout time.Duration) error {
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

func instanceStatus(ctx context.Context, conn *opsworks.Client, instanceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := conn.DescribeInstances(ctx, &opsworks.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if resp == nil || len(resp.Instances) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Instances[0]
		return i, aws.ToString(i.Status), nil
	}
}

func readBlockDevices(instance *awstypes.Instance) map[string]any {
	blockDevices := make(map[string]any)
	blockDevices["ebs"] = make([]map[string]any, 0)
	blockDevices["ephemeral"] = make([]map[string]any, 0)
	blockDevices["root"] = nil

	if len(instance.BlockDeviceMappings) == 0 {
		return nil
	}

	for _, bdm := range instance.BlockDeviceMappings {
		bd := make(map[string]any)
		if bdm.Ebs != nil && bdm.Ebs.DeleteOnTermination != nil {
			bd[names.AttrDeleteOnTermination] = aws.ToBool(bdm.Ebs.DeleteOnTermination)
		}
		if bdm.Ebs != nil && bdm.Ebs.VolumeSize != nil {
			bd[names.AttrVolumeSize] = aws.ToInt32(bdm.Ebs.VolumeSize)
		}
		if bdm.Ebs != nil {
			bd[names.AttrVolumeType] = bdm.Ebs.VolumeType
		}
		if bdm.Ebs != nil && bdm.Ebs.Iops != nil {
			bd[names.AttrIOPS] = aws.ToInt32(bdm.Ebs.Iops)
		}
		if aws.ToString(bdm.DeviceName) == "ROOT_DEVICE" {
			blockDevices["root"] = bd
		} else {
			if bdm.DeviceName != nil {
				bd[names.AttrDeviceName] = aws.ToString(bdm.DeviceName)
			}
			if bdm.VirtualName != nil {
				bd[names.AttrVirtualName] = aws.ToString(bdm.VirtualName)
				blockDevices["ephemeral"] = append(blockDevices["ephemeral"].([]map[string]any), bd)
			} else {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					bd[names.AttrSnapshotID] = aws.ToString(bdm.Ebs.SnapshotId)
				}
				blockDevices["ebs"] = append(blockDevices["ebs"].([]map[string]any), bd)
			}
		}
	}
	return blockDevices
}
