package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceThesaurus() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThesaurusCreate,
		ReadWithoutTimeout:   resourceThesaurusRead,
		UpdateWithoutTimeout: resourceThesaurusUpdate,
		DeleteWithoutTimeout: resourceThesaurusDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source_s3_path": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"thesaurus_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceThesaurusCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &kendra.CreateThesaurusInput{
		ClientToken:  aws.String(resource.UniqueId()),
		IndexId:      aws.String(d.Get("index_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		SourceS3Path: expandSourceS3Path(d.Get("source_s3_path").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateThesaurus(ctx, input)
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
		return diag.Errorf("creating Amazon Kendra Thesaurus (%s): %s", d.Get("name").(string), err)
	}

	if outputRaw == nil {
		return diag.Errorf("creating Amazon Kendra Thesaurus (%s): empty output", d.Get("name").(string))
	}

	output := outputRaw.(*kendra.CreateThesaurusOutput)

	id := aws.ToString(output.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitThesaurusCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Kendra Thesaurus (%s) create: %s", d.Id(), err)
	}

	return resourceThesaurusRead(ctx, d, meta)
}

func resourceThesaurusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id, indexId, err := ThesaurusParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	out, err := FindThesaurusByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Thesaurus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Kendra Thesaurus (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "kendra",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/thesaurus/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("description", out.Description)
	d.Set("index_id", out.IndexId)
	d.Set("name", out.Name)
	d.Set("role_arn", out.RoleArn)
	d.Set("status", out.Status)
	d.Set("thesaurus_id", out.Id)

	if err := d.Set("source_s3_path", flattenSourceS3Path(out.SourceS3Path)); err != nil {
		return diag.Errorf("setting complex argument: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)
	if err != nil {
		return diag.Errorf("listing tags for Kendra Thesaurus (%s): %s", d.Id(), err)
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

func resourceThesaurusUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	if d.HasChangesExcept("tags", "tags_all") {
		id, indexId, err := ThesaurusParseResourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		input := &kendra.UpdateThesaurusInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("source_s3_path") {
			input.SourceS3Path = expandSourceS3Path(d.Get("source_s3_path").([]interface{}))
		}

		log.Printf("[DEBUG] Updating Kendra Thesaurus (%s): %#v", d.Id(), input)

		_, err = tfresource.RetryWhen(
			propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateThesaurus(ctx, input)
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
			return diag.Errorf("updating Kendra Thesaurus(%s): %s", d.Id(), err)
		}

		if _, err := waitThesaurusUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for Kendra Thesaurus (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating Kendra Thesaurus (%s) tags: %s", d.Id(), err))
		}
	}

	return resourceThesaurusRead(ctx, d, meta)
}

func resourceThesaurusDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	log.Printf("[INFO] Deleting Kendra Thesaurus %s", d.Id())

	id, indexId, err := ThesaurusParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = conn.DeleteThesaurus(ctx, &kendra.DeleteThesaurusInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra Thesaurus (%s): %s", d.Id(), err)
	}

	if _, err := waitThesaurusDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra Thesaurus (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func statusThesaurus(ctx context.Context, conn *kendra.Client, id, indexId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindThesaurusByID(ctx, conn, id, indexId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func waitThesaurusCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeThesaurusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{string(types.ThesaurusStatusCreating)},
		Target:                    []string{string(types.ThesaurusStatusActive)},
		Refresh:                   statusThesaurus(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeThesaurusOutput); ok {
		if out.Status == types.ThesaurusStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}

func waitThesaurusUpdated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeThesaurusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{string(types.QuerySuggestionsBlockListStatusUpdating)},
		Target:                    []string{string(types.QuerySuggestionsBlockListStatusActive)},
		Refresh:                   statusThesaurus(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeThesaurusOutput); ok {
		if out.Status == types.ThesaurusStatusActiveButUpdateFailed || out.Status == types.ThesaurusStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}

func waitThesaurusDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeThesaurusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.QuerySuggestionsBlockListStatusDeleting)},
		Target:  []string{},
		Refresh: statusThesaurus(ctx, conn, id, indexId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeThesaurusOutput); ok {
		if out.Status == types.ThesaurusStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
		return out, err
	}

	return nil, err
}
