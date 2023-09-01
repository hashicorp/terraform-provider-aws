// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func virtualInterfaceRead(ctx context.Context, id string, conn *directconnect.DirectConnect) (*directconnect.VirtualInterface, error) {
	resp, state, err := virtualInterfaceStateRefresh(ctx, conn, id)()
	if err != nil {
		return nil, fmt.Errorf("reading Direct Connect virtual interface (%s): %s", id, err)
	}
	if state == directconnect.VirtualInterfaceStateDeleted {
		return nil, nil
	}

	return resp.(*directconnect.VirtualInterface), nil
}

func virtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	if d.HasChange("mtu") {
		req := &directconnect.UpdateVirtualInterfaceAttributesInput{
			Mtu:                aws.Int64(int64(d.Get("mtu").(int))),
			VirtualInterfaceId: aws.String(d.Id()),
		}
		log.Printf("[DEBUG] Modifying Direct Connect virtual interface attributes: %s", req)
		_, err := conn.UpdateVirtualInterfaceAttributesWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect virtual interface (%s) attributes: %s", d.Id(), err)
		}
	}
	if d.HasChange("sitelink_enabled") {
		req := &directconnect.UpdateVirtualInterfaceAttributesInput{
			EnableSiteLink:     aws.Bool(d.Get("sitelink_enabled").(bool)),
			VirtualInterfaceId: aws.String(d.Id()),
		}
		log.Printf("[DEBUG] Modifying Direct Connect virtual interface attributes: %s", req)
		_, err := conn.UpdateVirtualInterfaceAttributesWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect virtual interface (%s) attributes: %s", d.Id(), err)
		}
	}

	return diags
}

func virtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect virtual interface: %s", d.Id())
	_, err := conn.DeleteVirtualInterfaceWithContext(ctx, &directconnect.DeleteVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect virtual interface (%s): %s", d.Id(), err)
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStateDeleting,
			directconnect.VirtualInterfaceStateDown,
			directconnect.VirtualInterfaceStatePending,
			directconnect.VirtualInterfaceStateRejected,
			directconnect.VirtualInterfaceStateVerifying,
		},
		Target: []string{
			directconnect.VirtualInterfaceStateDeleted,
		},
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

func virtualInterfaceStateRefresh(ctx context.Context, conn *directconnect.DirectConnect, vifId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVirtualInterfacesWithContext(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(vifId),
		})
		if err != nil {
			return nil, "", err
		}

		n := len(resp.VirtualInterfaces)
		switch n {
		case 0:
			return "", directconnect.VirtualInterfaceStateDeleted, nil

		case 1:
			vif := resp.VirtualInterfaces[0]
			return vif, aws.StringValue(vif.VirtualInterfaceState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect virtual interfaces for %s, expected 1", n, vifId)
		}
	}
}

func virtualInterfaceWaitUntilAvailable(ctx context.Context, conn *directconnect.DirectConnect, vifId string, timeout time.Duration, pending, target []string) error {
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
