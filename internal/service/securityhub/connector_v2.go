// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_securityhub_connector_v2", name="Connector V2")
// @IdentityAttribute("connector_id")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub;securityhub;securityhub.GetConnectorV2Output")
// @Testing(serialize=true)
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="connector_id")
func newConnectorV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &connectorV2Resource{}, nil
}

type connectorV2Resource struct {
	framework.ResourceWithModel[connectorV2ResourceModel]
	framework.WithImportByIdentity
}

func (r *connectorV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:  framework.ARNAttributeComputedOnly(),
			"connector_id": framework.IDAttribute(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"health": framework.ResourceComputedListOfObjectsAttribute[healthCheckModel](ctx),
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"connector_provider": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[providerDetailModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"jira_cloud": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[jiraCloudDetailModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("jira_cloud"),
									path.MatchRelative().AtParent().AtName("service_now"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_status": schema.StringAttribute{
										Computed: true,
									},
									"auth_url": schema.StringAttribute{
										Computed: true,
									},
									"cloud_id": schema.StringAttribute{
										Computed: true,
									},
									names.AttrDomain: schema.StringAttribute{
										Computed: true,
									},
									"project_key": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"service_now": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serviceNowDetailModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_status": schema.StringAttribute{
										Computed: true,
									},
									"instance_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
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

func (r *connectorV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data connectorV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input securityhub.CreateConnectorV2Input
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	outputCC, err := conn.CreateConnectorV2(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Security Hub V2 Connector (%s)", name), err.Error())
		return
	}

	// Set values for unknowns.
	connectorID := aws.ToString(outputCC.ConnectorId)
	outputGC, err := findConnectorV2ByID(ctx, conn, connectorID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Connector (%s)", connectorID), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, outputGC, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *connectorV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data connectorV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	connectorID := fwflex.StringValueFromFramework(ctx, data.ConnectorID)
	output, err := findConnectorV2ByID(ctx, conn, connectorID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Connector (%s)", connectorID), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *connectorV2Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new connectorV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	diff, diags := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		connectorID := fwflex.StringValueFromFramework(ctx, new.ConnectorID)
		var input securityhub.UpdateConnectorV2Input
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateConnectorV2(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Hub V2 Connector (%s)", connectorID), err.Error())
			return
		}

		// Set values for unknowns.
		output, err := findConnectorV2ByID(ctx, conn, connectorID)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Connector (%s)", connectorID), err.Error())
			return
		}

		response.Diagnostics.Append(r.flatten(ctx, output, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *connectorV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data connectorV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	connectorID := fwflex.StringValueFromFramework(ctx, data.ConnectorID)
	input := securityhub.DeleteConnectorV2Input{
		ConnectorId: aws.String(connectorID),
	}
	_, err := conn.DeleteConnectorV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Hub V2 Connector (%s)", connectorID), err.Error())
		return
	}
}

func (r *connectorV2Resource) flatten(ctx context.Context, connectorV2 *securityhub.GetConnectorV2Output, data *connectorV2ResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, connectorV2, data, fwflex.WithFieldNameSuffix("Detail"))...)
	return diags
}

func findConnectorV2ByID(ctx context.Context, conn *securityhub.Client, connectorID string) (*securityhub.GetConnectorV2Output, error) {
	input := securityhub.GetConnectorV2Input{
		ConnectorId: aws.String(connectorID),
	}

	return findConnectorV2(ctx, conn, &input)
}

func findConnectorV2(ctx context.Context, conn *securityhub.Client, input *securityhub.GetConnectorV2Input) (*securityhub.GetConnectorV2Output, error) {
	output, err := conn.GetConnectorV2(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Security Hub V2 is not enabled") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type connectorV2ResourceModel struct {
	framework.WithRegionModel
	ConnectorARN types.String                                         `tfsdk:"arn"`
	ConnectorID  types.String                                         `tfsdk:"connector_id"`
	Description  types.String                                         `tfsdk:"description"`
	Health       fwtypes.ListNestedObjectValueOf[healthCheckModel]    `tfsdk:"health"`
	KmsKeyARN    fwtypes.ARN                                          `tfsdk:"kms_key_arn"`
	Name         types.String                                         `tfsdk:"name"`
	Provider     fwtypes.ListNestedObjectValueOf[providerDetailModel] `tfsdk:"connector_provider"`
	Tags         tftags.Map                                           `tfsdk:"tags"`
	TagsAll      tftags.Map                                           `tfsdk:"tags_all"`
}

type healthCheckModel struct {
	ConnectorStatus types.String      `tfsdk:"connector_status"`
	LastCheckedAt   timetypes.RFC3339 `tfsdk:"last_checked_at"`
	Message         types.String      `tfsdk:"message"`
}

type providerDetailModel struct {
	JiraCloud  fwtypes.ListNestedObjectValueOf[jiraCloudDetailModel]  `tfsdk:"jira_cloud"`
	ServiceNow fwtypes.ListNestedObjectValueOf[serviceNowDetailModel] `tfsdk:"service_now"`
}

var (
	_ fwflex.TypedExpander = providerDetailModel{}
	_ fwflex.Flattener     = &providerDetailModel{}
)

func (m providerDetailModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeFor[awstypes.ProviderConfiguration]():
		return m.expandToProviderConfiguration(ctx)
	case reflect.TypeFor[awstypes.ProviderUpdateConfiguration]():
		return m.expandToProviderUpdateConfiguration(ctx)
	}
	return nil, diags
}

func (m providerDetailModel) expandToProviderConfiguration(ctx context.Context) (awstypes.ProviderConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.JiraCloud.IsNull():
		data, d := m.JiraCloud.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ProviderConfigurationMemberJiraCloud
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.ServiceNow.IsNull():
		data, d := m.ServiceNow.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ProviderConfigurationMemberServiceNow
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m providerDetailModel) expandToProviderUpdateConfiguration(ctx context.Context) (awstypes.ProviderUpdateConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.JiraCloud.IsNull():
		data, d := m.JiraCloud.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ProviderUpdateConfigurationMemberJiraCloud
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.ServiceNow.IsNull():
		data, d := m.ServiceNow.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ProviderUpdateConfigurationMemberServiceNow
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *providerDetailModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ProviderDetailMemberJiraCloud:
		var data jiraCloudDetailModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.JiraCloud = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.ProviderDetailMemberServiceNow:
		var data serviceNowDetailModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.ServiceNow = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("provider flatten: %T", v),
		)
	}
	return diags
}

type jiraCloudDetailModel struct {
	AuthStatus types.String `tfsdk:"auth_status"`
	AuthURL    types.String `tfsdk:"auth_url"`
	CloudID    types.String `tfsdk:"cloud_id"`
	Domain     types.String `tfsdk:"domain"`
	ProjectKey types.String `tfsdk:"project_key"`
}

type serviceNowDetailModel struct {
	AuthStatus   types.String `tfsdk:"auth_status"`
	InstanceName types.String `tfsdk:"instance_name"`
	SecretARN    fwtypes.ARN  `tfsdk:"secret_arn"`
}
