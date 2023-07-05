// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	trustCreatedTimeout = 10 * time.Minute
	trustUpdatedTimeout = 10 * time.Minute
	trustDeleteTimeout  = 5 * time.Minute
)

// @FrameworkResource
func newResourceTrust(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTrust{}, nil
}

const (
	ResNameTrust = "Trust"
)

type resourceTrust struct {
	framework.ResourceWithConfigure
}

func (r *resourceTrust) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_directory_service_trust"
}

func (r *resourceTrust) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"conditional_forwarder_ip_addrs": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 4),
					setvalidator.ValueStringsAre(
						fwvalidators.IPv4Address(),
					),
				},
			},
			"created_date_time": schema.StringAttribute{
				Computed: true,
			},
			"delete_associated_conditional_forwarder": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					directoryIDValidator,
				},
			},
			"id": framework.IDAttribute(),
			"last_updated_date_time": schema.StringAttribute{
				Computed: true,
			},
			"remote_domain_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
					domainWithTrailingDotValidator,
				},
			},
			"selective_auth": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.SelectiveAuth](),
				},
			},
			"state_last_updated_date_time": schema.StringAttribute{
				Computed: true,
			},
			"trust_direction": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.TrustDirection](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trust_password": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(128),
					trustPasswordValidator,
				},
			},
			"trust_state": schema.StringAttribute{
				Computed: true,
			},
			"trust_state_reason": schema.StringAttribute{
				Computed: true,
			},
			"trust_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.TrustTypeForest)),
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.TrustType](),
				},
			},
		},
	}
}

func (r *resourceTrust) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DSClient(ctx)

	var plan resourceTrustData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	directoryID := plan.DirectoryID.ValueString()

	input := plan.createInput(ctx)

	output, err := conn.CreateTrust(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DS, create.ErrActionCreating, ResNameTrust, directoryID, nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = types.StringValue(aws.ToString(output.TrustId))

	// When Trust Direction is `One-Way: Incoming`, the Trust terminates at Created. Otherwise, it terminates at Verified
	var trust *awstypes.Trust
	if plan.TrustDirection.ValueString() == string(awstypes.TrustDirectionOneWayIncoming) {
		trust, err = waitTrustCreated(ctx, conn, state.DirectoryID.ValueString(), state.ID.ValueString(), trustCreatedTimeout)
	} else {
		trust, err = waitTrustVerified(ctx, conn, state.DirectoryID.ValueString(), state.ID.ValueString(), trustCreatedTimeout)
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.DS, create.ErrActionCreating, ResNameTrust, state.ID.ValueString(), err))
		return
	}

	state.update(ctx, trust)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceTrust) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DSClient(ctx)

	var data resourceTrustData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trust, err := FindTrustByID(ctx, conn, data.DirectoryID.ValueString(), data.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.DS, create.ErrActionReading, ResNameTrust, data.ID.ValueString(), err))
		return
	}

	data.update(ctx, trust)

	forwarder, err := findConditionalForwarder(ctx, conn, data.DirectoryID.ValueString(), data.RemoteDomainName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.DS, create.ErrActionReading, ResNameTrust, data.ID.ValueString(), fmt.Errorf("reading Conditional Forwarder: %w", err)))
		return
	}

	data.updateConditionalForwarder(ctx, forwarder)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceTrust) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, config, plan resourceTrustData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	if !plan.SelectiveAuth.IsUnknown() && !state.SelectiveAuth.Equal(plan.SelectiveAuth) {
		params := plan.updateInput(ctx)

		_, err := conn.UpdateTrust(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("updating Cognito User Pool Client (%s)", plan.ID.ValueString()),
				err.Error(),
			)
			return
		}

		trust, err := waitTrustUpdated(ctx, conn, state.DirectoryID.ValueString(), state.ID.ValueString(), trustUpdatedTimeout)
		if err != nil {
			resp.Diagnostics.Append(create.DiagErrorFramework(names.DS, create.ErrActionUpdating, ResNameTrust, state.ID.ValueString(), err))
			return
		}

		state.update(ctx, trust)
	}

	if !plan.ConditionalForwarderIpAddrs.IsUnknown() && !state.ConditionalForwarderIpAddrs.Equal(plan.ConditionalForwarderIpAddrs) {
		params := plan.updateConditionalForwarderInput(ctx)

		_, err := conn.UpdateConditionalForwarder(ctx, params)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("updating Cognito User Pool Client (%s) conditional forwarder IPs", plan.ID.ValueString()),
				err.Error(),
			)
			return
		}

		forwarder, err := findConditionalForwarder(ctx, conn, plan.DirectoryID.ValueString(), plan.RemoteDomainName.ValueString())
		if err != nil {
			// Outputting a NotFoundError does not include the original error.
			// Retrieve it to give the user an actionalble error message.
			if nfe, ok := errs.As[*retry.NotFoundError](err); ok {
				if nfe.LastError != nil {
					err = nfe.LastError
				}
			}
			resp.Diagnostics.Append(create.DiagErrorFramework(names.DS, create.ErrActionReading, ResNameTrust, plan.ID.ValueString(), fmt.Errorf("reading Conditional Forwarder: %w", err)))
			return
		}

		state.updateConditionalForwarder(ctx, forwarder)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTrust) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceTrustData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := state.deleteInput(ctx)

	conn := r.Meta().DSClient(ctx)

	_, err := conn.DeleteTrust(ctx, params)
	if isTrustNotFoundErr(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DS, create.ErrActionDeleting, ResNameTrust, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	_, err = waitTrustDeleted(ctx, conn, state.DirectoryID.ValueString(), state.ID.ValueString(), trustDeleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DS, create.ErrActionDeleting, ResNameTrust, state.ID.ValueString(), fmt.Errorf("waiting for completion: %w", err)),
			err.Error(),
		)
		return
	}
}

func (r *resourceTrust) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("Wrong format for import ID (%s), use: 'directory-id/remote-directory-domain'", req.ID))
		return
	}
	directoryID := parts[0]
	domain := parts[1]

	trust, err := findTrustByDomain(ctx, r.Meta().DSClient(ctx), directoryID, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Importing Resource",
			err.Error(),
		)
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), aws.ToString(trust.TrustId))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("directory_id"), directoryID)...)
}

type resourceTrustData struct {
	ConditionalForwarderIpAddrs          types.Set    `tfsdk:"conditional_forwarder_ip_addrs"`
	CreatedDateTime                      types.String `tfsdk:"created_date_time"`
	DeleteAssociatedConditionalForwarder types.Bool   `tfsdk:"delete_associated_conditional_forwarder"`
	DirectoryID                          types.String `tfsdk:"directory_id"`
	ID                                   types.String `tfsdk:"id"`
	LastUpdatedDateTime                  types.String `tfsdk:"last_updated_date_time"`
	RemoteDomainName                     types.String `tfsdk:"remote_domain_name"`
	SelectiveAuth                        types.String `tfsdk:"selective_auth"`
	StateLastUpdatedDateTime             types.String `tfsdk:"state_last_updated_date_time"`
	TrustDirection                       types.String `tfsdk:"trust_direction"`
	TrustPassword                        types.String `tfsdk:"trust_password"`
	TrustState                           types.String `tfsdk:"trust_state"`
	TrustStateReason                     types.String `tfsdk:"trust_state_reason"`
	TrustType                            types.String `tfsdk:"trust_type"`
}

func (data resourceTrustData) createInput(ctx context.Context) *directoryservice.CreateTrustInput {
	return &directoryservice.CreateTrustInput{
		ConditionalForwarderIpAddrs: flex.ExpandFrameworkStringValueSet(ctx, data.ConditionalForwarderIpAddrs),
		DirectoryId:                 flex.StringFromFramework(ctx, data.DirectoryID),
		RemoteDomainName:            flex.StringFromFramework(ctx, data.RemoteDomainName),
		SelectiveAuth:               stringlikeValueFromFramework[awstypes.SelectiveAuth](ctx, data.SelectiveAuth),
		TrustDirection:              stringlikeValueFromFramework[awstypes.TrustDirection](ctx, data.TrustDirection),
		TrustPassword:               flex.StringFromFramework(ctx, data.TrustPassword),
		TrustType:                   stringlikeValueFromFramework[awstypes.TrustType](ctx, data.TrustType),
	}
}

func (data resourceTrustData) updateInput(ctx context.Context) *directoryservice.UpdateTrustInput {
	return &directoryservice.UpdateTrustInput{
		TrustId:       flex.StringFromFramework(ctx, data.ID),
		SelectiveAuth: stringlikeValueFromFramework[awstypes.SelectiveAuth](ctx, data.SelectiveAuth),
	}
}

func (data resourceTrustData) updateConditionalForwarderInput(ctx context.Context) *directoryservice.UpdateConditionalForwarderInput {
	return &directoryservice.UpdateConditionalForwarderInput{
		DirectoryId:      flex.StringFromFramework(ctx, data.DirectoryID),
		RemoteDomainName: flex.StringFromFramework(ctx, data.RemoteDomainName),
		DnsIpAddrs:       flex.ExpandFrameworkStringValueSet(ctx, data.ConditionalForwarderIpAddrs),
	}
}

func (data resourceTrustData) deleteInput(ctx context.Context) *directoryservice.DeleteTrustInput {
	return &directoryservice.DeleteTrustInput{
		TrustId:                              flex.StringFromFramework(ctx, data.ID),
		DeleteAssociatedConditionalForwarder: data.DeleteAssociatedConditionalForwarder.ValueBool(),
	}
}

func (data *resourceTrustData) update(ctx context.Context, in *awstypes.Trust) {
	data.CreatedDateTime = flex.StringValueToFramework(ctx, in.CreatedDateTime.Format(time.RFC3339))
	data.LastUpdatedDateTime = flex.StringValueToFramework(ctx, in.LastUpdatedDateTime.Format(time.RFC3339))
	data.RemoteDomainName = flex.StringToFramework(ctx, in.RemoteDomainName)
	data.SelectiveAuth = flex.StringValueToFramework(ctx, in.SelectiveAuth)
	data.StateLastUpdatedDateTime = flex.StringValueToFramework(ctx, in.StateLastUpdatedDateTime.Format(time.RFC3339))
	data.TrustDirection = flex.StringValueToFramework(ctx, in.TrustDirection)
	// TrustPassword is not returned
	data.TrustState = flex.StringValueToFramework(ctx, in.TrustState)
	data.TrustStateReason = flex.StringToFramework(ctx, in.TrustStateReason)
	data.TrustType = flex.StringValueToFramework(ctx, in.TrustType)
}

func (data *resourceTrustData) updateConditionalForwarder(ctx context.Context, in *awstypes.ConditionalForwarder) {
	data.ConditionalForwarderIpAddrs = flex.FlattenFrameworkStringValueSet(ctx, in.DnsIpAddrs)
}

func stringlikeValueFromFramework[T ~string](_ context.Context, v types.String) T {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}

	return T(v.ValueString())
}

func FindTrustByID(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string) (*awstypes.Trust, error) {
	input := &directoryservice.DescribeTrustsInput{
		DirectoryId: aws.String(directoryID),
		TrustIds:    []string{trustID},
	}

	output, err := conn.DescribeTrusts(ctx, input)
	if isTrustNotFoundErr(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	trust, err := tfresource.AssertSingleValueResult(output.Trusts)
	if err != nil {
		return nil, err
	}
	return trust, nil
}

func findTrustByDomain(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, domain string) (*awstypes.Trust, error) {
	input := &directoryservice.DescribeTrustsInput{
		DirectoryId: aws.String(directoryID),
	}

	var results []awstypes.Trust
	paginator := directoryservice.NewDescribeTrustsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, trust := range page.Trusts {
			if aws.ToString(trust.RemoteDomainName) == domain {
				results = append(results, trust)
			}
		}
	}

	trust, err := tfresource.AssertSingleValueResult(results)
	if err != nil {
		return nil, err
	}
	return trust, nil
}

func isTrustNotFoundErr(err error) bool {
	return errs.IsA[*awstypes.EntityDoesNotExistException](err)
}

func isConditionalForwarderNotFoundErr(err error) bool {
	return errs.IsA[*awstypes.EntityDoesNotExistException](err)
}

// waitTrustCreated waits until a Trust is created.
func waitTrustCreated(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TrustStateCreating,
		),
		Target: enum.Slice(
			awstypes.TrustStateCreated,
		),
		Refresh: statusTrust(ctx, conn, directoryID, trustID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*awstypes.Trust); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.TrustStateReason)))

		return output, err
	}

	return nil, err
}

// waitTrustVerified waits until a Trust is created and verified.
// On first side of a Two-Way Trust relationship, `VerifyFailed` is expected. This then gets updated when the second side is created.
func waitTrustVerified(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TrustStateCreating,
			awstypes.TrustStateCreated,
			awstypes.TrustStateVerifying,
		),
		Target: enum.Slice(
			awstypes.TrustStateVerified,
			awstypes.TrustStateVerifyFailed,
		),
		Refresh: statusTrust(ctx, conn, directoryID, trustID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*awstypes.Trust); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.TrustStateReason)))

		return output, err
	}

	return nil, err
}

func waitTrustUpdated(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TrustStateUpdating,
			awstypes.TrustStateUpdated,
			awstypes.TrustStateVerifying,
		),
		Target: enum.Slice(
			awstypes.TrustStateVerified,
			awstypes.TrustStateVerifyFailed,
		),
		Refresh: statusTrust(ctx, conn, directoryID, trustID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*awstypes.Trust); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.TrustStateReason)))

		return output, err
	}

	return nil, err
}

func waitTrustDeleted(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TrustStateCreated,
			awstypes.TrustStateDeleting,
			awstypes.TrustStateDeleted,
		),
		Target:  []string{},
		Refresh: statusTrust(ctx, conn, directoryID, trustID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	// Wrap any error returned with waiting message
	defer func() {
		if err != nil {
			err = fmt.Errorf("waiting for completion: %w", err)
		}
	}()

	if output, ok := outputRaw.(*awstypes.Trust); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.TrustStateReason)))

		return output, err
	}

	return nil, err
}

func statusTrust(ctx context.Context, conn directoryservice.DescribeTrustsAPIClient, directoryID, trustID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTrustByID(ctx, conn, directoryID, trustID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TrustState), nil
	}
}

func findConditionalForwarder(ctx context.Context, conn *directoryservice.Client, directoryID, remoteDomainName string) (*awstypes.ConditionalForwarder, error) {
	// Directory Trust optionally accepts a remote domain name with a trailing period.
	// Conditional Forwarders
	remoteDomainName = strings.TrimRight(remoteDomainName, ".")

	input := &directoryservice.DescribeConditionalForwardersInput{
		DirectoryId:       aws.String(directoryID),
		RemoteDomainNames: []string{remoteDomainName},
	}

	output, err := conn.DescribeConditionalForwarders(ctx, input)
	if isConditionalForwarderNotFoundErr(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	forwarder, err := tfresource.AssertSingleValueResult(output.ConditionalForwarders)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}
