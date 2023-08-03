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
)

// @SDKResource("aws_networkmanager_link_association")
func ResourceLinkAssociation() *schema.Resource {
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

func resourceLinkAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID := d.Get("global_network_id").(string)
	linkID := d.Get("link_id").(string)
	deviceID := d.Get("device_id").(string)
	id := LinkAssociationCreateResourceID(globalNetworkID, linkID, deviceID)
	input := &networkmanager.AssociateLinkInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	}

	log.Printf("[DEBUG] Creating Network Manager Link Association: %s", input)
	_, err := conn.AssociateLinkWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Network Manager Link Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitLinkAssociationCreated(ctx, conn, globalNetworkID, linkID, deviceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Network Manager Link Association (%s) create: %s", d.Id(), err)
	}

	return resourceLinkAssociationRead(ctx, d, meta)
}

func resourceLinkAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID, linkID, deviceID, err := LinkAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Link Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Network Manager Link Association (%s): %s", d.Id(), err)
	}

	d.Set("device_id", output.DeviceId)
	d.Set("global_network_id", output.GlobalNetworkId)
	d.Set("link_id", output.LinkId)

	return nil
}

func resourceLinkAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	globalNetworkID, linkID, deviceID, err := LinkAssociationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Network Manager Link Association: %s", d.Id())
	_, err = conn.DisassociateLinkWithContext(ctx, &networkmanager.DisassociateLinkInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	})

	if globalNetworkIDNotFoundError(err) || tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Network Manager Link Association (%s): %s", d.Id(), err)
	}

	if _, err := waitLinkAssociationDeleted(ctx, conn, globalNetworkID, linkID, deviceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Network Manager Link Association (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindLinkAssociation(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetLinkAssociationsInput) (*networkmanager.LinkAssociation, error) {
	output, err := FindLinkAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLinkAssociations(ctx context.Context, conn *networkmanager.NetworkManager, input *networkmanager.GetLinkAssociationsInput) ([]*networkmanager.LinkAssociation, error) {
	var output []*networkmanager.LinkAssociation

	err := conn.GetLinkAssociationsPagesWithContext(ctx, input, func(page *networkmanager.GetLinkAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LinkAssociations {
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

func FindLinkAssociationByThreePartKey(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID, deviceID string) (*networkmanager.LinkAssociation, error) {
	input := &networkmanager.GetLinkAssociationsInput{
		DeviceId:        aws.String(deviceID),
		GlobalNetworkId: aws.String(globalNetworkID),
		LinkId:          aws.String(linkID),
	}

	output, err := FindLinkAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.LinkAssociationState); state == networkmanager.LinkAssociationStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalNetworkId) != globalNetworkID || aws.StringValue(output.LinkId) != linkID || aws.StringValue(output.DeviceId) != deviceID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func statusLinkAssociationState(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID, deviceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLinkAssociationByThreePartKey(ctx, conn, globalNetworkID, linkID, deviceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.LinkAssociationState), nil
	}
}

func waitLinkAssociationCreated(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID, deviceID string, timeout time.Duration) (*networkmanager.LinkAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.LinkAssociationStatePending},
		Target:  []string{networkmanager.LinkAssociationStateAvailable},
		Timeout: timeout,
		Refresh: statusLinkAssociationState(ctx, conn, globalNetworkID, linkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.LinkAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitLinkAssociationDeleted(ctx context.Context, conn *networkmanager.NetworkManager, globalNetworkID, linkID, deviceID string, timeout time.Duration) (*networkmanager.LinkAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.LinkAssociationStateDeleting},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusLinkAssociationState(ctx, conn, globalNetworkID, linkID, deviceID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.LinkAssociation); ok {
		return output, err
	}

	return nil, err
}

const linkAssociationIDSeparator = ","

func LinkAssociationCreateResourceID(globalNetworkID, linkID, deviceID string) string {
	parts := []string{globalNetworkID, linkID, deviceID}
	id := strings.Join(parts, linkAssociationIDSeparator)

	return id
}

func LinkAssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, linkAssociationIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected GLOBAL-NETWORK-ID%[2]sLINK-ID%[2]sDEVICE-ID", id, linkAssociationIDSeparator)
}
