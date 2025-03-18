// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_global_network", name="Global Network")
// @Tags(identifierAttribute="arn")
func resourceGlobalNetwork() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalNetworkCreate,
		ReadWithoutTimeout:   resourceGlobalNetworkRead,
		UpdateWithoutTimeout: resourceGlobalNetworkUpdate,
		DeleteWithoutTimeout: resourceGlobalNetworkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGlobalNetworkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	input := &networkmanager.CreateGlobalNetworkInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Network Manager Global Network: %#v", input)
	output, err := conn.CreateGlobalNetwork(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Global Network: %s", err)
	}

	d.SetId(aws.ToString(output.GlobalNetwork.GlobalNetworkId))

	if _, err := waitGlobalNetworkCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Global Network (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalNetworkRead(ctx, d, meta)...)
}

func resourceGlobalNetworkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetwork, err := findGlobalNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Global Network %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Global Network (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, globalNetwork.GlobalNetworkArn)
	d.Set(names.AttrDescription, globalNetwork.Description)

	setTagsOut(ctx, globalNetwork.Tags)

	return diags
}

func resourceGlobalNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &networkmanager.UpdateGlobalNetworkInput{
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			GlobalNetworkId: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating Network Manager Global Network: %#v", input)
		_, err := conn.UpdateGlobalNetwork(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager Global Network (%s): %s", d.Id(), err)
		}

		if _, err := waitGlobalNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Global Network (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGlobalNetworkRead(ctx, d, meta)...)
}

func resourceGlobalNetworkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if diags := disassociateCustomerGateways(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); diags.HasError() {
		return diags
	}

	if diags := disassociateTransitGatewayConnectPeers(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); diags.HasError() {
		return diags
	}

	if diags := deregisterTransitGateways(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Deleting Network Manager Global Network: %s", d.Id())
	_, err := tfresource.RetryWhen(ctx, globalNetworkValidationExceptionTimeout,
		func() (any, error) {
			return conn.DeleteGlobalNetwork(ctx, &networkmanager.DeleteGlobalNetworkInput{
				GlobalNetworkId: aws.String(d.Id()),
			})
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "cannot be deleted due to existing devices, sites, or links") {
				return true, err
			}

			return false, err
		},
	)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Global Network (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalNetworkDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Global Network (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func deregisterTransitGateways(ctx context.Context, conn *networkmanager.Client, globalNetworkID string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	output, err := findTransitGatewayRegistrations(ctx, conn, &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})

	if tfresource.NotFound(err) {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Transit Gateway Registrations (%s): %s", globalNetworkID, err)
	}

	for _, v := range output {
		err := deregisterTransitGateway(ctx, conn, globalNetworkID, aws.ToString(v.TransitGatewayArn), timeout)

		if err != nil {
			diags = sdkdiag.AppendFromErr(diags, err)
		}
	}

	if diags.HasError() {
		return diags
	}

	return diags
}

func disassociateCustomerGateways(ctx context.Context, conn *networkmanager.Client, globalNetworkID string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	output, err := findCustomerGatewayAssociations(ctx, conn, &networkmanager.GetCustomerGatewayAssociationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})

	if tfresource.NotFound(err) {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Customer Gateway Associations (%s): %s", globalNetworkID, err)
	}

	for _, v := range output {
		err := disassociateCustomerGateway(ctx, conn, globalNetworkID, aws.ToString(v.CustomerGatewayArn), timeout)

		if err != nil {
			diags = sdkdiag.AppendFromErr(diags, err)
		}
	}

	if diags.HasError() {
		return diags
	}

	return diags
}

func disassociateTransitGatewayConnectPeers(ctx context.Context, conn *networkmanager.Client, globalNetworkID string, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	output, err := findTransitGatewayConnectPeerAssociations(ctx, conn, &networkmanager.GetTransitGatewayConnectPeerAssociationsInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	})

	if tfresource.NotFound(err) {
		err = nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Network Manager Transit Gateway Connect Peer Associations (%s): %s", globalNetworkID, err)
	}

	for _, v := range output {
		err := disassociateTransitGatewayConnectPeer(ctx, conn, globalNetworkID, aws.ToString(v.TransitGatewayConnectPeerArn), timeout)

		if err != nil {
			diags = sdkdiag.AppendFromErr(diags, err)
		}
	}

	if diags.HasError() {
		return diags
	}

	return diags
}

func globalNetworkIDNotFoundError(err error) bool {
	return validationExceptionFieldsMessageContains(err, awstypes.ValidationExceptionReasonFieldValidationFailed, "Global network not found")
}

func findGlobalNetwork(ctx context.Context, conn *networkmanager.Client, input *networkmanager.DescribeGlobalNetworksInput) (*awstypes.GlobalNetwork, error) {
	output, err := findGlobalNetworks(ctx, conn, input)

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

func findGlobalNetworks(ctx context.Context, conn *networkmanager.Client, input *networkmanager.DescribeGlobalNetworksInput) ([]awstypes.GlobalNetwork, error) {
	var output []awstypes.GlobalNetwork

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.GlobalNetworks...)
	}

	return output, nil
}

func findGlobalNetworkByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.GlobalNetwork, error) {
	input := &networkmanager.DescribeGlobalNetworksInput{
		GlobalNetworkIds: []string{id},
	}

	output, err := findGlobalNetwork(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusGlobalNetworkState(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findGlobalNetworkByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitGlobalNetworkCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.GlobalNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GlobalNetworkStatePending),
		Target:  enum.Slice(awstypes.GlobalNetworkStateAvailable),
		Timeout: timeout,
		Refresh: statusGlobalNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalNetworkDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.GlobalNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.GlobalNetworkStateDeleting),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusGlobalNetworkState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalNetwork); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalNetworkUpdated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.GlobalNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GlobalNetworkStateUpdating),
		Target:  enum.Slice(awstypes.GlobalNetworkStateAvailable),
		Timeout: timeout,
		Refresh: statusGlobalNetworkState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalNetwork); ok {
		return output, err
	}

	return nil, err
}

const (
	globalNetworkValidationExceptionTimeout = 2 * time.Minute
)
