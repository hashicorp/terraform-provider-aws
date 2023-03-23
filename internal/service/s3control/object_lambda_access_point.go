package s3control

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceObjectLambdaAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectLambdaAccessPointCreate,
		Read:   resourceObjectLambdaAccessPointRead,
		Update: resourceObjectLambdaAccessPointUpdate,
		Delete: resourceObjectLambdaAccessPointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_features": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(s3control.ObjectLambdaAllowedFeature_Values(), false),
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
									"actions": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(s3control.ObjectLambdaTransformationConfigurationAction_Values(), false),
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
															"function_arn": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceObjectLambdaAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	name := d.Get("name").(string)
	resourceID := ObjectLambdaAccessPointCreateResourceID(accountID, name)

	input := &s3control.CreateAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandObjectLambdaConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating S3 Object Lambda Access Point: %s", input)
	_, err := conn.CreateAccessPointForObjectLambda(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Object Lambda Access Point (%s): %w", resourceID, err)
	}

	d.SetId(resourceID)

	return resourceObjectLambdaAccessPointRead(d, meta)
}

func resourceObjectLambdaAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	output, err := FindObjectLambdaAccessPointByAccountIDAndName(conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Object Lambda Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Object Lambda Access Point (%s): %w", d.Id(), err)
	}

	d.Set("account_id", accountID)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazons3objectlambda.html#amazons3objectlambda-resources-for-iam-policies.
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3-object-lambda",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", name),
	}.String()
	d.Set("arn", arn)
	if err := d.Set("configuration", []interface{}{flattenObjectLambdaConfiguration(output)}); err != nil {
		return fmt.Errorf("error setting configuration: %w", err)
	}
	d.Set("name", name)

	return nil
}

func resourceObjectLambdaAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &s3control.PutAccessPointConfigurationForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandObjectLambdaConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating S3 Object Lambda Access Point: %s", input)
	_, err = conn.PutAccessPointConfigurationForObjectLambda(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Object Lambda Access Point (%s): %w", d.Id(), err)
	}

	return resourceObjectLambdaAccessPointRead(d, meta)
}

func resourceObjectLambdaAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := ObjectLambdaAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Object Lambda Access Point: %s", d.Id())
	_, err = conn.DeleteAccessPointForObjectLambda(&s3control.DeleteAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Object Lambda Access Point (%s): %w", d.Id(), err)
	}

	return nil
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

func expandObjectLambdaConfiguration(tfMap map[string]interface{}) *s3control.ObjectLambdaConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.ObjectLambdaConfiguration{}

	if v, ok := tfMap["allowed_features"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowedFeatures = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["cloud_watch_metrics_enabled"].(bool); ok && v {
		apiObject.CloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["supporting_access_point"].(string); ok && v != "" {
		apiObject.SupportingAccessPoint = aws.String(v)
	}

	if v, ok := tfMap["transformation_configuration"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TransformationConfigurations = expandObjectLambdaTransformationConfigurations(v.List())
	}

	return apiObject
}

func expandObjectLambdaTransformationConfiguration(tfMap map[string]interface{}) *s3control.ObjectLambdaTransformationConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.ObjectLambdaTransformationConfiguration{}

	if v, ok := tfMap["actions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Actions = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["content_transformation"].([]interface{}); ok && len(v) > 0 {
		apiObject.ContentTransformation = expandObjectLambdaContentTransformation(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandObjectLambdaTransformationConfigurations(tfList []interface{}) []*s3control.ObjectLambdaTransformationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*s3control.ObjectLambdaTransformationConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandObjectLambdaTransformationConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandObjectLambdaContentTransformation(tfMap map[string]interface{}) *s3control.ObjectLambdaContentTransformation {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.ObjectLambdaContentTransformation{}

	if v, ok := tfMap["aws_lambda"].([]interface{}); ok && len(v) > 0 {
		apiObject.AwsLambda = expandLambdaTransformation(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLambdaTransformation(tfMap map[string]interface{}) *s3control.AwsLambdaTransformation {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.AwsLambdaTransformation{}

	if v, ok := tfMap["function_arn"].(string); ok && v != "" {
		apiObject.FunctionArn = aws.String(v)
	}

	if v, ok := tfMap["function_payload"].(string); ok && v != "" {
		apiObject.FunctionPayload = aws.String(v)
	}

	return apiObject
}

func flattenObjectLambdaConfiguration(apiObject *s3control.ObjectLambdaConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowedFeatures; v != nil {
		tfMap["allowed_features"] = aws.StringValueSlice(v)
	}

	if v := apiObject.CloudWatchMetricsEnabled; v != nil {
		tfMap["cloud_watch_metrics_enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.SupportingAccessPoint; v != nil {
		tfMap["supporting_access_point"] = aws.StringValue(v)
	}

	if v := apiObject.TransformationConfigurations; v != nil {
		tfMap["transformation_configuration"] = flattenObjectLambdaTransformationConfigurations(v)
	}

	return tfMap
}

func flattenObjectLambdaTransformationConfiguration(apiObject *s3control.ObjectLambdaTransformationConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Actions; v != nil {
		tfMap["actions"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ContentTransformation; v != nil {
		tfMap["content_transformation"] = []interface{}{flattenObjectLambdaContentTransformation(v)}
	}

	return tfMap
}

func flattenObjectLambdaTransformationConfigurations(apiObjects []*s3control.ObjectLambdaTransformationConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenObjectLambdaTransformationConfiguration(apiObject))
	}

	return tfList
}

func flattenObjectLambdaContentTransformation(apiObject *s3control.ObjectLambdaContentTransformation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AwsLambda; v != nil {
		tfMap["aws_lambda"] = []interface{}{flattenLambdaTransformation(v)}
	}

	return tfMap
}

func flattenLambdaTransformation(apiObject *s3control.AwsLambdaTransformation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FunctionArn; v != nil {
		tfMap["function_arn"] = aws.StringValue(v)
	}

	if v := apiObject.FunctionPayload; v != nil {
		tfMap["function_payload"] = aws.StringValue(v)
	}

	return tfMap
}
