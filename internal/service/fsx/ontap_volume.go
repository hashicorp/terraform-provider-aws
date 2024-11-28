// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_ontap_volume", name="ONTAP Volume")
// @Tags(identifierAttribute="arn")
func resourceONTAPVolume() *schema.Resource {
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
			"aggregate_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aggregates": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 12,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexache.MustCompile("^(aggr[0-9]{1,2})$"), "Each value must be in the format aggrX, where X is a number between 1 and the number of ha_pairs"),
							},
						},
						"constituents_per_aggregate": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 200),
						},
						"total_constituents": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrARN: {
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
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_backup_tags": tftags.TagsSchema(),
			"flexcache_endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"junction_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 203),
			},
			"ontap_volume_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InputOntapVolumeType](),
			},
			"security_style": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.StorageVirtualMachineRootVolumeSecurityStyle](),
			},
			"size_in_bytes": {
				Type:         nullable.TypeNullableInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 22517998000000000),
				ExactlyOneOf: []string{"size_in_bytes", "size_in_megabytes"},
			},
			"size_in_megabytes": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
				ExactlyOneOf: []string{"size_in_bytes", "size_in_megabytes"},
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
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.AutocommitPeriodType](),
									},
									names.AttrValue: {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 65535),
									},
								},
							},
						},
						"privileged_delete": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PrivilegedDeleteDisabled,
							ValidateDiagFunc: enum.Validate[awstypes.PrivilegedDelete](),
						},
						names.AttrRetentionPeriod: {
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
												names.AttrType: {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionPeriodType](),
												},
												names.AttrValue: {
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
												names.AttrType: {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionPeriodType](),
												},
												names.AttrValue: {
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
												names.AttrType: {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[awstypes.RetentionPeriodType](),
												},
												names.AttrValue: {
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
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SnaplockType](),
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
							Computed:     true,
							ValidateFunc: validation.IntBetween(2, 183),
						},
						names.AttrName: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TieringPolicyName](),
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
			"volume_style": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VolumeStyle](),
			},
			names.AttrVolumeType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.VolumeTypeOntap,
				ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceONTAPVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	ontapConfig := &awstypes.CreateOntapVolumeConfiguration{
		StorageVirtualMachineId: aws.String(d.Get("storage_virtual_machine_id").(string)),
	}

	if v, ok := d.GetOk("aggregate_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		ontapConfig.AggregateConfiguration = expandAggregateConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("copy_tags_to_backups"); ok {
		ontapConfig.CopyTagsToBackups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("junction_path"); ok {
		ontapConfig.JunctionPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ontap_volume_type"); ok {
		ontapConfig.OntapVolumeType = awstypes.InputOntapVolumeType(v.(string))
	}

	if v, ok := d.GetOk("security_style"); ok {
		ontapConfig.SecurityStyle = awstypes.SecurityStyle(v.(string))
	}

	if v, null, _ := nullable.Int(d.Get("size_in_bytes").(string)).ValueInt64(); !null && v > 0 {
		ontapConfig.SizeInBytes = aws.Int64(v)
	}

	if v, ok := d.GetOk("size_in_megabytes"); ok {
		ontapConfig.SizeInMegabytes = aws.Int32(int32(v.(int)))
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

	if v, ok := d.GetOk("volume_style"); ok {
		ontapConfig.VolumeStyle = awstypes.VolumeStyle(v.(string))
	}

	name := d.Get(names.AttrName).(string)
	input := &fsx.CreateVolumeInput{
		Name:               aws.String(name),
		OntapConfiguration: ontapConfig,
		Tags:               getTagsIn(ctx),
		VolumeType:         awstypes.VolumeType(d.Get(names.AttrVolumeType).(string)),
	}

	output, err := conn.CreateVolume(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for NetApp ONTAP Volume (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Volume.VolumeId))

	if _, err := waitVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceONTAPVolumeRead(ctx, d, meta)...)
}

func resourceONTAPVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	volume, err := findONTAPVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for NetApp ONTAP Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for NetApp ONTAP Volume (%s): %s", d.Id(), err)
	}

	ontapConfig := volume.OntapConfiguration

	if ontapConfig.AggregateConfiguration != nil {
		if err := d.Set("aggregate_configuration", []interface{}{flattenAggregateConfiguration(ontapConfig.AggregateConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aggregate_configuration: %s", err)
		}
	} else {
		d.Set("aggregate_configuration", nil)
	}
	d.Set(names.AttrARN, volume.ResourceARN)
	d.Set("copy_tags_to_backups", ontapConfig.CopyTagsToBackups)
	d.Set(names.AttrFileSystemID, volume.FileSystemId)
	d.Set("junction_path", ontapConfig.JunctionPath)
	d.Set(names.AttrName, volume.Name)
	d.Set("ontap_volume_type", ontapConfig.OntapVolumeType)
	d.Set("security_style", ontapConfig.SecurityStyle)
	d.Set("size_in_bytes", flex.Int64ToStringValue(ontapConfig.SizeInBytes))
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
	d.Set("volume_style", ontapConfig.VolumeStyle)
	d.Set(names.AttrVolumeType, volume.VolumeType)

	// Volume tags aren't set in the Describe response.
	// setTagsOut(ctx, volume.Tags)

	return diags
}

func resourceONTAPVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(
		"final_backup_tags",
		"skip_final_backup",
		names.AttrTags,
		names.AttrTagsAll,
	) {
		ontapConfig := &awstypes.UpdateOntapVolumeConfiguration{}

		if d.HasChange("copy_tags_to_backups") {
			ontapConfig.CopyTagsToBackups = aws.Bool(d.Get("copy_tags_to_backups").(bool))
		}

		if d.HasChange("junction_path") {
			ontapConfig.JunctionPath = aws.String(d.Get("junction_path").(string))
		}

		if d.HasChange("security_style") {
			ontapConfig.SecurityStyle = awstypes.SecurityStyle(d.Get("security_style").(string))
		}

		if d.HasChange("size_in_bytes") {
			if v, null, _ := nullable.Int(d.Get("size_in_bytes").(string)).ValueInt64(); !null && v > 0 {
				ontapConfig.SizeInBytes = aws.Int64(v)
			}
		}

		if d.HasChange("size_in_megabytes") {
			ontapConfig.SizeInMegabytes = aws.Int32(int32(d.Get("size_in_megabytes").(int)))
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
		_, err := conn.UpdateVolume(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for NetApp ONTAP Volume (%s): %s", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) update: %s", d.Id(), err)
		}

		if _, err := waitVolumeAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeVolumeUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP Volume (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeVolumeUpdate, err)
		}
	}

	return append(diags, resourceONTAPVolumeRead(ctx, d, meta)...)
}

func resourceONTAPVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.DeleteVolumeInput{
		OntapConfiguration: &awstypes.DeleteVolumeOntapConfiguration{
			BypassSnaplockEnterpriseRetention: aws.Bool(d.Get("bypass_snaplock_enterprise_retention").(bool)),
			SkipFinalBackup:                   aws.Bool(d.Get("skip_final_backup").(bool)),
		},
		VolumeId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("final_backup_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.OntapConfiguration.FinalBackupTags = Tags(tftags.New(ctx, v))
	}

	log.Printf("[DEBUG] Deleting FSx for NetApp ONTAP Volume: %s", d.Id())
	_, err := conn.DeleteVolume(ctx, input)

	if errs.IsA[*awstypes.VolumeNotFound](err) {
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

func findONTAPVolumeByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Volume, error) {
	output, err := findVolumeByIDAndType(ctx, conn, id, awstypes.VolumeTypeOntap)

	if err != nil {
		return nil, err
	}

	if output.OntapConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}

func findVolumeByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Volume, error) {
	input := &fsx.DescribeVolumesInput{
		VolumeIds: []string{id},
	}

	return findVolume(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Volume]())
}

func findVolumeByIDAndType(ctx context.Context, conn *fsx.Client, volID string, volType awstypes.VolumeType) (*awstypes.Volume, error) {
	input := &fsx.DescribeVolumesInput{
		VolumeIds: []string{volID},
	}
	filter := func(v *awstypes.Volume) bool {
		return v.VolumeType == volType
	}

	return findVolume(ctx, conn, input, filter)
}

func findVolume(ctx context.Context, conn *fsx.Client, input *fsx.DescribeVolumesInput, filter tfslices.Predicate[*awstypes.Volume]) (*awstypes.Volume, error) {
	output, err := findVolumes(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVolumes(ctx context.Context, conn *fsx.Client, input *fsx.DescribeVolumesInput, filter tfslices.Predicate[*awstypes.Volume]) ([]awstypes.Volume, error) {
	var output []awstypes.Volume

	pages := fsx.NewDescribeVolumesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.VolumeNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Volumes {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusVolume(ctx context.Context, conn *fsx.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitVolumeCreated(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VolumeLifecycleCreating, awstypes.VolumeLifecyclePending),
		Target:  enum.Slice(awstypes.VolumeLifecycleCreated, awstypes.VolumeLifecycleMisconfigured, awstypes.VolumeLifecycleAvailable),
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		if output.Lifecycle == awstypes.VolumeLifecycleFailed && output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeUpdated(ctx context.Context, conn *fsx.Client, id string, startTime time.Time, timeout time.Duration) (*awstypes.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VolumeLifecyclePending),
		Target:  enum.Slice(awstypes.VolumeLifecycleCreated, awstypes.VolumeLifecycleMisconfigured, awstypes.VolumeLifecycleAvailable),
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		switch output.Lifecycle {
		case awstypes.VolumeLifecycleFailed:
			// Report any failed non-VOLUME_UPDATE administrative actions.
			// See https://docs.aws.amazon.com/fsx/latest/APIReference/API_AdministrativeAction.html#FSx-Type-AdministrativeAction-AdministrativeActionType.
			administrativeActions := tfslices.Filter(output.AdministrativeActions, func(v awstypes.AdministrativeAction) bool {
				return v.Status == awstypes.StatusFailed && v.AdministrativeActionType != awstypes.AdministrativeActionTypeVolumeUpdate && v.FailureDetails != nil && startTime.Before(aws.ToTime(v.RequestTime))
			})
			administrativeActionsError := errors.Join(tfslices.ApplyToAll(administrativeActions, func(v awstypes.AdministrativeAction) error {
				return fmt.Errorf("%s: %s", string(v.AdministrativeActionType), aws.ToString(v.FailureDetails.Message))
			})...)

			if reason := output.LifecycleTransitionReason; reason != nil {
				if message := aws.ToString(reason.Message); administrativeActionsError != nil {
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

func waitVolumeDeleted(ctx context.Context, conn *fsx.Client, id string, timeout time.Duration) (*awstypes.Volume, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VolumeLifecycleCreated, awstypes.VolumeLifecycleMisconfigured, awstypes.VolumeLifecycleAvailable, awstypes.VolumeLifecycleDeleting),
		Target:  []string{},
		Refresh: statusVolume(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		if output.Lifecycle == awstypes.VolumeLifecycleFailed && output.LifecycleTransitionReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func findVolumeAdministrativeAction(ctx context.Context, conn *fsx.Client, volID string, actionType awstypes.AdministrativeActionType) (awstypes.AdministrativeAction, error) {
	output, err := findVolumeByID(ctx, conn, volID)

	if err != nil {
		return awstypes.AdministrativeAction{}, err
	}

	for _, v := range output.AdministrativeActions {
		if v.AdministrativeActionType == actionType {
			return v, nil
		}
	}

	// If the administrative action isn't found, assume it's complete.
	return awstypes.AdministrativeAction{Status: awstypes.StatusCompleted}, nil
}

func statusVolumeAdministrativeAction(ctx context.Context, conn *fsx.Client, volID string, actionType awstypes.AdministrativeActionType) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeAdministrativeAction(ctx, conn, volID, actionType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitVolumeAdministrativeActionCompleted(ctx context.Context, conn *fsx.Client, volID string, actionType awstypes.AdministrativeActionType, timeout time.Duration) (*awstypes.AdministrativeAction, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusInProgress, awstypes.StatusPending),
		Target:  enum.Slice(awstypes.StatusCompleted, awstypes.StatusUpdatedOptimizing),
		Refresh: statusVolumeAdministrativeAction(ctx, conn, volID, actionType),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AdministrativeAction); ok {
		if output.Status == awstypes.StatusFailed && output.FailureDetails != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func expandAggregateConfiguration(tfMap map[string]interface{}) *awstypes.CreateAggregateConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CreateAggregateConfiguration{}

	if v, ok := tfMap["aggregates"].([]interface{}); ok && v != nil {
		apiObject.Aggregates = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["constituents_per_aggregate"].(int); ok && v != 0 {
		apiObject.ConstituentsPerAggregate = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenAggregateConfiguration(apiObject *awstypes.AggregateConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	var aggregates int32

	if v := apiObject.Aggregates; v != nil {
		tfMap["aggregates"] = v
		//Need to get the count of aggregates for calculating constituents_per_aggregate
		aggregates = int32(len(v))
	}

	if v := apiObject.TotalConstituents; v != nil {
		tfMap["total_constituents"] = aws.ToInt32(v)
		//Since the api only returns totalConstituents, need to calculate the value of ConstituentsPerAggregate so state will be consistent with config
		if aggregates != 0 {
			tfMap["constituents_per_aggregate"] = aws.ToInt32(v) / aggregates
		} else {
			tfMap["constituents_per_aggregate"] = aws.ToInt32(v)
		}
	}

	return tfMap
}

const minTieringPolicyCoolingPeriod = 2

func expandTieringPolicy(tfMap map[string]interface{}) *awstypes.TieringPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TieringPolicy{}

	// Cooling period only accepts a minimum of 2 but int will return 0 not nil if unset.
	// Therefore we only set it if it is 2 or more.
	if tfMap[names.AttrName].(string) == string(awstypes.TieringPolicyNameAuto) || tfMap[names.AttrName].(string) == string(awstypes.TieringPolicyNameSnapshotOnly) {
		if v, ok := tfMap["cooling_period"].(int); ok && v >= minTieringPolicyCoolingPeriod {
			apiObject.CoolingPeriod = aws.Int32(int32(v))
		}
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = awstypes.TieringPolicyName(v)
	}

	return apiObject
}

func flattenTieringPolicy(apiObject *awstypes.TieringPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CoolingPeriod; v != nil {
		if v := aws.ToInt32(v); v >= minTieringPolicyCoolingPeriod {
			tfMap["cooling_period"] = v
		}
	}

	tfMap[names.AttrName] = string(apiObject.Name)

	return tfMap
}

func expandCreateSnaplockConfiguration(tfMap map[string]interface{}) *awstypes.CreateSnaplockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CreateSnaplockConfiguration{}

	if v, ok := tfMap["audit_log_volume"].(bool); ok && v {
		apiObject.AuditLogVolume = aws.Bool(v)
	}

	if v, ok := tfMap["autocommit_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutocommitPeriod = expandAutocommitPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["privileged_delete"].(string); ok && v != "" {
		apiObject.PrivilegedDelete = awstypes.PrivilegedDelete(v)
	}

	if v, ok := tfMap[names.AttrRetentionPeriod].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetentionPeriod = expandSnaplockRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["snaplock_type"].(string); ok && v != "" {
		apiObject.SnaplockType = awstypes.SnaplockType(v)
	}

	if v, ok := tfMap["volume_append_mode_enabled"].(bool); ok && v {
		apiObject.VolumeAppendModeEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandUpdateSnaplockConfiguration(tfMap map[string]interface{}) *awstypes.UpdateSnaplockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UpdateSnaplockConfiguration{}

	if v, ok := tfMap["audit_log_volume"].(bool); ok && v {
		apiObject.AuditLogVolume = aws.Bool(v)
	}

	if v, ok := tfMap["autocommit_period"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutocommitPeriod = expandAutocommitPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["privileged_delete"].(string); ok && v != "" {
		apiObject.PrivilegedDelete = awstypes.PrivilegedDelete(v)
	}

	if v, ok := tfMap[names.AttrRetentionPeriod].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.RetentionPeriod = expandSnaplockRetentionPeriod(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["volume_append_mode_enabled"].(bool); ok && v {
		apiObject.VolumeAppendModeEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandAutocommitPeriod(tfMap map[string]interface{}) *awstypes.AutocommitPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutocommitPeriod{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.AutocommitPeriodType(v)
	}

	if v, ok := tfMap[names.AttrValue].(int); ok && v != 0 {
		apiObject.Value = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSnaplockRetentionPeriod(tfMap map[string]interface{}) *awstypes.SnaplockRetentionPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SnaplockRetentionPeriod{}

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

func expandRetentionPeriod(tfMap map[string]interface{}) *awstypes.RetentionPeriod {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RetentionPeriod{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.RetentionPeriodType(v)
	}

	if v, ok := tfMap[names.AttrValue].(int); ok && v != 0 {
		apiObject.Value = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenSnaplockConfiguration(apiObject *awstypes.SnaplockConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuditLogVolume; v != nil {
		tfMap["audit_log_volume"] = aws.ToBool(v)
	}

	if v := apiObject.AutocommitPeriod; v != nil {
		tfMap["autocommit_period"] = []interface{}{flattenAutocommitPeriod(v)}
	}

	tfMap["privileged_delete"] = string(apiObject.PrivilegedDelete)

	if v := apiObject.RetentionPeriod; v != nil {
		tfMap[names.AttrRetentionPeriod] = []interface{}{flattenSnaplockRetentionPeriod(v)}
	}

	tfMap["snaplock_type"] = string(apiObject.SnaplockType)

	if v := apiObject.VolumeAppendModeEnabled; v != nil {
		tfMap["volume_append_mode_enabled"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenAutocommitPeriod(apiObject *awstypes.AutocommitPeriod) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrType] = string(apiObject.Type)

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenSnaplockRetentionPeriod(apiObject *awstypes.SnaplockRetentionPeriod) map[string]interface{} {
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

func flattenRetentionPeriod(apiObject *awstypes.RetentionPeriod) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrType] = string(apiObject.Type)

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToInt32(v)
	}

	return tfMap
}
