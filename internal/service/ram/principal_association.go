// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ram_principal_association")
func ResourcePrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrincipalAssociationCreate,
		ReadWithoutTimeout:   resourcePrincipalAssociationRead,
		DeleteWithoutTimeout: resourcePrincipalAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidAccountID,
					verify.ValidARN,
				),
			},
		},
	}
}

func resourcePrincipalAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareArn := d.Get("resource_share_arn").(string)
	principal := d.Get("principal").(string)

	request := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(id.UniqueId()),
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Create RAM principal association request:", request)
	_, err := conn.AssociateResourceShareWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating principal with RAM resource share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareArn, principal))

	// AWS Account ID Principals need to be accepted to become ASSOCIATED
	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		return append(diags, resourcePrincipalAssociationRead(ctx, d, meta)...)
	}

	if _, err := WaitResourceSharePrincipalAssociated(ctx, conn, resourceShareArn, principal); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM principal association (%s) to become ready: %s", d.Id(), err)
	}

	return append(diags, resourcePrincipalAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareArn, principal, err := PrincipalAssociationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Principal Association, parsing ID (%s): %s", d.Id(), err)
	}

	var association *ram.ResourceShareAssociation

	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		// AWS Account ID Principals need to be accepted to become ASSOCIATED
		association, err = FindResourceSharePrincipalAssociationByShareARNPrincipal(ctx, conn, resourceShareArn, principal)
	} else {
		association, err = WaitResourceSharePrincipalAssociated(ctx, conn, resourceShareArn, principal)
	}

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException) || tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException)) {
		log.Printf("[WARN] No RAM resource share principal association with ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if !d.IsNewResource() && (association == nil || aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusDisassociated) {
		log.Printf("[WARN] RAM resource share principal association with ARN (%s) found, but empty or disassociated - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociating {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) Principal Association (%s), status not associating or associated: %s", resourceShareArn, principal, aws.StringValue(association.Status))
	}

	d.Set("resource_share_arn", resourceShareArn)
	d.Set("principal", principal)

	return diags
}

func resourcePrincipalAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareArn, principal, err := PrincipalAssociationParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Principal Association (%s): %s", d.Id(), err)
	}

	request := &ram.DisassociateResourceShareInput{
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Delete RAM principal association request:", request)
	_, err = conn.DisassociateResourceShareWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Principal Association (%s): %s", d.Id(), err)
	}

	if _, err := WaitResourceSharePrincipalDisassociated(ctx, conn, resourceShareArn, principal); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Principal Association (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func PrincipalAssociationParseID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,PRINCIPAL", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}
