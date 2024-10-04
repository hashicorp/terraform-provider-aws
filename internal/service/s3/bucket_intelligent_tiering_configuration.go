// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_intelligent_tiering_configuration", name="Bucket Intelligent-Tiering Configuration")
func resourceBucketIntelligentTieringConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketIntelligentTieringConfigurationPut,
		ReadWithoutTimeout:   resourceBucketIntelligentTieringConfigurationRead,
		UpdateWithoutTimeout: resourceBucketIntelligentTieringConfigurationPut,
		DeleteWithoutTimeout: resourceBucketIntelligentTieringConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrFilter: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrPrefix: {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
						names.AttrTags: {
							Type:         schema.TypeMap,
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.IntelligentTieringStatusEnabled,
				ValidateDiagFunc: enum.Validate[types.IntelligentTieringStatus](),
			},
			"tiering": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_tier": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.IntelligentTieringAccessTier](),
						},
						"days": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceBucketIntelligentTieringConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	name := d.Get(names.AttrName).(string)
	intelligentTieringConfiguration := &types.IntelligentTieringConfiguration{
		Id:     aws.String(name),
		Status: types.IntelligentTieringStatus(d.Get(names.AttrStatus).(string)),
	}

	if v, ok := d.GetOk(names.AttrFilter); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		intelligentTieringConfiguration.Filter = expandIntelligentTieringFilter(ctx, v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("tiering"); ok && v.(*schema.Set).Len() > 0 {
		intelligentTieringConfiguration.Tierings = expandTierings(v.(*schema.Set).List())
	}

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3.PutBucketIntelligentTieringConfigurationInput{
		Bucket:                          aws.String(bucket),
		Id:                              aws.String(name),
		IntelligentTieringConfiguration: intelligentTieringConfiguration,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketIntelligentTieringConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "IntelligentTieringConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Intelligent-Tiering Configuration (%s): %s", bucket, name, err)
	}

	if d.IsNewResource() {
		d.SetId(BucketIntelligentTieringConfigurationCreateResourceID(bucket, name))

		_, err = tfresource.RetryWhenNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
			return findIntelligentTieringConfiguration(ctx, conn, bucket, name)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Intelligent-Tiering Configuration (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketIntelligentTieringConfigurationRead(ctx, d, meta)...)
}

func resourceBucketIntelligentTieringConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketIntelligentTieringConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findIntelligentTieringConfiguration(ctx, conn, bucket, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Intelligent-Tiering Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if output.Filter != nil {
		if err := d.Set(names.AttrFilter, []interface{}{flattenIntelligentTieringFilter(ctx, output.Filter)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
		}
	} else {
		d.Set(names.AttrFilter, nil)
	}
	d.Set(names.AttrName, output.Id)
	d.Set(names.AttrStatus, output.Status)
	if err := d.Set("tiering", flattenTierings(output.Tierings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tiering: %s", err)
	}

	return diags
}

func resourceBucketIntelligentTieringConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketIntelligentTieringConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Intelligent-Tiering Configuration: %s", d.Id())
	_, err = conn.DeleteBucketIntelligentTieringConfiguration(ctx, &s3.DeleteBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, bucketPropagationTimeout, func() (interface{}, error) {
		return findIntelligentTieringConfiguration(ctx, conn, bucket, name)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Intelligent-Tiering Configuration (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const bucketIntelligentTieringConfigurationResourceIDSeparator = ":"

func BucketIntelligentTieringConfigurationCreateResourceID(bucketName, configurationName string) string {
	parts := []string{bucketName, configurationName}
	id := strings.Join(parts, bucketIntelligentTieringConfigurationResourceIDSeparator)

	return id
}

func BucketIntelligentTieringConfigurationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, bucketIntelligentTieringConfigurationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected bucket-name%[2]sconfiguration-name", id, bucketIntelligentTieringConfigurationResourceIDSeparator)
}

func findIntelligentTieringConfiguration(ctx context.Context, conn *s3.Client, bucket, name string) (*types.IntelligentTieringConfiguration, error) {
	input := &s3.GetBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	output, err := conn.GetBucketIntelligentTieringConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IntelligentTieringConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.IntelligentTieringConfiguration, nil
}

func expandIntelligentTieringFilter(ctx context.Context, tfMap map[string]interface{}) *types.IntelligentTieringFilter {
	if tfMap == nil {
		return nil
	}

	var prefix string

	if v, ok := tfMap[names.AttrPrefix].(string); ok {
		prefix = v
	}

	var tags []types.Tag

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok {
		tags = Tags(tftags.New(ctx, v))
	}

	apiObject := &types.IntelligentTieringFilter{}

	if prefix == "" {
		switch len(tags) {
		case 0:
			return nil
		case 1:
			apiObject.Tag = &tags[0]
		default:
			apiObject.And = &types.IntelligentTieringAndOperator{
				Tags: tags,
			}
		}
	} else {
		switch len(tags) {
		case 0:
			apiObject.Prefix = aws.String(prefix)
		default:
			apiObject.And = &types.IntelligentTieringAndOperator{
				Prefix: aws.String(prefix),
				Tags:   tags,
			}
		}
	}

	return apiObject
}

func expandTiering(tfMap map[string]interface{}) *types.Tiering {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Tiering{}

	if v, ok := tfMap["access_tier"].(string); ok && v != "" {
		apiObject.AccessTier = types.IntelligentTieringAccessTier(v)
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTierings(tfList []interface{}) []types.Tiering {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Tiering

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTiering(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenIntelligentTieringFilter(ctx context.Context, apiObject *types.IntelligentTieringFilter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.And == nil {
		if v := apiObject.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}

		if v := apiObject.Tag; v != nil {
			tfMap[names.AttrTags] = keyValueTags(ctx, []types.Tag{*v}).Map()
		}
	} else {
		apiObject := apiObject.And

		if v := apiObject.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}

		if v := apiObject.Tags; v != nil {
			tfMap[names.AttrTags] = keyValueTags(ctx, v).Map()
		}
	}

	return tfMap
}

func flattenTiering(apiObject types.Tiering) map[string]interface{} {
	tfMap := map[string]interface{}{
		"access_tier": apiObject.AccessTier,
		"days":        apiObject.Days,
	}

	return tfMap
}

func flattenTierings(apiObjects []types.Tiering) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTiering(apiObject))
	}

	return tfList
}
