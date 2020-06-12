package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform/helper/resource"
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
			Create: schema.DefaultTimeout(15 * time.Minute),
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
			"has_default_path": {
				Type:     schema.TypeBool,
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
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"info": {
							Type:     schema.TypeMap,
							Required: true,
							Elem:     schema.TypeString,
							ForceNew: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate,
						}, // CLOUD_FORMATION_TEMPLATE  | MARKETPLACE_AMI | MARKETPLACE_CAR
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"active": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"created_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsServiceCatalogProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.CreateProductInput{}
	now := time.Now()

	if v, ok := d.GetOk("name"); ok {
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

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = tagsFromMapServiceCatalog(v.(map[string]interface{}))
	}

	pa := d.Get("provisioning_artifact")
	paList := pa.([]interface{})
	paParameters := paList[0].(map[string]interface{})
	artifactProperties := servicecatalog.ProvisioningArtifactProperties{}
	artifactProperties.Description = aws.String(paParameters["description"].(string))
	artifactProperties.Name = aws.String(paParameters["name"].(string))
	if v, ok := paParameters["type"]; ok && v != "" {
		artifactProperties.Type = aws.String(v.(string))
	} else {
		artifactProperties.Type = aws.String(servicecatalog.ProvisioningArtifactTypeCloudFormationTemplate)
	}
	artifactProperties.Info = make(map[string]*string)
	for k, v := range paParameters["info"].(map[string]interface{}) {
		artifactProperties.Info[k] = aws.String(v.(string))
	}
	input.IdempotencyToken = aws.String(fmt.Sprintf("%d", now.UnixNano()))
	input.SetProvisioningArtifactParameters(&artifactProperties)
	log.Printf("[DEBUG] Creating Service Catalog Product: %s %s", input, artifactProperties)

	resp, err := conn.CreateProduct(&input)
	if err != nil {
		return fmt.Errorf("creating ServiceCatalog product failed: %s", err)
	}
	productId := aws.StringValue(resp.ProductViewDetail.ProductViewSummary.ProductId)

	waitForCreated := &resource.StateChangeConf{
		Target: []string{"CREATED", servicecatalog.StatusAvailable},
		Refresh: func() (result interface{}, state string, err error) {
			resp, err := conn.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
				Id: aws.String(productId),
			})
			if err != nil {
				return nil, "", err
			}
			return resp, aws.StringValue(resp.ProductViewDetail.Status), nil
		},
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	if _, err := waitForCreated.WaitForState(); err != nil {
		return err
	}

	d.SetId(productId)

	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	log.Printf("[DEBUG] Reading Service Catalog Product with id %s", d.Id())
	resp, err := conn.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
		Id: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Service Catalog Product %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading ServiceCatalog product '%s' failed: %s", d.Id(), err)
	}

	d.Set("product_arn", resp.ProductViewDetail.ProductARN)
	d.Set("tags", tagsToMapServiceCatalog(resp.Tags))

	product := resp.ProductViewDetail.ProductViewSummary
	d.Set("has_default_path", aws.BoolValue(product.HasDefaultPath))
	d.Set("description", aws.StringValue(product.ShortDescription))
	d.Set("distributor", aws.StringValue(product.Distributor))
	d.Set("name", aws.StringValue(product.Name))
	d.Set("owner", aws.StringValue(product.Owner))
	d.Set("product_type", aws.StringValue(product.Type))
	d.Set("support_description", aws.StringValue(product.SupportDescription))
	d.Set("support_email", aws.StringValue(product.SupportEmail))
	d.Set("support_url", aws.StringValue(product.SupportUrl))

	provisioningArtifactList := make([]map[string]interface{}, 0)
	for _, pas := range resp.ProvisioningArtifactSummaries {
		artifact := make(map[string]interface{})
		artifact["description"] = *pas.Description
		artifact["id"] = *pas.Id
		artifact["name"] = *pas.Name

		paOutput, err := conn.DescribeProvisioningArtifact(&servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(d.Id()),
			ProvisioningArtifactId: pas.Id,
		})
		if err != nil {
			return fmt.Errorf("reading ProvisioningArtifact '%s' for product '%s' failed: %s", *pas.Id, d.Id(), err)
		}
		artifact["type"] = aws.StringValue(paOutput.ProvisioningArtifactDetail.Type)
		artifact["active"] = aws.BoolValue(paOutput.ProvisioningArtifactDetail.Active)
		artifact["created_time"] = paOutput.ProvisioningArtifactDetail.CreatedTime.Format(time.RFC3339)
		replaceProvisioningArtifactParametersKeys(paOutput.Info)
		log.Printf("[DEBUG] Info map coming from READ: %v", paOutput.Info)
		info := make(map[string]string)
		for k, v := range paOutput.Info {
			info[k] = aws.StringValue(v)
		}
		artifact["info"] = info
		provisioningArtifactList = append(provisioningArtifactList, artifact)
	}

	if err := d.Set("provisioning_artifact", provisioningArtifactList); err != nil {
		return fmt.Errorf("setting ProvisioningArtifact for product '%s' failed: %s", d.Id(), err)
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

	// figure out what tags to add and what tags to remove
	if d.HasChange("tags") {
		oldTags, newTags := d.GetChange("tags")
		removeTags := make([]*string, 0)
		for k := range oldTags.(map[string]interface{}) {
			if _, ok := (newTags.(map[string]interface{}))[k]; !ok {
				removeTags = append(removeTags, &k)
			}
		}
		addTags := make(map[string]interface{})
		for k, v := range newTags.(map[string]interface{}) {
			if _, ok := (oldTags.(map[string]interface{}))[k]; !ok {
				addTags[k] = v
			}
		}
		input.AddTags = tagsFromMapServiceCatalog(addTags)
		input.RemoveTags = removeTags
	}

	log.Printf("[DEBUG] Update Service Catalog Product: %s", input)
	_, err := conn.UpdateProduct(&input)
	if err != nil {
		return fmt.Errorf("updating ServiceCatalog product '%s' failed: %s", *input.Id, err)
	}

	// this change is slightly more complicated as basically we need to update the provisioning artifact
	if d.HasChange("provisioning_artifact") {
		_, newProvisioningArtifactList := d.GetChange("provisioning_artifact")
		newProvisioningArtifact := (newProvisioningArtifactList.([]interface{}))[0].(map[string]interface{})
		paId := newProvisioningArtifact["id"].(string)
		_, err := conn.UpdateProvisioningArtifact(&servicecatalog.UpdateProvisioningArtifactInput{
			ProductId:              aws.String(d.Id()),
			ProvisioningArtifactId: aws.String(paId),
			Name:        aws.String(newProvisioningArtifact["name"].(string)),
			Description: aws.String(newProvisioningArtifact["description"].(string)),
		})
		if err != nil {
			return fmt.Errorf("unable to update provisioning artifact %s for product %s due to %s", d.Id(), paId, err)
		}
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
		return fmt.Errorf("deleting ServiceCatalog product '%s' failed: %s", *input.Id, err)
	}
	return nil
}

// this is to workaround the issue of the `info` map which contains different keys between user-provided and READ operation
func replaceProvisioningArtifactParametersKeys(m map[string]*string) {
	replaceProvisioningArtifactParametersKey(m, "TemplateUrl", "LoadTemplateFromURL")
}

func replaceProvisioningArtifactParametersKey(m map[string]*string, replacedKey, withKey string) {
	m[withKey] = m[replacedKey]
	delete(m, replacedKey)
}
