// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			"connector_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[connectorConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("azure"),
						),
					},
					Blocks: map[string]schema.Block{
						"azure": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[azureConnectorConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"client_identifier": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"tenant_identifier": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 128),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
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

	var input configservice.PutConnectorInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.PutConnector(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating ConfigService Connector", err.Error())
		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.Arn)
	connector, err := findConnectorByARN(ctx, conn, arn)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ConfigService Connector (%s)", arn), err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, connector.Arn)
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

	response.Diagnostics.Append(r.flatten(ctx, connector, &data)...)
	if response.Diagnostics.HasError() {
		return
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
	// A connector cannot be deleted while service-linked configuration recorders
	// (created by a linked service such as Security Hub) still reference it. Those
	// recorders are removed asynchronously, so retry until the connector is deletable.
	const (
		timeout = 20 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteConnector(ctx, &input)
	}, "still in use")
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ConfigService Connector (%s)", arn), err.Error())
		return
	}
}

func (r *connectorResource) flatten(ctx context.Context, connector *awstypes.Connector, data *connectorResourceModel) diag.Diagnostics {
	return fwflex.Flatten(ctx, connector, data)
}

func findConnectorByARN(ctx context.Context, conn *configservice.Client, arn string) (*awstypes.Connector, error) {
	input := configservice.GetConnectorInput{
		Arn: aws.String(arn),
	}
	return findConnector(ctx, conn, &input)
}

func findConnector(ctx context.Context, conn *configservice.Client, input *configservice.GetConnectorInput) (*awstypes.Connector, error) {
	output, err := conn.GetConnector(ctx, input)

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
	ARN                    types.String                                                 `tfsdk:"arn"`
	ConnectorConfiguration fwtypes.ListNestedObjectValueOf[connectorConfigurationModel] `tfsdk:"connector_configuration"`
	Name                   types.String                                                 `tfsdk:"name"`
	Tags                   tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                tftags.Map                                                   `tfsdk:"tags_all"`
}

type connectorConfigurationModel struct {
	Azure fwtypes.ListNestedObjectValueOf[azureConnectorConfigurationModel] `tfsdk:"azure"`
}

type azureConnectorConfigurationModel struct {
	ClientIdentifier types.String `tfsdk:"client_identifier"`
	TenantIdentifier types.String `tfsdk:"tenant_identifier"`
}
