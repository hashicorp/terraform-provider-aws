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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_ontap_file_system", name="ONTAP File System")
// @Tags(identifierAttribute="arn")
func resourceONTAPFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceONTAPFileSystemCreate,
		ReadWithoutTimeout:   resourceONTAPFileSystemRead,
		UpdateWithoutTimeout: resourceONTAPFileSystemUpdate,
		DeleteWithoutTimeout: resourceONTAPFileSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 90),
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(5, 5),
					validation.StringMatch(regexache.MustCompile(`^([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format HH:MM"),
				),
			},
			"deployment_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OntapDeploymentType](),
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIOPS: {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 2400000),
						},
						names.AttrMode: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DiskIopsConfigurationModeAutomatic,
							ValidateDiagFunc: enum.Validate[awstypes.DiskIopsConfigurationMode](),
						},
					},
				},
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_ip_address_range": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},
			names.AttrEndpoints: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intercluster": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"management": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"fsx_admin_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 50),
			},
			"ha_pairs": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 12),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"network_interface_ids": {
				// As explained in https://docs.aws.amazon.com/fsx/latest/OntapGuide/mounting-on-premises.html, the first
				// network_interface_id is the primary one, so ordering matters. Use TypeList instead of TypeSet to preserve it.
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 50,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1024, 1024*1024),
			},
			names.AttrStorageType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.StorageTypeSsd,
				ValidateDiagFunc: enum.Validate[awstypes.StorageType](),
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 2,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144}),
				ExactlyOneOf: []string{"throughput_capacity", "throughput_capacity_per_ha_pair"},
			},
			"throughput_capacity_per_ha_pair": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144}),
				ExactlyOneOf: []string{"throughput_capacity", "throughput_capacity_per_ha_pair"},
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(7, 7),
					validation.StringMatch(regexache.MustCompile(`^[1-7]:([01]\d|2[0-3]):?([0-5]\d)$`), "must be in the format d:HH:MM"),
				),
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			resourceONTAPFileSystemThroughputCapacityPerHAPairCustomizeDiff,
			resourceONTAPFileSystemHAPairsCustomizeDiff,
		),
	}
}

func resourceONTAPFileSystemThroughputCapacityPerHAPairCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	// we want to force a new resource if the throughput_capacity_per_ha_pair is increased for Gen1 file systems
	if d.HasChange("throughput_capacity_per_ha_pair") {
		o, n := d.GetChange("throughput_capacity_per_ha_pair")
		if n != nil && n.(int) != 0 && n.(int) > o.(int) && (d.Get("deployment_type").(string) == string(awstypes.OntapDeploymentTypeSingleAz1) || d.Get("deployment_type").(string) == string(awstypes.OntapDeploymentTypeMultiAz1)) {
			if err := d.ForceNew("throughput_capacity_per_ha_pair"); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceONTAPFileSystemHAPairsCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta any) error {
	// we want to force a new resource if the ha_pairs is increased for Gen1 single AZ file systems. multiple ha_pairs is not supported on Multi AZ.
	if d.HasChange("ha_pairs") {
		o, n := d.GetChange("ha_pairs")
		if n != nil && n.(int) != 0 && n.(int) > o.(int) && (d.Get("deployment_type").(string) == string(awstypes.OntapDeploymentTypeSingleAz1)) {
			if err := d.ForceNew("ha_pairs"); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceONTAPFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     awstypes.FileSystemTypeOntap,
		OntapConfiguration: &awstypes.CreateFileSystemOntapConfiguration{
			AutomaticBackupRetentionDays: aws.Int32(int32(d.Get("automatic_backup_retention_days").(int))),
			DeploymentType:               awstypes.OntapDeploymentType(d.Get("deployment_type").(string)),
			PreferredSubnetId:            aws.String(d.Get("preferred_subnet_id").(string)),
		},
		StorageCapacity: aws.Int32(int32(d.Get("storage_capacity").(int))),
		StorageType:     awstypes.StorageType(d.Get(names.AttrStorageType).(string)),
		SubnetIds:       flex.ExpandStringValueList(d.Get(names.AttrSubnetIDs).([]interface{})),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("daily_automatic_backup_start_time"); ok {
		input.OntapConfiguration.DailyAutomaticBackupStartTime = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_iops_configuration"); ok {
		input.OntapConfiguration.DiskIopsConfiguration = expandOntapFileDiskIopsConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("endpoint_ip_address_range"); ok {
		input.OntapConfiguration.EndpointIpAddressRange = aws.String(v.(string))
	}

	if v, ok := d.GetOk("fsx_admin_password"); ok {
		input.OntapConfiguration.FsxAdminPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ha_pairs"); ok {
		v := int32(v.(int))
		input.OntapConfiguration.HAPairs = aws.Int32(v)

		if v > 0 {
			if v, ok := d.GetOk("throughput_capacity_per_ha_pair"); ok {
				input.OntapConfiguration.ThroughputCapacityPerHAPair = aws.Int32(int32(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_table_ids"); ok {
		input.OntapConfiguration.RouteTableIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroupIDs); ok {
		input.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("throughput_capacity"); ok {
		input.OntapConfiguration.ThroughputCapacity = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	output, err := conn.CreateFileSystem(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for NetApp ONTAP File System: %s", err)
	}

	d.SetId(aws.ToString(output.FileSystem.FileSystemId))

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceONTAPFileSystemRead(ctx, d, meta)...)
}

func resourceONTAPFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	filesystem, err := findONTAPFileSystemByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx for NetApp ONTAP File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for NetApp ONTAP File System (%s): %s", d.Id(), err)
	}

	ontapConfig := filesystem.OntapConfiguration

	d.Set(names.AttrARN, filesystem.ResourceARN)
	d.Set("automatic_backup_retention_days", ontapConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", ontapConfig.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", ontapConfig.DeploymentType)
	if err := d.Set("disk_iops_configuration", flattenOntapFileDiskIopsConfiguration(ontapConfig.DiskIopsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting disk_iops_configuration: %s", err)
	}
	d.Set(names.AttrDNSName, filesystem.DNSName)
	d.Set("endpoint_ip_address_range", ontapConfig.EndpointIpAddressRange)
	if err := d.Set(names.AttrEndpoints, flattenOntapFileSystemEndpoints(ontapConfig.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints: %s", err)
	}
	d.Set("fsx_admin_password", d.Get("fsx_admin_password").(string))
	haPairs := aws.ToInt32(ontapConfig.HAPairs)
	d.Set("ha_pairs", haPairs)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", filesystem.NetworkInterfaceIds)
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_subnet_id", ontapConfig.PreferredSubnetId)
	d.Set("route_table_ids", ontapConfig.RouteTableIds)
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set(names.AttrStorageType, filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, filesystem.SubnetIds)
	if ontapConfig.DeploymentType == awstypes.OntapDeploymentTypeSingleAz2 {
		d.Set("throughput_capacity", nil)
		d.Set("throughput_capacity_per_ha_pair", ontapConfig.ThroughputCapacityPerHAPair)
	} else {
		d.Set("throughput_capacity", ontapConfig.ThroughputCapacity)
		d.Set("throughput_capacity_per_ha_pair", ontapConfig.ThroughputCapacityPerHAPair)
	}
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", ontapConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceONTAPFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			FileSystemId:       aws.String(d.Id()),
			OntapConfiguration: &awstypes.UpdateFileSystemOntapConfiguration{},
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.OntapConfiguration.AutomaticBackupRetentionDays = aws.Int32(int32(d.Get("automatic_backup_retention_days").(int)))
		}

		if d.HasChange("daily_automatic_backup_start_time") {
			input.OntapConfiguration.DailyAutomaticBackupStartTime = aws.String(d.Get("daily_automatic_backup_start_time").(string))
		}

		if d.HasChange("disk_iops_configuration") {
			input.OntapConfiguration.DiskIopsConfiguration = expandOntapFileDiskIopsConfiguration(d.Get("disk_iops_configuration").([]interface{}))
		}

		if d.HasChange("fsx_admin_password") {
			input.OntapConfiguration.FsxAdminPassword = aws.String(d.Get("fsx_admin_password").(string))
		}

		if d.HasChange("ha_pairs") {
			input.OntapConfiguration.HAPairs = aws.Int32(int32(d.Get("ha_pairs").(int)))
			//for the ONTAP update API the ThroughputCapacityPerHAPair must explicitly be passed when adding ha_pairs even if it hasn't changed.
			input.OntapConfiguration.ThroughputCapacityPerHAPair = aws.Int32(int32(d.Get("throughput_capacity_per_ha_pair").(int)))
		}

		if d.HasChange("route_table_ids") {
			o, n := d.GetChange("route_table_ids")
			os, ns := o.(*schema.Set), n.(*schema.Set)
			add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

			if len(add) > 0 {
				input.OntapConfiguration.AddRouteTableIds = add
			}

			if len(del) > 0 {
				input.OntapConfiguration.RemoveRouteTableIds = del
			}
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int32(int32(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.OntapConfiguration.ThroughputCapacity = aws.Int32(int32(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("throughput_capacity_per_ha_pair") {
			input.OntapConfiguration.ThroughputCapacityPerHAPair = aws.Int32(int32(d.Get("throughput_capacity_per_ha_pair").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystem(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for NetApp ONTAP File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) administrative action (%s) complete: %s", d.Id(), awstypes.AdministrativeActionTypeFileSystemUpdate, err)
		}
	}

	return append(diags, resourceONTAPFileSystemRead(ctx, d, meta)...)
}

func resourceONTAPFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	log.Printf("[DEBUG] Deleting FSx for NetApp ONTAP File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(ctx, &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx for NetApp ONTAP File System (%s): %s", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandOntapFileDiskIopsConfiguration(cfg []interface{}) *awstypes.DiskIopsConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := awstypes.DiskIopsConfiguration{}

	if v, ok := conf[names.AttrMode].(string); ok && len(v) > 0 {
		out.Mode = awstypes.DiskIopsConfigurationMode(v)
	}
	if v, ok := conf[names.AttrIOPS].(int); ok {
		out.Iops = aws.Int64(int64(v))
	}

	return &out
}

func flattenOntapFileDiskIopsConfiguration(rs *awstypes.DiskIopsConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m[names.AttrMode] = string(rs.Mode)

	if rs.Iops != nil {
		m[names.AttrIOPS] = aws.ToInt64(rs.Iops)
	}

	return []interface{}{m}
}

func flattenOntapFileSystemEndpoints(rs *awstypes.FileSystemEndpoints) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Intercluster != nil {
		m["intercluster"] = flattenOntapFileSystemEndpoint(rs.Intercluster)
	}
	if rs.Management != nil {
		m["management"] = flattenOntapFileSystemEndpoint(rs.Management)
	}

	return []interface{}{m}
}

func flattenOntapFileSystemEndpoint(rs *awstypes.FileSystemEndpoint) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DNSName != nil {
		m[names.AttrDNSName] = aws.ToString(rs.DNSName)
	}
	if rs.IpAddresses != nil {
		m[names.AttrIPAddresses] = flex.FlattenStringValueSet(rs.IpAddresses)
	}

	return []interface{}{m}
}

func findONTAPFileSystemByID(ctx context.Context, conn *fsx.Client, id string) (*awstypes.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, awstypes.FileSystemTypeOntap)

	if err != nil {
		return nil, err
	}

	if output.OntapConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
