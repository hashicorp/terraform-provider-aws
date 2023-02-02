package apprunner

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCConnectorCreate,
		ReadWithoutTimeout:   resourceVPCConnectorRead,
		UpdateWithoutTimeout: resourceVPCConnectorUpdate,
		DeleteWithoutTimeout: resourceVPCConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_connector_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(4, 40),
			},
			"vpc_connector_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	vpcConnectorName := d.Get("vpc_connector_name").(string)
	input := &apprunner.CreateVpcConnectorInput{
		SecurityGroups:   flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		Subnets:          flex.ExpandStringSet(d.Get("subnets").(*schema.Set)),
		VpcConnectorName: aws.String(vpcConnectorName),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateVpcConnectorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner VPC Connector (%s): %s", vpcConnectorName, err)
	}

	d.SetId(aws.StringValue(output.VpcConnector.VpcConnectorArn))

	if err := waitVPCConnectorCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner VPC Connector (%s) create: %s", d.Id(), err)
	}

	return resourceVPCConnectorRead(ctx, d, meta)
}

func resourceVPCConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpcConnector, err := FindVPCConnectorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner VPC Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner VPC Connector (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(vpcConnector.VpcConnectorArn)
	d.Set("arn", vpcConnector.VpcConnectorArn)
	d.Set("security_groups", aws.StringValueSlice(vpcConnector.SecurityGroups))
	d.Set("status", vpcConnector.Status)
	d.Set("subnets", aws.StringValueSlice(vpcConnector.Subnets))
	d.Set("vpc_connector_name", vpcConnector.VpcConnectorName)
	d.Set("vpc_connector_revision", vpcConnector.VpcConnectorRevision)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for App Runner VPC Connector (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceVPCConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating App Runner VPC Connector (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceVPCConnectorRead(ctx, d, meta)
}

func resourceVPCConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

	log.Printf("[DEBUG] Deleting App Runner VPC Connector: %s", d.Id())
	_, err := conn.DeleteVpcConnectorWithContext(ctx, &apprunner.DeleteVpcConnectorInput{
		VpcConnectorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner VPC Connector (%s): %s", d.Id(), err)
	}

	if err := waitVPCConnectorDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner VPC Connector (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindVPCConnectorByARN(ctx context.Context, conn *apprunner.AppRunner, arn string) (*apprunner.VpcConnector, error) {
	input := &apprunner.DescribeVpcConnectorInput{
		VpcConnectorArn: aws.String(arn),
	}

	output, err := findVPCConnector(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == apprunner.VpcConnectorStatusInactive {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCConnector(ctx context.Context, conn *apprunner.AppRunner, input *apprunner.DescribeVpcConnectorInput) (*apprunner.VpcConnector, error) {
	output, err := conn.DescribeVpcConnectorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcConnector == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcConnector, nil
}

func statusVPCConnector(ctx context.Context, conn *apprunner.AppRunner, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCConnectorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	vpcConnectorCreateTimeout = 2 * time.Minute
	vpcConnectorDeleteTimeout = 2 * time.Minute
)

func waitVPCConnectorCreated(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Target:  []string{apprunner.VpcConnectorStatusActive},
		Refresh: statusVPCConnector(ctx, conn, arn),
		Timeout: vpcConnectorCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCConnectorDeleted(ctx context.Context, conn *apprunner.AppRunner, arn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apprunner.VpcConnectorStatusActive},
		Target:  []string{},
		Refresh: statusVPCConnector(ctx, conn, arn),
		Timeout: vpcConnectorDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
