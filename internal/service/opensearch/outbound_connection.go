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

// @SDKResource("aws_opensearch_outbound_connection", name="Outbound Connection")
func resourceOutboundConnection() *schema.Resource {
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
					Type:             schema.TypeString,
					Optional:         true,
					ForceNew:         true,
					Computed:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ConnectionMode](),
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

func resourceOutboundConnectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	connectionAlias := d.Get("connection_alias").(string)
	input := &opensearch.CreateOutboundConnectionInput{
		ConnectionAlias:      aws.String(connectionAlias),
		ConnectionMode:       awstypes.ConnectionMode(d.Get("connection_mode").(string)),
		ConnectionProperties: expandOutboundConnectionConnectionProperties(d.Get("connection_properties").([]any)),
		LocalDomainInfo:      expandOutboundConnectionDomainInfo(d.Get("local_domain_info").([]any)),
		RemoteDomainInfo:     expandOutboundConnectionDomainInfo(d.Get("remote_domain_info").([]any)),
	}

	output, err := conn.CreateOutboundConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Outbound Connection (%s): %s", connectionAlias, err)
	}

	d.SetId(aws.ToString(output.ConnectionId))

	if _, err := waitOutboundConnectionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Outbound Connection (%s) create: %s", d.Id(), err)
	}

	if d.Get("accept_connection").(bool) {
		input := &opensearch.AcceptInboundConnectionInput{
			ConnectionId: aws.String(d.Id()),
		}

		_, err := conn.AcceptInboundConnection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting OpenSearch Inbound Connection (%s): %s", d.Id(), err)
		}

		if _, err := waitInboundConnectionAccepted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Inbound Connection (%s) accept: %s", d.Id(), err)
		}
	}

	return append(diags, resourceOutboundConnectionRead(ctx, d, meta)...)
}

func resourceOutboundConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	connection, err := findOutboundConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Outbound Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
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

	return diags
}

func resourceOutboundConnectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Outbound Connection: %s", d.Id())
	_, err := conn.DeleteOutboundConnection(ctx, &opensearch.DeleteOutboundConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Outbound Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitOutboundConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Outbound Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findOutboundConnectionByID(ctx context.Context, conn *opensearch.Client, id string) (*awstypes.OutboundConnection, error) {
	input := &opensearch.DescribeOutboundConnectionsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("connection-id"),
				Values: []string{id},
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

	if status := output.ConnectionStatus.StatusCode; status == awstypes.OutboundConnectionStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, err
}

func findOutboundConnection(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribeOutboundConnectionsInput) (*awstypes.OutboundConnection, error) {
	output, err := findOutboundConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOutboundConnections(ctx context.Context, conn *opensearch.Client, input *opensearch.DescribeOutboundConnectionsInput) ([]awstypes.OutboundConnection, error) {
	var output []awstypes.OutboundConnection

	pages := opensearch.NewDescribeOutboundConnectionsPaginator(conn, input)
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

func statusOutboundConnection(ctx context.Context, conn *opensearch.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findOutboundConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConnectionStatus.StatusCode), nil
	}
}

func waitOutboundConnectionCreated(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*awstypes.OutboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.OutboundConnectionStatusCodeValidating, awstypes.OutboundConnectionStatusCodeProvisioning),
		Target: enum.Slice(
			awstypes.OutboundConnectionStatusCodePendingAcceptance,
			awstypes.OutboundConnectionStatusCodeActive,
			awstypes.OutboundConnectionStatusCodeApproved,
			awstypes.OutboundConnectionStatusCodeRejected,
			awstypes.OutboundConnectionStatusCodeValidationFailed,
		),
		Refresh: statusOutboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.OutboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func waitOutboundConnectionDeleted(ctx context.Context, conn *opensearch.Client, id string, timeout time.Duration) (*awstypes.OutboundConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.OutboundConnectionStatusCodeActive,
			awstypes.OutboundConnectionStatusCodePendingAcceptance,
			awstypes.OutboundConnectionStatusCodeDeleting,
			awstypes.OutboundConnectionStatusCodeRejecting,
		),
		Target:  []string{},
		Refresh: statusOutboundConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.OutboundConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ConnectionStatus.Message)))

		return output, err
	}

	return nil, err
}

func expandOutboundConnectionDomainInfo(vOptions []any) *awstypes.DomainInformationContainer {
	if len(vOptions) == 0 || vOptions[0] == nil {
		return nil
	}

	mOptions := vOptions[0].(map[string]any)

	return &awstypes.DomainInformationContainer{
		AWSDomainInformation: &awstypes.AWSDomainInformation{
			DomainName: aws.String(mOptions[names.AttrDomainName].(string)),
			OwnerId:    aws.String(mOptions[names.AttrOwnerID].(string)),
			Region:     aws.String(mOptions[names.AttrRegion].(string)),
		},
	}
}

func flattenOutboundConnectionDomainInfo(domainInfo *awstypes.DomainInformationContainer) []any {
	if domainInfo == nil || domainInfo.AWSDomainInformation == nil {
		return nil
	}
	return []any{map[string]any{
		names.AttrOwnerID:    aws.ToString(domainInfo.AWSDomainInformation.OwnerId),
		names.AttrDomainName: aws.ToString(domainInfo.AWSDomainInformation.DomainName),
		names.AttrRegion:     aws.ToString(domainInfo.AWSDomainInformation.Region),
	}}
}

func expandOutboundConnectionConnectionProperties(cProperties []any) *awstypes.ConnectionProperties {
	if len(cProperties) == 0 || cProperties[0] == nil {
		return nil
	}

	mOptions := cProperties[0].(map[string]any)

	return &awstypes.ConnectionProperties{
		CrossClusterSearch: expandOutboundConnectionCrossClusterSearchConnectionProperties(mOptions["cross_cluster_search"].([]any)),
	}
}

func flattenOutboundConnectionConnectionProperties(cProperties *awstypes.ConnectionProperties) []any {
	if cProperties == nil {
		return nil
	}
	return []any{map[string]any{
		"cross_cluster_search": flattenOutboundConnectionCrossClusterSearchConnectionProperties(cProperties.CrossClusterSearch),
		names.AttrEndpoint:     aws.ToString(cProperties.Endpoint),
	}}
}

func expandOutboundConnectionCrossClusterSearchConnectionProperties(cProperties []any) *awstypes.CrossClusterSearchConnectionProperties {
	if len(cProperties) == 0 || cProperties[0] == nil {
		return nil
	}

	mOptions := cProperties[0].(map[string]any)

	return &awstypes.CrossClusterSearchConnectionProperties{
		SkipUnavailable: awstypes.SkipUnavailableStatus(mOptions["skip_unavailable"].(string)),
	}
}

func flattenOutboundConnectionCrossClusterSearchConnectionProperties(cProperties *awstypes.CrossClusterSearchConnectionProperties) []any {
	if cProperties == nil {
		return nil
	}
	return []any{map[string]any{
		"skip_unavailable": cProperties.SkipUnavailable,
	}}
}
