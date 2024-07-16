// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var functionRegexp = `^(arn:[\w-]+:lambda:)?([a-z]{2}-(?:[a-z]+-){1,2}\d{1}:)?(\d{12}:)?(function:)?([0-9A-Za-z_-]+)(:(\$LATEST|[0-9A-Za-z_-]+))?$`

// @SDKResource("aws_lambda_permission", name="Permission")
func resourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionCreate,
		ReadWithoutTimeout:   resourcePermissionRead,
		DeleteWithoutTimeout: resourcePermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePermissionImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validPermissionAction(),
			},
			"event_source_token": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validPermissionEventSourceToken(),
			},
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validFunctionName(),
			},
			"function_url_auth_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FunctionUrlAuthType](),
			},
			names.AttrPrincipal: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal_org_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"qualifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validQualifier(),
			},
			"source_account": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"source_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"statement_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validPolicyStatementID(),
				ConflictsWith: []string{"statement_id_prefix"},
			},
			"statement_id_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validPolicyStatementID(),
				ConflictsWith: []string{"statement_id"},
			},
		},
	}
}

func resourcePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	statementID := create.Name(d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := &lambda.AddPermissionInput{
		Action:       aws.String(d.Get(names.AttrAction).(string)),
		FunctionName: aws.String(functionName),
		Principal:    aws.String(d.Get(names.AttrPrincipal).(string)),
		StatementId:  aws.String(statementID),
	}

	if v, ok := d.GetOk("event_source_token"); ok {
		input.EventSourceToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("function_url_auth_type"); ok {
		input.FunctionUrlAuthType = awstypes.FunctionUrlAuthType(v.(string))
	}

	if v, ok := d.GetOk("principal_org_id"); ok {
		input.PrincipalOrgID = aws.String(v.(string))
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_account"); ok {
		input.SourceAccount = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_arn"); ok {
		input.SourceArn = aws.String(v.(string))
	}

	// Retry for IAM and Lambda eventual consistency.
	_, err := tfresource.RetryWhenIsOneOf2[*awstypes.ResourceConflictException, *awstypes.ResourceNotFoundException](ctx, lambdaPropagationTimeout,
		func() (interface{}, error) {
			return conn.AddPermission(ctx, input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Lambda Permission (%s/%s): %s", functionName, statementID, err)
	}

	d.SetId(statementID)

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, lambdaPropagationTimeout, func() (interface{}, error) {
		return findPolicyStatementByTwoPartKey(ctx, conn, functionName, d.Id(), d.Get("qualifier").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Permission (%s/%s) not found, removing from state", functionName, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Permission (%s/%s): %s", functionName, d.Id(), err)
	}

	statement := outputRaw.(*PolicyStatement)
	qualifier, _ := getQualifierFromAliasOrVersionARN(statement.Resource)

	d.Set("qualifier", qualifier)

	// Save Lambda function name in the same format
	if strings.HasPrefix(functionName, "arn:"+meta.(*conns.AWSClient).Partition+":lambda:") {
		// Strip qualifier off
		trimmed := strings.TrimSuffix(statement.Resource, ":"+qualifier)
		d.Set("function_name", trimmed)
	} else {
		functionName, err := getFunctionNameFromARN(statement.Resource)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("function_name", functionName)
	}

	d.Set(names.AttrAction, statement.Action)
	// Check if the principal is a cross-account IAM role
	if v, ok := statement.Principal.(map[string]interface{}); ok {
		if _, ok := v["AWS"]; ok {
			d.Set(names.AttrPrincipal, v["AWS"])
		} else {
			d.Set(names.AttrPrincipal, v["Service"])
		}
	} else if v, ok := statement.Principal.(string); ok {
		d.Set(names.AttrPrincipal, v)
	}

	if v, ok := statement.Condition["StringEquals"]; ok {
		d.Set("event_source_token", v["lambda:EventSourceToken"])
		d.Set("function_url_auth_type", v["lambda:FunctionUrlAuthType"])
		d.Set("principal_org_id", v["aws:PrincipalOrgID"])
		d.Set("source_account", v["AWS:SourceAccount"])
	}

	if v, ok := statement.Condition["ArnLike"]; ok {
		d.Set("source_arn", v["AWS:SourceArn"])
	}

	d.Set("statement_id", statement.Sid)
	d.Set("statement_id_prefix", create.NamePrefixFromName(statement.Sid))

	return diags
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := &lambda.RemovePermissionInput{
		FunctionName: aws.String(functionName),
		StatementId:  aws.String(d.Id()),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[INFO] Deleting Lambda Permission: %s", d.Id())
	_, err := conn.RemovePermission(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing Lambda Permission (%s/%s): %s", functionName, d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, lambdaPropagationTimeout, func() (interface{}, error) {
		return findPolicyStatementByTwoPartKey(ctx, conn, functionName, d.Id(), d.Get("qualifier").(string))
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Permission (%s/%s) delete: %s", functionName, d.Id(), err)
	}

	return diags
}

func resourcePermissionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/STATEMENT_ID or FUNCTION_NAME:QUALIFIER/STATEMENT_ID", d.Id())
	}

	functionName := idParts[0]
	statementID := idParts[1]
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	var qualifier string
	if fnParts := strings.Split(functionName, ":"); len(fnParts) == 2 {
		qualifier = fnParts[1]
		input.Qualifier = aws.String(qualifier)
	}

	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	output, err := findFunction(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	d.SetId(statementID)
	d.Set("function_name", output.Configuration.FunctionName)
	if qualifier != "" {
		d.Set("qualifier", qualifier)
	}
	d.Set("statement_id", statementID)

	return []*schema.ResourceData{d}, nil
}

func findPolicy(ctx context.Context, conn *lambda.Client, input *lambda.GetPolicyInput) (*lambda.GetPolicyOutput, error) {
	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findPolicyStatementByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, statementID, qualifier string) (*PolicyStatement, error) {
	input := &lambda.GetPolicyInput{
		FunctionName: aws.String(functionName),
	}
	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	output, err := findPolicy(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	policy := &Policy{}
	err = json.Unmarshal([]byte(aws.ToString(output.Policy)), policy)

	if err != nil {
		return nil, err
	}

	for _, v := range policy.Statement {
		if v.Sid == statementID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest:  statementID,
		LastResponse: policy,
	}
}

func getQualifierFromAliasOrVersionARN(arn string) (string, error) {
	matches := regexache.MustCompile(functionRegexp).FindStringSubmatch(arn)
	if len(matches) < 8 || matches[7] == "" {
		return "", fmt.Errorf("Invalid ARN or otherwise unable to get qualifier from ARN (%s)", arn)
	}

	return matches[7], nil
}

func getFunctionNameFromARN(arn string) (string, error) {
	matches := regexache.MustCompile(functionRegexp).FindStringSubmatch(arn)
	if len(matches) < 6 || matches[5] == "" {
		return "", fmt.Errorf("Invalid ARN or otherwise unable to get qualifier from ARN (%q)",
			arn)
	}
	return matches[5], nil
}

type Policy struct {
	Version   string
	Statement []PolicyStatement
	Id        string
}

type PolicyStatement struct {
	Condition map[string]map[string]string
	Action    string
	Resource  string
	Effect    string
	Principal interface{}
	Sid       string
}
