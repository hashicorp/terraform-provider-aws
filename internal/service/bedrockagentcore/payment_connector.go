// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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

// @FrameworkResource("aws_bedrockagentcore_payment_connector", name="Payment Connector")
// @IdentityAttribute("payment_manager_id")
// @IdentityAttribute("payment_connector_id")
// @ImportIDHandler(paymentConnectorImportID)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccPaymentConnectorImportStateIDFunc")
// @Testing(importStateIdAttribute="payment_connector_id")
// @Testing(importIgnore="credential_provider_configuration")
// @Testing(generator="testAccRandomPaymentConnectorName(t)")
// @Testing(serialize=true)
func newPaymentConnectorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &paymentConnectorResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type paymentConnectorResource struct {
	framework.ResourceWithModel[paymentConnectorResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *paymentConnectorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 48),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`), "must start with a letter and contain only letters, numbers, and underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"payment_connector_id": framework.IDAttribute(),
			"payment_manager_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PaymentConnectorStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PaymentConnectorType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"credential_provider_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[credentialsProviderConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"coinbase_cdp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[paymentCredentialProviderConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("coinbase_cdp"),
									path.MatchRelative().AtParent().AtName("stripe_privy"),
								),
							},
							NestedObject: paymentCredentialProviderConfigurationSchema(),
						},
						"stripe_privy": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[paymentCredentialProviderConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("coinbase_cdp"),
									path.MatchRelative().AtParent().AtName("stripe_privy"),
								),
							},
							NestedObject: paymentCredentialProviderConfigurationSchema(),
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

func paymentCredentialProviderConfigurationSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"credential_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
	}
}

func (r *paymentConnectorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data paymentConnectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreatePaymentConnectorInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))

	out, err := conn.CreatePaymentConnector(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	paymentManagerID, paymentConnectorID := aws.ToString(out.PaymentManagerId), aws.ToString(out.PaymentConnectorId)

	paymentConnector, err := waitPaymentConnectorCreated(ctx, conn, paymentManagerID, paymentConnectorID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root("payment_manager_id"), paymentManagerID)
		response.State.SetAttribute(ctx, path.Root("payment_connector_id"), paymentConnectorID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
		return
	}

	// Preserve write-only configuration from the plan; the GET response does not echo it back identically.
	credentialProviderConfigurations := data.CredentialProviderConfigurations
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, paymentConnector, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.CredentialProviderConfigurations = credentialProviderConfigurations

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentConnectorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data paymentConnectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	paymentManagerID, paymentConnectorID := fwflex.StringValueFromFramework(ctx, data.PaymentManagerID), fwflex.StringValueFromFramework(ctx, data.PaymentConnectorID)
	out, err := findPaymentConnectorByTwoPartKey(ctx, conn, paymentManagerID, paymentConnectorID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
		return
	}

	// Preserve the configured credential provider configuration; the resource ARNs
	// round-trip, so flatten everything except the write-managed block.
	credentialProviderConfigurations := data.CredentialProviderConfigurations
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.CredentialProviderConfigurations = credentialProviderConfigurations

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *paymentConnectorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old paymentConnectorResourceModel
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
		paymentManagerID, paymentConnectorID := fwflex.StringValueFromFramework(ctx, new.PaymentManagerID), fwflex.StringValueFromFramework(ctx, new.PaymentConnectorID)
		var input bedrockagentcorecontrol.UpdatePaymentConnectorInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdatePaymentConnector(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
			return
		}

		if _, err := waitPaymentConnectorUpdated(ctx, conn, paymentManagerID, paymentConnectorID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *paymentConnectorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data paymentConnectorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	paymentManagerID, paymentConnectorID := fwflex.StringValueFromFramework(ctx, data.PaymentManagerID), fwflex.StringValueFromFramework(ctx, data.PaymentConnectorID)
	input := bedrockagentcorecontrol.DeletePaymentConnectorInput{
		PaymentConnectorId: aws.String(paymentConnectorID),
		PaymentManagerId:   aws.String(paymentManagerID),
	}

	_, err := conn.DeletePaymentConnector(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
		return
	}

	if _, err := waitPaymentConnectorDeleted(ctx, conn, paymentManagerID, paymentConnectorID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, paymentConnectorID)
		return
	}
}

func waitPaymentConnectorCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, paymentManagerID, paymentConnectorID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PaymentConnectorStatusCreating),
		Target:                    enum.Slice(awstypes.PaymentConnectorStatusReady),
		Refresh:                   statusPaymentConnector(conn, paymentManagerID, paymentConnectorID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPaymentConnectorUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, paymentManagerID, paymentConnectorID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.PaymentConnectorStatusUpdating),
		Target:                    enum.Slice(awstypes.PaymentConnectorStatusReady),
		Refresh:                   statusPaymentConnector(conn, paymentManagerID, paymentConnectorID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitPaymentConnectorDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, paymentManagerID, paymentConnectorID string, timeout time.Duration) (*bedrockagentcorecontrol.GetPaymentConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PaymentConnectorStatusDeleting, awstypes.PaymentConnectorStatusReady),
		Target:  []string{},
		Refresh: statusPaymentConnector(conn, paymentManagerID, paymentConnectorID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetPaymentConnectorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusPaymentConnector(conn *bedrockagentcorecontrol.Client, paymentManagerID, paymentConnectorID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findPaymentConnectorByTwoPartKey(ctx, conn, paymentManagerID, paymentConnectorID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findPaymentConnectorByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, paymentManagerID, paymentConnectorID string) (*bedrockagentcorecontrol.GetPaymentConnectorOutput, error) {
	input := bedrockagentcorecontrol.GetPaymentConnectorInput{
		PaymentConnectorId: aws.String(paymentConnectorID),
		PaymentManagerId:   aws.String(paymentManagerID),
	}

	out, err := conn.GetPaymentConnector(ctx, &input)

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

var (
	_ inttypes.ImportIDParser = paymentConnectorImportID{}
)

type paymentConnectorImportID struct{}

func (paymentConnectorImportID) Parse(id string) (string, map[string]any, error) {
	const (
		paymentConnectorIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, paymentConnectorIDParts, true)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"payment_manager_id":   parts[0],
		"payment_connector_id": parts[1],
	}

	return id, result, nil
}

type paymentConnectorResourceModel struct {
	framework.WithRegionModel
	CredentialProviderConfigurations fwtypes.ListNestedObjectValueOf[credentialsProviderConfigurationModel] `tfsdk:"credential_provider_configuration"`
	Description                      types.String                                                           `tfsdk:"description"`
	Name                             types.String                                                           `tfsdk:"name"`
	PaymentConnectorID               types.String                                                           `tfsdk:"payment_connector_id"`
	PaymentManagerID                 types.String                                                           `tfsdk:"payment_manager_id"`
	Status                           fwtypes.StringEnum[awstypes.PaymentConnectorStatus]                    `tfsdk:"status"`
	Type                             fwtypes.StringEnum[awstypes.PaymentConnectorType]                      `tfsdk:"type"`
	Timeouts                         timeouts.Value                                                         `tfsdk:"timeouts"`
}

type credentialsProviderConfigurationModel struct {
	CoinbaseCDP fwtypes.ListNestedObjectValueOf[paymentCredentialProviderConfigurationModel] `tfsdk:"coinbase_cdp"`
	StripePrivy fwtypes.ListNestedObjectValueOf[paymentCredentialProviderConfigurationModel] `tfsdk:"stripe_privy"`
}

type paymentCredentialProviderConfigurationModel struct {
	CredentialProviderARN fwtypes.ARN `tfsdk:"credential_provider_arn"`
}

var (
	_ fwflex.Expander  = credentialsProviderConfigurationModel{}
	_ fwflex.Flattener = &credentialsProviderConfigurationModel{}
)

func (m credentialsProviderConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CoinbaseCDP.IsNull():
		data, d := m.CoinbaseCDP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialsProviderConfigurationMemberCoinbaseCDP
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags

	case !m.StripePrivy.IsNull():
		data, d := m.StripePrivy.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.CredentialsProviderConfigurationMemberStripePrivy
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		return &r, diags
	}
	return nil, diags
}

func (m *credentialsProviderConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CredentialsProviderConfigurationMemberCoinbaseCDP:
		var model paymentCredentialProviderConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.CoinbaseCDP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	case awstypes.CredentialsProviderConfigurationMemberStripePrivy:
		var model paymentCredentialProviderConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.StripePrivy = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("credential provider configuration flatten: %T", v),
		)
	}
	return diags
}
