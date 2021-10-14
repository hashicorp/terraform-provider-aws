package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/apigateway/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_arns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &apigateway.CreateVpcLinkInput{
		Name:       aws.String(d.Get("name").(string)),
		TargetArns: flex.ExpandStringList(d.Get("target_arns").([]interface{})),
		Tags:       tags.IgnoreAws().ApigatewayTags(),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	resp, err := conn.CreateVpcLink(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.Id))

	if err := waiter.ApiGatewayVpcLinkAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway VPC Link (%s) availability after creation: %w", d.Id(), err)
	}

	return resourceVPCLinkRead(d, meta)
}

func resourceVPCLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &apigateway.GetVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	resp, err := conn.GetVpcLink(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] VPC Link %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	tags := tftags.ApigatewayKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("target_arns", flex.FlattenStringList(resp.TargetArns))
	return nil
}

func resourceVPCLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	input := &apigateway.UpdateVpcLinkInput{
		VpcLinkId:       aws.String(d.Id()),
		PatchOperations: operations,
	}

	_, err := conn.UpdateVpcLink(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] VPC Link %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := waiter.ApiGatewayVpcLinkAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway VPC Link (%s) availability after update: %w", d.Id(), err)
	}

	return resourceVPCLinkRead(d, meta)
}

func resourceVPCLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	input := &apigateway.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcLink(input)

	if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway VPC Link (%s): %w", d.Id(), err)
	}

	if err := waiter.ApiGatewayVpcLinkDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway VPC Link (%s) deletion: %w", d.Id(), err)
	}

	return nil
}
