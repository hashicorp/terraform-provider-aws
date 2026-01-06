// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_idc_application", name="IDC Application")
// @Tags(identifierAttribute="redshift_idc_application_arn")
// @Testing(tagsTest=false)
func newIDCApplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &idcApplicationResource{}

	return r, nil
}

type idcApplicationResource struct {
	framework.ResourceWithModel[idcApplicationResourceModel]
}

func (r *idcApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ApplicationType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrIAMRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"idc_display_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 127),
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				},
			},
			"idc_instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"idc_managed_application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identity_namespace": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 127),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_+.#@$-]+$`), "must match ^[a-zA-Z0-9_+.#@$-]+$"),
				},
			},
			"redshift_idc_application_arn": framework.ARNAttributeComputedOnly(),
			"redshift_idc_application_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-z][a-z0-9]*(-[a-z0-9]+)*`), "must match [a-z][a-z0-9]*(-[a-z0-9]+)*"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"authorized_token_issuer": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authorizedTokenIssuerModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"authorized_audiences_list": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"trusted_token_issuer_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"service_integration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[serviceIntegrationsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"lake_formation": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[lakeFormationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lake_formation_query": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lakeFormationQueryModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorization": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ServiceAuthorization](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
						"redshift": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[redshiftModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"connect": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[connectModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorization": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ServiceAuthorization](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
						"s3_access_grants": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3AccessGrantsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"read_write_access": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[readWriteAccessModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorization": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ServiceAuthorization](),
													Required:   true,
												},
											},
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

func (r *idcApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan idcApplicationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input redshift.CreateRedshiftIdcApplicationInput
	input.Tags = getTagsIn(ctx)

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateRedshiftIdcApplication(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RedshiftIDCApplicationName.String())
		return
	}
	if out == nil || out.RedshiftIdcApplication == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.RedshiftIDCApplicationName.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.RedshiftIdcApplication, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *idcApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state idcApplicationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRedshiftIDCApplicationByARN(ctx, conn, state.RedshiftIDCApplicationARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RedshiftIDCApplicationName.String())
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *idcApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var plan, state idcApplicationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input redshift.ModifyRedshiftIdcApplicationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyRedshiftIdcApplication(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RedshiftIDCApplicationName.String())
			return
		}
		if out == nil || out.RedshiftIdcApplication == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.RedshiftIDCApplicationARN.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.RedshiftIdcApplication, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *idcApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftClient(ctx)

	var state idcApplicationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := redshift.DeleteRedshiftIdcApplicationInput{
		RedshiftIdcApplicationArn: state.RedshiftIDCApplicationARN.ValueStringPointer(),
	}

	_, err := conn.DeleteRedshiftIdcApplication(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.RedshiftIdcApplicationNotExistsFault](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RedshiftIDCApplicationName.String())
		return
	}
}

func (r *idcApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("redshift_idc_application_arn"), req, resp)
}

type idcApplicationResourceModel struct {
	framework.WithRegionModel
	ApplicationType            fwtypes.StringEnum[awstypes.ApplicationType]                `tfsdk:"application_type"`
	AuthorizedTokenIssuerList  fwtypes.ListNestedObjectValueOf[authorizedTokenIssuerModel] `tfsdk:"authorized_token_issuer"`
	IAMRoleARN                 fwtypes.ARN                                                 `tfsdk:"iam_role_arn"`
	IDCDisplayName             types.String                                                `tfsdk:"idc_display_name"`
	IDCInstanceARN             fwtypes.ARN                                                 `tfsdk:"idc_instance_arn"`
	IDCManagedApplicationARN   fwtypes.ARN                                                 `tfsdk:"idc_managed_application_arn"`
	IdentityNamespace          types.String                                                `tfsdk:"identity_namespace"`
	RedshiftIDCApplicationARN  types.String                                                `tfsdk:"redshift_idc_application_arn"`
	RedshiftIDCApplicationName types.String                                                `tfsdk:"redshift_idc_application_name"`
	ServiceIntegrations        fwtypes.ListNestedObjectValueOf[serviceIntegrationsModel]   `tfsdk:"service_integration"`
	Tags                       tftags.Map                                                  `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                  `tfsdk:"tags_all"`
}

type authorizedTokenIssuerModel struct {
	AuthorizedAudiencesList fwtypes.ListOfString `tfsdk:"authorized_audiences_list"`
	TrustedTokenIssuerARN   fwtypes.ARN          `tfsdk:"trusted_token_issuer_arn"`
}

type serviceIntegrationsModel struct {
	LakeFormation  fwtypes.ListNestedObjectValueOf[lakeFormationModel]  `tfsdk:"lake_formation"`
	Redshift       fwtypes.ListNestedObjectValueOf[redshiftModel]       `tfsdk:"redshift"`
	S3AccessGrants fwtypes.ListNestedObjectValueOf[s3AccessGrantsModel] `tfsdk:"s3_access_grants"`
}

var (
	_ fwflex.Expander  = serviceIntegrationsModel{}
	_ fwflex.Flattener = &serviceIntegrationsModel{}
)

func (m serviceIntegrationsModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.LakeFormation.IsNull():
		lakeFormationData, d := m.LakeFormation.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		lfQuery, d := lakeFormationData.LakeFormationQuery.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ServiceIntegrationsUnionMemberLakeFormation
		if lfQuery != nil {
			var query awstypes.LakeFormationScopeUnionMemberLakeFormationQuery
			diags.Append(fwflex.Expand(ctx, lfQuery, &query.Value)...)
			if diags.HasError() {
				return nil, diags
			}
			r.Value = []awstypes.LakeFormationScopeUnion{&query}
		}

		return &r, diags

	case !m.Redshift.IsNull():
		redshiftData, d := m.Redshift.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		connect, d := redshiftData.Connect.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ServiceIntegrationsUnionMemberRedshift
		if connect != nil {
			var connectScope awstypes.RedshiftScopeUnionMemberConnect
			diags.Append(fwflex.Expand(ctx, connect, &connectScope.Value)...)
			if diags.HasError() {
				return nil, diags
			}
			r.Value = []awstypes.RedshiftScopeUnion{&connectScope}
		}

		return &r, diags

	case !m.S3AccessGrants.IsNull():
		s3AccessGrants, d := m.S3AccessGrants.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		readWriteAccess, d := s3AccessGrants.ReadWriteAccess.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.ServiceIntegrationsUnionMemberS3AccessGrants
		if readWriteAccess != nil {
			var rwaGrantsScope awstypes.S3AccessGrantsScopeUnionMemberReadWriteAccess
			diags.Append(fwflex.Expand(ctx, readWriteAccess, &rwaGrantsScope.Value)...)
			if diags.HasError() {
				return nil, diags
			}
			r.Value = []awstypes.S3AccessGrantsScopeUnion{&rwaGrantsScope}
		}

		return &r, diags
	}

	return nil, diags
}

func (m *serviceIntegrationsModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ServiceIntegrationsUnionMemberLakeFormation:
		var data lakeFormationModel
		if len(t.Value) > 0 {
			var lfQueryData lakeFormationQueryModel

			// Type switch on the LakeFormationScopeUnion to get LakeFormationQuery
			switch scopeUnion := t.Value[0].(type) {
			case *awstypes.LakeFormationScopeUnionMemberLakeFormationQuery:
				// Flatten the LakeFormationQuery value into the model
				diags.Append(fwflex.Flatten(ctx, scopeUnion.Value, &lfQueryData)...)
				if diags.HasError() {
					return diags
				}

				// Set the LakeFormationQuery in the parent model
				data.LakeFormationQuery = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &lfQueryData)
			}
		}
		m.LakeFormation = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case awstypes.ServiceIntegrationsUnionMemberRedshift:
		var data redshiftModel

		// Handle the nested RedshiftScope union
		if len(t.Value) > 0 {
			var connectData connectModel

			// Type switch on the RedshiftScopeUnion to get Connect
			switch scopeUnion := t.Value[0].(type) {
			case *awstypes.RedshiftScopeUnionMemberConnect:
				// Flatten the Connect value into the model
				diags.Append(fwflex.Flatten(ctx, scopeUnion.Value, &connectData)...)
				if diags.HasError() {
					return diags
				}

				// Set the Connect in the parent model
				data.Connect = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &connectData)
			}
		}
		m.Redshift = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	case awstypes.ServiceIntegrationsUnionMemberS3AccessGrants:
		var data s3AccessGrantsModel

		// Handle the nested S3AccessGrantsScope union
		if len(t.Value) > 0 {
			var readWriteAccessData readWriteAccessModel

			// Type switch on the S3AccessGrantsScopeUnion to get ReadWriteAccess
			switch scopeUnion := t.Value[0].(type) {
			case *awstypes.S3AccessGrantsScopeUnionMemberReadWriteAccess:
				// Flatten the ReadWriteAccess value into the model
				diags.Append(fwflex.Flatten(ctx, scopeUnion.Value, &readWriteAccessData)...)
				if diags.HasError() {
					return diags
				}

				// Set the ReadWriteAccess in the parent model
				data.ReadWriteAccess = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &readWriteAccessData)
			}
		}
		m.S3AccessGrants = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

type lakeFormationModel struct {
	LakeFormationQuery fwtypes.ListNestedObjectValueOf[lakeFormationQueryModel] `tfsdk:"lake_formation_query"`
}

type lakeFormationQueryModel struct {
	Authorization fwtypes.StringEnum[awstypes.ServiceAuthorization] `tfsdk:"authorization"`
}

type redshiftModel struct {
	Connect fwtypes.ListNestedObjectValueOf[connectModel] `tfsdk:"connect"`
}

type connectModel struct {
	Authorization fwtypes.StringEnum[awstypes.ServiceAuthorization] `tfsdk:"authorization"`
}

type s3AccessGrantsModel struct {
	ReadWriteAccess fwtypes.ListNestedObjectValueOf[readWriteAccessModel] `tfsdk:"read_write_access"`
}

type readWriteAccessModel struct {
	Authorization fwtypes.StringEnum[awstypes.ServiceAuthorization] `tfsdk:"authorization"`
}
