package networkmanager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayConnectPeerAssociation() *schema.Resource {
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

func resourceTransitGatewayConnectPeerAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID := d.Get("global_network_id").(string)
	connectPeerARN := d.Get("transit_gateway_connect_peer_arn").(string)
	id := TransitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN)
	input := &networkmanager.AssociateTransitGatewayConnectPeerInput{
		DeviceId:                     aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:              aws.String(globalNetworkID),
		TransitGatewayConnectPeerArn: aws.String(connectPeerARN),
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Connect Peer Association: %s", input)
	_, err := conn.AssociateTransitGatewayConnectPeerWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Network Manager Transit Gateway Connect Peer Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayConnectPeerAssociationCreated(ctx, conn, globalNetworkID, connectPeerARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Network Manager Transit Gateway Connect Peer Association (%s) create: %s", d.Id(), err)
	}

	return resourceTransitGatewayConnectPeerAssociationRead(ctx, d, meta)
}

func resourceTransitGatewayConnectPeerAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID, connectPeerARN, err := TransitGatewayConnectPeerAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Connect Peer Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Network Manager Transit Gateway Connect Peer Association (%s): %s", d.Id(), err)
	}

	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)
	d.Set("transit_gateway_connect_peer_arn", output.TransitGatewayConnectPeerArn)

	return nil
}

func resourceTransitGatewayConnectPeerAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID, connectPeerARN, err := TransitGatewayConnectPeerAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = disassociateTransitGatewayConnectPeer(ctx, conn, globalNetworkID, connectPeerARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func disassociateTransitGatewayConnectPeer(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectPeerARN string, timeout time.Duration) error {
	id := TransitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Connect Peer Association: %s", id)
	_, err := conn.DisassociateTransitGatewayConnectPeerWithContext(ctx, &networkmanager.DisassociateTransitGatewayConnectPeerInput{
		GlobalNetworkId:              aws.String(globalNetworkID),
		TransitGatewayConnectPeerArn: aws.String(connectPeerARN),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Transit Gateway Connect Peer Association (%s): %w", id, err)
	}

	if _, err := waitTransitGatewayConnectPeerAssociationDeleted(ctx, conn, globalNetworkID, connectPeerARN, timeout); err != nil {
		return fmt.Errorf("error waiting for Network Manager Transit Gateway Connect Peer Association (%s) delete: %w", id, err)
	}

	return nil
}

func FindTransitGatewayConnectPeerAssociation(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) (*networkmanager.TransitGatewayConnectPeerAssociation, error) {
	output, err := FindTransitGatewayConnectPeerAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].State == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayConnectPeerAssociations(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetTransitGatewayConnectPeerAssociationsInput) ([]*networkmanager.TransitGatewayConnectPeerAssociation, error) {
	var output []*networkmanager.TransitGatewayConnectPeerAssociation

	err := conn.GetTransitGatewayConnectPeerAssociationsPagesWithContext(ctx, input, func(page *networkmanager.GetTransitGatewayConnectPeerAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnectPeerAssociations {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if globalNetworkIDNotFoundError(err) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayConnectPeerAssociationByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectPeerARN string) (*networkmanager.TransitGatewayConnectPeerAssociation, error) {
	input := &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
		GlobalNetworkId:               aws.String(globalNetworkID),
		TransitGatewayConnectPeerArns: aws.StringSlice([]string{connectPeerARN}),
	}

	output, err := FindTransitGatewayConnectPeerAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == networkmanager.TransitGatewayConnectPeerAssociationStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.TransitGatewayConnectPeerArn) != connectPeerARN {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusTransitGatewayConnectPeerAssociationState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectPeerARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayConnectPeerAssociationByTwoPartKey(ctx, conn, globalNetworkID, connectPeerARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitTransitGatewayConnectPeerAssociationCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectPeerARN string, timeout time.Duration) (*networkmanager.TransitGatewayConnectPeerAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.TransitGatewayConnectPeerAssociationStatePending},
		Target:  []string{networkmanager.TransitGatewayConnectPeerAssociationStateAvailable},
		Timeout: timeout,
		Refresh: statusTransitGatewayConnectPeerAssociationState(ctx, conn, globalNetworkID, connectPeerARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayConnectPeerAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectPeerAssociationDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, connectPeerARN string, timeout time.Duration) (*networkmanager.TransitGatewayConnectPeerAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.TransitGatewayConnectPeerAssociationStateAvailable, networkmanager.TransitGatewayConnectPeerAssociationStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusTransitGatewayConnectPeerAssociationState(ctx, conn, globalNetworkID, connectPeerARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayConnectPeerAssociation); ok {
		return output, err
	}

	return nil, err
}

const transitGatewayConnectPeerAssociationIDSeparator = ","

func TransitGatewayConnectPeerAssociationCreateResourceID(globalNetworkID, connectPeerARN string) string {
	parts := []string{globalNetworkID, connectPeerARN}
	id := strings.Join(parts, transitGatewayConnectPeerAssociationIDSeparator)

	return id
}

func TransitGatewayConnectPeerAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayConnectPeerAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sCONNECT-PEER-ARN", id, transitGatewayConnectPeerAssociationIDSeparator)
}
