// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_registry", name="Registry")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="randomWithPrefixAndUnderscore(t)")
// @Testing(importStateIdAttribute="registry_id")
// @IdentityAttribute("registry_id")
func newRegistryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRegistry = "Registry"
)

type registryResource struct {
	framework.ResourceWithModel[registryResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *registryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auto_approval": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"authorizer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryAuthorizerType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z0-9_-]+$`),
						"must contain only letters, numbers, hyphens, and underscores",
					),
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"registry_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registry_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegistryStatus](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"authorizer_configuration": authorizerConfigurationSchema(ctx),
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *registryResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data registryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AuthorizerType.ValueEnum() == awstypes.RegistryAuthorizerTypeCustomJwt && data.AuthorizerConfiguration.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("authorizer_configuration"),
			"Missing Required Attribute",
			"authorizer_configuration is required when authorizer_type is CUSTOM_JWT",
		)
	}
}

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan registryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateRegistryInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))

	// auto_approval is Optional+Computed and maps to a nested AWS struct,
	// so it is handled outside of AutoFlex.
	if !plan.AutoApproval.IsNull() && !plan.AutoApproval.IsUnknown() {
		input.ApprovalConfiguration = &awstypes.ApprovalConfiguration{
			AutoApproval: plan.AutoApproval.ValueBool(),
		}
	}

	out, err := conn.CreateRegistry(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	registryARN := aws.ToString(out.RegistryArn)

	created, err := waitRegistryCreated(ctx, conn, registryARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryARN)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	flattenRegistryApproval(created, &plan)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state registryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRegistryByID(ctx, conn, state.RegistryID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RegistryID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	flattenRegistryApproval(out, &state)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state registryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.AutoApproval.Equal(state.AutoApproval) ||
		!plan.AuthorizerConfiguration.Equal(state.AuthorizerConfiguration) {
		input := bedrockagentcorecontrol.UpdateRegistryInput{
			RegistryId: plan.RegistryID.ValueStringPointer(),
		}

		if !plan.Name.Equal(state.Name) {
			input.Name = plan.Name.ValueStringPointer()
		}

		if !plan.Description.Equal(state.Description) {
			descriptionValue := plan.Description.ValueString()
			input.Description = &awstypes.UpdatedDescription{
				OptionalValue: &descriptionValue,
			}
		}

		if !plan.AutoApproval.Equal(state.AutoApproval) && !plan.AutoApproval.IsUnknown() {
			input.ApprovalConfiguration = &awstypes.UpdatedApprovalConfiguration{
				OptionalValue: &awstypes.ApprovalConfiguration{
					AutoApproval: plan.AutoApproval.ValueBool(),
				},
			}
		}

		if !plan.AuthorizerConfiguration.Equal(state.AuthorizerConfiguration) {
			authorizerConfiguration, d := plan.AuthorizerConfiguration.ToPtr(ctx)
			smerr.AddEnrich(ctx, &resp.Diagnostics, d)
			if resp.Diagnostics.HasError() {
				return
			}

			if authorizerConfiguration != nil {
				expanded, d := authorizerConfiguration.ExpandTo(ctx, reflect.TypeFor[awstypes.UpdatedAuthorizerConfiguration]())
				smerr.AddEnrich(ctx, &resp.Diagnostics, d)
				if resp.Diagnostics.HasError() {
					return
				}
				input.AuthorizerConfiguration = expanded.(*awstypes.UpdatedAuthorizerConfiguration)
			}
		}

		_, err := conn.UpdateRegistry(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RegistryID.String())
			return
		}

		updated, err := waitRegistryUpdated(ctx, conn, plan.RegistryID.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.RegistryID.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, updated, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
		flattenRegistryApproval(updated, &plan)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state registryResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	registryID := state.RegistryID.ValueString()

	input := bedrockagentcorecontrol.DeleteRegistryInput{
		RegistryId: aws.String(registryID),
	}

	_, err := conn.DeleteRegistry(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}

	if _, err := waitRegistryDeleted(ctx, conn, registryID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}
}

func waitRegistryCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusCreating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryStatusDeleting, awstypes.RegistryStatusReady),
		Target:  []string{},
		Refresh: statusRegistry(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetRegistryOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusRegistry(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRegistryByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findRegistryByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetRegistryOutput, error) {
	input := bedrockagentcorecontrol.GetRegistryInput{
		RegistryId: aws.String(id),
	}

	out, err := conn.GetRegistry(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type registryResourceModel struct {
	framework.WithRegionModel
	AutoApproval            types.Bool                                                    `tfsdk:"auto_approval"`
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
	AuthorizerType          fwtypes.StringEnum[awstypes.RegistryAuthorizerType]           `tfsdk:"authorizer_type"`
	Description             types.String                                                  `tfsdk:"description"`
	Name                    types.String                                                  `tfsdk:"name"`
	RegistryARN             types.String                                                  `tfsdk:"registry_arn"`
	RegistryID              types.String                                                  `tfsdk:"registry_id"`
	Status                  fwtypes.StringEnum[awstypes.RegistryStatus]                   `tfsdk:"status"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

// flattenRegistryApproval maps the nested AWS ApprovalConfiguration onto the flat
// auto_approval attribute. AWS always returns this configuration, defaulting
// auto_approval to false.
func flattenRegistryApproval(out *bedrockagentcorecontrol.GetRegistryOutput, m *registryResourceModel) {
	if out.ApprovalConfiguration != nil {
		m.AutoApproval = types.BoolValue(out.ApprovalConfiguration.AutoApproval)
		return
	}
	m.AutoApproval = types.BoolValue(false)
}
