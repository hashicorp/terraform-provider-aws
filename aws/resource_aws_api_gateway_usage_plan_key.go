package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayUsagePlanKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayUsagePlanKeyCreate,
		Read:   resourceAwsApiGatewayUsagePlanKeyRead,
		Delete: resourceAwsApiGatewayUsagePlanKeyDelete,

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"key_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"usage_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsApiGatewayUsagePlanKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Print("[DEBUG] Creating API Gateway Usage Plan Key")

	params := &apigateway.CreateUsagePlanKeyInput{
		KeyId:       aws.String(d.Get("key_id").(string)),
		KeyType:     aws.String(d.Get("key_type").(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	}

	up, err := conn.CreateUsagePlanKey(params)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Usage Plan Key: %s", err)
	}

	d.SetId(*up.Id)

	return resourceAwsApiGatewayUsagePlanKeyRead(d, meta)
}

func resourceAwsApiGatewayUsagePlanKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Reading API Gateway Usage Plan Key: %s", d.Id())

	up, err := conn.GetUsagePlanKey(&apigateway.GetUsagePlanKeyInput{
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
		KeyId:       aws.String(d.Get("key_id").(string)),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] API Gateway Usage Plan Key (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", up.Name)
	d.Set("value", up.Value)

	return nil
}

func resourceAwsApiGatewayUsagePlanKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan Key: %s", d.Id())
	_, err := conn.DeleteUsagePlanKey(&apigateway.DeleteUsagePlanKeyInput{
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
		KeyId:       aws.String(d.Get("key_id").(string)),
	})
	if isAWSErr(err, "NotFoundException", "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting API Gateway usage plan key: %s", err)
	}

	return nil

}
