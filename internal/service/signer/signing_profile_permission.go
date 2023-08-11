// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_signer_signing_profile_permission")
func ResourceSigningProfilePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSigningProfilePermissionCreate,
		ReadWithoutTimeout:   resourceSigningProfilePermissionRead,
		DeleteWithoutTimeout: resourceSigningProfilePermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceSigningProfilePermissionImport,
		},

		Schema: map[string]*schema.Schema{
			"profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 64),
			},
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"signer:StartSigningJob",
					"signer:GetSigningProfile",
					"signer:RevokeSignature"},
					false),
			},
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"profile_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(10, 10),
			},
			"statement_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id_prefix"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"statement_id_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
		},
	}
}

func resourceSigningProfilePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	profileName := d.Get("profile_name").(string)

	conns.GlobalMutexKV.Lock(profileName)
	defer conns.GlobalMutexKV.Unlock(profileName)

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	var revisionId string
	getProfilePermissionsOutput, err := conn.ListProfilePermissions(ctx, listProfilePermissionsInput)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			revisionId = ""
		} else {
			return sdkdiag.AppendErrorf(diags, "creating Signer Signing Profile Permission: %s", err)
		}
	} else {
		revisionId = aws.ToString(getProfilePermissionsOutput.RevisionId)
	}

	statementId := create.Name(d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))

	addProfilePermissionInput := &signer.AddProfilePermissionInput{
		Action:      aws.String(d.Get("action").(string)),
		Principal:   aws.String(d.Get("principal").(string)),
		ProfileName: aws.String(profileName),
		RevisionId:  aws.String(revisionId),
		StatementId: aws.String(statementId),
	}

	if v, ok := d.GetOk("profile_version"); ok {
		addProfilePermissionInput.ProfileVersion = aws.String(v.(string))
	}

	// Retry for IAM eventual consistency
	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err := conn.AddProfilePermission(ctx, addProfilePermissionInput)
		var ce *types.ConflictException
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &ce) || errors.As(err, &nfe) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AddProfilePermission(ctx, addProfilePermissionInput)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding new Signer Signing Profile Permission for %q: %s", profileName, err)
	}

	d.Set("statement_id", statementId)
	d.SetId(fmt.Sprintf("%s/%s", profileName, statementId))

	return append(diags, resourceSigningProfilePermissionRead(ctx, d, meta)...)
}

func resourceSigningProfilePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(d.Get("profile_name").(string)),
	}

	var listProfilePermissionsOutput *signer.ListProfilePermissionsOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		// IAM is eventually consistent :/
		var err error
		listProfilePermissionsOutput, err = conn.ListProfilePermissions(ctx, listProfilePermissionsInput)
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		listProfilePermissionsOutput, err = conn.ListProfilePermissions(ctx, listProfilePermissionsInput)
	}

	var nfe *types.ResourceNotFoundException
	if !d.IsNewResource() && errors.As(err, &nfe) {
		log.Printf("[WARN] No Signer Signing Profile Permissions found (%s), removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer Signing Profile Permissions (%s): %s", d.Id(), err)
	}

	statementId := d.Get("statement_id").(string)
	permission := getProfilePermission(listProfilePermissionsOutput.Permissions, statementId)
	if permission == (types.Permission{}) {
		log.Printf("[WARN] No Signer Signing Profile Permission found matching statement id: %s", statementId)
		d.SetId("")
		return diags
	}

	if err := d.Set("action", permission.Action); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile permission action: %s", err)
	}
	if err := d.Set("principal", permission.Principal); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile permission principal: %s", err)
	}
	if err := d.Set("profile_version", permission.ProfileVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile permission profile version: %s", err)
	}
	if err := d.Set("statement_id", permission.StatementId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile permission statement id: %s", err)
	}

	return diags
}

func getProfilePermission(permissions []types.Permission, statementId string) types.Permission {
	for _, permission := range permissions {
		if permission == (types.Permission{}) {
			continue
		}

		if aws.ToString(permission.StatementId) == statementId {
			return permission
		}
	}

	return types.Permission{}
}

func resourceSigningProfilePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	profileName := d.Get("profile_name").(string)

	conns.GlobalMutexKV.Lock(profileName)
	defer conns.GlobalMutexKV.Unlock(profileName)

	listProfilePermissionsInput := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	listProfilePermissionsOutput, err := conn.ListProfilePermissions(ctx, listProfilePermissionsInput)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			log.Printf("[WARN] No Signer Signing Profile Permission found for: %v", listProfilePermissionsInput)
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Signer Signing Profile Permission (%s): %s", d.Id(), err)
	}

	revisionId := aws.ToString(listProfilePermissionsOutput.RevisionId)

	statementId := d.Get("statement_id").(string)
	permission := getProfilePermission(listProfilePermissionsOutput.Permissions, statementId)
	if permission == (types.Permission{}) {
		log.Printf("[WARN] No Signer Signing Profile Permission found matching statement id: %s", statementId)
		return diags
	}

	removeProfilePermissionInput := &signer.RemoveProfilePermissionInput{
		ProfileName: aws.String(profileName),
		RevisionId:  aws.String(revisionId),
		StatementId: permission.StatementId,
	}

	_, err = conn.RemoveProfilePermission(ctx, removeProfilePermissionInput)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			log.Printf("[WARN] No Signer Signing Profile Permission found: %v", removeProfilePermissionInput)
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Signer Signing Profile Permission (%s): %s", d.Id(), err)
	}

	params := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	resp, err := conn.ListProfilePermissions(ctx, params)

	var nfe *types.ResourceNotFoundException
	if errors.As(err, &nfe) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Signer Signing Profile Permissions: %s", err)
	}

	if len(resp.Permissions) > 0 {
		permission := getProfilePermission(resp.Permissions, statementId)
		if permission != (types.Permission{}) {
			return sdkdiag.AppendErrorf(diags, "to delete Signer singing profile permission with ID %q", statementId)
		}
	}

	log.Printf("[DEBUG] Signer Signing Profile Permission with ID %q removed", statementId)

	return diags
}

func resourceSigningProfilePermissionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected PROFILE_NAME/STATEMENT_ID", d.Id())
	}

	profileName := idParts[0]
	statementId := idParts[1]

	d.Set("profile_name", profileName)
	d.Set("statement_id", statementId)
	d.SetId(statementId)
	return []*schema.ResourceData{d}, nil
}
