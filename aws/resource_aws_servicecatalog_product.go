package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
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
	input.IdempotencyToken = aws.String(resource.UniqueId())
	input.SetProvisioningArtifactParameters(&artifactProperties)
	log.Printf("[DEBUG] Creating Service Catalog Product: %s %s", input, artifactProperties)

	resp, err := conn.CreateProduct(&input)
	if err != nil {
		return fmt.Errorf("creating ServiceCatalog product failed: %s", err)
	}
	d.SetId(*resp.ProductViewDetail.ProductViewSummary.ProductId)
	if err := waitForServiceCatalogProductStatus(conn, d); err != nil {
		return err
	}
	return resourceAwsServiceCatalogProductRead(d, meta)
}

func waitForServiceCatalogProductStatus(conn *servicecatalog.ServiceCatalog, d *schema.ResourceData) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicecatalog.StatusCreating},
		// "CREATED" is not documented but seems to be the state it goes to
		Target: []string{servicecatalog.StatusAvailable, "CREATED"},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
				Id: aws.String(d.Id()),
			})
			if err != nil {
				return nil, "", err
			}
			return resp, aws.StringValue(resp.ProductViewDetail.Status), nil
		},
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func waitForServiceCatalogProductDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
	stateConf := resource.StateChangeConf{
		Pending:      []string{servicecatalog.StatusCreating},
		Target:       []string{""},
		Timeout:      15 * time.Minute,
		PollInterval: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeProductAsAdmin(&servicecatalog.DescribeProductAsAdminInput{
				Id: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
					return 42, "", nil
				}
				return 42, "", err
			}

			return resp, aws.StringValue(resp.ProductViewDetail.Status), nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
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

	err = d.Set("tags", tagsToMapServiceCatalog(keyvaluetags.ServicecatalogKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(meta.(*AWSClient).IgnoreTagsConfig).Map()))
	if err != nil {
		return fmt.Errorf("invalid tags read on ServiceCatalog product '%s': %s", d.Id(), err)
	}

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

	// TODO budgets
	// TODO tag options
	// TODO launch paths? (from describe product) -- probably not needed

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
		addTags := make(map[string]interface{})
		for k, v1 := range oldTags.(map[string]interface{}) {
			v2, ok := (newTags.(map[string]interface{}))[k]
			kk := k // copy, as &k is changing
			if !ok {
				removeTags = append(removeTags, &kk)
			} else if v2 != v1 {
				removeTags = append(removeTags, &kk)
				addTags[k] = v2
			}
		}
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

	if err := waitForServiceCatalogProductStatus(conn, d); err != nil {
		return err
	}

	// this change is slightly more complicated as basically we need to update the provisioning artifact
	if d.HasChange("provisioning_artifact") {
		newProvisioningArtifactList := d.Get("provisioning_artifact")
		newProvisioningArtifact := (newProvisioningArtifactList.([]interface{}))[0].(map[string]interface{})
		paId := newProvisioningArtifact["id"].(string)
		_, err := conn.UpdateProvisioningArtifact(&servicecatalog.UpdateProvisioningArtifactInput{
			ProductId:              aws.String(d.Id()),
			ProvisioningArtifactId: aws.String(paId),
			Name:                   aws.String(newProvisioningArtifact["name"].(string)),
			Description:            aws.String(newProvisioningArtifact["description"].(string)),
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
	if err := waitForServiceCatalogProductDeletion(conn, d.Id()); err != nil {
		return err
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

// tagsFromMap returns the tags for the given map of data.
func tagsFromMapServiceCatalog(m map[string]interface{}) []*servicecatalog.Tag {
	result := make([]*servicecatalog.Tag, 0, len(m))
	for k, v := range m {
		t := &servicecatalog.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		result = append(result, t)
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMapServiceCatalog(ts map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range ts {
		result[aws.StringValue(&k)] = aws.StringValue(&v)
	}
	return result
}
