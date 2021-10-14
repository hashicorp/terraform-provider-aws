package lambda

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				ValidateFunc: validPermissionAction,
			},
			"event_source_token": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validPermissionEventSourceToken,
			},
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validFunctionName,
			},
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"qualifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validQualifier,
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
				ConflictsWith: []string{"statement_id_prefix"},
				ValidateFunc:  validPolicyStatementID,
			},
			"statement_id_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"statement_id"},
				ValidateFunc:  validPolicyStatementID,
			},
		},
	}
}

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName := d.Get("function_name").(string)

	var statementId string
	if v, ok := d.GetOk("statement_id"); ok {
		statementId = v.(string)
	} else if v, ok := d.GetOk("statement_id_prefix"); ok {
		statementId = resource.PrefixedUniqueId(v.(string))
	} else {
		statementId = resource.UniqueId()
	}

	// There is a bug in the API (reported and acknowledged by AWS)
	// which causes some permissions to be ignored when API calls are sent in parallel
	// We work around this bug via mutex
	conns.GlobalMutexKV.Lock(functionName)
	defer conns.GlobalMutexKV.Unlock(functionName)

	input := lambda.AddPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		FunctionName: aws.String(functionName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementId),
	}

	if v, ok := d.GetOk("event_source_token"); ok {
		input.EventSourceToken = aws.String(v.(string))
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

	log.Printf("[DEBUG] Adding new Lambda permission: %s", input)
	var out *lambda.AddPermissionOutput
	// Retry for IAM and Lambda eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.AddPermission(&input)

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceConflictException, "") || tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceNotFoundException, "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.AddPermission(&input)
	}
	if err != nil {
		return fmt.Errorf("Error adding new Lambda Permission for %s: %s", functionName, err)
	}

	if out != nil && out.Statement != nil {
		log.Printf("[DEBUG] Created new Lambda permission: %s", *out.Statement)
	} else {
		log.Printf("[DEBUG] Created new Lambda permission, but no Statement was included")
	}

	d.SetId(statementId)

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		// IAM is eventually consistent :/
		err := resourcePermissionRead(d, meta)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Error reading Lambda policy: ResourceNotFoundException") {
				return resource.RetryableError(
					fmt.Errorf("Error reading newly created Lambda Permission for %s, retrying: %s",
						*input.FunctionName, err))
			}
			if strings.HasPrefix(err.Error(), "Failed to find statement \""+d.Id()) {
				return resource.RetryableError(
					fmt.Errorf("Error reading newly created Lambda Permission statement for %s, retrying: %s",
						*input.FunctionName, err))
			}

			log.Printf("[ERROR] An actual error occurred when expecting Lambda policy to be there: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		err = resourcePermissionRead(d, meta)
	}
	if err != nil {
		return fmt.Errorf("Error reading new Lambda permissions: %s", err)
	}
	return nil
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	input := lambda.GetPolicyInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}
	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Looking for Lambda permission: %s", input)
	var out *lambda.GetPolicyOutput
	var statement *PolicyStatement
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		// IAM is eventually consistent :/
		var err error
		out, err = conn.GetPolicy(&input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ResourceNotFoundException" {
					return resource.RetryableError(err)
				}
			}
			return resource.NonRetryableError(err)
		}

		statement, err = getLambdaPolicyStatement(out, d.Id())
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.GetPolicy(&input)

		if err == nil {
			var psErr error
			statement, psErr = getLambdaPolicyStatement(out, d.Id())

			// handle the resource not existing
			if awsErr, ok := psErr.(awserr.Error); ok {
				if awsErr.Code() == "ResourceNotFoundException" {
					log.Printf("[WARN] No Lambda Permission Policy found: %v", input)
					d.SetId("")
					return nil
				}
			}

			if psErr != nil {
				return psErr
			}
		}
	}

	if err != nil {
		// Missing whole policy or Lambda function (API error)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				log.Printf("[WARN] No Lambda Permission Policy found: %v", input)
				d.SetId("")
				return nil
			}
		}

		// Missing permission inside valid policy
		if tfresource.NotFound(err) {
			log.Printf("[WARN] %s", err)
			d.SetId("")
			return nil
		}

		return err
	}

	if statement == nil {
		return fmt.Errorf("No Lambda Permission policy found with ID %s", d.Id())
	}

	qualifier, err := GetQualifierFromAliasOrVersionARN(statement.Resource)
	if err != nil {
		log.Printf("[ERR] Error getting Lambda Qualifier: %s", err)
	}
	d.Set("qualifier", qualifier)

	// Save Lambda function name in the same format
	if strings.HasPrefix(d.Get("function_name").(string), "arn:"+meta.(*conns.AWSClient).Partition+":lambda:") {
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
	if _, ok := statement.Principal["AWS"]; ok {
		d.Set("principal", statement.Principal["AWS"])
	} else {
		d.Set("principal", statement.Principal["Service"])
	}

	if stringEquals, ok := statement.Condition["StringEquals"]; ok {
		d.Set("source_account", stringEquals["AWS:SourceAccount"])
		d.Set("event_source_token", stringEquals["lambda:EventSourceToken"])
	}

	if arnLike, ok := statement.Condition["ArnLike"]; ok {
		d.Set("source_arn", arnLike["AWS:SourceArn"])
	}

	d.Set("statement_id", statement.Sid)

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

	input := lambda.RemovePermissionInput{
		FunctionName: aws.String(functionName),
		StatementId:  aws.String(d.Id()),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Removing Lambda permission: %s", input)
	_, err := conn.RemovePermission(&input)
	if err != nil {
		// Missing whole policy or Lambda function (API error)
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				log.Printf("[WARN] No Lambda Permission Policy found: %v", input)
				return nil
			}
		}
		return err
	}

	params := &lambda.GetPolicyInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}
	if v, ok := d.GetOk("qualifier"); ok {
		params.Qualifier = aws.String(v.(string))
	}

	resp, err := conn.GetPolicy(params)

	if tfawserr.ErrMessageContains(err, "ResourceNotFoundException", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Lambda Permission policy: %s", err)
	}

	if resp.Policy == nil {
		return nil
	}

	statement, err := getLambdaPolicyStatement(resp, d.Id())

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Lambda Permission (%s) statement after deletion: %w", d.Id(), err)
	}

	if statement != nil {
		return fmt.Errorf("Failed to delete Lambda permission with ID %s", d.Id())
	}

	log.Printf("[DEBUG] Lambda permission with ID %q removed", d.Id())

	return nil
}
func getLambdaPolicyStatement(out *lambda.GetPolicyOutput, statemendId string) (statement *PolicyStatement, err error) {
	policyInBytes := []byte(*out.Policy)
	policy := Policy{}
	err = json.Unmarshal(policyInBytes, &policy)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Lambda policy: %s", err)
	}

	return FindPolicyStatementByID(&policy, statemendId)
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
	Principal map[string]string
	Sid       string
}
