// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_inbound_connection_accepter", name="Inbound Connection Accepter")
func resourceInboundConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInboundConnectionAccepterCreate,
		ReadWithoutTimeout:   resourceInboundConnectionRead,
		DeleteWithoutTimeout: resourceInboundConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m any) (result []*schema.ResourceData, err error) {
				d.Set(names.AttrConnectionID, d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInboundConnectionAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	input := &opensearch.AcceptInboundConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	_, err := conn.AcceptInboundConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting OpenSearch Inbound Connection (%s): %s", connectionID, err)
	}

	d.SetId(connectionID)

	if _, err := waitInboundConnectionAccepted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) accept: %s", d.Id(), err)
	}

	return append(diags, resourceInboundConnectionRead(ctx, d, meta)...)
}

func resourceInboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	connection, err := findInboundConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Inbound Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Inbound Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrConnectionID, connection.ConnectionId)
	d.Set("connection_status", connection.ConnectionStatus.StatusCode)
	return diags
}

func resourceInboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	if d.Get("connection_status").(string) == string(awstypes.InboundConnectionStatusCodePendingAcceptance) {
		log.Printf("[DEBUG] Rejecting OpenSearch Inbound Connection: %s", d.Id())
		_, err := conn.RejectInboundConnection(ctx, &opensearch.RejectInboundConnectionInput{
			ConnectionId: aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "rejecting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitInboundConnectionRejected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) reject: %s", d.Id(), err)
		}

		return diags
	}

	log.Printf("[DEBUG] Deleting OpenSearch Inbound Connection: %s", d.Id())
	_, err := conn.DeleteInboundConnection(ctx, &opensearch.DeleteInboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitInboundConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findInboundConnectionByID(ctx context.Context, conn *opensearch.Client, id string) (*awstypes.InboundConnection, error) {
	input := &opensearch.DescribeInboundConnectionsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("connection-id"),
				Values: []string{id},
			},
		},
	}

	output, err := findInboundConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.ConnectionStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.ConnectionStatus.StatusCode; status == awstypes.InboundConnectionStatusCodeDeleted || status == awstypes.InboundConnectionStatusCodeRejected {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, err
}

func findInboundConnection(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribeInboundConnectionsInput) (*awstypes.InboundConnection, error) {
	output, err := findInboundConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInboundConnections(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribeInboundConnectionsInput) ([]awstypes.InboundConnection, error) {
	var output []awstypes.InboundConnection

	pages := opensearch.NewDescribeInboundConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Connections...)
	}

	return output, nil
}

func statusInboundConnection(ctx context.Context, conn *opensearch.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findInboundConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConnectionStatus.StatusCode), nil
	}
}

func waitInboundConnectionAccepted(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*awstypes.InboundConnection, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InboundConnectionStatusCodeProvisioning, awstypes.InboundConnectionStatusCodeApproved),
		Target:  enum.Slice(awstypes.InboundConnectionStatusCodeActive),
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitInboundConnectionRejected(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*awstypes.InboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InboundConnectionStatusCodeRejecting),
		Target:  []string{},
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitInboundConnectionDeleted(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*awstypes.InboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InboundConnectionStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}
