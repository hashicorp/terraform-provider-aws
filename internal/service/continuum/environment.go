// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_continuum_environment", name="Environment")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum;continuum.GetEnvironmentOutput")
func newEnvironmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type environmentResource struct {
	framework.ResourceWithModel[environmentResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *environmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_domains": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Description: "Domains that code and context in this environment may reach over the network.",
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"collector_data_access_role_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the role that Continuum assumes to read data from the associated collector.",
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"patch_creation_policy": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PatchCreationPolicy](),
				Optional:   true,
				Computed:   true,
			},
			"scan_frequency": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ISO 8601 duration between automatic scans, such as `P1D`. `PT0H` turns automatic scans off.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnvironmentStatus](),
				Computed:   true,
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"repository_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[repositorySourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"default_branch": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"integration_id": schema.StringAttribute{
							Required: true,
						},
						"repository": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"s3_source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3SourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"code_prefix": schema.StringAttribute{
							Required: true,
						},
						"context_prefix": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"integration_id": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan environmentResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	envConfig, srcConfig, diags := plan.expand(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	input := continuum.CreateEnvironmentInput{
		Name: aws.String(name),
		Configuration: &awstypes.CreateEnvironmentConfiguration{
			CodeConfiguration: &awstypes.CreateCodeConfiguration{
				EnvironmentConfiguration: envConfig,
				SourceConfiguration:      srcConfig,
			},
		},
	}

	output, err := conn.CreateEnvironment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
	if output == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, name)
		return
	}

	id := aws.ToString(output.EnvironmentId)
	environment, err := waitEnvironmentActive(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, environment))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state environmentResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	output, err := findEnvironmentByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, state.flatten(ctx, output))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *environmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state environmentResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	if !plan.Name.Equal(state.Name) ||
		!plan.AllowedDomains.Equal(state.AllowedDomains) ||
		!plan.PatchCreationPolicy.Equal(state.PatchCreationPolicy) ||
		!plan.ScanFrequency.Equal(state.ScanFrequency) ||
		!plan.RepositorySources.Equal(state.RepositorySources) ||
		!plan.S3Sources.Equal(state.S3Sources) {
		envConfig, srcConfig, diags := plan.expand(ctx)
		smerr.AddEnrich(ctx, &response.Diagnostics, diags)
		if response.Diagnostics.HasError() {
			return
		}

		input := continuum.UpdateEnvironmentInput{
			EnvironmentId: plan.ID.ValueStringPointer(),
			Name:          plan.Name.ValueStringPointer(),
			Configuration: &awstypes.UpdateEnvironmentConfiguration{
				CodeConfiguration: &awstypes.UpdateCodeConfiguration{
					EnvironmentConfiguration: envConfig,
					SourceConfiguration:      srcConfig,
				},
			},
		}

		if _, err := conn.UpdateEnvironment(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	output, err := waitEnvironmentActive(ctx, conn, plan.ID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, plan.flatten(ctx, output))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state environmentResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	input := continuum.DeleteEnvironmentInput{
		EnvironmentId: state.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteEnvironment(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, state.ID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func findEnvironmentByID(ctx context.Context, conn *continuum.Client, id string) (*continuum.GetEnvironmentOutput, error) {
	input := continuum.GetEnvironmentInput{
		EnvironmentId: aws.String(id),
	}

	return findEnvironment(ctx, conn, &input)
}

func findEnvironment(ctx context.Context, conn *continuum.Client, input *continuum.GetEnvironmentInput) (*continuum.GetEnvironmentOutput, error) {
	output, err := conn.GetEnvironment(ctx, input)
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

func statusEnvironment(conn *continuum.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEnvironmentByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEnvironmentActive(ctx context.Context, conn *continuum.Client, id string, timeout time.Duration) (*continuum.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Target:  enum.Slice(awstypes.EnvironmentStatusActive),
		Refresh: statusEnvironment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*continuum.GetEnvironmentOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *continuum.Client, id string, timeout time.Duration) (*continuum.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EnvironmentStatusActive),
		Target:  []string{},
		Refresh: statusEnvironment(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*continuum.GetEnvironmentOutput); ok {
		return output, err
	}

	return nil, err
}

type environmentResourceModel struct {
	framework.WithRegionModel
	AllowedDomains             fwtypes.SetOfString                                    `tfsdk:"allowed_domains"`
	ARN                        types.String                                           `tfsdk:"arn"`
	CollectorDataAccessRoleARN types.String                                           `tfsdk:"collector_data_access_role_arn"`
	CreatedAt                  timetypes.RFC3339                                      `tfsdk:"created_at"`
	ID                         types.String                                           `tfsdk:"id"`
	Name                       types.String                                           `tfsdk:"name"`
	PatchCreationPolicy        fwtypes.StringEnum[awstypes.PatchCreationPolicy]       `tfsdk:"patch_creation_policy"`
	RepositorySources          fwtypes.ListNestedObjectValueOf[repositorySourceModel] `tfsdk:"repository_source"`
	S3Sources                  fwtypes.ListNestedObjectValueOf[s3SourceModel]         `tfsdk:"s3_source"`
	ScanFrequency              types.String                                           `tfsdk:"scan_frequency"`
	Status                     fwtypes.StringEnum[awstypes.EnvironmentStatus]         `tfsdk:"status"`
	StatusMessage              types.String                                           `tfsdk:"status_message"`
	Timeouts                   timeouts.Value                                         `tfsdk:"timeouts"`
	UpdatedAt                  timetypes.RFC3339                                      `tfsdk:"updated_at"`
}

type repositorySourceModel struct {
	DefaultBranch types.String `tfsdk:"default_branch"`
	IntegrationID types.String `tfsdk:"integration_id"`
	Repository    types.String `tfsdk:"repository"`
}

type s3SourceModel struct {
	CodePrefix    types.String `tfsdk:"code_prefix"`
	ContextPrefix types.String `tfsdk:"context_prefix"`
	IntegrationID types.String `tfsdk:"integration_id"`
}

func (m *environmentResourceModel) expand(ctx context.Context) (*awstypes.EnvironmentConfiguration, *awstypes.EnvironmentSourceConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	envConfig := &awstypes.EnvironmentConfiguration{}
	if !m.PatchCreationPolicy.IsNull() && !m.PatchCreationPolicy.IsUnknown() {
		envConfig.PatchCreationPolicy = m.PatchCreationPolicy.ValueEnum()
	}
	if !m.ScanFrequency.IsNull() && !m.ScanFrequency.IsUnknown() {
		envConfig.ScanFrequency = m.ScanFrequency.ValueStringPointer()
	}
	if !m.AllowedDomains.IsNull() {
		envConfig.Network = &awstypes.EnvironmentNetworkConfiguration{
			AllowedDomains: fwflex.ExpandFrameworkStringValueSet(ctx, m.AllowedDomains),
		}
	}

	srcConfig := &awstypes.EnvironmentSourceConfiguration{}

	repos, d := m.RepositorySources.ToSlice(ctx)
	diags.Append(d...)
	for _, v := range repos {
		srcConfig.RepositorySources = append(srcConfig.RepositorySources, awstypes.EnvironmentRepositorySource{
			DefaultBranch: fwflex.StringFromFramework(ctx, v.DefaultBranch),
			IntegrationId: v.IntegrationID.ValueStringPointer(),
			Repository:    v.Repository.ValueStringPointer(),
		})
	}

	s3s, d := m.S3Sources.ToSlice(ctx)
	diags.Append(d...)
	for _, v := range s3s {
		srcConfig.S3Sources = append(srcConfig.S3Sources, awstypes.EnvironmentS3IntegrationSource{
			CodePrefix:    v.CodePrefix.ValueStringPointer(),
			ContextPrefix: fwflex.StringFromFramework(ctx, v.ContextPrefix),
			IntegrationId: v.IntegrationID.ValueStringPointer(),
		})
	}

	return envConfig, srcConfig, diags
}

func (m *environmentResourceModel) flatten(ctx context.Context, output *continuum.GetEnvironmentOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ARN = types.StringPointerValue(output.EnvironmentArn)
	m.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	m.ID = types.StringPointerValue(output.EnvironmentId)
	m.Name = types.StringPointerValue(output.Name)
	m.Status = fwtypes.StringEnumValue(output.Status)
	m.StatusMessage = types.StringPointerValue(output.StatusMessage)
	m.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)

	var code *awstypes.CodeConfigurationDetails
	if output.Configuration != nil {
		code = output.Configuration.CodeConfiguration
	}
	if code == nil {
		m.RepositorySources = fwtypes.NewListNestedObjectValueOfNull[repositorySourceModel](ctx)
		m.S3Sources = fwtypes.NewListNestedObjectValueOfNull[s3SourceModel](ctx)
		return diags
	}

	if code.CollectorDetails != nil {
		m.CollectorDataAccessRoleARN = types.StringPointerValue(code.CollectorDetails.CollectorDataAccessRoleArn)
	} else {
		m.CollectorDataAccessRoleARN = types.StringNull()
	}

	if cfg := code.EnvironmentConfiguration; cfg != nil {
		m.PatchCreationPolicy = fwtypes.StringEnumValue(cfg.PatchCreationPolicy)
		m.ScanFrequency = types.StringPointerValue(cfg.ScanFrequency)
		if cfg.Network != nil && len(cfg.Network.AllowedDomains) > 0 {
			m.AllowedDomains = fwflex.FlattenFrameworkStringValueSetOfString(ctx, cfg.Network.AllowedDomains)
		} else {
			m.AllowedDomains = fwtypes.NewSetValueOfNull[types.String](ctx)
		}
	}

	if src := code.SourceConfiguration; src != nil {
		repos := make([]*repositorySourceModel, 0, len(src.RepositorySources))
		for _, v := range src.RepositorySources {
			repos = append(repos, &repositorySourceModel{
				DefaultBranch: types.StringPointerValue(v.DefaultBranch),
				IntegrationID: types.StringPointerValue(v.IntegrationId),
				Repository:    types.StringPointerValue(v.Repository),
			})
		}
		var d diag.Diagnostics
		m.RepositorySources, d = fwtypes.NewListNestedObjectValueOfSlice(ctx, repos, nil)
		diags.Append(d...)

		s3s := make([]*s3SourceModel, 0, len(src.S3Sources))
		for _, v := range src.S3Sources {
			s3s = append(s3s, &s3SourceModel{
				CodePrefix:    types.StringPointerValue(v.CodePrefix),
				ContextPrefix: types.StringPointerValue(v.ContextPrefix),
				IntegrationID: types.StringPointerValue(v.IntegrationId),
			})
		}
		m.S3Sources, d = fwtypes.NewListNestedObjectValueOfSlice(ctx, s3s, nil)
		diags.Append(d...)
	} else {
		m.RepositorySources = fwtypes.NewListNestedObjectValueOfNull[repositorySourceModel](ctx)
		m.S3Sources = fwtypes.NewListNestedObjectValueOfNull[s3SourceModel](ctx)
	}

	return diags
}
