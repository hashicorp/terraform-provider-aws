// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"errors"
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

// @SDKResource("aws_networkmanager_transit_gateway_registration", name="Transit Gateway Registration")
func resourceTransitGatewayRegistration() *schema.Resource {
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

func resourceTransitGatewayRegistrationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	transitGatewayARN := d.Get("transit_gateway_arn").(string)
	id := transitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN)
	input := &networkmanager.RegisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(globalNetworkID),
		TransitGatewayArn: aws.String(transitGatewayARN),
	}

	log.Printf("[DEBUG] Creating Network Manager Transit Gateway Registration: %#v", input)
	_, err := conn.RegisterTransitGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Transit Gateway Registration (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTransitGatewayRegistrationCreated(ctx, conn, globalNetworkID, transitGatewayARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayRegistrationRead(ctx, d, meta)...)
}

func resourceTransitGatewayRegistrationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, transitGatewayARN, err := transitGatewayRegistrationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	transitGatewayRegistration, err := findTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Transit Gateway Registration %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Registration (%s): %s", d.Id(), err)
	}

	d.Set("global_network_id", transitGatewayRegistration.GlobalNetworkId)
	d.Set("transit_gateway_arn", transitGatewayRegistration.TransitGatewayArn)

	return diags
}

func resourceTransitGatewayRegistrationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, transitGatewayARN, err := transitGatewayRegistrationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = deregisterTransitGateway(ctx, conn, globalNetworkID, transitGatewayARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deregisterTransitGateway(ctx context.Context, conn *networkmanager.Client, globalNetworkID, transitGatewayARN string, timeout time.Duration) error {
	id := transitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN)

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Registration: %s", id)
	_, err := conn.DeregisterTransitGateway(ctx, &networkmanager.DeregisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(globalNetworkID),
		TransitGatewayArn: aws.String(transitGatewayARN),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Network Manager Transit Gateway Registration (%s): %w", id, err)
	}

	if _, err := waitTransitGatewayRegistrationDeleted(ctx, conn, globalNetworkID, transitGatewayARN, timeout); err != nil {
		return fmt.Errorf("waiting for Network Manager Transit Gateway Registration (%s) delete: %w", id, err)
	}

	return nil
}

func findTransitGatewayRegistration(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetTransitGatewayRegistrationsInput) (*awstypes.TransitGatewayRegistration, error) {
	output, err := findTransitGatewayRegistrations(ctx, conn, input)

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

func findTransitGatewayRegistrations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetTransitGatewayRegistrationsInput) ([]awstypes.TransitGatewayRegistration, error) {
	var output []awstypes.TransitGatewayRegistration

	pages := networkmanager.NewGetTransitGatewayRegistrationsPaginator(conn, input)
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

		output = append(output, page.TransitGatewayRegistrations...)
	}

	return output, nil
}

func findTransitGatewayRegistrationByTwoPartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, transitGatewayARN string) (*awstypes.TransitGatewayRegistration, error) {
	input := &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId:    aws.String(globalNetworkID),
		TransitGatewayArns: []string{transitGatewayARN},
	}

	output, err := findTransitGatewayRegistration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State.Code; state == awstypes.TransitGatewayRegistrationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.TransitGatewayArn) != transitGatewayARN {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusTransitGatewayRegistrationState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, transitGatewayARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTransitGatewayRegistrationByTwoPartKey(ctx, conn, globalNetworkID, transitGatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Code), nil
	}
}

func waitTransitGatewayRegistrationCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, transitGatewayARN string, timeout time.Duration) (*awstypes.TransitGatewayRegistration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRegistrationStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayRegistrationStateAvailable),
		Timeout: timeout,
		Refresh: statusTransitGatewayRegistrationState(ctx, conn, globalNetworkID, transitGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRegistration); ok {
		if state := output.State.Code; state == awstypes.TransitGatewayRegistrationStateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.State.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTransitGatewayRegistrationDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, transitGatewayARN string, timeout time.Duration) (*awstypes.TransitGatewayRegistration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRegistrationStateAvailable, awstypes.TransitGatewayRegistrationStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusTransitGatewayRegistrationState(ctx, conn, globalNetworkID, transitGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRegistration); ok {
		if state := output.State.Code; state == awstypes.TransitGatewayRegistrationStateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.State.Message)))
		}

		return output, err
	}

	return nil, err
}

const transitGatewayRegistrationIDSeparator = ","

func transitGatewayRegistrationCreateResourceID(globalNetworkID, transitGatewayARN string) string {
	parts := []string{globalNetworkID, transitGatewayARN}
	id := strings.Join(parts, transitGatewayRegistrationIDSeparator)

	return id
}

func transitGatewayRegistrationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRegistrationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sTRANSIT-GATEWAY-ARN", id, transitGatewayRegistrationIDSeparator)
}
