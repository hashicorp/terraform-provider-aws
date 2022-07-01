package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceExperience() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceExperienceCreate,
		ReadWithoutTimeout:   resourceExperienceRead,
		UpdateWithoutTimeout: resourceExperienceUpdate,
		DeleteWithoutTimeout: resourceExperienceDelete,

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
			"configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_source_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							AtLeastOneOf: []string{
								"configuration.0.content_source_configuration",
								"configuration.0.user_identity_configuration",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_\-]*`), ""),
										},
									},
									"direct_put_content": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"faq_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_\-]*`), ""),
										},
									},
								},
							},
						},
						"user_identity_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							AtLeastOneOf: []string{
								"configuration.0.user_identity_configuration",
								"configuration.0.content_source_configuration",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"identity_attribute_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_\-]*`), ""),
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"endpoints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"experience_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9-]*`), ""),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_\-]*`), ""),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("description", func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
		),
	}
}

func resourceExperienceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	in := &kendra.CreateExperienceInput{
		ClientToken: aws.String(resource.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		RoleArn:     aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Configuration = expandConfiguration(v.([]interface{}))
	}

	out, err := conn.CreateExperience(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Kendra Experience (%s): %s", d.Get("name").(string), err)
	}

	if out == nil {
		return diag.Errorf("creating Amazon Kendra Experience (%s): empty output", d.Get("name").(string))
	}

	id := aws.ToString(out.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if err := waitExperienceCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Kendra Experience (%s) create: %s", d.Id(), err)
	}

	return resourceExperienceRead(ctx, d, meta)
}

func resourceExperienceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	out, err := FindExperienceByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Experience (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Kendra Experience (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "kendra",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/experience/%s", indexId, id),
	}.String()

	d.Set("arn", arn)
	d.Set("index_id", out.IndexId)
	d.Set("description", out.Description)
	d.Set("experience_id", out.Id)
	d.Set("name", out.Name)
	d.Set("role_arn", out.RoleArn)
	d.Set("status", out.Status)

	if err := d.Set("endpoints", flattenEndpoints(out.Endpoints)); err != nil {
		return diag.Errorf("setting endpoints argument: %s", err)
	}

	if err := d.Set("configuration", flattenConfiguration(out.Configuration)); err != nil {
		return diag.Errorf("setting configuration argument: %s", err)
	}

	return nil
}

func resourceExperienceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	in := &kendra.UpdateExperienceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	if d.HasChange("configuration") {
		in.Configuration = expandConfiguration(d.Get("configuration").([]interface{}))
	}

	if d.HasChange("description") {
		in.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		in.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("role_arn") {
		in.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	log.Printf("[DEBUG] Updating Kendra Experience (%s): %#v", d.Id(), in)
	_, err = conn.UpdateExperience(ctx, in)
	if err != nil {
		return diag.Errorf("updating Kendra Experience (%s): %s", d.Id(), err)
	}

	if err := waitExperienceUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for Kendra Experience (%s) update: %s", d.Id(), err)
	}

	return resourceExperienceRead(ctx, d, meta)
}

func resourceExperienceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	log.Printf("[INFO] Deleting Kendra Experience %s", d.Id())

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = conn.DeleteExperience(ctx, &kendra.DeleteExperienceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Kendra Experience (%s): %s", d.Id(), err)
	}

	if err := waitExperienceDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Kendra Experience (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func waitExperienceCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{string(types.ExperienceStatusCreating)},
		Target:                    []string{string(types.ExperienceStatusActive)},
		Refresh:                   statusExperience(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeExperienceOutput); ok {
		if out.Status == types.ExperienceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
	}

	return err
}

func waitExperienceUpdated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{string(types.ExperienceStatusActive)},
		Refresh:                   statusExperience(ctx, conn, id, indexId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeExperienceOutput); ok {
		if out.Status == types.ExperienceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
	}

	return err
}

func waitExperienceDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{string(types.ExperienceStatusDeleting)},
		Target:  []string{},
		Refresh: statusExperience(ctx, conn, id, indexId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kendra.DescribeExperienceOutput); ok {
		if out.Status == types.ExperienceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(out.ErrorMessage)))
		}
	}

	return err
}

func statusExperience(ctx context.Context, conn *kendra.Client, id, indexId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindExperienceByID(ctx, conn, id, indexId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindExperienceByID(ctx context.Context, conn *kendra.Client, id, indexId string) (*kendra.DescribeExperienceOutput, error) {
	in := &kendra.DescribeExperienceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	out, err := conn.DescribeExperience(ctx, in)
	var resourceNotFoundException *types.ResourceNotFoundException

	if errors.As(err, &resourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenConfiguration(apiObject *types.ExperienceConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.ContentSourceConfiguration; v != nil {
		m["content_source_configuration"] = flattenContentSourceConfiguration(v)
	}

	if v := apiObject.UserIdentityConfiguration; v != nil {
		m["user_identity_configuration"] = flattenUserIdentityConfiguration(v)
	}

	return []interface{}{m}
}

func flattenContentSourceConfiguration(apiObject *types.ContentSourceConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"direct_put_content": apiObject.DirectPutContent,
	}

	if v := apiObject.DataSourceIds; len(v) > 0 {
		m["data_source_ids"] = v
	}

	if v := apiObject.FaqIds; len(v) > 0 {
		m["faq_ids"] = v
	}
	return []interface{}{m}
}

func flattenEndpoints(apiObjects []types.ExperienceEndpoint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	l := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		m := make(map[string]interface{})

		if v := apiObject.Endpoint; v != nil {
			m["endpoint"] = aws.ToString(v)
		}

		if v := string(apiObject.EndpointType); v != "" {
			m["endpoint_type"] = v
		}

		l = append(l, m)
	}

	return l

}

func flattenUserIdentityConfiguration(apiObject *types.UserIdentityConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := make(map[string]interface{})

	if v := apiObject.IdentityAttributeName; v != nil {
		m["identity_attribute_name"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func expandConfiguration(tfList []interface{}) *types.ExperienceConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ExperienceConfiguration{}

	if v, ok := tfMap["content_source_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ContentSourceConfiguration = expandContentSourceConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["user_identity_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.UserIdentityConfiguration = expandUserIdentityConfiguration(v[0].(map[string]interface{}))
	}

	return result
}

func expandContentSourceConfiguration(tfMap map[string]interface{}) *types.ContentSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	result := &types.ContentSourceConfiguration{}

	if v, ok := tfMap["data_source_ids"].(*schema.Set); ok && v.Len() > 0 {
		result.DataSourceIds = expandStringList(v.List())
	}

	if v, ok := tfMap["direct_put_content"].(bool); ok {
		result.DirectPutContent = v
	}

	if v, ok := tfMap["faq_ids"].(*schema.Set); ok && v.Len() > 0 {
		result.FaqIds = expandStringList(v.List())
	}

	return result
}

func expandUserIdentityConfiguration(tfMap map[string]interface{}) *types.UserIdentityConfiguration {
	if tfMap == nil {
		return nil
	}

	result := &types.UserIdentityConfiguration{}

	if v, ok := tfMap["identity_attribute_name"].(string); ok && v != "" {
		result.IdentityAttributeName = aws.String(v)
	}

	return result
}

func expandStringList(tfList []interface{}) []string {
	var result []string

	for _, rawVal := range tfList {
		if v, ok := rawVal.(string); ok && v != "" {
			result = append(result, v)
		}
	}

	return result
}
