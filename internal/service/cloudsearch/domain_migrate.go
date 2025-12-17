// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// domainSchemaV0 returns the schema for version 0 (SDKv2 state)
func domainSchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"document_service_endpoint": schema.StringAttribute{
				Computed: true,
			},
			"domain_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			"multi_az": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"search_service_endpoint": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"endpoint_options": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enforce_https": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"tls_security_policy": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"index_field": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"analysis_scheme": schema.StringAttribute{
							Optional: true,
						},
						names.AttrDefaultValue: schema.StringAttribute{
							Optional: true,
						},
						"facet": schema.BoolAttribute{
							Optional: true,
						},
						"highlight": schema.BoolAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"return": schema.BoolAttribute{
							Optional: true,
						},
						"search": schema.BoolAttribute{
							Optional: true,
						},
						"sort": schema.BoolAttribute{
							Optional: true,
						},
						"source_fields": schema.StringAttribute{
							Optional: true,
						},
						names.AttrType: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"scaling_parameters": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"desired_instance_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"desired_partition_count": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
						"desired_replication_count": schema.Int64Attribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// V0 model types for state upgrade

type domainResourceModelV0 struct {
	ARN                     types.String `tfsdk:"arn"`
	DocumentServiceEndpoint types.String `tfsdk:"document_service_endpoint"`
	DomainID                types.String `tfsdk:"domain_id"`
	EndpointOptions         types.List   `tfsdk:"endpoint_options"`
	ID                      types.String `tfsdk:"id"`
	IndexFields             types.Set    `tfsdk:"index_field"`
	MultiAZ                 types.Bool   `tfsdk:"multi_az"`
	Name                    types.String `tfsdk:"name"`
	ScalingParameters       types.List   `tfsdk:"scaling_parameters"`
	SearchServiceEndpoint   types.String `tfsdk:"search_service_endpoint"`
}

type endpointOptionsModelV0 struct {
	EnforceHTTPS      types.Bool   `tfsdk:"enforce_https"`
	TLSSecurityPolicy types.String `tfsdk:"tls_security_policy"`
}

type indexFieldModelV0 struct {
	AnalysisScheme types.String `tfsdk:"analysis_scheme"`
	DefaultValue   types.String `tfsdk:"default_value"`
	Facet          types.Bool   `tfsdk:"facet"`
	Highlight      types.Bool   `tfsdk:"highlight"`
	Name           types.String `tfsdk:"name"`
	Return         types.Bool   `tfsdk:"return"`
	Search         types.Bool   `tfsdk:"search"`
	Sort           types.Bool   `tfsdk:"sort"`
	SourceFields   types.String `tfsdk:"source_fields"`
	Type           types.String `tfsdk:"type"`
}

type scalingParametersModelV0 struct {
	DesiredInstanceType     types.String `tfsdk:"desired_instance_type"`
	DesiredPartitionCount   types.Int64  `tfsdk:"desired_partition_count"`
	DesiredReplicationCount types.Int64  `tfsdk:"desired_replication_count"`
}

// upgradeDomainStateFromV0 upgrades state from version 0 (SDKv2) to version 1 (Framework)
func upgradeDomainStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var old domainResourceModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Convert endpoint options
	var endpointOptions fwtypes.ListNestedObjectValueOf[endpointOptionsModel]
	if !old.EndpointOptions.IsNull() && !old.EndpointOptions.IsUnknown() {
		var oldEndpointOptions []endpointOptionsModelV0
		response.Diagnostics.Append(old.EndpointOptions.ElementsAs(ctx, &oldEndpointOptions, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		if len(oldEndpointOptions) > 0 {
			newEndpointOpts := &endpointOptionsModel{
				EnforceHTTPS: oldEndpointOptions[0].EnforceHTTPS,
			}

			// Convert TLSSecurityPolicy from string to enum
			if !oldEndpointOptions[0].TLSSecurityPolicy.IsNull() && !oldEndpointOptions[0].TLSSecurityPolicy.IsUnknown() {
				newEndpointOpts.TLSSecurityPolicy = fwtypes.StringEnumValue(awstypes.TLSSecurityPolicy(oldEndpointOptions[0].TLSSecurityPolicy.ValueString()))
			}

			endpointOptions = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newEndpointOpts)
		} else {
			endpointOptions = fwtypes.NewListNestedObjectValueOfNull[endpointOptionsModel](ctx)
		}
	} else {
		endpointOptions = fwtypes.NewListNestedObjectValueOfNull[endpointOptionsModel](ctx)
	}

	// Convert scaling parameters
	var scalingParameters fwtypes.ListNestedObjectValueOf[scalingParametersModel]
	if !old.ScalingParameters.IsNull() && !old.ScalingParameters.IsUnknown() {
		var oldScalingParams []scalingParametersModelV0
		response.Diagnostics.Append(old.ScalingParameters.ElementsAs(ctx, &oldScalingParams, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		if len(oldScalingParams) > 0 {
			newScalingParams := &scalingParametersModel{
				DesiredPartitionCount:   oldScalingParams[0].DesiredPartitionCount,
				DesiredReplicationCount: oldScalingParams[0].DesiredReplicationCount,
			}

			// Convert DesiredInstanceType from string to enum
			if !oldScalingParams[0].DesiredInstanceType.IsNull() && !oldScalingParams[0].DesiredInstanceType.IsUnknown() {
				newScalingParams.DesiredInstanceType = fwtypes.StringEnumValue(awstypes.PartitionInstanceType(oldScalingParams[0].DesiredInstanceType.ValueString()))
			}

			scalingParameters = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, newScalingParams)
		} else {
			scalingParameters = fwtypes.NewListNestedObjectValueOfNull[scalingParametersModel](ctx)
		}
	} else {
		scalingParameters = fwtypes.NewListNestedObjectValueOfNull[scalingParametersModel](ctx)
	}

	// Convert index fields
	var indexFields fwtypes.SetNestedObjectValueOf[indexFieldModel]
	if !old.IndexFields.IsNull() && !old.IndexFields.IsUnknown() {
		var oldIndexFields []indexFieldModelV0
		response.Diagnostics.Append(old.IndexFields.ElementsAs(ctx, &oldIndexFields, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		if len(oldIndexFields) > 0 {
			var newIndexFields []*indexFieldModel
			for _, oldField := range oldIndexFields {
				newField := &indexFieldModel{
					AnalysisScheme: oldField.AnalysisScheme,
					DefaultValue:   oldField.DefaultValue,
					Facet:          oldField.Facet,
					Highlight:      oldField.Highlight,
					Name:           oldField.Name,
					Return:         oldField.Return,
					Search:         oldField.Search,
					Sort:           oldField.Sort,
					SourceFields:   oldField.SourceFields,
				}

				// Convert Type from string to enum
				if !oldField.Type.IsNull() && !oldField.Type.IsUnknown() {
					newField.Type = fwtypes.StringEnumValue(awstypes.IndexFieldType(oldField.Type.ValueString()))
				}

				newIndexFields = append(newIndexFields, newField)
			}

			var d diag.Diagnostics
			indexFields, d = fwtypes.NewSetNestedObjectValueOfSlice(ctx, newIndexFields, indexFieldSemanticEquality)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
		} else {
			indexFields = fwtypes.NewSetNestedObjectValueOfNull[indexFieldModel](ctx)
		}
	} else {
		indexFields = fwtypes.NewSetNestedObjectValueOfNull[indexFieldModel](ctx)
	}

	new := domainResourceModel{
		ARN:                     old.ARN,
		DocumentServiceEndpoint: old.DocumentServiceEndpoint,
		DomainID:                old.DomainID,
		EndpointOptions:         endpointOptions,
		ID:                      old.ID,
		IndexFields:             indexFields,
		MultiAZ:                 old.MultiAZ,
		Name:                    old.Name,
		ScalingParameters:       scalingParameters,
		SearchServiceEndpoint:   old.SearchServiceEndpoint,
		Timeouts:                timeouts.Value{},
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}
