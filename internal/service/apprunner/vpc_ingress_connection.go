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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIngressConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCIngressConnectionCreate,
		ReadWithoutTimeout:   resourceVPCIngressConnectionRead,
		UpdateWithoutTimeout: resourceVPCIngressConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCIngressConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ingress_vpc_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_endpoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCIngressConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &apprunner.CreateVpcIngressConnectionInput{
		ServiceArn:               aws.String(d.Get("service_arn").(string)),
		VpcIngressConnectionName: aws.String(name),
	}

	if v, ok := d.GetOk("ingress_vpc_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.IngressVpcConfiguration = expandIngressVPCConfiguration(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateVpcIngressConnectionWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner VPC Ingress Configuration (%s): %w", name, err))
	}

	if output == nil || output.VpcIngressConnection == nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner VPC Ingress Configuration (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.VpcIngressConnection.VpcIngressConnectionArn))

	if err := WaitVPCIngressConnectionActive(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for App Runner VPC Ingress Configuration (%s) creation: %w", d.Id(), err))
	}

	return resourceVPCIngressConnectionRead(ctx, d, meta)
}

func resourceVPCIngressConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &apprunner.DescribeVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeVpcIngressConnectionWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner VPC Ingress Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner VPC Ingress Configuration (%s): %w", d.Id(), err))
	}

	if output == nil || output.VpcIngressConnection == nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner VPC Ingress Configuration (%s): empty output", d.Id()))
	}

	if aws.StringValue(output.VpcIngressConnection.Status) == VPCIngressConnectionStatusDeleted {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner VPC Ingress Configuration (%s): %s after creation", d.Id(), aws.StringValue(output.VpcIngressConnection.Status)))
		}
		log.Printf("[WARN] App Runner VPC Ingress Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	config := output.VpcIngressConnection
	arn := aws.StringValue(config.VpcIngressConnectionArn)

	d.Set("arn", arn)
	d.Set("service_arn", config.ServiceArn)
	d.Set("name", config.VpcIngressConnectionName)
	d.Set("status", config.Status)
	d.Set("domain_name", config.DomainName)

	if err := d.Set("ingress_vpc_configuration", flattenIngressVPCConfiguration(config.IngressVpcConfiguration)); err != nil {
		return diag.Errorf("error setting ingress_vpc_configuration: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for App Runner VPC Ingress Configuration (%s): %s", arn, err))
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceVPCIngressConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating App Runner VPC Ingress Configuration (%s) tags: %s", d.Get("arn").(string), err))
		}
	}

	return resourceVPCIngressConnectionRead(ctx, d, meta)
}

func resourceVPCIngressConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn()

	input := &apprunner.DeleteVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcIngressConnectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting App Runner VPC Ingress Configuration (%s): %w", d.Id(), err))
	}

	if err := WaitVPCIngressConnectionDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for App Runner VPC Ingress Configuration (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func expandIngressVPCConfiguration(l []interface{}) *apprunner.IngressVpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &apprunner.IngressVpcConfiguration{}

	if v, ok := m["vpc_id"].(string); ok && v != "" {
		configuration.VpcId = aws.String(v)
	}

	if v, ok := m["vpc_endpoint_id"].(string); ok && v != "" {
		configuration.VpcEndpointId = aws.String(v)
	}

	return configuration
}

func flattenIngressVPCConfiguration(ingressVpcConfiguration *apprunner.IngressVpcConfiguration) []interface{} {
	if ingressVpcConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vpc_id":          aws.StringValue(ingressVpcConfiguration.VpcId),
		"vpc_endpoint_id": aws.StringValue(ingressVpcConfiguration.VpcEndpointId),
	}

	return []interface{}{m}
}
