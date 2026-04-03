// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package apigateway

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
				Default:          awstypes.AuthorizerTypeToken,
				ValidateDiagFunc: enum.Validate[awstypes.AuthorizerType](),
			},
		},
	}
}

func resourceAuthorizerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	var postCreateOps []awstypes.PatchOperation
	name := d.Get(names.AttrName).(string)
	input := apigateway.CreateAuthorizerInput{
		IdentitySource:               aws.String(d.Get("identity_source").(string)),
		Name:                         aws.String(name),
		RestApiId:                    aws.String(d.Get("rest_api_id").(string)),
		Type:                         awstypes.AuthorizerType(d.Get(names.AttrType).(string)),
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
		if input.Type != awstypes.AuthorizerTypeCognitoUserPools {
			input.AuthorizerCredentials = aws.String(v.(string))
		} else {
			postCreateOps = append(postCreateOps, awstypes.PatchOperation{
				Op:    awstypes.OpReplace,
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

	output, err := conn.CreateAuthorizer(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Authorizer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	if postCreateOps != nil {
		input := apigateway.UpdateAuthorizerInput{
			AuthorizerId:    aws.String(d.Id()),
			PatchOperations: postCreateOps,
			RestApiId:       input.RestApiId,
		}

		_, err := conn.UpdateAuthorizer(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	authorizer, err := findAuthorizerByTwoPartKey(ctx, conn, d.Id(), apiID)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] API Gateway Authorizer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, authorizerARN(ctx, c, apiID, d.Id()))
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

func resourceAuthorizerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]awstypes.PatchOperation, 0)

	if d.HasChange("authorizer_uri") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/authorizerUri"),
			Value: aws.String(d.Get("authorizer_uri").(string)),
		})
	}
	if d.HasChange("identity_source") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/identitySource"),
			Value: aws.String(d.Get("identity_source").(string)),
		})
	}
	if d.HasChange(names.AttrName) {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/name"),
			Value: aws.String(d.Get(names.AttrName).(string)),
		})
	}
	if d.HasChange(names.AttrType) {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/type"),
			Value: aws.String(d.Get(names.AttrType).(string)),
		})
	}
	if d.HasChange("authorizer_credentials") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/authorizerCredentials"),
			Value: aws.String(d.Get("authorizer_credentials").(string)),
		})
	}
	if d.HasChange("authorizer_result_ttl_in_seconds") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/authorizerResultTtlInSeconds"),
			Value: aws.String(strconv.Itoa(d.Get("authorizer_result_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("identity_validation_expression") {
		operations = append(operations, awstypes.PatchOperation{
			Op:    awstypes.OpReplace,
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
			operations = append(operations, awstypes.PatchOperation{
				Op:    awstypes.OpAdd,
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, awstypes.PatchOperation{
				Op:    awstypes.OpRemove,
				Path:  aws.String("/providerARNs"),
				Value: aws.String(v.(string)),
			})
		}
	}

	input := apigateway.UpdateAuthorizerInput{
		AuthorizerId:    aws.String(d.Id()),
		PatchOperations: operations,
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
	}

	_, err := conn.UpdateAuthorizer(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAuthorizerRead(ctx, d, meta)...)
}

func resourceAuthorizerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[INFO] Deleting API Gateway Authorizer: %s", d.Id())
	input := apigateway.DeleteAuthorizerInput{
		AuthorizerId: aws.String(d.Id()),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}
	_, err := conn.DeleteAuthorizer(ctx, &input)

	// TODO: Figure out a way to delete the method that depends on the authorizer first
	// otherwise the authorizer will be dangling until the API is deleted.
	if errs.IsA[*awstypes.ConflictException](err) || errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Authorizer (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAuthorizerCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, v any) error {
	var plan authorizerResourceModel
	if err := tfcty.GetFramework(ctx, diff.GetRawPlan(), &plan); err != nil {
		return fmt.Errorf("RawPlan to framework model: %w", err)
	}
	if plan.Type.IsUnknown() {
		return nil
	}

	// RawPlan contains no schema defaults in CustomizeDiff.
	authType := cmp.Or(plan.Type.ValueEnum(), awstypes.AuthorizerTypeToken)

	// switch type between COGNITO_USER_POOLS and TOKEN/REQUEST will create new resource.
	if rawState := diff.GetRawState(); !rawState.IsNull() { // RawState is null on Create.
		var state authorizerResourceModel
		if err := tfcty.GetFramework(ctx, rawState, &state); err != nil {
			return fmt.Errorf("RawState to framework model: %w", err)
		}

		if o, n := state.Type.ValueEnum(), authType; o != n && (o == awstypes.AuthorizerTypeCognitoUserPools || n == awstypes.AuthorizerTypeCognitoUserPools) {
			if err := diff.ForceNew(names.AttrType); err != nil {
				return err
			}
		}
	}

	var config authorizerResourceModel
	if err := tfcty.GetFramework(ctx, diff.GetRawConfig(), &config); err != nil {
		return fmt.Errorf("RawConfig to framework model: %w", err)
	}

	switch authType {
	case awstypes.AuthorizerTypeRequest, awstypes.AuthorizerTypeToken:
		// authorizer_uri is required for authorizer TOKEN/REQUEST.
		if authorizerURI := config.AuthorizerURI; !authorizerURI.IsUnknown() && (authorizerURI.IsNull() || authorizerURI.ValueString() == "") {
			return fmt.Errorf("authorizer_uri must be set non-empty when authorizer type is %s", authType)
		}
	case awstypes.AuthorizerTypeCognitoUserPools:
		// provider_arns is required for authorizer COGNITO_USER_POOLS.
		if providerARNs := config.ProviderARNs; !providerARNs.IsUnknown() && providerARNs.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}) == 0 {
			return fmt.Errorf("provider_arns must be set non-empty when authorizer type is %s", authType)
		}
	}

	return nil
}

func findAuthorizerByTwoPartKey(ctx context.Context, conn *apigateway.Client, authorizerID, apiID string) (*apigateway.GetAuthorizerOutput, error) {
	input := apigateway.GetAuthorizerInput{
		AuthorizerId: aws.String(authorizerID),
		RestApiId:    aws.String(apiID),
	}

	return findAuthorizer(ctx, conn, &input)
}

func findAuthorizer(ctx context.Context, conn *apigateway.Client, input *apigateway.GetAuthorizerInput) (*apigateway.GetAuthorizerOutput, error) {
	output, err := conn.GetAuthorizer(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func authorizerARN(ctx context.Context, c *conns.AWSClient, apiID, authorizerID string) string {
	return c.RegionalARNNoAccount(ctx, "apigateway", fmt.Sprintf("/restapis/%s/authorizers/%s", apiID, authorizerID))
}

type authorizerResourceModel struct {
	framework.WithRegionModel
	ARN                          types.String                                `tfsdk:"arn"`
	AuthorizerCredentials        fwtypes.ARN                                 `tfsdk:"authorizer_credentials"`
	AuthorizerResultTTLInSeconds types.Int64                                 `tfsdk:"authorizer_result_ttl_in_seconds"`
	AuthorizerURI                types.String                                `tfsdk:"authorizer_uri"`
	IdentitySource               types.String                                `tfsdk:"identity_source"`
	IdentityValidationExpression types.String                                `tfsdk:"identity_validation_expression"`
	Name                         types.String                                `tfsdk:"name"`
	ProviderARNs                 fwtypes.SetOfARN                            `tfsdk:"provider_arns"`
	RestApiID                    types.String                                `tfsdk:"rest_api_id"`
	Type                         fwtypes.StringEnum[awstypes.AuthorizerType] `tfsdk:"type"`
}
