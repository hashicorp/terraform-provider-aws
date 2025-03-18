// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func virtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if d.HasChange("mtu") {
		input := &directconnect.UpdateVirtualInterfaceAttributesInput{
			Mtu:                aws.Int32(int32(d.Get("mtu").(int))),
			VirtualInterfaceId: aws.String(d.Id()),
		}

		_, err := conn.UpdateVirtualInterfaceAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect Virtual Interface (%s) Mtu attribute: %s", d.Id(), err)
		}
	}

	if d.HasChange("sitelink_enabled") {
		input := &directconnect.UpdateVirtualInterfaceAttributesInput{
			EnableSiteLink:     aws.Bool(d.Get("sitelink_enabled").(bool)),
			VirtualInterfaceId: aws.String(d.Id()),
		}

		_, err := conn.UpdateVirtualInterfaceAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Direct Connect Virtual Interface (%s) EnableSiteLink attribute: %s", d.Id(), err)
		}
	}

	return diags
}

func virtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect Virtual Interface: %s", d.Id())
	input := directconnect.DeleteVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(d.Id()),
	}
	_, err := conn.DeleteVirtualInterface(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Virtual Interface (%s): %s", d.Id(), err)
	}

	if _, err := waitVirtualInterfaceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Virtual Interface (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findVirtualInterfaceByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.VirtualInterface, error) {
	input := &directconnect.DescribeVirtualInterfacesInput{
		VirtualInterfaceId: aws.String(id),
	}
	output, err := findVirtualInterface(ctx, conn, input, tfslices.PredicateTrue[*awstypes.VirtualInterface]())

	if err != nil {
		return nil, err
	}

	if state := output.VirtualInterfaceState; state == awstypes.VirtualInterfaceStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualInterface(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeVirtualInterfacesInput, filter tfslices.Predicate[*awstypes.VirtualInterface]) (*awstypes.VirtualInterface, error) {
	output, err := findVirtualInterfaces(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVirtualInterfaces(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeVirtualInterfacesInput, filter tfslices.Predicate[*awstypes.VirtualInterface]) ([]awstypes.VirtualInterface, error) {
	output, err := conn.DescribeVirtualInterfaces(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.VirtualInterfaces, tfslices.PredicateValue(filter)), nil
}

func statusVirtualInterface(ctx context.Context, conn *directconnect.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findVirtualInterfaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.VirtualInterfaceState), nil
	}
}

func waitVirtualInterfaceAvailable(ctx context.Context, conn *directconnect.Client, id string, pending, target []string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    statusVirtualInterface(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VirtualInterface); ok {
		return output, err
	}

	return nil, err
}

func waitVirtualInterfaceDeleted(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.VirtualInterfaceStateAvailable,
			awstypes.VirtualInterfaceStateConfirming,
			awstypes.VirtualInterfaceStateDeleting,
			awstypes.VirtualInterfaceStateDown,
			awstypes.VirtualInterfaceStatePending,
			awstypes.VirtualInterfaceStateRejected,
			awstypes.VirtualInterfaceStateVerifying,
		),
		Target:     []string{},
		Refresh:    statusVirtualInterface(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VirtualInterface); ok {
		return output, err
	}

	return nil, err
}
