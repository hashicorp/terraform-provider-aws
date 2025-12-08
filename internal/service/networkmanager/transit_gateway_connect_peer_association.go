// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_networkmanager_transit_gateway_connect_peer_association", name="Transit Gateway Connect Peer Association")
func resourceTransitGatewayConnectPeerAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayConnectPeerAssociationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayConnectPeerAssociationRead,
		DeleteWithoutTimeout: resourceTransitGatewayConnectPeerAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"transit_gateway_connect_peer_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceTransitGatewayConnectPeerAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	connectPeerARN := d.Get("transit_gateway_connect_peer_arn").(string)
	id := transitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN)
	input := &networkmanager.AssociateTransitGatewayConnectPeerInput{
		DeviceId:                     aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:              aws.String(globalNetworkID),
		TransitGatewayConnectPeerArn: aws.String(connectPeerARN),
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Connect Peer Association: %#v", input)
	_, err := conn.AssociateTransitGatewayConnectPeer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Transit Gateway Connect Peer Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayConnectPeerAssociationCreated(ctx, conn, globalNetworkID, connectPeerARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Connect Peer Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayConnectPeerAssociationRead(ctx, d, meta)...)
}

func resourceTransitGatewayConnectPeerAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, connectPeerARN, err := transitGatewayConnectPeerAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Connect Peer Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Connect Peer Association (%s): %s", d.Id(), err)
	}

	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)
	d.Set("transit_gateway_connect_peer_arn", output.TransitGatewayConnectPeerArn)

	return diags
}

func resourceTransitGatewayConnectPeerAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, connectPeerARN, err := transitGatewayConnectPeerAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = disassociateTransitGatewayConnectPeer(ctx, conn, globalNetworkID, connectPeerARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func disassociateTransitGatewayConnectPeer(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectPeerARN string, timeout time.Duration) error {
	id := transitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Connect Peer Association: %s", id)
	_, err := conn.DisassociateTransitGatewayConnectPeer(ctx, &networkmanager.DisassociateTransitGatewayConnectPeerInput{
		GlobalNetworkId:              aws.String(globalNetworkID),
		TransitGatewayConnectPeerArn: aws.String(connectPeerARN),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Network Manager Transit Gateway Connect Peer Association (%s): %w", id, err)
	}

	if _, err := waitTransitGatewayConnectPeerAssociationDeleted(ctx, conn, globalNetworkID, connectPeerARN, timeout); err != nil {
		return fmt.Errorf("waiting for Network Manager Transit Gateway Connect Peer Association (%s) delete: %w", id, err)
	}

	return nil
}

func findTransitGatewayConnectPeerAssociation(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) (*awstypes.TransitGatewayConnectPeerAssociation, error) {
	output, err := findTransitGatewayConnectPeerAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func findTransitGatewayConnectPeerAssociations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) ([]awstypes.TransitGatewayConnectPeerAssociation, error) {
	var output []awstypes.TransitGatewayConnectPeerAssociation

	pages := networkmanager.NewGetTransitGatewayConnectPeerAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if globalNetworkIDNotFoundError(err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayConnectPeerAssociations...)
	}

	return output, nil
}

func findTransitGatewayConnectPeerAssociationByTwoPartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectPeerARN string) (*awstypes.TransitGatewayConnectPeerAssociation, error) {
	input := &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
		GlobalNetworkId:               aws.String(globalNetworkID),
		TransitGatewayConnectPeerArns: []string{connectPeerARN},
	}

	output, err := findTransitGatewayConnectPeerAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayConnectPeerAssociationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.TransitGatewayConnectPeerArn) != connectPeerARN {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusTransitGatewayConnectPeerAssociationState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectPeerARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitTransitGatewayConnectPeerAssociationCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectPeerARN string, timeout time.Duration) (*awstypes.TransitGatewayConnectPeerAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayConnectPeerAssociationStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayConnectPeerAssociationStateAvailable),
		Timeout: timeout,
		Refresh: statusTransitGatewayConnectPeerAssociationState(ctx, conn, globalNetworkID, connectPeerARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnectPeerAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectPeerAssociationDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, connectPeerARN string, timeout time.Duration) (*awstypes.TransitGatewayConnectPeerAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayConnectPeerAssociationStateAvailable, awstypes.TransitGatewayConnectPeerAssociationStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusTransitGatewayConnectPeerAssociationState(ctx, conn, globalNetworkID, connectPeerARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnectPeerAssociation); ok {
		return output, err
	}

	return nil, err
}

const transitGatewayConnectPeerAssociationIDSeparator = ","

func transitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN string) string {
	parts := []string{globalNetworkID, connectPeerARN}
	id := strings.Join(parts, transitGatewayConnectPeerAssociationIDSeparator)

	return id
}

func transitGatewayConnectPeerAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayConnectPeerAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sCONNECT-PEER-ARN", id, transitGatewayConnectPeerAssociationIDSeparator)
}
