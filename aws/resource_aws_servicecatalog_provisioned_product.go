package aws

import (
	"fmt"
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
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"provisioned_product_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"provisioning_artifact_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
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
			"tags": tagsSchema(),
			
		},
	}
}
func resourceAwsServiceCatalogProvisionedProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.ProvisionProductInput{
		ProductId:              aws.String(d.Get("product_id").(string)),
		ProvisioningArtifactId: aws.String(d.Get("provisioning_artifact_id").(string)),
		ProvisionedProductName: aws.String(d.Get("provisioned_product_name").(string)),
		ProvisionToken:         aws.String(resource.UniqueId()),
		Tags:                   keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ServicecatalogTags(),
	}
    
    // TODO create with additional fields
//		PathId:                 aws.String(d.Get("path_id").(string)),
//		notificationArns
//		ProvisioningParameters: pHold,
/*
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
    if v := d.Get("instance_initiated_shutdown_behavior").(string); v != "" && ok {
        opts.InstanceInitiatedShutdownBehavior = aws.String(v)
    }
*/

	log.Printf("[DEBUG] Provision Service Catalog Product: %#v", input)
	resp, err := conn.ProvisionProduct(&input)
	if err != nil {
		return fmt.Errorf("Provisioning Service Catalog product failed: %s", err.Error())
	}
	
	d.SetId(*resp.RecordDetail.ProvisionedProductId)
	if err := waitForServiceCatalogProvisionedProductStatus("CREATED", conn, d); err != nil {
		return err
	}
	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func waitForServiceCatalogProvisionedProductStatus(status string, conn *servicecatalog.ServiceCatalog, d *schema.ResourceData) error {
	stateConf := &resource.StateChangeConf{
		Target: []string{status, servicecatalog.StatusAvailable},
		Refresh: refreshProvisionedProductStatus(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func refreshProvisionedProductStatus(conn *servicecatalog.ServiceCatalog, id string) resource.StateRefreshFunc {
	return func() (result interface{}, state string, err error) {
		resp, err := conn.DescribeProvisionedProduct(&servicecatalog.DescribeProvisionedProductInput{
			Id: aws.String(id),
		})
		if err != nil {
			return nil, "", err
		}
		return resp, aws.StringValue(resp.ProvisionedProductDetail.Status), nil
	}
}

func resourceAwsServiceCatalogProvisionedProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.DescribeProvisionedProductInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Service Catalog Provisioned Product: %#v", input)
	resp, err := conn.DescribeProvisionedProduct(&input)
	// TODO or SearchProvisionedProducts ???
	
	if err != nil {
		if scErr, ok := err.(awserr.Error); ok && scErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] Service Catalog Provisioned Product %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading ServiceCatalog Provisioned Product '%s' failed: %s", *input.Id, err.Error())
	}
	
	detail := resp.ProvisionedProductDetail
	if err := d.Set("created_time", detail.CreatedTime.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}
	
	// TODO read other fields
	
	//d.Set("arn", detail.ARN)
/*
    //  "ProvisionedProductDetail":	
      "Arn": "string",
      "CreatedTime": number,
      "Id": "string",
      "IdempotencyToken": "string",
      "LastRecordId": "string",
      "Name": "string",
      "ProductId": "string",
      "ProvisioningArtifactId": "string",
      "Status": "string",
      "StatusMessage": "string",
      "Type": "string"
      
      
      or ... first in ProvisionedProducts
      
         "Arn": "string",
         "CreatedTime": number,
         "Id": "string",
         "IdempotencyToken": "string",
         "LastRecordId": "string",
         "Name": "string",
         "ProductId": "string",
         "ProvisioningArtifactId": "string",
         "Status": "string",
         "StatusMessage": "string",
         "Type": "string",
         
         // extras:
         "PhysicalId": "string",
         "Tags": [ 
            { 
               "Key": "string",
               "Value": "string"
            }
         ],
         "UserArn": "string",
         "UserArnSession": "string"
*/

//	if err := d.Set("tags", keyvaluetags.ServicecatalogKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
//		return fmt.Errorf("error setting tags: %s", err)
//	}
	
	// ignored: cloudwatch dashboards

	return nil
}

func resourceAwsServiceCatalogProvisionedProductUpdate(d *schema.ResourceData, meta interface{}) error {
    // TODO update not done (code is for portfolio)
    
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdatePortfolioInput{
		Id:             aws.String(d.Id()),
	}

	if d.HasChange("name") {
		v, _ := d.GetOk("name")
		input.DisplayName = aws.String(v.(string))
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

	log.Printf("[DEBUG] Update Service Catalog Provisioned Product: %#v", input)
	_, err := conn.UpdatePortfolio(&input)
	if err != nil {
		return fmt.Errorf("Updating Service Catalog Provisioned Product '%s' failed: %s", *input.Id, err.Error())
	}
	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.TerminateProvisionedProductInput{
	    ProvisionedProductId : aws.String(d.Id()),
	    TerminateToken : aws.String(resource.UniqueId()),
	}

	log.Printf("[DEBUG] Delete Service Catalog Provisioned Product: %#v", input)
	_, err := conn.TerminateProvisionedProduct(&input)
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Provisioned Product '%s' failed: %s", *input.ProvisionedProductId, err.Error())
	}
	return nil
}
