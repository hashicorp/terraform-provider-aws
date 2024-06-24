// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const defaultAuthorizerTTL = 300

// @SDKResource("aws_api_gateway_authorizer", name="Authorizer")
func resourceAuthorizer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAuthorizerCreate,
		ReadWithoutTimeout:   resourceAuthorizerRead,
		UpdateWithoutTimeout: resourceAuthorizerUpdate,
		DeleteWithoutTimeout: resourceAuthorizerDelete,

		CustomizeDiff: resourceAuthorizerCustomizeDiff,

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
			names.AttrARN: {
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
				Default:      defaultAuthorizerTTL,
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
			names.AttrName: {
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
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.AuthorizerTypeToken,
				ValidateDiagFunc: enum.Validate[types.AuthorizerType](),
			},
		},
	}
}

func resourceAuthorizerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	var postCreateOps []types.PatchOperation
	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateAuthorizerInput{
		IdentitySource:               aws.String(d.Get("identity_source").(string)),
		Name:                         aws.String(name),
		RestApiId:                    aws.String(d.Get("rest_api_id").(string)),
		Type:                         types.AuthorizerType(d.Get(names.AttrType).(string)),
		AuthorizerResultTtlInSeconds: aws.Int32(int32(d.Get("authorizer_result_ttl_in_seconds").(int))),
	}

	if v, ok := d.GetOk("authorizer_uri"); ok {
		input.AuthorizerUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authorizer_credentials"); ok {
		// While the CreateAuthorizer method allows one to pass AuthorizerCredentials
		// regardless of authorizer Type, the API ignores this setting if the authorizer
		// is of Type "COGNITO_USER_POOLS"; thus, a PatchOperation is used as an alternative.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16613
		if input.Type != types.AuthorizerTypeCognitoUserPools {
			input.AuthorizerCredentials = aws.String(v.(string))
		} else {
			postCreateOps = append(postCreateOps, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/authorizerCredentials"),
				Value: aws.String(v.(string)),
			})
		}
	}

	if v, ok := d.GetOk("identity_validation_expression"); ok {
		input.IdentityValidationExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_arns"); ok {
		input.ProviderARNs = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateAuthorizer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	if postCreateOps != nil {
		input := &apigateway.UpdateAuthorizerInput{
			AuthorizerId:    aws.String(d.Id()),
			PatchOperations: postCreateOps,
			RestApiId:       input.RestApiId,
		}

		_, err := conn.UpdateAuthorizer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	authorizer, err := findAuthorizerByTwoPartKey(ctx, conn, d.Id(), apiID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, authorizerARN(meta.(*conns.AWSClient), apiID, d.Id()))
	d.Set("authorizer_credentials", authorizer.AuthorizerCredentials)
	if authorizer.AuthorizerResultTtlInSeconds != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("authorizer_result_ttl_in_seconds", authorizer.AuthorizerResultTtlInSeconds)
	} else {
		d.Set("authorizer_result_ttl_in_seconds", defaultAuthorizerTTL)
	}
	d.Set("authorizer_uri", authorizer.AuthorizerUri)
	d.Set("identity_source", authorizer.IdentitySource)
	d.Set("identity_validation_expression", authorizer.IdentityValidationExpression)
	d.Set(names.AttrName, authorizer.Name)
	d.Set("provider_arns", authorizer.ProviderARNs)
	d.Set(names.AttrType, authorizer.Type)

	return diags
}

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	if d.HasChange("authorizer_uri") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/authorizerUri"),
			Value: aws.String(d.Get("authorizer_uri").(string)),
		})
	}
	if d.HasChange("identity_source") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/identitySource"),
			Value: aws.String(d.Get("identity_source").(string)),
		})
	}
	if d.HasChange(names.AttrName) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/name"),
			Value: aws.String(d.Get(names.AttrName).(string)),
		})
	}
	if d.HasChange(names.AttrType) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/type"),
			Value: aws.String(d.Get(names.AttrType).(string)),
		})
	}
	if d.HasChange("authorizer_credentials") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/authorizerCredentials"),
			Value: aws.String(d.Get("authorizer_credentials").(string)),
		})
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/authorizerResultTtlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("authorizer_result_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("identity_validation_expression") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
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
			operations = append(operations, types.PatchOperation{
				Op:    types.OpAdd,
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpRemove,
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

	_, err := conn.UpdateAuthorizer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[INFO] Deleting API Gateway Authorizer: %s", d.Id())
	_, err := conn.DeleteAuthorizer(ctx, &apigateway.DeleteAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	})

	// XXX: Figure out a way to delete the method that depends on the authorizer first
	// otherwise the authorizer will be dangling until the API is deleted.
	if errs.IsA[*types.ConflictException](err) || errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAuthorizerCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// switch type between COGNITO_USER_POOLS and TOKEN/REQUEST will create new resource.
	if diff.HasChange(names.AttrType) {
		o, n := diff.GetChange(names.AttrType)
		if o.(string) == string(types.AuthorizerTypeCognitoUserPools) || n.(string) == string(types.AuthorizerTypeCognitoUserPools) {
			if err := diff.ForceNew(names.AttrType); err != nil {
				return err
			}
		}
	}

	switch authType, rawConfig := types.AuthorizerType(diff.Get(names.AttrType).(string)), diff.GetRawConfig(); authType {
	// authorizer_uri is required for authorizer TOKEN/REQUEST.
	case types.AuthorizerTypeRequest, types.AuthorizerTypeToken:
		if v := rawConfig.GetAttr("authorizer_uri"); v.IsKnown() && (v.IsNull() || v.AsString() == "") {
			return fmt.Errorf("authorizer_uri must be set non-empty when authorizer type is %s", authType)
		}
		// provider_arns is required for authorizer COGNITO_USER_POOLS.
	case types.AuthorizerTypeCognitoUserPools:
		if v := rawConfig.GetAttr("provider_arns"); v.IsKnown() && (v.IsNull() || v.AsValueSet().Length() == 0) {
			return fmt.Errorf("provider_arns must be set non-empty when authorizer type is %s", authType)
		}
	}

	return nil
}

func findAuthorizerByTwoPartKey(ctx context.Context, conn *apigateway.Client, authorizerID, apiID string) (*apigateway.GetAuthorizerOutput, error) {
	input := &apigateway.GetAuthorizerInput{
		AuthorizerId: aws.String(authorizerID),
		RestApiId:    aws.String(apiID),
	}

	output, err := conn.GetAuthorizer(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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
