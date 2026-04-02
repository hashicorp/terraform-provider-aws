// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var functionRegexp = `^(arn:[\w-]+:lambda:)?(` + inttypes.CanonicalRegionPatternNoAnchors + `:)?(\d{12}:)?(function:)?([0-9A-Za-z_-]+)(:(\$LATEST|[0-9A-Za-z_-]+))?$`

// @SDKResource("aws_lambda_permission", name="Permission")
// @IdentityAttribute("function_name")
// @IdentityAttribute("statement_id")
// @IdentityAttribute("qualifier", optional="true")
// @ImportIDHandler("permissionImportID")
// @Testing(preIdentityVersion="6.9.0")
// @Testing(existsType="github.com/hashicorp/terraform-provider-aws/internal/service/lambda;tflambda;tflambda.PolicyStatement")
// @Testing(importStateIdFunc="testAccPermissionImportStateIDFunc")
func resourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionCreate,
		ReadWithoutTimeout:   resourcePermissionRead,
		DeleteWithoutTimeout: resourcePermissionDelete,

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
			"invoked_via_function_url": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
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

func resourcePermissionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	statementID := create.Name(ctx, d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := lambda.AddPermissionInput{
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

	if v, ok := d.GetOk("invoked_via_function_url"); ok {
		input.InvokedViaFunctionUrl = aws.Bool(v.(bool))
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
	_, err := tfresource.RetryWhenIsOneOf2[any, *awstypes.ResourceConflictException, *awstypes.ResourceNotFoundException](ctx, lambdaPropagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.AddPermission(ctx, &input)
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Lambda Permission (%s/%s): %s", functionName, statementID, err)
	}

	d.SetId(statementID)

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	statement, err := tfresource.RetryWhenNewResourceNotFound(ctx, lambdaPropagationTimeout, func(ctx context.Context) (*policyStatement, error) {
		return findPolicyStatementByTwoPartKey(ctx, conn, functionName, d.Id(), d.Get("qualifier").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Lambda Permission (%s/%s) not found, removing from state", functionName, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Permission (%s/%s): %s", functionName, d.Id(), err)
	}

	return append(diags, resourcePermissionFlatten(ctx, d, meta.(*conns.AWSClient), statement, functionName)...)
}

func resourcePermissionFlatten(ctx context.Context, d *schema.ResourceData, awsClient *conns.AWSClient, statement *policyStatement, functionName string) diag.Diagnostics {
	var diags diag.Diagnostics
	qualifier, _ := getQualifierFromAliasOrVersionARN(statement.Resource)
	d.Set("qualifier", qualifier)

	// Save Lambda function name in the same format
	if strings.HasPrefix(functionName, "arn:"+awsClient.Partition(ctx)+":lambda:") {
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
	if v, ok := statement.Principal.(map[string]any); ok {
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

	if v, ok := statement.Condition["Bool"]; ok {
		d.Set("invoked_via_function_url", strings.EqualFold(v["lambda:InvokedViaFunctionUrl"], "true"))
	}

	if v, ok := statement.Condition["ArnLike"]; ok {
		d.Set("source_arn", v["AWS:SourceArn"])
	}

	d.Set("statement_id", statement.Sid)
	d.Set("statement_id_prefix", create.NamePrefixFromName(statement.Sid))

	return diags
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := lambda.RemovePermissionInput{
		FunctionName: aws.String(functionName),
		StatementId:  aws.String(d.Id()),
	}
	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[INFO] Deleting Lambda Permission: %s", d.Id())
	_, err := conn.RemovePermission(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "removing Lambda Permission (%s/%s): %s", functionName, d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, lambdaPropagationTimeout, func(ctx context.Context) (any, error) {
		return findPolicyStatementByTwoPartKey(ctx, conn, functionName, d.Id(), d.Get("qualifier").(string))
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Permission (%s/%s) delete: %s", functionName, d.Id(), err)
	}

	return diags
}

func findPolicy(ctx context.Context, conn *lambda.Client, input *lambda.GetPolicyInput) (*lambda.GetPolicyOutput, error) {
	output, err := conn.GetPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findPolicyStatementByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, statementID, qualifier string) (*policyStatement, error) {
	input := lambda.GetPolicyInput{
		FunctionName: aws.String(functionName),
	}
	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	output, err := findPolicy(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	policy := &policy{}
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
		Message: fmt.Sprintf("statement %s not found in policy", statementID),
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

type policy struct {
	Version   string
	Statement []policyStatement
	Id        string
}

type policyStatement struct {
	Condition map[string]map[string]string
	Action    string
	Resource  string
	Effect    string
	Principal any
	Sid       string
}

var _ inttypes.SDKv2ImportID = permissionImportID{}

type permissionImportID struct{}

func (permissionImportID) Create(d *schema.ResourceData) string {
	// For backward compatibility, the id attribute is set to the statement ID
	return d.Get("statement_id").(string)
}

func (permissionImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return id, nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/STATEMENT_ID or FUNCTION_NAME:QUALIFIER/STATEMENT_ID", id)
	}

	functionName := parts[0]
	statementID := parts[1]
	results := map[string]any{
		"function_name": functionName,
		"statement_id":  statementID,
	}

	if fnParts := strings.Split(functionName, ":"); len(fnParts) == 2 {
		results["qualifier"] = fnParts[1]
	}

	// For backward compatibility, the id attribute is set to the statement ID
	return statementID, results, nil
}
