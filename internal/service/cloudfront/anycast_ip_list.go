// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudfront

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_anycast_ip_list", name="Anycast IP List")
// @Tags(identifierAttribute="arn")
// @Testing(serialize=true)
func newAnycastIPListResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &anycastIPListResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)

	return r, nil
}

type anycastIPListResource struct {
	framework.ResourceWithModel[anycastIPListResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *anycastIPListResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"anycast_ips": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"etag": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"ip_count": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.OneOf(3, 21),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"ipam_config": schema.ListNestedAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipamConfigModel](ctx),
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ipam_cidr_configs": schema.ListNestedAttribute{
							CustomType: fwtypes.NewListNestedObjectTypeOf[ipamCidrConfigStatusModel](ctx),
							Computed:   true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"anycast_ip": schema.StringAttribute{
										Computed: true,
									},
									"cidr": schema.StringAttribute{
										Computed: true,
									},
									"ipam_pool_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Computed:   true,
									},
									names.AttrStatus: schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"quantity": schema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9-_]{1,64}$`), "must contain between 1 and 64 lowercase letters, numbers, hyphens, or underscores"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"ipam_cidr_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipamCidrConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"anycast_ip": schema.StringAttribute{
							Optional: true,
						},
						"cidr": schema.StringAttribute{
							Required: true,
						},
						"ipam_pool_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *anycastIPListResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudfront.CreateAnycastIpListInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	outputCAIL, err := conn.CreateAnycastIpList(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Anycast IP List (%s)", name), err.Error())

		return
	}
	if outputCAIL == nil || outputCAIL.AnycastIpList == nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudFront Anycast IP List (%s)", name), "empty result")
		return
	}

	// Set 'id' for the waiter and so as to taint the resource on error.
	data.ID = fwflex.StringToFramework(ctx, outputCAIL.AnycastIpList.Id)

	outputGAIL, err := waitAnycastIPListDeployed(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudFront Anycast IP List (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns. The configured 'ipam_cidr_config' block has no
	// equivalent on the response (server-assigned values are surfaced via the
	// computed 'ipam_config' attribute), so AutoFlex leaves it untouched.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGAIL.AnycastIpList, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, outputGAIL.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anycastIPListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findAnycastIPListByID(ctx, conn, id)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Anycast IP List (%s)", id), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AnycastIpList, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anycastIPListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data anycastIPListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	etag, err := anycastIPListETag(ctx, conn, id)

	if retry.NotFound(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudFront Anycast IP List (%s)", id), err.Error())

		return
	}

	input := cloudfront.DeleteAnycastIpListInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}
	_, err = conn.DeleteAnycastIpList(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudFront Anycast IP List (%s)", id), err.Error())

		return
	}
}

func anycastIPListETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findAnycastIPListByID(ctx, conn, id)

	if err != nil {
		return "", err
	}

	return aws.ToString(output.ETag), nil
}

func findAnycastIPListByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetAnycastIpListOutput, error) {
	input := cloudfront.GetAnycastIpListInput{
		Id: aws.String(id),
	}

	return findAnycastIPList(ctx, conn, &input)
}

func findAnycastIPList(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetAnycastIpListInput) (*cloudfront.GetAnycastIpListOutput, error) {
	output, err := conn.GetAnycastIpList(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AnycastIpList == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func anycastIPListStatus(conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAnycastIPListByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		anycastIPList := output.AnycastIpList

		// When the list is backed by IPAM CIDRs, provisioning is asynchronous and
		// reported per-CIDR (e.g. "provisioning" -> "provisioned" -> "advertising"
		// -> "advertised", plus "failed-*" variants). Surface a failure and keep
		// waiting while any CIDR is still transitioning.
		if ipamConfig := anycastIPList.IpamConfig; ipamConfig != nil {
			for _, cidrConfig := range ipamConfig.IpamCidrConfigs {
				status := string(cidrConfig.Status)

				if strings.HasPrefix(status, "failed-") {
					return output, status, fmt.Errorf("IPAM CIDR (%s): %s", aws.ToString(cidrConfig.Cidr), status)
				}

				if strings.HasSuffix(status, "ing") {
					return output, anycastIPListDeploying, nil
				}
			}
		}

		return output, aws.ToString(anycastIPList.Status), nil
	}
}

func waitAnycastIPListDeployed(ctx context.Context, conn *cloudfront.Client, id string, timeout time.Duration) (*cloudfront.GetAnycastIpListOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{anycastIPListDeploying},
		Target:  []string{anycastIPListDeployed},
		Refresh: anycastIPListStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetAnycastIpListOutput); ok {
		return output, err
	}

	return nil, err
}

type anycastIPListResourceModel struct {
	AnycastIPs      fwtypes.ListOfString                                 `tfsdk:"anycast_ips"`
	ARN             types.String                                         `tfsdk:"arn"`
	ETag            types.String                                         `tfsdk:"etag"`
	ID              types.String                                         `tfsdk:"id"`
	IPCount         types.Int32                                          `tfsdk:"ip_count"`
	IpamCidrConfigs fwtypes.ListNestedObjectValueOf[ipamCidrConfigModel] `tfsdk:"ipam_cidr_config"`
	IpamConfig      fwtypes.ListNestedObjectValueOf[ipamConfigModel]     `tfsdk:"ipam_config"`
	Name            types.String                                         `tfsdk:"name"`
	Tags            tftags.Map                                           `tfsdk:"tags"`
	TagsAll         tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts        timeouts.Value                                       `tfsdk:"timeouts"`
}

// ipamCidrConfigModel is the configurable 'ipam_cidr_config' block, mapped to
// the CreateAnycastIpList 'IpamCidrConfigs' input.
type ipamCidrConfigModel struct {
	AnycastIP   types.String `tfsdk:"anycast_ip"`
	Cidr        types.String `tfsdk:"cidr"`
	IpamPoolARN fwtypes.ARN  `tfsdk:"ipam_pool_arn"`
}

// ipamConfigModel is the computed 'ipam_config' attribute, mapped from the
// AnycastIpList 'IpamConfig' response object.
type ipamConfigModel struct {
	IpamCidrConfigs fwtypes.ListNestedObjectValueOf[ipamCidrConfigStatusModel] `tfsdk:"ipam_cidr_configs"`
	Quantity        types.Int32                                                `tfsdk:"quantity"`
}

type ipamCidrConfigStatusModel struct {
	AnycastIP   types.String `tfsdk:"anycast_ip"`
	Cidr        types.String `tfsdk:"cidr"`
	IpamPoolARN fwtypes.ARN  `tfsdk:"ipam_pool_arn"`
	Status      types.String `tfsdk:"status"`
}
