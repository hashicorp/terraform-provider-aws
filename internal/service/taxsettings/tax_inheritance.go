// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package taxsettings

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/taxsettings"
	awstypes "github.com/aws/aws-sdk-go-v2/service/taxsettings/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_taxsettings_tax_inheritance", name="Tax Inheritance")
// @Region(global=true)
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
func newTaxInheritanceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &taxInheritanceResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameTaxInheritance = "Tax Inheritance"
)

type taxInheritanceResource struct {
	framework.ResourceWithModel[taxInheritanceResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
	framework.WithNoOpDelete
}

func (r *taxInheritanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"heritage_status": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *taxInheritanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan taxInheritanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.putTaxInheritanceHeritageStatus(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *taxInheritanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TaxSettingsClient(ctx)
	var state taxInheritanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	heritageStatus, err := findTaxInheritanceHeritageStatus(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.HeritageStatus)
		return
	}

	state.HeritageStatus = types.StringValue(string(*heritageStatus))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *taxInheritanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan taxInheritanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.putTaxInheritanceHeritageStatus(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *taxInheritanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("heritage_status"), req, resp)
}

func (r *taxInheritanceResource) putTaxInheritanceHeritageStatus(ctx context.Context, data *taxInheritanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := r.Meta().TaxSettingsClient(ctx)

	var input taxsettings.PutTaxInheritanceInput
	diags.Append(flex.Expand(ctx, data, &input)...)
	if diags.HasError() {
		return diags
	}

	// A read check is introduced to avoid unnecessary updates, i.e. when
	// the desired heritage status matches the remote.
	heritageStatus, err := findTaxInheritanceHeritageStatus(ctx, conn)
	if err != nil {
		diags.AddError("put heritage status for tax inheritance", err.Error())
		return diags
	}

	if input.HeritageStatus == *heritageStatus {
		return diags
	}

	// The heritage status can only be updated once within 15 minutes. The API returns a conflict exception
	// (409 http status code) if this constraint is not respected.
	_, err = conn.PutTaxInheritance(ctx, &input)
	if err != nil {
		diags.AddError("put heritage status for tax inheritance", err.Error())
		return diags
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	_, err = waitHeritageStatusUpdated(ctx, conn, data.HeritageStatus.ValueString(), createTimeout)
	if err != nil {
		diags.AddError("waiting heritage status update for tax inheritance", err.Error())
		return diags
	}

	return diags
}

func findTaxInheritanceHeritageStatus(ctx context.Context, conn *taxsettings.Client) (*awstypes.HeritageStatus, error) {
	input := taxsettings.GetTaxInheritanceInput{}

	out, err := conn.GetTaxInheritance(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.HeritageStatus == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return &out.HeritageStatus, nil
}

func waitHeritageStatusUpdated(ctx context.Context, conn *taxsettings.Client, expectedHeritageStatus string, timeout time.Duration) (*awstypes.HeritageStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{expectedHeritageStatus},
		Refresh:                   statusTaxInheritance(conn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.HeritageStatus); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusTaxInheritance(conn *taxsettings.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findTaxInheritanceHeritageStatus(ctx, conn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, aws.ToString((*string)(out)), nil
	}
}

type taxInheritanceResourceModel struct {
	HeritageStatus types.String   `tfsdk:"heritage_status"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}
