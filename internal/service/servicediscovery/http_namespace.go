// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_service_discovery_http_namespace", name="HTTP Namespace")
// @Tags(identifierAttribute="arn")
func resourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHTTPNamespaceCreate,
		ReadWithoutTimeout:   resourceHTTPNamespaceRead,
		UpdateWithoutTimeout: resourceHTTPNamespaceUpdate,
		DeleteWithoutTimeout: resourceHTTPNamespaceDelete,

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
			"http_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validNamespaceName,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHTTPNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicediscovery.CreateHttpNamespaceInput{
		CreatorRequestId: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHttpNamespace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Discovery HTTP Namespace (%s): %s", name, err)
	}

	operation, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery HTTP Namespace (%s) create: %s", name, err)
	}

	d.SetId(operation.Targets[string(awstypes.OperationTargetTypeNamespace)])

	return append(diags, resourceHTTPNamespaceRead(ctx, d, meta)...)
}

func resourceHTTPNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	ns, err := findNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery HTTP Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(ns.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, ns.Description)
	if ns.Properties != nil && ns.Properties.HttpProperties != nil {
		d.Set("http_name", ns.Properties.HttpProperties.HttpName)
	} else {
		d.Set("http_name", nil)
	}
	d.Set(names.AttrName, ns.Name)

	return diags
}

func resourceHTTPNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceHTTPNamespaceRead(ctx, d, meta)
}

func resourceHTTPNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceDiscoveryClient(ctx)

	log.Printf("[INFO] Deleting Service Discovery HTTP Namespace: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ResourceInUse](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteNamespace(ctx, &servicediscovery.DeleteNamespaceInput{
			Id: aws.String(d.Id()),
		})
	})

	if errs.IsA[*awstypes.NamespaceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	if output := outputRaw.(*servicediscovery.DeleteNamespaceOutput); output != nil && output.OperationId != nil {
		if _, err := waitOperationSucceeded(ctx, conn, aws.ToString(output.OperationId)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Service Discovery HTTP Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func findNamespace(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListNamespacesInput, filter tfslices.Predicate[*awstypes.NamespaceSummary]) (*awstypes.NamespaceSummary, error) {
	output, err := findNamespaces(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNamespaces(ctx context.Context, conn *servicediscovery.Client, input *servicediscovery.ListNamespacesInput, filter tfslices.Predicate[*awstypes.NamespaceSummary]) ([]awstypes.NamespaceSummary, error) {
	var output []awstypes.NamespaceSummary

	pages := servicediscovery.NewListNamespacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Namespaces {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func namespaceTypeFilter(nsType awstypes.NamespaceType) awstypes.NamespaceFilter {
	return awstypes.NamespaceFilter{
		Condition: awstypes.FilterConditionEq,
		Name:      awstypes.NamespaceFilterNameType,
		Values:    enum.Slice(nsType),
	}
}

func findNamespacesByType(ctx context.Context, conn *servicediscovery.Client, nsType awstypes.NamespaceType) ([]awstypes.NamespaceSummary, error) {
	input := &servicediscovery.ListNamespacesInput{
		Filters: []awstypes.NamespaceFilter{namespaceTypeFilter(nsType)},
	}

	return findNamespaces(ctx, conn, input, tfslices.PredicateTrue[*awstypes.NamespaceSummary]())
}

func findNamespaceByNameAndType(ctx context.Context, conn *servicediscovery.Client, name string, nsType awstypes.NamespaceType) (*awstypes.NamespaceSummary, error) {
	input := &servicediscovery.ListNamespacesInput{
		Filters: []awstypes.NamespaceFilter{namespaceTypeFilter(nsType)},
	}

	return findNamespace(ctx, conn, input, func(v *awstypes.NamespaceSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findNamespaceByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Namespace, error) {
	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetNamespace(ctx, input)

	if errs.IsA[*awstypes.NamespaceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Namespace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}
