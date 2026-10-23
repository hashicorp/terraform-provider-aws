// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_environment_blueprint", name="Environment Blueprint")
// @IdentityAttribute("domain_identifier")
// @IdentityAttribute("id")
// @ImportIDHandler("environmentBlueprintImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/datazone;datazone.GetEnvironmentBlueprintOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(importStateIdAttributes="domain_identifier;id", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newEnvironmentBlueprintResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentBlueprintResource{}

	r.SetDefaultCreateTimeout(2 * time.Minute)

	return r, nil
}

type environmentBlueprintResource struct {
	framework.ResourceWithModel[environmentBlueprintResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *environmentBlueprintResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"blueprint_category": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BlueprintCategory](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"glossary_terms": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrProviderName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"provisioning_properties": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[provisioningPropertiesModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloud_formation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudFormationPropertiesModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"template_url": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"user_parameters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customParameterModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDefaultValue: schema.StringAttribute{
							Optional: true,
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"field_type": schema.StringAttribute{
							Required: true,
						},
						"is_editable": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"is_optional": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"is_update_supported": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
						"key_name": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *environmentBlueprintResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentBlueprintResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	var input datazone.CreateEnvironmentBlueprintInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// The domain execution role's access to the CloudFormation template in S3
	// is subject to IAM eventual consistency, so retry AccessDenied until the
	// configurable create timeout elapses (mirrors the datazone domain resource).
	output, err := tfresource.RetryWhenAWSErrCodeContains(ctx, r.CreateTimeout(ctx, data.Timeouts), func(ctx context.Context) (*datazone.CreateEnvironmentBlueprintOutput, error) {
		return conn.CreateEnvironmentBlueprint(ctx, &input)
	}, errCodeAccessDenied)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.ProviderName = fwflex.StringToFramework(ctx, output.Provider)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *environmentBlueprintResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data environmentBlueprintResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	output, err := findEnvironmentBlueprintByTwoPartKey(ctx, conn, data.DomainIdentifier.ValueString(), data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	// Preserve the user-configured domain identifier (not returned by the API).
	domainID := data.DomainIdentifier
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.DomainIdentifier = domainID
	data.ProviderName = fwflex.StringToFramework(ctx, output.Provider)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *environmentBlueprintResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state environmentBlueprintResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if !diff.HasChanges() {
		// Only non-resource fields (e.g. timeouts) changed; avoid an
		// unnecessary API call that would require update permission and
		// mutate updated_at.
		smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
		return
	}

	var input datazone.UpdateEnvironmentBlueprintInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}
	// Update addresses the blueprint by its ID, not its name.
	input.Identifier = state.ID.ValueStringPointer()

	// AutoFlex omits a null string / null list from the request, so the API
	// keeps the previous value. When the user removes a configured
	// description or user_parameters block, send the cleared value explicitly.
	if plan.Description.IsNull() {
		input.Description = aws.String("")
	}
	if plan.UserParameters.IsNull() && !state.UserParameters.IsNull() {
		input.UserParameters = []awstypes.CustomParameter{}
	}

	output, err := conn.UpdateEnvironmentBlueprint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	// UpdateEnvironmentBlueprint is read-after-write eventually consistent: an
	// immediate GetEnvironmentBlueprint can still return the pre-update values.
	// Wait until a fresh read reflects this update (matched on UpdatedAt) so the
	// next refresh/plan does not observe stale data.
	if output.UpdatedAt != nil {
		const consistencyTimeout = 2 * time.Minute
		if _, err := tfresource.RetryUntilEqual(ctx, consistencyTimeout, true, func(ctx context.Context) (bool, error) {
			fresh, err := findEnvironmentBlueprintByTwoPartKey(ctx, conn, plan.DomainIdentifier.ValueString(), aws.ToString(output.Id))
			if err != nil {
				return false, err
			}
			return fresh.UpdatedAt != nil && !fresh.UpdatedAt.Before(*output.UpdatedAt), nil
		}); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.ValueString())
			return
		}
	}

	domainID := plan.DomainIdentifier
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if response.Diagnostics.HasError() {
		return
	}
	plan.DomainIdentifier = domainID
	plan.ProviderName = fwflex.StringToFramework(ctx, output.Provider)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *environmentBlueprintResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentBlueprintResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	input := datazone.DeleteEnvironmentBlueprintInput{
		DomainIdentifier: data.DomainIdentifier.ValueStringPointer(),
		Identifier:       data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteEnvironmentBlueprint(ctx, &input)

	if isResourceMissing(err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}
}

func findEnvironmentBlueprintByTwoPartKey(ctx context.Context, conn *datazone.Client, domainID, id string) (*datazone.GetEnvironmentBlueprintOutput, error) {
	input := datazone.GetEnvironmentBlueprintInput{
		DomainIdentifier: aws.String(domainID),
		Identifier:       aws.String(id),
	}
	output, err := conn.GetEnvironmentBlueprint(ctx, &input)

	if isResourceMissing(err) {
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

var (
	_ inttypes.ImportIDParser = environmentBlueprintImportID{}
)

type environmentBlueprintImportID struct{}

func (environmentBlueprintImportID) Parse(id string) (string, map[string]any, error) {
	domainID, blueprintID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <domain-identifier>%s<id>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"domain_identifier": domainID,
		names.AttrID:        blueprintID,
	}

	return id, result, nil
}

type environmentBlueprintResourceModel struct {
	framework.WithRegionModel
	BlueprintCategory      fwtypes.StringEnum[awstypes.BlueprintCategory]               `tfsdk:"blueprint_category"`
	CreatedAt              timetypes.RFC3339                                            `tfsdk:"created_at"`
	Description            types.String                                                 `tfsdk:"description"`
	DomainIdentifier       types.String                                                 `tfsdk:"domain_identifier"`
	GlossaryTerms          fwtypes.ListOfString                                         `tfsdk:"glossary_terms"`
	ID                     types.String                                                 `tfsdk:"id"`
	Name                   types.String                                                 `tfsdk:"name"`
	ProviderName           types.String                                                 `tfsdk:"provider_name" autoflex:"-"`
	ProvisioningProperties fwtypes.ListNestedObjectValueOf[provisioningPropertiesModel] `tfsdk:"provisioning_properties"`
	Timeouts               timeouts.Value                                               `tfsdk:"timeouts"`
	UpdatedAt              timetypes.RFC3339                                            `tfsdk:"updated_at"`
	UserParameters         fwtypes.ListNestedObjectValueOf[customParameterModel]        `tfsdk:"user_parameters"`
}

// provisioningPropertiesModel maps to the awstypes.ProvisioningProperties tagged
// union. Autoflex cannot dispatch union interfaces automatically, so the model
// implements TypedExpander (plan -> SDK) and Flattener (SDK -> state) explicitly.
// Today the union has a single member: CloudFormation.
type provisioningPropertiesModel struct {
	CloudFormation fwtypes.ListNestedObjectValueOf[cloudFormationPropertiesModel] `tfsdk:"cloud_formation"`
}

var (
	_ fwflex.TypedExpander = provisioningPropertiesModel{}
	_ fwflex.Flattener     = &provisioningPropertiesModel{}
)

func (m provisioningPropertiesModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch targetType {
	case reflect.TypeFor[awstypes.ProvisioningProperties]():
		if !m.CloudFormation.IsNull() {
			cf, d := m.CloudFormation.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			var r awstypes.ProvisioningPropertiesMemberCloudFormation
			diags.Append(fwflex.Expand(ctx, cf, &r.Value)...)
			if diags.HasError() {
				return nil, diags
			}

			return &r, diags
		}
	}

	diags.AddError(
		"Unexpected ProvisioningProperties target type",
		fmt.Sprintf("expand target %q is not supported; this is a bug in the provider", targetType.String()),
	)
	return nil, diags
}

func (m *provisioningPropertiesModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ProvisioningPropertiesMemberCloudFormation:
		var cf cloudFormationPropertiesModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &cf)...)
		if diags.HasError() {
			return diags
		}

		m.CloudFormation = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cf)
	default:
		diags.AddError(
			"Unexpected ProvisioningProperties union member",
			fmt.Sprintf("received %T; this is a bug in the provider", t),
		)
	}

	return diags
}

type cloudFormationPropertiesModel struct {
	TemplateURL types.String `tfsdk:"template_url"`
}

type customParameterModel struct {
	DefaultValue      types.String `tfsdk:"default_value"`
	Description       types.String `tfsdk:"description"`
	FieldType         types.String `tfsdk:"field_type"`
	IsEditable        types.Bool   `tfsdk:"is_editable"`
	IsOptional        types.Bool   `tfsdk:"is_optional"`
	IsUpdateSupported types.Bool   `tfsdk:"is_update_supported"`
	KeyName           types.String `tfsdk:"key_name"`
}
