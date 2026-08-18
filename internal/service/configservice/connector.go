// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_config_connector", name="Connector")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/configservice/types;awstypes;awstypes.Connector")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(tagsTest=false)
func newConnectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &connectorResource{}, nil
}

type connectorResource struct {
	framework.ResourceWithModel[connectorResourceModel]
	framework.WithImportByIdentity
}

func (r *connectorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"azure": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[connectorAzureConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"client_identifier": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"tenant_identifier": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *connectorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data connectorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	azure, diags := data.Azure.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	input := configservice.PutConnectorInput{
		ConnectorConfiguration: &awstypes.ConnectorConfiguration{
			Azure: &awstypes.AzureConnectorConfiguration{
				ClientIdentifier: fwflex.StringFromFramework(ctx, azure.ClientIdentifier),
				TenantIdentifier: fwflex.StringFromFramework(ctx, azure.TenantIdentifier),
			},
		},
		Tags: getTagsIn(ctx),
	}

	output, err := conn.PutConnector(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating ConfigService Connector", err.Error())

		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.Arn)
	data.ARN = fwflex.StringValueToFramework(ctx, arn)

	connector, err := findConnectorByARN(ctx, conn, arn)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ConfigService Connector (%s)", arn), err.Error())

		return
	}

	data.Name = fwflex.StringToFramework(ctx, connector.Name)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *connectorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data connectorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	connector, err := findConnectorByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ConfigService Connector (%s)", arn), err.Error())

		return
	}

	data.Name = fwflex.StringToFramework(ctx, connector.Name)

	// The connector configuration attributes (client_identifier, tenant_identifier)
	// force replacement and are set at creation. Only refresh them from the API when
	// they are returned, to avoid spurious diffs if the service omits them.
	if cc := connector.ConnectorConfiguration; cc != nil && cc.Azure != nil {
		azure := &connectorAzureConfigurationModel{
			ClientIdentifier: fwflex.StringToFramework(ctx, cc.Azure.ClientIdentifier),
			TenantIdentifier: fwflex.StringToFramework(ctx, cc.Azure.TenantIdentifier),
		}
		data.Azure = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, azure)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *connectorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data connectorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := configservice.DeleteConnectorInput{
		Arn: aws.String(arn),
	}
	_, err := conn.DeleteConnector(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ConfigService Connector (%s)", arn), err.Error())

		return
	}
}

func findConnectorByARN(ctx context.Context, conn *configservice.Client, arn string) (*awstypes.Connector, error) {
	input := configservice.GetConnectorInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetConnector(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connector == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Connector, nil
}

type connectorResourceModel struct {
	framework.WithRegionModel
	ARN     types.String                                                      `tfsdk:"arn"`
	Azure   fwtypes.ListNestedObjectValueOf[connectorAzureConfigurationModel] `tfsdk:"azure"`
	Name    types.String                                                      `tfsdk:"name"`
	Tags    tftags.Map                                                        `tfsdk:"tags"`
	TagsAll tftags.Map                                                        `tfsdk:"tags_all"`
}

type connectorAzureConfigurationModel struct {
	ClientIdentifier types.String `tfsdk:"client_identifier"`
	TenantIdentifier types.String `tfsdk:"tenant_identifier"`
}
