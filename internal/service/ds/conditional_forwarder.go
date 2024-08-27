// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_directory_service_conditional_forwarder", name="Conditional Forwarder")
func resourceConditionalForwarder() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID := d.Get("directory_id").(string)
	domainName := d.Get("remote_domain_name").(string)
	id := conditionalForwarderCreateResourceID(directoryID, domainName)
	input := &directoryservice.CreateConditionalForwarderInput{
		DirectoryId:      aws.String(directoryID),
		DnsIpAddrs:       flex.ExpandStringValueList(d.Get("dns_ips").([]interface{})),
		RemoteDomainName: aws.String(domainName),
	}

	_, err := conn.CreateConditionalForwarder(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Conditional Forwarder (%s): %s", id, err)
	}

	d.SetId(id)

	const (
		timeout = 1 * time.Minute
	)
	_, err = tfresource.RetryWhenNotFound(ctx, timeout, func() (interface{}, error) {
		return findConditionalForwarderByTwoPartKey(ctx, conn, directoryID, domainName)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Directory Service Conditional Forwarder (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConditionalForwarderRead(ctx, d, meta)...)
}

func resourceConditionalForwarderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, domainName, err := conditionalForwarderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	cfd, err := findConditionalForwarderByTwoPartKey(ctx, conn, directoryID, domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Conditional Forwarder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	d.Set("directory_id", directoryID)
	d.Set("dns_ips", cfd.DnsIpAddrs)
	d.Set("remote_domain_name", cfd.RemoteDomainName)

	return diags
}

func resourceConditionalForwarderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, domainName, err := conditionalForwarderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &directoryservice.UpdateConditionalForwarderInput{
		DirectoryId:      aws.String(directoryID),
		DnsIpAddrs:       flex.ExpandStringValueList(d.Get("dns_ips").([]interface{})),
		RemoteDomainName: aws.String(domainName),
	}

	_, err = conn.UpdateConditionalForwarder(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConditionalForwarderRead(ctx, d, meta)...)
}

func resourceConditionalForwarderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)

	directoryID, domainName, err := conditionalForwarderParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Directory Conditional Forwarder: %s", d.Id())
	_, err = conn.DeleteConditionalForwarder(ctx, &directoryservice.DeleteConditionalForwarderInput{
		DirectoryId:      aws.String(directoryID),
		RemoteDomainName: aws.String(domainName),
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Conditional Forwarder (%s): %s", d.Id(), err)
	}

	return diags
}

const conditionalForwarderResourceIDSeparator = ":" // nosemgrep:ci.ds-in-const-name,ci.ds-in-var-name

func conditionalForwarderCreateResourceID(directoryID, domainName string) string {
	parts := []string{directoryID, domainName}
	id := strings.Join(parts, conditionalForwarderResourceIDSeparator)

	return id
}

func conditionalForwarderParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, conditionalForwarderResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DIRECTORY_ID%[2]sDOMAIN_NAME", id, conditionalForwarderResourceIDSeparator)
}

func findConditionalForwarder(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeConditionalForwardersInput) (*awstypes.ConditionalForwarder, error) {
	output, err := findConditionalForwarders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConditionalForwarders(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeConditionalForwardersInput) ([]awstypes.ConditionalForwarder, error) {
	output, err := conn.DescribeConditionalForwarders(ctx, input)

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConditionalForwarders, nil
}

func findConditionalForwarderByTwoPartKey(ctx context.Context, conn *directoryservice.Client, directoryID, domainName string) (*awstypes.ConditionalForwarder, error) {
	input := &directoryservice.DescribeConditionalForwardersInput{
		DirectoryId:       aws.String(directoryID),
		RemoteDomainNames: []string{domainName},
	}

	return findConditionalForwarder(ctx, conn, input)
}
