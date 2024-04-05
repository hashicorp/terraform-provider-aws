// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_flow_log", name="Flow Log")
// @Tags(identifierAttribute="id")
func ResourceFlowLog() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLogFlowCreate,
		ReadWithoutTimeout:   resourceLogFlowRead,
		UpdateWithoutTimeout: resourceLogFlowUpdate,
		DeleteWithoutTimeout: resourceLogFlowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deliver_cross_account_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id", "transit_gateway_id", "transit_gateway_attachment_id"},
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
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id", "transit_gateway_id", "transit_gateway_attachment_id"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"traffic_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.TrafficType_Values(), false),
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id", "transit_gateway_id", "transit_gateway_attachment_id"},
			},
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id", "transit_gateway_id", "transit_gateway_attachment_id"},
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", "subnet_id", "vpc_id", "transit_gateway_id", "transit_gateway_attachment_id"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLogFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

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
			ID:   d.Get("transit_gateway_id").(string),
			Type: ec2.FlowLogsResourceTypeTransitGateway,
		},
		{
			ID:   d.Get("transit_gateway_attachment_id").(string),
			Type: ec2.FlowLogsResourceTypeTransitGatewayAttachment,
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
		ClientToken:        aws.String(id.UniqueId()),
		LogDestinationType: aws.String(d.Get("log_destination_type").(string)),
		ResourceIds:        aws.StringSlice([]string{resourceID}),
		ResourceType:       aws.String(resourceType),
		TagSpecifications:  getTagSpecificationsIn(ctx, ec2.ResourceTypeVpcFlowLog),
	}

	if resourceType != ec2.FlowLogsResourceTypeTransitGateway && resourceType != ec2.FlowLogsResourceTypeTransitGatewayAttachment {
		if v, ok := d.GetOk("traffic_type"); ok {
			input.TrafficType = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("destination_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationOptions = expandDestinationOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("deliver_cross_account_role"); ok {
		input.DeliverCrossAccountRole = aws.String(v.(string))
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

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, iamPropagationTimeout, func() (interface{}, error) {
		return conn.CreateFlowLogsWithContext(ctx, input)
	}, errCodeInvalidParameter, "Unable to assume given IAM role")

	if err == nil && outputRaw != nil {
		err = UnsuccessfulItemsError(outputRaw.(*ec2.CreateFlowLogsOutput).Unsuccessful)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Flow Log (%s): %s", resourceID, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*ec2.CreateFlowLogsOutput).FlowLogIds[0]))

	return append(diags, resourceLogFlowRead(ctx, d, meta)...)
}

func resourceLogFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	fl, err := FindFlowLogByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Flow Log %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Flow Log (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-flow-log/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("deliver_cross_account_role", fl.DeliverCrossAccountRole)
	if fl.DestinationOptions != nil {
		if err := d.Set("destination_options", []interface{}{flattenDestinationOptionsResponse(fl.DestinationOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting destination_options: %s", err)
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
	switch resourceID := aws.StringValue(fl.ResourceId); {
	case strings.HasPrefix(resourceID, "vpc-"):
		d.Set("vpc_id", resourceID)
	case strings.HasPrefix(resourceID, "tgw-"):
		if strings.HasPrefix(resourceID, "tgw-attach-") {
			d.Set("transit_gateway_attachment_id", resourceID)
		} else {
			d.Set("transit_gateway_id", resourceID)
		}
	case strings.HasPrefix(resourceID, "subnet-"):
		d.Set("subnet_id", resourceID)
	case strings.HasPrefix(resourceID, "eni-"):
		d.Set("eni_id", resourceID)
	}
	if !strings.HasPrefix(aws.StringValue(fl.ResourceId), "tgw-") {
		d.Set("traffic_type", fl.TrafficType)
	}

	setTagsOut(ctx, fl.Tags)

	return diags
}

func resourceLogFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLogFlowRead(ctx, d, meta)...)
}

func resourceLogFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[INFO] Deleting Flow Log: %s", d.Id())
	output, err := conn.DeleteFlowLogsWithContext(ctx, &ec2.DeleteFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{d.Id()}),
	})

	if err == nil && output != nil {
		err = UnsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidFlowLogIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Flow Log (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDestinationOptionsRequest(tfMap map[string]interface{}) *ec2.DestinationOptionsRequest {
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

func flattenDestinationOptionsResponse(apiObject *ec2.DestinationOptionsResponse) map[string]interface{} {
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
