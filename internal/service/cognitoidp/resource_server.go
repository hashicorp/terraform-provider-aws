// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_resource_server", name="Resource Server")
func resourceResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceServerCreate,
		ReadWithoutTimeout:   resourceResourceServerRead,
		UpdateWithoutTimeout: resourceResourceServerUpdate,
		DeleteWithoutTimeout: resourceResourceServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrScope: {
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
			names.AttrUserPoolID: {
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
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	identifier := d.Get(names.AttrIdentifier).(string)
	userPoolID := d.Get(names.AttrUserPoolID).(string)
	id := resourceServerCreateResourceID(userPoolID, identifier)
	input := &cognitoidentityprovider.CreateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get(names.AttrName).(string)),
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk(names.AttrScope); ok {
		input.Scopes = expandResourceServerScopeTypes(v.(*schema.Set).List())
	}

	_, err := conn.CreateResourceServer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito Resource Server (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceResourceServerRead(ctx, d, meta)...)
}

func resourceResourceServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, identifier, err := resourceServerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resourceServer, err := findResourceServerByTwoPartKey(ctx, conn, userPoolID, identifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito Resource Server %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito Resource Server (%s): %s", d.Id(), err)
	}

	identifier = aws.ToString(resourceServer.Identifier)
	d.Set(names.AttrIdentifier, identifier)
	d.Set(names.AttrName, resourceServer.Name)
	scopes := flattenResourceServerScopeTypes(resourceServer.Scopes)
	if err := d.Set(names.AttrScope, scopes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting scope: %s", err)
	}
	d.Set("scope_identifiers", tfslices.ApplyToAll(scopes, func(tfMap map[string]interface{}) string {
		return identifier + "/" + tfMap["scope_name"].(string)
	}))
	d.Set(names.AttrUserPoolID, resourceServer.UserPoolId)

	return diags
}

func resourceResourceServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, identifier, err := resourceServerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &cognitoidentityprovider.UpdateResourceServerInput{
		Identifier: aws.String(identifier),
		Name:       aws.String(d.Get(names.AttrName).(string)),
		Scopes:     expandResourceServerScopeTypes(d.Get(names.AttrScope).(*schema.Set).List()),
		UserPoolId: aws.String(userPoolID),
	}

	_, err = conn.UpdateResourceServer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Cognito Resource Server (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResourceServerRead(ctx, d, meta)...)
}

func resourceResourceServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, identifier, err := resourceServerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Cognito Resource Server: %s", d.Id())
	_, err = conn.DeleteResourceServer(ctx, &cognitoidentityprovider.DeleteResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito Resource Server (%s): %s", d.Id(), err)
	}

	return diags
}

const resourceServerResourceIDSeparator = "|"

func resourceServerCreateResourceID(userPoolID, identifier string) string {
	parts := []string{userPoolID, identifier}
	id := strings.Join(parts, resourceServerResourceIDSeparator)

	return id
}

func resourceServerParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, resourceServerResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sIdentifier", id, resourceServerResourceIDSeparator)
}

func findResourceServerByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, identifier string) (*awstypes.ResourceServerType, error) {
	input := &cognitoidentityprovider.DescribeResourceServerInput{
		Identifier: aws.String(identifier),
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.DescribeResourceServer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourceServer == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourceServer, nil
}

func expandResourceServerScopeTypes(tfList []interface{}) []awstypes.ResourceServerScopeType {
	apiObjects := make([]awstypes.ResourceServerScopeType, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := awstypes.ResourceServerScopeType{}

		if v, ok := tfMap["scope_description"]; ok {
			apiObject.ScopeDescription = aws.String(v.(string))
		}

		if v, ok := tfMap["scope_name"]; ok {
			apiObject.ScopeName = aws.String(v.(string))
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func flattenResourceServerScopeTypes(apiObjects []awstypes.ResourceServerScopeType) []map[string]interface{} {
	tfList := make([]map[string]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"scope_description": aws.ToString(apiObject.ScopeDescription),
			"scope_name":        aws.ToString(apiObject.ScopeName),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
