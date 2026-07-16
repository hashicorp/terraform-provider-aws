// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_network_insights_access_scope", name="Network Insights Access Scope")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @ArnFormat("network-insights-access-scope/{id}", attribute="arn")
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
func newNetworkInsightsAccessScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &networkInsightsAccessScopeResource{}, nil
}

type networkInsightsAccessScopeResource struct {
	framework.ResourceWithModel[networkInsightsAccessScopeResourceModel]
	framework.WithImportByIdentity
}

func (r *networkInsightsAccessScopeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceStatementBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Validators: []validator.Object{
				tfobjectvalidator.ExactlyOneOfChildren(
					path.MatchRelative().AtName(names.AttrResources),
					path.MatchRelative().AtName("resource_types"),
				),
			},
			Attributes: map[string]schema.Attribute{
				names.AttrResources: schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					Optional:    true,
					ElementType: types.StringType,
				},
				"resource_types": schema.ListAttribute{
					CustomType:  fwtypes.ListOfStringType,
					Optional:    true,
					ElementType: types.StringType,
				},
			},
		},
	}

	pathStatementBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[pathStatementModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"packet_header_statement": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[packetHeaderStatementModel](ctx),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"source_addresses": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"destination_addresses": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"source_ports": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"destination_ports": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"source_prefix_lists": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"destination_prefix_lists": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
							"protocols": schema.ListAttribute{
								CustomType:  fwtypes.ListOfStringType,
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
				"resource_statement": resourceStatementBlock,
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"match_paths": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[matchPathModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrSource:      pathStatementBlock,
						names.AttrDestination: pathStatementBlock,
					},
				},
			},
			"exclude_paths": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[excludePathModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrSource:      pathStatementBlock,
						names.AttrDestination: pathStatementBlock,
						"through_resources": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[throughResourcesStatementModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"resource_statement": resourceStatementBlock,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *networkInsightsAccessScopeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data networkInsightsAccessScopeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	var input ec2.CreateNetworkInsightsAccessScopeInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeNetworkInsightsAccessScope)

	output, err := conn.CreateNetworkInsightsAccessScope(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating EC2 Network Insights Access Scope", err.Error())
		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.NetworkInsightsAccessScope, &data, fwflex.WithFieldNamePrefix("NetworkInsightsAccessScope"))...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *networkInsightsAccessScopeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data networkInsightsAccessScopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.ID.ValueString()

	scope, err := findNetworkInsightsAccessScopeByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Network Insights Access Scope (%s)", id), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, scope, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *networkInsightsAccessScopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data networkInsightsAccessScopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	id := data.ID.ValueString()
	input := ec2.DeleteNetworkInsightsAccessScopeInput{
		NetworkInsightsAccessScopeId: aws.String(id),
	}
	_, err := conn.DeleteNetworkInsightsAccessScope(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAccessScopeIdNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Network Insights Access Scope (%s)", id), err.Error())
	}
}

func findNetworkInsightsAccessScopes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsAccessScopesInput) ([]awstypes.NetworkInsightsAccessScope, error) {
	var output []awstypes.NetworkInsightsAccessScope

	pages := ec2.NewDescribeNetworkInsightsAccessScopesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAccessScopeIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NetworkInsightsAccessScopes...)
	}

	return output, nil
}

func findNetworkInsightsAccessScope(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsAccessScopesInput) (*awstypes.NetworkInsightsAccessScope, error) {
	output, err := findNetworkInsightsAccessScopes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkInsightsAccessScopeByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInsightsAccessScope, error) {
	input := ec2.DescribeNetworkInsightsAccessScopesInput{
		NetworkInsightsAccessScopeIds: []string{id},
	}

	output, err := findNetworkInsightsAccessScope(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.NetworkInsightsAccessScopeId) != id {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

type networkInsightsAccessScopeResourceModel struct {
	framework.WithRegionModel
	ARN          types.String                                      `tfsdk:"arn"`
	ExcludePaths fwtypes.ListNestedObjectValueOf[excludePathModel] `tfsdk:"exclude_paths"`
	ID           types.String                                      `tfsdk:"id"`
	MatchPaths   fwtypes.ListNestedObjectValueOf[matchPathModel]   `tfsdk:"match_paths"`
	Tags         tftags.Map                                        `tfsdk:"tags"`
	TagsAll      tftags.Map                                        `tfsdk:"tags_all"`
}

type matchPathModel struct {
	Source      fwtypes.ListNestedObjectValueOf[pathStatementModel] `tfsdk:"source"`
	Destination fwtypes.ListNestedObjectValueOf[pathStatementModel] `tfsdk:"destination"`
}

type excludePathModel struct {
	Source           fwtypes.ListNestedObjectValueOf[pathStatementModel]             `tfsdk:"source"`
	Destination      fwtypes.ListNestedObjectValueOf[pathStatementModel]             `tfsdk:"destination"`
	ThroughResources fwtypes.ListNestedObjectValueOf[throughResourcesStatementModel] `tfsdk:"through_resources"`
}

type pathStatementModel struct {
	PacketHeaderStatement fwtypes.ListNestedObjectValueOf[packetHeaderStatementModel] `tfsdk:"packet_header_statement"`
	ResourceStatement     fwtypes.ListNestedObjectValueOf[resourceStatementModel]     `tfsdk:"resource_statement"`
}

type packetHeaderStatementModel struct {
	SourceAddresses        fwtypes.ListOfString `tfsdk:"source_addresses"`
	DestinationAddresses   fwtypes.ListOfString `tfsdk:"destination_addresses"`
	SourcePorts            fwtypes.ListOfString `tfsdk:"source_ports"`
	DestinationPorts       fwtypes.ListOfString `tfsdk:"destination_ports"`
	SourcePrefixLists      fwtypes.ListOfString `tfsdk:"source_prefix_lists"`
	DestinationPrefixLists fwtypes.ListOfString `tfsdk:"destination_prefix_lists"`
	Protocols              fwtypes.ListOfString `tfsdk:"protocols"`
}

type resourceStatementModel struct {
	Resources     fwtypes.ListOfString `tfsdk:"resources"`
	ResourceTypes fwtypes.ListOfString `tfsdk:"resource_types"`
}

type throughResourcesStatementModel struct {
	ResourceStatement fwtypes.ListNestedObjectValueOf[resourceStatementModel] `tfsdk:"resource_statement"`
}
