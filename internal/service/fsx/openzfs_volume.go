package fsx

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenzfsVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceOepnzfsVolumeCreate,
		Read:   resourceOpenzfsVolumeRead,
		Update: resourceOpenzfsVolumeUpdate,
		Delete: resourceOpenzfsVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"copy_tags_to_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"data_compression_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "NONE",
				ValidateFunc: validation.StringInSlice(fsx.OpenZFSDataCompressionType_Values(), false),
			},
			"name": {
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
											validation.StringMatch(regexp.MustCompile(`^[ -~]{1,128}$`), "must be either IP Address or CIDR"),
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
							ValidateFunc: validation.StringInSlice(fsx.OpenZFSCopyStrategy_Values(), false),
						},
						"snapshot_arn": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(8, 512),
								validation.StringMatch(regexp.MustCompile(`^arn:.*`), "must specify the full ARN of the snapshot"),
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
					validation.StringMatch(regexp.MustCompile(`^(fsvol-[0-9a-f]{17,})$`), "must specify a filesystem id i.e. fs-12345678"),
				),
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
						"id": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},
						"storage_capacity_quota_gib": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 2147483647),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(fsx.OpenZFSQuotaType_Values(), false),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"volume_type": {
				Type:         schema.TypeString,
				Default:      fsx.VolumeTypeOpenzfs,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(fsx.VolumeType_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOepnzfsVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &fsx.CreateVolumeInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		Name:               aws.String(d.Get("name").(string)),
		VolumeType:         aws.String(d.Get("volume_type").(string)),
		OpenZFSConfiguration: &fsx.CreateOpenZFSVolumeConfiguration{
			ParentVolumeId: aws.String(d.Get("parent_volume_id").(string)),
		},
	}

	if v, ok := d.GetOk("copy_tags_to_snapshots"); ok {
		input.OpenZFSConfiguration.CopyTagsToSnapshots = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("data_compression_type"); ok {
		input.OpenZFSConfiguration.DataCompressionType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("nfs_exports"); ok {
		input.OpenZFSConfiguration.NfsExports = expandOpenzfsVolumeNFSExports(v.([]interface{}))
	}

	if v, ok := d.GetOk("read_only"); ok {
		input.OpenZFSConfiguration.ReadOnly = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("storage_capacity_quota_gib"); ok {
		input.OpenZFSConfiguration.StorageCapacityQuotaGiB = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("storage_capacity_reservation_gib"); ok {
		input.OpenZFSConfiguration.StorageCapacityReservationGiB = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("user_and_group_quotas"); ok {
		input.OpenZFSConfiguration.UserAndGroupQuotas = expandOpenzfsVolumeUserAndGroupQuotas(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("origin_snapshot"); ok {
		input.OpenZFSConfiguration.OriginSnapshot = expandOpenzfsCreateVolumeOriginSnapshot(v.([]interface{}))

		log.Printf("[DEBUG] Creating FSx OpenZFS Volume: %s", input)
		result, err := conn.CreateVolume(input)

		if err != nil {
			return fmt.Errorf("error creating FSx OpenZFS Volume from snapshot: %w", err)
		}

		d.SetId(aws.StringValue(result.Volume.VolumeId))
	} else {
		log.Printf("[DEBUG] Creating FSx OpenZFS Volume: %s", input)
		result, err := conn.CreateVolume(input)

		if err != nil {
			return fmt.Errorf("error creating FSx OpenZFS Volume: %w", err)
		}

		d.SetId(aws.StringValue(result.Volume.VolumeId))
	}

	if _, err := waitVolumeCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for FSx OpenZFS Volume(%s) create: %w", d.Id(), err)
	}

	return resourceOpenzfsVolumeRead(d, meta)
}

func resourceOpenzfsVolumeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	volume, err := FindVolumeByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx OpenZFS volume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading FSx OpenZFS Volume (%s): %w", d.Id(), err)
	}

	openzfsConfig := volume.OpenZFSConfiguration

	if volume.OntapConfiguration != nil {
		return fmt.Errorf("expected FSx OpeZFS Volume, found FSx ONTAP Volume: %s", d.Id())
	}

	if openzfsConfig == nil {
		return fmt.Errorf("error describing FSx OpenZFS Volume (%s): empty Openzfs configuration", d.Id())
	}

	d.Set("arn", volume.ResourceARN)
	d.Set("copy_tags_to_snapshots", openzfsConfig.CopyTagsToSnapshots)
	d.Set("data_compression_type", openzfsConfig.DataCompressionType)
	d.Set("name", volume.Name)
	d.Set("parent_volume_id", openzfsConfig.ParentVolumeId)
	d.Set("read_only", openzfsConfig.ReadOnly)
	d.Set("storage_capacity_quota_gib", openzfsConfig.StorageCapacityQuotaGiB)
	d.Set("storage_capacity_reservation_gib", openzfsConfig.StorageCapacityReservationGiB)
	d.Set("volume_type", volume.VolumeType)

	//Volume tags do not get returned with describe call so need to make a separate list tags call
	tags, tagserr := ListTags(conn, *volume.ResourceARN)

	if tagserr != nil {
		return fmt.Errorf("error reading Tags for FSx OpenZFS Volume (%s): %w", d.Id(), err)
	} else {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("origin_snapshot", flattenOpenzfsVolumeOriginSnapshot(openzfsConfig.OriginSnapshot)); err != nil {
		return fmt.Errorf("error setting nfs_exports: %w", err)
	}

	if err := d.Set("nfs_exports", flattenOpenzfsVolumeNFSExports(openzfsConfig.NfsExports)); err != nil {
		return fmt.Errorf("error setting nfs_exports: %w", err)
	}

	if err := d.Set("user_and_group_quotas", flattenOpenzfsVolumeUserAndGroupQuotas(openzfsConfig.UserAndGroupQuotas)); err != nil {
		return fmt.Errorf("error setting user_and_group_quotas: %w", err)
	}

	return nil
}

func resourceOpenzfsVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating FSx OpenZFS Volume (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	if d.HasChangesExcept("tags_all", "tags") {
		input := &fsx.UpdateVolumeInput{
			ClientRequestToken:   aws.String(resource.UniqueId()),
			VolumeId:             aws.String(d.Id()),
			OpenZFSConfiguration: &fsx.UpdateOpenZFSVolumeConfiguration{},
		}

		if d.HasChange("data_compression_type") {
			input.OpenZFSConfiguration.DataCompressionType = aws.String(d.Get("data_compression_type").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("nfs_exports") {
			input.OpenZFSConfiguration.NfsExports = expandOpenzfsVolumeNFSExports(d.Get("nfs_exports").([]interface{}))
		}

		if d.HasChange("read_only") {
			input.OpenZFSConfiguration.ReadOnly = aws.Bool(d.Get("read_only").(bool))
		}

		if d.HasChange("storage_capacity_quota_gib") {
			input.OpenZFSConfiguration.StorageCapacityQuotaGiB = aws.Int64(int64(d.Get("storage_capacity_quota_gib").(int)))
		}

		if d.HasChange("storage_capacity_reservation_gib") {
			input.OpenZFSConfiguration.StorageCapacityReservationGiB = aws.Int64(int64(d.Get("storage_capacity_reservation_gib").(int)))
		}

		if d.HasChange("user_and_group_quotas") {
			input.OpenZFSConfiguration.UserAndGroupQuotas = expandOpenzfsVolumeUserAndGroupQuotas(d.Get("user_and_group_quotas").(*schema.Set).List())
		}

		_, err := conn.UpdateVolume(input)

		if err != nil {
			return fmt.Errorf("error updating FSx OpenZFS Volume (%s): %w", d.Id(), err)
		}

		if _, err := waitVolumeUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for FSx OpenZFS Volume (%s) update: %w", d.Id(), err)
		}

	}

	return resourceOpenzfsVolumeRead(d, meta)
}

func resourceOpenzfsVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FSxConn

	log.Printf("[DEBUG] Deleting FSx OpenZFS Volume: %s", d.Id())
	_, err := conn.DeleteVolume(&fsx.DeleteVolumeInput{
		VolumeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting FSx OpenZFS Volume (%s): %w", d.Id(), err)
	}

	if _, err := waitVolumeDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for FSx OpenZFS Volume (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandOpenzfsVolumeUserAndGroupQuotas(cfg []interface{}) []*fsx.OpenZFSUserOrGroupQuota {
	quotas := []*fsx.OpenZFSUserOrGroupQuota{}

	for _, quota := range cfg {
		expandedQuota := expandOpenzfsVolumeUserAndGroupQuota(quota.(map[string]interface{}))
		if expandedQuota != nil {
			quotas = append(quotas, expandedQuota)
		}
	}

	return quotas

}

func expandOpenzfsVolumeUserAndGroupQuota(conf map[string]interface{}) *fsx.OpenZFSUserOrGroupQuota {
	if len(conf) < 1 {
		return nil
	}

	out := fsx.OpenZFSUserOrGroupQuota{}

	if v, ok := conf["id"].(int); ok {
		out.Id = aws.Int64(int64(v))
	}

	if v, ok := conf["storage_capacity_quota_gib"].(int); ok {
		out.StorageCapacityQuotaGiB = aws.Int64(int64(v))
	}

	if v, ok := conf["type"].(string); ok {
		out.Type = aws.String(v)
	}

	return &out

}

func expandOpenzfsVolumeNFSExports(cfg []interface{}) []*fsx.OpenZFSNfsExport {
	exports := []*fsx.OpenZFSNfsExport{}

	for _, export := range cfg {
		expandedExport := expandOpenzfsVolumeNFSExport(export.(map[string]interface{}))
		if expandedExport != nil {
			exports = append(exports, expandedExport)
		}
	}

	return exports

}

func expandOpenzfsVolumeNFSExport(cfg map[string]interface{}) *fsx.OpenZFSNfsExport {
	out := fsx.OpenZFSNfsExport{}

	if v, ok := cfg["client_configurations"]; ok {
		out.ClientConfigurations = expandOpenzfsVolumeClinetConfigurations(v.(*schema.Set).List())
	}

	return &out
}

func expandOpenzfsVolumeClinetConfigurations(cfg []interface{}) []*fsx.OpenZFSClientConfiguration {
	configurations := []*fsx.OpenZFSClientConfiguration{}

	for _, configuration := range cfg {
		expandedConfiguration := expandOpenzfsVolumeClientConfiguration(configuration.(map[string]interface{}))
		if expandedConfiguration != nil {
			configurations = append(configurations, expandedConfiguration)
		}
	}

	return configurations

}

func expandOpenzfsVolumeClientConfiguration(conf map[string]interface{}) *fsx.OpenZFSClientConfiguration {
	out := fsx.OpenZFSClientConfiguration{}

	if v, ok := conf["clients"].(string); ok && len(v) > 0 {
		out.Clients = aws.String(v)
	}

	if v, ok := conf["options"].([]interface{}); ok {
		out.Options = flex.ExpandStringList(v)
	}

	return &out
}

func expandOpenzfsCreateVolumeOriginSnapshot(cfg []interface{}) *fsx.CreateOpenZFSOriginSnapshotConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.CreateOpenZFSOriginSnapshotConfiguration{}

	if v, ok := conf["copy_strategy"].(string); ok {
		out.CopyStrategy = aws.String(v)
	}

	if v, ok := conf["snapshot_arn"].(string); ok {
		out.SnapshotARN = aws.String(v)
	}

	return &out
}

func flattenOpenzfsVolumeNFSExports(rs []*fsx.OpenZFSNfsExport) []map[string]interface{} {
	exports := make([]map[string]interface{}, 0)

	for _, export := range rs {
		if export != nil {
			cfg := make(map[string]interface{})
			cfg["client_configurations"] = flattenOpenzfsVolumeClientConfigurations(export.ClientConfigurations)
			exports = append(exports, cfg)
		}
	}

	if len(exports) > 0 {
		return exports
	}

	return nil
}

func flattenOpenzfsVolumeClientConfigurations(rs []*fsx.OpenZFSClientConfiguration) []map[string]interface{} {
	configurations := make([]map[string]interface{}, 0)

	for _, configuration := range rs {
		if configuration != nil {
			cfg := make(map[string]interface{})
			cfg["clients"] = aws.StringValue(configuration.Clients)
			cfg["options"] = flex.FlattenStringList(configuration.Options)
			configurations = append(configurations, cfg)
		}
	}

	if len(configurations) > 0 {
		return configurations
	}

	return nil
}

func flattenOpenzfsVolumeUserAndGroupQuotas(rs []*fsx.OpenZFSUserOrGroupQuota) []map[string]interface{} {
	quotas := make([]map[string]interface{}, 0)

	for _, quota := range rs {
		if quota != nil {
			cfg := make(map[string]interface{})
			cfg["id"] = aws.Int64Value(quota.Id)
			cfg["storage_capacity_quota_gib"] = aws.Int64Value(quota.StorageCapacityQuotaGiB)
			cfg["type"] = aws.StringValue(quota.Type)
			quotas = append(quotas, cfg)
		}
	}

	if len(quotas) > 0 {
		return quotas
	}

	return nil
}

func flattenOpenzfsVolumeOriginSnapshot(rs *fsx.OpenZFSOriginSnapshotConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.CopyStrategy != nil {
		m["copy_strategy"] = aws.StringValue(rs.CopyStrategy)
	}
	if rs.SnapshotARN != nil {
		m["snapshot_arn"] = aws.StringValue(rs.SnapshotARN)
	}

	return []interface{}{m}
}
