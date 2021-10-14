package apigatewayv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCLinkCreate,
		Read:   resourceVPCLinkRead,
		Update: resourceVPCLinkUpdate,
		Delete: resourceVPCLinkDelete,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &apigatewayv2.CreateVpcLinkInput{
		Name:             aws.String(d.Get("name").(string)),
		SecurityGroupIds: flex.ExpandStringSet(d.Get("security_group_ids").(*schema.Set)),
		SubnetIds:        flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             tags.IgnoreAws().Apigatewayv2Tags(),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 VPC Link: %s", req)
	resp, err := conn.CreateVpcLink(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 VPC Link: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpcLinkId))

	if _, err := WaitVPCLinkAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 deployment (%s) availability: %s", d.Id(), err)
	}

	return resourceVPCLinkRead(d, meta)
}

func resourceVPCLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, _, err := StatusVPCLink(conn, d.Id())()
	if tfawserr.ErrMessageContains(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 VPC Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*apigatewayv2.GetVpcLinkOutput)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", output.Name)
	if err := d.Set("security_group_ids", flex.FlattenStringSet(output.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	if err := d.Set("subnet_ids", flex.FlattenStringSet(output.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}

	tags := tftags.Apigatewayv2KeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPCLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 VPC Link (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceVPCLinkRead(d, meta)
}

func resourceVPCLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 VPC Link (%s)", d.Id())
	_, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})
	if tfawserr.ErrMessageContains(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	_, err = WaitVPCLinkDeleted(conn, d.Id())
	if tfawserr.ErrMessageContains(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
