// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_ontap_volume", name="ONTAP Volume")
// @Tags(identifierAttribute="arn")
func ResourceONTAPVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceONTAPVolumeCreate,
		ReadWithoutTimeout:   resourceONTAPVolumeRead,
		UpdateWithoutTimeout: resourceONTAPVolumeUpdate,
		DeleteWithoutTimeout: resourceONTAPVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("bypass_snaplock_enterprise_retention", false)
				d.Set("skip_final_backup", false)

				return []*schema.ResourceData{d}, nil
			},
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
			"bypass_snaplock_enterprise_retention": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flexcache_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"junction_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 203),
			},
			"ontap_volume_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.InputOntapVolumeType_Values(), false),
			},
			"security_style": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(fsx.StorageVirtualMachineRootVolumeSecurityStyle_Values(), false),
			},
			"size_in_megabytes": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snaplock_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit_log_volume": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"autocommit_period": {
							Type:             schema.TypeList,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(fsx.AutocommitPeriodType_Values(), false),
									},
									"value": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 65535),
									},
								},
							},
						},
						"privileged_delete": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      fsx.PrivilegedDeleteDisabled,
							ValidateFunc: validation.StringInSlice(fsx.PrivilegedDelete_Values(), false),
						},
						"retention_period": {
							Type:             schema.TypeList,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_retention": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(fsx.RetentionPeriodType_Values(), false),
												},
												"value": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
											},
										},
									},
									"maximum_retention": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(fsx.RetentionPeriodType_Values(), false),
												},
												"value": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
											},
										},
									},
									"minimum_retention": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(fsx.RetentionPeriodType_Values(), false),
												},
												"value": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 65535),
												},
											},
										},
									},
								},
							},
						},
						"snaplock_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(fsx.SnaplockType_Values(), false),
						},
						"volume_append_mode_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"snapshot_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"storage_efficiency_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"storage_virtual_machine_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(21, 21),
			},
			"tiering_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cooling_period": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 183),
						},
						"name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(fsx.TieringPolicyName_Values(), false),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.VolumeTypeOntap,
				ValidateFunc: validation.StringInSlice(fsx.VolumeType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceONTAPVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	ontapConfig := &fsx.CreateOntapVolumeConfiguration{
		SizeInMegabytes:         aws.Int64(int64(d.Get("size_in_megabytes").(int))),
		StorageVirtualMachineId: aws.String(d.Get("storage_virtual_machine_id").(string)),
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		ontapConfig.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("junction_path"); ok {
		ontapConfig.JunctionPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ontap_volume_type"); ok {
		ontapConfig.OntapVolumeType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_style"); ok {
		ontapConfig.SecurityStyle = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snaplock_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		ontapConfig.SnaplockConfiguration = expandCreateSnaplockConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("snapshot_policy"); ok {
		ontapConfig.SnapshotPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("storage_efficiency_enabled"); ok {
		ontapConfig.StorageEfficiencyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("tiering_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		ontapConfig.TieringPolicy = expandTieringPolicy(v.([]interface{})[0].(map[string]interface{}))
	}

	name := d.Get("name").(string)
	input := &fsx.CreateVolumeInput{
		Name:               aws.String(name),
		OntapConfiguration: ontapConfig,
		Tags:               getTagsIn(ctx),
		VolumeType:         aws.String(d.Get("volume_type").(string)),
	}

	output, err := conn.CreateVolumeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for NetApp ONTAP Volume (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Volume.VolumeId))

	if _, err := waitVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceONTAPVolumeRead(ctx, d, meta)...)
}

func resourceONTAPVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	volume, err := FindONTAPVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for NetApp ONTAP Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for NetApp ONTAP Volume (%s): %s", d.Id(), err)
	}

	ontapConfig := volume.OntapConfiguration

	d.Set("arn", volume.ResourceARN)
	d.Set("copy_tags_to_backups", ontapConfig.CopyTagsToBackups)
	d.Set("file_system_id", volume.FileSystemId)
	d.Set("junction_path", ontapConfig.JunctionPath)
	d.Set("name", volume.Name)
	d.Set("ontap_volume_type", ontapConfig.OntapVolumeType)
	d.Set("security_style", ontapConfig.SecurityStyle)
	d.Set("size_in_megabytes", ontapConfig.SizeInMegabytes)
	if ontapConfig.SnaplockConfiguration != nil {
		if err := d.Set("snaplock_configuration", []interface{}{flattenSnaplockConfiguration(ontapConfig.SnaplockConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting snaplock_configuration: %s", err)
		}
	} else {
		d.Set("snaplock_configuration", nil)
	}
	d.Set("snapshot_policy", ontapConfig.SnapshotPolicy)
	d.Set("storage_efficiency_enabled", ontapConfig.StorageEfficiencyEnabled)
	d.Set("storage_virtual_machine_id", ontapConfig.StorageVirtualMachineId)
	if ontapConfig.TieringPolicy != nil {
		if err := d.Set("tiering_policy", []interface{}{flattenTieringPolicy(ontapConfig.TieringPolicy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tiering_policy: %s", err)
		}
	} else {
		d.Set("tiering_policy", nil)
	}
	d.Set("uuid", ontapConfig.UUID)
	d.Set("volume_type", volume.VolumeType)

	return diags
}

func resourceONTAPVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		ontapConfig := &fsx.UpdateOntapVolumeConfiguration{}

		if d.HasChange("copy_tags_to_backups") {
			ontapConfig.CopyTagsToBackups = aws.Bool(d.Get("copy_tags_to_backups").(bool))
		}

		if d.HasChange("junction_path") {
			ontapConfig.JunctionPath = aws.String(d.Get("junction_path").(string))
		}

		if d.HasChange("security_style") {
			ontapConfig.SecurityStyle = aws.String(d.Get("security_style").(string))
		}

		if d.HasChange("size_in_megabytes") {
			ontapConfig.SizeInMegabytes = aws.Int64(int64(d.Get("size_in_megabytes").(int)))
		}

		if d.HasChange("snaplock_configuration") {
			if v, ok := d.GetOk("snaplock_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				ontapConfig.SnaplockConfiguration = expandUpdateSnaplockConfiguration(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("snapshot_policy") {
			ontapConfig.SnapshotPolicy = aws.String(d.Get("snapshot_policy").(string))
		}

		if d.HasChange("storage_efficiency_enabled") {
			ontapConfig.StorageEfficiencyEnabled = aws.Bool(d.Get("storage_efficiency_enabled").(bool))
		}

		if d.HasChange("tiering_policy") {
			if v, ok := d.GetOk("tiering_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				ontapConfig.TieringPolicy = expandTieringPolicy(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		input := &fsx.UpdateVolumeInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			OntapConfiguration: ontapConfig,
			VolumeId:           aws.String(d.Id()),
		}

		startTime := time.Now()
		_, err := conn.UpdateVolumeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for NetApp ONTAP Volume (%s): %s", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) update: %s", d.Id(), err)
		}

		if _, err := waitVolumeAdministrativeActionCompleted(ctx, conn, d.Id(), fsx.AdministrativeActionTypeVolumeUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) administrative action (%s) complete: %s", d.Id(), fsx.AdministrativeActionTypeVolumeUpdate, err)
		}
	}

	return append(diags, resourceONTAPVolumeRead(ctx, d, meta)...)
}

func resourceONTAPVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx for NetApp ONTAP Volume: %s", d.Id())
	_, err := conn.DeleteVolumeWithContext(ctx, &fsx.DeleteVolumeInput{
		OntapConfiguration: &fsx.DeleteVolumeOntapConfiguration{
			BypassSnaplockEnterpriseRetention: aws.Bool(d.Get("bypass_snaplock_enterprise_retention").(bool)),
			SkipFinalBackup:                   aws.Bool(d.Get("skip_final_backup").(bool)),
		},
		VolumeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for NetApp ONTAP Volume (%s): %s", d.Id(), err)
	}

	if _, err := waitVolumeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const minTieringPolicyCoolingPeriod = 2

func expandTieringPolicy(tfMap map[string]interface{}) *fsx.TieringPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.TieringPolicy{}

	// Cooling period only accepts a minimum of 2 but int will return 0 not nil if unset.
	// Therefore we only set it if it is 2 or more.
	if v, ok := tfMap["cooling_period"].(int); ok && v >= minTieringPolicyCoolingPeriod {
		apiObject.CoolingPeriod = aws.Int64(int64(v))
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenTieringPolicy(apiObject *fsx.TieringPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CoolingPeriod; v != nil {
		if v := aws.Int64Value(v); v >= minTieringPolicyCoolingPeriod {
			tfMap["cooling_period"] = v
		}
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func expandCreateSnaplockConfiguration(tfMap map[string]interface{}) *fsx.CreateSnaplockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.CreateSnaplockConfiguration{}

	if v, ok := tfMap["audit_log_volume"].(bool); ok && v {
		apiObject.AuditLogVolume = aws.Bool(v)
	}

	if v, ok := tfMap["autocommit_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutocommitPeriod = expandAutocommitPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["privileged_delete"].(string); ok && v != "" {
		apiObject.PrivilegedDelete = aws.String(v)
	}

	if v, ok := tfMap["retention_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetentionPeriod = expandSnaplockRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["snaplock_type"].(string); ok && v != "" {
		apiObject.SnaplockType = aws.String(v)
	}

	if v, ok := tfMap["volume_append_mode_enabled"].(bool); ok && v {
		apiObject.VolumeAppendModeEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandUpdateSnaplockConfiguration(tfMap map[string]interface{}) *fsx.UpdateSnaplockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.UpdateSnaplockConfiguration{}

	if v, ok := tfMap["audit_log_volume"].(bool); ok && v {
		apiObject.AuditLogVolume = aws.Bool(v)
	}

	if v, ok := tfMap["autocommit_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutocommitPeriod = expandAutocommitPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["privileged_delete"].(string); ok && v != "" {
		apiObject.PrivilegedDelete = aws.String(v)
	}

	if v, ok := tfMap["retention_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetentionPeriod = expandSnaplockRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["volume_append_mode_enabled"].(bool); ok && v {
		apiObject.VolumeAppendModeEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandAutocommitPeriod(tfMap map[string]interface{}) *fsx.AutocommitPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.AutocommitPeriod{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(int); ok && v != 0 {
		apiObject.Value = aws.Int64(int64(v))
	}

	return apiObject
}

func expandSnaplockRetentionPeriod(tfMap map[string]interface{}) *fsx.SnaplockRetentionPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.SnaplockRetentionPeriod{}

	if v, ok := tfMap["default_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DefaultRetention = expandRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["maximum_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.MaximumRetention = expandRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["minimum_retention"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.MinimumRetention = expandRetentionPeriod(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandRetentionPeriod(tfMap map[string]interface{}) *fsx.RetentionPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &fsx.RetentionPeriod{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(int); ok && v != 0 {
		apiObject.Value = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenSnaplockConfiguration(apiObject *fsx.SnaplockConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuditLogVolume; v != nil {
		tfMap["audit_log_volume"] = aws.BoolValue(v)
	}

	if v := apiObject.AutocommitPeriod; v != nil {
		tfMap["autocommit_period"] = []interface{}{flattenAutocommitPeriod(v)}
	}

	if v := apiObject.PrivilegedDelete; v != nil {
		tfMap["privileged_delete"] = aws.StringValue(v)
	}

	if v := apiObject.RetentionPeriod; v != nil {
		tfMap["retention_period"] = []interface{}{flattenSnaplockRetentionPeriod(v)}
	}

	if v := apiObject.SnaplockType; v != nil {
		tfMap["snaplock_type"] = aws.StringValue(v)
	}

	if v := apiObject.VolumeAppendModeEnabled; v != nil {
		tfMap["volume_append_mode_enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenAutocommitPeriod(apiObject *fsx.AutocommitPeriod) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenSnaplockRetentionPeriod(apiObject *fsx.SnaplockRetentionPeriod) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultRetention; v != nil {
		tfMap["default_retention"] = []interface{}{flattenRetentionPeriod(v)}
	}

	if v := apiObject.MaximumRetention; v != nil {
		tfMap["maximum_retention"] = []interface{}{flattenRetentionPeriod(v)}
	}

	if v := apiObject.MinimumRetention; v != nil {
		tfMap["minimum_retention"] = []interface{}{flattenRetentionPeriod(v)}
	}

	return tfMap
}

func flattenRetentionPeriod(apiObject *fsx.RetentionPeriod) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.Int64Value(v)
	}

	return tfMap
}

func FindONTAPVolumeByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Volume, error) {
	output, err := findVolumeByIDAndType(ctx, conn, id, fsx.VolumeTypeOntap)

	if err != nil {
		return nil, err
	}

	if output.OntapConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findVolumeByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Volume, error) {
	input := &fsx.DescribeVolumesInput{
		VolumeIds: aws.StringSlice([]string{id}),
	}

	return findVolume(ctx, conn, input, tfslices.PredicateTrue[*fsx.Volume]())
}

func findVolumeByIDAndType(ctx context.Context, conn *fsx.FSx, volID, volType string) (*fsx.Volume, error) {
	input := &fsx.DescribeVolumesInput{
		VolumeIds: aws.StringSlice([]string{volID}),
	}
	filter := func(fs *fsx.Volume) bool {
		return aws.StringValue(fs.VolumeType) == volType
	}

	return findVolume(ctx, conn, input, filter)
}

func findVolume(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeVolumesInput, filter tfslices.Predicate[*fsx.Volume]) (*fsx.Volume, error) {
	output, err := findVolumes(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findVolumes(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeVolumesInput, filter tfslices.Predicate[*fsx.Volume]) ([]*fsx.Volume, error) {
	var output []*fsx.Volume

	err := conn.DescribeVolumesPagesWithContext(ctx, input, func(page *fsx.DescribeVolumesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Volumes {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusVolume(ctx context.Context, conn *fsx.FSx, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func waitVolumeCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.VolumeLifecycleCreating, fsx.VolumeLifecyclePending},
		Target:  []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable},
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Volume); ok {
		if status, reason := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.VolumeLifecycleFailed && reason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(reason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeUpdated(ctx context.Context, conn *fsx.FSx, id string, startTime time.Time, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.VolumeLifecyclePending},
		Target:  []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable},
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Volume); ok {
		switch status := aws.StringValue(output.Lifecycle); status {
		case fsx.VolumeLifecycleFailed:
			// Report any failed non-VOLUME_UPDATE administrative actions.
			// See https://docs.aws.amazon.com/fsx/latest/APIReference/API_AdministrativeAction.html#FSx-Type-AdministrativeAction-AdministrativeActionType.
			administrativeActions := tfslices.Filter(output.AdministrativeActions, func(v *fsx.AdministrativeAction) bool {
				return v != nil && aws.StringValue(v.Status) == fsx.StatusFailed && aws.StringValue(v.AdministrativeActionType) != fsx.AdministrativeActionTypeVolumeUpdate && v.FailureDetails != nil && startTime.Before(aws.TimeValue(v.RequestTime))
			})
			administrativeActionsError := errors.Join(tfslices.ApplyToAll(administrativeActions, func(v *fsx.AdministrativeAction) error {
				return fmt.Errorf("%s: %s", aws.StringValue(v.AdministrativeActionType), aws.StringValue(v.FailureDetails.Message))
			})...)

			if reason := output.LifecycleTransitionReason; reason != nil {
				if message := aws.StringValue(reason.Message); administrativeActionsError != nil {
					tfresource.SetLastError(err, fmt.Errorf("%s: %w", message, administrativeActionsError))
				} else {
					tfresource.SetLastError(err, errors.New(message))
				}
			} else {
				tfresource.SetLastError(err, administrativeActionsError)
			}
		}

		return output, err
	}

	return nil, err
}

func waitVolumeDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable, fsx.VolumeLifecycleDeleting},
		Target:  []string{},
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Volume); ok {
		if status, reason := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.VolumeLifecycleFailed && reason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(reason.Message)))
		}

		return output, err
	}

	return nil, err
}

func findVolumeAdministrativeAction(ctx context.Context, conn *fsx.FSx, volID, actionType string) (*fsx.AdministrativeAction, error) {
	output, err := findVolumeByID(ctx, conn, volID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.AdministrativeActions {
		if v == nil {
			continue
		}

		if aws.StringValue(v.AdministrativeActionType) == actionType {
			return v, nil
		}
	}

	// If the administrative action isn't found, assume it's complete.
	return &fsx.AdministrativeAction{Status: aws.String(fsx.StatusCompleted)}, nil
}

func statusVolumeAdministrativeAction(ctx context.Context, conn *fsx.FSx, volID, actionType string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeAdministrativeAction(ctx, conn, volID, actionType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitVolumeAdministrativeActionCompleted(ctx context.Context, conn *fsx.FSx, volID, actionType string, timeout time.Duration) (*fsx.AdministrativeAction, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.StatusInProgress, fsx.StatusPending},
		Target:  []string{fsx.StatusCompleted, fsx.StatusUpdatedOptimizing},
		Refresh: statusVolumeAdministrativeAction(ctx, conn, volID, actionType),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.AdministrativeAction); ok {
		if status, details := aws.StringValue(output.Status), output.FailureDetails; status == fsx.StatusFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}
