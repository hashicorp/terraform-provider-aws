// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_vpc_encryption_control", name="VPC Encryption Control")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.VpcEncryptionControl")
// @Testing(hasNoPreExistingResource=true)
func newResourceVPCEncryptionControl(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceVPCEncryptionControl{}

	return r, nil
}

const (
	ResNameVPCEncryptionControl = "VPC Encryption Control"
)

type resourceVPCEncryptionControl struct {
	framework.ResourceWithModel[resourceVPCEncryptionControlModel]
	// framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *resourceVPCEncryptionControl) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VpcEncryptionControlMode](),
				Required:   true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"state_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Required: true,
			},
		},
		// Blocks: map[string]schema.Block{
		// },
	}
}

func (r *resourceVPCEncryptionControl) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceVPCEncryptionControlModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.CreateVpcEncryptionControlInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateVpcEncryptionControl(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "vpc_id", plan.VPCID.String())
		return
	}
	if out == nil || out.VpcEncryptionControl == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), "vpc_id", plan.VPCID.String())
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.VpcEncryptionControl.VpcEncryptionControlId)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID))
	if resp.Diagnostics.HasError() {
		return
	}

	// createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	createTimeout := 20 * time.Minute
	ec, err := waitVPCEncryptionControlAvailable(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "vpc_id", plan.VPCID.String())
		return
	}

	if plan.Mode.ValueEnum() == awstypes.VpcEncryptionControlModeEnforce {
		var modifyInput ec2.ModifyVpcEncryptionControlInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &modifyInput, flex.WithFieldNamePrefix("VpcEncryptionControl")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.ModifyVpcEncryptionControl(ctx, &modifyInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, "vpc_id", plan.VPCID.String())
			return
		}
		if out == nil || out.VpcEncryptionControl == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), "vpc_id", plan.VPCID.String())
			return
		}

		ec, err = waitVPCEncryptionControlAvailable(ctx, conn, plan.ID.ValueString(), createTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, "vpc_id", plan.VPCID.String())
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, ec, &plan, flex.WithFieldNamePrefix("VpcEncryptionControl")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceVPCEncryptionControl) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCEncryptionControlModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findVPCEncryptionControlByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceVPCEncryptionControl) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceVPCEncryptionControl) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceVPCEncryptionControlModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DeleteVpcEncryptionControlInput{
		VpcEncryptionControlId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteVpcEncryptionControl(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeInvalidVpcEncryptionControlIdNotFound) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	deleteTimeout := 5 * time.Minute
	_, err = waitVPCEncryptionControlDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitVPCEncryptionControlAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcEncryptionControl, error) {
	stateConf := &retry.StateChangeConfOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState]{
		Pending: enum.EnumSlice(awstypes.VpcEncryptionControlStateCreating, awstypes.VpcEncryptionControlStateEnforceInProgress, awstypes.VpcEncryptionControlStateMonitorInProgress),
		Target:  enum.EnumSlice(awstypes.VpcEncryptionControlStateAvailable),
		Refresh: statusVPCEncryptionControl(conn, id),
		Timeout: timeout,
	}

	return wrapError(stateConf.WaitForStateContext(ctx))
}

func waitVPCEncryptionControlDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcEncryptionControl, error) {
	stateConf := &retry.StateChangeConfOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState]{
		Pending: enum.EnumSlice(awstypes.VpcEncryptionControlStateDeleting),
		Target:  []awstypes.VpcEncryptionControlState{},
		Refresh: statusVPCEncryptionControl(conn, id),
		Timeout: timeout,
	}

	return wrapError(stateConf.WaitForStateContext(ctx))
}

func wrapError[T any](v T, err error) (T, error) {
	return v, smarterr.NewError(err)
}

func statusVPCEncryptionControl(conn *ec2.Client, id string) retry.StateRefreshFuncOf[*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState] {
	return func(ctx context.Context) (*awstypes.VpcEncryptionControl, awstypes.VpcEncryptionControlState, error) {
		encryptionControl, err := findVPCEncryptionControlByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return &encryptionControl, encryptionControl.State, nil
	}
}

func findVPCEncryptionControlByID(ctx context.Context, conn *ec2.Client, id string) (awstypes.VpcEncryptionControl, error) {
	output, err := findVPCEncryptionControlsByIDs(ctx, conn, []string{id})
	if err != nil {
		return awstypes.VpcEncryptionControl{}, err
	}

	result, err := tfresource.AssertSingleValueResult(output)

	return *result, err
}

func findVPCEncryptionControlsByIDs(ctx context.Context, conn *ec2.Client, ids []string) ([]awstypes.VpcEncryptionControl, error) {
	input := ec2.DescribeVpcEncryptionControlsInput{
		VpcEncryptionControlIds: ids,
	}

	var output []awstypes.VpcEncryptionControl

	pages := NewDescribeVpcEncryptionControlsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if tfawserr.ErrCodeEquals(err, errCodeInvalidVpcEncryptionControlIdNotFound) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.VpcEncryptionControls...)
	}

	return output, nil
}

type resourceVPCEncryptionControlModel struct {
	framework.WithRegionModel
	ID           types.String                                          `tfsdk:"id"`
	Mode         fwtypes.StringEnum[awstypes.VpcEncryptionControlMode] `tfsdk:"mode"`
	State        types.String                                          `tfsdk:"state"`
	StateMessage types.String                                          `tfsdk:"state_message"`
	VPCID        types.String                                          `tfsdk:"vpc_id"`
}

// DescribeVpcEncryptionControlsPaginatorOptions is the paginator options for
// DescribeVpcEncryptionControls
type DescribeVpcEncryptionControlsPaginatorOptions struct {
	// The maximum number of items to return for this request. To get the next page of
	// items, make another request with the token returned in the output. For more
	// information, see [Pagination].
	//
	// [Pagination]: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Query-Requests.html#api-pagination
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// DescribeVpcEncryptionControlsPaginator is a paginator for DescribeVpcEncryptionControls
type DescribeVpcEncryptionControlsPaginator struct {
	options   DescribeVpcEncryptionControlsPaginatorOptions
	client    DescribeVpcEncryptionControlsAPIClient
	params    *ec2.DescribeVpcEncryptionControlsInput
	nextToken *string
	firstPage bool
}

// NewDescribeVpcEncryptionControlsPaginator returns a new DescribeVpcEncryptionControlsPaginator
func NewDescribeVpcEncryptionControlsPaginator(client DescribeVpcEncryptionControlsAPIClient, params *ec2.DescribeVpcEncryptionControlsInput, optFns ...func(*DescribeVpcEncryptionControlsPaginatorOptions)) *DescribeVpcEncryptionControlsPaginator {
	if params == nil {
		params = &ec2.DescribeVpcEncryptionControlsInput{}
	}

	options := DescribeVpcEncryptionControlsPaginatorOptions{}
	if params.MaxResults != nil {
		options.Limit = *params.MaxResults
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &DescribeVpcEncryptionControlsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *DescribeVpcEncryptionControlsPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next DescribeVpcEncryptionControls page.
func (p *DescribeVpcEncryptionControlsPaginator) NextPage(ctx context.Context, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcEncryptionControlsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxResults = limit

	result, err := p.client.DescribeVpcEncryptionControls(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

// DescribeVpcEncryptionControlsAPIClient is a client that implements the
// DescribeVpcEncryptionControls operation.
type DescribeVpcEncryptionControlsAPIClient interface {
	DescribeVpcEncryptionControls(context.Context, *ec2.DescribeVpcEncryptionControlsInput, ...func(*ec2.Options)) (*ec2.DescribeVpcEncryptionControlsOutput, error)
}

var _ DescribeVpcEncryptionControlsAPIClient = (*ec2.Client)(nil)
