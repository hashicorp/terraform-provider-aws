// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func virtualInterfaceRead(ctx context.Context, id string, conn *directconnect.Client) (*awstypes.VirtualInterface, error) {
	resp, state, err := virtualInterfaceStateRefresh(ctx, conn, id)()
	if err != nil {
		return nil, fmt.Errorf("reading Direct Connect virtual interface (%s): %s", id, err)
	}
	if state == string(awstypes.VirtualInterfaceStateDeleted) {
		return nil, nil
	}

	return resp.(*awstypes.VirtualInterface), nil
}

func virtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange("mtu") {
		req := &directconnect.UpdateVirtualInterfaceAttributesInput{
			Mtu:                aws.Int32(int32(d.Get("mtu").(int))),
			VirtualInterfaceId: aws.String(d.Id()),
		}
		log.Printf("[DEBUG] Modifying Direct Connect virtual interface attributes: %#v", req)
		_, err := conn.UpdateVirtualInterfaceAttributes(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect virtual interface (%s) attributes: %s", d.Id(), err)
		}
	}
	if d.HasChange("sitelink_enabled") {
		req := &directconnect.UpdateVirtualInterfaceAttributesInput{
			EnableSiteLink:     aws.Bool(d.Get("sitelink_enabled").(bool)),
			VirtualInterfaceId: aws.String(d.Id()),
		}
		log.Printf("[DEBUG] Modifying Direct Connect virtual interface attributes: %#v", req)
		_, err := conn.UpdateVirtualInterfaceAttributes(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect virtual interface (%s) attributes: %s", d.Id(), err)
		}
	}

	return diags
}

func virtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect virtual interface: %s", d.Id())
	_, err := conn.DeleteVirtualInterface(ctx, &directconnect.DeleteVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(d.Id()),
	})
	if err != nil {
		if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect virtual interface (%s): %s", d.Id(), err)
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.VirtualInterfaceStateAvailable,
			awstypes.VirtualInterfaceStateConfirming,
			awstypes.VirtualInterfaceStateDeleting,
			awstypes.VirtualInterfaceStateDown,
			awstypes.VirtualInterfaceStatePending,
			awstypes.VirtualInterfaceStateRejected,
			awstypes.VirtualInterfaceStateVerifying),
		Target:     enum.Slice(awstypes.VirtualInterfaceStateDeleted),
		Refresh:    virtualInterfaceStateRefresh(ctx, conn, d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect virtual interface (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func virtualInterfaceStateRefresh(ctx context.Context, conn *directconnect.Client, vifId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVirtualInterfaces(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(vifId),
		})
		if err != nil {
			return nil, "", err
		}

		n := len(resp.VirtualInterfaces)
		switch n {
		case 0:
			return "", string(awstypes.VirtualInterfaceStateDeleted), nil

		case 1:
			vif := resp.VirtualInterfaces[0]
			return vif, string(vif.VirtualInterfaceState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect virtual interfaces for %s, expected 1", n, vifId)
		}
	}
}

func virtualInterfaceWaitUntilAvailable(ctx context.Context, conn *directconnect.Client, vifId string, timeout time.Duration, pending, target []string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    virtualInterfaceStateRefresh(ctx, conn, vifId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for Direct Connect virtual interface (%s) to become available: %s", vifId, err)
	}

	return nil
}
