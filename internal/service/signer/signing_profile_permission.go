// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
			"profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 64),
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
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"statement_id_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id"},
				ValidateFunc:  validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
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

	var revisionID string
	output, err := conn.ListProfilePermissions(ctx, &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	})

	if err == nil {
		revisionID = aws.ToString(output.RevisionId)
	} else if !errs.IsA[*types.ResourceNotFoundException](err) {
		return sdkdiag.AppendErrorf(diags, "reading Signer Signing Profile (%s) Permissions: %s", profileName, err)
	}

	statementID := create.Name(d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))
	input := &signer.AddProfilePermissionInput{
		Action:      aws.String(d.Get("action").(string)),
		Principal:   aws.String(d.Get("principal").(string)),
		ProfileName: aws.String(profileName),
		RevisionId:  aws.String(revisionID),
		StatementId: aws.String(statementID),
	}

	if v, ok := d.GetOk("profile_version"); ok {
		input.ProfileVersion = aws.String(v.(string))
	}

	_, err = tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.AddProfilePermission(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*types.ConflictException](err) || errs.IsA[*types.ResourceNotFoundException](err) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Signer Signing Profile (%s) Permission: %s", profileName, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", profileName, statementID))
	d.Set("statement_id", statementID)

	return append(diags, resourceSigningProfilePermissionRead(ctx, d, meta)...)
}

func resourceSigningProfilePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)

	profileName, statementID := d.Get("profile_name").(string), d.Get("statement_id").(string)
	permission, err := findPermissionByTwoPartKey(ctx, conn, profileName, statementID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Signer Signing Profile Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer Signing Profile Permission (%s): %s", d.Id(), err)
	}

	d.Set("action", permission.Action)
	d.Set("principal", permission.Principal)
	d.Set("profile_version", permission.ProfileVersion)
	d.Set("statement_id", permission.StatementId)
	d.Set("statement_id_prefix", create.NamePrefixFromName(aws.ToString(permission.StatementId)))

	return diags
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

func findPermissionByTwoPartKey(ctx context.Context, conn *signer.Client, profileName, statementID string) (*types.Permission, error) {
	input := &signer.ListProfilePermissionsInput{
		ProfileName: aws.String(profileName),
	}

	return findPermission(ctx, conn, input, func(v types.Permission) bool {
		return aws.ToString(v.StatementId) == statementID
	})
}

func findPermission(ctx context.Context, conn *signer.Client, input *signer.ListProfilePermissionsInput, filter tfslices.Predicate[types.Permission]) (*types.Permission, error) {
	output, err := findPermissions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPermissions(ctx context.Context, conn *signer.Client, input *signer.ListProfilePermissionsInput, filter tfslices.Predicate[types.Permission]) ([]types.Permission, error) {
	var permissions []types.Permission

	// No paginator in the AWS SDK for Go v2.
	for {
		output, err := conn.ListProfilePermissions(ctx, input)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range output.Permissions {
			if filter(v) {
				permissions = append(permissions, v)
			}
		}

		if aws.ToString(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return permissions, nil
}
