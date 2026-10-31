// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_db_proxy_policy_attachment", name="DB Proxy Policy Attachment")
func resourceProxyPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceProxyPolicyAttachmentRead,
		UpdateWithoutTimeout: resourceProxyPolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourceProxyPolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProxyPolicyAttachmentImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"db_proxy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"policy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"policy_document": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceProxyPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	iamConn := meta.(*conns.AWSClient).IAMClient(ctx)

	dbProxyName := d.Get("db_proxy_name").(string)
	policyName := d.Get("policy_name").(string)
	policyDocument := d.Get("policy_document").(string)
	id := proxyPolicyAttachmentCreateResourceID(dbProxyName, policyName)

	roleName, err := findDBProxyRoleName(ctx, conn, dbProxyName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s) role: %s", dbProxyName, err)
	}

	policy, err := structure.NormalizeJsonString(policyDocument)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyName, err)
	}

	input := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(policyName),
		RoleName:       aws.String(roleName),
	}

	_, err = iamConn.PutRolePolicy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Proxy Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	// Wait for eventual consistency.
	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func(ctx context.Context) (any, error) {
		return findRolePolicyForProxy(ctx, iamConn, roleName, policyName)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy Policy Attachment (%s) create: %s", id, err)
	}

	return append(diags, resourceProxyPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceProxyPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	iamConn := meta.(*conns.AWSClient).IAMClient(ctx)

	dbProxyName, policyName, err := proxyPolicyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	roleName, err := findDBProxyRoleName(ctx, conn, dbProxyName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Policy Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s) role: %s", dbProxyName, err)
	}

	policyDocument, err := findRolePolicyForProxy(ctx, iamConn, roleName, policyName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy Policy Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy Policy Attachment (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy_document").(string), policyDocument)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("policy_name", policyName)
	d.Set("policy_document", policyToSet)

	return diags
}

func resourceProxyPolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	iamConn := meta.(*conns.AWSClient).IAMClient(ctx)

	dbProxyName, policyName, err := proxyPolicyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	roleName, err := findDBProxyRoleName(ctx, conn, dbProxyName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s) role: %s", dbProxyName, err)
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyName, err)
	}

	input := &iam.PutRolePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(policyName),
		RoleName:       aws.String(roleName),
	}

	_, err = iamConn.PutRolePolicy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy Policy Attachment (%s): %s", d.Id(), err)
	}

	return append(diags, resourceProxyPolicyAttachmentRead(ctx, d, meta)...)
}

func resourceProxyPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	iamConn := meta.(*conns.AWSClient).IAMClient(ctx)

	dbProxyName, policyName, err := proxyPolicyAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	roleName, err := findDBProxyRoleName(ctx, conn, dbProxyName)

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s) role: %s", dbProxyName, err)
	}

	log.Printf("[DEBUG] Deleting RDS DB Proxy Policy Attachment: %s", d.Id())
	_, err = iamConn.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	})

	if errs.IsA[*awsiam.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Proxy Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceProxyPolicyAttachmentImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	dbProxyName, policyName, err := proxyPolicyAttachmentParseResourceID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("db_proxy_name", dbProxyName)
	d.Set("policy_name", policyName)

	return []*schema.ResourceData{d}, nil
}

const proxyPolicyAttachmentResourceIDSeparator = "/"

func proxyPolicyAttachmentCreateResourceID(dbProxyName, policyName string) string {
	return dbProxyName + proxyPolicyAttachmentResourceIDSeparator + policyName
}

func proxyPolicyAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, proxyPolicyAttachmentResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected db_proxy_name%[2]spolicy_name", id, proxyPolicyAttachmentResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

// findDBProxyRoleName looks up the proxy and extracts the IAM role name from its RoleArn.
func findDBProxyRoleName(ctx context.Context, conn *rds.Client, dbProxyName string) (string, error) {
	dbProxy, err := findDBProxyByName(ctx, conn, dbProxyName)
	if err != nil {
		return "", err
	}

	roleARN := aws.ToString(dbProxy.RoleArn)
	parsed, err := arn.Parse(roleARN)
	if err != nil {
		return "", fmt.Errorf("parsing role ARN (%s): %w", roleARN, err)
	}

	// ARN resource format: "role/role-name" or "role/path/role-name"
	resource := parsed.Resource
	if !strings.HasPrefix(resource, "role/") {
		return "", fmt.Errorf("unexpected IAM role ARN resource format: %s", resource)
	}

	// Extract role name (last segment after "role/")
	rolePath := strings.TrimPrefix(resource, "role/")
	parts := strings.Split(rolePath, "/")
	roleName := parts[len(parts)-1]

	return roleName, nil
}

// findRolePolicyForProxy retrieves an inline role policy document.
func findRolePolicyForProxy(ctx context.Context, conn *iam.Client, roleName, policyName string) (string, error) {
	input := &iam.GetRolePolicyInput{
		PolicyName: aws.String(policyName),
		RoleName:   aws.String(roleName),
	}

	output, err := conn.GetRolePolicy(ctx, input)

	if errs.IsA[*awsiam.NoSuchEntityException](err) {
		return "", &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.PolicyDocument == nil {
		return "", tfresource.NewEmptyResultError()
	}

	policyDocument, err := url.QueryUnescape(aws.ToString(output.PolicyDocument))
	if err != nil {
		return "", err
	}

	return policyDocument, nil
}
