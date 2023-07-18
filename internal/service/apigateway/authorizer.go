// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const DefaultAuthorizerTTL = 300

// @SDKResource("aws_api_gateway_authorizer")
func ResourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,
		CustomizeDiff:        resourceAuthorizerCustomizeDiff,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/AUTHORIZER-ID", d.Id())
				}
				restAPIId := idParts[0]
				authorizerId := idParts[1]
				d.Set("rest_api_id", restAPIId)
				d.SetId(authorizerId)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorizer_credentials": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"authorizer_result_ttl_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 3600),
				Default:      DefaultAuthorizerTTL,
			},
			"authorizer_uri": {
				Type:     schema.TypeString,
				Optional: true, // authorizer_uri is required for authorizer TOKEN/REQUEST
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
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      apigateway.AuthorizerTypeToken,
				ValidateFunc: validation.StringInSlice(apigateway.AuthorizerType_Values(), false),
			},
		},
	}
}

func resourceAuthorizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	if err := validateAuthorizerType(d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Authorizer: %s", err)
	}

	var postCreateOps []*apigateway.PatchOperation
	name := d.Get("name").(string)
	input := &apigateway.CreateAuthorizerInput{
		IdentitySource:               aws.String(d.Get("identity_source").(string)),
		Name:                         aws.String(name),
		RestApiId:                    aws.String(d.Get("rest_api_id").(string)),
		Type:                         aws.String(d.Get("type").(string)),
		AuthorizerResultTtlInSeconds: aws.Int64(int64(d.Get("authorizer_result_ttl_in_seconds").(int))),
	}

	if v, ok := d.GetOk("authorizer_uri"); ok {
		input.AuthorizerUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authorizer_credentials"); ok {
		// While the CreateAuthorizer method allows one to pass AuthorizerCredentials
		// regardless of authorizer Type, the API ignores this setting if the authorizer
		// is of Type "COGNITO_USER_POOLS"; thus, a PatchOperation is used as an alternative.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16613
		if aws.StringValue(input.Type) != apigateway.AuthorizerTypeCognitoUserPools {
			input.AuthorizerCredentials = aws.String(v.(string))
		} else {
			postCreateOps = append(postCreateOps, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/authorizerCredentials"),
				Value: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("identity_validation_expression"); ok {
		input.IdentityValidationExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_arns"); ok {
		input.ProviderARNs = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateAuthorizerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	if postCreateOps != nil {
		input := &apigateway.UpdateAuthorizerInput{
			AuthorizerId:    aws.String(d.Id()),
			PatchOperations: postCreateOps,
			RestApiId:       input.RestApiId,
		}

		_, err := conn.UpdateAuthorizerWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	apiID := d.Get("rest_api_id").(string)
	authorizer, err := FindAuthorizerByTwoPartKey(ctx, conn, d.Id(), apiID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	d.Set("arn", authorizerARN(meta.(*conns.AWSClient), apiID, d.Id()))
	d.Set("authorizer_credentials", authorizer.AuthorizerCredentials)
	if authorizer.AuthorizerResultTtlInSeconds != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("authorizer_result_ttl_in_seconds", authorizer.AuthorizerResultTtlInSeconds)
	} else {
		d.Set("authorizer_result_ttl_in_seconds", DefaultAuthorizerTTL)
	}
	d.Set("authorizer_uri", authorizer.AuthorizerUri)
	d.Set("identity_source", authorizer.IdentitySource)
	d.Set("identity_validation_expression", authorizer.IdentityValidationExpression)
	d.Set("name", authorizer.Name)
	d.Set("provider_arns", aws.StringValueSlice(authorizer.ProviderARNs))
	d.Set("type", authorizer.Type)

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("authorizer_uri") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/authorizerUri"),
			Value: aws.String(d.Get("authorizer_uri").(string)),
		})
	}
	if d.HasChange("identity_source") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/identitySource"),
			Value: aws.String(d.Get("identity_source").(string)),
		})
	}
	if d.HasChange("name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}
	if d.HasChange("type") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/type"),
			Value: aws.String(d.Get("type").(string)),
		})
	}
	if d.HasChange("authorizer_credentials") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/authorizerCredentials"),
			Value: aws.String(d.Get("authorizer_credentials").(string)),
		})
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/authorizerResultTtlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("authorizer_result_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("identity_validation_expression") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
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
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpAdd),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpRemove),
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
	}

	input := &apigateway.UpdateAuthorizerInput{
		AuthorizerId:    aws.String(d.Id()),
		PatchOperations: operations,
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
	}

	_, err := conn.UpdateAuthorizerWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	log.Printf("[INFO] Deleting API Gateway Authorizer: %s", d.Id())
	_, err := conn.DeleteAuthorizerWithContext(ctx, &apigateway.DeleteAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	})

	if err != nil {
		// XXX: Figure out a way to delete the method that depends on the authorizer first
		// otherwise the authorizer will be dangling until the API is deleted
		if !strings.Contains(err.Error(), apigateway.ErrCodeConflictException) {
			return sdkdiag.AppendErrorf(diags, "deleting API Gateway Authorizer failed: %s", err)
		}
	}

	return diags
}

func resourceAuthorizerCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
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

func validateAuthorizerType(d *schema.ResourceData) error {
	authType := d.Get("type").(string)
	// authorizer_uri is required for authorizer TOKEN/REQUEST
	if authType == apigateway.AuthorizerTypeRequest || authType == apigateway.AuthorizerTypeToken {
		if v, ok := d.GetOk("authorizer_uri"); !ok || v.(string) == "" {
			return fmt.Errorf("authorizer_uri must be set non-empty when authorizer type is %s", authType)
		}
	}
	// provider_arns is required for authorizer COGNITO_USER_POOLS.
	if authType == apigateway.AuthorizerTypeCognitoUserPools {
		if v, ok := d.GetOk("provider_arns"); !ok || v.(*schema.Set).Len() == 0 {
			return fmt.Errorf("provider_arns must be set non-empty when authorizer type is %s", authType)
		}
	}

	return nil
}

func FindAuthorizerByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, authorizerID, apiID string) (*apigateway.Authorizer, error) {
	input := &apigateway.GetAuthorizerInput{
		AuthorizerId: aws.String(authorizerID),
		RestApiId:    aws.String(apiID),
	}

	output, err := conn.GetAuthorizerWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
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

func authorizerARN(c *conns.AWSClient, apiID, authorizerID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "apigateway",
		Region:    c.Region,
		Resource:  fmt.Sprintf("/restapis/%s/authorizers/%s", apiID, authorizerID),
	}.String()
}
