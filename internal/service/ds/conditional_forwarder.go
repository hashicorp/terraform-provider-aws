// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_directory_service_conditional_forwarder")
func ResourceConditionalForwarder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConditionalForwarderCreate,
		ReadWithoutTimeout:   resourceConditionalForwarderRead,
		UpdateWithoutTimeout: resourceConditionalForwarderUpdate,
		DeleteWithoutTimeout: resourceConditionalForwarderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dns_ips": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"remote_domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// Documentation is incorrect, the API call fails if a trailing period is included
				ValidateFunc: domainValidator,
			},
		},
	}
}

func resourceConditionalForwarderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	dnsIps := flex.ExpandStringList(d.Get("dns_ips").([]interface{}))

	directoryId := d.Get("directory_id").(string)
	domainName := d.Get("remote_domain_name").(string)

	_, err := conn.CreateConditionalForwarderWithContext(ctx, &directoryservice.CreateConditionalForwarderInput{
		DirectoryId:      aws.String(directoryId),
		DnsIpAddrs:       dnsIps,
		RemoteDomainName: aws.String(domainName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Conditional Forwarder: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", directoryId, domainName))

	return diags
}

func resourceConditionalForwarderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId, domainName, err := ParseConditionalForwarderID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	res, err := conn.DescribeConditionalForwardersWithContext(ctx, &directoryservice.DescribeConditionalForwardersInput{
		DirectoryId:       aws.String(directoryId),
		RemoteDomainNames: []*string{aws.String(domainName)},
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
			log.Printf("[WARN] Directory Service Conditional Forwarder (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	if len(res.ConditionalForwarders) == 0 {
		log.Printf("[WARN] Directory Service Conditional Forwarder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	cfd := res.ConditionalForwarders[0]

	d.Set("dns_ips", flex.FlattenStringList(cfd.DnsIpAddrs))
	d.Set("directory_id", directoryId)
	d.Set("remote_domain_name", cfd.RemoteDomainName)

	return diags
}

func resourceConditionalForwarderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId, domainName, err := ParseConditionalForwarderID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	dnsIps := flex.ExpandStringList(d.Get("dns_ips").([]interface{}))

	_, err = conn.UpdateConditionalForwarderWithContext(ctx, &directoryservice.UpdateConditionalForwarderInput{
		DirectoryId:      aws.String(directoryId),
		DnsIpAddrs:       dnsIps,
		RemoteDomainName: aws.String(domainName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConditionalForwarderRead(ctx, d, meta)...)
}

func resourceConditionalForwarderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId, domainName, err := ParseConditionalForwarderID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	_, err = conn.DeleteConditionalForwarderWithContext(ctx, &directoryservice.DeleteConditionalForwarderInput{
		DirectoryId:      aws.String(directoryId),
		RemoteDomainName: aws.String(domainName),
	})

	if err != nil && !tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	return diags
}

func ParseConditionalForwarderID(id string) (directoryId, domainName string, err error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("please make sure ID is in format DIRECTORY_ID:DOMAIN_NAME")
	}

	return parts[0], parts[1], nil
}
