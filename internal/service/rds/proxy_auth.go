// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"log"
	"strings"
	"time"
)

// @SDKResource("aws_db_proxy_auth_item", name="DB Proxy Auth Item")
func resourceProxyAuthItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyAuthItemCreate,
		ReadWithoutTimeout:   resourceProxyAuthItemRead,
		DeleteWithoutTimeout: resourceProxyAuthItemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"auth_scheme": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.AuthScheme](),
			},
			"client_password_auth_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.ClientPasswordAuthType](),
			},
			"description": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"iam_auth": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.IAMAuthMode](),
			},
			"secret_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"username": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceProxyAuthItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	secretArn := d.Get("secret_arn").(string)

	existingAuthItems, err := findDBProxyAuthItems(ctx, conn, &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(dbProxyName),
	}, func(v *types.UserAuthConfig) bool {
		return true
	})
	if err != nil {
		return diags
	}

	var authInfo types.UserAuthConfig

	if v, ok := d.GetOk("auth_scheme"); ok && v != "" {
		authInfo.AuthScheme = types.AuthScheme(v.(string))
	}

	if v, ok := d.GetOk("client_password_auth_type"); ok && v != "" {
		authInfo.ClientPasswordAuthType = types.ClientPasswordAuthType(v.(string))
	}

	if v, ok := d.GetOk("description"); ok && v != "" {
		authInfo.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_auth"); ok && v != "" {
		authInfo.IAMAuth = types.IAMAuthMode(v.(string))
	}

	if v, ok := d.GetOk("secret_arn"); ok && v != "" {
		authInfo.SecretArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("username"); ok && v != "" {
		authInfo.UserName = aws.String(v.(string))
	}

	newAuthItems := append(existingAuthItems, authInfo)

	input := &rds.ModifyDBProxyInput{
		DBProxyName: aws.String(dbProxyName),
		Auth:        newAuthItems,
	}

	const (
		timeout = 5 * time.Minute
	)
	_, err = tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidDBInstanceStateFault](ctx, timeout,
		func() (interface{}, error) {
			return conn.ModifyDBProxy(ctx, input)
		},
		"CREATING")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering RDS DB Proxy Auth Item (%s): %s", dbProxyName, err)
	}

	d.SetId(proxyAuthItemCreateResourceID(dbProxyName, secretArn))

	if _, err := waitDBProxyUpdated(ctx, conn, dbProxyName, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", dbProxyName, err)
	}

	return append(diags, resourceProxyAuthItemRead(ctx, d, meta)...)
}

func resourceProxyAuthItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, secretArn, err := proxyAuthItemParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dbAuthItem, err := findDBProxyAuthItemByArn(ctx, conn, dbProxyName, secretArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Auth Item ARN (%s) not found, removing from state", secretArn)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Auth Item (%s): %s", d.Id(), err)
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("iam_auth", dbAuthItem.IAMAuth)
	d.Set("auth_scheme", dbAuthItem.AuthScheme)
	d.Set("client_password_auth_type", dbAuthItem.ClientPasswordAuthType)
	d.Set("description", dbAuthItem.Description)
	d.Set("username", dbAuthItem.UserName)
	d.Set("secret_arn", dbAuthItem.SecretArn)

	return diags

}

func resourceProxyAuthItemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxyName, secretArn, err := proxyAuthItemParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	existingAuthItems, err := findDBProxyAuthItems(ctx, conn, &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(dbProxyName),
	}, func(v *types.UserAuthConfig) bool {
		return true
	})
	if err != nil {
		return diags
	}

	newAuthItems := func(authItems []types.UserAuthConfig, secretArn string) []types.UserAuthConfig {
		for i, item := range authItems {
			if aws.ToString(item.SecretArn) == secretArn {
				return append(authItems[:i], authItems[i+1:]...)
			}
		}
		return authItems
	}(existingAuthItems, secretArn)

	input := &rds.ModifyDBProxyInput{
		DBProxyName: aws.String(dbProxyName),
		Auth:        newAuthItems,
	}

	const (
		timeout = 5 * time.Minute
	)
	_, err = tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidDBInstanceStateFault](ctx, timeout,
		func() (interface{}, error) {
			return conn.ModifyDBProxy(ctx, input)
		},
		"CREATING")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering RDS DB Proxy Auth Item (%s): %s", dbProxyName, err)
	}

	d.SetId(proxyAuthItemCreateResourceID(dbProxyName, secretArn))

	if _, err := waitDBProxyUpdated(ctx, conn, dbProxyName, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", dbProxyName, err)
	}

	return append(diags, resourceProxyAuthItemRead(ctx, d, meta)...)
}

const proxyAuthItemResourceIDSeparator = "/"

func proxyAuthItemCreateResourceID(dbProxyName, secretArn string) string {
	parts := []string{dbProxyName, secretArn}
	id := strings.Join(parts, proxyAuthItemResourceIDSeparator)

	return id
}

func proxyAuthItemParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, proxyAuthItemResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBPROXYNAME%[2]sSECRETARN", id, proxyAuthItemResourceIDSeparator)
}

func findDBProxyAuthItemByArn(ctx context.Context, conn *rds.Client, dbProxyName, secretArn string) (*types.UserAuthConfig, error) {
	input := &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(dbProxyName),
	}

	return findDBProxyAuthItem(ctx, conn, input, func(v *types.UserAuthConfig) bool {
		return aws.ToString(v.SecretArn) == secretArn
	})
}

func findDBProxyAuthItem(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxiesInput, filter tfslices.Predicate[*types.UserAuthConfig]) (*types.UserAuthConfig, error) {
	output, err := findDBProxyAuthItems(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBProxyAuthItems(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxiesInput, filter tfslices.Predicate[*types.UserAuthConfig]) ([]types.UserAuthConfig, error) {
	var output []types.UserAuthConfig

	pages := rds.NewDescribeDBProxiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBProxyNotFoundFault](err) || errs.IsA[*types.DBProxyTargetGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBProxies {
			for _, item := range v.Auth {
				authConfig := types.UserAuthConfig{
					AuthScheme:             item.AuthScheme,
					ClientPasswordAuthType: item.ClientPasswordAuthType,
					Description:            item.Description,
					IAMAuth:                item.IAMAuth,
					SecretArn:              item.SecretArn,
					UserName:               item.UserName,
				}

				if filter(&authConfig) {
					output = append(output, authConfig)
				}
			}
		}
	}

	return output, nil
}
