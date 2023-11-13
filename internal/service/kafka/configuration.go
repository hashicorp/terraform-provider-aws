// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_msk_configuration")
func ResourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationCreate,
		ReadWithoutTimeout:   resourceConfigurationRead,
		UpdateWithoutTimeout: resourceConfigurationUpdate,
		DeleteWithoutTimeout: resourceConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("latest_revision", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("server_properties")
			}),
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kafka_versions": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_properties": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	input := &kafka.CreateConfigurationInput{
		Name:             aws.String(d.Get("name").(string)),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kafka_versions"); ok && v.(*schema.Set).Len() > 0 {
		input.KafkaVersions = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Configuration: %s", err)
	}

	d.SetId(aws.StringValue(output.Arn))

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	configurationInput := &kafka.DescribeConfigurationInput{
		Arn: aws.String(d.Id()),
	}

	configurationOutput, err := conn.DescribeConfigurationWithContext(ctx, configurationInput)

	if tfawserr.ErrMessageContains(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
		log.Printf("[WARN] MSK Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s): %s", d.Id(), err)
	}

	if configurationOutput == nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s): missing result", d.Id())
	}

	if configurationOutput.LatestRevision == nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s): missing latest revision", d.Id())
	}

	revision := configurationOutput.LatestRevision.Revision
	revisionInput := &kafka.DescribeConfigurationRevisionInput{
		Arn:      aws.String(d.Id()),
		Revision: revision,
	}

	revisionOutput, err := conn.DescribeConfigurationRevisionWithContext(ctx, revisionInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s) Revision (%d): %s", d.Id(), aws.Int64Value(revision), err)
	}

	if revisionOutput == nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s) Revision (%d): missing result", d.Id(), aws.Int64Value(revision))
	}

	d.Set("arn", configurationOutput.Arn)
	d.Set("description", revisionOutput.Description)

	if err := d.Set("kafka_versions", aws.StringValueSlice(configurationOutput.KafkaVersions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kafka_versions: %s", err)
	}

	d.Set("latest_revision", revision)
	d.Set("name", configurationOutput.Name)
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	return diags
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	input := &kafka.UpdateConfigurationInput{
		Arn:              aws.String(d.Id()),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating MSK Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func resourceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn(ctx)

	input := &kafka.DeleteConfigurationInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting MSK Configuration: %s", d.Id())
	_, err := conn.DeleteConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitConfigurationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Configuration (%s): %s", d.Id(), err)
	}

	return diags
}
