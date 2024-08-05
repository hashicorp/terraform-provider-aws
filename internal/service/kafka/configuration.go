// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_configuration", name="Configuration")
func resourceConfiguration() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	input := &kafka.CreateConfigurationInput{
		Name:             aws.String(d.Get(names.AttrName).(string)),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kafka_versions"); ok && v.(*schema.Set).Len() > 0 {
		input.KafkaVersions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Configuration: %s", err)
	}

	d.SetId(aws.ToString(output.Arn))

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	configurationOutput, err := findConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Configuration (%s): %s", d.Id(), err)
	}

	revision := aws.ToInt64(configurationOutput.LatestRevision.Revision)
	revisionOutput, err := findConfigurationRevisionByTwoPartKey(ctx, conn, d.Id(), revision)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Configuration (%s) revision (%d): %s", d.Id(), revision, err)
	}

	d.Set(names.AttrARN, configurationOutput.Arn)
	d.Set(names.AttrDescription, revisionOutput.Description)
	d.Set("kafka_versions", configurationOutput.KafkaVersions)
	d.Set("latest_revision", revision)
	d.Set(names.AttrName, configurationOutput.Name)
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	return diags
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	input := &kafka.UpdateConfigurationInput{
		Arn:              aws.String(d.Id()),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating MSK Configuration (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func resourceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[DEBUG] Deleting MSK Configuration: %s", d.Id())
	_, err := conn.DeleteConfiguration(ctx, &kafka.DeleteConfigurationInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Configuration ARN does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Configuration (%s): %s", d.Id(), err)
	}

	if _, err := waitConfigurationDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeConfigurationOutput, error) {
	input := &kafka.DescribeConfigurationInput{
		Arn: aws.String(arn),
	}

	output, err := conn.DescribeConfiguration(ctx, input)

	if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Configuration ARN does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LatestRevision == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findConfigurationRevisionByTwoPartKey(ctx context.Context, conn *kafka.Client, arn string, revision int64) (*kafka.DescribeConfigurationRevisionOutput, error) {
	input := &kafka.DescribeConfigurationRevisionInput{
		Arn:      aws.String(arn),
		Revision: aws.Int64(revision),
	}

	output, err := conn.DescribeConfigurationRevision(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusConfigurationState(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitConfigurationDeleted(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeConfigurationOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ConfigurationStateDeleting),
		Target:  []string{},
		Refresh: statusConfigurationState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafka.DescribeConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
