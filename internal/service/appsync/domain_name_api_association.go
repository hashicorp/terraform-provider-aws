// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package appsync

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
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

func resourceDomainNameAPIAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(domainName),
	}

	output, err := conn.AssociateApi(ctx, input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, domainName)
	}

	d.SetId(aws.ToString(output.ApiAssociation.DomainName))

	if _, err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return smerr.AppendEnrich(ctx, diags, resourceDomainNameAPIAssociationRead(ctx, d, meta))
}

func resourceDomainNameAPIAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	association, err := findDomainNameAPIAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		smerr.AppendOne(ctx, diags, sdkdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	d.Set("api_id", association.ApiId)
	d.Set(names.AttrDomainName, association.DomainName)

	return diags
}

func resourceDomainNameAPIAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	input := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Id()),
	}

	_, err := conn.AssociateApi(ctx, input)

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	if _, err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return smerr.AppendEnrich(ctx, diags, resourceDomainNameAPIAssociationRead(ctx, d, meta))
}

func resourceDomainNameAPIAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	log.Printf("[INFO] Deleting Appsync Domain Name API Association: %s", d.Id())
	input := appsync.DisassociateApiInput{
		DomainName: aws.String(d.Id()),
	}
	_, err := conn.DisassociateApi(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	if _, err := waitDomainNameAPIDisassociation(ctx, conn, d.Id()); err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return diags
}

func findDomainNameAPIAssociationByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) {
	input := &appsync.GetApiAssociationInput{
		DomainName: aws.String(id),
	}

	output, err := conn.GetApiAssociation(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.ApiAssociation == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.ApiAssociation, nil
}

func statusDomainNameAPIAssociation(conn *appsync.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDomainNameAPIAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
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
		Refresh: statusDomainNameAPIAssociation(conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.DeploymentDetail)))
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitDomainNameAPIDisassociation(ctx context.Context, conn *appsync.Client, id string) (*awstypes.ApiAssociation, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusProcessing),
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApiAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.DeploymentDetail)))
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}
