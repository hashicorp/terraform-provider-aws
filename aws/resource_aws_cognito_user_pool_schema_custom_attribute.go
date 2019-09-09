package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsCognitoUserPoolSchemaCustomAttribute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolSchemaCustomAttributCreate,
		Read:   resourceAwsCognitoUserPoolSchemaAttributRead,
		Update: resourceAwsCognitoUserPoolSchemaCustomAttributUpdate,
		Delete: resourceAwsCognitoUserPoolSchemaCustomAttributDelete,

		Schema: map[string]*schema.Schema{
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"schema": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCognitoUserPoolSchemaName,
						},
						"attribute_data_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								cognitoidentityprovider.AttributeDataTypeString,
								cognitoidentityprovider.AttributeDataTypeNumber,
								cognitoidentityprovider.AttributeDataTypeDateTime,
								cognitoidentityprovider.AttributeDataTypeBoolean,
							}, false),
						},
						"developer_only_attribute": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"mutable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"number_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"string_attribute_constraints": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_length": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_length": {
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
	}
}

func resourceAwsCognitoUserPoolSchemaCustomAttributCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	params := &cognitoidentityprovider.AddCustomAttributesInput{
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("schema"); ok {
		configs := v.(*schema.Set).List()
		params.CustomAttributes = expandCognitoUserPoolSchema(configs)
	}

	resp, err := conn.AddCustomAttributes(params)

	if err != nil {
		return errwrap.Wrapf("Error creating Cognito User Pool Custom Attribute: {{err}}", err)
	}

	log.Printf("[DEBUG] Created the custom attribute on the user pool: %s", resp.String())

	return resourceAwsCognitoUserPoolSchemaAttributRead(d, meta)
}

func resourceAwsCognitoUserPoolSchemaAttributRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool: %s", params)

	resp, err := conn.DescribeUserPool(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Cognito User Pool %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if resp.UserPool.AliasAttributes != nil {
		d.Set("alias_attributes", flattenStringList(resp.UserPool.AliasAttributes))
	}
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "cognito-idp",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("userpool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.SetId(*resp.UserPool.Id)
	return nil
}

func resourceAwsCognitoUserPoolSchemaCustomAttributUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("update custom attribute operation is not supported")
}
func resourceAwsCognitoUserPoolSchemaCustomAttributDelete(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("delete custom attribute operation is not supported")
}
