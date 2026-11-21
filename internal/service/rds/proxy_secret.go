// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy_secret", name="DB Proxy Secret")
func resourceProxySecret() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxySecretCreate,
		ReadWithoutTimeout:   resourceProxySecretRead,
		UpdateWithoutTimeout: resourceProxySecretUpdate,
		DeleteWithoutTimeout: resourceProxySecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProxySecretImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"auth_scheme": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(types.AuthSchemeSecrets),
				ValidateDiagFunc: enum.Validate[types.AuthScheme](),
			},
			"client_password_auth_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ClientPasswordAuthType](),
			},
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"iam_auth": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(types.IAMAuthModeDisabled),
				ValidateDiagFunc: enum.Validate[types.IAMAuthMode](),
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrUsername: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceProxySecretCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	secretARN := d.Get("secret_arn").(string)
	id := proxySecretCreateResourceID(dbProxyName, secretARN)

	dbProxy, err := findDBProxyByName(ctx, conn, dbProxyName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", dbProxyName, err)
	}

	// Check if the secret is already associated.
	for _, auth := range dbProxy.Auth {
		if aws.ToString(auth.SecretArn) == secretARN {
			return sdkdiag.AppendErrorf(diags, "RDS DB Proxy Secret (%s) already exists", id)
		}
	}

	// Build the new auth list: existing auths + the new secret.
	newAuth := make([]types.UserAuthConfig, 0, len(dbProxy.Auth)+1)
	for _, auth := range dbProxy.Auth {
		newAuth = append(newAuth, types.UserAuthConfig{
			AuthScheme:             auth.AuthScheme,
			ClientPasswordAuthType: auth.ClientPasswordAuthType,
			Description:            auth.Description,
			IAMAuth:                auth.IAMAuth,
			SecretArn:              auth.SecretArn,
			UserName:               auth.UserName,
		})
	}

	newAuthEntry := types.UserAuthConfig{
		SecretArn: aws.String(secretARN),
	}
	if v, ok := d.GetOk("auth_scheme"); ok {
		newAuthEntry.AuthScheme = types.AuthScheme(v.(string))
	}
	if v, ok := d.GetOk("client_password_auth_type"); ok {
		newAuthEntry.ClientPasswordAuthType = types.ClientPasswordAuthType(v.(string))
	}
	if v, ok := d.GetOk(names.AttrDescription); ok {
		newAuthEntry.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("iam_auth"); ok {
		newAuthEntry.IAMAuth = types.IAMAuthMode(v.(string))
	}
	if v, ok := d.GetOk(names.AttrUsername); ok {
		newAuthEntry.UserName = aws.String(v.(string))
	}
	newAuth = append(newAuth, newAuthEntry)

	input := &rds.ModifyDBProxyInput{
		DBProxyName: aws.String(dbProxyName),
		Auth:        newAuth,
	}

	_, err = conn.ModifyDBProxy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Proxy Secret (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDBProxyUpdated(ctx, conn, dbProxyName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", dbProxyName, err)
	}

	return append(diags, resourceProxySecretRead(ctx, d, meta)...)
}

func resourceProxySecretRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, secretARN, err := proxySecretParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbProxy, err := findDBProxyByName(ctx, conn, dbProxyName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Secret %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Secret (%s): %s", d.Id(), err)
	}

	// Find the matching auth entry by secret ARN.
	var matchedAuth *types.UserAuthConfigInfo
	for _, auth := range dbProxy.Auth {
		if aws.ToString(auth.SecretArn) == secretARN {
			matchedAuth = &auth
			break
		}
	}

	if matchedAuth == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Secret (%s): secret not found in proxy auth configuration", d.Id())
		}
		log.Printf("[WARN] RDS DB Proxy Secret %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("auth_scheme", string(matchedAuth.AuthScheme))
	d.Set("client_password_auth_type", string(matchedAuth.ClientPasswordAuthType))
	d.Set("db_proxy_name", dbProxyName)
	d.Set(names.AttrDescription, aws.ToString(matchedAuth.Description))
	d.Set("iam_auth", string(matchedAuth.IAMAuth))
	d.Set("secret_arn", secretARN)
	d.Set(names.AttrUsername, aws.ToString(matchedAuth.UserName))

	return diags
}

func resourceProxySecretUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, secretARN, err := proxySecretParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbProxy, err := findDBProxyByName(ctx, conn, dbProxyName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", dbProxyName, err)
	}

	// Rebuild auth list, replacing the entry for this secret with updated values.
	newAuth := make([]types.UserAuthConfig, 0, len(dbProxy.Auth))
	for _, auth := range dbProxy.Auth {
		if aws.ToString(auth.SecretArn) == secretARN {
			updatedEntry := types.UserAuthConfig{
				SecretArn: aws.String(secretARN),
			}
			if v, ok := d.GetOk("auth_scheme"); ok {
				updatedEntry.AuthScheme = types.AuthScheme(v.(string))
			}
			if v, ok := d.GetOk("client_password_auth_type"); ok {
				updatedEntry.ClientPasswordAuthType = types.ClientPasswordAuthType(v.(string))
			}
			if v, ok := d.GetOk(names.AttrDescription); ok {
				updatedEntry.Description = aws.String(v.(string))
			}
			if v, ok := d.GetOk("iam_auth"); ok {
				updatedEntry.IAMAuth = types.IAMAuthMode(v.(string))
			}
			if v, ok := d.GetOk(names.AttrUsername); ok {
				updatedEntry.UserName = aws.String(v.(string))
			}
			newAuth = append(newAuth, updatedEntry)
		} else {
			newAuth = append(newAuth, types.UserAuthConfig{
				AuthScheme:             auth.AuthScheme,
				ClientPasswordAuthType: auth.ClientPasswordAuthType,
				Description:            auth.Description,
				IAMAuth:                auth.IAMAuth,
				SecretArn:              auth.SecretArn,
				UserName:               auth.UserName,
			})
		}
	}

	input := &rds.ModifyDBProxyInput{
		DBProxyName: aws.String(dbProxyName),
		Auth:        newAuth,
	}

	_, err = conn.ModifyDBProxy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy Secret (%s): %s", d.Id(), err)
	}

	if _, err := waitDBProxyUpdated(ctx, conn, dbProxyName, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", dbProxyName, err)
	}

	return append(diags, resourceProxySecretRead(ctx, d, meta)...)
}

func resourceProxySecretDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, secretARN, err := proxySecretParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbProxy, err := findDBProxyByName(ctx, conn, dbProxyName)

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", dbProxyName, err)
	}

	// Build auth list without the secret being removed.
	newAuth := make([]types.UserAuthConfig, 0, len(dbProxy.Auth))
	for _, auth := range dbProxy.Auth {
		if aws.ToString(auth.SecretArn) == secretARN {
			continue
		}
		newAuth = append(newAuth, types.UserAuthConfig{
			AuthScheme:             auth.AuthScheme,
			ClientPasswordAuthType: auth.ClientPasswordAuthType,
			Description:            auth.Description,
			IAMAuth:                auth.IAMAuth,
			SecretArn:              auth.SecretArn,
			UserName:               auth.UserName,
		})
	}

	input := &rds.ModifyDBProxyInput{
		DBProxyName: aws.String(dbProxyName),
		Auth:        newAuth,
	}

	log.Printf("[DEBUG] Deleting RDS DB Proxy Secret: %s", d.Id())
	_, err = conn.ModifyDBProxy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Proxy Secret (%s): %s", d.Id(), err)
	}

	if _, err := waitDBProxyUpdated(ctx, conn, dbProxyName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", dbProxyName, err)
	}

	return diags
}

func resourceProxySecretImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	dbProxyName, secretARN, err := proxySecretParseResourceID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("secret_arn", secretARN)

	return []*schema.ResourceData{d}, nil
}

const proxySecretResourceIDSeparator = "/"

func proxySecretCreateResourceID(dbProxyName, secretARN string) string {
	return dbProxyName + proxySecretResourceIDSeparator + secretARN
}

func proxySecretParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, proxySecretResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected db_proxy_name%[2]ssecret_arn", id, proxySecretResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
