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

func ResourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDataSourceCreate,
		ReadContext:   resourceDataSourceRead,
		UpdateContext: resourceDataSourceUpdate,
		DeleteContext: resourceDataSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
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
			"data_source_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
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
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dataSourceTypeValues(types.DataSourceType("").Values()...), false),
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

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &kendra.CreateDataSourceInput{
		ClientToken: aws.String(resource.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(name),
		Type:        types.DataSourceType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("language_code"); ok {
		input.LanguageCode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Kendra Data Source %#v", input)

	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateDataSource(ctx, input)
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
		return diag.Errorf("creating Kendra Data Source (%s): %s", name, err)
	}

	if outputRaw == nil {
		return diag.Errorf("creating Kendra Data Source (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateDataSourceOutput)

	id := aws.ToString(output.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitDataSourceCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Data Source (%s) creation: %s", d.Id(), err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id, indexId, err := DataSourceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := FindDataSourceByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("getting Kendra Data Source (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/data-source/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("data_source_id", resp.Id)
	d.Set("description", resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("index_id", resp.IndexId)
	d.Set("language_code", resp.LanguageCode)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("schedule", resp.Schedule)
	d.Set("status", resp.Status)
	d.Set("type", resp.Type)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

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

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	if d.HasChanges("description", "language_code", "role_arn", "schedule") {
		id, indexId, err := DataSourceParseResourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		input := &kendra.UpdateDataSourceInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("language_code") {
			input.LanguageCode = aws.String(d.Get("language_code").(string))
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("schedule") {
			input.Schedule = aws.String(d.Get("schedule").(string))
		}

		log.Printf("[DEBUG] Updating Kendra Data Source (%s): %#v", d.Id(), input)

		_, err = tfresource.RetryWhen(
			propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateDataSource(ctx, input)
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
			return diag.Errorf("updating Kendra Data Source(%s): %s", d.Id(), err)
		}

		if _, err := waitDataSourceUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Kendra Data Source (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("updating Kendra Data Source (%s) tags: %s", d.Id(), err))
		}
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	log.Printf("[INFO] Deleting Kendra Data Source %s", d.Id())

	id, indexId, err := DataSourceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteDataSource(ctx, &kendra.DeleteDataSourceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra Data Source (%s): %s", d.Id(), err)
	}

	if _, err := waitDataSourceDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra Data Source (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func waitDataSourceCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {

	stateConf := &resource.StateChangeConf{
		Pending:                   dataSourceStatusValues(types.DataSourceStatusCreating),
		Target:                    dataSourceStatusValues(types.DataSourceStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusDataSource(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		if output.Status == types.DataSourceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitDataSourceUpdated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {

	stateConf := &resource.StateChangeConf{
		Pending:                   dataSourceStatusValues(types.DataSourceStatusUpdating),
		Target:                    dataSourceStatusValues(types.DataSourceStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusDataSource(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		if output.Status == types.DataSourceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitDataSourceDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: dataSourceStatusValues(types.DataSourceStatusDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusDataSource(ctx, conn, id, indexId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func statusDataSource(ctx context.Context, conn *kendra.Client, id, indexId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDataSourceByID(ctx, conn, id, indexId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

// Helpers added. Could be generated or somehow use go 1.18 generics?
func dataSourceTypeValues(input ...types.DataSourceType) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func dataSourceStatusValues(input ...types.DataSourceStatus) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}
