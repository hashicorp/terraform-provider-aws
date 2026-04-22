// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package fsx

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/fsx/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline. Improve this data source's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_fsx_lustre_file_system", name="Lustre File System")
func newLustreFileSystemDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &lustreFileSystemDataSource{}, nil
}

const (
	DSNameLustreFileSystem = "Lustre File System Data Source"
)

type lustreFileSystemDataSource struct {
	framework.DataSourceWithModel[lustreFileSystemDataSourceModel]
}

func (d *lustreFileSystemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_import_policy": {
				Type:             schema.TypeString,
				Computed: true,
			},
			"automatic_backup_retention_days": {
				Type:         schema.TypeInt,
				Computed: true,
			},
			"backup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags_to_backups": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_compression_type": {
				Type:             schema.TypeString,
				Computed: true,
			},
			"data_read_cache_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"sizing_mode": {
							Type:             schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"deployment_type": {
				Type:             schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"drive_cache_type": {
				Type:             schema.TypeString,
				Computed: true,
			},
			"efa_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"export_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_type_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_backup_tags": tftags.TagsSchemaComputed(),
			"imported_file_chunk_size": {
				Type:         schema.TypeInt,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Computed: true,
			},
			"log_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDestination: {
							Type:         schema.TypeString,
							Computed:     true,
						"level": {
							Type:             schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"metadata_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:             schema.TypeString,
							Computed:         true,
						},
						names.AttrIOPS: {
							Type:             schema.TypeInt,
							Computed:         true,
						},
					},
				},
			},
			"mount_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				// As explained in https://docs.aws.amazon.com/fsx/latest/LustreGuide/mounting-on-premises.html, the first
				// network_interface_id is the primary one, so ordering matters. Use TypeList instead of TypeSet to preserve it.
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"per_unit_storage_throughput": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"root_squash_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"no_squash_nids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
							},
						},
						"root_squash": {
							Type:         schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"skip_final_backup": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"storage_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func (d *lustreFileSystemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().FSxClient(ctx)

	// TIP: -- 2. Fetch the config
	var data lustreFileSystemDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get information about a resource from AWS
	out, err := findLustreFileSystemByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	// TIP: -- 4. Set the ID, arguments, and attributes
	// Using a field name prefix allows mapping fields such as `LustreFileSystemId` to `ID`
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("LustreFileSystem")), smerr.ID, data.Name.String())
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 5. Set the tags

	// TIP: -- 6. Set the state
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.Name.String())
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type lustreFileSystemDataSourceModel struct {
	framework.WithRegionModel
	ARN     types.String `tfsdk:"arn"`
	ID      types.String `tfsdk:"id"`
	DNSName types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
}
