// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_documentation_part", name="Documentation Part")
func resourceDocumentationPart() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDocumentationPartCreate,
		ReadWithoutTimeout:   resourceDocumentationPartRead,
		UpdateWithoutTimeout: resourceDocumentationPartUpdate,
		DeleteWithoutTimeout: resourceDocumentationPartDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"documentation_part_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrStatusCode: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrProperties: {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDocumentationPartCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	input := &apigateway.CreateDocumentationPartInput{
		Location:   expandDocumentationPartLocation(d.Get(names.AttrLocation).([]interface{})),
		Properties: aws.String(d.Get(names.AttrProperties).(string)),
		RestApiId:  aws.String(apiID),
	}

	output, err := conn.CreateDocumentationPart(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Documentation Part: %s", err)
	}

	d.SetId(documentationPartCreateResourceID(apiID, aws.ToString(output.Id)))

	return append(diags, resourceDocumentationPartRead(ctx, d, meta)...)
}

func resourceDocumentationPartRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationPartID, err := documentationPartParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	docPart, err := findDocumentationPartByTwoPartKey(ctx, conn, apiID, documentationPartID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Documentation Part (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Documentation Part (%s): %s", d.Id(), err)
	}

	d.Set("documentation_part_id", docPart.Id)
	d.Set(names.AttrLocation, flattenDocumentationPartLocation(docPart.Location))
	d.Set(names.AttrProperties, docPart.Properties)
	d.Set("rest_api_id", apiID)

	return diags
}

func resourceDocumentationPartUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationPartID, err := documentationPartParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &apigateway.UpdateDocumentationPartInput{
		DocumentationPartId: aws.String(documentationPartID),
		RestApiId:           aws.String(apiID),
	}
	operations := make([]types.PatchOperation, 0)

	if d.HasChange(names.AttrProperties) {
		properties := d.Get(names.AttrProperties).(string)
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/properties"),
			Value: aws.String(properties),
		})
	}

	input.PatchOperations = operations

	_, err = conn.UpdateDocumentationPart(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Documentation Part (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDocumentationPartRead(ctx, d, meta)...)
}

func resourceDocumentationPartDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID, documentationPartID, err := documentationPartParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting API Gateway Documentation Part: %s", d.Id())
	_, err = conn.DeleteDocumentationPart(ctx, &apigateway.DeleteDocumentationPartInput{
		DocumentationPartId: aws.String(documentationPartID),
		RestApiId:           aws.String(apiID),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Documentation Part (%s): %s", d.Id(), err)
	}

	return diags
}

func findDocumentationPartByTwoPartKey(ctx context.Context, conn *apigateway.Client, apiID, documentationPartID string) (*apigateway.GetDocumentationPartOutput, error) {
	input := &apigateway.GetDocumentationPartInput{
		DocumentationPartId: aws.String(documentationPartID),
		RestApiId:           aws.String(apiID),
	}

	output, err := conn.GetDocumentationPart(ctx, input)

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

const documentationPartResourceIDSeparator = "/"

func documentationPartCreateResourceID(apiID, documentationPartID string) string {
	parts := []string{apiID, documentationPartID}
	id := strings.Join(parts, documentationPartResourceIDSeparator)

	return id
}

func documentationPartParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, documentationPartResourceIDSeparator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%[1]s), expected REST-API-ID%[2]sID", id, documentationPartResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func expandDocumentationPartLocation(l []interface{}) *types.DocumentationPartLocation {
	if len(l) == 0 {
		return nil
	}
	loc := l[0].(map[string]interface{})
	out := &types.DocumentationPartLocation{
		Type: types.DocumentationPartType(loc[names.AttrType].(string)),
	}
	if v, ok := loc["method"]; ok {
		out.Method = aws.String(v.(string))
	}
	if v, ok := loc[names.AttrName]; ok {
		out.Name = aws.String(v.(string))
	}
	if v, ok := loc[names.AttrPath]; ok {
		out.Path = aws.String(v.(string))
	}
	if v, ok := loc[names.AttrStatusCode]; ok {
		out.StatusCode = aws.String(v.(string))
	}
	return out
}

func flattenDocumentationPartLocation(l *types.DocumentationPartLocation) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := l.Method; v != nil {
		m["method"] = aws.ToString(v)
	}

	if v := l.Name; v != nil {
		m[names.AttrName] = aws.ToString(v)
	}

	if v := l.Path; v != nil {
		m[names.AttrPath] = aws.ToString(v)
	}

	if v := l.StatusCode; v != nil {
		m[names.AttrStatusCode] = aws.ToString(v)
	}

	m[names.AttrType] = string(l.Type)

	return []interface{}{m}
}
