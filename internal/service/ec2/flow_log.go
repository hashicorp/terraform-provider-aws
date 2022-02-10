package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFlowLog() *schema.Resource {
	return &schema.Resource{
		Create: resourceLogFlowCreate,
		Read:   resourceLogFlowRead,
		Update: resourceLogFlowUpdate,
		Delete: resourceLogFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_format": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(ec2.DestinationFileFormat_Values(), false),
							Optional:     true,
							Default:      ec2.DestinationFileFormatPlainText,
							ForceNew:     true,
						},
						"hive_compatible_partitions": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
						"per_hour_partition": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},
			"eni_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"log_destination": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"log_group_name"},
			},
			"log_destination_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.LogDestinationTypeCloudWatchLogs,
				ValidateFunc: validation.StringInSlice(ec2.LogDestinationType_Values(), false),
			},
			"log_format": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"log_group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"log_destination"},
				Deprecated:    "use 'log_destination' argument instead",
			},
			"max_aggregation_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      600,
				ValidateFunc: validation.IntInSlice([]int{60, 600}),
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"traffic_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrafficType_Values(), false),
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLogFlowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var resourceID string
	var resourceType string
	for _, v := range []struct {
		ID   string
		Type string
	}{
		{
			ID:   d.Get("vpc_id").(string),
			Type: ec2.FlowLogsResourceTypeVpc,
		},
		{
			ID:   d.Get("subnet_id").(string),
			Type: ec2.FlowLogsResourceTypeSubnet,
		},
		{
			ID:   d.Get("eni_id").(string),
			Type: ec2.FlowLogsResourceTypeNetworkInterface,
		},
	} {
		if v.ID != "" {
			resourceID = v.ID
			resourceType = v.Type
			break
		}
	}

	input := &ec2.CreateFlowLogsInput{
		LogDestinationType: aws.String(d.Get("log_destination_type").(string)),
		ResourceIds:        aws.StringSlice([]string{resourceID}),
		ResourceType:       aws.String(resourceType),
		TrafficType:        aws.String(d.Get("traffic_type").(string)),
	}

	if v, ok := d.GetOk("destination_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationOptions = expandEc2DestinationOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.DeliverLogsPermissionArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_destination"); ok {
		input.LogDestination = aws.String(strings.TrimSuffix(v.(string), ":*"))
	}

	if v, ok := d.GetOk("log_format"); ok {
		input.LogFormat = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_group_name"); ok {
		input.LogGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_aggregation_interval"); ok {
		input.MaxAggregationInterval = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		input.TagSpecifications = ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpcFlowLog)
	}

	log.Printf("[DEBUG] Creating Flow Log: %s", input)
	output, err := conn.CreateFlowLogs(input)

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if err != nil {
		return fmt.Errorf("error creating Flow Log (%s): %w", resourceID, err)
	}

	d.SetId(aws.StringValue(output.FlowLogIds[0]))

	return resourceLogFlowRead(d, meta)
}

func resourceLogFlowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	fl, err := FindFlowLogByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Flow Log %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Flow Log (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-flow-log/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if fl.DestinationOptions != nil {
		if err := d.Set("destination_options", []interface{}{flattenEc2DestinationOptionsResponse(fl.DestinationOptions)}); err != nil {
			return fmt.Errorf("error setting destination_options: %w", err)
		}
	} else {
		d.Set("destination_options", nil)
	}
	d.Set("iam_role_arn", fl.DeliverLogsPermissionArn)
	d.Set("log_destination", fl.LogDestination)
	d.Set("log_destination_type", fl.LogDestinationType)
	d.Set("log_format", fl.LogFormat)
	d.Set("log_group_name", fl.LogGroupName)
	d.Set("max_aggregation_interval", fl.MaxAggregationInterval)
	d.Set("traffic_type", fl.TrafficType)

	switch resourceID := aws.StringValue(fl.ResourceId); {
	case strings.HasPrefix(resourceID, "vpc-"):
		d.Set("vpc_id", resourceID)
	case strings.HasPrefix(resourceID, "subnet-"):
		d.Set("subnet_id", resourceID)
	case strings.HasPrefix(resourceID, "eni-"):
		d.Set("eni_id", resourceID)
	}

	tags := KeyValueTags(fl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLogFlowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Flow Log (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLogFlowRead(d, meta)
}

func resourceLogFlowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting Flow Log: %s", d.Id())
	output, err := conn.DeleteFlowLogs(&ec2.DeleteFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidFlowLogIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Flow Log (%s): %w", d.Id(), err)
	}

	return nil
}

func expandEc2DestinationOptionsRequest(tfMap map[string]interface{}) *ec2.DestinationOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.DestinationOptionsRequest{}

	if v, ok := tfMap["file_format"].(string); ok && v != "" {
		apiObject.FileFormat = aws.String(v)
	}

	if v, ok := tfMap["hive_compatible_partitions"].(bool); ok {
		apiObject.HiveCompatiblePartitions = aws.Bool(v)
	}

	if v, ok := tfMap["per_hour_partition"].(bool); ok {
		apiObject.PerHourPartition = aws.Bool(v)
	}

	return apiObject
}

func flattenEc2DestinationOptionsResponse(apiObject *ec2.DestinationOptionsResponse) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.FileFormat; v != nil {
		tfMap["file_format"] = aws.StringValue(v)
	}

	if v := apiObject.HiveCompatiblePartitions; v != nil {
		tfMap["hive_compatible_partitions"] = aws.BoolValue(v)
	}

	if v := apiObject.PerHourPartition; v != nil {
		tfMap["per_hour_partition"] = aws.BoolValue(v)
	}

	return tfMap
}
