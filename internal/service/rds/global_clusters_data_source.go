// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_global_clusters", name="Global Clusters")
func newGlobalClustersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &globalClustersDataSource{}, nil
}

type globalClustersDataSource struct {
	framework.DataSourceWithModel[globalClustersDataSourceModel]
}

func (d *globalClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"global_cluster_identifiers": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"global_cluster_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[globalClustersFilterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (d *globalClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data globalClustersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters, diags := data.Filter.ToSlice(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate filter names up front so a typo surfaces an error instead of
	// silently returning no results.
	for _, f := range filters {
		if _, ok := globalClusterFilterFields[f.Name.ValueString()]; !ok {
			resp.Diagnostics.AddError(
				"Invalid Filter Name",
				fmt.Sprintf("%q is not a supported filter name. Valid names are: %s.", f.Name.ValueString(), strings.Join(globalClusterFilterNames(), ", ")),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().RDSClient(ctx)
	input := &rds.DescribeGlobalClustersInput{}

	clusters, err := findGlobalClusters(ctx, conn, input, globalClustersPredicate(ctx, filters))
	if err != nil {
		resp.Diagnostics.AddError("listing RDS Global Clusters", err.Error())
		return
	}

	identifiers := tfslices.ApplyToAll(clusters, func(c awstypes.GlobalCluster) string {
		return aws.ToString(c.GlobalClusterIdentifier)
	})
	arns := tfslices.ApplyToAll(clusters, func(c awstypes.GlobalCluster) string {
		return aws.ToString(c.GlobalClusterArn)
	})

	data.GlobalClusterIdentifiers = fwflex.FlattenFrameworkStringValueListOfString(ctx, identifiers)
	data.GlobalClusterArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, arns)
	data.ID = types.StringValue(d.Meta().Region(ctx))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// globalClusterFilterFields maps a supported filter name to the corresponding
// value on a global cluster. The DescribeGlobalClusters API does not support
// server-side filtering (other than by identifier), so filtering is client-side.
var globalClusterFilterFields = map[string]func(*awstypes.GlobalCluster) string{
	names.AttrDatabaseName:       func(c *awstypes.GlobalCluster) string { return aws.ToString(c.DatabaseName) },
	names.AttrDeletionProtection: func(c *awstypes.GlobalCluster) string { return strconv.FormatBool(aws.ToBool(c.DeletionProtection)) },
	names.AttrEngine:             func(c *awstypes.GlobalCluster) string { return aws.ToString(c.Engine) },
	names.AttrEngineVersion:      func(c *awstypes.GlobalCluster) string { return aws.ToString(c.EngineVersion) },
	"global_cluster_identifier":  func(c *awstypes.GlobalCluster) string { return aws.ToString(c.GlobalClusterIdentifier) },
	names.AttrStatus:             func(c *awstypes.GlobalCluster) string { return aws.ToString(c.Status) },
	names.AttrStorageEncrypted:   func(c *awstypes.GlobalCluster) string { return strconv.FormatBool(aws.ToBool(c.StorageEncrypted)) },
}

func globalClusterFilterNames() []string {
	filterNames := make([]string, 0, len(globalClusterFilterFields))
	for name := range globalClusterFilterFields {
		filterNames = append(filterNames, name)
	}
	slices.Sort(filterNames)
	return filterNames
}

func globalClustersPredicate(ctx context.Context, filters []*globalClustersFilterModel) tfslices.Predicate[*awstypes.GlobalCluster] {
	if len(filters) == 0 {
		return tfslices.PredicateTrue[*awstypes.GlobalCluster]()
	}

	// Extract each filter's field accessor and values once, rather than per cluster.
	type compiledFilter struct {
		field  func(*awstypes.GlobalCluster) string
		values []string
	}
	compiled := make([]compiledFilter, 0, len(filters))
	for _, f := range filters {
		compiled = append(compiled, compiledFilter{
			field:  globalClusterFilterFields[f.Name.ValueString()], // name validated in Read
			values: fwflex.ExpandFrameworkStringValueList(ctx, f.Values),
		})
	}

	return func(cluster *awstypes.GlobalCluster) bool {
		for _, cf := range compiled {
			if !slices.Contains(cf.values, cf.field(cluster)) {
				return false
			}
		}
		return true
	}
}

type globalClustersDataSourceModel struct {
	framework.WithRegionModel
	ID                       types.String                                              `tfsdk:"id"`
	Filter                   fwtypes.SetNestedObjectValueOf[globalClustersFilterModel] `tfsdk:"filter"`
	GlobalClusterIdentifiers fwtypes.ListOfString                                      `tfsdk:"global_cluster_identifiers"`
	GlobalClusterArns        fwtypes.ListOfString                                      `tfsdk:"global_cluster_arns"`
}

type globalClustersFilterModel struct {
	Name   types.String         `tfsdk:"name"`
	Values fwtypes.ListOfString `tfsdk:"values"`
}
