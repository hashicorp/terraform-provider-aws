// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssoadmin

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

// @FrameworkResource("aws_ssoadmin_application", name="Application")
// @Tags
// @ArnIdentity(identityDuplicateAttributes="id;application_arn")
// @ArnFormat(global=true)
// @IdentityFix
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssoadmin;ssoadmin.DescribeApplicationOutput")
// @Testing(preCheckWithRegion="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstancesWithRegion")
// @Testing(v60NullValuesError=true)
func newApplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationResource{}, nil
}

type applicationResource struct {
	framework.ResourceWithModel[applicationResourceModel]
	framework.WithImportByIdentity
}

func (r *applicationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_account": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_arn": framework.ARNAttributeComputedOnlyDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"application_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ApplicationStatus](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"portal_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[portalOptionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"visibility": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ApplicationVisibility](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								// If explicitly set, require that sign_in_options also be configured
								// to ensure the flattener correctly reads both values into state.
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("sign_in_options")),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"sign_in_options": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"application_url": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
									"origin": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SignInOrigin](),
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

func (r *applicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input ssoadmin.CreateApplicationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateApplication(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SSO Application (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.ApplicationArn)
	data.ApplicationARN = data.ARN
	data.ID = data.ARN

	// Read after create to get computed attributes omitted from the create response.
	app, err := findApplicationByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Application (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Skip writing to state if only the visibilty attribute is returned
	// to avoid a nested computed attribute causing a diff.
	if app.PortalOptions != nil && app.PortalOptions.SignInOptions == nil {
		app.PortalOptions = nil
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, app, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *applicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	output, err := findApplicationByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Application (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Skip writing to state if only the visibilty attribute is returned
	// to avoid a nested computed attribute causing a diff.
	if output.PortalOptions != nil && output.PortalOptions.SignInOptions == nil {
		output.PortalOptions = nil
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = data.ApplicationARN

	// listTags requires both application and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	tags, err := listTags(ctx, conn, data.ARN.ValueString(), data.InstanceARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Application (%s) tags", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, svcTags(tags))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *applicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old applicationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.Name.Equal(old.Name) ||
		!new.PortalOptions.Equal(old.PortalOptions) ||
		!new.Status.Equal(old.Status) {
		var input ssoadmin.UpdateApplicationInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApplication(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SSO Application (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	// updateTags requires both application and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.ARN.ValueString(), new.InstanceARN.ValueString(), oldTagsAll, newTagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SSO Application (%s) tags", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *applicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data applicationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	input := ssoadmin.DeleteApplicationInput{
		ApplicationArn: fwflex.StringFromFramework(ctx, data.ARN),
	}
	_, err := conn.DeleteApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SSO Application (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findApplicationByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.DescribeApplicationOutput, error) {
	input := ssoadmin.DescribeApplicationInput{
		ApplicationArn: aws.String(id),
	}
	output, err := conn.DescribeApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

type applicationResourceModel struct {
	framework.WithRegionModel
	ApplicationAccount     types.String                                        `tfsdk:"application_account"`
	ApplicationARN         types.String                                        `tfsdk:"application_arn"`
	ApplicationProviderARN fwtypes.ARN                                         `tfsdk:"application_provider_arn"`
	ARN                    types.String                                        `tfsdk:"arn"`
	ClientToken            types.String                                        `tfsdk:"client_token"`
	Description            types.String                                        `tfsdk:"description"`
	ID                     types.String                                        `tfsdk:"id"`
	InstanceARN            fwtypes.ARN                                         `tfsdk:"instance_arn"`
	Name                   types.String                                        `tfsdk:"name"`
	PortalOptions          fwtypes.ListNestedObjectValueOf[portalOptionsModel] `tfsdk:"portal_options"`
	Status                 fwtypes.StringEnum[awstypes.ApplicationStatus]      `tfsdk:"status"`
	Tags                   tftags.Map                                          `tfsdk:"tags"`
	TagsAll                tftags.Map                                          `tfsdk:"tags_all"`
}

type portalOptionsModel struct {
	SignInOptions fwtypes.ListNestedObjectValueOf[signInOptionsModel] `tfsdk:"sign_in_options"`
	Visibility    fwtypes.StringEnum[awstypes.ApplicationVisibility]  `tfsdk:"visibility"`
}

type signInOptionsModel struct {
	ApplicationURL types.String                              `tfsdk:"application_url"`
	Origin         fwtypes.StringEnum[awstypes.SignInOrigin] `tfsdk:"origin"`
}
