package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceBucketIntelligentTieringConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketIntelligentTieringConfigurationPut,
		ReadWithoutTimeout:   resourceBucketIntelligentTieringConfigurationRead,
		UpdateWithoutTimeout: resourceBucketIntelligentTieringConfigurationPut,
		DeleteWithoutTimeout: resourceBucketIntelligentTieringConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
						"tags": {
							Type:         schema.TypeMap,
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							AtLeastOneOf: []string{"filter.0.prefix", "filter.0.tags"},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      s3.IntelligentTieringStatusEnabled,
				ValidateFunc: validation.StringInSlice(s3.IntelligentTieringStatus_Values(), false),
			},
			"tiering": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_tier": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.IntelligentTieringAccessTier_Values(), false),
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
	conn := meta.(*conns.AWSClient).S3Conn()

	bucketName := d.Get("bucket").(string)
	configurationName := d.Get("name").(string)
	resourceID := BucketIntelligentTieringConfigurationCreateResourceID(bucketName, configurationName)
	apiObject := &s3.IntelligentTieringConfiguration{
		Id:     aws.String(configurationName),
		Status: aws.String(d.Get("status").(string)),
	}

	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Filter = expandIntelligentTieringFilter(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("tiering"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.Tierings = expandTierings(v.(*schema.Set).List())
	}

	input := &s3.PutBucketIntelligentTieringConfigurationInput{
		Bucket:                          aws.String(bucketName),
		Id:                              aws.String(configurationName),
		IntelligentTieringConfiguration: apiObject,
	}

	log.Printf("[DEBUG] Creating S3 Intelligent-Tiering Configuration: %s", input)
	_, err := retryWhenBucketNotFound(ctx, func() (interface{}, error) {
		return conn.PutBucketIntelligentTieringConfigurationWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Intelligent-Tiering Configuration (%s): %s", resourceID, err)
	}

	d.SetId(resourceID)

	return append(diags, resourceBucketIntelligentTieringConfigurationRead(ctx, d, meta)...)
}

func resourceBucketIntelligentTieringConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucketName, configurationName, err := BucketIntelligentTieringConfigurationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
	}

	output, err := FindBucketIntelligentTieringConfiguration(ctx, conn, bucketName, configurationName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Intelligent-Tiering Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucketName)
	if output.Filter != nil {
		if err := d.Set("filter", []interface{}{flattenIntelligentTieringFilter(output.Filter)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
		}
	} else {
		d.Set("filter", nil)
	}
	d.Set("name", output.Id)
	d.Set("status", output.Status)
	if err := d.Set("tiering", flattenTierings(output.Tierings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tiering: %s", err)
	}

	return diags
}

func resourceBucketIntelligentTieringConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Conn()

	bucketName, configurationName, err := BucketIntelligentTieringConfigurationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting S3 Intelligent-Tiering Configuration: (%s)", d.Id())
	_, err = conn.DeleteBucketIntelligentTieringConfigurationWithContext(ctx, &s3.DeleteBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucketName),
		Id:     aws.String(configurationName),
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Intelligent-Tiering Configuration (%s): %s", d.Id(), err)
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

func FindBucketIntelligentTieringConfiguration(ctx context.Context, conn *s3.S3, bucketName, configurationName string) (*s3.IntelligentTieringConfiguration, error) {
	input := &s3.GetBucketIntelligentTieringConfigurationInput{
		Bucket: aws.String(bucketName),
		Id:     aws.String(configurationName),
	}

	output, err := conn.GetBucketIntelligentTieringConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeNoSuchConfiguration) {
		return nil, &resource.NotFoundError{
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

func expandIntelligentTieringFilter(tfMap map[string]interface{}) *s3.IntelligentTieringFilter {
	if tfMap == nil {
		return nil
	}

	var prefix string

	if v, ok := tfMap["prefix"].(string); ok {
		prefix = v
	}

	var tags []*s3.Tag

	if v, ok := tfMap["tags"].(map[string]interface{}); ok {
		tags = Tags(tftags.New(v))
	}

	apiObject := &s3.IntelligentTieringFilter{}

	if prefix == "" {
		switch len(tags) {
		case 0:
			return nil
		case 1:
			apiObject.Tag = tags[0]
		default:
			apiObject.And = &s3.IntelligentTieringAndOperator{
				Tags: tags,
			}
		}
	} else {
		switch len(tags) {
		case 0:
			apiObject.Prefix = aws.String(prefix)
		default:
			apiObject.And = &s3.IntelligentTieringAndOperator{
				Prefix: aws.String(prefix),
				Tags:   tags,
			}
		}
	}

	return apiObject
}

func expandTiering(tfMap map[string]interface{}) *s3.Tiering {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3.Tiering{}

	if v, ok := tfMap["access_tier"].(string); ok && v != "" {
		apiObject.AccessTier = aws.String(v)
	}

	if v, ok := tfMap["days"].(int); ok && v != 0 {
		apiObject.Days = aws.Int64(int64(v))
	}

	return apiObject
}

func expandTierings(tfList []interface{}) []*s3.Tiering {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*s3.Tiering

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTiering(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenIntelligentTieringFilter(apiObject *s3.IntelligentTieringFilter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.And == nil {
		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.Tag; v != nil {
			tfMap["tags"] = KeyValueTags([]*s3.Tag{v}).Map()
		}
	} else {
		apiObject := apiObject.And

		if v := apiObject.Prefix; v != nil {
			tfMap["prefix"] = aws.StringValue(v)
		}

		if v := apiObject.Tags; v != nil {
			tfMap["tags"] = KeyValueTags(v).Map()
		}
	}

	return tfMap
}

func flattenTiering(apiObject *s3.Tiering) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccessTier; v != nil {
		tfMap["access_tier"] = aws.StringValue(v)
	}

	if v := apiObject.Days; v != nil {
		tfMap["days"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenTierings(apiObjects []*s3.Tiering) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTiering(apiObject))
	}

	return tfList
}
