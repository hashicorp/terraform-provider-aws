// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_flow_log", name="Flow Log")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceFlowLog() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLogFlowCreate,
		ReadWithoutTimeout:   resourceLogFlowRead,
		UpdateWithoutTimeout: resourceLogFlowUpdate,
		DeleteWithoutTimeout: resourceLogFlowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DestinationFileFormatPlainText,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DestinationFileFormat](),
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
				ExactlyOneOf: []string{"eni_id", names.AttrSubnetID, names.AttrVPCID, names.AttrTransitGatewayID, names.AttrTransitGatewayAttachmentID},
			},
			names.AttrIAMRoleARN: {
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
				ConflictsWith: []string{names.AttrLogGroupName},
			},
			"log_destination_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.LogDestinationTypeCloudWatchLogs,
				ValidateDiagFunc: enum.Validate[awstypes.LogDestinationType](),
			},
			"log_format": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			names.AttrLogGroupName: {
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
			names.AttrSubnetID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", names.AttrSubnetID, names.AttrVPCID, names.AttrTransitGatewayID, names.AttrTransitGatewayAttachmentID},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"traffic_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TrafficType](),
			},
			names.AttrTransitGatewayAttachmentID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", names.AttrSubnetID, names.AttrVPCID, names.AttrTransitGatewayID, names.AttrTransitGatewayAttachmentID},
			},
			names.AttrTransitGatewayID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", names.AttrSubnetID, names.AttrVPCID, names.AttrTransitGatewayID, names.AttrTransitGatewayAttachmentID},
			},
			names.AttrVPCID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"eni_id", names.AttrSubnetID, names.AttrVPCID, names.AttrTransitGatewayID, names.AttrTransitGatewayAttachmentID},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLogFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	var resourceID string
	var resourceType awstypes.FlowLogsResourceType
	for _, v := range []struct {
		ID   string
		Type awstypes.FlowLogsResourceType
	}{
		{
			ID:   d.Get(names.AttrVPCID).(string),
			Type: awstypes.FlowLogsResourceTypeVpc,
		},
		{
			ID:   d.Get(names.AttrTransitGatewayID).(string),
			Type: awstypes.FlowLogsResourceTypeTransitGateway,
		},
		{
			ID:   d.Get(names.AttrTransitGatewayAttachmentID).(string),
			Type: awstypes.FlowLogsResourceTypeTransitGatewayAttachment,
		},
		{
			ID:   d.Get(names.AttrSubnetID).(string),
			Type: awstypes.FlowLogsResourceTypeSubnet,
		},
		{
			ID:   d.Get("eni_id").(string),
			Type: awstypes.FlowLogsResourceTypeNetworkInterface,
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
		LogDestinationType: awstypes.LogDestinationType(d.Get("log_destination_type").(string)),
		ResourceIds:        []string{resourceID},
		ResourceType:       resourceType,
		TagSpecifications:  getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpcFlowLog),
	}

	if resourceType != awstypes.FlowLogsResourceTypeTransitGateway && resourceType != awstypes.FlowLogsResourceTypeTransitGatewayAttachment {
		if v, ok := d.GetOk("traffic_type"); ok {
			input.TrafficType = awstypes.TrafficType(v.(string))
		}
	}

	if v, ok := d.GetOk("destination_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationOptions = expandDestinationOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("deliver_cross_account_role"); ok {
		input.DeliverCrossAccountRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIAMRoleARN); ok {
		input.DeliverLogsPermissionArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_destination"); ok {
		input.LogDestination = aws.String(strings.TrimSuffix(v.(string), ":*"))
	}

	if v, ok := d.GetOk("log_format"); ok {
		input.LogFormat = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrLogGroupName); ok {
		input.LogGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_aggregation_interval"); ok {
		input.MaxAggregationInterval = aws.Int32(int32(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, iamPropagationTimeout, func() (interface{}, error) {
		return conn.CreateFlowLogs(ctx, input)
	}, errCodeInvalidParameter, "Unable to assume given IAM role")

	if err == nil && outputRaw != nil {
		err = unsuccessfulItemsError(outputRaw.(*ec2.CreateFlowLogsOutput).Unsuccessful)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Flow Log (%s): %s", resourceID, err)
	}

	d.SetId(outputRaw.(*ec2.CreateFlowLogsOutput).FlowLogIds[0])

	return append(diags, resourceLogFlowRead(ctx, d, meta)...)
}

func resourceLogFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	fl, err := findFlowLogByID(ctx, conn, d.Id())

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
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-flow-log/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("deliver_cross_account_role", fl.DeliverCrossAccountRole)
	if fl.DestinationOptions != nil {
		if err := d.Set("destination_options", []interface{}{flattenDestinationOptionsResponse(fl.DestinationOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting destination_options: %s", err)
		}
	} else {
		d.Set("destination_options", nil)
	}
	d.Set(names.AttrIAMRoleARN, fl.DeliverLogsPermissionArn)
	d.Set("log_destination", fl.LogDestination)
	d.Set("log_destination_type", fl.LogDestinationType)
	d.Set("log_format", fl.LogFormat)
	d.Set(names.AttrLogGroupName, fl.LogGroupName)
	d.Set("max_aggregation_interval", fl.MaxAggregationInterval)
	switch resourceID := aws.ToString(fl.ResourceId); {
	case strings.HasPrefix(resourceID, "vpc-"):
		d.Set(names.AttrVPCID, resourceID)
	case strings.HasPrefix(resourceID, "tgw-"):
		if strings.HasPrefix(resourceID, "tgw-attach-") {
			d.Set(names.AttrTransitGatewayAttachmentID, resourceID)
		} else {
			d.Set(names.AttrTransitGatewayID, resourceID)
		}
	case strings.HasPrefix(resourceID, "subnet-"):
		d.Set(names.AttrSubnetID, resourceID)
	case strings.HasPrefix(resourceID, "eni-"):
		d.Set("eni_id", resourceID)
	}
	if !strings.HasPrefix(aws.ToString(fl.ResourceId), "tgw-") {
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Flow Log: %s", d.Id())
	output, err := conn.DeleteFlowLogs(ctx, &ec2.DeleteFlowLogsInput{
		FlowLogIds: []string{d.Id()},
	})

	if err == nil && output != nil {
		err = unsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidFlowLogIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Flow Log (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDestinationOptionsRequest(tfMap map[string]interface{}) *awstypes.DestinationOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DestinationOptionsRequest{}

	if v, ok := tfMap["file_format"].(string); ok && v != "" {
		apiObject.FileFormat = awstypes.DestinationFileFormat(v)
	}

	if v, ok := tfMap["hive_compatible_partitions"].(bool); ok {
		apiObject.HiveCompatiblePartitions = aws.Bool(v)
	}

	if v, ok := tfMap["per_hour_partition"].(bool); ok {
		apiObject.PerHourPartition = aws.Bool(v)
	}

	return apiObject
}

func flattenDestinationOptionsResponse(apiObject *awstypes.DestinationOptionsResponse) map[string]interface{} {
	tfMap := map[string]interface{}{
		"file_format": apiObject.FileFormat,
	}

	if v := apiObject.HiveCompatiblePartitions; v != nil {
		tfMap["hive_compatible_partitions"] = aws.ToBool(v)
	}

	if v := apiObject.PerHourPartition; v != nil {
		tfMap["per_hour_partition"] = aws.ToBool(v)
	}

	return tfMap
}
