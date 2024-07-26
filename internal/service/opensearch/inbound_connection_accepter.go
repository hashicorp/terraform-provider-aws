// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_inbound_connection_accepter")
func ResourceInboundConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInboundConnectionAccepterCreate,
		ReadWithoutTimeout:   resourceInboundConnectionRead,
		DeleteWithoutTimeout: resourceInboundConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) (result []*schema.ResourceData, err error) {
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

func resourceInboundConnectionAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	input := &opensearchservice.AcceptInboundConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	_, err := conn.AcceptInboundConnectionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting OpenSearch Inbound Connection (%s): %s", connectionID, err)
	}

	d.SetId(connectionID)

	if _, err := waitInboundConnectionAccepted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) accept: %s", d.Id(), err)
	}

	return append(diags, resourceInboundConnectionRead(ctx, d, meta)...)
}

func resourceInboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	connection, err := FindInboundConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Inbound Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Inbound Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrConnectionID, connection.ConnectionId)
	d.Set("connection_status", connection.ConnectionStatus.StatusCode)
	return nil
}

func resourceInboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	if d.Get("connection_status").(string) == opensearchservice.InboundConnectionStatusCodePendingAcceptance {
		log.Printf("[DEBUG] Rejecting OpenSearch Inbound Connection: %s", d.Id())
		_, err := conn.RejectInboundConnectionWithContext(ctx, &opensearchservice.RejectInboundConnectionInput{
			ConnectionId: aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "rejecting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitInboundConnectionRejected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) reject: %s", d.Id(), err)
		}

		return nil
	}

	log.Printf("[DEBUG] Deleting OpenSearch Inbound Connection: %s", d.Id())
	_, err := conn.DeleteInboundConnectionWithContext(ctx, &opensearchservice.DeleteInboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitInboundConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindInboundConnectionByID(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) (*opensearchservice.InboundConnection, error) {
	input := &opensearchservice.DescribeInboundConnectionsInput{
		Filters: []*opensearchservice.Filter{
			{
				Name:   aws.String("connection-id"),
				Values: aws.StringSlice([]string{id}),
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

	if status := aws.StringValue(output.ConnectionStatus.StatusCode); status == opensearchservice.InboundConnectionStatusCodeDeleted || status == opensearchservice.InboundConnectionStatusCodeRejected {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, err
}

func findInboundConnection(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribeInboundConnectionsInput) (*opensearchservice.InboundConnection, error) {
	output, err := findInboundConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findInboundConnections(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribeInboundConnectionsInput) ([]*opensearchservice.InboundConnection, error) {
	var output []*opensearchservice.InboundConnection

	err := conn.DescribeInboundConnectionsPagesWithContext(ctx, input, func(page *opensearchservice.DescribeInboundConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connections {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusInboundConnection(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInboundConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectionStatus.StatusCode), nil
	}
}

func waitInboundConnectionAccepted(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) (*opensearchservice.InboundConnection, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{opensearchservice.InboundConnectionStatusCodeProvisioning, opensearchservice.InboundConnectionStatusCodeApproved},
		Target:  []string{opensearchservice.InboundConnectionStatusCodeActive},
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitInboundConnectionRejected(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) (*opensearchservice.InboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{opensearchservice.InboundConnectionStatusCodeRejecting},
		Target:  []string{},
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitInboundConnectionDeleted(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) (*opensearchservice.InboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{opensearchservice.InboundConnectionStatusCodeDeleting},
		Target:  []string{},
		Refresh: statusInboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.InboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}
