// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
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

// Each resource is one primary->secondary pair. Multiple secondaries, if the service allows
// them, are modelled as multiple resources.
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
			"peer_role": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRole: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusMessage: schema.StringAttribute{
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

	domainName, peerDomainARN := fwflex.StringValueFromFramework(ctx, plan.DomainName), fwflex.StringValueFromFramework(ctx, plan.PeerDomainARN)
	id := domainReplicationCreateResourceID(domainName, peerDomainARN)
	input := opensearch.AddReplicationPeerInput{
		DomainName:    aws.String(domainName),
		PeerDomainArn: aws.String(peerDomainARN),
	}

	// ResourceAlreadyExistsException: a domain is already replicating.
	// ValidationException: non-empty secondary, cross-account peer, or malformed ARN; the
	// service message names the cause, so it is passed through unchanged.
	if _, err := conn.AddReplicationPeer(ctx, &input); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Domain Replication (%s)", id), err.Error())
		return
	}

	out, err := waitDomainReplicationActive(ctx, conn, domainName, peerDomainARN, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Domain Replication (%s) create", id), err.Error())
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, id)
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

	domainName, peerDomainARN, err := domainReplicationParseResourceID(fwflex.StringValueFromFramework(ctx, state.ID))
	if err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findDomainReplicationByTwoPartKey(ctx, conn, domainName, peerDomainARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Domain Replication (%s)", state.ID.ValueString()), err.Error())
		return
	}

	// Set the identifying arguments for import.
	state.DomainName = fwflex.StringValueToFramework(ctx, domainName)
	state.PeerDomainARN = fwtypes.ARNValue(peerDomainARN)
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

	domainName, peerDomainARN := fwflex.StringValueFromFramework(ctx, state.DomainName), fwflex.StringValueFromFramework(ctx, state.PeerDomainARN)
	input := opensearch.RemoveReplicationPeerInput{
		DomainName:    aws.String(domainName),
		PeerDomainArn: aws.String(peerDomainARN),
	}
	_, err := conn.RemoveReplicationPeer(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Domain Replication (%s)", state.ID.ValueString()), err.Error())
		return
	}

	if _, err := waitDomainReplicationRemoved(ctx, conn, domainName, peerDomainARN, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for OpenSearch Domain Replication (%s) delete", state.ID.ValueString()), err.Error())
		return
	}
}

// ACTIVE means the relationship is established on both clusters; the first copy may still be
// running. A FAILED peer is terminal, so the waiter stops instead of polling to the timeout.
func waitDomainReplicationActive(ctx context.Context, conn *opensearch.Client, domainName, peerDomainARN string, timeout time.Duration) (*opensearch.DescribeDomainReplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicationStatusPendingSetup),
		Target:  enum.Slice(awstypes.ReplicationStatusActive),
		Refresh: statusDomainReplication(conn, domainName, peerDomainARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.DescribeDomainReplicationOutput); ok {
		return out, err
	}

	return nil, err
}

// The pair is gone once describe reports not-replicating, the peer is no longer listed, or the
// peer is DETACHED.
func waitDomainReplicationRemoved(ctx context.Context, conn *opensearch.Client, domainName, peerDomainARN string, timeout time.Duration) (*opensearch.DescribeDomainReplicationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicationStatusPendingRemoval, awstypes.ReplicationStatusActive, awstypes.ReplicationStatusEjected),
		Target:  []string{},
		Refresh: statusDomainReplication(conn, domainName, peerDomainARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*opensearch.DescribeDomainReplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDomainReplication(conn *opensearch.Client, domainName, peerDomainARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findDomainReplicationByTwoPartKey(ctx, conn, domainName, peerDomainARN)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		peer := findReplicationPeerByARN(out, peerDomainARN)
		switch peer.Status {
		case awstypes.ReplicationStatusDetached:
			return nil, "", nil
		case awstypes.ReplicationStatusFailed:
			return out, string(peer.Status), fmt.Errorf("replication peer status %s: %s", peer.Status, aws.ToString(peer.StatusMessage))
		}

		return out, string(peer.Status), nil
	}
}

// The returned output is guaranteed to list peerDomainARN among its peers. A domain that is not
// replicating returns ResourceNotFoundException; an empty peer list, or one without this peer, is
// treated the same way. EJECTED peers (after a failover) are still a relationship.
func findDomainReplicationByTwoPartKey(ctx context.Context, conn *opensearch.Client, domainName, peerDomainARN string) (*opensearch.DescribeDomainReplicationOutput, error) {
	input := opensearch.DescribeDomainReplicationInput{
		DomainName: aws.String(domainName),
	}

	output, err := findDomainReplication(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if findReplicationPeerByARN(output, peerDomainARN) == nil {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("OpenSearch Domain Replication: domain %s has no replication peer %s", domainName, peerDomainARN),
		}
	}

	return output, nil
}

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

// findReplicationPeerByARN returns the entry for peerDomainARN in the describe output, or nil.
func findReplicationPeerByARN(out *opensearch.DescribeDomainReplicationOutput, peerDomainARN string) *awstypes.ReplicationPeer {
	if out == nil {
		return nil
	}

	for i := range out.Peers {
		if aws.ToString(out.Peers[i].PeerDomainArn) == peerDomainARN {
			return &out.Peers[i]
		}
	}

	return nil
}

// The pair's attributes come from two levels of the output (the domain and its entry for this
// peer), so this is set by hand rather than with fwflex.Flatten. out must come from
// findDomainReplicationByTwoPartKey for the model's peer_domain_arn.
func (data *domainReplicationResourceModel) setFromDescribe(ctx context.Context, out *opensearch.DescribeDomainReplicationOutput) {
	data.DomainARN = fwflex.StringToFramework(ctx, out.DomainArn)
	data.Role = fwflex.StringValueToFramework(ctx, string(out.Role))

	peer := findReplicationPeerByARN(out, fwflex.StringValueFromFramework(ctx, data.PeerDomainARN))
	data.PeerRole = fwflex.StringValueToFramework(ctx, string(peer.PeerRole))
	data.Status = fwflex.StringValueToFramework(ctx, string(peer.Status))
	data.StatusMessage = fwflex.StringToFramework(ctx, peer.StatusMessage)
}

type domainReplicationResourceModel struct {
	framework.WithRegionModel
	DomainARN     types.String   `tfsdk:"domain_arn"`
	DomainName    types.String   `tfsdk:"domain_name"`
	ID            types.String   `tfsdk:"id"`
	PeerDomainARN fwtypes.ARN    `tfsdk:"peer_domain_arn"`
	PeerRole      types.String   `tfsdk:"peer_role"`
	Role          types.String   `tfsdk:"role"`
	Status        types.String   `tfsdk:"status"`
	StatusMessage types.String   `tfsdk:"status_message"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

const domainReplicationResourceIDPartCount = 2

func domainReplicationCreateResourceID(domainName, peerDomainARN string) string {
	return strings.Join([]string{domainName, peerDomainARN}, intflex.ResourceIdSeparator)
}

func domainReplicationParseResourceID(id string) (string, string, error) {
	parts, err := intflex.ExpandResourceId(id, domainReplicationResourceIDPartCount, false)
	if err != nil {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected domain-name%[2]speer-domain-arn: %w", id, intflex.ResourceIdSeparator, err)
	}

	return parts[0], parts[1], nil
}
