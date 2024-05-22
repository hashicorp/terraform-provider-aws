// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_service", name="Service")
// @Tags(identifierAttribute="arn")
func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.AuthType](),
			},
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"custom_domain_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 255),
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 40),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameService = "Service"
)

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateServiceInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(d.Get(names.AttrName).(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auth_type"); ok {
		in.AuthType = types.AuthType(v.(string))
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		in.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_domain_name"); ok {
		in.CustomDomainName = aws.String(v.(string))
	}

	out, err := conn.CreateService(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameService, d.Get(names.AttrName).(string), err)
	}

	d.SetId(aws.ToString(out.Id))

	if _, err := waitServiceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForCreation, ResNameService, d.Id(), err)
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	out, err := findServiceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameService, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("auth_type", out.AuthType)
	d.Set(names.AttrCertificateARN, out.CertificateArn)
	d.Set("custom_domain_name", out.CustomDomainName)
	if out.DnsEntry != nil {
		if err := d.Set("dns_entry", []interface{}{flattenDNSEntry(out.DnsEntry)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_entry: %s", err)
		}
	} else {
		d.Set("dns_entry", nil)
	}
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrStatus, out.Status)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &vpclattice.UpdateServiceInput{
			ServiceIdentifier: aws.String(d.Id()),
		}

		if d.HasChanges("auth_type") {
			in.AuthType = types.AuthType(d.Get("auth_type").(string))
		}

		if d.HasChanges(names.AttrCertificateARN) {
			in.CertificateArn = aws.String(d.Get(names.AttrCertificateARN).(string))
		}

		_, err := conn.UpdateService(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionUpdating, ResNameService, d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPC Lattice Service: %s", d.Id())
	_, err := conn.DeleteService(ctx, &vpclattice.DeleteServiceInput{
		ServiceIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameService, d.Id(), err)
	}

	if _, err := waitServiceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameService, d.Id(), err)
	}

	return diags
}

func waitServiceCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.ServiceStatusCreateInProgress),
		Target:                    enum.Slice(types.ServiceStatusActive),
		Refresh:                   statusService(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitServiceDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.GetServiceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusDeleteInProgress, types.ServiceStatusActive),
		Target:  []string{},
		Refresh: statusService(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.GetServiceOutput); ok {
		return out, err
	}

	return nil, err
}

func statusService(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findServiceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findServiceByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.GetServiceOutput, error) {
	in := &vpclattice.GetServiceInput{
		ServiceIdentifier: aws.String(id),
	}
	out, err := conn.GetService(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func findService(ctx context.Context, conn *vpclattice.Client, filter tfslices.Predicate[types.ServiceSummary]) (*types.ServiceSummary, error) {
	output, err := findServices(ctx, conn, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServices(ctx context.Context, conn *vpclattice.Client, filter tfslices.Predicate[types.ServiceSummary]) ([]types.ServiceSummary, error) {
	input := &vpclattice.ListServicesInput{}
	var output []types.ServiceSummary
	paginator := vpclattice.NewListServicesPaginator(conn, input, func(options *vpclattice.ListServicesPaginatorOptions) {
		options.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func flattenDNSEntry(apiObject *types.DnsEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DomainName; v != nil {
		tfMap[names.AttrDomainName] = aws.ToString(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap[names.AttrHostedZoneID] = aws.ToString(v)
	}

	return tfMap
}
