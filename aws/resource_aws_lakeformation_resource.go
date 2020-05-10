package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLakeFormationResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationResourceRegister,
		Read:   resourceAwsLakeFormationResourceDescribe,
		Delete: resourceAwsLakeFormationResourceDeregister,

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"use_service_linked_role": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLakeFormationResourceRegister(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)
	useServiceLinkedRole := d.Get("use_service_linked_role").(bool)

	input := &lakeformation.RegisterResourceInput{
		ResourceArn:          aws.String(resourceArn),
		UseServiceLinkedRole: aws.Bool(useServiceLinkedRole),
	}
	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	_, err := conn.RegisterResource(input)
	if err != nil {
		return fmt.Errorf("Error registering LakeFormation Resource: %s", err)
	}

	d.SetId(fmt.Sprintf("lakeformation:resource:%s", resourceArn))

	return resourceAwsLakeFormationResourceDescribe(d, meta)
}

func resourceAwsLakeFormationResourceDescribe(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)

	input := &lakeformation.DescribeResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	out, err := conn.DescribeResource(input)
	if err != nil {
		return fmt.Errorf("Error reading LakeFormation Resource: %s", err)
	}

	d.Set("resource_arn", resourceArn)
	d.Set("role_arn", out.ResourceInfo.RoleArn)
	if out.ResourceInfo.LastModified != nil {
		d.Set("last_modified", out.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return nil
}

func resourceAwsLakeFormationResourceDeregister(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)

	input := &lakeformation.DeregisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	_, err := conn.DeregisterResource(input)
	if err != nil {
		return fmt.Errorf("Error deregistering LakeFormation Resource: %s", err)
	}

	return nil
}
