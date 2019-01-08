package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsApiGatewayV2Authorizer() *schema.Resource {
	return &schema.Resource{
		Create:        resourceAwsApiGatewayV2AuthorizerCreate,
		Read:          resourceAwsApiGatewayV2AuthorizerRead,
		Update:        resourceAwsApiGatewayV2AuthorizerUpdate,
		Delete:        resourceAwsApiGatewayV2AuthorizerDelete,
		CustomizeDiff: resourceAwsApiGatewayV2AuthorizerCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorizer_credentials_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"authorizer_result_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 3600),
			},
			"authorizer_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "REQUEST",
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.AuthorizerTypeRequest,
				}, false),
			},
			"authorizer_uri": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity_source": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "method.request.header.Authorization",
			},
			"identity_validation_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_arns": {
				Type:     schema.TypeSet,
				Optional: true, // provider_arns is required for authorizer COGNITO_USER_POOLS.
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsApiGatewayV2AuthorizerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2

	input := apigatewayv2.CreateAuthorizerInput{
		ApiId:          aws.String(d.Get("api_id").(string)),
		IdentitySource: []*string{aws.String(d.Get("identity_source").(string))},
		Name:           aws.String(d.Get("name").(string)),
		AuthorizerType: aws.String(d.Get("authorizer_type").(string)),
	}

	if err := validateAuthorizerV2Type(d); err != nil {
		return err
	}
	if v, ok := d.GetOk("authorizer_uri"); ok {
		input.AuthorizerUri = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorizer_result_ttl_in_seconds"); ok {
		input.AuthorizerResultTtlInSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("identity_validation_expression"); ok {
		input.IdentityValidationExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("provider_arns"); ok {
		input.ProviderArns = expandStringList(v.(*schema.Set).List())
	}

	log.Printf("[INFO] Creating API GatewayV2 Authorizer: %s", input)
	out, err := conn.CreateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Error creating API GatewayV2 Authorizer: %s", err)
	}

	d.SetId(*out.AuthorizerId)

	return resourceAwsApiGatewayV2AuthorizerRead(d, meta)
}

func resourceAwsApiGatewayV2AuthorizerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2

	log.Printf("[INFO] Reading API GatewayV2 Authorizer %s", d.Id())
	input := apigatewayv2.GetAuthorizerInput{
		AuthorizerId: aws.String(d.AuthorizerId()),
		ApiId:        aws.String(d.Get("api_id").(string)),
	}

	authorizer, err := conn.GetAuthorizer(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] No API GatewayV2 Authorizer found: %s", input)
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API GatewayV2 Authorizer: %s", authorizer)

	d.Set("authorizer_credentials_arn", authorizer.AuthorizerCredentialsArn)
	d.Set("authorizer_result_ttl_in_seconds", authorizer.AuthorizerResultTtlInSeconds)
	d.Set("authorizer_uri", authorizer.AuthorizerUri)
	d.Set("identity_source", authorizer.IdentitySource)
	d.Set("identity_validation_expression", authorizer.IdentityValidationExpression)
	d.Set("name", authorizer.Name)
	d.Set("authorizer_type", authorizer.AuthorizerType)
	d.Set("provider_arns", flattenStringList(authorizer.ProviderArns))

	return nil
}

func resourceAwsApiGatewayV2AuthorizerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2

	input := apigatewayv2.UpdateAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}

	operations := make([]*apigatewayv2.PatchOperation, 0)

	if d.HasChange("authorizer_uri") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerUri"),
			Value: aws.String(d.Get("authorizer_uri").(string)),
		})
	}
	if d.HasChange("identity_source") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/identitySource"),
			Value: aws.String(d.Get("identity_source").(string)),
		})
	}
	if d.HasChange("name") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}
	if d.HasChange("type") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/type"),
			Value: aws.String(d.Get("type").(string)),
		})
	}
	if d.HasChange("authorizer_credentials") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerCredentials"),
			Value: aws.String(d.Get("authorizer_credentials").(string)),
		})
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/authorizerResultTtlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("authorizer_result_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("identity_validation_expression") {
		operations = append(operations, &apigatewayv2.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/identityValidationExpression"),
			Value: aws.String(d.Get("identity_validation_expression").(string)),
		})
	}
	if d.HasChange("provider_arns") {
		old, new := d.GetChange("provider_arns")
		os := old.(*schema.Set)
		ns := new.(*schema.Set)
		// providerARNs can't be empty, so add first and then remove
		additionList := ns.Difference(os)
		for _, v := range additionList.List() {
			operations = append(operations, &apigatewayv2.PatchOperation{
				Op:    aws.String("add"),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, &apigatewayv2.PatchOperation{
				Op:    aws.String("remove"),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
	}

	input.PatchOperations = operations

	log.Printf("[INFO] Updating API GatewayV2 Authorizer: %s", input)
	_, err := conn.UpdateAuthorizer(&input)
	if err != nil {
		return fmt.Errorf("Updating API GatewayV2 Authorizer failed: %s", err)
	}

	return resourceAwsApiGatewayV2AuthorizerRead(d, meta)
}

func resourceAwsApiGatewayV2AuthorizerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2
	input := apigatewayv2.DeleteAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}
	log.Printf("[INFO] Deleting API GatewayV2 Authorizer: %s", input)
	_, err := conn.DeleteAuthorizer(&input)
	if err != nil {
		// XXX: Figure out a way to delete the method that depends on the authorizer first
		// otherwise the authorizer will be dangling until the API is deleted
		if !strings.Contains(err.Error(), "ConflictException") {
			return fmt.Errorf("Deleting API GatewayV2 Authorizer failed: %s", err)
		}
	}

	return nil
}

func resourceAwsApiGatewayV2AuthorizerCustomizeDiff(diff *schema.ResourceDiff, v interface{}) error {
	// switch type between COGNITO_USER_POOLS and TOKEN/REQUEST will create new resource.
	if diff.HasChange("type") {
		o, n := diff.GetChange("type")
		if o.(string) == apigatewayv2.AuthorizerTypeCognitoUserPools || n.(string) == apigatewayv2.AuthorizerTypeCognitoUserPools {
			if err := diff.ForceNew("type"); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateAuthorizerV2Type(d *schema.ResourceData) error {
	authType := d.Get("type").(string)
	// authorizer_uri is required for authorizer TOKEN/REQUEST
	if authType == apigatewayv2.AuthorizerTypeRequest || authType == apigatewayv2.AuthorizerTypeToken {
		if v, ok := d.GetOk("authorizer_uri"); !ok || v.(string) == "" {
			return fmt.Errorf("authorizer_uri must be set non-empty when authorizer type is %s", authType)
		}
	}
	// provider_arns is required for authorizer COGNITO_USER_POOLS.
	if authType == apigatewayv2.AuthorizerTypeCognitoUserPools {
		if v, ok := d.GetOk("provider_arns"); !ok || len(v.(*schema.Set).List()) == 0 {
			return fmt.Errorf("provider_arns must be set non-empty when authorizer type is %s", authType)
		}
	}

	return nil
}
