// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflextypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_global_clusters", name="Global Clusters")
func newDataSourceGlobalClusters(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceGlobalClusters{}, nil
}

type dataSourceGlobalClusters struct {
	framework.DataSourceWithModel[dataSourceGlobalClustersModel]
}

func (d *dataSourceGlobalClusters) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"global_cluster_identifiers": schema.ListAttribute{
				CustomType:  fwflextypes.ListOfStringType,
				ElementType: fwtypes.StringType,
				Computed:    true,
			},
			"global_cluster_arns": schema.ListAttribute{
				CustomType:  fwflextypes.ListOfStringType,
				ElementType: fwtypes.StringType,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.ListAttribute{
							Required:    true,
							ElementType: fwtypes.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceGlobalClusters) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceGlobalClustersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().RDSClient(ctx)
	input := &rds.DescribeGlobalClustersInput{}

	var filters []globalClustersFilterModel
	resp.Diagnostics.Append(data.Filter.ElementsAs(ctx, &filters, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	predicate := globalClustersPredicate(filters)

	clusters, err := findGlobalClusters(ctx, conn, input, predicate)
	if err != nil {
		resp.Diagnostics.AddError("listing RDS Global Clusters", err.Error())
		return
	}

	var identifiers []string
	var arns []string
	for _, cluster := range clusters {
		identifiers = append(identifiers, aws.ToString(cluster.GlobalClusterIdentifier))
		arns = append(arns, aws.ToString(cluster.GlobalClusterArn))
	}

	data.GlobalClusterIdentifiers = fwflex.FlattenFrameworkStringValueListOfString(ctx, identifiers)
	data.GlobalClusterArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, arns)
	data.ID = fwtypes.StringValue(d.Meta().Region(ctx))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func globalClustersPredicate(filters []globalClustersFilterModel) tfslices.Predicate[*types.GlobalCluster] {
	if len(filters) == 0 {
		return tfslices.PredicateTrue[*types.GlobalCluster]()
	}

	return func(cluster *types.GlobalCluster) bool {
		for _, f := range filters {
			name := f.Name.ValueString()

			var values []string
			for _, v := range f.Values.Elements() {
				values = append(values, v.(fwtypes.String).ValueString())
			}

			var fieldValue string
			switch name {
			case "engine":
				fieldValue = aws.ToString(cluster.Engine)
			case "engine_version":
				fieldValue = aws.ToString(cluster.EngineVersion)
			case "status":
				fieldValue = aws.ToString(cluster.Status)
			case "global_cluster_identifier":
				fieldValue = aws.ToString(cluster.GlobalClusterIdentifier)
			case "database_name":
				fieldValue = aws.ToString(cluster.DatabaseName)
			case "storage_encrypted":
				if aws.ToBool(cluster.StorageEncrypted) {
					fieldValue = "true"
				} else {
					fieldValue = "false"
				}
			case "deletion_protection":
				if aws.ToBool(cluster.DeletionProtection) {
					fieldValue = "true"
				} else {
					fieldValue = "false"
				}
			default:
				return false
			}

			if !slices.Contains(values, fieldValue) {
				return false
			}
		}
		return true
	}
}

type dataSourceGlobalClustersModel struct {
	framework.WithRegionModel
	ID                       fwtypes.String       `tfsdk:"id"`
	Filter                   fwtypes.Set          `tfsdk:"filter"`
	GlobalClusterIdentifiers fwflextypes.ListOfString `tfsdk:"global_cluster_identifiers"`
	GlobalClusterArns        fwflextypes.ListOfString `tfsdk:"global_cluster_arns"`
}

type globalClustersFilterModel struct {
	Name   fwtypes.String `tfsdk:"name"`
	Values fwtypes.List   `tfsdk:"values"`
}
