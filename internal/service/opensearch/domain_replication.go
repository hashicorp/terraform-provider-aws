// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearch_domain_replication", name="Domain Replication")
func newDomainReplicationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainReplicationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type domainReplicationResource struct {
	framework.ResourceWithModel[domainReplicationResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *domainReplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_arn": framework.ARNAttributeComputedOnly(),
			names.AttrDomainName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"peer_domain_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"peers": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[replicationPeerModel](ctx),
				ElementType: fwtypes.NewObjectTypeOf[replicationPeerModel](ctx),
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRole: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *domainReplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainReplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	domainName := fwflex.StringValueFromFramework(ctx, plan.DomainName)
	input := opensearch.AddReplicationPeerInput{
		DomainName:    aws.String(domainName),
		PeerDomainArn: fwflex.StringFromFramework(ctx, plan.PeerDomainARN),
	}

	// ResourceAlreadyExistsException: a domain is already replicating.
	// ValidationException: non-empty secondary, cross-account peer, or malformed ARN; the
	// service message names the cause, so it is passed through unchanged.
	if _, err := conn.AddReplicationPeer(ctx, &input); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Domain Replication (%s)", domainName), err.Error())
		return
	}

	out, err := waitDomainReplicationActive(ctx, conn, domainName, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Domain Replication (%s) create", domainName), err.Error())
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, domainName)
	plan.setFromDescribe(ctx, out)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *domainReplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	domainName := fwflex.StringValueFromFramework(ctx, state.ID)
	out, err := findDomainReplicationByDomainName(ctx, conn, domainName)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Domain Replication (%s)", domainName), err.Error())
		return
	}

	// Set the identifying arguments for import. The peer comes from the live relationship.
	state.DomainName = fwflex.StringValueToFramework(ctx, domainName)
	if state.PeerDomainARN.IsNull() || state.PeerDomainARN.IsUnknown() {
		state.PeerDomainARN = fwtypes.ARNValue(aws.ToString(out.Peers[0].PeerDomainArn))
	}

	state.setFromDescribe(ctx, out)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *domainReplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchClient(ctx)

	domainName := fwflex.StringValueFromFramework(ctx, state.ID)
	input := opensearch.RemoveReplicationPeerInput{
		DomainName:    aws.String(domainName),
		PeerDomainArn: fwflex.StringFromFramework(ctx, state.PeerDomainARN),
	}
	_, err := conn.RemoveReplicationPeer(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Domain Replication (%s)", domainName), err.Error())
		return
	}

	if _, err := waitDomainReplicationRemoved(ctx, conn, domainName, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Domain Replication (%s) delete", domainName), err.Error())
		return
	}
}

// ACTIVE means the relationship is established on both clusters; the first copy may still be
// running. A FAILED peer is terminal, so the waiter stops instead of polling to the timeout.
func waitDomainReplicationActive(ctx context.Context, conn *opensearch.Client, domainName string, timeout time.Duration) (*opensearch.DescribeDomainReplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicationStatusPendingSetup),
		Target:  enum.Slice(awstypes.ReplicationStatusActive),
		Refresh: statusDomainReplication(conn, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.DescribeDomainReplicationOutput); ok {
		return out, err
	}

	return nil, err
}

// The relationship is gone once describe reports not-replicating, or every peer is DETACHED.
func waitDomainReplicationRemoved(ctx context.Context, conn *opensearch.Client, domainName string, timeout time.Duration) (*opensearch.DescribeDomainReplicationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicationStatusPendingRemoval, awstypes.ReplicationStatusActive, awstypes.ReplicationStatusEjected),
		Target:  []string{},
		Refresh: statusDomainReplication(conn, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.DescribeDomainReplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDomainReplication(conn *opensearch.Client, domainName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findDomainReplicationByDomainName(ctx, conn, domainName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		status := relationshipStatus(out)
		switch status {
		case awstypes.ReplicationStatusDetached:
			return nil, "", nil
		case awstypes.ReplicationStatusFailed:
			return out, string(status), fmt.Errorf("replication peer status %s: %s", status, aws.ToString(out.Peers[0].StatusMessage))
		}

		return out, string(status), nil
	}
}

func findDomainReplicationByDomainName(ctx context.Context, conn *opensearch.Client, domainName string) (*opensearch.DescribeDomainReplicationOutput, error) {
	input := opensearch.DescribeDomainReplicationInput{
		DomainName: aws.String(domainName),
	}

	return findDomainReplication(ctx, conn, &input)
}

// A domain that is not replicating returns ResourceNotFoundException; an empty peer list is
// treated the same way. EJECTED peers (after a failover) are still a relationship.
func findDomainReplication(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribeDomainReplicationInput) (*opensearch.DescribeDomainReplicationOutput, error) {
	output, err := conn.DescribeDomainReplication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Peers) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// One peer per domain, so the relationship status is the single peer's status.
func relationshipStatus(out *opensearch.DescribeDomainReplicationOutput) awstypes.ReplicationStatus {
	if out == nil || len(out.Peers) == 0 {
		return ""
	}

	return out.Peers[0].Status
}

// The top-level status is derived from the single peer, so this is set by hand rather than
// with fwflex.Flatten.
func (data *domainReplicationResourceModel) setFromDescribe(ctx context.Context, out *opensearch.DescribeDomainReplicationOutput) {
	data.DomainARN = fwflex.StringToFramework(ctx, out.DomainArn)
	data.Role = fwflex.StringValueToFramework(ctx, string(out.Role))
	data.Status = fwflex.StringValueToFramework(ctx, string(relationshipStatus(out)))

	peers := make([]*replicationPeerModel, 0, len(out.Peers))
	for _, p := range out.Peers {
		peers = append(peers, &replicationPeerModel{
			PeerDomainARN: fwflex.StringToFramework(ctx, p.PeerDomainArn),
			PeerRole:      fwflex.StringValueToFramework(ctx, string(p.PeerRole)),
			Status:        fwflex.StringValueToFramework(ctx, string(p.Status)),
			StatusMessage: fwflex.StringToFramework(ctx, p.StatusMessage),
		})
	}
	data.Peers = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, peers)
}

type domainReplicationResourceModel struct {
	framework.WithRegionModel
	DomainARN     types.String                                          `tfsdk:"domain_arn"`
	DomainName    types.String                                          `tfsdk:"domain_name"`
	ID            types.String                                          `tfsdk:"id"`
	PeerDomainARN fwtypes.ARN                                           `tfsdk:"peer_domain_arn"`
	Peers         fwtypes.ListNestedObjectValueOf[replicationPeerModel] `tfsdk:"peers"`
	Role          types.String                                          `tfsdk:"role"`
	Status        types.String                                          `tfsdk:"status"`
	Timeouts      timeouts.Value                                        `tfsdk:"timeouts"`
}

type replicationPeerModel struct {
	PeerDomainARN types.String `tfsdk:"peer_domain_arn"`
	PeerRole      types.String `tfsdk:"peer_role"`
	Status        types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
}
