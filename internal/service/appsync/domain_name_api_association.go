// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
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

// @SDKResource("aws_appsync_domain_name_api_association", name="Domain Name API Association")
func resourceDomainNameAPIAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameAPIAssociationCreate,
		ReadWithoutTimeout:   resourceDomainNameAPIAssociationRead,
		UpdateWithoutTimeout: resourceDomainNameAPIAssociationUpdate,
		DeleteWithoutTimeout: resourceDomainNameAPIAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainNameAPIAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(domainName),
	}

	output, err := conn.AssociateApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Domain Name API Association (%s): %s", domainName, err)
	}

	d.SetId(aws.ToString(output.ApiAssociation.DomainName))

	if _, err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameAPIAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAPIAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	association, err := findDomainNameAPIAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Appsync Domain Name API Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Domain Name API Association (%s): %s", d.Id(), err)
	}

	d.Set("api_id", association.ApiId)
	d.Set(names.AttrDomainName, association.DomainName)

	return diags
}

func resourceDomainNameAPIAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Id()),
	}

	_, err := conn.AssociateApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Domain Name API Association (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API Association (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameAPIAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAPIAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	log.Printf("[INFO] Deleting Appsync Domain Name API Association: %s", d.Id())
	_, err := conn.DisassociateApi(ctx, &appsync.DisassociateApiInput{
		DomainName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Domain Name API Association (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainNameAPIDisassociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDomainNameAPIAssociationByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) {
	input := &appsync.GetApiAssociationInput{
		DomainName: aws.String(id),
	}

	output, err := conn.GetApiAssociation(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApiAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApiAssociation, nil
}

func statusDomainNameAPIAssociation(ctx context.Context, conn *appsync.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDomainNameAPIAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AssociationStatus), nil
	}
}

func waitDomainNameAPIAssociation(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) { //nolint:unparam
	const (
		domainNameAPIAssociationTimeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusProcessing),
		Target:  enum.Slice(awstypes.AssociationStatusSuccess),
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DeploymentDetail)))
		return output, err
	}

	return nil, err
}

func waitDomainNameAPIDisassociation(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusProcessing),
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DeploymentDetail)))
		return output, err
	}

	return nil, err
}
