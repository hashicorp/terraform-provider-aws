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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCustomDomainAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomDomainAssociationCreate,
		ReadWithoutTimeout:   resourceCustomDomainAssociationRead,
		DeleteWithoutTimeout: resourceCustomDomainAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"certificate_validation_records": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_target": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"enable_www_subdomain": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCustomDomainAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn

	domainName := d.Get("domain_name").(string)
	serviceArn := d.Get("service_arn").(string)

	input := &apprunner.AssociateCustomDomainInput{
		DomainName:         aws.String(domainName),
		EnableWWWSubdomain: aws.Bool(d.Get("enable_www_subdomain").(bool)),
		ServiceArn:         aws.String(serviceArn),
	}

	output, err := conn.AssociateCustomDomainWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error associating App Runner Custom Domain (%s) for Service (%s): %w", domainName, serviceArn, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error associating App Runner Custom Domain (%s) for Service (%s): empty output", domainName, serviceArn))
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(output.CustomDomain.DomainName), aws.StringValue(output.ServiceArn)))
	d.Set("dns_target", output.DNSTarget)

	if err := WaitCustomDomainAssociationCreated(ctx, conn, domainName, serviceArn); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for App Runner Custom Domain Association (%s) creation: %w", d.Id(), err))
	}

	return resourceCustomDomainAssociationRead(ctx, d, meta)
}

func resourceCustomDomainAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := d.Set("certificate_validation_records", flattenCustomDomainCertificateValidationRecords(customDomain.CertificateValidationRecords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting certificate_validation_records: %w", err))
	}

	d.Set("domain_name", customDomain.DomainName)
	d.Set("enable_www_subdomain", customDomain.EnableWWWSubdomain)
	d.Set("service_arn", serviceArn)
	d.Set("status", customDomain.Status)

	return nil
}

func resourceCustomDomainAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func flattenCustomDomainCertificateValidationRecords(records []*apprunner.CertificateValidationRecord) []interface{} {
	var results []interface{}

	for _, record := range records {
		if record == nil {
			continue
		}

		m := map[string]interface{}{
			"name":   aws.StringValue(record.Name),
			"status": aws.StringValue(record.Status),
			"type":   aws.StringValue(record.Type),
			"value":  aws.StringValue(record.Value),
		}

		results = append(results, m)
	}

	return results
}
