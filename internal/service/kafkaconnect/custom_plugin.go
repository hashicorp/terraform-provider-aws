// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafkaconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_mskconnect_custom_plugin", name="Custom Plugin")
// @Tags(identifierAttribute="arn")
func resourceCustomPlugin() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomPluginCreate,
		ReadWithoutTimeout:   resourceCustomPluginRead,
		UpdateWithoutTimeout: resourceCustomPluginUpdate,
		DeleteWithoutTimeout: resourceCustomPluginDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContentType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CustomPluginContentType](),
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
			names.AttrLocation: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3": {
							Type:     schema.TypeList,
							MaxItems: 1,
							ForceNew: true,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"file_key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"object_version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomPluginCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kafkaconnect.CreateCustomPluginInput{
		ContentType: awstypes.CustomPluginContentType(d.Get(names.AttrContentType).(string)),
		Location:    expandCustomPluginLocation(d.Get(names.AttrLocation).([]interface{})[0].(map[string]interface{})),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateCustomPlugin(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Connect Custom Plugin (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CustomPluginArn))

	if _, err := waitCustomPluginCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Custom Plugin (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCustomPluginRead(ctx, d, meta)...)
}

func resourceCustomPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	plugin, err := findCustomPluginByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Custom Plugin (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Connect Custom Plugin (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, plugin.CustomPluginArn)
	d.Set(names.AttrDescription, plugin.Description)
	d.Set(names.AttrName, plugin.Name)
	d.Set(names.AttrState, plugin.CustomPluginState)

	if plugin.LatestRevision != nil {
		d.Set(names.AttrContentType, plugin.LatestRevision.ContentType)
		d.Set("latest_revision", plugin.LatestRevision.Revision)
		if plugin.LatestRevision.Location != nil {
			if err := d.Set(names.AttrLocation, []interface{}{flattenCustomPluginLocationDescription(plugin.LatestRevision.Location)}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting location: %s", err)
			}
		} else {
			d.Set(names.AttrLocation, nil)
		}
	} else {
		d.Set(names.AttrContentType, nil)
		d.Set("latest_revision", nil)
		d.Set(names.AttrLocation, nil)
	}

	return diags
}

func resourceCustomPluginUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// This update function is for updating tags only - there is no update action for this resource.

	return append(diags, resourceCustomPluginRead(ctx, d, meta)...)
}

func resourceCustomPluginDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	log.Printf("[DEBUG] Deleting MSK Connect Custom Plugin: %s", d.Id())
	_, err := conn.DeleteCustomPlugin(ctx, &kafkaconnect.DeleteCustomPluginInput{
		CustomPluginArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Connect Custom Plugin (%s): %s", d.Id(), err)
	}

	if _, err := waitCustomPluginDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Custom Plugin (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findCustomPluginByARN(ctx context.Context, conn *kafkaconnect.Client, arn string) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	input := &kafkaconnect.DescribeCustomPluginInput{
		CustomPluginArn: aws.String(arn),
	}

	output, err := conn.DescribeCustomPlugin(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusCustomPlugin(ctx context.Context, conn *kafkaconnect.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCustomPluginByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.CustomPluginState), nil
	}
}

func waitCustomPluginCreated(ctx context.Context, conn *kafkaconnect.Client, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CustomPluginStateCreating),
		Target:  enum.Slice(awstypes.CustomPluginStateActive),
		Refresh: statusCustomPlugin(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		if state, stateDescription := output.CustomPluginState, output.StateDescription; state == awstypes.CustomPluginStateCreateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateDescription.Code), aws.ToString(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitCustomPluginDeleted(ctx context.Context, conn *kafkaconnect.Client, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CustomPluginStateDeleting),
		Target:  []string{},
		Refresh: statusCustomPlugin(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		return output, err
	}

	return nil, err
}

func expandCustomPluginLocation(tfMap map[string]interface{}) *awstypes.CustomPluginLocation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomPluginLocation{}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3Location = expandS3Location(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandS3Location(tfMap map[string]interface{}) *awstypes.S3Location {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Location{}

	if v, ok := tfMap["bucket_arn"].(string); ok && v != "" {
		apiObject.BucketArn = aws.String(v)
	}

	if v, ok := tfMap["file_key"].(string); ok && v != "" {
		apiObject.FileKey = aws.String(v)
	}

	if v, ok := tfMap["object_version"].(string); ok && v != "" {
		apiObject.ObjectVersion = aws.String(v)
	}

	return apiObject
}

func flattenCustomPluginLocationDescription(apiObject *awstypes.CustomPluginLocationDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Location; v != nil {
		tfMap["s3"] = []interface{}{flattenS3LocationDescription(v)}
	}

	return tfMap
}

func flattenS3LocationDescription(apiObject *awstypes.S3LocationDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketArn; v != nil {
		tfMap["bucket_arn"] = aws.ToString(v)
	}

	if v := apiObject.FileKey; v != nil {
		tfMap["file_key"] = aws.ToString(v)
	}

	if v := apiObject.ObjectVersion; v != nil {
		tfMap["object_version"] = aws.ToString(v)
	}

	return tfMap
}
