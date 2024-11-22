// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package swf

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/swf"
	"github.com/aws/aws-sdk-go-v2/service/swf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_swf_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"workflow_execution_retention_period_in_days": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value, err := strconv.Atoi(v.(string))
					if err != nil || value > 90 || value < 0 {
						es = append(es, fmt.Errorf(
							"%q must be between 0 and 90 days inclusive", k))
					}
					return
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SWFClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &swf.RegisterDomainInput{
		Name:                                   aws.String(name),
		Tags:                                   getTagsIn(ctx),
		WorkflowExecutionRetentionPeriodInDays: aws.String(d.Get("workflow_execution_retention_period_in_days").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.RegisterDomain(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SWF Domain (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SWFClient(ctx)

	output, err := findDomainByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SWF Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SWF Domain (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(output.DomainInfo.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, output.DomainInfo.Description)
	d.Set(names.AttrName, output.DomainInfo.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.DomainInfo.Name)))
	d.Set("workflow_execution_retention_period_in_days", output.Configuration.WorkflowExecutionRetentionPeriodInDays)

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceDomainRead(ctx, d, meta)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SWFClient(ctx)

	_, err := conn.DeprecateDomain(ctx, &swf.DeprecateDomainInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*types.DomainDeprecatedFault](err) || errs.IsA[*types.UnknownResourceFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SWF Domain (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainByName(ctx context.Context, conn *swf.Client, name string) (*swf.DescribeDomainOutput, error) {
	input := &swf.DescribeDomainInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDomain(ctx, input)

	if errs.IsA[*types.UnknownResourceFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil || output.DomainInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.DomainInfo.Status; status == types.RegistrationStatusDeprecated {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}
