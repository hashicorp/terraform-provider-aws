// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_networkmanager_customer_gateway_association")
func ResourceCustomerGatewayAssociation() *schema.Resource {
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

func resourceCustomerGatewayAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	customerGatewayARN := d.Get("customer_gateway_arn").(string)
	id := CustomerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN)
	input := &networkmanager.AssociateCustomerGatewayInput{
		CustomerGatewayArn: aws.String(customerGatewayARN),
		DeviceId:           aws.String(d.Get("device_id").(string)),
		GlobalNetworkId:    aws.String(globalNetworkID),
	}

	if v, ok := d.GetOk("link_id"); ok {
		input.LinkId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Customer Gateway Association: %s", input)
	_, err := tfresource.RetryWhen(ctx, customerGatewayAssociationResourceNotFoundExceptionTimeout,
		func() (interface{}, error) {
			return conn.AssociateCustomerGatewayWithContext(ctx, input)
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
		return diag.Errorf("creating Network Manager Customer Gateway Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitCustomerGatewayAssociationCreated(ctx, conn, globalNetworkID, customerGatewayARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Network Manager Customer Gateway Association (%s) create: %s", d.Id(), err)
	}

	return resourceCustomerGatewayAssociationRead(ctx, d, meta)
}

func resourceCustomerGatewayAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID, customerGatewayARN, err := CustomerGatewayAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Customer Gateway Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Network Manager Customer Gateway Association (%s): %s", d.Id(), err)
	}

	d.Set("customer_gateway_arn", output.CustomerGatewayArn)
	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)

	return nil
}

func resourceCustomerGatewayAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID, customerGatewayARN, err := CustomerGatewayAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	err = disassociateCustomerGateway(ctx, conn, globalNetworkID, customerGatewayARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func disassociateCustomerGateway(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, customerGatewayARN string, timeout time.Duration) error {
	id := CustomerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN)

	log.Printf("[DEBUG] Deleting Network Manager Customer Gateway Association: %s", id)
	_, err := conn.DisassociateCustomerGatewayWithContext(ctx, &networkmanager.DisassociateCustomerGatewayInput{
		CustomerGatewayArn: aws.String(customerGatewayARN),
		GlobalNetworkId:    aws.String(globalNetworkID),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindCustomerGatewayAssociation(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetCustomerGatewayAssociationsInput) (*networkmanager.CustomerGatewayAssociation, error) {
	output, err := FindCustomerGatewayAssociations(ctx, conn, input)

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

func FindCustomerGatewayAssociations(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetCustomerGatewayAssociationsInput) ([]*networkmanager.CustomerGatewayAssociation, error) {
	var output []*networkmanager.CustomerGatewayAssociation

	err := conn.GetCustomerGatewayAssociationsPagesWithContext(ctx, input, func(page *networkmanager.GetCustomerGatewayAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CustomerGatewayAssociations {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if globalNetworkIDNotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCustomerGatewayAssociationByTwoPartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, customerGatewayARN string) (*networkmanager.CustomerGatewayAssociation, error) {
	input := &networkmanager.GetCustomerGatewayAssociationsInput{
		CustomerGatewayArns: aws.StringSlice([]string{customerGatewayARN}),
		GlobalNetworkId:     aws.String(globalNetworkID),
	}

	output, err := FindCustomerGatewayAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == networkmanager.CustomerGatewayAssociationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.CustomerGatewayArn) != customerGatewayARN {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusCustomerGatewayAssociationState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, customerGatewayARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomerGatewayAssociationByTwoPartKey(ctx, conn, globalNetworkID, customerGatewayARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitCustomerGatewayAssociationCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, customerGatewayARN string, timeout time.Duration) (*networkmanager.CustomerGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.CustomerGatewayAssociationStatePending},
		Target:  []string{networkmanager.CustomerGatewayAssociationStateAvailable},
		Timeout: timeout,
		Refresh: statusCustomerGatewayAssociationState(ctx, conn, globalNetworkID, customerGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CustomerGatewayAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitCustomerGatewayAssociationDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, customerGatewayARN string, timeout time.Duration) (*networkmanager.CustomerGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.CustomerGatewayAssociationStateAvailable, networkmanager.CustomerGatewayAssociationStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusCustomerGatewayAssociationState(ctx, conn, globalNetworkID, customerGatewayARN),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.CustomerGatewayAssociation); ok {
		return output, err
	}

	return nil, err
}

const customerGatewayAssociationIDSeparator = ","

func CustomerGatewayAssociationCreateResourceID(globalNetworkID, customerGatewayARN string) string {
	parts := []string{globalNetworkID, customerGatewayARN}
	id := strings.Join(parts, customerGatewayAssociationIDSeparator)

	return id
}

func CustomerGatewayAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, customerGatewayAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sCUSTOMER-GATEWAY-ARN", id, customerGatewayAssociationIDSeparator)
}

const (
	customerGatewayAssociationResourceNotFoundExceptionTimeout = 1 * time.Minute
)
