package networkmanager

import (
	"context"
	"errors"
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

func ResourceTransitGatewayRegistration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRegistrationCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRegistrationRead,
		DeleteWithoutTimeout: resourceTransitGatewayRegistrationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceTransitGatewayRegistrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID := d.Get("global_network_id").(string)
	transitGatewayARN := d.Get("transit_gateway_arn").(string)
	id := TransitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN)
	input := &networkmanager.RegisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(globalNetworkID),
		TransitGatewayArn: aws.String(transitGatewayARN),
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Registration: %s", input)
	_, err := conn.RegisterTransitGatewayWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating Network Manager Transit Gateway Registration (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayRegistrationCreated(ctx, conn, globalNetworkID, transitGatewayARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Network Manager Transit Gateway Attachment (%s) create: %s", d.Id(), err)
	}

	return resourceTransitGatewayRegistrationRead(ctx, d, meta)
}

func resourceTransitGatewayRegistrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID, transitGatewayARN, err := TransitGatewayRegistrationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	transitGatewayRegistration, err := FindTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Registration %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Network Manager Transit Gateway Registration (%s): %s", d.Id(), err)
	}

	d.Set("global_network_id", transitGatewayRegistration.GlobalNetworkId)
	d.Set("transit_gateway_arn", transitGatewayRegistration.TransitGatewayArn)

	return nil
}

func resourceTransitGatewayRegistrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()

	globalNetworkID, transitGatewayARN, err := TransitGatewayRegistrationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = deregisterTransitGateway(ctx, conn, globalNetworkID, transitGatewayARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func deregisterTransitGateway(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayARN string, timeout time.Duration) error {
	id := TransitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Registration: %s", id)
	_, err := conn.DeregisterTransitGatewayWithContext(ctx, &networkmanager.DeregisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(globalNetworkID),
		TransitGatewayArn: aws.String(transitGatewayARN),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Transit Gateway Registration (%s): %w", id, err)
	}

	if _, err := waitTransitGatewayRegistrationDeleted(ctx, conn, globalNetworkID, transitGatewayARN, timeout); err != nil {
		return fmt.Errorf("error waiting for Network Manager Transit Gateway Registration (%s) delete: %w", id, err)
	}

	return nil
}

func FindTransitGatewayRegistration(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetTransitGatewayRegistrationsInput) (*networkmanager.TransitGatewayRegistration, error) {
	output, err := FindTransitGatewayRegistrations(ctx, conn, input)

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

func FindTransitGatewayRegistrations(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetTransitGatewayRegistrationsInput) ([]*networkmanager.TransitGatewayRegistration, error) {
	var output []*networkmanager.TransitGatewayRegistration

	err := conn.GetTransitGatewayRegistrationsPagesWithContext(ctx, input, func(page *networkmanager.GetTransitGatewayRegistrationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayRegistrations {
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

func FindTransitGatewayRegistrationByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayARN string) (*networkmanager.TransitGatewayRegistration, error) {
	input := &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId:    aws.String(globalNetworkID),
		TransitGatewayArns: aws.StringSlice([]string{transitGatewayARN}),
	}

	output, err := FindTransitGatewayRegistration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State.Code); state == networkmanager.TransitGatewayRegistrationStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.TransitGatewayArn) != transitGatewayARN {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusTransitGatewayRegistrationState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State.Code), nil
	}
}

func waitTransitGatewayRegistrationCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayARN string, timeout time.Duration) (*networkmanager.TransitGatewayRegistration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.TransitGatewayRegistrationStatePending},
		Target:  []string{networkmanager.TransitGatewayRegistrationStateAvailable},
		Timeout: timeout,
		Refresh: statusTransitGatewayRegistrationState(ctx, conn, globalNetworkID, transitGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayRegistration); ok {
		if state := aws.StringValue(output.State.Code); state == networkmanager.TransitGatewayRegistrationStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.State.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRegistrationDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayARN string, timeout time.Duration) (*networkmanager.TransitGatewayRegistration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.TransitGatewayRegistrationStateAvailable, networkmanager.TransitGatewayRegistrationStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusTransitGatewayRegistrationState(ctx, conn, globalNetworkID, transitGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.TransitGatewayRegistration); ok {
		if state := aws.StringValue(output.State.Code); state == networkmanager.TransitGatewayRegistrationStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.State.Message)))
		}

		return output, err
	}

	return nil, err
}

const transitGatewayRegistrationIDSeparator = ","

func TransitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN string) string {
	parts := []string{globalNetworkID, transitGatewayARN}
	id := strings.Join(parts, transitGatewayRegistrationIDSeparator)

	return id
}

func TransitGatewayRegistrationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRegistrationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sTRANSIT-GATEWAY-ARN", id, transitGatewayRegistrationIDSeparator)
}
