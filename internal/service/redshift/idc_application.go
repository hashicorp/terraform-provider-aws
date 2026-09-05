// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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

func (r *idcApplicationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		lakehouseServiceIntegrationsValidator{},
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

	si, d := expandServiceIntegrations(ctx, plan.ServiceIntegrations)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ServiceIntegrations = si

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

	flatSI, d := flattenServiceIntegrations(ctx, out.RedshiftIdcApplication.ServiceIntegrations)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ServiceIntegrations = flatSI

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

	flatSI, d := flattenServiceIntegrations(ctx, out.ServiceIntegrations)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ServiceIntegrations = flatSI

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

	// service_integration is handled outside autoflex (autoflex:"-"), so detect changes manually.
	siChanged := !plan.ServiceIntegrations.Equal(state.ServiceIntegrations)

	if diff.HasChanges() || siChanged {
		var input redshift.ModifyRedshiftIdcApplicationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		si, d := expandServiceIntegrations(ctx, plan.ServiceIntegrations)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ServiceIntegrations = si

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

		flatSI, d := flattenServiceIntegrations(ctx, out.RedshiftIdcApplication.ServiceIntegrations)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ServiceIntegrations = flatSI
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
	// Handled manually so a single block can fan out into multiple ServiceIntegrationsUnion entries.
	ServiceIntegrations fwtypes.ListNestedObjectValueOf[serviceIntegrationsModel] `tfsdk:"service_integration" autoflex:"-"`
	Tags                tftags.Map                                                `tfsdk:"tags"`
	TagsAll             tftags.Map                                                `tfsdk:"tags_all"`
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

// expandServiceIntegrations turns the single HCL service_integration block into
// the array of ServiceIntegrationsUnion entries the Redshift API expects.
// This is required for application_type = "Lakehouse", which requires both
// LakeFormation:LakeFormationQuery and Redshift:Connect to be present.
func expandServiceIntegrations(ctx context.Context, l fwtypes.ListNestedObjectValueOf[serviceIntegrationsModel]) ([]awstypes.ServiceIntegrationsUnion, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}

	m, d := l.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() || m == nil {
		return nil, diags
	}

	var out []awstypes.ServiceIntegrationsUnion

	if !m.LakeFormation.IsNull() && !m.LakeFormation.IsUnknown() {
		lf, d := m.LakeFormation.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if lf != nil {
			q, d := lf.LakeFormationQuery.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			if q != nil {
				var scope awstypes.LakeFormationScopeUnionMemberLakeFormationQuery
				diags.Append(fwflex.Expand(ctx, q, &scope.Value)...)
				if diags.HasError() {
					return nil, diags
				}
				out = append(out, &awstypes.ServiceIntegrationsUnionMemberLakeFormation{
					Value: []awstypes.LakeFormationScopeUnion{&scope},
				})
			}
		}
	}

	if !m.Redshift.IsNull() && !m.Redshift.IsUnknown() {
		rs, d := m.Redshift.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if rs != nil {
			c, d := rs.Connect.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			if c != nil {
				var scope awstypes.RedshiftScopeUnionMemberConnect
				diags.Append(fwflex.Expand(ctx, c, &scope.Value)...)
				if diags.HasError() {
					return nil, diags
				}
				out = append(out, &awstypes.ServiceIntegrationsUnionMemberRedshift{
					Value: []awstypes.RedshiftScopeUnion{&scope},
				})
			}
		}
	}

	if !m.S3AccessGrants.IsNull() && !m.S3AccessGrants.IsUnknown() {
		sg, d := m.S3AccessGrants.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if sg != nil {
			rwa, d := sg.ReadWriteAccess.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			if rwa != nil {
				var scope awstypes.S3AccessGrantsScopeUnionMemberReadWriteAccess
				diags.Append(fwflex.Expand(ctx, rwa, &scope.Value)...)
				if diags.HasError() {
					return nil, diags
				}
				out = append(out, &awstypes.ServiceIntegrationsUnionMemberS3AccessGrants{
					Value: []awstypes.S3AccessGrantsScopeUnion{&scope},
				})
			}
		}
	}

	return out, diags
}

// flattenServiceIntegrations folds an API []ServiceIntegrationsUnion back into
// a single service_integration block with multiple sub-blocks set.
func flattenServiceIntegrations(ctx context.Context, unions []awstypes.ServiceIntegrationsUnion) (fwtypes.ListNestedObjectValueOf[serviceIntegrationsModel], diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(unions) == 0 {
		return fwtypes.NewListNestedObjectValueOfNull[serviceIntegrationsModel](ctx), diags
	}

	model := serviceIntegrationsModel{
		LakeFormation:  fwtypes.NewListNestedObjectValueOfNull[lakeFormationModel](ctx),
		Redshift:       fwtypes.NewListNestedObjectValueOfNull[redshiftModel](ctx),
		S3AccessGrants: fwtypes.NewListNestedObjectValueOfNull[s3AccessGrantsModel](ctx),
	}

	for _, u := range unions {
		switch t := u.(type) {
		case *awstypes.ServiceIntegrationsUnionMemberLakeFormation:
			data := lakeFormationModel{
				LakeFormationQuery: fwtypes.NewListNestedObjectValueOfNull[lakeFormationQueryModel](ctx),
			}
			if len(t.Value) > 0 {
				if s, ok := t.Value[0].(*awstypes.LakeFormationScopeUnionMemberLakeFormationQuery); ok {
					var q lakeFormationQueryModel
					diags.Append(fwflex.Flatten(ctx, s.Value, &q)...)
					if diags.HasError() {
						return fwtypes.NewListNestedObjectValueOfNull[serviceIntegrationsModel](ctx), diags
					}
					data.LakeFormationQuery = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &q)
				}
			}
			model.LakeFormation = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

		case *awstypes.ServiceIntegrationsUnionMemberRedshift:
			data := redshiftModel{
				Connect: fwtypes.NewListNestedObjectValueOfNull[connectModel](ctx),
			}
			if len(t.Value) > 0 {
				if s, ok := t.Value[0].(*awstypes.RedshiftScopeUnionMemberConnect); ok {
					var c connectModel
					diags.Append(fwflex.Flatten(ctx, s.Value, &c)...)
					if diags.HasError() {
						return fwtypes.NewListNestedObjectValueOfNull[serviceIntegrationsModel](ctx), diags
					}
					data.Connect = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &c)
				}
			}
			model.Redshift = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

		case *awstypes.ServiceIntegrationsUnionMemberS3AccessGrants:
			data := s3AccessGrantsModel{
				ReadWriteAccess: fwtypes.NewListNestedObjectValueOfNull[readWriteAccessModel](ctx),
			}
			if len(t.Value) > 0 {
				if s, ok := t.Value[0].(*awstypes.S3AccessGrantsScopeUnionMemberReadWriteAccess); ok {
					var r readWriteAccessModel
					diags.Append(fwflex.Flatten(ctx, s.Value, &r)...)
					if diags.HasError() {
						return fwtypes.NewListNestedObjectValueOfNull[serviceIntegrationsModel](ctx), diags
					}
					data.ReadWriteAccess = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &r)
				}
			}
			model.S3AccessGrants = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
		}
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model), diags
}

// lakehouseServiceIntegrationsValidator surfaces the AWS Lakehouse precondition
// (must enable both LakeFormation:LakeFormationQuery and Redshift:Connect) at
// plan time instead of waiting for the API to return InvalidParameterValue.
type lakehouseServiceIntegrationsValidator struct{}

func (lakehouseServiceIntegrationsValidator) Description(_ context.Context) string {
	return `When application_type = "Lakehouse", service_integration must enable both lake_formation.lake_formation_query and redshift.connect.`
}

func (v lakehouseServiceIntegrationsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (lakehouseServiceIntegrationsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg idcApplicationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.ApplicationType.IsNull() || cfg.ApplicationType.IsUnknown() {
		return
	}
	if cfg.ApplicationType.ValueString() != string(awstypes.ApplicationTypeLakehouse) {
		return
	}
	if cfg.ServiceIntegrations.IsNull() || cfg.ServiceIntegrations.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("service_integration"),
			"Invalid Lakehouse service_integration",
			`application_type = "Lakehouse" requires both lake_formation.lake_formation_query.authorization = "Enabled" and redshift.connect.authorization = "Enabled".`,
		)
		return
	}

	si, diags := cfg.ServiceIntegrations.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || si == nil {
		return
	}

	if si.LakeFormation.IsNull() || si.LakeFormation.IsUnknown() ||
		si.Redshift.IsNull() || si.Redshift.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("service_integration"),
			"Invalid Lakehouse service_integration",
			`application_type = "Lakehouse" requires both lake_formation.lake_formation_query.authorization = "Enabled" and redshift.connect.authorization = "Enabled".`,
		)
	}
}
