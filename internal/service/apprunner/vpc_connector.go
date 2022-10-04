package apprunner

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceVPCConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCConnectorCreate,
		ReadWithoutTimeout:   resourceVPCConnectorRead,
		DeleteWithoutTimeout: resourceVPCConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vpc_connector_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(4, 40),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tftags.TagsSchemaComputed(),
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_connector_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceVPCConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	vpcConnectorName := d.Get("vpc_connector_name").(string)
	subnets := flex.ExpandStringSet(d.Get("subnets").(*schema.Set))
	securityGroups := flex.ExpandStringSet(d.Get("security_groups").(*schema.Set))

	input := &apprunner.CreateVpcConnectorInput{
		VpcConnectorName: aws.String(vpcConnectorName),
		Subnets:          subnets,
		SecurityGroups:   securityGroups,
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateVpcConnectorWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner vpc (%s): %w", vpcConnectorName, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner vpc (%s): empty output", vpcConnectorName))
	}

	d.SetId(aws.StringValue(output.VpcConnector.VpcConnectorArn))

	if err := waitVPCConnectorActive(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for creating App Runner vpc (%s) creation: %w", d.Id(), err))
	}

	return resourceVPCConnectorRead(ctx, d, meta)
}

func resourceVPCConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &apprunner.DescribeVpcConnectorInput{
		VpcConnectorArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeVpcConnectorWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner vpc connector (%s): %w", d.Id(), err))
	}

	if output == nil || output.VpcConnector == nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner vpc connector (%s): empty output", d.Id()))
	}

	if aws.StringValue(output.VpcConnector.Status) == apprunner.VpcConnectorStatusInactive {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner vpc connector (%s): %s after creation", d.Id(), aws.StringValue(output.VpcConnector.Status)))
		}
		log.Printf("[WARN] App Runner vpc connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	vpcConnector := output.VpcConnector
	arn := aws.StringValue(vpcConnector.VpcConnectorArn)

	d.Set("vpc_connector_name", vpcConnector.VpcConnectorName)
	d.Set("vpc_connector_revision", vpcConnector.VpcConnectorRevision)
	d.Set("arn", vpcConnector.VpcConnectorArn)
	d.Set("status", vpcConnector.Status)

	var subnets []string
	for _, sn := range vpcConnector.Subnets {
		subnets = append(subnets, aws.StringValue(sn))
	}
	if err := d.Set("subnets", subnets); err != nil {
		return diag.FromErr(fmt.Errorf("Error saving Subnet IDs to state for App Runner vpc connector (%s): %s", d.Id(), err))
	}

	var securityGroups []string
	for _, sn := range vpcConnector.SecurityGroups {
		securityGroups = append(securityGroups, aws.StringValue(sn))
	}
	if err := d.Set("security_groups", securityGroups); err != nil {
		return diag.FromErr(fmt.Errorf("Error saving securityGroup IDs to state for App Runner vpc connector (%s): %s", d.Id(), err))
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return diag.Errorf("error listing tags for App Runner vpc connector (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}

func resourceVPCConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	input := &apprunner.DeleteVpcConnectorInput{
		VpcConnectorArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcConnectorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting App Runner vpc connector (%s): %w", d.Id(), err))
	}

	if err := waitVPCConnectorInactive(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error waiting for App Runner vpc connector (%s) deletion: %w", d.Id(), err))
	}

	return nil
}
