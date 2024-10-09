// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_object_lambda_access_point")
func resourceObjectLambdaAccessPoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceObjectLambdaAccessPointCreate,
		ReadWithoutTimeout:   resourceObjectLambdaAccessPointRead,
		UpdateWithoutTimeout: resourceObjectLambdaAccessPointUpdate,
		DeleteWithoutTimeout: resourceObjectLambdaAccessPointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_features": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.ObjectLambdaAllowedFeature](),
							},
						},
						"cloud_watch_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"supporting_access_point": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"transformation_configuration": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrActions: {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[types.ObjectLambdaTransformationConfigurationAction](),
										},
									},
									"content_transformation": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aws_lambda": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrFunctionARN: {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
															"function_payload": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceObjectLambdaAccessPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	name := d.Get(names.AttrName).(string)
	id := ObjectLambdaAccessPointCreateResourceID(accountID, name)
	input := &s3control.CreateAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandObjectLambdaConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateAccessPointForObjectLambda(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Object Lambda Access Point (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceObjectLambdaAccessPointRead(ctx, d, meta)...)
}

func resourceObjectLambdaAccessPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputConfiguration, err := findObjectLambdaAccessPointConfigurationByTwoPartKey(ctx, conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Object Lambda Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Object Lambda Access Point (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazons3objectlambda.html#amazons3objectlambda-resources-for-iam-policies.
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3-object-lambda",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", name),
	}.String()
	d.Set(names.AttrARN, arn)
	if err := d.Set(names.AttrConfiguration, []interface{}{flattenObjectLambdaConfiguration(outputConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	d.Set(names.AttrName, name)

	outputAlias, err := findObjectLambdaAccessPointAliasByTwoPartKey(ctx, conn, accountID, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Object Lambda Access Point (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAlias, outputAlias.Value)

	return diags
}

func resourceObjectLambdaAccessPointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutAccessPointConfigurationForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandObjectLambdaConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err = conn.PutAccessPointConfigurationForObjectLambda(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Object Lambda Access Point (%s): %s", d.Id(), err)
	}

	return append(diags, resourceObjectLambdaAccessPointRead(ctx, d, meta)...)
}

func resourceObjectLambdaAccessPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Object Lambda Access Point: %s", d.Id())
	_, err = conn.DeleteAccessPointForObjectLambda(ctx, &s3control.DeleteAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Object Lambda Access Point (%s): %s", d.Id(), err)
	}

	return diags
}

func findObjectLambdaAccessPointConfigurationByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*types.ObjectLambdaConfiguration, error) {
	input := &s3control.GetAccessPointConfigurationForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPointConfigurationForObjectLambda(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Configuration, nil
}

func findObjectLambdaAccessPointAliasByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*types.ObjectLambdaAccessPointAlias, error) {
	input := &s3control.GetAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPointForObjectLambda(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Alias == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Alias, nil
}

const objectLambdaAccessPointResourceIDSeparator = ":"

func ObjectLambdaAccessPointCreateResourceID(accountID, accessPointName string) string {
	parts := []string{accountID, accessPointName}
	id := strings.Join(parts, objectLambdaAccessPointResourceIDSeparator)

	return id
}

func ObjectLambdaAccessPointParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, objectLambdaAccessPointResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]saccess-point-name", id, objectLambdaAccessPointResourceIDSeparator)
}

func expandObjectLambdaConfiguration(tfMap map[string]interface{}) *types.ObjectLambdaConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ObjectLambdaConfiguration{}

	if v, ok := tfMap["allowed_features"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowedFeatures = flex.ExpandStringyValueSet[types.ObjectLambdaAllowedFeature](v)
	}

	if v, ok := tfMap["cloud_watch_metrics_enabled"].(bool); ok && v {
		apiObject.CloudWatchMetricsEnabled = v
	}

	if v, ok := tfMap["supporting_access_point"].(string); ok && v != "" {
		apiObject.SupportingAccessPoint = aws.String(v)
	}

	if v, ok := tfMap["transformation_configuration"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TransformationConfigurations = expandObjectLambdaTransformationConfigurations(v.List())
	}

	return apiObject
}

func expandObjectLambdaTransformationConfiguration(tfMap map[string]interface{}) *types.ObjectLambdaTransformationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ObjectLambdaTransformationConfiguration{}

	if v, ok := tfMap[names.AttrActions].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Actions = flex.ExpandStringyValueSet[types.ObjectLambdaTransformationConfigurationAction](v)
	}

	if v, ok := tfMap["content_transformation"].([]interface{}); ok && len(v) > 0 {
		apiObject.ContentTransformation = expandObjectLambdaContentTransformation(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandObjectLambdaTransformationConfigurations(tfList []interface{}) []types.ObjectLambdaTransformationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ObjectLambdaTransformationConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandObjectLambdaTransformationConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandObjectLambdaContentTransformation(tfMap map[string]interface{}) types.ObjectLambdaContentTransformation {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ObjectLambdaContentTransformationMemberAwsLambda{}

	if v, ok := tfMap["aws_lambda"].([]interface{}); ok && len(v) > 0 {
		apiObject.Value = expandLambdaTransformation(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLambdaTransformation(tfMap map[string]interface{}) types.AwsLambdaTransformation {
	apiObject := types.AwsLambdaTransformation{}

	if v, ok := tfMap[names.AttrFunctionARN].(string); ok && v != "" {
		apiObject.FunctionArn = aws.String(v)
	}

	if v, ok := tfMap["function_payload"].(string); ok && v != "" {
		apiObject.FunctionPayload = aws.String(v)
	}

	return apiObject
}

func flattenObjectLambdaConfiguration(apiObject *types.ObjectLambdaConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cloud_watch_metrics_enabled": apiObject.CloudWatchMetricsEnabled,
	}

	if v := apiObject.AllowedFeatures; v != nil {
		tfMap["allowed_features"] = v
	}

	if v := apiObject.SupportingAccessPoint; v != nil {
		tfMap["supporting_access_point"] = aws.ToString(v)
	}

	if v := apiObject.TransformationConfigurations; v != nil {
		tfMap["transformation_configuration"] = flattenObjectLambdaTransformationConfigurations(v)
	}

	return tfMap
}

func flattenObjectLambdaTransformationConfiguration(apiObject types.ObjectLambdaTransformationConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap[names.AttrActions] = v
	}

	if v := apiObject.ContentTransformation; v != nil {
		tfMap["content_transformation"] = []interface{}{flattenObjectLambdaContentTransformation(v)}
	}

	return tfMap
}

func flattenObjectLambdaTransformationConfigurations(apiObjects []types.ObjectLambdaTransformationConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenObjectLambdaTransformationConfiguration(apiObject))
	}

	return tfList
}

func flattenObjectLambdaContentTransformation(apiObject types.ObjectLambdaContentTransformation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v, ok := apiObject.(*types.ObjectLambdaContentTransformationMemberAwsLambda); ok {
		tfMap["aws_lambda"] = []interface{}{flattenLambdaTransformation(v.Value)}
	}

	return tfMap
}

func flattenLambdaTransformation(apiObject types.AwsLambdaTransformation) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.FunctionArn; v != nil {
		tfMap[names.AttrFunctionARN] = aws.ToString(v)
	}

	if v := apiObject.FunctionPayload; v != nil {
		tfMap["function_payload"] = aws.ToString(v)
	}

	return tfMap
}
