// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"log"
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

// @SDKResource("aws_qbusiness_index", name="Index")
// @Tags(identifierAttribute="arn")
func ResourceIndex() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceIndexCreate,
		ReadWithoutTimeout:   resourceIndexRead,
		UpdateWithoutTimeout: resourceIndexUpdate,
		DeleteWithoutTimeout: resourceIndexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the Amazon Q application associated with the index.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) of the Amazon Q index.",
			},
			"capacity_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"units": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "The number of additional storage units for the Amazon Q index.",
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for the Amazon Q index.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"document_attribute_configurations": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 50,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the document attribute.",
										ValidateFunc: validation.All(
											validation.StringLenBetween(0, 2048),
											validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
										),
									},
									"search": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Information about whether the document attribute can be used by an end user to search for information on their web experience.",
										ValidateFunc: validation.StringInSlice(qbusiness.Status_Values(), false),
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "The type of document attribute.",
										ValidateFunc: validation.StringInSlice(qbusiness.AttributeType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"index_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the Amazon Q index.",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	application_id := d.Get("application_id").(string)
	display_name := d.Get("display_name").(string)

	input := &qbusiness.CreateIndexInput{
		ApplicationId: aws.String(application_id),
		DisplayName:   aws.String(display_name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("capacity_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CapacityConfiguration = &qbusiness.IndexCapacityConfiguration{
			Units: aws.Int64(int64(v.([]interface{})[0].(map[string]interface{})["units"].(int))),
		}
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIndexWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness index: %s", err)
	}

	d.SetId(application_id + "/" + aws.StringValue(output.IndexId))

	if _, err := waitIndexCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness index (%s) to be created: %s", d.Id(), err)
	}

	updateInput := &qbusiness.UpdateIndexInput{
		ApplicationId: aws.String(application_id),
		IndexId:       output.IndexId,
	}

	updateInput.DocumentAttributeConfigurations = expandDocumentAttributeConfigurations(d.Get("document_attribute_configurations").([]interface{}))

	_, err = conn.UpdateIndexWithContext(ctx, updateInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness index (%s): %s", d.Id(), err)
	}
	return append(diags, resourceIndexRead(ctx, d, meta)...)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	output, err := FindIndexByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] qbusiness index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness index (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.IndexArn)
	d.Set("description", output.Description)
	d.Set("display_name", output.DisplayName)
	d.Set("index_id", output.IndexId)
	d.Set("application_id", output.ApplicationId)

	if err := d.Set("capacity_configuration", flattenIndexCapacityConfiguration(output.CapacityConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting qbusiness index capacity_configuration: %s", err)
	}

	omitDefaults := filterDefaultDocumentAttributeConfigurations(output.DocumentAttributeConfigurations)
	if err := d.Set("document_attribute_configurations", flattenDocumentAttributeConfigurations(omitDefaults)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting qbusiness index document_attribute_configurations: %s", err)
	}

	return diags
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	id := strings.Split(d.Id(), "/")
	input := &qbusiness.UpdateIndexInput{
		ApplicationId: aws.String(id[0]),
		IndexId:       aws.String(id[1]),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
	}

	if d.HasChange("capacity_configuration") {
		input.CapacityConfiguration = expandIndexCapacityConfiguration(d.Get("capacity_configuration").([]interface{}))
	}

	_, err := conn.UpdateIndexWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness index (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	id := strings.Split(d.Id(), "/")
	_, err := conn.DeleteIndexWithContext(ctx, &qbusiness.DeleteIndexInput{
		ApplicationId: aws.String(id[0]),
		IndexId:       aws.String(id[1]),
	})

	if tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness index (%s): %s", d.Id(), err)
	}

	if _, err := waitIndexDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness index (%s) to be deleted: %s", d.Id(), err)
	}
	return diags
}

func FindIndexByID(ctx context.Context, conn *qbusiness.QBusiness, index_id string) (*qbusiness.GetIndexOutput, error) {

	id := strings.Split(index_id, "/")
	input := &qbusiness.GetIndexInput{
		ApplicationId: aws.String(id[0]),
		IndexId:       aws.String(id[1]),
	}

	output, err := conn.GetIndexWithContext(ctx, input)

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

func expandIndexCapacityConfiguration(v []interface{}) *qbusiness.IndexCapacityConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	return &qbusiness.IndexCapacityConfiguration{
		Units: aws.Int64(int64(v[0].(map[string]interface{})["units"].(int))),
	}
}

func flattenIndexCapacityConfiguration(v *qbusiness.IndexCapacityConfiguration) []interface{} {
	if v == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"units": aws.Int64Value(v.Units),
		},
	}
}

func flattenDocumentAttributeConfigurations(v []*qbusiness.DocumentAttributeConfiguration) []interface{} {
	if v == nil {
		return nil
	}
	var attributes []interface{}
	for _, attribute := range v {
		attributes = append(attributes, map[string]interface{}{
			"name":   aws.StringValue(attribute.Name),
			"search": aws.StringValue(attribute.Search),
			"type":   aws.StringValue(attribute.Type),
		})
	}
	return []interface{}{
		map[string]interface{}{
			"attribute": attributes,
		},
	}
}

func expandDocumentAttributeConfigurations(v []interface{}) []*qbusiness.DocumentAttributeConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})["attribute"].([]interface{})

	var attributes []*qbusiness.DocumentAttributeConfiguration
	for _, attribute := range m {
		attributes = append(attributes, &qbusiness.DocumentAttributeConfiguration{
			Name:   aws.String(attribute.(map[string]interface{})["name"].(string)),
			Search: aws.String(attribute.(map[string]interface{})["search"].(string)),
			Type:   aws.String(attribute.(map[string]interface{})["type"].(string)),
		})
	}
	return attributes
}

func filterDefaultDocumentAttributeConfigurations(conf []*qbusiness.DocumentAttributeConfiguration) []*qbusiness.DocumentAttributeConfiguration {
	var attributes []*qbusiness.DocumentAttributeConfiguration
	for _, attribute := range conf {
		filter := false
		if strings.HasPrefix(aws.StringValue(attribute.Name), "_") {
			filter = true
		}
		if aws.StringValue(attribute.Name) == "_document_title" && aws.StringValue(attribute.Search) == qbusiness.StatusDisabled {
			filter = false
		}
		if filter {
			continue
		}
		attributes = append(attributes, attribute)
	}
	return attributes
}
