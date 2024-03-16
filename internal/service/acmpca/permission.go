// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_acmpca_permission")
func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionCreate,
		ReadWithoutTimeout:   resourcePermissionRead,
		DeleteWithoutTimeout: resourcePermissionDelete,

		Schema: map[string]*schema.Schema{
			"actions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.ActionType](),
				},
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"principal": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"acm.amazonaws.com",
				}, false),
			},
			"source_account": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourcePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	caARN := d.Get("certificate_authority_arn").(string)
	principal := d.Get("principal").(string)
	sourceAccount := d.Get("source_account").(string)
	id := PermissionCreateResourceID(caARN, principal, sourceAccount)
	input := &acmpca.CreatePermissionInput{
		Actions:                 expandPermissionActions(d.Get("actions").(*schema.Set)),
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

	log.Printf("[DEBUG] Creating ACM PCA Permission: %+v", input)
	_, err := conn.CreatePermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ACM PCA Permission (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	caARN, principal, sourceAccount, err := PermissionParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Permission (%s): %s", d.Id(), err)
	}

	permission, err := FindPermission(ctx, conn, caARN, principal, sourceAccount)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Permission (%s): %s", d.Id(), err)
	}

	d.Set("actions", flattenPermissionActions(permission.Actions))
	d.Set("certificate_authority_arn", permission.CertificateAuthorityArn)
	d.Set("policy", permission.Policy)
	d.Set("principal", permission.Principal)
	d.Set("source_account", permission.SourceAccount)

	return diags
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	caARN, principal, sourceAccount, err := PermissionParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Permission (%s): %s", d.Id(), err)
	}

	input := &acmpca.DeletePermissionInput{
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

	log.Printf("[DEBUG] Deleting ACM PCA Permission: %s", d.Id())
	_, err = conn.DeletePermission(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Permission (%s): %s", d.Id(), err)
	}

	return diags
}

const permissionIDSeparator = ","

func PermissionCreateResourceID(caARN, principal, sourceAccount string) string {
	parts := []string{caARN, principal, sourceAccount}
	id := strings.Join(parts, permissionIDSeparator)

	return id
}

func PermissionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, permissionIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CertificateAuthorityARN%[2]sPrincipal%[2]sSourceAccount", id, permissionIDSeparator)
}

func expandPermissionActions(s *schema.Set) []awstypes.ActionType {
	actions := make([]awstypes.ActionType, 0)

	for _, a := range s.List() {
		action := awstypes.ActionType(a.(string))
		actions = append(actions, action)
	}
	return actions
}

func flattenPermissionActions(list []awstypes.ActionType) []string {
	if len(list) == 0 {
		return nil
	}

	result := make([]string, 0, len(list))
	for _, a := range list {
		action := string(a)
		result = append(result, action)
	}
	return result
}
