// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_resource", name="Resource")
func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		UpdateWithoutTimeout: resourceResourceUpdate,
		DeleteWithoutTimeout: resourceResourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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
			"full_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"parent_id", "path_part"},
				ValidateFunc: func(val any, key string) (warns []string, errs []error) {
					v := val.(string)
					if !strings.HasPrefix(v, "/") {
						errs = append(errs, fmt.Errorf("%q must start with '/'", key))
					}
					if strings.HasSuffix(v, "/") && v != "/" {
						errs = append(errs, fmt.Errorf("%q must not end with '/' unless it is the root '/'", key))
					}
					return
				},
			},
			"parent_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"full_path"},
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_part": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"full_path"},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf(names.AttrPath, func(ctx context.Context, diff *schema.ResourceDiff, meta any) bool {
				return diff.HasChange("path_part") || diff.HasChange("full_path")
			}),
			func(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
				fullPath := diff.Get("full_path").(string)
				parentID := diff.Get("parent_id").(string)
				pathPart := diff.Get("path_part").(string)

				// Require either full_path OR (parent_id AND path_part)
				if fullPath == "" && (parentID == "" || pathPart == "") {
					return fmt.Errorf("either 'full_path' or both 'parent_id' and 'path_part' must be specified")
				}

				return nil
			},
		),
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)
	restApiID := d.Get("rest_api_id").(string)

	if fullPath := d.Get("full_path").(string); fullPath != "" {
		// Handle full_path creation logic
		resourceID, err := createResourcesFromFullPath(ctx, conn, restApiID, fullPath)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating API Gateway Resource from full path %q: %s", fullPath, err)
		}
		d.SetId(resourceID)
	} else {
		// Handle traditional parent_id + path_part creation
		input := apigateway.CreateResourceInput{
			ParentId:  aws.String(d.Get("parent_id").(string)),
			PathPart:  aws.String(d.Get("path_part").(string)),
			RestApiId: aws.String(restApiID),
		}

		output, err := conn.CreateResource(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating API Gateway Resource: %s", err)
		}

		d.SetId(aws.ToString(output.Id))
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	resource, err := findResourceByTwoPartKey(ctx, conn, d.Id(), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] API Gateway Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Resource (%s): %s", d.Id(), err)
	}

	// Set common attributes
	d.Set(names.AttrPath, resource.Path)

	// Set attributes based on creation mode
	if fullPath := d.Get("full_path").(string); fullPath != "" {
		// For full_path mode, only set the computed path
		d.Set("full_path", aws.ToString(resource.Path))
	} else {
		// For traditional mode, set parent_id and path_part
		d.Set("parent_id", resource.ParentId)
		d.Set("path_part", resource.PathPart)
	}

	return diags
}

func resourceResourceUpdateOperations(d *schema.ResourceData) []types.PatchOperation {
	operations := make([]types.PatchOperation, 0)
	
	// Only allow updates for traditional mode (parent_id + path_part)
	// full_path mode forces recreation via ForceNew
	if d.HasChange("path_part") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/pathPart"),
			Value: aws.String(d.Get("path_part").(string)),
		})
	}

	if d.HasChange("parent_id") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/parentId"),
			Value: aws.String(d.Get("parent_id").(string)),
		})
	}
	return operations
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := apigateway.UpdateResourceInput{
		ResourceId:      aws.String(d.Id()),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: resourceResourceUpdateOperations(d),
	}

	_, err := conn.UpdateResource(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Resource (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Resource: %s", d.Id())
	input := apigateway.DeleteResourceInput{
		ResourceId: aws.String(d.Id()),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	}
	_, err := conn.DeleteResource(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Resource (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourceByTwoPartKey(ctx context.Context, conn *apigateway.Client, resourceID, apiID string) (*apigateway.GetResourceOutput, error) {
	input := apigateway.GetResourceInput{
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
	}

	output, err := conn.GetResource(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
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

// createResourcesFromFullPath creates all necessary parent resources and returns the final resource ID
func createResourcesFromFullPath(ctx context.Context, conn *apigateway.Client, restApiID, fullPath string) (string, error) {
	// Get root resource ID first
	restAPI, err := conn.GetRestApi(ctx, &apigateway.GetRestApiInput{
		RestApiId: aws.String(restApiID),
	})
	if err != nil {
		return "", fmt.Errorf("getting REST API %s: %w", restApiID, err)
	}

	rootResourceID := aws.ToString(restAPI.RootResourceId)

	// Handle root path special case
	if fullPath == "/" {
		return rootResourceID, nil
	}

	// Split path into segments
	pathSegments := strings.Split(strings.Trim(fullPath, "/"), "/")
	
	// Check if resources already exist and create missing ones
	currentParentID := rootResourceID
	
	for _, segment := range pathSegments {
		// Try to find existing resource
		existingResource, err := findChildResource(ctx, conn, restApiID, currentParentID, segment)
		if err == nil {
			// Resource already exists, use it as parent for next iteration
			currentParentID = aws.ToString(existingResource.Id)
			continue
		}
		
		// Resource doesn't exist, create it
		input := apigateway.CreateResourceInput{
			ParentId:  aws.String(currentParentID),
			PathPart:  aws.String(segment),
			RestApiId: aws.String(restApiID),
		}

		output, err := conn.CreateResource(ctx, &input)
		if err != nil {
			return "", fmt.Errorf("creating resource for path segment %q: %w", segment, err)
		}
		
		currentParentID = aws.ToString(output.Id)
	}

	return currentParentID, nil
}

// findChildResource looks for a child resource with the given path part under the specified parent
func findChildResource(ctx context.Context, conn *apigateway.Client, restApiID, parentID, pathPart string) (*apigateway.GetResourceOutput, error) {
	// List all resources in the API
	resources, err := conn.GetResources(ctx, &apigateway.GetResourcesInput{
		RestApiId: aws.String(restApiID),
	})
	if err != nil {
		return nil, fmt.Errorf("listing resources: %w", err)
	}

	// Find the resource with matching parent and path part
	for _, resource := range resources.Items {
		if aws.ToString(resource.ParentId) == parentID && aws.ToString(resource.PathPart) == pathPart {
			// Found matching resource, get its details
			return conn.GetResource(ctx, &apigateway.GetResourceInput{
				RestApiId:  aws.String(restApiID),
				ResourceId: resource.Id,
			})
		}
	}

	return nil, fmt.Errorf("resource not found")
}
