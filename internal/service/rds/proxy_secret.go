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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_db_proxy_secret", name="DB Proxy Secret")
func resourceProxySecret() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxySecretCreate,
		ReadWithoutTimeout:   resourceProxySecretRead,
		DeleteWithoutTimeout: resourceProxySecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProxySecretImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
			AuthScheme:  auth.AuthScheme,
			Description: auth.Description,
			IAMAuth:     auth.IAMAuth,
			SecretArn:   auth.SecretArn,
			UserName:    auth.UserName,
		})
	}
	newAuth = append(newAuth, types.UserAuthConfig{
		AuthScheme: types.AuthSchemeSecrets,
		SecretArn:  aws.String(secretARN),
	})

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

	// Verify the secret ARN is still in the proxy's auth list.
	found := false
	for _, auth := range dbProxy.Auth {
		if aws.ToString(auth.SecretArn) == secretARN {
			found = true
			break
		}
	}

	if !found {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Secret (%s): secret not found in proxy auth configuration", d.Id())
		}
		log.Printf("[WARN] RDS DB Proxy Secret %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("secret_arn", secretARN)

	return diags
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
			AuthScheme:  auth.AuthScheme,
			Description: auth.Description,
			IAMAuth:     auth.IAMAuth,
			SecretArn:   auth.SecretArn,
			UserName:    auth.UserName,
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
