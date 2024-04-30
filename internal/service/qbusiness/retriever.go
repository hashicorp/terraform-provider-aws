// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkchema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_qbusiness_retriever", name="Retriever")
// @Tags(identifierAttribute="arn")
func newResourceRetriever(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRetriever{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRetriever = "Retriever"
)

type resourceRetriever struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceRetriever) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_retriever"
}

func (r *resourceRetriever) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"application_id": schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the retriever",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"retriever_id": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the Amazon Q retriever.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of an IAM role used by Amazon Q to access the basic authentication credentials stored in a Secrets Manager secret.",
				Optional:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"kendra_index_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.Expressions{path.MatchRoot("native_index_configuration")}...),
					listvalidator.AtLeastOneOf(path.MatchRoot("kendra_index_configuration"), path.MatchRoot("native_index_configuration")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index_id": schema.StringAttribute{
							Required:    true,
							Description: "Identifier of the Amazon Kendra index.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
							},
						},
					},
				},
			},
			"native_index_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.ConflictsWith(path.Expressions{path.MatchRoot("kendra_index_configuration")}...),
					listvalidator.AtLeastOneOf(path.MatchRoot("kendra_index_configuration"), path.MatchRoot("native_index_configuration")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"index_id": schema.StringAttribute{
							Required:    true,
							Description: "Identifier for the Amazon Q index.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceRetriever) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *resourceRetriever) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *resourceRetriever) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceRetriever) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceRetriever) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

/*
// @SDKResource("aws_qbusiness_retriever", name="Retriever")
// @Tags(identifierAttribute="arn")
func ResourceRetriever() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRetrieverCreate,
		ReadWithoutTimeout:   resourceRetrieverRead,
		UpdateWithoutTimeout: resourceRetrieverUpdate,
		DeleteWithoutTimeout: resourceRetrieverDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

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

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}
*/

func resourceRetrieverCreate(ctx context.Context, d *sdkchema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id := d.Get("application_id").(string)
	input := &qbusiness.CreateRetrieverInput{
		ApplicationId: aws.String(application_id),
		DisplayName:   aws.String(d.Get("display_name").(string)),
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kendra_index_configuration"); ok {
		input.Configuration = types.RetrieverConfiguration(expandKendraIndexConfiguration(v.([]interface{})))
		input.Type = types.RetrieverTypeKendraIndex
	}

	if v, ok := d.GetOk("native_index_configuration"); ok {
		input.Configuration = types.RetrieverConfiguration(expandNativeIndexConfiguration(v.([]interface{})))
		input.Type = types.RetrieverTypeNativeIndex
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateRetriever(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness retriever: %s", err)
	}

	d.SetId(application_id + "/" + aws.ToString(output.RetrieverId))

	if _, err := waitRetrieverCreated(ctx, conn, d.Id(), d.Timeout(sdkchema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness retriever (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceRetrieverRead(ctx, d, meta)...)
}

func resourceRetrieverUpdate(ctx context.Context, d *sdkchema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

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

	if d.HasChange("kendra_index_configuration") {
		input.Configuration = types.RetrieverConfiguration(expandKendraIndexConfiguration(d.Get("kendra_index_configuration").([]interface{})))
		changed = true
	}

	if d.HasChange("native_index_configuration") {
		input.Configuration = types.RetrieverConfiguration(expandNativeIndexConfiguration(d.Get("native_index_configuration").([]interface{})))
		changed = true
	}

	if changed {
		if _, err := conn.UpdateRetriever(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating qbusiness retriever: %s", err)
		}
	}

	return append(diags, resourceRetrieverRead(ctx, d, meta)...)
}

func resourceRetrieverRead(ctx context.Context, d *sdkchema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

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
	if output.Type == types.RetrieverTypeKendraIndex {
		d.Set("kendra_index_configuration", flattenKendraIndexConfiguration(output.Configuration.(*types.RetrieverConfigurationMemberKendraIndexConfiguration)))
	}
	if output.Type == types.RetrieverTypeNativeIndex {
		d.Set("native_index_configuration", flattenNativeIndexConfiguration(output.Configuration.(*types.RetrieverConfigurationMemberNativeIndexConfiguration)))
	}
	d.Set("display_name", output.DisplayName)
	d.Set("iam_service_role_arn", output.RoleArn)
	d.Set("retriever_id", output.RetrieverId)

	return diags
}

func resourceRetrieverDelete(ctx context.Context, d *sdkchema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, retriever_id, err := parseRetrieverID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid qbusiness retriever ID: %s", err)
	}

	input := &qbusiness.DeleteRetrieverInput{
		ApplicationId: aws.String(application_id),
		RetrieverId:   aws.String(retriever_id),
	}

	_, err = conn.DeleteRetriever(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness retriever: %s", err)
	}

	if _, err := waitRetrieverDeleted(ctx, conn, d.Id(), d.Timeout(sdkchema.TimeoutDelete)); err != nil {
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

func FindRetrieverByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetRetrieverOutput, error) {
	application_id, retriever_id, err := parseRetrieverID(id)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetRetrieverInput{
		ApplicationId: aws.String(application_id),
		RetrieverId:   aws.String(retriever_id),
	}

	output, err := conn.GetRetriever(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func expandKendraIndexConfiguration(v []interface{}) *types.RetrieverConfigurationMemberKendraIndexConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	return &types.RetrieverConfigurationMemberKendraIndexConfiguration{
		Value: types.KendraIndexConfiguration{
			IndexId: aws.String(m["index_id"].(string)),
		},
	}
}

func expandNativeIndexConfiguration(v []interface{}) *types.RetrieverConfigurationMemberNativeIndexConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})
	return &types.RetrieverConfigurationMemberNativeIndexConfiguration{
		Value: types.NativeIndexConfiguration{
			IndexId: aws.String(m["index_id"].(string)),
		},
	}
}

func flattenKendraIndexConfiguration(c *types.RetrieverConfigurationMemberKendraIndexConfiguration) []interface{} {
	m := map[string]interface{}{}
	m["index_id"] = aws.ToString(c.Value.IndexId)
	return []interface{}{m}
}

func flattenNativeIndexConfiguration(c *types.RetrieverConfigurationMemberNativeIndexConfiguration) []interface{} {
	m := map[string]interface{}{}
	m["index_id"] = aws.ToString(c.Value.IndexId)
	return []interface{}{m}
}
