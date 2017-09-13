package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"artifact_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"artifact_id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cloud_formation_template_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"distributor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"product_arn": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: false,
			},
			"product_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"support_description": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"support_email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"support_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
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

	artifactProperties := servicecatalog.ProvisioningArtifactProperties{Type: aws.String("CLOUD_FORMATION_TEMPLATE")}

	if v, ok := d.GetOk("artifact_description"); ok {
		artifactProperties.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("artifact_name"); ok {
		artifactProperties.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cloud_formation_template_url"); ok {
		info := map[string]*string{
			"LoadTemplateFromURL": aws.String(v.(string)),
		}
		artifactProperties.Info = info
	}

	input.SetProvisioningArtifactParameters(&artifactProperties)

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

	resp, err := conn.DescribeProductAsAdmin(&input)
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog product '%s' failed: %s", *input.Id, err.Error())
	}

	d.Set("product_arn", resp.ProductViewDetail.ProductARN)
	pvs := resp.ProductViewDetail.ProductViewSummary

	d.Set("description", pvs.ShortDescription)
	d.Set("distributor", pvs.Distributor)
	d.Set("id", d.Id())
	d.Set("name", pvs.Name)
	d.Set("owner", pvs.Owner)
	d.Set("product_type", pvs.Type)
	d.Set("support_description", pvs.SupportDescription)
	d.Set("support_email", pvs.SupportEmail)
	d.Set("support_url", pvs.SupportUrl)

	provisioningArtifactSummary := getProvisioningArtifactSummary(resp)
	d.Set("artifact_description", provisioningArtifactSummary.Description)
	d.Set("artifact_id", provisioningArtifactSummary.Id)
	d.Set("artifact_name", provisioningArtifactSummary.Name)
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
		input.Name = aws.String(v.(string))
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

	_, err := conn.DeleteProduct(&input)
	if err != nil {
		return fmt.Errorf("Deleting ServiceCatalog product '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
