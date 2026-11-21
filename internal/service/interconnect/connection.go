// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_interconnect_connection", name="Connection")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/interconnect/types;awstypes;awstypes.Connection")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(tagsTest=false)
// @Testing(identityTest=false)
func newConnectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &connectionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type connectionResource struct {
	framework.ResourceWithModel[connectionResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *connectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"activation_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"bandwidth": schema.StringAttribute{
				Required: true,
			},
			"billing_tier": schema.Int32Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"interconnect_provider": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrLocation: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_account": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"shared_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionState](),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"attach_point": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[attachPointModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName(names.AttrARN),
									path.MatchRelative().AtParent().AtName("direct_connect_gateway"),
								),
							},
						},
						"direct_connect_gateway": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"remote_account": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[remoteAccountModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIdentifier: schema.StringAttribute{
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

func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().InterconnectClient(ctx)

	var plan connectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input interconnect.CreateConnectionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateConnection(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Description.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.Connection, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	connection, err := waitConnectionCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, connection, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.InterconnectProvider = flattenProvider(connection.Provider)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().InterconnectClient(ctx)

	var state connectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// On import via ARN, the short ID is not yet populated; derive it from the ARN.
	if state.ID.ValueString() == "" && state.ARN.ValueString() != "" {
		id, err := connectionIDFromARN(state.ARN.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
			return
		}
		state.ID = types.StringValue(id)
	}

	out, err := findConnectionByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.InterconnectProvider = flattenProvider(out.Provider)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().InterconnectClient(ctx)

	var plan, state connectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Bandwidth.Equal(state.Bandwidth) || !plan.Description.Equal(state.Description) {
		input := interconnect.UpdateConnectionInput{
			Identifier: plan.ID.ValueStringPointer(),
		}
		if !plan.Bandwidth.Equal(state.Bandwidth) {
			input.Bandwidth = plan.Bandwidth.ValueStringPointer()
		}
		if !plan.Description.Equal(state.Description) {
			input.Description = plan.Description.ValueStringPointer()
		}

		_, err := conn.UpdateConnection(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		connection, err := waitConnectionUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, connection, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.InterconnectProvider = flattenProvider(connection.Provider)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().InterconnectClient(ctx)

	var state connectionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := interconnect.DeleteConnectionInput{
		Identifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteConnection(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	err = waitConnectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

// connectionIDFromARN extracts the short connection ID from a connection ARN.
// The resource portion of the ARN has the form "connection/mcc-abcd1234".
func connectionIDFromARN(s string) (string, error) {
	parsed, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	_, id, found := strings.Cut(parsed.Resource, "/")
	if !found {
		return "", fmt.Errorf("unexpected ARN resource format: %q", parsed.Resource)
	}

	return id, nil
}

func findConnectionByID(ctx context.Context, conn *interconnect.Client, id string) (*awstypes.Connection, error) {
	input := interconnect.GetConnectionInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetConnection(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Connection == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	// A deleted connection lingers in the API in the "deleted" state rather than
	// returning a not-found error. Treat it as not found.
	if out.Connection.State == awstypes.ConnectionStateDeleted {
		return nil, &retry.NotFoundError{
			LastError: fmt.Errorf("Interconnect Connection (%s) in state %q", id, out.Connection.State),
		}
	}

	return out.Connection, nil
}

func statusConnection(conn *interconnect.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findConnectionByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func waitConnectionCreated(ctx context.Context, conn *interconnect.Client, id string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ConnectionStatePending),
		Target:                    enum.Slice(awstypes.ConnectionStateRequested, awstypes.ConnectionStateAvailable),
		Refresh:                   statusConnection(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Connection); ok {
		return out, err
	}

	return nil, err
}

func waitConnectionUpdated(ctx context.Context, conn *interconnect.Client, id string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStateUpdating),
		Target:  enum.Slice(awstypes.ConnectionStateRequested, awstypes.ConnectionStateAvailable),
		Refresh: statusConnection(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Connection); ok {
		return out, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *interconnect.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStateDeleting, awstypes.ConnectionStateAvailable, awstypes.ConnectionStateRequested, awstypes.ConnectionStatePending),
		Target:  []string{},
		Refresh: statusConnection(conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

type connectionResourceModel struct {
	framework.WithRegionModel
	ActivationKey        types.String                                        `tfsdk:"activation_key"`
	ARN                  types.String                                        `tfsdk:"arn"`
	AttachPoint          fwtypes.ListNestedObjectValueOf[attachPointModel]   `tfsdk:"attach_point"`
	Bandwidth            types.String                                        `tfsdk:"bandwidth"`
	BillingTier          types.Int32                                         `tfsdk:"billing_tier"`
	Description          types.String                                        `tfsdk:"description"`
	EnvironmentID        types.String                                        `tfsdk:"environment_id"`
	ID                   types.String                                        `tfsdk:"id"`
	InterconnectProvider types.String                                        `tfsdk:"interconnect_provider" autoflex:"-"`
	Location             types.String                                        `tfsdk:"location"`
	OwnerAccount         types.String                                        `tfsdk:"owner_account"`
	RemoteAccount        fwtypes.ListNestedObjectValueOf[remoteAccountModel] `tfsdk:"remote_account"`
	SharedID             types.String                                        `tfsdk:"shared_id"`
	State                fwtypes.StringEnum[awstypes.ConnectionState]        `tfsdk:"state"`
	Tags                 tftags.Map                                          `tfsdk:"tags"`
	TagsAll              tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                      `tfsdk:"timeouts"`
	Type                 types.String                                        `tfsdk:"type"`
}

type attachPointModel struct {
	ARN                  fwtypes.ARN  `tfsdk:"arn"`
	DirectConnectGateway types.String `tfsdk:"direct_connect_gateway"`
}

var _ fwflex.Expander = attachPointModel{}

func (m attachPointModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.DirectConnectGateway.IsNull():
		return &awstypes.AttachPointMemberDirectConnectGateway{
			Value: m.DirectConnectGateway.ValueString(),
		}, diags
	case !m.ARN.IsNull():
		return &awstypes.AttachPointMemberArn{
			Value: m.ARN.ValueString(),
		}, diags
	}

	return nil, diags
}

var _ fwflex.Flattener = &attachPointModel{}

func (m *attachPointModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.AttachPointMemberDirectConnectGateway:
		m.DirectConnectGateway = types.StringValue(t.Value)
	case awstypes.AttachPointMemberArn:
		m.ARN = fwtypes.ARNValue(t.Value)
	}

	return diags
}

// flattenProvider returns the provider name from the SDK Provider tagged union.
// The connection's "type" attribute distinguishes Multicloud vs LastMile, so a
// single string attribute is sufficient.
func flattenProvider(provider awstypes.Provider) types.String {
	switch v := provider.(type) {
	case *awstypes.ProviderMemberCloudServiceProvider:
		return types.StringValue(v.Value)
	case *awstypes.ProviderMemberLastMileProvider:
		return types.StringValue(v.Value)
	default:
		return types.StringNull()
	}
}

type remoteAccountModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

var _ fwflex.Expander = remoteAccountModel{}

func (m remoteAccountModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	return &awstypes.RemoteAccountIdentifierMemberIdentifier{
		Value: m.Identifier.ValueString(),
	}, diags
}
