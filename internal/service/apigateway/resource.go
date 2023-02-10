package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		UpdateWithoutTimeout: resourceResourceUpdate,
		DeleteWithoutTimeout: resourceResourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				d.Set("rest_api_id", restApiID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_part": {
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

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := &apigateway.CreateResourceInput{
		ParentId:  aws.String(d.Get("parent_id").(string)),
		PathPart:  aws.String(d.Get("path_part").(string)),
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
	}

	output, err := conn.CreateResourceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Resource: %s", err)
	}

	d.SetId(aws.StringValue(output.Id))

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	resource, err := FindResourceByTwoPartKey(ctx, conn, d.Id(), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Resource (%s): %s", d.Id(), err)
	}

	d.Set("parent_id", resource.ParentId)
	d.Set("path_part", resource.PathPart)
	d.Set("path", resource.Path)

	return diags
}

func resourceResourceUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("path_part") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/pathPart"),
			Value: aws.String(d.Get("path_part").(string)),
		})
	}

	if d.HasChange("parent_id") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/parentId"),
			Value: aws.String(d.Get("parent_id").(string)),
		})
	}
	return operations
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := &apigateway.UpdateResourceInput{
		ResourceId:      aws.String(d.Id()),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: resourceResourceUpdateOperations(d),
	}

	_, err := conn.UpdateResourceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Resource (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	log.Printf("[DEBUG] Deleting API Gateway Resource: %s", d.Id())
	_, err := conn.DeleteResourceWithContext(ctx, &apigateway.DeleteResourceInput{
		ResourceId: aws.String(d.Id()),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Resource (%s): %s", d.Id(), err)
	}

	return diags
}

func FindResourceByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, resourceID, apiID string) (*apigateway.Resource, error) {
	input := &apigateway.GetResourceInput{
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
	}

	output, err := conn.GetResourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
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
