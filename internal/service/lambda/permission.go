package lambda

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var LambdaFunctionRegexp = `^(arn:[\w-]+:lambda:)?([a-z]{2}-(?:[a-z]+-){1,2}\d{1}:)?(\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:(\$LATEST|[a-zA-Z0-9-_]+))?$`

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Delete: resourcePermissionDelete,

		Importer: &schema.ResourceImporter{
			State: resourcePermissionImport,
		},

		Schema: map[string]*schema.Schema{
			"action": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(lambda.FunctionUrlAuthType_Values(), false),
			},
			"principal": {
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

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName := d.Get("function_name").(string)
	statementID := create.Name(d.Get("statement_id").(string), d.Get("statement_id_prefix").(string))

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := &lambda.AddPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		FunctionName: aws.String(functionName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementID),
	}

	if v, ok := d.GetOk("event_source_token"); ok {
		input.EventSourceToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("function_url_auth_type"); ok {
		input.FunctionUrlAuthType = aws.String(v.(string))
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

	log.Printf("[DEBUG] Adding Lambda Permission: %s", input)
	// Retry for IAM and Lambda eventual consistency.
	_, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout,
		func() (interface{}, error) {
			return conn.AddPermission(input)
		},
		lambda.ErrCodeResourceConflictException, lambda.ErrCodeResourceNotFoundException)

	if err != nil {
		return fmt.Errorf("adding Lambda Permission (%s/%s): %w", functionName, statementID, err)
	}

	d.SetId(statementID)

	return resourcePermissionRead(d, meta)
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName := d.Get("function_name").(string)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout,
		func() (interface{}, error) {
			return FindPolicyStatementByTwoPartKey(conn, functionName, d.Id(), d.Get("qualifier").(string))
		}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Permission (%s/%s) not found, removing from state", functionName, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Lambda Permission (%s/%s): %w", functionName, d.Id(), err)
	}

	statement := outputRaw.(*PolicyStatement)
	qualifier, _ := GetQualifierFromAliasOrVersionARN(statement.Resource)

	d.Set("qualifier", qualifier)

	// Save Lambda function name in the same format
	if strings.HasPrefix(functionName, "arn:"+meta.(*conns.AWSClient).Partition+":lambda:") {
		// Strip qualifier off
		trimmedArn := strings.TrimSuffix(statement.Resource, ":"+qualifier)
		d.Set("function_name", trimmedArn)
	} else {
		functionName, err := GetFunctionNameFromARN(statement.Resource)

		if err != nil {
			return err
		}

		d.Set("function_name", functionName)
	}

	d.Set("action", statement.Action)
	// Check if the principal is a cross-account IAM role
	if v, ok := statement.Principal.(map[string]interface{}); ok {
		if _, ok := v["AWS"]; ok {
			d.Set("principal", v["AWS"])
		} else {
			d.Set("principal", v["Service"])
		}
	} else if v, ok := statement.Principal.(string); ok {
		d.Set("principal", v)
	}

	if stringEquals, ok := statement.Condition["StringEquals"]; ok {
		d.Set("source_account", stringEquals["AWS:SourceAccount"])
		d.Set("event_source_token", stringEquals["lambda:EventSourceToken"])
		d.Set("principal_org_id", stringEquals["aws:PrincipalOrgID"])
		d.Set("function_url_auth_type", stringEquals["lambda:FunctionUrlAuthType"])
	}

	if arnLike, ok := statement.Condition["ArnLike"]; ok {
		d.Set("source_arn", arnLike["AWS:SourceArn"])
	}

	d.Set("statement_id", statement.Sid)
	d.Set("statement_id_prefix", create.NamePrefixFromName(statement.Sid))

	return nil
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

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

	log.Printf("[DEBUG] Removing Lambda Permission: %s", input)
	_, err := conn.RemovePermission(input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("removing Lambda Permission (%s/%s): %w", functionName, d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(propagationTimeout, func() (interface{}, error) {
		return FindPolicyStatementByTwoPartKey(conn, functionName, d.Id(), d.Get("qualifier").(string))
	})

	if err != nil {
		return fmt.Errorf("waiting for Lambda Permission (%s/%s) delete: %w", functionName, d.Id(), err)
	}

	return nil
}

func findPolicy(conn *lambda.Lambda, input *lambda.GetPolicyInput) (*lambda.GetPolicyOutput, error) {
	output, err := conn.GetPolicy(input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindPolicyStatementByTwoPartKey(conn *lambda.Lambda, functionName, statementID, qualifier string) (*PolicyStatement, error) {
	input := &lambda.GetPolicyInput{
		FunctionName: aws.String(functionName),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	output, err := findPolicy(conn, input)

	if err != nil {
		return nil, err
	}

	policy := Policy{}
	err = json.Unmarshal([]byte(aws.StringValue(output.Policy)), &policy)

	if err != nil {
		return nil, err
	}

	for _, v := range policy.Statement {
		if v.Sid == statementID {
			return &v, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastRequest:  statementID,
		LastResponse: policy,
		Message:      fmt.Sprintf("Failed to find statement %q in Lambda policy:\n%s", statementID, policy.Statement),
	}
}

func FindPolicyStatementByID(policy *Policy, id string) (*PolicyStatement, error) {

	log.Printf("[DEBUG] Received %d statements in Lambda policy: %s", len(policy.Statement), policy.Statement)
	for _, statement := range policy.Statement {
		if statement.Sid == id {
			return &statement, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastRequest:  id,
		LastResponse: policy,
		Message:      fmt.Sprintf("Failed to find statement %q in Lambda policy:\n%s", id, policy.Statement),
	}
}

func GetQualifierFromAliasOrVersionARN(arn string) (string, error) {
	matches := regexp.MustCompile(LambdaFunctionRegexp).FindStringSubmatch(arn)
	if len(matches) < 8 || matches[7] == "" {
		return "", fmt.Errorf("Invalid ARN or otherwise unable to get qualifier from ARN (%q)",
			arn)
	}

	return matches[7], nil
}

func GetFunctionNameFromARN(arn string) (string, error) {
	matches := regexp.MustCompile(LambdaFunctionRegexp).FindStringSubmatch(arn)
	if len(matches) < 6 || matches[5] == "" {
		return "", fmt.Errorf("Invalid ARN or otherwise unable to get qualifier from ARN (%q)",
			arn)
	}
	return matches[5], nil
}

func resourcePermissionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/STATEMENT_ID or FUNCTION_NAME:QUALIFIER/STATEMENT_ID", d.Id())
	}

	functionName := idParts[0]

	input := &lambda.GetFunctionInput{FunctionName: &functionName}

	var qualifier string
	fnParts := strings.Split(functionName, ":")
	if len(fnParts) == 2 {
		functionName = fnParts[0]
		qualifier = fnParts[1]
		input.Qualifier = &qualifier
	}
	statementId := idParts[1]
	log.Printf("[DEBUG] Importing Lambda Permission %s for function name %s", statementId, functionName)

	conn := meta.(*conns.AWSClient).LambdaConn
	getFunctionOutput, err := conn.GetFunction(input)
	if err != nil {
		return nil, err
	}

	d.Set("function_name", getFunctionOutput.Configuration.FunctionArn)
	d.Set("statement_id", statementId)
	if qualifier != "" {
		d.Set("qualifier", qualifier)
	}
	d.SetId(statementId)
	return []*schema.ResourceData{d}, nil
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
