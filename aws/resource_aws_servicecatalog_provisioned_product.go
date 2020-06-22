package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
			"notification_arns": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"path_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"product_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"provisioned_product_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"provisioning_artifact_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"provisioning_parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			/*
				// TODO stack set preferences
				"provisioning_preferences": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     <various>
				},
			*/
			"tags": tagsSchemaForceNew(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_record_errors": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: schema.TypeString,
				},
			},
			"last_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_record_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_record_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
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

	if v, ok := d.GetOk("notification_arns"); ok {
		input.NotificationArns = []*string{}
		for _, na := range v.([]string) {
			input.NotificationArns = append(input.NotificationArns, aws.String(na))
		}
	}
	if v, ok := d.GetOk("path_id"); ok {
		input.PathId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("provisioning_parameters"); ok {
		input.ProvisioningParameters = make([]*servicecatalog.ProvisioningParameter, 0)
		for k, vv := range v.(map[string]interface{}) {
			input.ProvisioningParameters = append(input.ProvisioningParameters,
				&servicecatalog.ProvisioningParameter{Key: aws.String(k), Value: aws.String(vv.(string))})
		}
	}
	/*
	   // TODO stack set prefs - and update docs
	   if v, ok := d.GetOk("provisioning_preferences"); ok {
	       input.Description = aws.String(v.(string))
	   }
	*/

	log.Printf("[DEBUG] Provision Service Catalog Product: %#v", input)
	resp, err := conn.ProvisionProduct(&input)
	if err != nil {
		return fmt.Errorf("Provisioning Service Catalog product failed: %s", err.Error())
	}

	d.SetId(*resp.RecordDetail.ProvisionedProductId)
	if err := waitForServiceCatalogProvisionedProductStatus(conn, d); err != nil {
		return err
	}
	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func waitForServiceCatalogProvisionedProductStatus(conn *servicecatalog.ServiceCatalog, d *schema.ResourceData) error {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{servicecatalog.ProvisionedProductStatusUnderChange},
		Target:       []string{servicecatalog.ProvisionedProductStatusAvailable},
		Refresh:      refreshProvisionedProductStatus(conn, d),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		PollInterval: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func refreshProvisionedProductStatus(conn *servicecatalog.ServiceCatalog, d *schema.ResourceData) resource.StateRefreshFunc {
	return func() (result interface{}, state string, err error) {
		resp, err := conn.DescribeProvisionedProduct(&servicecatalog.DescribeProvisionedProductInput{
			Id: aws.String(d.Id()),
		})
		if err != nil {
			// to help debug if there's a problem
			d.Set("status", resp.ProvisionedProductDetail.Status)
			d.Set("status_message", resp.ProvisionedProductDetail.StatusMessage)
			return nil, "", err
		}
		return resp, aws.StringValue(resp.ProvisionedProductDetail.Status), nil
	}
}

func waitForServiceCatalogProvisionedProductDeletion(conn *servicecatalog.ServiceCatalog, id string) error {
	stateConf := resource.StateChangeConf{
		Pending:      []string{servicecatalog.ProvisionedProductStatusUnderChange},
		Target:       []string{""},
		Timeout:      15 * time.Minute,
		PollInterval: 3 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeProvisionedProduct(&servicecatalog.DescribeProvisionedProductInput{
				Id: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
					return 42, "", nil
				}
				return 42, "", err
			}

			return resp, aws.StringValue(resp.ProvisionedProductDetail.Status), nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceAwsServiceCatalogProvisionedProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := servicecatalog.DescribeProvisionedProductInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Service Catalog Provisioned Product: %#v", input)
	resp, err := conn.DescribeProvisionedProduct(&input)

	if err != nil {
		if isAWSErr(err, servicecatalog.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Service Catalog Provisioned Product %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading ServiceCatalog Provisioned Product '%s' failed: %s", *input.Id, err.Error())
	}

	detail := resp.ProvisionedProductDetail

	record, err := conn.DescribeRecord(&servicecatalog.DescribeRecordInput{Id: aws.String(*detail.LastRecordId)})
	if err != nil {
		return fmt.Errorf("Reading ServiceCatalog Provisioned Product '%s' failed on record '%s': %s", *input.Id, *detail.LastRecordId, err.Error())
	}

	// from ProvisionedProductDetail
	d.Set("id", detail.Id)
	d.Set("arn", detail.Arn)
	d.Set("created_time", detail.CreatedTime.Format(time.RFC3339))
	d.Set("provisioned_product_name", detail.Name)
	d.Set("product_id", detail.ProductId)
	d.Set("provisioning_artifact_id", detail.ProvisioningArtifactId)
	d.Set("status", detail.Status)
	d.Set("status_message", detail.StatusMessage)
	d.Set("last_record_id", detail.LastRecordId)
	// detail.Type omitted

	// from DescribeRecord
	d.Set("path_id", record.RecordDetail.PathId)
	d.Set("updated_time", record.RecordDetail.UpdatedTime.Format(time.RFC3339))
	d.Set("last_record_type", record.RecordDetail.RecordType)
	d.Set("last_record_status", record.RecordDetail.Status)

	recordErrors := make([]map[string]string, 0)
	for _, b := range record.RecordDetail.RecordErrors {
		bb := make(map[string]string)
		bb["code"] = aws.StringValue(b.Code)
		bb["description"] = aws.StringValue(b.Description)
		recordErrors = append(recordErrors, bb)
	}
	err = d.Set("last_record_errors", recordErrors)
	if err != nil {
		return fmt.Errorf("invalid errors read on ServiceCatalog provisioned product '%s': %s", d.Id(), err)
	}

	outputs := make(map[string]string)
	for _, b := range record.RecordOutputs {
		outputs[aws.StringValue(b.OutputKey)] = aws.StringValue(b.OutputValue)
	}
	err = d.Set("outputs", outputs)
	if err != nil {
		return fmt.Errorf("invalid outputs read on ServiceCatalog provisioned product '%s': %s", d.Id(), err)
	}

	//not returned (they should be what we set):
	// notification_arns
	// provisioning_parameters
	// provisioning_preferences
	// tags

	// ignored: record.CloudWatchDashboards

	return nil
}

func resourceAwsServiceCatalogProvisionedProductUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.UpdateProvisionedProductInput{
		ProvisionedProductId: aws.String(d.Id()),
		UpdateToken:          aws.String(resource.UniqueId()),
	}

	if d.HasChange("notification_arns") {
		return fmt.Errorf("Update/changes to notification_arns not supported (should force new)")
	}

	if d.HasChange("path_id") {
		v, _ := d.GetOk("path_id")
		input.PathId = aws.String(v.(string))
	}

	if d.HasChange("product_id") {
		v, _ := d.GetOk("product_id")
		input.ProductId = aws.String(v.(string))
	}

	if d.HasChange("provisioned_product_name") {
		v, _ := d.GetOk("provisioned_product_name")
		input.ProvisionedProductName = aws.String(v.(string))
	}

	if d.HasChange("provisioning_artifact_id") {
		v, _ := d.GetOk("provisioning_artifact_id")
		input.ProvisioningArtifactId = aws.String(v.(string))
	}

	if d.HasChange("provisioning_parameters") {
		v, _ := d.GetOk("provisioning_parameters")
		input.ProvisioningParameters = make([]*servicecatalog.UpdateProvisioningParameter, 0)
		for k, vv := range v.(map[string]interface{}) {
			input.ProvisioningParameters = append(input.ProvisioningParameters,
				&servicecatalog.UpdateProvisioningParameter{Key: aws.String(k), Value: aws.String(vv.(string))})
		}
	}

	// TODO stack set preferences

	if d.HasChange("tags") {
		n, _ := d.GetOk("tags")
		input.Tags = keyvaluetags.New(n.(map[string]interface{})).IgnoreAws().ServicecatalogTags()
	}

	log.Printf("[DEBUG] Update Service Catalog Provisioned Product: %#v", input)
	_, err := conn.UpdateProvisionedProduct(&input)
	if err != nil {
		return fmt.Errorf("Updating Service Catalog Provisioned Product '%s' failed: %s", *input.ProvisionedProductId, err.Error())
	}
	if err := waitForServiceCatalogProvisionedProductStatus(conn, d); err != nil {
		return err
	}
	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	input := servicecatalog.TerminateProvisionedProductInput{
		ProvisionedProductId: aws.String(d.Id()),
		TerminateToken:       aws.String(resource.UniqueId()),
	}

	// not available on servicecatalog, but returned here if under change
	errCodeValidationException := "ValidationException"

	log.Printf("[DEBUG] Delete Service Catalog Provisioned Product: %#v", input)
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.TerminateProvisionedProduct(&input)
		if err != nil {
			if isAWSErr(err, servicecatalog.ErrCodeResourceInUseException, "") || isAWSErr(err, errCodeValidationException, "") {
				// delay and retry, other things eg associations might still be getting deleted
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Deleting Service Catalog Provisioned Product '%s' failed: %s", *input.ProvisionedProductId, err.Error())
	}
	if err := waitForServiceCatalogProvisionedProductDeletion(conn, d.Id()); err != nil {
		return err
	}
	return nil
}
