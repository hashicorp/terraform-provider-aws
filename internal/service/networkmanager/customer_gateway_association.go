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

// @SDKResource("aws_networkmanager_customer_gateway_association", name="Customer Gateway Association")
func resourceCustomerGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomerGatewayAssociationCreate,
		ReadWithoutTimeout:   resourceCustomerGatewayAssociationRead,
		DeleteWithoutTimeout: resourceCustomerGatewayAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"customer_gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
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
		},
	}
}

func resourceCustomerGatewayAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	customerGatewayARN := d.Get("customer_gateway_arn").(string)
	id := customerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN)
	input := &networkmanager.AssociateCustomerGatewayInput{
		CustomerGatewayArn: aws.String(customerGatewayARN),
		DeviceId:           aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:    aws.String(globalNetworkID),
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Customer Gateway Association: %#v", input)
	_, err := tfresource.RetryWhen(ctx, customerGatewayAssociationResourceNotFoundExceptionTimeout,
		func() (any, error) {
			return conn.AssociateCustomerGateway(ctx, input)
		},
		func(err error) (bool, error) {
			// Wait out eventual consistency errors like:
			//
			// ResourceNotFoundException: Resource not found.
			// {
			//   RespMetadata: {
			// 	  StatusCode: 404,
			// 	  RequestID: "530d124c-2af8-4adf-be73-cee3793042f3"
			//   },
			//   Message_: "Resource not found.",
			//   ResourceId: "arn:aws:ec2:us-west-2:123456789012:customer-gateway/cgw-07c83f17516ae28fd",
			//   ResourceType: "customer-gateway"
			// }
			if resourceNotFoundExceptionResourceIDEquals(err, customerGatewayARN) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Customer Gateway Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitCustomerGatewayAssociationCreated(ctx, conn, globalNetworkID, customerGatewayARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Customer Gateway Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCustomerGatewayAssociationRead(ctx, d, meta)...)
}

func resourceCustomerGatewayAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, customerGatewayARN, err := customerGatewayAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Customer Gateway Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Customer Gateway Association (%s): %s", d.Id(), err)
	}

	d.Set("customer_gateway_arn", output.CustomerGatewayArn)
	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)

	return diags
}

func resourceCustomerGatewayAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, customerGatewayARN, err := customerGatewayAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = disassociateCustomerGateway(ctx, conn, globalNetworkID, customerGatewayARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func disassociateCustomerGateway(ctx context.Context, conn *networkmanager.Client, globalNetworkID, customerGatewayARN string, timeout time.Duration) error {
	id := customerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN)

	log.Printf("[DEBUG] Deleting Network Manager Customer Gateway Association: %s", id)
	_, err := conn.DisassociateCustomerGateway(ctx, &networkmanager.DisassociateCustomerGatewayInput{
		CustomerGatewayArn: aws.String(customerGatewayARN),
		GlobalNetworkId:    aws.String(globalNetworkID),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Network Manager Customer Gateway Association (%s): %w", id, err)
	}

	if _, err := waitCustomerGatewayAssociationDeleted(ctx, conn, globalNetworkID, customerGatewayARN, timeout); err != nil {
		return fmt.Errorf("waiting for Network Manager Customer Gateway Association (%s) delete: %w", id, err)
	}

	return nil
}

func findCustomerGatewayAssociation(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetCustomerGatewayAssociationsInput) (*awstypes.CustomerGatewayAssociation, error) {
	output, err := findCustomerGatewayAssociations(ctx, conn, input)

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

func findCustomerGatewayAssociations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetCustomerGatewayAssociationsInput) ([]awstypes.CustomerGatewayAssociation, error) {
	var output []awstypes.CustomerGatewayAssociation

	pages := networkmanager.NewGetCustomerGatewayAssociationsPaginator(conn, input)
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

		output = append(output, page.CustomerGatewayAssociations...)
	}

	return output, nil
}

func findCustomerGatewayAssociationByTwoPartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, customerGatewayARN string) (*awstypes.CustomerGatewayAssociation, error) {
	input := &networkmanager.GetCustomerGatewayAssociationsInput{
		CustomerGatewayArns: []string{customerGatewayARN},
		GlobalNetworkId:     aws.String(globalNetworkID),
	}

	output, err := findCustomerGatewayAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.CustomerGatewayAssociationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.CustomerGatewayArn) != customerGatewayARN {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusCustomerGatewayAssociationState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, customerGatewayARN string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitCustomerGatewayAssociationCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, customerGatewayARN string, timeout time.Duration) (*awstypes.CustomerGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CustomerGatewayAssociationStatePending),
		Target:  enum.Slice(awstypes.CustomerGatewayAssociationStateAvailable),
		Timeout: timeout,
		Refresh: statusCustomerGatewayAssociationState(ctx, conn, globalNetworkID, customerGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CustomerGatewayAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitCustomerGatewayAssociationDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, customerGatewayARN string, timeout time.Duration) (*awstypes.CustomerGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CustomerGatewayAssociationStateAvailable, awstypes.CustomerGatewayAssociationStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusCustomerGatewayAssociationState(ctx, conn, globalNetworkID, customerGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CustomerGatewayAssociation); ok {
		return output, err
	}

	return nil, err
}

const customerGatewayAssociationIDSeparator = ","

func customerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN string) string {
	parts := []string{globalNetworkID, customerGatewayARN}
	id := strings.Join(parts, customerGatewayAssociationIDSeparator)

	return id
}

func customerGatewayAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, customerGatewayAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sCUSTOMER-GATEWAY-ARN", id, customerGatewayAssociationIDSeparator)
}

const (
	customerGatewayAssociationResourceNotFoundExceptionTimeout = 1 * time.Minute
)
