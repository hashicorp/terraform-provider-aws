// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_network_insights_path", name="Network Insights Path")
// @Tags(identifierAttribute="id")
func ResourceNetworkInsightsPath() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInsightsPathCreate,
		ReadWithoutTimeout:   resourceNetworkInsightsPathRead,
		UpdateWithoutTimeout: resourceNetworkInsightsPathUpdate,
		DeleteWithoutTimeout: resourceNetworkInsightsPathDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.Protocol_Values(), false),
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_ip": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkInsightsPathCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.CreateNetworkInsightsPathInput{
		Destination:       aws.String(d.Get("destination").(string)),
		Protocol:          aws.String(d.Get("protocol").(string)),
		Source:            aws.String(d.Get("source").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeNetworkInsightsPath),
	}

	if v, ok := d.GetOk("destination_ip"); ok {
		input.DestinationIp = aws.String(v.(string))
	}

	if v, ok := d.GetOk("destination_port"); ok {
		input.DestinationPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("source_ip"); ok {
		input.SourceIp = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Network Insights Path: %s", input)
	output, err := conn.CreateNetworkInsightsPathWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating EC2 Network Insights Path: %s", err)
	}

	d.SetId(aws.StringValue(output.NetworkInsightsPath.NetworkInsightsPathId))

	return resourceNetworkInsightsPathRead(ctx, d, meta)
}

func resourceNetworkInsightsPathRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	nip, err := FindNetworkInsightsPathByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Insights Path %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EC2 Network Insights Path (%s): %s", d.Id(), err)
	}

	d.Set("arn", nip.NetworkInsightsPathArn)
	d.Set("destination", nip.Destination)
	d.Set("destination_ip", nip.DestinationIp)
	d.Set("destination_port", nip.DestinationPort)
	d.Set("protocol", nip.Protocol)
	d.Set("source", nip.Source)
	d.Set("source_ip", nip.SourceIp)

	setTagsOut(ctx, nip.Tags)

	return nil
}

func resourceNetworkInsightsPathUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceNetworkInsightsPathRead(ctx, d, meta)
}

func resourceNetworkInsightsPathDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Network Insights Path: %s", d.Id())
	_, err := conn.DeleteNetworkInsightsPathWithContext(ctx, &ec2.DeleteNetworkInsightsPathInput{
		NetworkInsightsPathId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsPathIdNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EC2 Network Insights Path (%s): %s", d.Id(), err)
	}

	return nil
}
