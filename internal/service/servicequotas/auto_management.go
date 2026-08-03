// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_servicequotas_auto_management", name="Auto Management")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(serialize=true)
func newAutoManagementResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &autoManagementResource{}

	return r, nil
}

type autoManagementResource struct {
	framework.ResourceWithModel[autoManagementResourceModel]
	framework.WithImportByIdentity
}

func (r *autoManagementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"exclusion_list": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				CustomType: fwtypes.NewMapTypeOf[fwtypes.ListOfString](ctx),
				Optional:   true,
				Validators: []validator.Map{
					mapvalidator.NoNullValues(),
					mapvalidator.SizeAtLeast(1),
					mapvalidator.ValueListsAre(listvalidator.SizeAtLeast(1)),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			"notification_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							// If the value is being removed (Config is Null, but State is not Null)
							resp.RequiresReplace = req.ConfigValue.IsNull() && !req.StateValue.IsNull()
						},
						"The Service Quotas API does not support unsetting the Notification ARN once it has been configured. The resource must be recreated to reset this setting.",
						"The Service Quotas API does not support unsetting the Notification ARN once it has been configured. The resource must be recreated to reset this setting.",
					),
				},
			},
			"opt_in_level": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OptInLevel](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"opt_in_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OptInType](),
				Required:   true,
			},
		},
	}
}

func (r *autoManagementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var plan autoManagementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input servicequotas.StartAutoManagementInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithIgnoredFieldNames([]string{"ExclusionList"})))
	if resp.Diagnostics.HasError() {
		return
	}

	exclusionList, diags := expandExclusionList(ctx, plan.ExclusionList)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ExclusionList = exclusionList

	_, err := conn.StartAutoManagement(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *autoManagementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state autoManagementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutoManagement(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithIgnoredFieldNames([]string{"ExclusionList"})))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ExclusionList = flattenExclusionList(ctx, out.ExclusionList)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *autoManagementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var plan, state autoManagementResourceModel
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
		var input servicequotas.UpdateAutoManagementInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithIgnoredFieldNames([]string{"ExclusionList"})))
		if resp.Diagnostics.HasError() {
			return
		}

		exclusionList, diags := expandExclusionList(ctx, plan.ExclusionList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ExclusionList = exclusionList

		_, err := conn.UpdateAutoManagement(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *autoManagementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ServiceQuotasClient(ctx)

	var state autoManagementResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := servicequotas.StopAutoManagementInput{}
	_, err := conn.StopAutoManagement(ctx, &input)
	if errs.IsA[*awstypes.NoSuchResourceException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}
}

func findAutoManagement(ctx context.Context, conn *servicequotas.Client) (*servicequotas.GetAutoManagementConfigurationOutput, error) {
	var input servicequotas.GetAutoManagementConfigurationInput
	out, err := conn.GetAutoManagementConfiguration(ctx, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if status := out.OptInStatus; status == awstypes.OptInStatusDisabled {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: string(status),
		})
	}

	return out, nil
}

type autoManagementResourceModel struct {
	framework.WithRegionModel
	ExclusionList   fwtypes.MapValueOf[fwtypes.ListOfString] `tfsdk:"exclusion_list"`
	ID              types.String                             `tfsdk:"id"`
	NotificationARN fwtypes.ARN                              `tfsdk:"notification_arn"`
	OptInLevel      fwtypes.StringEnum[awstypes.OptInLevel]  `tfsdk:"opt_in_level"`
	OptInType       fwtypes.StringEnum[awstypes.OptInType]   `tfsdk:"opt_in_type"`
}

func expandExclusionList(ctx context.Context, tfMap fwtypes.MapValueOf[fwtypes.ListOfString]) (map[string][]string, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-expander-functions
	var diags diag.Diagnostics

	if tfMap.IsNull() || tfMap.IsUnknown() {
		// Return an empty map to clear a configured exclusion list in the API; nil would leave it unchanged
		return make(map[string][]string, 0), diags
	}

	apiMap := make(map[string][]string, len(tfMap.Elements()))

	diags.Append(tfMap.ElementsAs(ctx, &apiMap, false)...)

	return apiMap, diags
}

func flattenExclusionList(ctx context.Context, apiMap map[string][]awstypes.QuotaInfo) fwtypes.MapValueOf[fwtypes.ListOfString] { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	if len(apiMap) == 0 {
		return fwtypes.NewMapValueOfNull[fwtypes.ListOfString](ctx)
	}

	elements := make(map[string]attr.Value, len(apiMap))

	for k, v := range apiMap {
		attrValues := make([]attr.Value, len(v))

		for i, q := range v {
			attrValues[i] = types.StringValue(aws.ToString(q.QuotaCode))
		}

		elements[k] = fwtypes.NewListValueOfMust[types.String](ctx, attrValues)
	}

	return fwtypes.NewMapValueOfMust[fwtypes.ListOfString](ctx, elements)
}
