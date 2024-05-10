// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_custom_domain_association", name="Custom Domain Association")
func resourceCustomDomainAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomDomainAssociationCreate,
		ReadWithoutTimeout:   resourceCustomDomainAssociationRead,
		DeleteWithoutTimeout: resourceCustomDomainAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"certificate_validation_records": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_target": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"enable_www_subdomain": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCustomDomainAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	serviceARN := d.Get("service_arn").(string)
	id := customDomainAssociationCreateResourceID(domainName, serviceARN)
	input := &apprunner.AssociateCustomDomainInput{
		DomainName:         aws.String(domainName),
		EnableWWWSubdomain: aws.Bool(d.Get("enable_www_subdomain").(bool)),
		ServiceArn:         aws.String(serviceARN),
	}

	output, err := conn.AssociateCustomDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner Custom Domain Association (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("dns_target", output.DNSTarget)

	if _, err := waitCustomDomainAssociationCreated(ctx, conn, domainName, serviceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Custom Domain Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCustomDomainAssociationRead(ctx, d, meta)...)
}

func resourceCustomDomainAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	domainName, serviceArn, err := customDomainAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	customDomain, err := findCustomDomainByTwoPartKey(ctx, conn, domainName, serviceArn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner Custom Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner Custom Domain Association (%s): %s", d.Id(), err)
	}

	if err := d.Set("certificate_validation_records", flattenCustomDomainCertificateValidationRecords(customDomain.CertificateValidationRecords)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_validation_records: %s", err)
	}
	d.Set(names.AttrDomainName, customDomain.DomainName)
	d.Set("enable_www_subdomain", customDomain.EnableWWWSubdomain)
	d.Set("service_arn", serviceArn)
	d.Set(names.AttrStatus, customDomain.Status)

	return diags
}

func resourceCustomDomainAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	domainName, serviceARN, err := customDomainAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting App Runner Custom Domain Association: %s", d.Id())
	_, err = conn.DisassociateCustomDomain(ctx, &apprunner.DisassociateCustomDomainInput{
		DomainName: aws.String(domainName),
		ServiceArn: aws.String(serviceARN),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner Custom Domain Association (%s): %s", d.Id(), err)
	}

	if _, err := waitCustomDomainAssociationDeleted(ctx, conn, domainName, serviceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Custom Domain Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const customDomainAssociationIDSeparator = ","

func customDomainAssociationCreateResourceID(domainName, serviceARN string) string {
	parts := []string{domainName, serviceARN}
	id := strings.Join(parts, customDomainAssociationIDSeparator)

	return id
}

func customDomainAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, customDomainAssociationIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected domain_name%[2]service_arn", id, customDomainAssociationIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findCustomDomainByTwoPartKey(ctx context.Context, conn *apprunner.Client, domainName, serviceARN string) (*types.CustomDomain, error) {
	input := &apprunner.DescribeCustomDomainsInput{
		ServiceArn: aws.String(serviceARN),
	}

	return findCustomDomain(ctx, conn, input, func(v *types.CustomDomain) bool {
		return aws.ToString(v.DomainName) == domainName
	})
}

func findCustomDomain(ctx context.Context, conn *apprunner.Client, input *apprunner.DescribeCustomDomainsInput, filter tfslices.Predicate[*types.CustomDomain]) (*types.CustomDomain, error) {
	output, err := findCustomDomains(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findCustomDomains(ctx context.Context, conn *apprunner.Client, input *apprunner.DescribeCustomDomainsInput, filter tfslices.Predicate[*types.CustomDomain]) ([]*types.CustomDomain, error) {
	var output []*types.CustomDomain

	err := forEachCustomDomainPage(ctx, conn, input, func(page *apprunner.DescribeCustomDomainsOutput) {
		for _, v := range page.CustomDomains {
			v := v
			if v := &v; filter(v) {
				output = append(output, v)
			}
		}
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func forEachCustomDomainPage(ctx context.Context, conn *apprunner.Client, input *apprunner.DescribeCustomDomainsInput, fn func(page *apprunner.DescribeCustomDomainsOutput)) error {
	pages := apprunner.NewDescribeCustomDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return err
		}

		fn(page)
	}

	return nil
}

const (
	customDomainAssociationStatusActive                          = "active"
	customDomainAssociationStatusBindingCertificate              = "binding_certificate"
	customDomainAssociationStatusCreating                        = "creating"
	customDomainAssociationStatusDeleting                        = "deleting"
	customDomainAssociationStatusPendingCertificateDNSValidation = "pending_certificate_dns_validation"
)

func statusCustomDomain(ctx context.Context, conn *apprunner.Client, domainName, serviceARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCustomDomainByTwoPartKey(ctx, conn, domainName, serviceARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitCustomDomainAssociationCreated(ctx context.Context, conn *apprunner.Client, domainName, serviceARN string) (*types.CustomDomain, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{customDomainAssociationStatusCreating},
		Target:  []string{customDomainAssociationStatusPendingCertificateDNSValidation, customDomainAssociationStatusBindingCertificate},
		Refresh: statusCustomDomain(ctx, conn, domainName, serviceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CustomDomain); ok {
		return output, err
	}

	return nil, err
}

func waitCustomDomainAssociationDeleted(ctx context.Context, conn *apprunner.Client, domainName, serviceARN string) (*types.CustomDomain, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{customDomainAssociationStatusActive, customDomainAssociationStatusDeleting},
		Target:  []string{},
		Refresh: statusCustomDomain(ctx, conn, domainName, serviceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CustomDomain); ok {
		return output, err
	}

	return nil, err
}

func flattenCustomDomainCertificateValidationRecords(records []types.CertificateValidationRecord) []interface{} {
	var results []interface{}

	for _, record := range records {
		m := map[string]interface{}{
			names.AttrName:   aws.ToString(record.Name),
			names.AttrStatus: record.Status,
			names.AttrType:   aws.ToString(record.Type),
			names.AttrValue:  aws.ToString(record.Value),
		}

		results = append(results, m)
	}

	return results
}
