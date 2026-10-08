// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_api_gateway_domain_name_share", name="Domain Name Share")
// @IdentityAttribute("domain_name_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetDomainNameOutput")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain(t)")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
func newDomainNameShareResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainNameShareResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDomainNameShare = "Domain Name Share"
)

type domainNameShareResource struct {
	framework.ResourceWithModel[domainNameShareResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *domainNameShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_accounts": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(validators.AWSAccountID()),
				},
			},
			"domain_name_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root("domain_name_id")),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *domainNameShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data domainNameShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.APIGatewayClient(ctx)

	domainNameID := data.DomainNameID.ValueString()
	if err := putDomainNameShare(ctx, c, conn, domainNameID, fwflex.ExpandFrameworkStringValueSet(ctx, data.AllowedAccounts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	output, err := findDomainNameShareByDomainNameID(ctx, conn, domainNameID)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	allowedAccounts, err := domainNameShareAllowedAccounts(aws.ToString(output.ManagementPolicy))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	data.AllowedAccounts = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, allowedAccounts)
	data.DomainNameID = fwflex.StringToFramework(ctx, output.DomainNameId)
	data.ID = data.DomainNameID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *domainNameShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data domainNameShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	domainNameID := data.DomainNameID.ValueString()
	if domainNameID == "" {
		domainNameID = data.ID.ValueString()
	}

	output, err := findDomainNameShareByDomainNameID(ctx, conn, domainNameID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	allowedAccounts, err := domainNameShareAllowedAccounts(aws.ToString(output.ManagementPolicy))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}
	if len(allowedAccounts) == 0 {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(&retry.NotFoundError{}))
		resp.State.RemoveResource(ctx)
		return
	}

	data.AllowedAccounts = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, allowedAccounts)
	data.DomainNameID = fwflex.StringToFramework(ctx, output.DomainNameId)
	data.ID = data.DomainNameID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *domainNameShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data domainNameShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.Meta()
	conn := c.APIGatewayClient(ctx)

	domainNameID := data.DomainNameID.ValueString()
	if err := putDomainNameShare(ctx, c, conn, domainNameID, fwflex.ExpandFrameworkStringValueSet(ctx, data.AllowedAccounts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	output, err := findDomainNameShareByDomainNameID(ctx, conn, domainNameID)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	allowedAccounts, err := domainNameShareAllowedAccounts(aws.ToString(output.ManagementPolicy))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	data.AllowedAccounts = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, allowedAccounts)
	data.DomainNameID = fwflex.StringToFramework(ctx, output.DomainNameId)
	data.ID = data.DomainNameID

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *domainNameShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data domainNameShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	domainNameID := data.DomainNameID.ValueString()
	if domainNameID == "" {
		domainNameID = data.ID.ValueString()
	}

	output, err := findDomainNameByID(ctx, conn, domainNameID)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting API Gateway Domain Name Share (%s)", domainNameID), err.Error())
		return
	}

	input := apigateway.UpdateDomainNameInput{
		DomainName:   aws.String(aws.ToString(output.DomainName)),
		DomainNameId: aws.String(domainNameID),
		PatchOperations: []awstypes.PatchOperation{
			{
				Op:   awstypes.OpRemove,
				Path: aws.String("/managementPolicy"),
			},
		},
	}

	_, err = conn.UpdateDomainName(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "Invalid patch path") {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting API Gateway Domain Name Share (%s)", domainNameID), err.Error())
	}
}

func putDomainNameShare(ctx context.Context, c *conns.AWSClient, conn *apigateway.Client, domainNameID string, allowedAccounts []string) error {
	output, err := findDomainNameByID(ctx, conn, domainNameID)
	if err != nil {
		return err
	}

	managementPolicy, err := domainNameShareManagementPolicy(ctx, c, aws.ToString(output.DomainName), domainNameID, allowedAccounts)
	if err != nil {
		return err
	}

	input := apigateway.UpdateDomainNameInput{
		DomainName:   aws.String(aws.ToString(output.DomainName)),
		DomainNameId: aws.String(domainNameID),
		PatchOperations: []awstypes.PatchOperation{
			{
				Op:    awstypes.OpReplace,
				Path:  aws.String("/managementPolicy"),
				Value: aws.String(managementPolicy),
			},
		},
	}

	if _, err := conn.UpdateDomainName(ctx, &input); err != nil {
		return err
	}

	return nil
}

func findDomainNameShareByDomainNameID(ctx context.Context, conn *apigateway.Client, domainNameID string) (*apigateway.GetDomainNameOutput, error) {
	output, err := findDomainNameByID(ctx, conn, domainNameID)
	if err != nil {
		return nil, err
	}

	if aws.ToString(output.ManagementPolicy) == "" {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findDomainNameByID(ctx context.Context, conn *apigateway.Client, domainNameID string) (*apigateway.GetDomainNameOutput, error) {
	input := apigateway.GetDomainNamesInput{}

	for {
		output, err := conn.GetDomainNames(ctx, &input)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range output.Items {
			if aws.ToString(v.DomainNameId) == domainNameID {
				return findDomainNameByTwoPartKey(ctx, conn, aws.ToString(v.DomainName), domainNameID)
			}
		}

		if aws.ToString(output.Position) == "" {
			break
		}

		input.Position = output.Position
	}

	return nil, &retry.NotFoundError{LastError: tfresource.NewEmptyResultError()}
}

func domainNameShareManagementPolicy(ctx context.Context, c *conns.AWSClient, domainName, domainNameID string, allowedAccounts []string) (string, error) {
	accounts := slices.Clone(allowedAccounts)
	slices.Sort(accounts)

	principalARNs := make([]string, 0, len(accounts))
	for _, accountID := range accounts {
		principalARNs = append(principalARNs, fmt.Sprintf("arn:%s:iam::%s:root", c.Partition(ctx), accountID))
	}

	document := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{
			map[string]any{
				"Effect": "Allow",
				"Principal": map[string]any{
					"AWS": principalARNs,
				},
				"Action":   "apigateway:CreateAccessAssociation",
				"Resource": domainNameShareResourceARN(ctx, c, domainName, domainNameID),
			},
		},
	}

	b, err := json.Marshal(document)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func domainNameShareAllowedAccounts(policy string) ([]string, error) {
	type principal struct {
		AWS any `json:"AWS"`
	}

	type statement struct {
		Action    any       `json:"Action"`
		Principal principal `json:"Principal"`
	}

	type policyDocument struct {
		Statement []statement `json:"Statement"`
	}

	var document policyDocument
	if err := json.Unmarshal([]byte(policy), &document); err != nil {
		return nil, fmt.Errorf("parsing management policy JSON: %w", err)
	}

	accounts := make(map[string]struct{})

	for _, statement := range document.Statement {
		if !domainNameShareHasAction(statement.Action, "apigateway:CreateAccessAssociation") {
			continue
		}

		for _, principalARN := range domainNameShareStringSlice(statement.Principal.AWS) {
			if accountID, err := domainNameSharePrincipalAccountID(principalARN); err == nil && accountID != "" {
				accounts[accountID] = struct{}{}
			}
		}
	}

	results := make([]string, 0, len(accounts))
	for accountID := range accounts {
		results = append(results, accountID)
	}

	slices.Sort(results)

	return results, nil
}

func domainNameShareResourceARN(ctx context.Context, c *conns.AWSClient, domainName, domainNameID string) string {
	return fmt.Sprintf("arn:%s:apigateway:%s:%s:/domainnames/%s+%s", c.Partition(ctx), c.Region(ctx), c.AccountID(ctx), domainName, domainNameID)
}

func domainNameShareHasAction(actions any, expected string) bool {
	for _, action := range domainNameShareStringSlice(actions) {
		if strings.EqualFold(action, expected) {
			return true
		}
	}

	return false
}

func domainNameSharePrincipalAccountID(principalARN string) (string, error) {
	if principalARN == "" {
		return "", nil
	}

	if !strings.Contains(principalARN, ":") {
		return principalARN, nil
	}

	parsedARN, err := awsarn.Parse(principalARN)
	if err != nil {
		return "", err
	}

	return parsedARN.AccountID, nil
}

func domainNameShareStringSlice(v any) []string {
	switch v := v.(type) {
	case string:
		if v == "" {
			return nil
		}

		return []string{v}
	case []any:
		result := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok && s != "" {
				result = append(result, s)
			}
		}

		return result
	case []string:
		return v
	default:
		return nil
	}
}

type domainNameShareResourceModel struct {
	framework.WithRegionModel
	AllowedAccounts fwtypes.SetOfString `tfsdk:"allowed_accounts"`
	DomainNameID    types.String        `tfsdk:"domain_name_id"`
	ID              types.String        `tfsdk:"id"`
	Timeouts        timeouts.Value      `tfsdk:"timeouts"`
}
