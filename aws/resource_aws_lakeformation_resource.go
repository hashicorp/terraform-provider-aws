package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsLakeFormationResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationResourceCreate,
		Read:   resourceAwsLakeFormationResourceRead,
		Update: resourceAwsLakeFormationResourceUpdate,
		Delete: resourceAwsLakeFormationResourceDelete,

		Schema: map[string]*schema.Schema{
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsLakeFormationResourceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)

	input := &lakeformation.RegisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	} else {
		input.UseServiceLinkedRole = aws.Bool(true)
	}

	_, err := conn.RegisterResource(input)

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeAlreadyExistsException) {
		log.Printf("[WARN] Lake Formation Resource (%s) already exists", resourceArn)
	} else if err != nil {
		return fmt.Errorf("error registering Lake Formation Resource (%s): %s", resourceArn, err)
	}

	d.SetId(resourceArn)
	return resourceAwsLakeFormationResourceRead(d, meta)
}

func resourceAwsLakeFormationResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)

	input := &lakeformation.DescribeResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.DescribeResource(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Lake Formation Resource (%s): %w", d.Id(), err)
	}

	if output == nil || output.ResourceInfo == nil {
		return fmt.Errorf("error getting Lake Formation Resource (%s): empty response", d.Id())
	}

	if err != nil {
		return fmt.Errorf("error reading Lake Formation Resource (%s): %w", d.Id(), err)
	}

	// d.Set("resource_arn", output.ResourceInfo.ResourceArn) // output not including resource arn currently
	d.Set("role_arn", output.ResourceInfo.RoleArn)
	if output.ResourceInfo.LastModified != nil { // output not including last modified currently
		d.Set("last_modified", output.ResourceInfo.LastModified.Format(time.RFC3339))
	}

	return nil
}

func resourceAwsLakeFormationResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("role_arn"); !ok {
		return resourceAwsLakeFormationResourceCreate(d, meta)
	}

	conn := meta.(*AWSClient).lakeformationconn

	input := &lakeformation.UpdateResourceInput{
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
		RoleArn:     aws.String(d.Get("role_arn").(string)),
	}

	_, err := conn.UpdateResource(input)
	if err != nil {
		return fmt.Errorf("error updating Lake Formation Resource (%s): %w", d.Id(), err)
	}

	return resourceAwsLakeFormationResourceRead(d, meta)
}

func resourceAwsLakeFormationResourceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	resourceArn := d.Get("resource_arn").(string)

	input := &lakeformation.DeregisterResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	_, err := conn.DeregisterResource(input)
	if err != nil {
		return fmt.Errorf("error deregistering Lake Formation Resource (%s): %w", d.Id(), err)
	}

	return nil
}
