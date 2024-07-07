// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_mskconnect_worker_configuration")
// @Tags(identifierAttribute="arn")
func ResourceWorkerConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkerConfigurationCreate,
		ReadWithoutTimeout:   resourceWorkerConfigurationRead,
		UpdateWithoutTimeout: resourceWorkerConfigurationUpdate,
		DeleteWithoutTimeout: resourceWorkerConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"properties_file_content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return decodePropertiesFileContent(v)
					default:
						return ""
					}
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkerConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaConnectConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kafkaconnect.CreateWorkerConfigurationInput{
		Name:                  aws.String(name),
		PropertiesFileContent: flex.StringValueToBase64String(d.Get("properties_file_content").(string)),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating MSK Connect Worker Configuration: %s", input)
	output, err := conn.CreateWorkerConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Connect Worker Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.WorkerConfigurationArn))

	return append(diags, resourceWorkerConfigurationRead(ctx, d, meta)...)
}

func resourceWorkerConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaConnectConn(ctx)

	config, err := FindWorkerConfigurationByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Worker Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Connect Worker Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, config.WorkerConfigurationArn)
	d.Set(names.AttrDescription, config.Description)
	d.Set(names.AttrName, config.Name)

	if config.LatestRevision != nil {
		d.Set("latest_revision", config.LatestRevision.Revision)
		d.Set("properties_file_content", decodePropertiesFileContent(aws.StringValue(config.LatestRevision.PropertiesFileContent)))
	} else {
		d.Set("latest_revision", nil)
		d.Set("properties_file_content", nil)
	}

	return diags
}

func resourceWorkerConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// This update function is for updating tags only - there is no update action for this resource

	return append(diags, resourceWorkerConfigurationRead(ctx, d, meta)...)
}

func decodePropertiesFileContent(content string) string {
	v, err := itypes.Base64Decode(content)
	if err != nil {
		return content
	}

	return string(v)
}

func resourceWorkerConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaConnectConn(ctx)

	log.Printf("[DEBUG] Deleting MSK Connect Worker Configuration: %s", d.Id())
	_, err := conn.DeleteWorkerConfigurationWithContext(ctx, &kafkaconnect.DeleteWorkerConfigurationInput{
		WorkerConfigurationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kafkaconnect.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Connect Worker Configuration (%s): %s", d.Id(), err)
	}

	_, err = waitWorkerConfigurationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Worker Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}
