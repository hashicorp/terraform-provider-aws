package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	apigatewayv2VpcLinkStatusDeleted = "DELETED"
)

func resourceAwsApiGatewayV2VpcLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2VpcLinkCreate,
		Read:   resourceAwsApiGatewayV2VpcLinkRead,
		Update: resourceAwsApiGatewayV2VpcLinkUpdate,
		Delete: resourceAwsApiGatewayV2VpcLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGatewayV2VpcLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateVpcLinkInput{
		Name:             aws.String(d.Get("name").(string)),
		SecurityGroupIds: expandStringSet(d.Get("security_group_ids").(*schema.Set)),
		SubnetIds:        expandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 VPC Link: %s", req)
	resp, err := conn.CreateVpcLink(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 VPC Link: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpcLinkId))

	if err := waitForApigatewayv2VpcLinkCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsApiGatewayV2VpcLinkRead(d, meta)
}

func resourceAwsApiGatewayV2VpcLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	respRaw, state, err := apigatewayv2VpcLinkStateRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}
	if state == apigatewayv2VpcLinkStatusDeleted {
		log.Printf("[WARN] API Gateway v2 VPC Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	resp := respRaw.(*apigatewayv2.GetVpcLinkOutput)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", resp.Name)
	if err := d.Set("security_group_ids", flattenStringSet(resp.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	if err := d.Set("subnet_ids", flattenStringSet(resp.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(resp.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2VpcLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	if d.HasChange("name") {
		req := &apigatewayv2.UpdateVpcLinkInput{
			VpcLinkId: aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating API Gateway v2 VPC Link: %s", req)
		_, err := conn.UpdateVpcLink(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 VPC Link (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 VPC Link (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsApiGatewayV2VpcLinkRead(d, meta)
}

func resourceAwsApiGatewayV2VpcLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 VPC Link (%s)", d.Id())
	_, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	if err := waitForApigatewayv2VpcLinkDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func apigatewayv2VpcLinkStateRefresh(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetVpcLink(&apigatewayv2.GetVpcLinkInput{
			VpcLinkId: aws.String(vpcLinkId),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			return "", apigatewayv2VpcLinkStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.VpcLinkStatus), nil
	}
}

func waitForApigatewayv2VpcLinkCreation(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusPending},
		Target:  []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh: apigatewayv2VpcLinkStateRefresh(conn, vpcLinkId),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for API Gateway v2 VPC Link (%s) availability", vpcLinkId)
	_, err := stateConf.WaitForState()

	return err
}

func waitForApigatewayv2VpcLinkDeletion(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusDeleting},
		Target:  []string{apigatewayv2VpcLinkStatusDeleted},
		Refresh: apigatewayv2VpcLinkStateRefresh(conn, vpcLinkId),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for API Gateway v2 VPC Link (%s) deletion", vpcLinkId)
	_, err := stateConf.WaitForState()

	return err
}
