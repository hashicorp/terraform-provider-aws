// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"fmt"
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	"github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_m2_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func ResourceEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentCreate,
		ReadWithoutTimeout:   resourceEnvironmentRead,
		UpdateWithoutTimeout: resourceEnvironmentUpdate,
		DeleteWithoutTimeout: resourceEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"apply_changes_during_maintenance_window": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Apply the changes during maintenance window",
			},
			"client_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"efs_mount": {
				Type:          schema.TypeSet,
				Optional:      true,
				MinItems:      0,
				MaxItems:      1,
				ForceNew:      true,
				ConflictsWith: []string{"fsx_mount"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_system_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mount_point": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile("^/m2/mount(/[\\w-]+)+$"), "Mount point must start /m2/mount"),
						},
					},
				},
			},
			"engine_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(enum.Values[types.EngineType](), false),
				ForceNew:     true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fsx_mount": {
				Type:          schema.TypeSet,
				Optional:      true,
				ForceNew:      true,
				MinItems:      0,
				MaxItems:      1,
				ConflictsWith: []string{"efs_mount"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_system_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mount_point": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"high_availability_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_capacity": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},
					},
				},
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"load_balancer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"preferred_maintenance_window": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Preferred maintenance window, if not provided a random one will be generated.",
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				MinItems: 1,
				ForceNew: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				MinItems: 1,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNameEnvironment = "Environment"
)

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).M2Client(ctx)

	var clientToken string
	if v, ok := d.GetOk("client_token"); ok {
		clientToken = v.(string)
	} else {
		clientToken = id.UniqueId()
	}

	in := &m2.CreateEnvironmentInput{
		EngineType:       types.EngineType(d.Get("engine_type").(string)),
		InstanceType:     aws.String(d.Get("instance_type").(string)),
		Name:             aws.String(d.Get("name").(string)),
		ClientToken:      aws.String(clientToken),
		Description:      aws.String(d.Get("description").(string)),
		SecurityGroupIds: flex.ExpandStringValueSet(d.Get("security_groups").(*schema.Set)),
		SubnetIds:        flex.ExpandStringValueSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("engine_version"); ok {
		in.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("high_availability_config"); ok && len(v.([]interface{})) > 0 {
		in.HighAvailabilityConfig = expandHaConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		in.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		in.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		in.PubliclyAccessible = v.(bool)
	}

	in.StorageConfigurations = expandStorageConfigurations(d)

	out, err := conn.CreateEnvironment(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionCreating, ResNameEnvironment, d.Get("name").(string), err)
	}

	if out == nil || out.EnvironmentId == nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionCreating, ResNameEnvironment, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.EnvironmentId))

	// TIP: -- 5. Use a waiter to wait for create to complete
	if _, err := waitEnvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionWaitingForCreation, ResNameEnvironment, d.Id(), err)
	}

	// TIP: -- 6. Call the Read function in the Create return
	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).M2Client(ctx)

	out, err := findEnvironmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] M2 Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionReading, ResNameEnvironment, d.Id(), err)
	}

	d.Set("arn", out.EnvironmentArn)
	d.Set("description", out.Description)
	d.Set("engine_type", string(out.EngineType))
	d.Set("engine_version", out.EngineVersion)
	d.Set("instance_type", out.InstanceType)
	d.Set("kms_key_id", out.KmsKeyId)

	d.Set("name", out.Name)

	efsConfig, fsxConfig := flattenStorageConfig(out.StorageConfigurations)

	d.Set("efs_mount", efsConfig)
	d.Set("fsx_mount", fsxConfig)

	d.Set("preferred_maintenance_window", out.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", out.PubliclyAccessible)
	d.Set("security_groups", out.SecurityGroupIds)
	d.Set("subnet_ids", out.SubnetIds)
	d.Set("high_availability_config", flattenHaConfig(out.HighAvailabilityConfig))
	d.Set("load_balancer_arn", out.LoadBalancerArn)

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).M2Client(ctx)

	update := false

	in := &m2.UpdateEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("apply_changes_during_maintenance_window"); ok && v.(bool) {
		if d.HasChangesExcept("apply_changes_during_maintenance_window", "engine_version") {
			return create.AppendDiagError(diags, names.M2, create.ErrActionUpdating, ResNameEnvironment, d.Id(), fmt.Errorf("cannot make changes to any configuration except `engine_version` during maintenance window"))
		}
		in.ApplyDuringMaintenanceWindow = d.Get("apply_changes_during_maintenance_window").(bool)
		in.EngineVersion = aws.String(d.Get("engine_version").(string))
		update = true
	} else {
		if d.HasChange("engine_version") {
			in.EngineVersion = aws.String(d.Get("engine_version").(string))
			update = true
		}

		if d.HasChange("high_availability_config") {
			if v, ok := d.GetOk("high_availability_config"); ok && len(v.([]interface{})) > 0 {
				config := v.([]interface{})[0].(map[string]interface{})
				in.DesiredCapacity = aws.Int32(int32(config["desired_capacity"].(int)))
			} else {
				in.DesiredCapacity = aws.Int32(1)
			}
			update = true
		}

		if d.HasChange("instance_type") {
			in.InstanceType = aws.String(d.Get("instance_type").(string))
			update = true
		}

		if d.HasChange("preferred_maintenance_window") {
			in.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
			update = true
		}
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating M2 Environment (%s): %#v", d.Id(), in)
	out, err := conn.UpdateEnvironment(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionUpdating, ResNameEnvironment, d.Id(), err)
	}

	if _, err := waitEnvironmentUpdated(ctx, conn, aws.ToString(out.EnvironmentId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionWaitingForUpdate, ResNameEnvironment, d.Id(), err)
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).M2Client(ctx)

	log.Printf("[INFO] Deleting M2 Environment %s", d.Id())

	_, err := conn.DeleteEnvironment(ctx, &m2.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionDeleting, ResNameEnvironment, d.Id(), err)
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.M2, create.ErrActionWaitingForDeletion, ResNameEnvironment, d.Id(), err)
	}

	return diags
}

func waitEnvironmentCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.EnvironmentLifecycleCreating),
		Target:                    enum.Slice(types.EnvironmentLifecycleAvailable),
		Refresh:                   statusEnvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEnvironmentUpdated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.EnvironmentLifecycleUpdating),
		Target:                    enum.Slice(types.EnvironmentLifecycleAvailable),
		Refresh:                   statusEnvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.EnvironmentLifecycleAvailable, types.EnvironmentLifecycleCreating, types.EnvironmentLifecycleDeleting),
		Target:  []string{},
		Refresh: statusEnvironment(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusEnvironment(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findEnvironmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findEnvironmentByID(ctx context.Context, conn *m2.Client, id string) (*m2.GetEnvironmentOutput, error) {
	in := &m2.GetEnvironmentInput{
		EnvironmentId: aws.String(id),
	}
	out, err := conn.GetEnvironment(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandStorageConfigurations(d *schema.ResourceData) []types.StorageConfiguration {
	configs := make([]types.StorageConfiguration, 0)

	if efsMounts, ok := d.GetOk("efs_mount"); ok {
		for _, config := range efsMounts.(*schema.Set).List() {
			configMap := config.(map[string]interface{})
			configs = append(configs, expandStorageConfiguration(configMap, true))
		}
	}

	if fsxMounts, ok := d.GetOk("fsx_mount"); ok {
		for _, config := range fsxMounts.(*schema.Set).List() {
			configMap := config.(map[string]interface{})
			configs = append(configs, expandStorageConfiguration(configMap, false))
		}
	}
	return configs
}

func expandHaConfig(tfList []interface{}) *types.HighAvailabilityConfig {
	if len(tfList) == 0 {
		return nil
	}

	v, ok := tfList[0].(map[string]interface{})

	if ok {

		return &types.HighAvailabilityConfig{
			DesiredCapacity: aws.Int32(int32(v["desired_capacity"].(int))),
		}
	}
	return nil
}

func flattenHaConfig(haConfig *types.HighAvailabilityConfig) []interface{} {
	if haConfig == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"desired_capacity": haConfig.DesiredCapacity,
	}}
}

func expandStorageConfiguration(m map[string]interface{}, efs bool) types.StorageConfiguration {
	if efs {
		return &types.StorageConfigurationMemberEfs{
			Value: types.EfsStorageConfiguration{
				FileSystemId: aws.String(m["file_system_id"].(string)),
				MountPoint:   aws.String(m["mount_point"].(string)),
			},
		}
	} else {
		return &types.StorageConfigurationMemberFsx{
			Value: types.FsxStorageConfiguration{
				FileSystemId: aws.String(m["file_system_id"].(string)),
				MountPoint:   aws.String(m["mount_point"].(string)),
			},
		}
	}
}

func flattenStorageConfig(storageConfig []types.StorageConfiguration) ([]interface{}, []interface{}) {
	efsConfig := make([]interface{}, 0)
	fsxConfig := make([]interface{}, 0)
	for _, config := range storageConfig {
		switch config.(type) {
		case *types.StorageConfigurationMemberEfs:
			r := config.(*types.StorageConfigurationMemberEfs)
			efsConfig = appendStorageConfig(r.Value.FileSystemId, r.Value.MountPoint, efsConfig)
		case *types.StorageConfigurationMemberFsx:
			r := config.(*types.StorageConfigurationMemberFsx)
			fsxConfig = appendStorageConfig(r.Value.FileSystemId, r.Value.MountPoint, fsxConfig)
		}
	}
	return efsConfig, fsxConfig
}

func appendStorageConfig(fileSystemId *string, mountPoint *string, config []interface{}) []interface{} {
	return append(config, map[string]string{
		"file_system_id": *fileSystemId,
		"mount_point":    *mountPoint,
	})
}
