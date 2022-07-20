package apigatewayv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAuthorizerCreate,
		Read:   resourceAuthorizerRead,
		Update: resourceAuthorizerUpdate,
		Delete: resourceAuthorizerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAuthorizerImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authorizer_credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"authorizer_payload_format_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1.0", "2.0"}, false),
			},
			"authorizer_result_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(0, 3600),
			},
			"authorizer_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(apigatewayv2.AuthorizerType_Values(), false),
			},
			"authorizer_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"enable_simple_responses": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"identity_sources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"jwt_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audience": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"issuer": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAuthorizerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	apiId := d.Get("api_id").(string)
	authorizerType := d.Get("authorizer_type").(string)

	apiOutput, err := FindAPIByID(conn, apiId)

	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %s", apiId, err)
	}

	protocolType := aws.StringValue(apiOutput.ProtocolType)

	req := &apigatewayv2.CreateAuthorizerInput{
		ApiId:          aws.String(apiId),
		AuthorizerType: aws.String(authorizerType),
		IdentitySource: flex.ExpandStringSet(d.Get("identity_sources").(*schema.Set)),
		Name:           aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("authorizer_credentials_arn"); ok {
		req.AuthorizerCredentialsArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("authorizer_payload_format_version"); ok {
		req.AuthorizerPayloadFormatVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOkExists("authorizer_result_ttl_in_seconds"); ok {
		req.AuthorizerResultTtlInSeconds = aws.Int64(int64(v.(int)))
	} else if protocolType == apigatewayv2.ProtocolTypeHttp && authorizerType == apigatewayv2.AuthorizerTypeRequest {
		// Default in the AWS Console is 300 seconds.
		// Explicitly set on creation so that we can correctly detect changes to the 0 value.
		req.AuthorizerResultTtlInSeconds = aws.Int64(300)
	}
	if v, ok := d.GetOk("authorizer_uri"); ok {
		req.AuthorizerUri = aws.String(v.(string))
	}
	if v, ok := d.GetOk("enable_simple_responses"); ok {
		req.EnableSimpleResponses = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("jwt_configuration"); ok {
		req.JwtConfiguration = expandJWTConfiguration(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 authorizer: %s", req)
	resp, err := conn.CreateAuthorizer(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 authorizer: %s", err)
	}

	d.SetId(aws.StringValue(resp.AuthorizerId))

	return resourceAuthorizerRead(d, meta)
}

func resourceAuthorizerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	resp, err := conn.GetAuthorizer(&apigatewayv2.GetAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 authorizer: %s", err)
	}

	d.Set("authorizer_credentials_arn", resp.AuthorizerCredentialsArn)
	d.Set("authorizer_payload_format_version", resp.AuthorizerPayloadFormatVersion)
	d.Set("authorizer_result_ttl_in_seconds", resp.AuthorizerResultTtlInSeconds)
	d.Set("authorizer_type", resp.AuthorizerType)
	d.Set("authorizer_uri", resp.AuthorizerUri)
	d.Set("enable_simple_responses", resp.EnableSimpleResponses)
	if err := d.Set("identity_sources", flex.FlattenStringSet(resp.IdentitySource)); err != nil {
		return fmt.Errorf("error setting identity_sources: %s", err)
	}
	if err := d.Set("jwt_configuration", flattenJWTConfiguration(resp.JwtConfiguration)); err != nil {
		return fmt.Errorf("error setting jwt_configuration: %s", err)
	}
	d.Set("name", resp.Name)

	return nil
}

func resourceAuthorizerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	req := &apigatewayv2.UpdateAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	}
	if d.HasChange("authorizer_credentials_arn") {
		req.AuthorizerCredentialsArn = aws.String(d.Get("authorizer_credentials_arn").(string))
	}
	if d.HasChange("authorizer_payload_format_version") {
		req.AuthorizerPayloadFormatVersion = aws.String(d.Get("authorizer_payload_format_version").(string))
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		req.AuthorizerResultTtlInSeconds = aws.Int64(int64(d.Get("authorizer_result_ttl_in_seconds").(int)))
	}
	if d.HasChange("authorizer_type") {
		req.AuthorizerType = aws.String(d.Get("authorizer_type").(string))
	}
	if d.HasChange("authorizer_uri") {
		req.AuthorizerUri = aws.String(d.Get("authorizer_uri").(string))
	}
	if d.HasChange("enable_simple_responses") {
		req.EnableSimpleResponses = aws.Bool(d.Get("enable_simple_responses").(bool))
	}
	if d.HasChange("identity_sources") {
		req.IdentitySource = flex.ExpandStringSet(d.Get("identity_sources").(*schema.Set))
	}
	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}
	if d.HasChange("jwt_configuration") {
		req.JwtConfiguration = expandJWTConfiguration(d.Get("jwt_configuration").([]interface{}))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 authorizer: %s", req)
	_, err := conn.UpdateAuthorizer(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 authorizer: %s", err)
	}

	return resourceAuthorizerRead(d, meta)
}

func resourceAuthorizerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 authorizer (%s)", d.Id())
	_, err := conn.DeleteAuthorizer(&apigatewayv2.DeleteAuthorizerInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		AuthorizerId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 authorizer: %s", err)
	}

	return nil
}

func resourceAuthorizerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/authorizer-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}

func expandJWTConfiguration(vConfiguration []interface{}) *apigatewayv2.JWTConfiguration {
	configuration := &apigatewayv2.JWTConfiguration{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]interface{})

	if vAudience, ok := mConfiguration["audience"].(*schema.Set); ok && vAudience.Len() > 0 {
		configuration.Audience = flex.ExpandStringSet(vAudience)
	}
	if vIssuer, ok := mConfiguration["issuer"].(string); ok && vIssuer != "" {
		configuration.Issuer = aws.String(vIssuer)
	}

	return configuration
}

func flattenJWTConfiguration(configuration *apigatewayv2.JWTConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"audience": flex.FlattenStringSet(configuration.Audience),
		"issuer":   aws.StringValue(configuration.Issuer),
	}}
}
