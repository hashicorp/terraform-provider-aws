// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_opensearch_inbound_connection_accepter")
func ResourceInboundConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInboundConnectionAccepterCreate,
		ReadWithoutTimeout:   resourceInboundConnectionRead,
		DeleteWithoutTimeout: resourceInboundConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) (result []*schema.ResourceData, err error) {
				d.Set("connection_id", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"connection_id": {
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
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	// Create the Inbound Connection
	acceptOpts := &opensearchservice.AcceptInboundConnectionInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
	}

	log.Printf("[DEBUG] Inbound Connection Accept options: %#v", acceptOpts)

	resp, err := conn.AcceptInboundConnectionWithContext(ctx, acceptOpts)
	if err != nil {
		return diag.Errorf("accepting Inbound Connection: %s", err)
	}

	// Get the ID and store it
	d.SetId(aws.StringValue(resp.Connection.ConnectionId))
	log.Printf("[INFO] Inbound Connection ID: %s", d.Id())

	err = inboundConnectionWaitUntilActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("waiting for Inbound Connection to become active: %s", err)
	}

	return resourceInboundConnectionRead(ctx, d, meta)
}

func resourceInboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	ccscRaw, statusCode, err := inboundConnectionRefreshState(ctx, conn, d.Id())()

	if err != nil {
		return diag.Errorf("reading Inbound Connection: %s", err)
	}

	ccsc := ccscRaw.(*opensearchservice.InboundConnection)
	log.Printf("[DEBUG] Inbound Connection response: %#v", ccsc)

	d.Set("connection_id", ccsc.ConnectionId)
	d.Set("connection_status", statusCode)
	return nil
}

func resourceInboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	req := &opensearchservice.DeleteInboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteInboundConnectionWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Inbound Connection (%s): %s", d.Id(), err)
	}

	if err := waitForInboundConnectionDeletion(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for VPC Peering Connection (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func inboundConnectionRefreshState(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInboundConnectionsWithContext(ctx, &opensearchservice.DescribeInboundConnectionsInput{
			Filters: []*opensearchservice.Filter{
				{
					Name:   aws.String("connection-id"),
					Values: []*string{aws.String(id)},
				},
			},
		})
		if err != nil {
			return nil, "", err
		}

		if resp == nil || resp.Connections == nil ||
			len(resp.Connections) == 0 || resp.Connections[0] == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our connection yet. Return an empty state.
			return nil, "", nil
		}
		ccsc := resp.Connections[0]
		if ccsc.ConnectionStatus == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our connection yet. Return an empty state.
			return nil, "", nil
		}
		statusCode := aws.StringValue(ccsc.ConnectionStatus.StatusCode)

		return ccsc, statusCode, nil
	}
}

func inboundConnectionWaitUntilActive(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for Inbound Connection (%s) to become available.", id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.InboundConnectionStatusCodeProvisioning,
			opensearchservice.InboundConnectionStatusCodeApproved,
		},
		Target: []string{
			opensearchservice.InboundConnectionStatusCodeActive,
		},
		Refresh: inboundConnectionRefreshState(ctx, conn, id),
		Timeout: timeout,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for Inbound Connection (%s) to become available: %s", id, err)
	}
	return nil
}

func waitForInboundConnectionDeletion(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.InboundConnectionStatusCodeDeleting,
		},
		Target: []string{
			opensearchservice.InboundConnectionStatusCodeDeleted,
		},
		Refresh: inboundConnectionRefreshState(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
