// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qbusiness_retriever", name="Retriever")
// @Tags(identifierAttribute="arn")
func ResourceRetriever() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRetrieverCreate,
		ReadWithoutTimeout:   resourceRetrieverRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Identifier of Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ARN of the the retriever.",
			},
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kendra_index_configuration": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Description:   "Information on how the Amazon Kendra index used as a retriever for your Amazon Q application is configured.",
							ConflictsWith: []string{"native_index_configuration"},
							AtLeastOneOf:  []string{"kendra_index_configuration", "native_index_configuration"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Identifier of the Amazon Kendra index.",
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
										),
									},
								},
							},
						},
						"native_index_configuration": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Description:   "Information on how a Amazon Q index used as a retriever for your Amazon Q application is configured.",
							ConflictsWith: []string{"kendra_index_configuration"},
							AtLeastOneOf:  []string{"kendra_index_configuration", "native_index_configuration"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Identifier for the Amazon Q index.",
										ValidateFunc: validation.All(
											validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
										),
									},
								},
							},
						},
					},
				},
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of retriever.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"iam_service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
				ValidateFunc: verify.ValidARN,
			},
			"retriever_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the retriever.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Type of retriever.",
				ValidateFunc: validation.StringInSlice(qbusiness.RetrieverType_Values(), false),
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRetrieverCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	application_id := d.Get("application_id").(string)
	input := &qbusiness.CreateRetrieverInput{
		ApplicationId: aws.String(application_id),
		Configuration: expandRetrieverConfiguration(d.Get("configuration").([]interface{})),
		DisplayName:   aws.String(d.Get("display_name").(string)),
		Type:          aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateRetrieverWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness retriever: %s", err)
	}

	d.SetId(application_id + "/" + aws.StringValue(output.RetrieverId))

	if _, err := waitRetrieverCreated(ctx, conn, aws.StringValue(output.RetrieverId), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness retriever (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceIndexRead(ctx, d, meta)...)
}

func resourceRetrieverUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	application_id, retriever_id, err := parseRetrieverID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid qbusiness retriever ID: %s", err)
	}

	changed := false
	input := &qbusiness.UpdateRetrieverInput{
		ApplicationId: aws.String(application_id),
		RetrieverId:   aws.String(retriever_id),
	}

	if d.HasChange("iam_service_role_arn") {
		input.RoleArn = aws.String(d.Get("iam_service_role_arn").(string))
		changed = true
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
		changed = true
	}

	if d.HasChange("configuration") {
		input.Configuration = expandRetrieverConfiguration(d.Get("configuration").([]interface{}))
		changed = true
	}

	if changed {
		if _, err := conn.UpdateRetrieverWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating qbusiness retriever: %s", err)
		}
	}

	return append(diags, resourceRetrieverRead(ctx, d, meta)...)
}

func resourceRetrieverRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	output, err := FindRetrieverByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness retriever: %s", err)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("arn", output.RetrieverArn)
	d.Set("configuration", flattenRetrieverConfiguration(output.Configuration))
	d.Set("display_name", output.DisplayName)
	d.Set("iam_service_role_arn", output.RoleArn)
	d.Set("retriever_id", output.RetrieverId)
	d.Set("type", output.Type)

	return diags
}

func resourceRetrieverDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	application_id, retriever_id, err := parseRetrieverID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid qbusiness retriever ID: %s", err)
	}

	input := &qbusiness.DeleteRetrieverInput{
		ApplicationId: aws.String(application_id),
		RetrieverId:   aws.String(retriever_id),
	}

	if _, err := conn.DeleteRetrieverWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness retriever: %s", err)
	}

	if _, err := waitRetrieverDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness retriever (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func parseRetrieverID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid retriever ID: %s", id)
	}

	return parts[0], parts[1], nil
}

func FindRetrieverByID(ctx context.Context, conn *qbusiness.QBusiness, id string) (*qbusiness.GetRetrieverOutput, error) {
	application_id, retriever_id, err := parseRetrieverID(id)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetRetrieverInput{
		ApplicationId: aws.String(application_id),
		RetrieverId:   aws.String(retriever_id),
	}

	output, err := conn.GetRetrieverWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
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

func expandRetrieverConfiguration(v []interface{}) *qbusiness.RetrieverConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	m := v[0].(map[string]interface{})

	return &qbusiness.RetrieverConfiguration{
		KendraIndexConfiguration: expandKendraIndexConfiguration(m["kendra_index_configuration"].([]interface{})),
		NativeIndexConfiguration: expandNativeIndexConfiguration(m["native_index_configuration"].([]interface{})),
	}
}

func expandKendraIndexConfiguration(v []interface{}) *qbusiness.KendraIndexConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	m := v[0].(map[string]interface{})

	return &qbusiness.KendraIndexConfiguration{
		IndexId: aws.String(m["index_id"].(string)),
	}
}

func expandNativeIndexConfiguration(v []interface{}) *qbusiness.NativeIndexConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	m := v[0].(map[string]interface{})

	return &qbusiness.NativeIndexConfiguration{
		IndexId: aws.String(m["index_id"].(string)),
	}
}

func flattenRetrieverConfiguration(c *qbusiness.RetrieverConfiguration) []interface{} {
	if c == nil {
		return nil
	}

	m := map[string]interface{}{}

	if c.KendraIndexConfiguration != nil {
		m["kendra_index_configuration"] = []interface{}{flattenKendraIndexConfiguration(c.KendraIndexConfiguration)}
	}

	if c.NativeIndexConfiguration != nil {
		m["native_index_configuration"] = []interface{}{flattenNativeIndexConfiguration(c.NativeIndexConfiguration)}
	}

	return []interface{}{m}
}

func flattenKendraIndexConfiguration(c *qbusiness.KendraIndexConfiguration) map[string]interface{} {
	if c == nil {
		return nil
	}

	m := map[string]interface{}{}

	if c.IndexId != nil {
		m["index_id"] = aws.StringValue(c.IndexId)
	}

	return m
}

func flattenNativeIndexConfiguration(c *qbusiness.NativeIndexConfiguration) map[string]interface{} {
	if c == nil {
		return nil
	}

	m := map[string]interface{}{}

	if c.IndexId != nil {
		m["index_id"] = aws.StringValue(c.IndexId)
	}

	return m
}
