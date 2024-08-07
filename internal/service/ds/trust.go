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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Trust")
func newTrustResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &trustResource{}, nil
}

type trustResource struct {
	framework.ResourceWithConfigure
}

func (*trustResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_directory_service_trust"
}

func (r *trustResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	trustType := fwtypes.StringEnumType[awstypes.TrustType]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"conditional_forwarder_ip_addrs": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
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
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			names.AttrID: framework.IDAttribute(),
			"last_updated_date_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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
				CustomType: fwtypes.StringEnumType[awstypes.SelectiveAuth](),
				Optional:   true,
				Computed:   true,
			},
			"state_last_updated_date_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"trust_direction": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TrustDirection](),
				Required:   true,
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
				CustomType: trustType,
				Optional:   true,
				Computed:   true,
				Default:    trustType.AttributeDefault(awstypes.TrustTypeForest),
			},
		},
	}
}

func (r *trustResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data trustResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := data.DirectoryID.ValueString()
	input := &directoryservice.CreateTrustInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateTrust(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Directory Service Trust (%s)", directoryID), err.Error())

		return
	}

	// Set values for unknowns.
	trustID := aws.ToString(output.TrustId)
	data.ID = types.StringValue(trustID)

	// When Trust Direction is `One-Way: Incoming`, the Trust terminates at Created. Otherwise, it terminates at Verified.
	const (
		timeout = 10 * time.Minute
	)
	var trust *awstypes.Trust
	if data.TrustDirection.ValueEnum() == awstypes.TrustDirectionOneWayIncoming {
		trust, err = waitTrustCreated(ctx, conn, directoryID, trustID, timeout)
	} else {
		trust, err = waitTrustVerified(ctx, conn, directoryID, trustID, timeout)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Directory Service Trust (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	response.Diagnostics.Append(fwflex.Flatten(ctx, trust, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *trustResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data trustResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := data.DirectoryID.ValueString()
	trustID := data.ID.ValueString()
	trust, err := findTrustByTwoPartKey(ctx, conn, directoryID, trustID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Directory Service Trust (%s)", trustID), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, trust, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Directory Trust optionally accepts a remote domain name with a trailing period.
	domainName := strings.TrimRight(data.RemoteDomainName.ValueString(), ".")
	forwarder, err := findConditionalForwarderByTwoPartKey(ctx, conn, directoryID, domainName)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Directory Service Conditional Forwarder (%s)", conditionalForwarderCreateResourceID(directoryID, domainName)), err.Error())

		return
	}

	data.ConditionalForwarderIPAddrs = fwflex.FlattenFrameworkStringValueSet(ctx, forwarder.DnsIpAddrs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *trustResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new trustResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := new.DirectoryID.ValueString()
	trustID := new.ID.ValueString()

	if !new.SelectiveAuth.IsUnknown() && !old.SelectiveAuth.Equal(new.SelectiveAuth) {
		input := &directoryservice.UpdateTrustInput{
			SelectiveAuth: new.SelectiveAuth.ValueEnum(),
			TrustId:       aws.String(trustID),
		}

		_, err := conn.UpdateTrust(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Directory Service Trust (%s)", trustID), err.Error())

			return
		}

		const (
			timeout = 10 * time.Minute
		)
		trust, err := waitTrustUpdated(ctx, conn, directoryID, trustID, timeout)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Directory Service Trust (%s) update", trustID), err.Error())

			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, trust, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		// Set values for unknowns.
		new.LastUpdatedDateTime = old.LastUpdatedDateTime
		new.SelectiveAuth = old.SelectiveAuth
		new.StateLastUpdatedDateTime = old.StateLastUpdatedDateTime
		new.TrustState = old.TrustState
		new.TrustStateReason = old.TrustStateReason
	}

	if !new.ConditionalForwarderIPAddrs.IsUnknown() && !old.ConditionalForwarderIPAddrs.Equal(new.ConditionalForwarderIPAddrs) {
		input := &directoryservice.UpdateConditionalForwarderInput{
			DirectoryId:      aws.String(directoryID),
			DnsIpAddrs:       fwflex.ExpandFrameworkStringValueSet(ctx, new.ConditionalForwarderIPAddrs),
			RemoteDomainName: fwflex.StringFromFramework(ctx, new.RemoteDomainName),
		}

		_, err := conn.UpdateConditionalForwarder(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Directory Service Conditional Forwarder (%s)", conditionalForwarderCreateResourceID(new.DirectoryID.ValueString(), new.RemoteDomainName.ValueString())), err.Error())

			return
		}

		// Directory Trust optionally accepts a remote domain name with a trailing period.
		domainName := strings.TrimRight(new.RemoteDomainName.ValueString(), ".")
		forwarder, err := findConditionalForwarderByTwoPartKey(ctx, conn, directoryID, domainName)

		if err != nil {
			// Outputting a NotFoundError does not include the original error.
			// Retrieve it to give the user an actionalble error message.
			if nfe, ok := errs.As[*retry.NotFoundError](err); ok {
				if nfe.LastError != nil {
					err = nfe.LastError
				}
			}

			response.Diagnostics.AddError(fmt.Sprintf("reading Directory Service Conditional Forwarder (%s)", conditionalForwarderCreateResourceID(directoryID, domainName)), err.Error())

			return
		}

		new.ConditionalForwarderIPAddrs = fwflex.FlattenFrameworkStringValueSet(ctx, forwarder.DnsIpAddrs)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *trustResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data trustResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	_, err := conn.DeleteTrust(ctx, &directoryservice.DeleteTrustInput{
		DeleteAssociatedConditionalForwarder: data.DeleteAssociatedConditionalForwarder.ValueBool(),
		TrustId:                              fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Directory Service Trust (%s)", data.ID.ValueString()), err.Error())

		return
	}

	const (
		timeout = 5 * time.Minute
	)
	if _, err := waitTrustDeleted(ctx, conn, data.DirectoryID.ValueString(), data.ID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Directory Service Trust (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *trustResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, "/")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("Wrong format for import ID (%s), use: 'directory-id/remote-directory-domain'", request.ID))
		return
	}
	directoryID := parts[0]
	domain := parts[1]

	trust, err := findTrustByDomain(ctx, r.Meta().DSClient(ctx), directoryID, domain)
	if err != nil {
		response.Diagnostics.AddError(
			"Importing Resource",
			err.Error(),
		)
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), aws.ToString(trust.TrustId))...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("directory_id"), directoryID)...)
}

type trustResourceModel struct {
	ConditionalForwarderIPAddrs          types.Set                                   `tfsdk:"conditional_forwarder_ip_addrs"`
	CreatedDateTime                      timetypes.RFC3339                           `tfsdk:"created_date_time"`
	DeleteAssociatedConditionalForwarder types.Bool                                  `tfsdk:"delete_associated_conditional_forwarder"`
	DirectoryID                          types.String                                `tfsdk:"directory_id"`
	ID                                   types.String                                `tfsdk:"id"`
	LastUpdatedDateTime                  timetypes.RFC3339                           `tfsdk:"last_updated_date_time"`
	RemoteDomainName                     types.String                                `tfsdk:"remote_domain_name"`
	SelectiveAuth                        fwtypes.StringEnum[awstypes.SelectiveAuth]  `tfsdk:"selective_auth"`
	StateLastUpdatedDateTime             timetypes.RFC3339                           `tfsdk:"state_last_updated_date_time"`
	TrustDirection                       fwtypes.StringEnum[awstypes.TrustDirection] `tfsdk:"trust_direction"`
	TrustPassword                        types.String                                `tfsdk:"trust_password"`
	TrustState                           types.String                                `tfsdk:"trust_state"`
	TrustStateReason                     types.String                                `tfsdk:"trust_state_reason"`
	TrustType                            fwtypes.StringEnum[awstypes.TrustType]      `tfsdk:"trust_type"`
}

func findTrust(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeTrustsInput, filter tfslices.Predicate[*awstypes.Trust]) (*awstypes.Trust, error) {
	output, err := findTrusts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrusts(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeTrustsInput, filter tfslices.Predicate[*awstypes.Trust]) ([]awstypes.Trust, error) {
	var output []awstypes.Trust

	pages := directoryservice.NewDescribeTrustsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Trusts {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findTrustByTwoPartKey(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string) (*awstypes.Trust, error) {
	input := &directoryservice.DescribeTrustsInput{
		DirectoryId: aws.String(directoryID),
		TrustIds:    []string{trustID},
	}

	return findTrust(ctx, conn, input, tfslices.PredicateTrue[*awstypes.Trust]())
}

func findTrustByDomain(ctx context.Context, conn *directoryservice.Client, directoryID, domain string) (*awstypes.Trust, error) {
	input := &directoryservice.DescribeTrustsInput{
		DirectoryId: aws.String(directoryID),
	}

	return findTrust(ctx, conn, input, func(v *awstypes.Trust) bool {
		return aws.ToString(v.RemoteDomainName) == domain
	})
}

func statusTrust(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTrustByTwoPartKey(ctx, conn, directoryID, trustID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TrustState), nil
	}
}

func waitTrustCreated(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TrustStateCreating),
		Target:  enum.Slice(awstypes.TrustStateCreated),
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

func waitTrustVerified(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TrustStateCreating,
			awstypes.TrustStateCreated,
			awstypes.TrustStateVerifying,
		),
		// On first side of a Two-Way Trust relationship, `VerifyFailed` is expected. This then gets updated when the second side is created.
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

func waitTrustUpdated(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
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

func waitTrustDeleted(ctx context.Context, conn *directoryservice.Client, directoryID, trustID string, timeout time.Duration) (*awstypes.Trust, error) {
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
