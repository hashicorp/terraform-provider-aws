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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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

		SchemaFunc: func() map[string]*schema.Schema {
			outboundConnectionDomainInfoSchema := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrDomainName: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							names.AttrOwnerID: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							names.AttrRegion: {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
						},
					},
				}
			}

			return map[string]*schema.Schema{
				"accept_connection": {
					Type:     schema.TypeBool,
					Optional: true,
					ForceNew: true,
					Default:  false,
				},
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
							names.AttrEndpoint: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"connection_status": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"local_domain_info":  outboundConnectionDomainInfoSchema(),
				"remote_domain_info": outboundConnectionDomainInfoSchema(),
			}
		},
	}
}

func resourceOutboundConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	connectionAlias := d.Get("connection_alias").(string)
	input := &opensearchservice.CreateOutboundConnectionInput{
		ConnectionAlias:      aws.String(connectionAlias),
		ConnectionMode:       aws.String(d.Get("connection_mode").(string)),
		ConnectionProperties: expandOutboundConnectionConnectionProperties(d.Get("connection_properties").([]interface{})),
		LocalDomainInfo:      expandOutboundConnectionDomainInfo(d.Get("local_domain_info").([]interface{})),
		RemoteDomainInfo:     expandOutboundConnectionDomainInfo(d.Get("remote_domain_info").([]interface{})),
	}

	output, err := conn.CreateOutboundConnectionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Outbound Connection (%s): %s", connectionAlias, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	if _, err := waitOutboundConnectionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Outbound Connection (%s) create: %s", d.Id(), err)
	}

	if d.Get("accept_connection").(bool) {
		input := &opensearchservice.AcceptInboundConnectionInput{
			ConnectionId: aws.String(d.Id()),
		}

		_, err := conn.AcceptInboundConnectionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitInboundConnectionAccepted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) accept: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOutboundConnectionRead(ctx, d, meta)...)
}

func resourceOutboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	connection, err := FindOutboundConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Outbound Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Outbound Connection (%s): %s", d.Id(), err)
	}

	d.Set("connection_alias", connection.ConnectionAlias)
	d.Set("connection_mode", connection.ConnectionMode)
	d.Set("connection_properties", flattenOutboundConnectionConnectionProperties(connection.ConnectionProperties))
	d.Set("connection_status", connection.ConnectionStatus.StatusCode)
	d.Set("remote_domain_info", flattenOutboundConnectionDomainInfo(connection.RemoteDomainInfo))
	d.Set("local_domain_info", flattenOutboundConnectionDomainInfo(connection.LocalDomainInfo))

	return nil
}

func resourceOutboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Outbound Connection: %s", d.Id())
	_, err := conn.DeleteOutboundConnectionWithContext(ctx, &opensearchservice.DeleteOutboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Outbound Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitOutboundConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Outbound Connection (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindOutboundConnectionByID(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) (*opensearchservice.OutboundConnection, error) {
	input := &opensearchservice.DescribeOutboundConnectionsInput{
		Filters: []*opensearchservice.Filter{
			{
				Name:   aws.String("connection-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	output, err := findOutboundConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.ConnectionStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := aws.StringValue(output.ConnectionStatus.StatusCode); status == opensearchservice.OutboundConnectionStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, err
}

func findOutboundConnection(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribeOutboundConnectionsInput) (*opensearchservice.OutboundConnection, error) {
	output, err := findOutboundConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOutboundConnections(ctx context.Context, conn *opensearchservice.OpenSearchService, input *opensearchservice.DescribeOutboundConnectionsInput) ([]*opensearchservice.OutboundConnection, error) {
	var output []*opensearchservice.OutboundConnection

	err := conn.DescribeOutboundConnectionsPagesWithContext(ctx, input, func(page *opensearchservice.DescribeOutboundConnectionsOutput, lastPage bool) bool {
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

func statusOutboundConnection(ctx context.Context, conn *opensearchservice.OpenSearchService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindOutboundConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectionStatus.StatusCode), nil
	}
}

func waitOutboundConnectionCreated(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) (*opensearchservice.OutboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{opensearchservice.OutboundConnectionStatusCodeValidating, opensearchservice.OutboundConnectionStatusCodeProvisioning},
		Target: []string{
			opensearchservice.OutboundConnectionStatusCodePendingAcceptance,
			opensearchservice.OutboundConnectionStatusCodeActive,
			opensearchservice.OutboundConnectionStatusCodeApproved,
			opensearchservice.OutboundConnectionStatusCodeRejected,
			opensearchservice.OutboundConnectionStatusCodeValidationFailed,
		},
		Refresh: statusOutboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.OutboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitOutboundConnectionDeleted(ctx context.Context, conn *opensearchservice.OpenSearchService, id string, timeout time.Duration) (*opensearchservice.OutboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			opensearchservice.OutboundConnectionStatusCodeActive,
			opensearchservice.OutboundConnectionStatusCodePendingAcceptance,
			opensearchservice.OutboundConnectionStatusCodeDeleting,
			opensearchservice.OutboundConnectionStatusCodeRejecting,
		},
		Target:  []string{},
		Refresh: statusOutboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*opensearchservice.OutboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func expandOutboundConnectionDomainInfo(vOptions []interface{}) *opensearchservice.DomainInformationContainer {
	if len(vOptions) == 0 || vOptions[0] == nil {
		return nil
	}

	mOptions := vOptions[0].(map[string]interface{})

	return &opensearchservice.DomainInformationContainer{
		AWSDomainInformation: &opensearchservice.AWSDomainInformation{
			DomainName: aws.String(mOptions[names.AttrDomainName].(string)),
			OwnerId:    aws.String(mOptions[names.AttrOwnerID].(string)),
			Region:     aws.String(mOptions[names.AttrRegion].(string)),
		},
	}
}

func flattenOutboundConnectionDomainInfo(domainInfo *opensearchservice.DomainInformationContainer) []interface{} {
	if domainInfo == nil || domainInfo.AWSDomainInformation == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		names.AttrOwnerID:    aws.StringValue(domainInfo.AWSDomainInformation.OwnerId),
		names.AttrDomainName: aws.StringValue(domainInfo.AWSDomainInformation.DomainName),
		names.AttrRegion:     aws.StringValue(domainInfo.AWSDomainInformation.Region),
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
		names.AttrEndpoint:     aws.StringValue(cProperties.Endpoint),
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
