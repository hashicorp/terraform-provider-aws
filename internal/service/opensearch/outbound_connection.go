// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_opensearch_outbound_connection")
func ResourceOutboundConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOutboundConnectionCreate,
		ReadWithoutTimeout:   resourceOutboundConnectionRead,
		DeleteWithoutTimeout: resourceOutboundConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"connection_alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(opensearchservice.ConnectionMode_Values(), false),
			},
			"connection_properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cross_cluster_search": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"skip_unavailable": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"local_domain_info":  outboundConnectionDomainInfoSchema(),
			"remote_domain_info": outboundConnectionDomainInfoSchema(),
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accept_connection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceOutboundConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	// Create the Outbound Connection
	createOpts := &opensearchservice.CreateOutboundConnectionInput{
		ConnectionAlias:      aws.String(d.Get("connection_alias").(string)),
		ConnectionMode:       aws.String(d.Get("connection_mode").(string)),
		ConnectionProperties: expandOutboundConnectionConnectionProperties(d.Get("connection_properties").([]interface{})),
		LocalDomainInfo:      expandOutboundConnectionDomainInfo(d.Get("local_domain_info").([]interface{})),
		RemoteDomainInfo:     expandOutboundConnectionDomainInfo(d.Get("remote_domain_info").([]interface{})),
	}

	log.Printf("[DEBUG] Outbound Connection Create options: %#v", createOpts)

	resp, err := conn.CreateOutboundConnectionWithContext(ctx, createOpts)
	if err != nil {
		return diag.Errorf("creating Outbound Connection: %s", err)
	}

	// Get the ID and store it
	d.SetId(aws.StringValue(resp.ConnectionId))
	log.Printf("[INFO] Outbound Connection ID: %s", d.Id())

	err = outboundConnectionWaitUntilAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.Errorf("waiting for Outbound Connection to become available: %s", err)
	}

	if d.Get("accept_connection").(bool) {
		if err := inboundConnectionAccept(ctx, d, conn); err != nil {
			return diag.Errorf("unable to accept Connection: %s", err)
		}
	}

	return resourceOutboundConnectionRead(ctx, d, meta)
}

func resourceOutboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	ccscRaw, statusCode, err := outboundConnectionRefreshState(ctx, conn, d.Id())()

	if err != nil {
		return diag.Errorf("reading Outbound Connection: %s", err)
	}

	ccsc := ccscRaw.(*opensearchservice.OutboundConnection)
	log.Printf("[DEBUG] Outbound Connection response: %#v", ccsc)

	if !d.IsNewResource() && statusCode == opensearchservice.OutboundConnectionStatusCodeDeleted {
		log.Printf("[INFO] Outbound Connection (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("connection_alias", ccsc.ConnectionAlias)
	d.Set("connection_mode", ccsc.ConnectionMode)
	d.Set("connection_properties", flattenOutboundConnectionConnectionProperties(ccsc.ConnectionProperties))
	d.Set("remote_domain_info", flattenOutboundConnectionDomainInfo(ccsc.RemoteDomainInfo))
	d.Set("local_domain_info", flattenOutboundConnectionDomainInfo(ccsc.LocalDomainInfo))
	d.Set("connection_status", statusCode)

	return nil
}

func resourceOutboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	req := &opensearchservice.DeleteOutboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	}

	_, err := conn.DeleteOutboundConnectionWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Outbound Connection (%s): %s", d.Id(), err)
	}

	if err := waitForOutboundConnectionDeletion(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for VPC Peering Connection (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func outboundConnectionRefreshState(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeOutboundConnectionsWithContext(ctx, &opensearchservice.DescribeOutboundConnectionsInput{
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

		// A Outbound Connection can exist in a failed state,
		// thus we short circuit before the time out would occur.
		if statusCode == opensearchservice.OutboundConnectionStatusCodeValidationFailed {
			return nil, statusCode, errors.New(aws.StringValue(ccsc.ConnectionStatus.Message))
		}

		return ccsc, statusCode, nil
	}
}

func outboundConnectionWaitUntilAvailable(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Waiting for Outbound Connection (%s) to become available.", id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.OutboundConnectionStatusCodeValidating,
			opensearchservice.OutboundConnectionStatusCodeProvisioning,
		},
		Target: []string{
			opensearchservice.OutboundConnectionStatusCodePendingAcceptance,
			opensearchservice.OutboundConnectionStatusCodeActive,
			opensearchservice.OutboundConnectionStatusCodeApproved,
			opensearchservice.OutboundConnectionStatusCodeRejected,
			opensearchservice.OutboundConnectionStatusCodeValidationFailed,
		},
		Refresh: outboundConnectionRefreshState(ctx, conn, id),
		Timeout: timeout,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("waiting for Outbound Connection (%s) to become available: %s", id, err)
	}
	return nil
}

func waitForOutboundConnectionDeletion(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.OutboundConnectionStatusCodeActive,
			opensearchservice.OutboundConnectionStatusCodePendingAcceptance,
			opensearchservice.OutboundConnectionStatusCodeDeleting,
			opensearchservice.OutboundConnectionStatusCodeRejecting,
		},
		Target: []string{
			opensearchservice.OutboundConnectionStatusCodeDeleted,
		},
		Refresh: outboundConnectionRefreshState(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func outboundConnectionDomainInfoSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"owner_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"domain_name": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"region": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
			},
		},
	}
}

func inboundConnectionAccept(ctx context.Context, d *schema.ResourceData, conn *opensearchservice.OpenSearchService) error {
	// Create the Inbound Connection
	acceptOpts := &opensearchservice.AcceptInboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Inbound Connection Accept options: %#v", acceptOpts)

	_, err := conn.AcceptInboundConnectionWithContext(ctx, acceptOpts)
	if err != nil {
		return err
	}

	err = inboundConnectionWaitUntilActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	return err
}

func expandOutboundConnectionDomainInfo(vOptions []interface{}) *opensearchservice.DomainInformationContainer {
	if len(vOptions) == 0 || vOptions[0] == nil {
		return nil
	}

	mOptions := vOptions[0].(map[string]interface{})

	return &opensearchservice.DomainInformationContainer{
		AWSDomainInformation: &opensearchservice.AWSDomainInformation{
			DomainName: aws.String(mOptions["domain_name"].(string)),
			OwnerId:    aws.String(mOptions["owner_id"].(string)),
			Region:     aws.String(mOptions["region"].(string)),
		},
	}
}

func flattenOutboundConnectionDomainInfo(domainInfo *opensearchservice.DomainInformationContainer) []interface{} {
	if domainInfo == nil || domainInfo.AWSDomainInformation == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"owner_id":    aws.StringValue(domainInfo.AWSDomainInformation.OwnerId),
		"domain_name": aws.StringValue(domainInfo.AWSDomainInformation.DomainName),
		"region":      aws.StringValue(domainInfo.AWSDomainInformation.Region),
	}}
}

func expandOutboundConnectionConnectionProperties(cProperties []interface{}) *opensearchservice.ConnectionProperties {
	if len(cProperties) == 0 || cProperties[0] == nil {
		return nil
	}

	mOptions := cProperties[0].(map[string]interface{})

	return &opensearchservice.ConnectionProperties{
		CrossClusterSearch: expandOutboundConnectionCrossClusterSearchConnectionProperties(mOptions["cross_cluster_search"].([]interface{})),
	}
}

func flattenOutboundConnectionConnectionProperties(cProperties *opensearchservice.ConnectionProperties) []interface{} {
	if cProperties == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"cross_cluster_search": flattenOutboundConnectionCrossClusterSearchConnectionProperties(cProperties.CrossClusterSearch),
		"endpoint":             aws.StringValue(cProperties.Endpoint),
	}}
}

func expandOutboundConnectionCrossClusterSearchConnectionProperties(cProperties []interface{}) *opensearchservice.CrossClusterSearchConnectionProperties {
	if len(cProperties) == 0 || cProperties[0] == nil {
		return nil
	}

	mOptions := cProperties[0].(map[string]interface{})

	return &opensearchservice.CrossClusterSearchConnectionProperties{
		SkipUnavailable: aws.String(mOptions["skip_unavailable"].(string)),
	}
}

func flattenOutboundConnectionCrossClusterSearchConnectionProperties(cProperties *opensearchservice.CrossClusterSearchConnectionProperties) []interface{} {
	if cProperties == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"skip_unavailable": aws.StringValue(cProperties.SkipUnavailable),
	}}
}
