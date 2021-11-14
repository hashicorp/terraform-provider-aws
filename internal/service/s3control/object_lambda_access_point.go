package s3control

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
							Type:     schema.TypeList,
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
													Optional: true,
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

	configuration := &s3control.ObjectLambdaConfiguration{
		AllowedFeatures:          flex.ExpandStringSet(d.Get("allowed_features").(*schema.Set)),
		CloudWatchMetricsEnabled: aws.Bool(d.Get("cloud_watch_metrics_enabled").(bool)),
		SupportingAccessPoint:    aws.String(d.Get("supporting_access_point").(string)),
		//TransformationConfigurations: expandObjectLambdaTransformationConfiguration(d.Get("transformation_configurations").([]*interface{})),
	}

	input := &s3control.CreateAccessPointForObjectLambdaInput{
		AccountId:     aws.String(accountID),
		Configuration: configuration,
		Name:          aws.String(name),
	}

	log.Printf("[DEBUG] Creating S3 Object Lambda Access Point: %s", input)
	output, err := conn.CreateAccessPointForObjectLambda(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Control Access Point (%s): %w", name, err)
	}

	if output == nil {
		return fmt.Errorf("error creating S3 Control Access Point (%s): empty response", name)
	}

	parsedARN, err := arn.Parse(aws.StringValue(output.ObjectLambdaAccessPointArn))

	if err == nil && strings.HasPrefix(parsedARN.Resource, "outpost/") {
		d.SetId(aws.StringValue(output.ObjectLambdaAccessPointArn))
		name = aws.StringValue(output.ObjectLambdaAccessPointArn)
	} else {
		d.SetId(fmt.Sprintf("%s:%s", accountID, name))
	}

	return resourceObjectLambdaAccessPointRead(d, meta)
}

func resourceObjectLambdaAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountId, name, err := S3ObjectLambdaAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	output, err := conn.GetAccessPoint(&s3control.GetAccessPointInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		log.Printf("[WARN] S3 Object Lambda Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Object Lambda Access Point (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading S3 Object Lambda Access Point (%s): empty response", d.Id())
	}

	if strings.HasPrefix(name, "arn:") {
		parsedAccessPointARN, err := arn.Parse(name)

		if err != nil {
			return fmt.Errorf("error parsing S3 Control Access Point ARN (%s): %w", name, err)
		}

		bucketARN := arn.ARN{
			AccountID: parsedAccessPointARN.AccountID,
			Partition: parsedAccessPointARN.Partition,
			Region:    parsedAccessPointARN.Region,
			Resource: strings.Replace(
				parsedAccessPointARN.Resource,
				fmt.Sprintf("accesspoint/%s", aws.StringValue(output.Name)),
				fmt.Sprintf("bucket/%s", aws.StringValue(output.Bucket)),
				1,
			),
			Service: parsedAccessPointARN.Service,
		}

		d.Set("arn", name)
		d.Set("bucket", bucketARN.String())
	} else {
		accessPointARN := arn.ARN{
			AccountID: accountId,
			Partition: meta.(*conns.AWSClient).Partition,
			Region:    meta.(*conns.AWSClient).Region,
			Resource:  fmt.Sprintf("accesspoint/%s", aws.StringValue(output.Name)),
			Service:   "s3",
		}

		d.Set("arn", accessPointARN.String())
		d.Set("bucket", output.Bucket)
	}

	d.Set("account_id", accountId)
	d.Set("domain_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s-%s.s3-accesspoint", aws.StringValue(output.Name), accountId)))
	d.Set("name", output.Name)
	d.Set("network_origin", output.NetworkOrigin)
	if err := d.Set("public_access_block_configuration", flattenS3ObjectLambdaAccessPointPublicAccessBlockConfiguration(output.PublicAccessBlockConfiguration)); err != nil {
		return fmt.Errorf("error setting public_access_block_configuration: %s", err)
	}
	if err := d.Set("vpc_configuration", flattenS3ObjectLambdaAccessPointVpcConfiguration(output.VpcConfiguration)); err != nil {
		return fmt.Errorf("error setting vpc_configuration: %s", err)
	}

	// Return early since S3 on Outposts cannot have public policies
	if strings.HasPrefix(name, "arn:") {
		d.Set("has_public_access_policy", false)

		return nil
	}

	return nil
}

func resourceObjectLambdaAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceObjectLambdaAccessPointRead(d, meta)
}

func resourceObjectLambdaAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountId, name, err := S3ObjectLambdaAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Object Lambda Access Point: %s", d.Id())
	_, err = conn.DeleteAccessPointForObjectLambda(&s3control.DeleteAccessPointForObjectLambdaInput{
		AccountId: aws.String(accountId),
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

// func expandObjectLambdaContentTransformation(vConfig []interface{}) *s3control.ObjectLambdaContentTransformation {
// 	if len(vConfig) == 0 || vConfig[0] == nil {
// 		return nil
// 	}

// 	mConfig := vConfig[0].(map[string]interface{})

// 	return &s3control.ObjectLambdaContentTransformation{
// 		AwsLambda: &s3control.AwsLambdaTransformation{
// 			FunctionArn:     aws.String(mConfig["aws_lambda"]["function_arn"]),
// 			FunctionPayload: aws.String(mConfig["aws_lambda"]["function_payload"]),
// 		},
// 	}

// }

// func expandObjectLambdaTransformationConfiguration(vConfig []interface{}) *s3control.ObjectLambdaTransformationConfiguration {
// 	if len(vConfig) == 0 || vConfig[0] == nil {
// 		return nil
// 	}
// 	mConfig := vConfig[0].(map[string]interface{})

// 	return &s3control.ObjectLambdaTransformationConfiguration{
// 		Actions:               expandStringSet(mConfig["actions"].(*schema.Set)),
// 		ContentTransformation: expandObjectLambdaContentTransformation(mConfig["content_transformation"].([]interface{})),
// 	}
// }

// S3ObjectLambdaAccessPointParseId returns the Account ID and Access Point Name (S3) or ARN (S3 on Outposts)
func S3ObjectLambdaAccessPointParseId(id string) (string, string, error) {
	parsedARN, err := arn.Parse(id)

	if err == nil {
		return parsedARN.AccountID, id, nil
	}

	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected ACCOUNT_ID:NAME", id)
	}

	return parts[0], parts[1], nil
}

func expandS3ObjectLambdaAccessPointVpcConfiguration(vConfig []interface{}) *s3control.VpcConfiguration {
	if len(vConfig) == 0 || vConfig[0] == nil {
		return nil
	}

	mConfig := vConfig[0].(map[string]interface{})

	return &s3control.VpcConfiguration{
		VpcId: aws.String(mConfig["vpc_id"].(string)),
	}
}

func flattenS3ObjectLambdaAccessPointVpcConfiguration(config *s3control.VpcConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"vpc_id": aws.StringValue(config.VpcId),
	}}
}

func expandS3ObjectLambdaAccessPointPublicAccessBlockConfiguration(vConfig []interface{}) *s3control.PublicAccessBlockConfiguration {
	if len(vConfig) == 0 || vConfig[0] == nil {
		return nil
	}

	mConfig := vConfig[0].(map[string]interface{})

	return &s3control.PublicAccessBlockConfiguration{
		BlockPublicAcls:       aws.Bool(mConfig["block_public_acls"].(bool)),
		BlockPublicPolicy:     aws.Bool(mConfig["block_public_policy"].(bool)),
		IgnorePublicAcls:      aws.Bool(mConfig["ignore_public_acls"].(bool)),
		RestrictPublicBuckets: aws.Bool(mConfig["restrict_public_buckets"].(bool)),
	}
}

func flattenS3ObjectLambdaAccessPointPublicAccessBlockConfiguration(config *s3control.PublicAccessBlockConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"block_public_acls":       aws.BoolValue(config.BlockPublicAcls),
		"block_public_policy":     aws.BoolValue(config.BlockPublicPolicy),
		"ignore_public_acls":      aws.BoolValue(config.IgnorePublicAcls),
		"restrict_public_buckets": aws.BoolValue(config.RestrictPublicBuckets),
	}}
}
