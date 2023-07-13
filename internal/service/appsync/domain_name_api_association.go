// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_appsync_domain_name_api_association")
func ResourceDomainNameAPIAssociation() *schema.Resource {
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainNameAPIAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	resp, err := conn.AssociateApiWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Domain Name API Association: %s", err)
	}

	d.SetId(aws.StringValue(resp.ApiAssociation.DomainName))

	if err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API (%s) Association: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameAPIAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAPIAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	association, err := FindDomainNameAPIAssociationByID(ctx, conn, d.Id())
	if association == nil && !d.IsNewResource() {
		log.Printf("[WARN] Appsync Domain Name API Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Appsync Domain Name API Association %q: %s", d.Id(), err)
	}

	d.Set("domain_name", association.DomainName)
	d.Set("api_id", association.ApiId)

	return diags
}

func resourceDomainNameAPIAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	_, err := conn.AssociateApiWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Domain Name API Association: %s", err)
	}

	if err := waitDomainNameAPIAssociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API (%s) Association: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameAPIAssociationRead(ctx, d, meta)...)
}

func resourceDomainNameAPIAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	input := &appsync.DisassociateApiInput{
		DomainName: aws.String(d.Id()),
	}
	_, err := conn.DisassociateApiWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Domain Name API Association: %s", err)
	}

	if err := waitDomainNameAPIDisassociation(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Appsync Domain Name API (%s) Disassociation: %s", d.Id(), err)
	}

	return diags
}
