package s3

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketInventory() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketInventoryPut,
		Read:   resourceBucketInventoryRead,
		Update: resourceBucketInventoryPut,
		Delete: resourceBucketInventoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"format": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											s3.InventoryFormatCsv,
											s3.InventoryFormatOrc,
											s3.InventoryFormatParquet,
										}, false),
									},
									"bucket_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"encryption": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"sse_kms": {
													Type:          schema.TypeList,
													Optional:      true,
													MaxItems:      1,
													ConflictsWith: []string{"destination.0.bucket.0.encryption.0.sse_s3"},
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key_id": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
														},
													},
												},
												"sse_s3": {
													Type:          schema.TypeList,
													Optional:      true,
													MaxItems:      1,
													ConflictsWith: []string{"destination.0.bucket.0.encryption.0.sse_kms"},
													Elem: &schema.Resource{
														// No options currently; just existence of "sse_s3".
														Schema: map[string]*schema.Schema{},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"frequency": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								s3.InventoryFrequencyDaily,
								s3.InventoryFrequencyWeekly,
							}, false),
						},
					},
				},
			},
			// TODO: Is there a sensible default for this?
			"included_object_versions": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					s3.InventoryIncludedObjectVersionsCurrent,
					s3.InventoryIncludedObjectVersionsAll,
				}, false),
			},
			"optional_fields": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(s3.InventoryOptionalField_Values(), false),
				},
				Set: schema.HashString,
			},
		},
	}
}

func resourceBucketInventoryPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn
	bucket := d.Get("bucket").(string)
	name := d.Get("name").(string)

	inventoryConfiguration := &s3.InventoryConfiguration{
		Id:        aws.String(name),
		IsEnabled: aws.Bool(d.Get("enabled").(bool)),
	}

	if v, ok := d.GetOk("included_object_versions"); ok {
		inventoryConfiguration.IncludedObjectVersions = aws.String(v.(string))
	}

	if v, ok := d.GetOk("optional_fields"); ok {
		inventoryConfiguration.OptionalFields = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("schedule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		scheduleList := v.([]interface{})
		scheduleMap := scheduleList[0].(map[string]interface{})
		inventoryConfiguration.Schedule = &s3.InventorySchedule{
			Frequency: aws.String(scheduleMap["frequency"].(string)),
		}
	}

	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		filterList := v.([]interface{})
		filterMap := filterList[0].(map[string]interface{})
		inventoryConfiguration.Filter = expandInventoryFilter(filterMap)
	}

	if v, ok := d.GetOk("destination"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		destinationList := v.([]interface{})
		destinationMap := destinationList[0].(map[string]interface{})
		bucketList := destinationMap["bucket"].([]interface{})
		bucketMap := bucketList[0].(map[string]interface{})

		inventoryConfiguration.Destination = &s3.InventoryDestination{
			S3BucketDestination: expandInventoryBucketDestination(bucketMap),
		}
	}

	input := &s3.PutBucketInventoryConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		InventoryConfiguration: inventoryConfiguration,
	}

	log.Printf("[DEBUG] Putting S3 bucket inventory configuration: %s", input)
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err := conn.PutBucketInventoryConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBucketInventoryConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error putting S3 Bucket Inventory Configuration: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", bucket, name))

	return resourceBucketInventoryRead(d, meta)
}

func resourceBucketInventoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, name, err := BucketInventoryParseID(d.Id())
	if err != nil {
		return err
	}

	input := &s3.DeleteBucketInventoryConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 bucket inventory configuration: %s", input)
	_, err = conn.DeleteBucketInventoryConfiguration(input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket Inventory Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceBucketInventoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, name, err := BucketInventoryParseID(d.Id())
	if err != nil {
		return err
	}

	d.Set("bucket", bucket)
	d.Set("name", name)

	input := &s3.GetBucketInventoryConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Reading S3 bucket inventory configuration: %s", input)
	var output *s3.GetBucketInventoryConfigurationOutput
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.GetBucketInventoryConfiguration(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetBucketInventoryConfiguration(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Inventory Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchConfiguration) {
		log.Printf("[WARN] S3 Bucket Inventory Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket Inventory Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.InventoryConfiguration == nil {
		return fmt.Errorf("error getting S3 Bucket Inventory Configuration (%s): empty response", d.Id())
	}

	d.Set("enabled", output.InventoryConfiguration.IsEnabled)
	d.Set("included_object_versions", output.InventoryConfiguration.IncludedObjectVersions)

	if err := d.Set("optional_fields", flex.FlattenStringList(output.InventoryConfiguration.OptionalFields)); err != nil {
		return fmt.Errorf("error setting optional_fields: %s", err)
	}

	if err := d.Set("filter", flattenInventoryFilter(output.InventoryConfiguration.Filter)); err != nil {
		return fmt.Errorf("error setting filter: %s", err)
	}

	if err := d.Set("schedule", flattenInventorySchedule(output.InventoryConfiguration.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %s", err)
	}

	if output.InventoryConfiguration.Destination != nil {
		destination := map[string]interface{}{
			"bucket": flattenInventoryBucketDestination(output.InventoryConfiguration.Destination.S3BucketDestination),
		}

		if err := d.Set("destination", []map[string]interface{}{destination}); err != nil {
			return fmt.Errorf("error setting destination: %s", err)
		}
	}

	return nil
}

func expandInventoryFilter(m map[string]interface{}) *s3.InventoryFilter {
	v, ok := m["prefix"]
	if !ok {
		return nil
	}
	return &s3.InventoryFilter{
		Prefix: aws.String(v.(string)),
	}
}

func flattenInventoryFilter(filter *s3.InventoryFilter) []map[string]interface{} {
	if filter == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, 1)

	m := make(map[string]interface{})
	if filter.Prefix != nil {
		m["prefix"] = aws.StringValue(filter.Prefix)
	}

	result = append(result, m)

	return result
}

func flattenInventorySchedule(schedule *s3.InventorySchedule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	m := make(map[string]interface{}, 1)
	m["frequency"] = aws.StringValue(schedule.Frequency)

	result = append(result, m)

	return result
}

func expandInventoryBucketDestination(m map[string]interface{}) *s3.InventoryS3BucketDestination {
	destination := &s3.InventoryS3BucketDestination{
		Format: aws.String(m["format"].(string)),
		Bucket: aws.String(m["bucket_arn"].(string)),
	}

	if v, ok := m["account_id"]; ok && v.(string) != "" {
		destination.AccountId = aws.String(v.(string))
	}

	if v, ok := m["prefix"]; ok && v.(string) != "" {
		destination.Prefix = aws.String(v.(string))
	}

	if v, ok := m["encryption"].([]interface{}); ok && len(v) > 0 {
		encryptionMap := v[0].(map[string]interface{})

		encryption := &s3.InventoryEncryption{}

		for k, v := range encryptionMap {
			data := v.([]interface{})

			if len(data) == 0 {
				continue
			}

			switch k {
			case "sse_kms":
				m := data[0].(map[string]interface{})
				encryption.SSEKMS = &s3.SSEKMS{
					KeyId: aws.String(m["key_id"].(string)),
				}
			case "sse_s3":
				encryption.SSES3 = &s3.SSES3{}
			}
		}

		destination.Encryption = encryption
	}

	return destination
}

func flattenInventoryBucketDestination(destination *s3.InventoryS3BucketDestination) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	m := map[string]interface{}{
		"format":     aws.StringValue(destination.Format),
		"bucket_arn": aws.StringValue(destination.Bucket),
	}

	if destination.AccountId != nil {
		m["account_id"] = aws.StringValue(destination.AccountId)
	}
	if destination.Prefix != nil {
		m["prefix"] = aws.StringValue(destination.Prefix)
	}

	if destination.Encryption != nil {
		encryption := make(map[string]interface{}, 1)
		if destination.Encryption.SSES3 != nil {
			encryption["sse_s3"] = []map[string]interface{}{{}}
		} else if destination.Encryption.SSEKMS != nil {
			encryption["sse_kms"] = []map[string]interface{}{
				{
					"key_id": aws.StringValue(destination.Encryption.SSEKMS.KeyId),
				},
			}
		}
		m["encryption"] = []map[string]interface{}{encryption}
	}

	result = append(result, m)

	return result
}

func BucketInventoryParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}
