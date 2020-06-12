package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func GetLaunchPathByName(ProductName string) string {

	productId := GetProductIdByName(ProductName)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := servicecatalog.New(sess)

	params :=&servicecatalog.ListLaunchPathsInput{
		ProductId: &productId,
	}

	resp, err := svc.ListLaunchPaths(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	}

	var LaunchPathId interface{}

	for i := range resp.LaunchPathSummaries {
		if *resp.LaunchPathSummaries[i].Name == ProductName {
			LaunchPathId = *resp.LaunchPathSummaries[i].Id
		}
	}
	return LaunchPathId.(string)
}

func GetProductIdByName(ProductName string) string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := servicecatalog.New(sess)

	params :=&servicecatalog.SearchProductsInput{
		AcceptLanguage: aws.String("en"),
	}

	resp, err := svc.SearchProducts(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	}

	var ProductId interface{}

	for i := range resp.ProductViewSummaries {
		if *resp.ProductViewSummaries[i].Name == ProductName {
			ProductId = *resp.ProductViewSummaries[i].ProductId
		}
	}
	return ProductId.(string)
}

func GetProductArtifactIdByName(ProductId string, ProductArtifactName string, meta interface{}) string {
	conn := meta.(*AWSClient).scconn

	params :=&servicecatalog.DescribeProductInput{
		AcceptLanguage: aws.String("en"),
		Id:             aws.String(ProductId),
	}

	resp, err := conn.DescribeProduct(params)

	if err != nil {
		return fmt.Errorf("retrieving product artificate id by name failed: %s", err.Error())
	}

	var ArtifactId interface{}

	for i := range resp.ProvisioningArtifacts {
		if *resp.ProvisioningArtifacts[i].Name == ProductArtifactName {
			ArtifactId = *resp.ProvisioningArtifacts[i].Id
		}
	}

	return ArtifactId.(string)
}

func resourceAwsServiceCatalogProvisionedProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProvisionedProductCreate,
		Read:   resourceAwsServiceCatalogProvisionedProductRead,
		Update: resourceAwsServiceCatalogProvisionedProductUpdate,
		Delete: resourceAwsServiceCatalogProvisionedProductDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tags": tagsSchema(),
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"notification_arn": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"path_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"product_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"product_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"provisioned_product_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"provisioning_artifact_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"provisioning_artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"provisioning_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
			"provisioning_preferences": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
func resourceAwsServiceCatalogProvisionedProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	//input := servicecatalog.CreatePortfolioInput{
	//	AcceptLanguage:   aws.String("en"),
	//	DisplayName:      aws.String(d.Get("name").(string)),
	//	IdempotencyToken: aws.String(resource.UniqueId()),
	//	Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ServicecatalogTags(),
	//}

	fmt.Println(d.Get("provisioning_parameters"))
	log.Println("matt")
	log.Println(d.Get("provisioning_parameters"))

	tempParams := d.Get("provisioning_parameters").([]interface{})

	var pHold []*servicecatalog.ProvisioningParameter

	for i, s := range tempParams {
		log.Println(i, s)
		for key, element := range s.(map[string]interface{}) {
			log.Println("Key:", key, "=>", "Element:", element)
			var tempelement = element.(string)
			log.Println(element.(string))
			temppp := servicecatalog.ProvisioningParameter{Key: &key, Value: &tempelement}
			pHold = append(pHold, &temppp)
		}
	}

	log.Println(pHold)

	input := servicecatalog.ProvisionProductInput{
		AcceptLanguage:         aws.String(d.Get("accept_language").(string)),
		ProvisionedProductName: aws.String(d.Get("provisioned_product_name").(string)),
		ProvisionToken:         aws.String(resource.UniqueId()),
		ProductId:              aws.String(d.Get("product_id").(string)),
		ProvisioningArtifactId: aws.String(d.Get("provisioning_artifact_id").(string)),
		PathId:                 aws.String(d.Get("path_id").(string)),
		ProvisioningParameters: pHold,
	}

	//if v, ok := d.GetOk("description"); ok {
	//	input.Description = aws.String(v.(string))
	//}
	//
	//if v, ok := d.GetOk("provider_name"); ok {
	//	input.ProviderName = aws.String(v.(string))
	//}

	log.Printf("[DEBUG] Provision Service Catalog Product: %#v", input)
	resp, err := conn.ProvisionProduct(&input)
	if err != nil {
		return fmt.Errorf("Provisioning Service Catalog product failed: %s", err.Error())
	}
	d.SetId(*resp.RecordDetail.RecordId)

	return resourceAwsServiceCatalogPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := servicecatalog.DescribePortfolioInput{
		AcceptLanguage: aws.String("en"),
	}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Reading Service Catalog Portfolio: %#v", input)
	resp, err := conn.DescribePortfolio(&input)
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Service Catalog Portfolio %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading ServiceCatalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	portfolioDetail := resp.PortfolioDetail
	if err := d.Set("created_time", portfolioDetail.CreatedTime.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}
	d.Set("arn", portfolioDetail.ARN)
	d.Set("description", portfolioDetail.Description)
	d.Set("name", portfolioDetail.DisplayName)
	d.Set("provider_name", portfolioDetail.ProviderName)

	if err := d.Set("tags", keyvaluetags.ServicecatalogKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsServiceCatalogProvisionedProductUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdatePortfolioInput{
		AcceptLanguage: aws.String("en"),
		Id:             aws.String(d.Id()),
	}

	if d.HasChange("name") {
		v, _ := d.GetOk("name")
		input.DisplayName = aws.String(v.(string))
	}

	if d.HasChange("accept_language") {
		v, _ := d.GetOk("accept_language")
		input.AcceptLanguage = aws.String(v.(string))
	}

	if d.HasChange("description") {
		v, _ := d.GetOk("description")
		input.Description = aws.String(v.(string))
	}

	if d.HasChange("provider_name") {
		v, _ := d.GetOk("provider_name")
		input.ProviderName = aws.String(v.(string))
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		input.AddTags = keyvaluetags.New(n).IgnoreAws().ServicecatalogTags()
		input.RemoveTags = aws.StringSlice(keyvaluetags.New(o).IgnoreAws().Keys())
	}

	log.Printf("[DEBUG] Update Service Catalog Portfolio: %#v", input)
	_, err := conn.UpdatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Updating Service Catalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return resourceAwsServiceCatalogPortfolioRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.DeletePortfolioInput{}
	input.Id = aws.String(d.Id())

	log.Printf("[DEBUG] Delete Service Catalog Portfolio: %#v", input)
	_, err := conn.DeletePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Portfolio '%s' failed: %s", *input.Id, err.Error())
	}
	return nil
}
