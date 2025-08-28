// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"iter"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_group", name="Log Group")
// @Tags(identifierAttribute="arn")
// @Testing(destroyTakesT=true)
// @Testing(existsTakesT=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.LogGroup")
// @IdentityAttribute("name")
// @Testing(idAttrDuplicates="name")
// @Testing(preIdentityVersion="v6.7.0")
func resourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"log_group_class": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LogGroupClass](),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validLogGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validLogGroupNamePrefix,
			},
			"retention_in_days": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653}),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.HasChange("log_group_class") {
						return false
					}
					if v, ok := d.GetOk("log_group_class"); ok {
						if awstypes.LogGroupClass(v.(string)) == awstypes.LogGroupClassDelivery {
							return true
						}
					}
					return false
				},
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

var _ list.ListResourceWithRawV5Schemas = &logGroupListResource{}

// @SDKListResource("aws_cloudwatch_log_group")
func logGroupResourceAsListResource() list.ListResourceWithConfigure {
	l := logGroupListResource{}
	l.SetResource(resourceGroup())

	return &l
}

type logGroupListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Identity
	framework.WithImportByIdentity
}

type logGroupListResourceModel struct {
	framework.WithRegionModel
}

func (l *logGroupListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"region": listschema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (l *logGroupListResource) RawV5Schemas(ctx context.Context, request list.RawV5SchemaRequest, response *list.RawV5SchemaResponse) {
	response.ProtoV5Schema = l.GetResource().ProtoSchema(ctx)()
	response.ProtoV5IdentitySchema = l.GetResource().ProtoIdentitySchema(ctx)()
}

func (l *logGroupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.LogsClient(ctx)

	var query logGroupListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	interceptors := []listResultInterceptor{
		populateIdentityInterceptor{},
		identityInterceptor{
			attributes: l.IdentitySpec().Attributes,
		},
		//setResourceInterceptor{},
		//tagsInterceptor{
		//	HTags: interceptors.HTags(unique.Make(inttypes.ServicePackageResourceTags{
		//		IdentifierAttribute: names.AttrARN,
		//	})),
		//},
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		for output, err := range listLogGroups(ctx, conn, &cloudwatchlogs.DescribeLogGroupsInput{}, tfslices.PredicateTrue[*awstypes.LogGroup]()) {
			if err != nil {
				result = list.ListResult{
					Diagnostics: fwdiag.Diagnostics{
						fwdiag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			params := interceptorParams{
				c:      awsClient,
				result: &result,
			}

			params.when = Before
			for interceptor := range slices.Values(interceptors) {
				d := interceptor.read(ctx, params)
				result.Diagnostics.Append(d...)
				if d.HasError() {
					result = list.ListResult{Diagnostics: result.Diagnostics}
					yield(result)
					return
				}
			}

			logGroupResource := l.GetResource()
			rd := logGroupResource.Data(&terraform.InstanceState{})
			rd.SetId(aws.ToString(output.LogGroupName))
			resourceGroupFlatten(ctx, rd, output)

			tfTypeResource, err := rd.TfTypeResourceState()
			if err != nil {
				result = list.ListResult{
					Diagnostics: fwdiag.Diagnostics{
						fwdiag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			if err := result.Resource.Set(ctx, *tfTypeResource); err != nil {
				result = list.ListResult{
					Diagnostics: fwdiag.Diagnostics{
						fwdiag.NewErrorDiagnostic(
							"Error Listing Remote Resources",
							fmt.Sprintf("Error: %s", err),
						),
					},
				}
				yield(result)
				return
			}

			result.DisplayName = fmt.Sprintf("%s: (%s)", aws.ToString(output.LogGroupName), aws.ToString(output.Arn))

			params.when = After
			for interceptor := range tfslices.BackwardValues(interceptors) {
				d := interceptor.read(ctx, params)
				result.Diagnostics.Append(d...)
				if d.HasError() {
					result = list.ListResult{Diagnostics: result.Diagnostics}
					yield(result)
					return
				}
			}

			if result.Diagnostics.HasError() {
				result = list.ListResult{Diagnostics: result.Diagnostics}
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func resourceGroupFlatten(_ context.Context, d *schema.ResourceData, lg awstypes.LogGroup) {
	d.Set(names.AttrARN, trimLogGroupARNWildcardSuffix(aws.ToString(lg.Arn)))
	d.Set(names.AttrKMSKeyID, lg.KmsKeyId)
	d.Set("log_group_class", lg.LogGroupClass)
	d.Set(names.AttrName, lg.LogGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lg.LogGroupName)))
	d.Set("retention_in_days", lg.RetentionInDays)
	// Support in-place update of non-refreshable attribute.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := cloudwatchlogs.CreateLogGroupInput{
		LogGroupClass: awstypes.LogGroupClass(d.Get("log_group_class").(string)),
		LogGroupName:  aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	_, err := conn.CreateLogGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Logs Log Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("retention_in_days"); ok && input.LogGroupClass != awstypes.LogGroupClassDelivery {
		input := cloudwatchlogs.PutRetentionPolicyInput{
			LogGroupName:    aws.String(d.Id()),
			RetentionInDays: aws.Int32(int32(v.(int))),
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate), func(ctx context.Context) (any, error) {
			return conn.PutRetentionPolicy(ctx, &input)
		}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	lg, err := findLogGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Log Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, trimLogGroupARNWildcardSuffix(aws.ToString(lg.Arn)))
	d.Set(names.AttrKMSKeyID, lg.KmsKeyId)
	d.Set("log_group_class", lg.LogGroupClass)
	d.Set(names.AttrName, lg.LogGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lg.LogGroupName)))
	d.Set("retention_in_days", lg.RetentionInDays)
	// Support in-place update of non-refreshable attribute.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	if d.HasChange("retention_in_days") {
		if v, ok := d.GetOk("retention_in_days"); ok {
			input := cloudwatchlogs.PutRetentionPolicyInput{
				LogGroupName:    aws.String(d.Id()),
				RetentionInDays: aws.Int32(int32(v.(int))),
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutUpdate), func(ctx context.Context) (any, error) {
				return conn.PutRetentionPolicy(ctx, &input)
			}, "AccessDeniedException", "no identity-based policy allows the logs:PutRetentionPolicy action")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		} else {
			input := cloudwatchlogs.DeleteRetentionPolicyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DeleteRetentionPolicy(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s) retention policy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrKMSKeyID) {
		if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
			input := cloudwatchlogs.AssociateKmsKeyInput{
				KmsKeyId:     aws.String(v.(string)),
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.AssociateKmsKey(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating CloudWatch Logs Log Group (%s) KMS key: %s", d.Id(), err)
			}
		} else {
			input := cloudwatchlogs.DisassociateKmsKeyInput{
				LogGroupName: aws.String(d.Id()),
			}

			_, err := conn.DisassociateKmsKey(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating CloudWatch Logs Log Group (%s) KMS key: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining CloudWatch Logs Log Group: %s", d.Id())
		return diags
	}

	log.Printf("[INFO] Deleting CloudWatch Logs Log Group: %s", d.Id())
	input := cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(d.Id()),
	}
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.OperationAbortedException](ctx, 1*time.Minute, func(ctx context.Context) (any, error) {
		return conn.DeleteLogGroup(ctx, &input)
	}, "try again")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Log Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findLogGroupByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.LogGroup, error) {
	input := cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(name),
	}

	return findLogGroup(ctx, conn, &input, func(v *awstypes.LogGroup) bool {
		return aws.ToString(v.LogGroupName) == name
	})
}

func findLogGroup(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup]) (*awstypes.LogGroup, error) {
	var output []awstypes.LogGroup
	for value, err := range listLogGroups(ctx, conn, input, filter) {
		if err != nil {
			return nil, err
		}

		output = append(output, value)
	}

	return tfresource.AssertSingleValueResult(output)
}

func listLogGroups(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogGroupsInput, filter tfslices.Predicate[*awstypes.LogGroup]) iter.Seq2[awstypes.LogGroup, error] {
	return func(yield func(awstypes.LogGroup, error) bool) {
		pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.LogGroup{}, fmt.Errorf("listing CloudWatch Logs Log Groups: %w", err))
				return
			}

			for _, v := range page.LogGroups {
				if filter(&v) {
					if !yield(v, nil) {
						return
					}
				}
			}
		}
	}
}

// when represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type when uint16

const (
	Before  when = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

type interceptorParams struct {
	c      *conns.AWSClient
	result *list.ListResult
	//object *jobQueueResourceModel // Because tfsdk.Resource doesn't have SetAttribute
	when when
}

type listResultInterceptor interface {
	read(ctx context.Context, params interceptorParams) fwdiag.Diagnostics
}

type identityInterceptor struct {
	attributes []inttypes.IdentityAttribute
}

func (r identityInterceptor) read(ctx context.Context, params interceptorParams) (diags fwdiag.Diagnostics) {
	awsClient := params.c

	switch params.when {
	case After:
		for _, att := range r.attributes {
			switch att.Name() {
			case names.AttrAccountID:
				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.AccountID(ctx))...)
				if diags.HasError() {
					return
				}

			case names.AttrRegion:
				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), awsClient.Region(ctx))...)
				if diags.HasError() {
					return
				}

			default:
				var attrVal attr.Value
				diags.Append(params.result.Resource.GetAttribute(ctx, path.Root(att.ResourceAttributeName()), &attrVal)...)
				if diags.HasError() {
					return
				}

				diags.Append(params.result.Identity.SetAttribute(ctx, path.Root(att.Name()), attrVal)...)
				if diags.HasError() {
					return
				}
			}
		}
	}

	return
}

// This interceptor will not be needed if Framework pre-populates the Identity as it does with CRUD operations
type populateIdentityInterceptor struct{}

// This interceptor will not be needed if Framework pre-populates the Identity as it does with CRUD operations

func (r populateIdentityInterceptor) read(ctx context.Context, params interceptorParams) (diags fwdiag.Diagnostics) {
	switch params.when {
	case Before:
		identityType := params.result.Identity.Schema.Type()

		obj, d := newEmptyObject(identityType)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		diags.Append(params.result.Identity.Set(ctx, obj)...)
		if diags.HasError() {
			return
		}
	}

	return
}

func newEmptyObject(typ attr.Type) (obj basetypes.ObjectValue, diags fwdiag.Diagnostics) {
	i, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddError(
			"Internal Error",
			"An unexpected error occurred. "+
				"This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected value type to implement attr.TypeWithAttributeTypes, got: %T", typ),
		)
		return
	}

	attrTypes := i.AttributeTypes()
	attrValues := make(map[string]attr.Value, len(attrTypes))
	for attrName := range attrTypes {
		attrValues[attrName] = types.StringNull()
	}
	obj, d := basetypes.NewObjectValue(attrTypes, attrValues)
	diags.Append(d...)
	if d.HasError() {
		return basetypes.ObjectValue{}, diags
	}

	return obj, diags
}
