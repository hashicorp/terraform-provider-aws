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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCustomVpcAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomVpcAssociationCreate,
		ReadWithoutTimeout:   resourceCustomVpcAssociationRead,
		DeleteWithoutTimeout: resourceCustomVpcAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vpc_connector_name": {
				Type:         schema.TypeString,
				Computed:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(4, 40),
			},
			"vpc_connector_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCustomVpcAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	vpcConnectorName := d.Get("vpc_connector_name").(string)
	subnets := flex.ExpandStringSet(d.Get("subnets").(*schema.Set))
	securityGroups := flex.ExpandStringSet(d.Get("security_groups").(*schema.Set))

	input := &apprunner.CreateVpcConnectorInput{
		VpcConnectorName: aws.String(vpcConnectorName),
		Subnets:          subnets,
		SecurityGroups:   securityGroups,
	}

	output, err := conn.CreateVpcConnectorWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating App Runner Custom VPC (%s): %w", vpcConnectorName, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error associating App Runner Custom VPC (%s): empty output", vpcConnectorName))
	}

	d.SetId(aws.StringValue(output.VpcConnector.VpcConnectorArn))

	if err := WaitAutoScalingConfigurationActive(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for associating App Runner Custom VPC(%s) creation: %w", d.Id(), err))
	}

	return resourceCustomVpcAssociationRead(ctx, d, meta)
}

func resourceCustomVpcAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	domainName, serviceArn, err := CustomDomainAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	customDomain, err := FindCustomDomain(ctx, conn, domainName, serviceArn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner Custom Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if customDomain == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner Custom Domain Association (%s): empty output after creation", d.Id()))
		}
		log.Printf("[WARN] App Runner Custom Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("certificate_validation_records", flattenAppRunnerCustomDomainCertificateValidationRecords(customDomain.CertificateValidationRecords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting certificate_validation_records: %w", err))
	}

	d.Set("domain_name", customDomain.DomainName)
	d.Set("enable_www_subdomain", customDomain.EnableWWWSubdomain)
	d.Set("service_arn", serviceArn)
	d.Set("status", customDomain.Status)

	return nil
}

func resourceCustomVpcAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	domainName, serviceArn, err := CustomDomainAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &apprunner.DisassociateCustomDomainInput{
		DomainName: aws.String(domainName),
		ServiceArn: aws.String(serviceArn),
	}

	_, err = conn.DisassociateCustomDomainWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error disassociating App Runner Custom Domain (%s) for Service (%s): %w", domainName, serviceArn, err))
	}

	if err := WaitCustomDomainAssociationDeleted(ctx, conn, domainName, serviceArn); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error waiting for App Runner Custom Domain Association (%s) deletion: %w", d.Id(), err))
	}

	return nil
}
