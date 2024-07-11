// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_openzfs_volume", name="OpenZFS Volume")
// @Tags(identifierAttribute="arn")
func resourceOpenZFSVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOpenZFSVolumeCreate,
		ReadWithoutTimeout:   resourceOpenZFSVolumeRead,
		UpdateWithoutTimeout: resourceOpenZFSVolumeUpdate,
		DeleteWithoutTimeout: resourceOpenZFSVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("delete_volume_options", nil)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags_to_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"data_compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      awstypes.OpenZFSDataCompressionTypeNone,
				ValidateFunc: enum.Validate[awstypes.OpenZFSDataCompressionType](),
			},
			"delete_volume_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: enum.Validate[awstypes.DeleteFileSystemOpenZFSOption](),
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 203),
			},
			"nfs_exports": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_configurations": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"clients": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 128),
											validation.StringMatch(regexache.MustCompile(`^[ -~]{1,128}$`), "must be either IP Address or CIDR"),
										),
									},
									"options": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 20,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 128),
										},
									},
								},
							},
						},
					},
				},
			},
			"origin_snapshot": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"copy_strategy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: enum.Validate[awstypes.OpenZFSCopyStrategy](),
						},
						"snapshot_arn": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(8, 512),
								validation.StringMatch(regexache.MustCompile(`^arn:.*`), "must specify the full ARN of the snapshot"),
							),
						},
					},
				},
			},
			"parent_volume_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(23, 23),
					validation.StringMatch(regexache.MustCompile(`^(fsvol-[0-9a-f]{17,})$`), "must specify a filesystem id i.e. fs-12345678"),
				),
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"record_size_kib": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      128,
				ValidateFunc: validation.IntInSlice([]int{4, 8, 16, 32, 64, 128, 256, 512, 1024}),
			},
			"storage_capacity_quota_gib": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},
			"storage_capacity_reservation_gib": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},
			"user_and_group_quotas": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},
						"storage_capacity_quota_gib": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: enum.Validate[awstypes.OpenZFSQuotaType](),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVolumeType: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      awstypes.VolumeTypeOpenzfs,
				ValidateFunc: enum.Validate[awstypes.VolumeType](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOpenZFSVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	openzfsConfig := &awstypes.CreateOpenZFSVolumeConfiguration{
		ParentVolumeId: aws.String(d.Get("parent_volume_id").(string)),
	}

	if v, ok := d.GetOk("copy_tags_to_snapshots"); ok {
		openzfsConfig.CopyTagsToSnapshots = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("data_compression_type"); ok {
		openzfsConfig.DataCompressionType = awstypes.DataCompressionType(v.(string))
	}

	if v, ok := d.GetOk("nfs_exports"); ok {
		openzfsConfig.NfsExports = expandOpenZFSNfsExports(v.([]interface{}))
	}

	if v, ok := d.GetOk("origin_snapshot"); ok {
		openzfsConfig.OriginSnapshot = expandCreateOpenZFSOriginSnapshotConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("read_only"); ok {
		openzfsConfig.ReadOnly = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("record_size_kib"); ok {
		openzfsConfig.RecordSizeKiB = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("storage_capacity_quota_gib"); ok {
		openzfsConfig.StorageCapacityQuotaGiB = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("storage_capacity_reservation_gib"); ok {
		openzfsConfig.StorageCapacityReservationGiB = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("user_and_group_quotas"); ok {
		openzfsConfig.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(v.(*schema.Set).List())
	}

	name := d.Get(names.AttrName).(string)
	input := &fsx.CreateVolumeInput{
		ClientRequestToken:   aws.String(id.UniqueId()),
		Name:                 aws.String(name),
		OpenZFSConfiguration: openzfsConfig,
		Tags:                 getTagsIn(ctx),
		VolumeType:           aws.String(d.Get(names.AttrVolumeType).(string)),
	}

	output, err := conn.CreateVolume(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for OpenZFS Volume (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Volume.VolumeId))

	if _, err := waitVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceOpenZFSVolumeRead(ctx, d, meta)...)
}

func resourceOpenZFSVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	volume, err := findOpenZFSVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for OpenZFS Volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for OpenZFS Volume (%s): %s", d.Id(), err)
	}

	openzfsConfig := volume.OpenZFSConfiguration

	d.Set(names.AttrARN, volume.ResourceARN)
	d.Set("copy_tags_to_snapshots", openzfsConfig.CopyTagsToSnapshots)
	d.Set("data_compression_type", openzfsConfig.DataCompressionType)
	d.Set(names.AttrName, volume.Name)
	if err := d.Set("nfs_exports", flattenOpenZFSNfsExports(openzfsConfig.NfsExports)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting nfs_exports: %s", err)
	}
	if err := d.Set("origin_snapshot", flattenOpenZFSOriginSnapshotConfiguration(openzfsConfig.OriginSnapshot)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting nfs_exports: %s", err)
	}
	d.Set("parent_volume_id", openzfsConfig.ParentVolumeId)
	d.Set("read_only", openzfsConfig.ReadOnly)
	d.Set("record_size_kib", openzfsConfig.RecordSizeKiB)
	d.Set("storage_capacity_quota_gib", openzfsConfig.StorageCapacityQuotaGiB)
	d.Set("storage_capacity_reservation_gib", openzfsConfig.StorageCapacityReservationGiB)
	if err := d.Set("user_and_group_quotas", flattenOpenZFSUserOrGroupQuotas(openzfsConfig.UserAndGroupQuotas)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_and_group_quotas: %s", err)
	}
	d.Set(names.AttrVolumeType, volume.VolumeType)

	setTagsOut(ctx, volume.Tags)

	return diags
}

func resourceOpenZFSVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		openzfsConfig := &awstypes.UpdateOpenZFSVolumeConfiguration{}

		if d.HasChange("data_compression_type") {
			openzfsConfig.DataCompressionType = awstypes.DataCompressionType(d.Get("data_compression_type").(string))
		}

		if d.HasChange("nfs_exports") {
			openzfsConfig.NfsExports = expandOpenZFSNfsExports(d.Get("nfs_exports").([]interface{}))
		}

		if d.HasChange("read_only") {
			openzfsConfig.ReadOnly = aws.Bool(d.Get("read_only").(bool))
		}

		if d.HasChange("record_size_kib") {
			openzfsConfig.RecordSizeKiB = aws.Int64(int64(d.Get("record_size_kib").(int)))
		}

		if d.HasChange("storage_capacity_quota_gib") {
			openzfsConfig.StorageCapacityQuotaGiB = aws.Int64(int64(d.Get("storage_capacity_quota_gib").(int)))
		}

		if d.HasChange("storage_capacity_reservation_gib") {
			openzfsConfig.StorageCapacityReservationGiB = aws.Int64(int64(d.Get("storage_capacity_reservation_gib").(int)))
		}

		if d.HasChange("user_and_group_quotas") {
			openzfsConfig.UserAndGroupQuotas = expandOpenZFSUserOrGroupQuotas(d.Get("user_and_group_quotas").(*schema.Set).List())
		}

		input := &fsx.UpdateVolumeInput{
			ClientRequestToken:   aws.String(id.UniqueId()),
			OpenZFSConfiguration: openzfsConfig,
			VolumeId:             aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateVolume(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for OpenZFS Volume (%s): %s", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) update: %s", d.Id(), err)
		}

		if _, err := waitVolumeAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeVolumeUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeVolumeUpdate, err)
		}
	}

	return append(diags, resourceOpenZFSVolumeRead(ctx, d, meta)...)
}

func resourceOpenZFSVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.DeleteVolumeInput{
		VolumeId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("delete_volume_options"); ok && len(v.([]interface{})) > 0 {
		input.OpenZFSConfiguration = &awstypes.DeleteVolumeOpenZFSConfiguration{
			Options: flex.ExpandStringList(v.([]interface{})),
		}
	}

	log.Printf("[DEBUG] Deleting FSx for OpenZFS Volume: %s", d.Id())
	_, err := conn.DeleteVolume(ctx, input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeVolumeNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for OpenZFS Volume (%s): %s", d.Id(), err)
	}

	if _, err := waitVolumeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for OpenZFS Volume (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandOpenZFSUserOrGroupQuotas(cfg []interface{}) []*awstypes.OpenZFSUserOrGroupQuota {
	quotas := []*awstypes.OpenZFSUserOrGroupQuota{}

	for _, quota := range cfg {
		expandedQuota := expandOpenZFSUserOrGroupQuota(quota.(map[string]interface{}))
		if expandedQuota != nil {
			quotas = append(quotas, expandedQuota)
		}
	}

	return quotas
}

func expandOpenZFSUserOrGroupQuota(conf map[string]interface{}) *awstypes.OpenZFSUserOrGroupQuota {
	if len(conf) < 1 {
		return nil
	}

	out := awstypes.OpenZFSUserOrGroupQuota{}

	if v, ok := conf[names.AttrID].(int); ok {
		out.Id = aws.Int64(int64(v))
	}

	if v, ok := conf["storage_capacity_quota_gib"].(int); ok {
		out.StorageCapacityQuotaGiB = aws.Int64(int64(v))
	}

	if v, ok := conf[names.AttrType].(string); ok {
		out.Type = aws.String(v)
	}

	return &out
}

func expandOpenZFSNfsExports(cfg []interface{}) []*awstypes.OpenZFSNfsExport { // nosemgrep:ci.caps4-in-func-name
	exports := []*awstypes.OpenZFSNfsExport{}

	for _, export := range cfg {
		expandedExport := expandOpenZFSNfsExport(export.(map[string]interface{}))
		if expandedExport != nil {
			exports = append(exports, expandedExport)
		}
	}

	return exports
}

func expandOpenZFSNfsExport(cfg map[string]interface{}) *awstypes.OpenZFSNfsExport { // nosemgrep:ci.caps4-in-func-name
	out := awstypes.OpenZFSNfsExport{}

	if v, ok := cfg["client_configurations"]; ok {
		out.ClientConfigurations = expandOpenZFSClientConfigurations(v.(*schema.Set).List())
	}

	return &out
}

func expandOpenZFSClientConfigurations(cfg []interface{}) []*awstypes.OpenZFSClientConfiguration {
	configurations := []*awstypes.OpenZFSClientConfiguration{}

	for _, configuration := range cfg {
		expandedConfiguration := expandOpenZFSClientConfiguration(configuration.(map[string]interface{}))
		if expandedConfiguration != nil {
			configurations = append(configurations, expandedConfiguration)
		}
	}

	return configurations
}

func expandOpenZFSClientConfiguration(conf map[string]interface{}) *awstypes.OpenZFSClientConfiguration {
	out := awstypes.OpenZFSClientConfiguration{}

	if v, ok := conf["clients"].(string); ok && len(v) > 0 {
		out.Clients = aws.String(v)
	}

	if v, ok := conf["options"].([]interface{}); ok {
		out.Options = flex.ExpandStringList(v)
	}

	return &out
}

func expandCreateOpenZFSOriginSnapshotConfiguration(cfg []interface{}) *awstypes.CreateOpenZFSOriginSnapshotConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := awstypes.CreateOpenZFSOriginSnapshotConfiguration{}

	if v, ok := conf["copy_strategy"].(string); ok {
		out.CopyStrategy = aws.String(v)
	}

	if v, ok := conf["snapshot_arn"].(string); ok {
		out.SnapshotARN = aws.String(v)
	}

	return &out
}

func flattenOpenZFSNfsExports(rs []*awstypes.OpenZFSNfsExport) []map[string]interface{} { // nosemgrep:ci.caps4-in-func-name
	exports := make([]map[string]interface{}, 0)

	for _, export := range rs {
		if export != nil {
			cfg := make(map[string]interface{})
			cfg["client_configurations"] = flattenOpenZFSClientConfigurations(export.ClientConfigurations)
			exports = append(exports, cfg)
		}
	}

	if len(exports) > 0 {
		return exports
	}

	return nil
}

func flattenOpenZFSClientConfigurations(rs []*awstypes.OpenZFSClientConfiguration) []map[string]interface{} {
	configurations := make([]map[string]interface{}, 0)

	for _, configuration := range rs {
		if configuration != nil {
			cfg := make(map[string]interface{})
			cfg["clients"] = aws.ToString(configuration.Clients)
			cfg["options"] = flex.FlattenStringList(configuration.Options)
			configurations = append(configurations, cfg)
		}
	}

	if len(configurations) > 0 {
		return configurations
	}

	return nil
}

func flattenOpenZFSUserOrGroupQuotas(rs []*awstypes.OpenZFSUserOrGroupQuota) []map[string]interface{} {
	quotas := make([]map[string]interface{}, 0)

	for _, quota := range rs {
		if quota != nil {
			cfg := make(map[string]interface{})
			cfg[names.AttrID] = aws.ToInt64(quota.Id)
			cfg["storage_capacity_quota_gib"] = aws.ToInt64(quota.StorageCapacityQuotaGiB)
			cfg[names.AttrType] = aws.ToString(quota.Type)
			quotas = append(quotas, cfg)
		}
	}

	if len(quotas) > 0 {
		return quotas
	}

	return nil
}

func flattenOpenZFSOriginSnapshotConfiguration(rs *awstypes.OpenZFSOriginSnapshotConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.CopyStrategy != nil {
		m["copy_strategy"] = aws.ToString(rs.CopyStrategy)
	}
	if rs.SnapshotARN != nil {
		m["snapshot_arn"] = aws.ToString(rs.SnapshotARN)
	}

	return []interface{}{m}
}

func findOpenZFSVolumeByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.Volume, error) {
	output, err := findVolumeByIDAndType(ctx, conn, id, awstypes.VolumeTypeOpenzfs)

	if err != nil {
		return nil, err
	}

	if output.OpenZFSConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
