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
)

// @SDKResource("aws_networkmanager_link_association", name="Link Association")
func resourceLinkAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkAssociationCreate,
		ReadWithoutTimeout:   resourceLinkAssociationRead,
		DeleteWithoutTimeout: resourceLinkAssociationDelete,

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
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLinkAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	linkID := d.Get("link_id").(string)
	deviceID := d.Get("device_id").(string)
	id := linkAssociationCreateResourceID(globalNetworkID, linkID, deviceID)
	input := &networkmanager.AssociateLinkInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	}

	log.Printf("[DEBUG] Creating Network Manager Link Association: %#v", input)
	_, err := conn.AssociateLink(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Link Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitLinkAssociationCreated(ctx, conn, globalNetworkID, linkID, deviceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLinkAssociationRead(ctx, d, meta)...)
}

func resourceLinkAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, linkID, deviceID, err := linkAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Link Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Link Association (%s): %s", d.Id(), err)
	}

	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)

	return diags
}

func resourceLinkAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	globalNetworkID, linkID, deviceID, err := linkAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Network Manager Link Association: %s", d.Id())
	_, err = conn.DisassociateLink(ctx, &networkmanager.DisassociateLinkInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	})

	if globalNetworkIDNotFoundError(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Link Association (%s): %s", d.Id(), err)
	}

	if _, err := waitLinkAssociationDeleted(ctx, conn, globalNetworkID, linkID, deviceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Link Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findLinkAssociation(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetLinkAssociationsInput) (*awstypes.LinkAssociation, error) {
	output, err := findLinkAssociations(ctx, conn, input)

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

func findLinkAssociations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetLinkAssociationsInput) ([]awstypes.LinkAssociation, error) {
	var output []awstypes.LinkAssociation

	pages := networkmanager.NewGetLinkAssociationsPaginator(conn, input)
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

		output = append(output, page.LinkAssociations...)
	}

	return output, nil
}

func findLinkAssociationByThreePartKey(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID, deviceID string) (*awstypes.LinkAssociation, error) {
	input := &networkmanager.GetLinkAssociationsInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	}

	output, err := findLinkAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.LinkAssociationState; state == awstypes.LinkAssociationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalNetworkId) != globalNetworkID || aws.ToString(output.LinkId) != linkID || aws.ToString(output.DeviceId) != deviceID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLinkAssociationState(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID, deviceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LinkAssociationState), nil
	}
}

func waitLinkAssociationCreated(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID, deviceID string, timeout time.Duration) (*awstypes.LinkAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LinkAssociationStatePending),
		Target:  enum.Slice(awstypes.LinkAssociationStateAvailable),
		Timeout: timeout,
		Refresh: statusLinkAssociationState(ctx, conn, globalNetworkID, linkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LinkAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitLinkAssociationDeleted(ctx context.Context, conn *networkmanager.Client, globalNetworkID, linkID, deviceID string, timeout time.Duration) (*awstypes.LinkAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LinkAssociationStateDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusLinkAssociationState(ctx, conn, globalNetworkID, linkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LinkAssociation); ok {
		return output, err
	}

	return nil, err
}

const linkAssociationIDSeparator = ","

func linkAssociationCreateResourceID(globalNetworkID, linkID, deviceID string) string {
	parts := []string{globalNetworkID, linkID, deviceID}
	id := strings.Join(parts, linkAssociationIDSeparator)

	return id
}

func linkAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, linkAssociationIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sLINK-ID%[2]sDEVICE-ID", id, linkAssociationIDSeparator)
}
