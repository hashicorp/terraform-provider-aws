package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/schema"
)

func generateCustomerManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateServiceManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
}

func generateStorageSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"customer_managed_s3": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateCustomerManagedS3Schema(),
			},
		},
	}
}

func generateRetentionPeriodSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number_of_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"unlimited": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsIotAnalyticsChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsChannelCreate,
		Read:   resourceAwsIotAnalyticsChannelRead,
		Update: resourceAwsIotAnalyticsChannelUpdate,
		Delete: resourceAwsIotAnalyticsChannelDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"storage": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateStorageSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
		},
	}
}

func parseCustomerManagedS3(rawCustomerManagedS3 map[string]interface{}) *iotanalytics.CustomerManagedChannelS3Storage {
	bucket := rawCustomerManagedS3["bucket"].(string)
	roleArn := rawCustomerManagedS3["role_arn"].(string)
	customerManagedS3 := &iotanalytics.CustomerManagedChannelS3Storage{
		Bucket:  aws.String(bucket),
		RoleArn: aws.String(roleArn),
	}

	if v, ok := rawCustomerManagedS3["key_prefix"]; ok && len(v.(string)) >= 1 {
		customerManagedS3.KeyPrefix = aws.String(v.(string))
	}

	return customerManagedS3
}

func parseServiceManagedS3(rawServiceManagedS3 map[string]interface{}) *iotanalytics.ServiceManagedChannelS3Storage {
	return &iotanalytics.ServiceManagedChannelS3Storage{}
}

func parseStorage(rawChannelStorage map[string]interface{}) *iotanalytics.ChannelStorage {

	var customerManagedS3 *iotanalytics.CustomerManagedChannelS3Storage
	if set := rawChannelStorage["customer_managed_s3"].(*schema.Set).List(); len(set) > 0 {
		rawCustomerManagedS3 := set[0].(map[string]interface{})
		customerManagedS3 = parseCustomerManagedS3(rawCustomerManagedS3)
	}
	return &iotanalytics.ChannelStorage{
		CustomerManagedS3: customerManagedS3,
	}
}

func parseRetentionPeriod(rawRetentionPeriod map[string]interface{}) *iotanalytics.RetentionPeriod {

	var numberOfDays *int64
	if v, ok := rawRetentionPeriod["numberOfDays"]; ok && v.(int64) > 1 {
		numberOfDays = aws.Int64(v.(int64))
	}
	var unlimited *bool
	if v, ok := rawRetentionPeriod["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.RetentionPeriod{
		NumberOfDays: numberOfDays,
		Unlimited:    unlimited,
	}
}

func resourceAwsIotAnalyticsChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.CreateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}

	channelStorageSet := d.Get("storage").(*schema.Set).List()
	if len(channelStorageSet) >= 1 {
		rawChannelStorage := channelStorageSet[0].(map[string]interface{})
		params.ChannelStorage = parseStorage(rawChannelStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Create IoTAnalytics Channel: %s", params)
	_, err := conn.CreateChannel(params)

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsIotAnalyticsChannelRead(d, meta)
}

func flattenCustomerManagedS3(customerManagedS3 *iotanalytics.CustomerManagedChannelS3Storage) map[string]interface{} {
	if customerManagedS3 == nil {
		return nil
	}

	rawCustomerManagedS3 := make(map[string]interface{})

	rawCustomerManagedS3["bucket"] = aws.StringValue(customerManagedS3.Bucket)
	rawCustomerManagedS3["role_arn"] = aws.StringValue(customerManagedS3.RoleArn)

	if customerManagedS3.KeyPrefix != nil {
		rawCustomerManagedS3["key_prefix"] = aws.StringValue(customerManagedS3.KeyPrefix)
	}

	return rawCustomerManagedS3
}

func flattenServiceManagedS3(serviceManagedS3 *iotanalytics.ServiceManagedChannelS3Storage) map[string]interface{} {
	if serviceManagedS3 == nil {
		return nil
	}

	rawServiceManagedS3 := make(map[string]interface{})
	return rawServiceManagedS3
}

func flattenStorage(channelStorage *iotanalytics.ChannelStorage) map[string]interface{} {
	customerManagedS3 := flattenCustomerManagedS3(channelStorage.CustomerManagedS3)
	serviceManagedS3 := flattenServiceManagedS3(channelStorage.ServiceManagedS3)

	if customerManagedS3 == nil && serviceManagedS3 == nil {
		return nil
	}

	rawStorage := make(map[string]interface{})
	rawStorage["customer_managed_s3"] = wrapMapInList(customerManagedS3)
	return rawStorage
}

func flattenRetentionPeriod(retentionPeriod *iotanalytics.RetentionPeriod) map[string]interface{} {
	rawRetentionPeriod := make(map[string]interface{})

	if retentionPeriod.NumberOfDays != nil {
		rawRetentionPeriod["number_of_days"] = aws.Int64Value(retentionPeriod.NumberOfDays)
	}
	if retentionPeriod.Unlimited != nil {
		rawRetentionPeriod["unlimited"] = aws.BoolValue(retentionPeriod.Unlimited)
	}

	return rawRetentionPeriod
}

func wrapMapInList(mapping map[string]interface{}) []interface{} {
	if mapping == nil {
		return make([]interface{}, 0)
	} else {
		return []interface{}{mapping}
	}
}

func resourceAwsIotAnalyticsChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeChannelInput{
		ChannelName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading IoTAnalytics Channel: %s", params)

	out, err := conn.DescribeChannel(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Channel.Name)
	storage := flattenStorage(out.Channel.Storage)
	d.Set("storage", wrapMapInList(storage))
	retentionPeriod := flattenRetentionPeriod(out.Channel.RetentionPeriod)
	d.Set("retention_period", wrapMapInList(retentionPeriod))

	return nil
}

func resourceAwsIotAnalyticsChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.UpdateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}

	channelStorageSet := d.Get("storage").(*schema.Set).List()
	if len(channelStorageSet) >= 1 {
		rawChannelStorage := channelStorageSet[0].(map[string]interface{})
		params.ChannelStorage = parseStorage(rawChannelStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Updating IoTAnalytics Channel: %s", params)
	_, err := conn.UpdateChannel(params)

	if err != nil {
		return err
	}

	return resourceAwsIotAnalyticsChannelRead(d, meta)
}

func resourceAwsIotAnalyticsChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteChannelInput{
		ChannelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Delete IoTAnalytics Channel: %s", params)
	_, err := conn.DeleteChannel(params)

	return err
}
