// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_rest_api", name="REST API")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetRestApiOutput", importIgnore="put_rest_api_mode")
func resourceRestAPI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRestAPICreate,
		ReadWithoutTimeout:   resourceRestAPIRead,
		UpdateWithoutTimeout: resourceRestAPIUpdate,
		DeleteWithoutTimeout: resourceRestAPIDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("put_rest_api_mode", types.PutModeOverwrite)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"api_key_source": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.ApiKeySourceType](),
			},
			names.AttrARN: {
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
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.EndpointType](),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrPolicy: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.PutModeOverwrite,
				ValidateDiagFunc: enum.Validate[types.PutMode](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "" && new == string(types.PutModeOverwrite) {
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateRestApiInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("api_key_source"); ok {
		input.ApiKeySource = types.ApiKeySourceType(v.(string))
	}

	if v, ok := d.GetOk("binary_media_types"); ok {
		input.BinaryMediaTypes = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok {
		input.DisableExecuteApiEndpoint = v.(bool)
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		input.EndpointConfiguration = expandEndpointConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("minimum_compression_size"); ok && v.(string) != "" && v.(string) != "-1" {
		mcs, err := strconv.ParseInt(v.(string), 0, 32)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "converting minimum_compression_size (%s): %s", v, err)
		}
		input.MinimumCompressionSize = aws.Int32(int32(mcs))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.Policy = aws.String(policy)
	}

	output, err := conn.CreateRestApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway REST API (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	if body, ok := d.GetOk("body"); ok {
		// Terraform implementation uses the `overwrite` mode by default.
		// Overwrite mode will delete existing literal properties if they are not explicitly set in the OpenAPI definition.
		// The VPC endpoints deletion and immediate recreation can cause a race condition.
		// 		Impacted properties: ApiKeySourceType, BinaryMediaTypes, Description, EndpointConfiguration, MinimumCompressionSize, Name, Policy
		// The `merge` mode will not delete literal properties of a RestApi if they’re not explicitly set in the OAS definition.
		input := &apigateway.PutRestApiInput{
			Body:      []byte(body.(string)),
			Mode:      types.PutMode(modeConfigOrDefault(d)),
			RestApiId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("fail_on_warnings"); ok {
			input.FailOnWarnings = v.(bool)
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
			input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}

		api, err := conn.PutRestApi(ctx, input)

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

			_, err := conn.UpdateRestApi(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s) after OpenAPI import: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceRestAPIRead(ctx, d, meta)...)
}

func resourceRestAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	api, err := findRestAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway REST API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API (%s): %s", d.Id(), err)
	}

	d.Set("api_key_source", api.ApiKeySource)
	d.Set(names.AttrARN, apiARN(meta.(*conns.AWSClient), d.Id()))
	d.Set("binary_media_types", api.BinaryMediaTypes)
	d.Set(names.AttrCreatedDate, api.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(api.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	d.Set("execution_arn", apiInvokeARN(meta.(*conns.AWSClient), d.Id()))
	if api.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", nil)
	} else {
		d.Set("minimum_compression_size", strconv.FormatInt(int64(aws.ToInt32(api.MinimumCompressionSize)), 10))
	}
	d.Set(names.AttrName, api.Name)

	input := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	}

	rootResource, err := findResource(ctx, conn, input, func(v *types.Resource) bool {
		return aws.ToString(v.Path) == "/"
	})

	switch {
	case err == nil:
		d.Set("root_resource_id", rootResource.Id)
	case tfresource.NotFound(err):
		d.Set("root_resource_id", nil)
	default:
		return sdkdiag.AppendErrorf(diags, "reading API Gateway REST API (%s) root resource: %s", d.Id(), err)
	}

	policy, err := flattenAPIPolicy(api.Policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	setTagsOut(ctx, api.Tags)

	return diags
}

func resourceRestAPIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		operations := make([]types.PatchOperation, 0)

		if d.HasChange("api_key_source") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
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
					operations = append(operations, types.PatchOperation{
						Op:   types.OpRemove,
						Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJSONPointer(e))),
					})
				}
			}

			// Handle additions
			if len(new) > 0 {
				for _, v := range new {
					if e, ok := v.(string); ok {
						operations = append(operations, types.PatchOperation{
							Op:   types.OpAdd,
							Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJSONPointer(e))),
						})
					}
				}
			}
		}

		if d.HasChange(names.AttrDescription) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/description"),
				Value: aws.String(d.Get(names.AttrDescription).(string)),
			})
		}

		if d.HasChange("disable_execute_api_endpoint") {
			value := strconv.FormatBool(d.Get("disable_execute_api_endpoint").(bool))
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/disableExecuteApiEndpoint"),
				Value: aws.String(value),
			})
		}

		if d.HasChange("endpoint_configuration.0.types") {
			// The REST API must have an endpoint type.
			// If attempting to remove the configuration, do nothing.
			if v, ok := d.GetOk("endpoint_configuration"); ok && len(v.([]interface{})) > 0 {
				m := v.([]interface{})[0].(map[string]interface{})

				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
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
				operations = append(operations, types.PatchOperation{
					Op:    types.OpRemove,
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
				operations = append(operations, types.PatchOperation{
					Op:    types.OpAdd,
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
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/minimumCompressionSize"),
				Value: value,
			})
		}

		if d.HasChange(names.AttrName) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/name"),
				Value: aws.String(d.Get(names.AttrName).(string)),
			})
		}

		if d.HasChange(names.AttrPolicy) {
			policy, _ := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string)) // validation covers error

			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/policy"),
				Value: aws.String(policy),
			})
		}

		if len(operations) > 0 {
			_, err := conn.UpdateRestApi(ctx, &apigateway.UpdateRestApiInput{
				PatchOperations: operations,
				RestApiId:       aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating API Gateway REST API (%s): %s", d.Id(), err)
			}
		}

		if d.HasChanges("body", names.AttrParameters) {
			if body, ok := d.GetOk("body"); ok {
				// Terraform implementation uses the `overwrite` mode by default.
				// Overwrite mode will delete existing literal properties if they are not explicitly set in the OpenAPI definition.
				// The VPC endpoints deletion and immediate recreation can cause a race condition.
				// 		Impacted properties: ApiKeySourceType, BinaryMediaTypes, Description, EndpointConfiguration, MinimumCompressionSize, Name, Policy
				// The `merge` mode will not delete literal properties of a RestApi if they’re not explicitly set in the OAS definition.
				input := &apigateway.PutRestApiInput{
					Body:      []byte(body.(string)),
					Mode:      types.PutMode(modeConfigOrDefault(d)),
					RestApiId: aws.String(d.Id()),
				}

				if v, ok := d.GetOk("fail_on_warnings"); ok {
					input.FailOnWarnings = v.(bool)
				}

				if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
					input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
				}

				output, err := conn.PutRestApi(ctx, input)

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

					_, err := conn.UpdateRestApi(ctx, input)

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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway REST API: %s", d.Id())
	_, err := conn.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway REST API (%s): %s", d.Id(), err)
	}

	return diags
}

func findRestAPIByID(ctx context.Context, conn *apigateway.Client, id string) (*apigateway.GetRestApiOutput, error) {
	input := &apigateway.GetRestApiInput{
		RestApiId: aws.String(id),
	}

	output, err := conn.GetRestApi(ctx, input)

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

func resourceRestAPIWithBodyUpdateOperations(d *schema.ResourceData, output *apigateway.PutRestApiOutput) []types.PatchOperation {
	operations := make([]types.PatchOperation, 0)

	if v, ok := d.GetOk("api_key_source"); ok && v.(string) != string(output.ApiKeySource) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/apiKeySource"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk("binary_media_types"); ok && len(v.([]interface{})) > 0 {
		if len(output.BinaryMediaTypes) > 0 {
			for _, elem := range output.BinaryMediaTypes {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpRemove,
					Path: aws.String("/binaryMediaTypes/" + escapeJSONPointer(elem)),
				})
			}
		}

		for _, elem := range v.([]interface{}) {
			if el, ok := elem.(string); ok {
				operations = append(operations, types.PatchOperation{
					Op:   types.OpAdd,
					Path: aws.String("/binaryMediaTypes/" + escapeJSONPointer(el)),
				})
			}
		}
	}

	if v, ok := d.GetOk(names.AttrDescription); ok && v.(string) != aws.ToString(output.Description) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/description"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk("disable_execute_api_endpoint"); ok && v.(bool) != output.DisableExecuteApiEndpoint {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
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
						if v == x {
							break
						}
					}
					operations = append(operations, types.PatchOperation{
						Op:    types.OpRemove,
						Path:  aws.String(prefix),
						Value: aws.String(v),
					})
				}
			}

			for _, v := range endpointConfiguration.VpcEndpointIds {
				for _, x := range output.EndpointConfiguration.VpcEndpointIds {
					if v == x {
						break
					}
				}
				operations = append(operations, types.PatchOperation{
					Op:    types.OpAdd,
					Path:  aws.String(prefix),
					Value: aws.String(v),
				})
			}
		}
	}

	if v, ok := d.GetOk("minimum_compression_size"); ok && v.(string) != strconv.FormatInt(int64(aws.ToInt32(output.MinimumCompressionSize)), 10) {
		value := aws.String(v.(string))
		if v.(string) == "-1" {
			value = nil
		}
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/minimumCompressionSize"),
			Value: value,
		})
	}

	if v, ok := d.GetOk(names.AttrName); ok && v.(string) != aws.ToString(output.Name) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/name"),
			Value: aws.String(v.(string)),
		})
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		if equivalent, err := awspolicy.PoliciesAreEquivalent(v.(string), aws.ToString(output.Policy)); err != nil || !equivalent {
			policy, _ := structure.NormalizeJsonString(v.(string)) // validation covers error

			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/policy"),
				Value: aws.String(policy),
			})
		}
	}

	return operations
}

// escapeJSONPointer escapes string per RFC 6901
// so it can be used as path in JSON patch operations
func escapeJSONPointer(path string) string {
	path = strings.Replace(path, "~", "~0", -1)
	path = strings.Replace(path, "/", "~1", -1)
	return path
}

func modeConfigOrDefault(d *schema.ResourceData) string {
	if v, ok := d.GetOk("put_rest_api_mode"); ok {
		return v.(string)
	} else {
		return string(types.PutModeOverwrite)
	}
}

func expandEndpointConfiguration(l []interface{}) *types.EndpointConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	endpointConfiguration := &types.EndpointConfiguration{
		Types: flex.ExpandStringyValueList[types.EndpointType](m["types"].([]interface{})),
	}

	if endpointIds, ok := m["vpc_endpoint_ids"]; ok {
		endpointConfiguration.VpcEndpointIds = flex.ExpandStringValueSet(endpointIds.(*schema.Set))
	}

	return endpointConfiguration
}

func flattenEndpointConfiguration(endpointConfiguration *types.EndpointConfiguration) []interface{} {
	if endpointConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"types": endpointConfiguration.Types,
	}

	if len(endpointConfiguration.VpcEndpointIds) > 0 {
		m["vpc_endpoint_ids"] = endpointConfiguration.VpcEndpointIds
	}

	return []interface{}{m}
}

func flattenAPIPolicy(apiObject *string) (string, error) {
	// The API returns policy as an escaped JSON string
	// {\\\"Version\\\":\\\"2012-10-17\\\",...}
	// The string must be normalized before unquoting as it may contain escaped
	// forward slashes in CIDR blocks, which will break strconv.Unquote

	// I'm not sure why it needs to be wrapped with double quotes first, but it does
	normalizedPolicy, err := structure.NormalizeJsonString(`"` + aws.ToString(apiObject) + `"`)
	if err != nil {
		return "", err
	}

	policy, err := strconv.Unquote(normalizedPolicy)
	if err != nil {
		return "", err
	}

	return policy, nil
}

func apiARN(c *conns.AWSClient, apiID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "apigateway",
		Region:    c.Region,
		Resource:  fmt.Sprintf("/restapis/%s", apiID),
	}.String()
}

func apiInvokeARN(c *conns.AWSClient, apiID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "execute-api",
		Region:    c.Region,
		AccountID: c.AccountID,
		Resource:  apiID,
	}.String()
}
