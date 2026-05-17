// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_api_gateway_domain_name_share", name="Domain Name Share")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetDomainNameOutput")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain(t)")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
func resourceDomainNameShare() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameShareCreate,
		ReadWithoutTimeout:   resourceDomainNameShareRead,
		UpdateWithoutTimeout: resourceDomainNameShareUpdate,
		DeleteWithoutTimeout: resourceDomainNameShareDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allowed_accounts": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
			"domain_name_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainNameShareCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	if err := putDomainNameShare(ctx, d, meta); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Domain Name Share (%s): %s", d.Get("domain_name_id").(string), err)
	}

	d.SetId(d.Get("domain_name_id").(string))

	return append(diags, resourceDomainNameShareRead(ctx, d, meta)...)
}

func resourceDomainNameShareRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	output, err := findDomainNameShareByDomainNameID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] API Gateway Domain Name Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name Share (%s): %s", d.Id(), err)
	}

	allowedAccounts, err := domainNameShareAllowedAccounts(aws.ToString(output.ManagementPolicy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name Share (%s): %s", d.Id(), err)
	}

	if len(allowedAccounts) == 0 {
		log.Printf("[WARN] API Gateway Domain Name Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("allowed_accounts", allowedAccounts)
	d.Set("domain_name_id", output.DomainNameId)

	return diags
}

func resourceDomainNameShareUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.HasChange("allowed_accounts") {
		if err := putDomainNameShare(ctx, d, meta); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Domain Name Share (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainNameShareRead(ctx, d, meta)...)
}

func resourceDomainNameShareDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.APIGatewayClient(ctx)

	output, err := findDomainNameByID(ctx, conn, d.Id())

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Domain Name Share (%s): %s", d.Id(), err)
	}

	input := apigateway.UpdateDomainNameInput{
		DomainName:   aws.String(aws.ToString(output.DomainName)),
		DomainNameId: aws.String(d.Id()),
		PatchOperations: []types.PatchOperation{
			{
				Op:   types.OpRemove,
				Path: aws.String("/managementPolicy"),
			},
		},
	}

	_, err = conn.UpdateDomainName(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Invalid patch path") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Domain Name Share (%s): %s", d.Id(), err)
	}

	return diags
}

func putDomainNameShare(ctx context.Context, d *schema.ResourceData, meta any) error {
	c := meta.(*conns.AWSClient)
	conn := c.APIGatewayClient(ctx)

	domainNameID := d.Get("domain_name_id").(string)
	output, err := findDomainNameByID(ctx, conn, domainNameID)
	if err != nil {
		return err
	}

	allowedAccounts := flex.ExpandStringValueSet(d.Get("allowed_accounts").(*schema.Set))
	managementPolicy, err := domainNameShareManagementPolicy(ctx, c, aws.ToString(output.DomainName), domainNameID, allowedAccounts)
	if err != nil {
		return err
	}

	input := apigateway.UpdateDomainNameInput{
		DomainName:   aws.String(aws.ToString(output.DomainName)),
		DomainNameId: aws.String(domainNameID),
		PatchOperations: []types.PatchOperation{
			{
				Op:    types.OpReplace,
				Path:  aws.String("/managementPolicy"),
				Value: aws.String(managementPolicy),
			},
		},
	}

	if _, err := conn.UpdateDomainName(ctx, &input); err != nil {
		return err
	}

	return nil
}

func findDomainNameShareByDomainNameID(ctx context.Context, conn *apigateway.Client, domainNameID string) (*apigateway.GetDomainNameOutput, error) {
	output, err := findDomainNameByID(ctx, conn, domainNameID)
	if err != nil {
		return nil, err
	}

	if aws.ToString(output.ManagementPolicy) == "" {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findDomainNameByID(ctx context.Context, conn *apigateway.Client, domainNameID string) (*apigateway.GetDomainNameOutput, error) {
	input := apigateway.GetDomainNamesInput{}

	for {
		output, err := conn.GetDomainNames(ctx, &input)

		if errs.IsA[*types.NotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range output.Items {
			if aws.ToString(v.DomainNameId) == domainNameID {
				return findDomainNameByTwoPartKey(ctx, conn, aws.ToString(v.DomainName), domainNameID)
			}
		}

		if aws.ToString(output.Position) == "" {
			break
		}

		input.Position = output.Position
	}

	return nil, &retry.NotFoundError{LastError: tfresource.NewEmptyResultError()}
}

func domainNameShareManagementPolicy(ctx context.Context, c *conns.AWSClient, domainName, domainNameID string, allowedAccounts []string) (string, error) {
	accounts := slices.Clone(allowedAccounts)
	slices.Sort(accounts)

	principalARNs := make([]string, 0, len(accounts))
	for _, accountID := range accounts {
		principalARNs = append(principalARNs, fmt.Sprintf("arn:%s:iam::%s:root", c.Partition(ctx), accountID))
	}

	document := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{
			map[string]any{
				"Effect": "Allow",
				"Principal": map[string]any{
					"AWS": principalARNs,
				},
				"Action":   "apigateway:CreateAccessAssociation",
				"Resource": domainNameShareResourceARN(ctx, c, domainName, domainNameID),
			},
		},
	}

	b, err := json.Marshal(document)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func domainNameShareAllowedAccounts(policy string) ([]string, error) {
	type principal struct {
		AWS any `json:"AWS"`
	}

	type statement struct {
		Action    any       `json:"Action"`
		Principal principal `json:"Principal"`
	}

	type policyDocument struct {
		Statement []statement `json:"Statement"`
	}

	var document policyDocument
	if err := json.Unmarshal([]byte(policy), &document); err != nil {
		return nil, fmt.Errorf("parsing management policy JSON: %w", err)
	}

	accounts := make(map[string]struct{})

	for _, statement := range document.Statement {
		if !domainNameShareHasAction(statement.Action, "apigateway:CreateAccessAssociation") {
			continue
		}

		for _, principalARN := range domainNameShareStringSlice(statement.Principal.AWS) {
			if accountID, err := domainNameSharePrincipalAccountID(principalARN); err == nil && accountID != "" {
				accounts[accountID] = struct{}{}
			}
		}
	}

	results := make([]string, 0, len(accounts))
	for accountID := range accounts {
		results = append(results, accountID)
	}

	slices.Sort(results)

	return results, nil
}

func domainNameShareResourceARN(ctx context.Context, c *conns.AWSClient, domainName, domainNameID string) string {
	return fmt.Sprintf("arn:%s:apigateway:%s:%s:/domainnames/%s+%s", c.Partition(ctx), c.Region(ctx), c.AccountID(ctx), domainName, domainNameID)
}

func domainNameShareHasAction(actions any, expected string) bool {
	for _, action := range domainNameShareStringSlice(actions) {
		if strings.EqualFold(action, expected) {
			return true
		}
	}

	return false
}

func domainNameSharePrincipalAccountID(principalARN string) (string, error) {
	if principalARN == "" {
		return "", nil
	}

	if !strings.Contains(principalARN, ":") {
		return principalARN, nil
	}

	parsedARN, err := awsarn.Parse(principalARN)
	if err != nil {
		return "", err
	}

	return parsedARN.AccountID, nil
}

func domainNameShareStringSlice(v any) []string {
	switch v := v.(type) {
	case string:
		if v == "" {
			return nil
		}

		return []string{v}
	case []any:
		result := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok && s != "" {
				result = append(result, s)
			}
		}

		return result
	case []string:
		return v
	default:
		return nil
	}
}
