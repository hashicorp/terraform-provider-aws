package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsServiceCatalogProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProductCreate,
		Read:   resourceAwsServiceCatalogProductRead,
		Update: resourceAwsServiceCatalogProductUpdate,
		Delete: resourceAwsServiceCatalogProductDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"distributor": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"product_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"support_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"support_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"support_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provisioning_artifact": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Required: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"load_template_from_url": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Only check for a diff on initial create.
								if d.Id() != "" {
									return true
								}
								return false
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsServiceCatalogProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.CreateProductInput{}

	if v, ok := d.GetOk("name"); ok {
		now := time.Now()
		input.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distributor"); ok {
		input.Distributor = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner"); ok {
		input.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_type"); ok {
		input.ProductType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_description"); ok {
		input.SupportDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_email"); ok {
		input.SupportEmail = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_url"); ok {
		input.SupportUrl = aws.String(v.(string))
	}

	artifactSettings := d.Get("provisioning_artifact").(*schema.Set).List()[0].(map[string]interface{})
	artifactProperties := servicecatalog.ProvisioningArtifactProperties{}
	artifactProperties.Description = aws.String(artifactSettings["description"].(string))
	artifactProperties.Name = aws.String(artifactSettings["name"].(string))
	artifactProperties.Type = aws.String("CLOUD_FORMATION_TEMPLATE")

	url := aws.String(artifactSettings["load_template_from_url"].(string))
	info := map[string]*string{
		"LoadTemplateFromURL": url,
	}
	artifactProperties.Info = info
	input.SetProvisioningArtifactParameters(&artifactProperties)

	log.Printf("[DEBUG] Creating Service Catalog Product: %s %s", input, artifactProperties)
	resp, err := conn.CreateProduct(&input)
	if err != nil {
		return fmt.Errorf("Creating ServiceCatalog product failed: %s", err.Error())
	}
	d.SetId(*resp.ProductViewDetail.ProductViewSummary.ProductId)

	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DescribeProductAsAdminInput{}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Reading Service Catalog Product: %s", input)
	resp, err := conn.DescribeProductAsAdmin(&input)
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Service Catalog Product %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading ServiceCatalog product '%s' failed: %s", *input.Id, err.Error())
	}

	d.Set("product_arn", resp.ProductViewDetail.ProductARN)

	pvs := resp.ProductViewDetail.ProductViewSummary
	d.Set("description", pvs.ShortDescription)
	d.Set("distributor", pvs.Distributor)
	d.Set("name", pvs.Name)
	d.Set("owner", pvs.Owner)
	d.Set("product_type", pvs.Type)
	d.Set("support_description", pvs.SupportDescription)
	d.Set("support_email", pvs.SupportEmail)
	d.Set("support_url", pvs.SupportUrl)

	provisioningArtifactSummary := getProvisioningArtifactSummary(resp)
	var a []map[string]interface{}
	artifact := make(map[string]interface{})
	artifact["description"] = *provisioningArtifactSummary.Description
	artifact["id"] = *provisioningArtifactSummary.Id
	artifact["load_template_from_url"] = "only_used_on_initial_create"
	artifact["name"] = *provisioningArtifactSummary.Name
	a = append(a, artifact)

	if err := d.Set("provisioning_artifact", a); err != nil {
		return err
	}
	return nil
}

// Lookup the first artifact, which was the one created on inital build, by comparing created at time
func getProvisioningArtifactSummary(resp *servicecatalog.DescribeProductAsAdminOutput) *servicecatalog.ProvisioningArtifactSummary {
	firstPas := resp.ProvisioningArtifactSummaries[0]
	for _, pas := range resp.ProvisioningArtifactSummaries {
		if pas.CreatedTime.UnixNano() < firstPas.CreatedTime.UnixNano() {
			firstPas = pas
		}
	}
	return firstPas
}

func resourceAwsServiceCatalogProductUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdateProductInput{}
	input.Id = aws.String(d.Id())

	if d.HasChange("description") {
		v, _ := d.GetOk("description")
		input.Description = aws.String(v.(string))
	}

	if d.HasChange("distributor") {
		v, _ := d.GetOk("distributor")
		input.Distributor = aws.String(v.(string))
	}

	if d.HasChange("name") {
		v, _ := d.GetOk("name")
		input.Name = aws.String(v.(string))
	}

	if d.HasChange("owner") {
		v, _ := d.GetOk("owner")
		input.Owner = aws.String(v.(string))
	}

	if d.HasChange("support_description") {
		v, _ := d.GetOk("support_description")
		input.SupportDescription = aws.String(v.(string))
	}

	if d.HasChange("support_email") {
		v, _ := d.GetOk("support_email")
		input.SupportEmail = aws.String(v.(string))
	}

	if d.HasChange("support_url") {
		v, _ := d.GetOk("support_url")
		input.SupportUrl = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Update Service Catalog Product: %s", input)
	_, err := conn.UpdateProduct(&input)
	if err != nil {
		return fmt.Errorf("Updating ServiceCatalog product '%s' failed: %s", *input.Id, err.Error())
	}
	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DeleteProductInput{}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Delete Service Catalog Product: %s", input)
	_, err := conn.DeleteProduct(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog product '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
