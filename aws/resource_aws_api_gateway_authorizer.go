package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsApiGatewayAuthorizer() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsApiGatewayAuthorizerCreate,
		Read:          resourceAwsApiGatewayAuthorizerRead,
		Update:        resourceAwsApiGatewayAuthorizerUpdate,
		Delete:        resourceAwsApiGatewayAuthorizerDelete,
		CustomizeDiff: resourceAwsApiGatewayAuthorizerCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"authorizer_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"identity_source": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "method.request.header.Authorization",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "TOKEN",
				ValidateFunc: validation.StringInSlice([]string{
					apigateway.AuthorizerTypeCognitoUserPools,
					apigateway.AuthorizerTypeRequest,
					apigateway.AuthorizerTypeToken,
				}, false),
			},
			"authorizer_credentials": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorizer_result_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validateIntegerInRange(0, 3600),
			},
			"identity_validation_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provider_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsApiGatewayAuthorizerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	input := apigateway.CreateAuthorizerInput{
		IdentitySource: aws.String(d.Get("identity_source").(string)),
		Name:           aws.String(d.Get("name").(string)),
		RestApiId:      aws.String(d.Get("rest_api_id").(string)),
		Type:           aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("authorizer_uri"); ok {
		input.AuthorizerUri = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorizer_credentials"); ok {
		input.AuthorizerCredentials = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorizer_result_ttl_in_seconds"); ok {
		input.AuthorizerResultTtlInSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("identity_validation_expression"); ok {
		input.IdentityValidationExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("provider_arns"); ok {
		input.ProviderARNs = expandStringList(v.(*schema.Set).List())
	}

	log.Printf("[INFO] Creating API Gateway Authorizer: %s", input)
	out, err := conn.CreateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Authorizer: %s", err)
	}

	d.SetId(*out.Id)

	return resourceAwsApiGatewayAuthorizerRead(d, meta)
}

func resourceAwsApiGatewayAuthorizerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[INFO] Reading API Gateway Authorizer %s", d.Id())
	input := apigateway.GetAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}

	authorizer, err := conn.GetAuthorizer(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] No API Gateway Authorizer found: %s", input)
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Authorizer: %s", authorizer)

	d.Set("authorizer_credentials", authorizer.AuthorizerCredentials)
	d.Set("authorizer_result_ttl_in_seconds", authorizer.AuthorizerResultTtlInSeconds)
	d.Set("authorizer_uri", authorizer.AuthorizerUri)
	d.Set("identity_source", authorizer.IdentitySource)
	d.Set("identity_validation_expression", authorizer.IdentityValidationExpression)
	d.Set("name", authorizer.Name)
	d.Set("type", authorizer.Type)
	d.Set("provider_arns", flattenStringList(authorizer.ProviderARNs))

	return nil
}

func resourceAwsApiGatewayAuthorizerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	input := apigateway.UpdateAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}

	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("authorizer_uri") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerUri"),
			Value: aws.String(d.Get("authorizer_uri").(string)),
		})
	}
	if d.HasChange("identity_source") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/identitySource"),
			Value: aws.String(d.Get("identity_source").(string)),
		})
	}
	if d.HasChange("name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}
	if d.HasChange("type") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/type"),
			Value: aws.String(d.Get("type").(string)),
		})
	}
	if d.HasChange("authorizer_credentials") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerCredentials"),
			Value: aws.String(d.Get("authorizer_credentials").(string)),
		})
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerResultTtlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("authorizer_result_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("identity_validation_expression") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/identityValidationExpression"),
			Value: aws.String(d.Get("identity_validation_expression").(string)),
		})
	}
	if d.HasChange("provider_arns") {
		old, new := d.GetChange("provider_arns")
		oldValue := old.(*schema.Set).List()
		newValue := new.(*schema.Set).List()
		operations = append(operations, diffProviderARNsOp("/providerARNs", oldValue, newValue)...)
	}
	input.PatchOperations = operations

	log.Printf("[INFO] Updating API Gateway Authorizer: %s", input)
	_, err := conn.UpdateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Updating API Gateway Authorizer failed: %s", err)
	}

	return resourceAwsApiGatewayAuthorizerRead(d, meta)
}

func resourceAwsApiGatewayAuthorizerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	input := apigateway.DeleteAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}
	log.Printf("[INFO] Deleting API Gateway Authorizer: %s", input)
	_, err := conn.DeleteAuthorizer(&input)
	if err != nil {
		// XXX: Figure out a way to delete the method that depends on the authorizer first
		// otherwise the authorizer will be dangling until the API is deleted
		if !strings.Contains(err.Error(), "ConflictException") {
			return fmt.Errorf("Deleting API Gateway Authorizer failed: %s", err)
		}
	}

	return nil
}

func diffProviderARNsOp(prefix string, old, new []interface{}) (ops []*apigateway.PatchOperation) {
	// providerARNs can't be empty, so add first and then remove
	for _, n := range new {
		add := true
		for _, o := range old {
			if n.(string) == o.(string) {
				add = false
			}
		}
		if add {
			ops = append(ops, &apigateway.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(n.(string)),
			})
		}
	}
	for _, o := range old {
		remove := true
		for _, n := range new {
			if o.(string) == n.(string) {
				remove = false
			}
		}
		if remove {
			ops = append(ops, &apigateway.PatchOperation{
				Op:    aws.String("remove"),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(o.(string)),
			})
		}
	}
	return
}

func resourceAwsApiGatewayAuthorizerCustomizeDiff(diff *schema.ResourceDiff, v interface{}) error {
	args := []string{"authorizer_uri", "name", "rest_api_id", "identity_source", "type", "identity_validation_expression", "authorizer_credentials"}
	for _, arg := range args {
		val, ok := diff.GetOk(arg)
		log.Printf("[DEBUG] %s: #%s#, #%v#", arg, val.(string), ok)
	}

	authType := diff.Get("type").(string)
	// authorizer_uri is required for authorizer TOKEN/REQUEST
	if authType == apigateway.AuthorizerTypeRequest || authType == apigateway.AuthorizerTypeToken {
		if val, ok := diff.GetOk("authorizer_uri"); !ok || val.(string) == "" {
			return fmt.Errorf("authorizer_uri must be set non-empty when authorizer type is %s", authType)
		}
	}
	// provider_arns is required for authorizer COGNITO_USER_POOLS.
	if authType == apigateway.AuthorizerTypeCognitoUserPools {
		if val, ok := diff.GetOk("provider_arns"); !ok || len(val.(*schema.Set).List()) == 0 {
			return fmt.Errorf("provider_arns must be set non-empty when authorizer type is %s", authType)
		}
	}

	// switch type between COGNITO_USER_POOLS and TOKEN/REQUEST will create new resource.
	if diff.HasChange("type") {
		o, n := diff.GetChange("type")
		if o.(string) == apigateway.AuthorizerTypeCognitoUserPools || n.(string) == apigateway.AuthorizerTypeCognitoUserPools {
			if err := diff.ForceNew("type"); err != nil {
				return err
			}
		}
	}

	return nil
}
