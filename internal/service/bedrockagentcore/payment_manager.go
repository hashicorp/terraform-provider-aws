// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_payment_manager", name="Payment Manager")
// @Tags(identifierAttribute="payment_manager_arn")
// @IdentityAttribute("payment_manager_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="payment_manager_id")
// @Testing(generator="testAccRandomPaymentManagerName(t)")
// @Testing(serialize=true)
func newPaymentManagerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &paymentManagerResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type paymentManagerResource struct {
	framework.ResourceWithModel[paymentManagerResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *paymentManagerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authorizer_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PaymentsAuthorizerType](),
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 48),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`), "must start with a letter and contain only letters and numbers"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"payment_manager_arn": framework.ARNAttributeComputedOnly(),
			"payment_manager_id":  framework.IDAttribute(),
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PaymentManagerStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
			"workload_identity_details": framework.ResourceComputedListOfObjectsAttribute[workloadIdentityDetailsModel](ctx, listplanmodifier.UseStateForUnknown()),
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

func (r *paymentManagerResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data paymentManagerResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if data.AuthorizerType.ValueEnum() == awstypes.PaymentsAuthorizerTypeCustomJwt && data.AuthorizerConfiguration.IsNull() {
		response.Diagnostics.AddAttributeError(
			path.Root("authorizer_configuration"),
			"Missing Configuration",
			"authorizer_configuration is required when authorizer_type is CUSTOM_JWT",
		)
	}
}

func (r *paymentManagerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data paymentManagerResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreatePaymentManagerInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	// The execution role's permissions may not have propagated yet.
	var out *bedrockagentcorecontrol.CreatePaymentManagerOutput
	err := tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		out, err = conn.CreatePaymentManager(ctx, &input)

		if tfawserr.ErrMessageContains(err, errCodeValidationException, "execution role is missing the required permission") {
			return tfresource.RetryableError(err)
		}
		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	paymentManagerID := aws.ToString(out.PaymentManagerId)

	paymentManager, err := waitPaymentManagerCreated(ctx, conn, paymentManagerID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("payment_manager_id"), paymentManagerID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, paymentManager, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentManagerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data paymentManagerResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	paymentManagerID := fwflex.StringValueFromFramework(ctx, data.PaymentManagerID)
	out, err := findPaymentManagerByID(ctx, conn, paymentManagerID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentManagerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old paymentManagerResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		paymentManagerID := fwflex.StringValueFromFramework(ctx, new.PaymentManagerID)
		var input bedrockagentcorecontrol.UpdatePaymentManagerInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdatePaymentManager(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
			return
		}

		if _, err := waitPaymentManagerUpdated(ctx, conn, paymentManagerID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *paymentManagerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data paymentManagerResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	paymentManagerID := fwflex.StringValueFromFramework(ctx, data.PaymentManagerID)
	input := bedrockagentcorecontrol.DeletePaymentManagerInput{
		PaymentManagerId: aws.String(paymentManagerID),
	}

	_, err := conn.DeletePaymentManager(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
		return
	}

	if _, err := waitPaymentManagerDeleted(ctx, conn, paymentManagerID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentManagerID)
		return
	}
}

func waitPaymentManagerCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PaymentManagerStatusCreating),
		Target:                    enum.Slice(awstypes.PaymentManagerStatusReady),
		Refresh:                   statusPaymentManager(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentManagerOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPaymentManagerUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PaymentManagerStatusUpdating),
		Target:                    enum.Slice(awstypes.PaymentManagerStatusReady),
		Refresh:                   statusPaymentManager(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentManagerOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPaymentManagerDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentManagerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PaymentManagerStatusDeleting, awstypes.PaymentManagerStatusReady),
		Target:  []string{},
		Refresh: statusPaymentManager(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentManagerOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPaymentManager(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPaymentManagerByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findPaymentManagerByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetPaymentManagerOutput, error) {
	input := bedrockagentcorecontrol.GetPaymentManagerInput{
		PaymentManagerId: aws.String(id),
	}

	out, err := conn.GetPaymentManager(ctx, &input)

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

type paymentManagerResourceModel struct {
	framework.WithRegionModel
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
	AuthorizerType          fwtypes.StringEnum[awstypes.PaymentsAuthorizerType]           `tfsdk:"authorizer_type"`
	Description             types.String                                                  `tfsdk:"description"`
	Name                    types.String                                                  `tfsdk:"name"`
	PaymentManagerARN       types.String                                                  `tfsdk:"payment_manager_arn"`
	PaymentManagerID        types.String                                                  `tfsdk:"payment_manager_id"`
	RoleARN                 fwtypes.ARN                                                   `tfsdk:"role_arn"`
	Status                  fwtypes.StringEnum[awstypes.PaymentManagerStatus]             `tfsdk:"status"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
	WorkloadIdentityDetails fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel] `tfsdk:"workload_identity_details"`
}
