// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource
func newResourceCIDRLocation(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCIDRLocation{}

	return r, nil
}

type resourceCIDRLocation struct {
	framework.ResourceWithConfigure
}

func (r *resourceCIDRLocation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_route53_cidr_location"
}

func (r *resourceCIDRLocation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cidr_blocks": schema.SetAttribute{
				Required:    true,
				ElementType: fwtypes.CIDRBlockType,
			},
			"cidr_collection_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(16),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`), `can include letters, digits, underscore (_) and the dash (-) character`),
				},
			},
		},
	}
}

func (r *resourceCIDRLocation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceCIDRLocationData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().Route53Conn(ctx)

	collectionID := data.CIDRCollectionID.ValueString()
	collection, err := findCIDRCollectionByID(ctx, conn, collectionID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 CIDR Collection (%s)", collectionID), err.Error())

		return
	}

	name := data.Name.ValueString()
	input := &route53.ChangeCidrCollectionInput{
		Changes: []*route53.CidrCollectionChange{{
			Action:       aws.String(route53.CidrCollectionChangeActionPut),
			CidrList:     flex.ExpandFrameworkStringSet(ctx, data.CIDRBlocks),
			LocationName: aws.String(name),
		}},
		CollectionVersion: collection.Version,
		Id:                aws.String(collectionID),
	}

	_, err = conn.ChangeCidrCollectionWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Route 53 CIDR Location (%s)", name), err.Error())

		return
	}

	data.ID = types.StringValue(cidrLocationCreateResourceID(collectionID, name))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCIDRLocation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceCIDRLocationData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	collectionID, name, err := cidrLocationParseResourceID(data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().Route53Conn(ctx)

	cidrBlocks, err := findCIDRLocationByTwoPartKey(ctx, conn, collectionID, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 CIDR Location (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if n := len(cidrBlocks); n > 0 {
		elems := make([]attr.Value, n)
		for i, cidrBlock := range cidrBlocks {
			elems[i] = fwtypes.CIDRBlockValue(cidrBlock)
		}
		data.CIDRBlocks = types.SetValueMust(fwtypes.CIDRBlockType, elems)
	} else {
		data.CIDRBlocks = types.SetNull(fwtypes.CIDRBlockType)
	}
	data.CIDRCollectionID = types.StringValue(collectionID)
	data.Name = types.StringValue(name)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCIDRLocation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceCIDRLocationData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	collectionID, name, err := cidrLocationParseResourceID(new.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().Route53Conn(ctx)

	collection, err := findCIDRCollectionByID(ctx, conn, collectionID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 CIDR Collection (%s)", collectionID), err.Error())

		return
	}

	oldCIDRBlocks := flex.ExpandFrameworkStringValueSet(ctx, old.CIDRBlocks)
	newCIDRBlocks := flex.ExpandFrameworkStringValueSet(ctx, new.CIDRBlocks)
	add := newCIDRBlocks.Difference(oldCIDRBlocks)
	del := oldCIDRBlocks.Difference(newCIDRBlocks)
	collectionVersion := collection.Version

	if len(add) > 0 {
		input := &route53.ChangeCidrCollectionInput{
			Changes: []*route53.CidrCollectionChange{{
				Action:       aws.String(route53.CidrCollectionChangeActionPut),
				CidrList:     aws.StringSlice(add),
				LocationName: aws.String(name),
			}},
			CollectionVersion: collectionVersion,
			Id:                aws.String(collectionID),
		}

		_, err = conn.ChangeCidrCollectionWithContext(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("adding CIDR blocks to Route 53 CIDR Location (%s)", new.ID.ValueString()), err.Error())

			return
		}

		collectionVersion = nil // Clear the collection version as it will have changed after the last operation.
	}

	if len(del) > 0 {
		input := &route53.ChangeCidrCollectionInput{
			Changes: []*route53.CidrCollectionChange{{
				Action:       aws.String(route53.CidrCollectionChangeActionDeleteIfExists),
				CidrList:     aws.StringSlice(del),
				LocationName: aws.String(name),
			}},
			CollectionVersion: collectionVersion,
			Id:                aws.String(collectionID),
		}

		_, err = conn.ChangeCidrCollectionWithContext(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("removing CIDR blocks from Route 53 CIDR Location (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceCIDRLocation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceCIDRLocationData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	collectionID, name, err := cidrLocationParseResourceID(data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().Route53Conn(ctx)

	collection, err := findCIDRCollectionByID(ctx, conn, collectionID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Route 53 CIDR Collection (%s)", collectionID), err.Error())

		return
	}

	tflog.Debug(ctx, "deleting Route 53 CIDR Location", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	input := &route53.ChangeCidrCollectionInput{
		Changes: []*route53.CidrCollectionChange{{
			Action:       aws.String(route53.CidrCollectionChangeActionDeleteIfExists),
			CidrList:     flex.ExpandFrameworkStringSet(ctx, data.CIDRBlocks),
			LocationName: aws.String(name),
		}},
		CollectionVersion: collection.Version,
		Id:                aws.String(collectionID),
	}

	_, err = conn.ChangeCidrCollectionWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Route 53 CIDR Location (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourceCIDRLocation) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

type resourceCIDRLocationData struct {
	CIDRBlocks       types.Set    `tfsdk:"cidr_blocks"`
	CIDRCollectionID types.String `tfsdk:"cidr_collection_id"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
}

func findCIDRLocationByTwoPartKey(ctx context.Context, conn *route53.Route53, collectionID, locationName string) ([]string, error) {
	input := &route53.ListCidrBlocksInput{
		CollectionId: aws.String(collectionID),
		LocationName: aws.String(locationName),
	}
	var output []string

	err := conn.ListCidrBlocksPagesWithContext(ctx, input, func(page *route53.ListCidrBlocksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CidrBlocks {
			if v == nil {
				continue
			}

			output = append(output, aws.StringValue(v.CidrBlock))
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchCidrCollectionException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

const cidrLocationResourceIDSeparator = ","

func cidrLocationCreateResourceID(collectionID, locationName string) string {
	parts := []string{collectionID, locationName}
	id := strings.Join(parts, cidrLocationResourceIDSeparator)

	return id
}

func cidrLocationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, cidrLocationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected CIDRCOLLECTIONID%[2]sLOCATIONNAME", id, cidrLocationResourceIDSeparator)
}
