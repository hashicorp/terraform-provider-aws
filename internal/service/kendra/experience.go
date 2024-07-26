// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kendra_experience")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
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
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`), ""),
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
											ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`), ""),
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
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`), ""),
									},
								},
							},
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			names.AttrEndpoints: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEndpoint: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEndpointType: {
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
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]*`), ""),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`), ""),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange(names.AttrDescription, func(_ context.Context, old, new, meta interface{}) bool {
				// Any existing value cannot be cleared.
				return new.(string) == ""
			}),
		),
	}
}

func resourceExperienceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	in := &kendra.CreateExperienceInput{
		ClientToken: aws.String(id.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(d.Get(names.AttrName).(string)),
		RoleArn:     aws.String(d.Get(names.AttrRoleARN).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Configuration = expandConfiguration(v.([]interface{}))
	}

	out, err := conn.CreateExperience(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amazon Kendra Experience (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "creating Amazon Kendra Experience (%s): empty output", d.Get(names.AttrName).(string))
	}

	id := aws.ToString(out.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if err := waitExperienceCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Amazon Kendra Experience (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceExperienceRead(ctx, d, meta)...)
}

func resourceExperienceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := FindExperienceByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Experience (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kendra Experience (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "kendra",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/experience/%s", indexId, id),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set("index_id", out.IndexId)
	d.Set(names.AttrDescription, out.Description)
	d.Set("experience_id", out.Id)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrRoleARN, out.RoleArn)
	d.Set(names.AttrStatus, out.Status)

	if err := d.Set(names.AttrEndpoints, flattenEndpoints(out.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints argument: %s", err)
	}

	if err := d.Set(names.AttrConfiguration, flattenConfiguration(out.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration argument: %s", err)
	}

	return diags
}

func resourceExperienceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	in := &kendra.UpdateExperienceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	}

	if d.HasChange(names.AttrConfiguration) {
		in.Configuration = expandConfiguration(d.Get(names.AttrConfiguration).([]interface{}))
	}

	if d.HasChange(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrName) {
		in.Name = aws.String(d.Get(names.AttrName).(string))
	}

	if d.HasChange(names.AttrRoleARN) {
		in.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
	}

	log.Printf("[DEBUG] Updating Kendra Experience (%s): %#v", d.Id(), in)
	_, err = conn.UpdateExperience(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Kendra Experience (%s): %s", d.Id(), err)
	}

	if err := waitExperienceUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kendra Experience (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceExperienceRead(ctx, d, meta)...)
}

func resourceExperienceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	log.Printf("[INFO] Deleting Kendra Experience %s", d.Id())

	id, indexId, err := ExperienceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	_, err = conn.DeleteExperience(ctx, &kendra.DeleteExperienceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kendra Experience (%s): %s", d.Id(), err)
	}

	if err := waitExperienceDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kendra Experience (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func waitExperienceCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ExperienceStatusCreating),
		Target:                    enum.Slice(types.ExperienceStatusActive),
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
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(types.ExperienceStatusActive),
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
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ExperienceStatusDeleting),
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

func statusExperience(ctx context.Context, conn *kendra.Client, id, indexId string) retry.StateRefreshFunc {
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
		return nil, &retry.NotFoundError{
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
			m[names.AttrEndpoint] = aws.ToString(v)
		}

		if v := string(apiObject.EndpointType); v != "" {
			m[names.AttrEndpointType] = v
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
