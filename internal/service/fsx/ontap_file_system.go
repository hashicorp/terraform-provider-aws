// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.OntapDeploymentType_Values(), false),
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(0, 2400000),
						},
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      fsx.DiskIopsConfigurationModeAutomatic,
							ValidateFunc: validation.StringInSlice(fsx.DiskIopsConfigurationMode_Values(), false),
						},
					},
				},
			},
			"dns_name": {
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
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intercluster": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
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
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
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
				ForceNew:     true,
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
			"security_group_ids": {
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
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      fsx.StorageTypeSsd,
				ValidateFunc: validation.StringInSlice(fsx.StorageType_Values(), false),
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
				ValidateFunc: validation.IntInSlice([]int{128, 256, 512, 1024, 2048, 4096}),
				ExactlyOneOf: []string{"throughput_capacity", "throughput_capacity_per_ha_pair"},
			},
			"throughput_capacity_per_ha_pair": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{128, 256, 512, 1024, 2048, 3072, 4096, 6144}),
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
			customdiff.ForceNewIfChange("throughput_capacity_per_ha_pair", func(ctx context.Context, old, new, meta any) bool {
				if new != nil && new != 0 {
					return new.(int) != old.(int)
				} else {
					return false
				}
			}),
		),
	}
}

func resourceONTAPFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.CreateFileSystemInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		FileSystemType:     aws.String(fsx.FileSystemTypeOntap),
		OntapConfiguration: &fsx.CreateFileSystemOntapConfiguration{
			AutomaticBackupRetentionDays: aws.Int64(int64(d.Get("automatic_backup_retention_days").(int))),
			DeploymentType:               aws.String(d.Get("deployment_type").(string)),
			PreferredSubnetId:            aws.String(d.Get("preferred_subnet_id").(string)),
		},
		StorageCapacity: aws.Int64(int64(d.Get("storage_capacity").(int))),
		StorageType:     aws.String(d.Get("storage_type").(string)),
		SubnetIds:       flex.ExpandStringList(d.Get(names.AttrSubnetIDs).([]interface{})),
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
		v := int64(v.(int))
		input.OntapConfiguration.HAPairs = aws.Int64(v)

		if v > 0 {
			if v, ok := d.GetOk("throughput_capacity_per_ha_pair"); ok {
				input.OntapConfiguration.ThroughputCapacityPerHAPair = aws.Int64(int64(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_table_ids"); ok {
		input.OntapConfiguration.RouteTableIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("security_group_ids"); ok {
		input.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("throughput_capacity"); ok {
		input.OntapConfiguration.ThroughputCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("weekly_maintenance_start_time"); ok {
		input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(v.(string))
	}

	output, err := conn.CreateFileSystemWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx for NetApp ONTAP File System: %s", err)
	}

	d.SetId(aws.StringValue(output.FileSystem.FileSystemId))

	if _, err := waitFileSystemCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceONTAPFileSystemRead(ctx, d, meta)...)
}

func resourceONTAPFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

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
	d.Set("dns_name", filesystem.DNSName)
	d.Set("endpoint_ip_address_range", ontapConfig.EndpointIpAddressRange)
	if err := d.Set("endpoints", flattenOntapFileSystemEndpoints(ontapConfig.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints: %s", err)
	}
	d.Set("fsx_admin_password", d.Get("fsx_admin_password").(string))
	haPairs := aws.Int64Value(ontapConfig.HAPairs)
	d.Set("ha_pairs", haPairs)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds))
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_subnet_id", ontapConfig.PreferredSubnetId)
	d.Set("route_table_ids", aws.StringValueSlice(ontapConfig.RouteTableIds))
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, aws.StringValueSlice(filesystem.SubnetIds))
	if aws.StringValue(ontapConfig.DeploymentType) == fsx.OntapDeploymentTypeSingleAz2 {
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
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateFileSystemInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			FileSystemId:       aws.String(d.Id()),
			OntapConfiguration: &fsx.UpdateFileSystemOntapConfiguration{},
		}

		if d.HasChange("automatic_backup_retention_days") {
			input.OntapConfiguration.AutomaticBackupRetentionDays = aws.Int64(int64(d.Get("automatic_backup_retention_days").(int)))
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

		if d.HasChange("route_table_ids") {
			o, n := d.GetChange("route_table_ids")
			os, ns := o.(*schema.Set), n.(*schema.Set)
			add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

			if len(add) > 0 {
				input.OntapConfiguration.AddRouteTableIds = aws.StringSlice(add)
			}

			if len(del) > 0 {
				input.OntapConfiguration.RemoveRouteTableIds = aws.StringSlice(del)
			}
		}

		if d.HasChange("storage_capacity") {
			input.StorageCapacity = aws.Int64(int64(d.Get("storage_capacity").(int)))
		}

		if d.HasChange("throughput_capacity") {
			input.OntapConfiguration.ThroughputCapacity = aws.Int64(int64(d.Get("throughput_capacity").(int)))
		}

		if d.HasChange("weekly_maintenance_start_time") {
			input.OntapConfiguration.WeeklyMaintenanceStartTime = aws.String(d.Get("weekly_maintenance_start_time").(string))
		}

		startTime := time.Now()
		_, err := conn.UpdateFileSystemWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx for NetApp ONTAP File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemUpdated(ctx, conn, d.Id(), startTime, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) update: %s", d.Id(), err)
		}

		if _, err := waitFileSystemAdministrativeActionCompleted(ctx, conn, d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx for NetApp ONTAP File System (%s) administrative action (%s) complete: %s", d.Id(), fsx.AdministrativeActionTypeFileSystemUpdate, err)
		}
	}

	return append(diags, resourceONTAPFileSystemRead(ctx, d, meta)...)
}

func resourceONTAPFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx for NetApp ONTAP File System: %s", d.Id())
	_, err := conn.DeleteFileSystemWithContext(ctx, &fsx.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
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

func expandOntapFileDiskIopsConfiguration(cfg []interface{}) *fsx.DiskIopsConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.DiskIopsConfiguration{}

	if v, ok := conf["mode"].(string); ok && len(v) > 0 {
		out.Mode = aws.String(v)
	}
	if v, ok := conf["iops"].(int); ok {
		out.Iops = aws.Int64(int64(v))
	}

	return &out
}

func flattenOntapFileDiskIopsConfiguration(rs *fsx.DiskIopsConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Mode != nil {
		m["mode"] = aws.StringValue(rs.Mode)
	}
	if rs.Iops != nil {
		m["iops"] = aws.Int64Value(rs.Iops)
	}

	return []interface{}{m}
}

func flattenOntapFileSystemEndpoints(rs *fsx.FileSystemEndpoints) []interface{} {
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

func flattenOntapFileSystemEndpoint(rs *fsx.FileSystemEndpoint) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DNSName != nil {
		m["dns_name"] = aws.StringValue(rs.DNSName)
	}
	if rs.IpAddresses != nil {
		m["ip_addresses"] = flex.FlattenStringSet(rs.IpAddresses)
	}

	return []interface{}{m}
}

func findONTAPFileSystemByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	output, err := findFileSystemByIDAndType(ctx, conn, id, fsx.FileSystemTypeOntap)

	if err != nil {
		return nil, err
	}

	if output.OntapConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output, nil
}
