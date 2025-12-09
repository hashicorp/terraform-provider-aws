// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func networkInsightsFilterPortRangeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"from_port": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 65535),
		},
		"to_port": {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 65535),
		},
	}
}

func networkInsightsPathFilterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"destination_address": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(0, 15),
				validation.StringMatch(
					regexache.MustCompile(`^([0-9]{1,3}.){3}[0-9]{1,3}$`),
					"must be a valid IPv4 address"),
			),
		},
		"destination_port_range": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: networkInsightsFilterPortRangeSchema(),
			},
		},
		"source_address": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(0, 15),
				validation.StringMatch(
					regexache.MustCompile(`^([0-9]{1,3}.){3}[0-9]{1,3}$`),
					"must be a valid IPv4 address"),
			),
		},
		"source_port_range": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 1,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: networkInsightsFilterPortRangeSchema(),
			},
		},
	}
}

// @SDKResource("aws_ec2_network_insights_path", name="Network Insights Path")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceNetworkInsightsPath() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInsightsPathCreate,
		ReadWithoutTimeout:   resourceNetworkInsightsPathRead,
		UpdateWithoutTimeout: resourceNetworkInsightsPathUpdate,
		DeleteWithoutTimeout: resourceNetworkInsightsPathDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestinationARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDestination: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentIDOrARN,
			},
			"destination_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"filter_at_destination": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: networkInsightsPathFilterSchema(),
				},
			},
			"filter_at_source": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: networkInsightsPathFilterSchema(),
				},
			},
			names.AttrProtocol: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Protocol](),
			},
			names.AttrSource: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressEquivalentIDOrARN,
			},
			"source_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceNetworkInsightsPathCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := ec2.CreateNetworkInsightsPathInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		Protocol:          awstypes.Protocol(d.Get(names.AttrProtocol).(string)),
		Source:            aws.String(d.Get(names.AttrSource).(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeNetworkInsightsPath),
	}

	if v, ok := d.GetOk(names.AttrDestination); ok {
		input.Destination = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_ip"); ok {
		input.DestinationIp = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_port"); ok {
		input.DestinationPort = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("filter_at_destination"); ok && len(v.([]any)) > 0 {
		input.FilterAtDestination = expandPathRequestFilter(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("filter_at_source"); ok && len(v.([]any)) > 0 {
		input.FilterAtSource = expandPathRequestFilter(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("source_ip"); ok {
		input.SourceIp = aws.String(v.(string))
	}

	output, err := conn.CreateNetworkInsightsPath(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network Insights Path: %s", err)
	}

	d.SetId(aws.ToString(output.NetworkInsightsPath.NetworkInsightsPathId))

	return append(diags, resourceNetworkInsightsPathRead(ctx, d, meta)...)
}

func resourceNetworkInsightsPathRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	nip, err := findNetworkInsightsPathByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Insights Path %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Insights Path (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, nip.NetworkInsightsPathArn)
	d.Set(names.AttrDestination, nip.Destination)
	d.Set(names.AttrDestinationARN, nip.DestinationArn)
	d.Set("destination_ip", nip.DestinationIp)
	d.Set("destination_port", nip.DestinationPort)
	if v := nip.FilterAtDestination; v != nil {
		d.Set("filter_at_destination", []any{flattenPathFilter(v)})
	}
	if v := nip.FilterAtSource; v != nil {
		d.Set("filter_at_source", []any{flattenPathFilter(v)})
	}
	d.Set(names.AttrProtocol, nip.Protocol)
	d.Set(names.AttrSource, nip.Source)
	d.Set("source_arn", nip.SourceArn)
	d.Set("source_ip", nip.SourceIp)

	setTagsOut(ctx, nip.Tags)

	return diags
}

func resourceNetworkInsightsPathUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceNetworkInsightsPathRead(ctx, d, meta)
}

func resourceNetworkInsightsPathDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Network Insights Path: %s", d.Id())
	input := ec2.DeleteNetworkInsightsPathInput{
		NetworkInsightsPathId: aws.String(d.Id()),
	}
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.DeleteNetworkInsightsPath(ctx, &input)
	}, errCodeAnalysisExistsForNetworkInsightsPath)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsPathIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Network Insights Path (%s): %s", d.Id(), err)
	}

	return diags
}

// idFromIDOrARN return a resource ID from an ID or ARN.
func idFromIDOrARN(idOrARN string) string {
	// e.g. "eni-02ae120b80627a68f" or
	// "arn:aws:ec2:ap-southeast-2:123456789012:network-interface/eni-02ae120b80627a68f".
	return idOrARN[strings.LastIndex(idOrARN, "/")+1:]
}

// suppressEquivalentIDOrARN provides custom difference suppression
// for strings that represent equal resource IDs or ARNs.
func suppressEquivalentIDOrARN(_, old, new string, _ *schema.ResourceData) bool {
	return idFromIDOrARN(old) == idFromIDOrARN(new)
}

func expandRequestFilterPortRange(tfMap map[string]any) *awstypes.RequestFilterPortRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RequestFilterPortRange{}

	if v, ok := tfMap["from_port"]; ok {
		apiObject.FromPort = aws.Int32(int32(v.(int)))
	}
	if v, ok := tfMap["to_port"]; ok {
		apiObject.ToPort = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func expandPathRequestFilter(tfMap map[string]any) *awstypes.PathRequestFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.PathRequestFilter{}

	if v, ok := tfMap["destination_address"].(string); ok && v != "" {
		apiObject.DestinationAddress = aws.String(v)
	}
	if v, ok := tfMap["destination_port_range"].([]any); ok && len(v) > 0 {
		apiObject.DestinationPortRange = expandRequestFilterPortRange(v[0].(map[string]any))
	}
	if v, ok := tfMap["source_address"].(string); ok && v != "" {
		apiObject.SourceAddress = aws.String(v)
	}
	if v, ok := tfMap["source_port_range"].([]any); ok && len(v) > 0 {
		apiObject.SourcePortRange = expandRequestFilterPortRange(v[0].(map[string]any))
	}

	return apiObject
}

func flattenFilterPortRange(portRange *awstypes.FilterPortRange) map[string]any {
	if portRange == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := portRange.FromPort; v != nil {
		tfMap["from_port"] = aws.ToInt32(v)
	}

	if v := portRange.ToPort; v != nil {
		tfMap["to_port"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenPathFilter(requestFilter *awstypes.PathFilter) map[string]any {
	if requestFilter == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := requestFilter.DestinationAddress; v != nil {
		tfMap["destination_address"] = aws.ToString(v)
	}

	if v := requestFilter.DestinationPortRange; v != nil {
		tfMap["destination_port_range"] = []any{flattenFilterPortRange(v)}
	}

	if v := requestFilter.SourceAddress; v != nil {
		tfMap["source_address"] = aws.ToString(v)
	}

	if v := requestFilter.SourcePortRange; v != nil {
		tfMap["source_port_range"] = []any{flattenFilterPortRange(v)}
	}

	return tfMap
}
