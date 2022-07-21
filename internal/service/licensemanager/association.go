package licensemanager

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAssociationCreate,
		ReadWithoutTimeout:   resourceAssociationRead,
		DeleteWithoutTimeout: resourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"license_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn

	resourceArn := d.Get("resource_arn").(string)
	licenseConfigurationArn := d.Get("license_configuration_arn").(string)

	input := &licensemanager.UpdateLicenseSpecificationsForResourceInput{
		AddLicenseSpecifications: []*licensemanager.LicenseSpecification{{
			LicenseConfigurationArn: aws.String(licenseConfigurationArn),
		}},
		ResourceArn: aws.String(resourceArn),
	}

	log.Printf("[DEBUG] Creating License Manager Association: %s", input)

	_, err := conn.UpdateLicenseSpecificationsForResourceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating License Manager Association: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceArn, licenseConfigurationArn))

	return resourceAssociationRead(ctx, d, meta)
}

func resourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn

	resourceArn, licenseConfigurationArn, err := AssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	licenseSpecification, err := AssociationFindSpecification(ctx, conn, resourceArn, licenseConfigurationArn)

	if err != nil {
		return diag.Errorf("reading License Manager Association (%s): %s", d.Id(), err)
	}

	if licenseSpecification == nil {
		log.Printf("[WARN] License Manager association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("license_configuration_arn", licenseConfigurationArn)
	d.Set("resource_arn", resourceArn)

	return nil
}

func resourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LicenseManagerConn

	resourceArn, licenseConfigurationArn, err := AssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &licensemanager.UpdateLicenseSpecificationsForResourceInput{
		RemoveLicenseSpecifications: []*licensemanager.LicenseSpecification{{
			LicenseConfigurationArn: aws.String(licenseConfigurationArn),
		}},
		ResourceArn: aws.String(resourceArn),
	}

	log.Printf("[DEBUG] License Manager association: %s", input)

	_, err = conn.UpdateLicenseSpecificationsForResourceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("deleting License Manager Association (%s): %s", d.Id(), err)
	}

	return nil
}

func AssociationFindSpecification(ctx context.Context, conn *licensemanager.LicenseManager, resourceArn, licenseConfigurationArn string) (*licensemanager.LicenseSpecification, error) {
	opts := &licensemanager.ListLicenseSpecificationsForResourceInput{
		ResourceArn: aws.String(resourceArn),
	}

	for {
		resp, err := conn.ListLicenseSpecificationsForResourceWithContext(ctx, opts)

		if err != nil {
			return nil, err
		}

		for _, licenseSpecification := range resp.LicenseSpecifications {
			if aws.StringValue(licenseSpecification.LicenseConfigurationArn) == licenseConfigurationArn {
				return licenseSpecification, nil
			}
		}

		if len(resp.LicenseSpecifications) == 0 || resp.NextToken == nil {
			return nil, nil
		}

		opts.NextToken = resp.NextToken
	}
}

func AssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected License Manager Association ID in the form resource_arn,license_configuration_arn - received: %s", id)
	}
	return parts[0], parts[1], nil
}
