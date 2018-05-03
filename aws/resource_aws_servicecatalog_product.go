package aws

import (
	"fmt"
	"log"
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
				ForceNew: true,
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
				Type:     schema.TypeList,
				MinItems: 1,
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
	now := time.Now()

	if v, ok := d.GetOk("name"); ok {
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

	if pa, ok := d.GetOk("provisioning_artifact"); ok {
		pvs := pa.([]interface{})
		v := pvs[0]
		bd := v.(map[string]interface{})
		artifactProperties := servicecatalog.ProvisioningArtifactProperties{}
		artifactProperties.Description = aws.String(bd["description"].(string))
		artifactProperties.Name = aws.String(bd["name"].(string))
		artifactProperties.Type = aws.String("CLOUD_FORMATION_TEMPLATE")
		url := aws.String(bd["load_template_from_url"].(string))
		info := map[string]*string{
			"LoadTemplateFromURL": url,
		}
		artifactProperties.Info = info
		input.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))
		input.SetProvisioningArtifactParameters(&artifactProperties)
		log.Printf("[DEBUG] Creating Service Catalog Product: %s %s", input, artifactProperties)

	}

	resp, err := conn.CreateProduct(&input)
	if err != nil {
		return fmt.Errorf("Creating ServiceCatalog product failed: %s", err.Error())
	}
	d.SetId(*resp.ProductViewDetail.ProductViewSummary.ProductId)

	if pa, ok := d.GetOk("provisioning_artifact"); ok {
		pvs := pa.([]interface{})
		if len(pvs) > 1 {
			for _, pa := range pvs[1:len(pvs)] {
				bd := pa.(map[string]interface{})
				parameters := servicecatalog.ProvisioningArtifactProperties{}
				parameters.Description = aws.String(bd["description"].(string))
				parameters.Name = aws.String(bd["name"].(string))
				parameters.Type = aws.String("CLOUD_FORMATION_TEMPLATE")

				url := aws.String(bd["load_template_from_url"].(string))
				info := map[string]*string{
					"LoadTemplateFromURL": url,
				}
				parameters.Info = info

				cpai := servicecatalog.CreateProvisioningArtifactInput{}
				cpai.Parameters = &parameters
				cpai.ProductId = aws.String(d.Id())
				cpai.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))
				log.Printf("[DEBUG] Adding Service Catalog Provisioning Artifact : %s", cpai)
				resp, err := conn.CreateProvisioningArtifact(&cpai)
				if err != nil {
					return fmt.Errorf("Creating ServiceCatalog provisioning artifact failed: %s %s", err.Error(), resp)
				}
			}
		}
	}

	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Service Catalog Product: %s", input)
	resp, err := conn.DescribeProductAsAdmin(&input)
	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
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

	var a []map[string]interface{}
	for _, pas := range resp.ProvisioningArtifactSummaries {
		artifact := make(map[string]interface{})
		artifact["description"] = *pas.Description
		artifact["id"] = *pas.Id
		artifact["name"] = *pas.Name

		i := servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(d.Id()),
			ProvisioningArtifactId: aws.String(*pas.Id),
		}
		dpao, err := conn.DescribeProvisioningArtifact(&i)
		if err != nil {
			return err
		}
		artifact["load_template_from_url"] = *dpao.Info["TemplateUrl"]
		a = append(a, artifact)
	}

	if err := d.Set("provisioning_artifact", a); err != nil {
		return err
	}
	return nil
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
