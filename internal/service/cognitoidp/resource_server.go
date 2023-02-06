package cognitoidp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceServerCreate,
		ReadWithoutTimeout:   resourceResourceServerRead,
		UpdateWithoutTimeout: resourceResourceServerUpdate,
		DeleteWithoutTimeout: resourceResourceServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateResourceServer.html
		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scope_description": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"scope_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validResourceServerScopeName,
						},
					},
				},
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scope_identifiers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceResourceServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	identifier := d.Get("identifier").(string)
	userPoolID := d.Get("user_pool_id").(string)

	params := &cognitoidentityprovider.CreateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk("scope"); ok {
		configs := v.(*schema.Set).List()
		params.Scopes = expandServerScope(configs)
	}

	log.Printf("[DEBUG] Creating Cognito Resource Server: %s", params)

	_, err := conn.CreateResourceServerWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Cognito Resource Server: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", userPoolID, identifier))

	return append(diags, resourceResourceServerRead(ctx, d, meta)...)
}

func resourceResourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameResourceServer, d.Id(), err)
	}

	params := &cognitoidentityprovider.DescribeResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	}

	log.Printf("[DEBUG] Reading Cognito Resource Server: %s", params)

	resp, err := conn.DescribeResourceServerWithContext(ctx, params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameResourceServer, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameResourceServer, d.Id(), err)
	}

	if !d.IsNewResource() && (resp == nil || resp.ResourceServer == nil) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameResourceServer, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && (resp == nil || resp.ResourceServer == nil) {
		return create.DiagError(names.CognitoIDP, create.ErrActionReading, ResNameResourceServer, d.Id(), errors.New("not found after creation"))
	}

	d.Set("identifier", resp.ResourceServer.Identifier)
	d.Set("name", resp.ResourceServer.Name)
	d.Set("user_pool_id", resp.ResourceServer.UserPoolId)

	scopes := flattenServerScope(resp.ResourceServer.Scopes)
	if err := d.Set("scope", scopes); err != nil {
		return sdkdiag.AppendErrorf(diags, "Failed setting schema: %s", err)
	}

	var scopeIdentifiers []string
	for _, elem := range scopes {
		scopeIdentifier := fmt.Sprintf("%s/%s", aws.StringValue(resp.ResourceServer.Identifier), elem["scope_name"].(string))
		scopeIdentifiers = append(scopeIdentifiers, scopeIdentifier)
	}
	if err := d.Set("scope_identifiers", scopeIdentifiers); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scope_identifiers: %s", err)
	}
	return diags
}

func resourceResourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Resource Server (%s): %s", d.Id(), err)
	}

	params := &cognitoidentityprovider.UpdateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get("name").(string)),
		Scopes:     expandServerScope(d.Get("scope").(*schema.Set).List()),
		UserPoolId: aws.String(userPoolID),
	}

	log.Printf("[DEBUG] Updating Cognito Resource Server: %s", params)

	_, err = conn.UpdateResourceServerWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Resource Server (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResourceServerRead(ctx, d, meta)...)
}

func resourceResourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPConn()

	userPoolID, identifier, err := DecodeResourceServerID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Resource Server (%s): %s", d.Id(), err)
	}

	params := &cognitoidentityprovider.DeleteResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	}

	_, err = conn.DeleteResourceServerWithContext(ctx, params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Resource Server (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeResourceServerID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format UserPoolID|Identifier, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
