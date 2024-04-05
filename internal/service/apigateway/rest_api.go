// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_rest_api", name="REST API")
// @Tags(identifierAttribute="arn")
func ResourceRestAPI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRestAPICreate,
		ReadWithoutTimeout:   resourceRestAPIRead,
		UpdateWithoutTimeout: resourceRestAPIUpdate,
		DeleteWithoutTimeout: resourceRestAPIDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("put_rest_api_mode", apigateway.PutModeOverwrite)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"api_key_source": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(apigateway.ApiKeySourceType_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"binary_media_types": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disable_execute_api_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(apigateway.EndpointType_Values(), false),
							},
						},
						"vpc_endpoint_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fail_on_warnings": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"minimum_compression_size": {
				Type:         nullable.TypeNullableInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(-1, 10485760),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// suppress null trigger when value is already null
					return old == "" && new == "-1"
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"put_rest_api_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      apigateway.PutModeOverwrite,
				ValidateFunc: validation.StringInSlice(apigateway.PutMode_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "" && new == apigateway.PutModeOverwrite {
						return true
					}
					return false
				},
			},
			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRestAPICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	name := d.Get("name").(string)
	input := &apigateway.CreateRestApiInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("api_key_source"); ok {
		input.ApiKeySource = aws.String(v.(string))
	}

	if v, ok := d.GetOk("binary_media_types"); ok {
		input.BinaryMediaTypes = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok {
		input.DisableExecuteApiEndpoint = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		input.EndpointConfiguration = expandEndpointConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("minimum_compression_size"); ok && v.(string) != "" && v.(string) != "-1" {
		mcs, err := strconv.Atoi(v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "converting minimum_compression_size (%s): %s", v, err)
		}
		input.MinimumCompressionSize = aws.Int64(int64(mcs))
	}

	if v, ok := d.GetOk("policy"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		input.Policy = aws.String(policy)
	}

	output, err := conn.CreateRestApiWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway REST API (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	if body, ok := d.GetOk("body"); ok {
		// Terraform implementation uses the `overwrite` mode by default.
		// Overwrite mode will delete existing literal properties if they are not explicitly set in the OpenAPI definition.
		// The VPC endpoints deletion and immediate recreation can cause a race condition.
		// 		Impacted properties: ApiKeySourceType, BinaryMediaTypes, Description, EndpointConfiguration, MinimumCompressionSize, Name, Policy
		// The `merge` mode will not delete literal properties of a RestApi if they’re not explicitly set in the OAS definition.
		input := &apigateway.PutRestApiInput{
			Body:      []byte(body.(string)),
			Mode:      aws.String(modeConfigOrDefault(d)),
			RestApiId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("fail_on_warnings"); ok {
			input.FailOnWarnings = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
			input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
		}

		api, err := conn.PutRestApiWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating API Gateway REST API (%s) specification: %s", d.Id(), err)
		}

		// Using PutRestApi with mode overwrite will remove any configuration
		// that was done with CreateRestApi. Reconcile these changes by having
		// any Terraform configured values overwrite imported configuration.
		if operations := resourceRestAPIWithBodyUpdateOperations(d, api); len(operations) > 0 {
			input := &apigateway.UpdateRestApiInput{
				PatchOperations: operations,
				RestApiId:       aws.String(d.Id()),
			}

			_, err := conn.UpdateRestApiWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s) after OpenAPI import: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceRestAPIRead(ctx, d, meta)...)
}

func resourceRestAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	api, err := FindRESTAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		log.Printf("[WARN] API Gateway REST API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API (%s): %s", d.Id(), err)
	}

	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	}

	err = conn.GetResourcesPagesWithContext(ctx, input, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			if aws.StringValue(v.Path) == "/" {
				d.Set("root_resource_id", v.Id)
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API (%s) resources: %s", d.Id(), err)
	}

	d.Set("api_key_source", api.ApiKeySource)
	apiARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("arn", apiARN)
	d.Set("binary_media_types", api.BinaryMediaTypes)
	d.Set("created_date", api.CreatedDate.Format(time.RFC3339))
	d.Set("description", api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(api.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	executionARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionARN)
	if api.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", nil)
	} else {
		d.Set("minimum_compression_size", strconv.FormatInt(aws.Int64Value(api.MinimumCompressionSize), 10))
	}
	d.Set("name", api.Name)

	// The API returns policy as an escaped JSON string
	// {\\\"Version\\\":\\\"2012-10-17\\\",...}
	// The string must be normalized before unquoting as it may contain escaped
	// forward slashes in CIDR blocks, which will break strconv.Unquote

	// I'm not sure why it needs to be wrapped with double quotes first, but it does
	normalized_policy, err := structure.NormalizeJsonString(`"` + aws.StringValue(api.Policy) + `"`)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "normalizing policy JSON: %s", err)
	}

	policy, err := strconv.Unquote(normalized_policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unescaping policy: %s", err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), policy)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	setTagsOut(ctx, api.Tags)

	return diags
}

func resourceRestAPIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		operations := make([]*apigateway.PatchOperation, 0)

		if d.HasChange("api_key_source") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/apiKeySource"),
				Value: aws.String(d.Get("api_key_source").(string)),
			})
		}

		if d.HasChange("binary_media_types") {
			o, n := d.GetChange("binary_media_types")
			prefix := "binaryMediaTypes"

			old := o.([]interface{})
			new := n.([]interface{})

			// Remove every binary media types. Simpler to remove and add new ones,
			// since there are no replacings.
			for _, v := range old {
				if e, ok := v.(string); ok {
					operations = append(operations, &apigateway.PatchOperation{
						Op:   aws.String(apigateway.OpRemove),
						Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJSONPointer(e))),
					})
				}
			}

			// Handle additions
			if len(new) > 0 {
				for _, v := range new {
					if e, ok := v.(string); ok {
						operations = append(operations, &apigateway.PatchOperation{
							Op:   aws.String(apigateway.OpAdd),
							Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJSONPointer(e))),
						})
					}
				}
			}
		}

		if d.HasChange("description") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			})
		}

		if d.HasChange("disable_execute_api_endpoint") {
			value := strconv.FormatBool(d.Get("disable_execute_api_endpoint").(bool))
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/disableExecuteApiEndpoint"),
				Value: aws.String(value),
			})
		}

		if d.HasChange("endpoint_configuration.0.types") {
			// The REST API must have an endpoint type.
			// If attempting to remove the configuration, do nothing.
			if v, ok := d.GetOk("endpoint_configuration"); ok && len(v.([]interface{})) > 0 {
				m := v.([]interface{})[0].(map[string]interface{})

				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpReplace),
					Path:  aws.String("/endpointConfiguration/types/0"),
					Value: aws.String(m["types"].([]interface{})[0].(string)),
				})
			}
		}

		// Compare the old and new values, don't blindly remove as they can cause race conditions with DNS and endpoint creation
		if d.HasChange("endpoint_configuration.0.vpc_endpoint_ids") {
			o, n := d.GetChange("endpoint_configuration.0.vpc_endpoint_ids")
			prefix := "/endpointConfiguration/vpcEndpointIds"

			old := o.(*schema.Set).List()
			new := n.(*schema.Set).List()

			for _, v := range old {
				for _, x := range new {
					if v.(string) == x.(string) {
						break
					}
				}
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpRemove),
					Path:  aws.String(prefix),
					Value: aws.String(v.(string)),
				})
			}

			for _, v := range new {
				for _, x := range old {
					if v.(string) == x.(string) {
						break
					}
				}
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpAdd),
					Path:  aws.String(prefix),
					Value: aws.String(v.(string)),
				})
			}
		}

		if d.HasChange("minimum_compression_size") {
			v := d.Get("minimum_compression_size").(string)
			value := aws.String(v)
			if v == "-1" {
				value = nil
			}
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/minimumCompressionSize"),
				Value: value,
			})
		}

		if d.HasChange("name") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/name"),
				Value: aws.String(d.Get("name").(string)),
			})
		}

		if d.HasChange("policy") {
			policy, _ := structure.NormalizeJsonString(d.Get("policy").(string)) // validation covers error

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/policy"),
				Value: aws.String(policy),
			})
		}

		if len(operations) > 0 {
			_, err := conn.UpdateRestApiWithContext(ctx, &apigateway.UpdateRestApiInput{
				PatchOperations: operations,
				RestApiId:       aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s): %s", d.Id(), err)
			}
		}

		if d.HasChanges("body", "parameters") {
			if body, ok := d.GetOk("body"); ok {
				// Terraform implementation uses the `overwrite` mode by default.
				// Overwrite mode will delete existing literal properties if they are not explicitly set in the OpenAPI definition.
				// The VPC endpoints deletion and immediate recreation can cause a race condition.
				// 		Impacted properties: ApiKeySourceType, BinaryMediaTypes, Description, EndpointConfiguration, MinimumCompressionSize, Name, Policy
				// The `merge` mode will not delete literal properties of a RestApi if they’re not explicitly set in the OAS definition.
				input := &apigateway.PutRestApiInput{
					Body:      []byte(body.(string)),
					Mode:      aws.String(modeConfigOrDefault(d)),
					RestApiId: aws.String(d.Id()),
				}

				if v, ok := d.GetOk("fail_on_warnings"); ok {
					input.FailOnWarnings = aws.Bool(v.(bool))
				}

				if v, ok := d.GetOk("parameters"); ok && len(v.(map[string]interface{})) > 0 {
					input.Parameters = flex.ExpandStringMap(v.(map[string]interface{}))
				}

				output, err := conn.PutRestApiWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s) specification: %s", d.Id(), err)
				}

				// Using PutRestApi with mode overwrite will remove any configuration
				// that was done previously. Reconcile these changes by having
				// any Terraform configured values overwrite imported configuration.
				if operations := resourceRestAPIWithBodyUpdateOperations(d, output); len(operations) > 0 {
					input := &apigateway.UpdateRestApiInput{
						PatchOperations: operations,
						RestApiId:       aws.String(d.Id()),
					}

					_, err := conn.UpdateRestApiWithContext(ctx, input)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s) after OpenAPI import: %s", d.Id(), err)
					}
				}
			}
		}
	}

	return append(diags, resourceRestAPIRead(ctx, d, meta)...)
}

func resourceRestAPIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway REST API: %s", d.Id())
	_, err := conn.DeleteRestApiWithContext(ctx, &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway REST API (%s): %s", d.Id(), err)
	}

	return diags
}

func FindRESTAPIByID(ctx context.Context, conn *apigateway.APIGateway, id string) (*apigateway.RestApi, error) {
	input := &apigateway.GetRestApiInput{
		RestApiId: aws.String(id),
	}

	output, err := conn.GetRestApiWithContext(ctx, input)

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

func resourceRestAPIWithBodyUpdateOperations(d *schema.ResourceData, output *apigateway.RestApi) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	if v, ok := d.GetOk("api_key_source"); ok && v.(string) != aws.StringValue(output.ApiKeySource) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/apiKeySource"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk("binary_media_types"); ok && len(v.([]interface{})) > 0 {
		if len(output.BinaryMediaTypes) > 0 {
			for _, elem := range aws.StringValueSlice(output.BinaryMediaTypes) {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String("/binaryMediaTypes/" + escapeJSONPointer(elem)),
				})
			}
		}

		for _, elem := range v.([]interface{}) {
			if el, ok := elem.(string); ok {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpAdd),
					Path: aws.String("/binaryMediaTypes/" + escapeJSONPointer(el)),
				})
			}
		}
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != aws.StringValue(output.Description) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok && v.(bool) != aws.BoolValue(output.DisableExecuteApiEndpoint) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/disableExecuteApiEndpoint"),
			Value: aws.String(strconv.FormatBool(v.(bool))),
		})
	}

	// Compare the defined values to the output values, don't blindly remove as they can cause race conditions with DNS and endpoint creation
	if v, ok := d.GetOk("endpoint_configuration"); ok {
		endpointConfiguration := expandEndpointConfiguration(v.([]interface{}))
		prefix := "/endpointConfiguration/vpcEndpointIds"
		if endpointConfiguration != nil && len(endpointConfiguration.VpcEndpointIds) > 0 {
			if output.EndpointConfiguration != nil {
				for _, v := range output.EndpointConfiguration.VpcEndpointIds {
					for _, x := range endpointConfiguration.VpcEndpointIds {
						if aws.StringValue(v) == aws.StringValue(x) {
							break
						}
					}
					operations = append(operations, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpRemove),
						Path:  aws.String(prefix),
						Value: v,
					})
				}
			}

			for _, v := range endpointConfiguration.VpcEndpointIds {
				for _, x := range output.EndpointConfiguration.VpcEndpointIds {
					if aws.StringValue(v) == aws.StringValue(x) {
						break
					}
				}
				operations = append(operations, &apigateway.PatchOperation{
					Op:    aws.String(apigateway.OpAdd),
					Path:  aws.String(prefix),
					Value: v,
				})
			}
		}
	}

	if v, ok := d.GetOk("minimum_compression_size"); ok && v.(string) != strconv.FormatInt(aws.Int64Value(output.MinimumCompressionSize), 10) {
		value := aws.String(v.(string))
		if v.(string) == "-1" {
			value = nil
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/minimumCompressionSize"),
			Value: value,
		})
	}

	if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(output.Name) {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/name"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk("policy"); ok {
		if equivalent, err := awspolicy.PoliciesAreEquivalent(v.(string), aws.StringValue(output.Policy)); err != nil || !equivalent {
			policy, _ := structure.NormalizeJsonString(v.(string)) // validation covers error

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/policy"),
				Value: aws.String(policy),
			})
		}
	}

	return operations
}

func modeConfigOrDefault(d *schema.ResourceData) string {
	if v, ok := d.GetOk("put_rest_api_mode"); ok {
		return v.(string)
	} else {
		return apigateway.PutModeOverwrite
	}
}

func expandEndpointConfiguration(l []interface{}) *apigateway.EndpointConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	endpointConfiguration := &apigateway.EndpointConfiguration{
		Types: flex.ExpandStringList(m["types"].([]interface{})),
	}

	if endpointIds, ok := m["vpc_endpoint_ids"]; ok {
		endpointConfiguration.VpcEndpointIds = flex.ExpandStringSet(endpointIds.(*schema.Set))
	}

	return endpointConfiguration
}

func flattenEndpointConfiguration(endpointConfiguration *apigateway.EndpointConfiguration) []interface{} {
	if endpointConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"types": flex.FlattenStringList(endpointConfiguration.Types),
	}

	if len(endpointConfiguration.VpcEndpointIds) > 0 {
		m["vpc_endpoint_ids"] = aws.StringValueSlice(endpointConfiguration.VpcEndpointIds)
	}

	return []interface{}{m}
}
