// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	permissionResourceIDPartCount = 3
)

// @SDKResource("aws_acmpca_permission", name="Permission")
func resourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionCreate,
		ReadWithoutTimeout:   resourcePermissionRead,
		DeleteWithoutTimeout: resourcePermissionDelete,

		Schema: map[string]*schema.Schema{
			names.AttrActions: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.ActionType](),
				},
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrincipal: {
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
	principal := d.Get(names.AttrPrincipal).(string)
	sourceAccount := d.Get("source_account").(string)
	id := errs.Must(flex.FlattenResourceId([]string{caARN, principal, sourceAccount}, permissionResourceIDPartCount, true))
	input := &acmpca.CreatePermissionInput{
		Actions:                 expandPermissionActions(d.Get(names.AttrActions).(*schema.Set)),
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

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

	parts, err := flex.ExpandResourceId(d.Id(), permissionResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	caARN, principal, sourceAccount := parts[0], parts[1], parts[2]
	permission, err := findPermissionByThreePartKey(ctx, conn, caARN, principal, sourceAccount)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Permission (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrActions, flattenPermissionActions(permission.Actions))
	d.Set("certificate_authority_arn", permission.CertificateAuthorityArn)
	d.Set(names.AttrPolicy, permission.Policy)
	d.Set(names.AttrPrincipal, permission.Principal)
	d.Set("source_account", permission.SourceAccount)

	return diags
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), permissionResourceIDPartCount, true)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	caARN, principal, sourceAccount := parts[0], parts[1], parts[2]
	input := &acmpca.DeletePermissionInput{
		CertificateAuthorityArn: aws.String(caARN),
		Principal:               aws.String(principal),
	}

	if sourceAccount != "" {
		input.SourceAccount = aws.String(sourceAccount)
	}

	log.Printf("[DEBUG] Deleting ACM PCA Permission: %s", d.Id())
	_, err = conn.DeletePermission(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Permission (%s): %s", d.Id(), err)
	}

	return diags
}

func findPermissionByThreePartKey(ctx context.Context, conn *acmpca.Client, certificateAuthorityARN, principal, sourceAccount string) (*types.Permission, error) {
	input := &acmpca.ListPermissionsInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityARN),
	}

	return findPermission(ctx, conn, input, func(v *types.Permission) bool {
		return aws.ToString(v.Principal) == principal && (sourceAccount == "" || aws.ToString(v.SourceAccount) == sourceAccount)
	})
}

func findPermission(ctx context.Context, conn *acmpca.Client, input *acmpca.ListPermissionsInput, filter tfslices.Predicate[*types.Permission]) (*types.Permission, error) {
	output, err := findPermissions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPermissions(ctx context.Context, conn *acmpca.Client, input *acmpca.ListPermissionsInput, filter tfslices.Predicate[*types.Permission]) ([]types.Permission, error) {
	var output []types.Permission

	pages := acmpca.NewListPermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*types.InvalidStateException](err, "The certificate authority is in the DELETED state") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Permissions {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandPermissionActions(s *schema.Set) []types.ActionType {
	actions := make([]types.ActionType, 0)

	for _, a := range s.List() {
		action := types.ActionType(a.(string))
		actions = append(actions, action)
	}
	return actions
}

func flattenPermissionActions(list []types.ActionType) []string {
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
