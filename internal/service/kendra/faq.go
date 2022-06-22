package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFaq() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFaqCreate,
		ReadContext:   resourceFaqRead,
		UpdateContext: resourceFaqUpdate,
		DeleteContext: resourceFaqDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"faq_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_format": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(faqFileFormatValues(types.FaqFileFormat("").Values()...), false),
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"language_code": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z-]*`),
						"Must have alphanumeric characters or hyphens.",
					),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_path": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceFaqCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &kendra.CreateFaqInput{
		ClientToken: aws.String(resource.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(name),
		RoleArn:     aws.String(d.Get("role_arn").(string)),
		S3Path:      expandS3Path(d.Get("s3_path").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("file_format"); ok {
		input.FileFormat = types.FaqFileFormat(v.(string))
	}

	if v, ok := d.GetOk("language_code"); ok {
		input.LanguageCode = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Kendra Faq %#v", input)

	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateFaq(ctx, input)
		},
		func(err error) (bool, error) {
			var validationException *types.ValidationException

			if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return diag.Errorf("creating Kendra Faq (%s): %s", name, err)
	}

	if outputRaw == nil {
		return diag.Errorf("creating Kendra Faq (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateFaqOutput)

	id := aws.ToString(output.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitFaqCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Faq (%s) creation: %s", d.Id(), err)
	}

	return resourceFaqRead(ctx, d, meta)
}

func resourceFaqRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id, indexId, err := FaqParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := FindFaqByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Faq (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Kendra Faq (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/faq/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("faq_id", resp.Id)
	d.Set("file_format", resp.FileFormat)
	d.Set("index_id", resp.IndexId)
	d.Set("language_code", resp.LanguageCode)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("status", resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set("s3_path", flattenS3Path(resp.S3Path)); err != nil {
		return diag.FromErr(err)
	}

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return diag.Errorf("listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceFaqUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("updating Kendra Faq (%s) tags: %s", d.Id(), err))
		}
	}

	return resourceFaqRead(ctx, d, meta)
}

func resourceFaqDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	log.Printf("[INFO] Deleting Kendra Faq %s", d.Id())

	id, indexId, err := FaqParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteFaq(ctx, &kendra.DeleteFaqInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra Faq (%s): %s", d.Id(), err)
	}

	if _, err := waitFaqDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra Faq (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func waitFaqCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeFaqOutput, error) {

	stateConf := &resource.StateChangeConf{
		Pending:                   FaqStatusValues(types.FaqStatusCreating, "PENDING_CREATION"), // API currently returns PENDING_CREATION instead of CREATING
		Target:                    FaqStatusValues(types.FaqStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusFaq(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeFaqOutput); ok {
		if output.Status == types.FaqStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitFaqDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeFaqOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: FaqStatusValues(types.FaqStatusDeleting, "PENDING_DELETION"), // API currently returns PENDING_DELETION instead of DELETING
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusFaq(ctx, conn, id, indexId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeFaqOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func statusFaq(ctx context.Context, conn *kendra.Client, id, indexId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFaqByID(ctx, conn, id, indexId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func expandS3Path(tfList []interface{}) *types.S3Path {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3Path{}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		result.Key = aws.String(v)
	}

	return result
}

func flattenS3Path(apiObject *types.S3Path) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		m["bucket"] = aws.ToString(v)
	}

	if v := apiObject.Key; v != nil {
		m["key"] = aws.ToString(v)
	}

	return []interface{}{m}
}

// Helpers added. Could be generated or somehow use go 1.18 generics?
func faqFileFormatValues(input ...types.FaqFileFormat) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func FaqStatusValues(input ...types.FaqStatus) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	output = append(output, "PENDING_CREATION")
	output = append(output, "PENDING_DELETION")

	return output
}
