// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_parameter_group", name="Parameter Group")
// @Tags(identifierAttribute="arn")
func resourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

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
				Default:  "Managed by Terraform",
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: parameterHash,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticache.CreateCacheParameterGroupInput{
		CacheParameterGroupName:   aws.String(name),
		CacheParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:               aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                      getTagsIn(ctx),
	}

	output, err := conn.CreateCacheParameterGroupWithContext(ctx, input)

	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache Parameter Group with tags: %s. Trying create without tags.", err)

		input.Tags = nil
		output, err = conn.CreateCacheParameterGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache Parameter Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.CacheParameterGroup.CacheParameterGroupName))
	d.Set(names.AttrARN, output.CacheParameterGroup.ARN)

	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	parameterGroup, err := findCacheParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, parameterGroup.ARN)
	d.Set(names.AttrDescription, parameterGroup.Description)
	d.Set("family", parameterGroup.CacheParameterGroupFamily)
	d.Set(names.AttrName, parameterGroup.CacheParameterGroupName)

	// Only include user customized parameters as there's hundreds of system/default ones.
	input := &elasticache.DescribeCacheParametersInput{
		CacheParameterGroupName: aws.String(d.Id()),
		Source:                  aws.String("user"),
	}

	output, err := conn.DescribeCacheParametersWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	d.Set("parameter", flattenParameters(output.Parameters))

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		toRemove, toAdd := parameterChanges(o, n)

		// We can only modify 20 parameters at a time, so walk them until
		// we've got them all.
		const maxParams = 20

		for len(toRemove) > 0 {
			var paramsToModify []*elasticache.ParameterNameValue
			if len(toRemove) <= maxParams {
				paramsToModify, toRemove = toRemove[:], nil
			} else {
				paramsToModify, toRemove = toRemove[:maxParams], toRemove[maxParams:]
			}

			err := resourceResetParameterGroup(ctx, conn, d.Get(names.AttrName).(string), paramsToModify)

			// When attempting to reset the reserved-memory parameter, the API
			// can return two types of error.
			//
			// In the commercial partition, it will return a 400 error with:
			//   InvalidParameterValue: Parameter reserved-memory doesn't exist
			//
			// In the GovCloud partition it will return the below 500 error,
			// which causes the AWS Go SDK to automatically retry and timeout:
			//   InternalFailure: An internal error has occurred. Please try your query again at a later time.
			//
			// Instead of hardcoding the reserved-memory parameter removal
			// above, which may become out of date, here we add logic to
			// workaround this API behavior

			if tfresource.TimedOut(err) || tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Parameter reserved-memory doesn't exist") {
				for i, paramToModify := range paramsToModify {
					if aws.StringValue(paramToModify.ParameterName) != "reserved-memory" {
						continue
					}

					// Always reset the top level error and remove the reset for reserved-memory
					err = nil
					paramsToModify = append(paramsToModify[:i], paramsToModify[i+1:]...)

					// If we are only trying to remove reserved-memory and not perform
					// an update to reserved-memory or reserved-memory-percent, we
					// can attempt to workaround the API issue by switching it to
					// reserved-memory-percent first then reset that temporary parameter.

					tryReservedMemoryPercentageWorkaround := true
					for _, configuredParameter := range toAdd {
						if aws.StringValue(configuredParameter.ParameterName) == "reserved-memory-percent" {
							tryReservedMemoryPercentageWorkaround = false
							break
						}
					}

					if !tryReservedMemoryPercentageWorkaround {
						break
					}

					// The reserved-memory-percent parameter does not exist in redis2.6 and redis2.8
					family := d.Get("family").(string)
					if family == "redis2.6" || family == "redis2.8" {
						log.Printf("[WARN] Cannot reset ElastiCache Parameter Group (%s) reserved-memory parameter with %s family", d.Id(), family)
						break
					}

					workaroundParams := []*elasticache.ParameterNameValue{
						{
							ParameterName:  aws.String("reserved-memory-percent"),
							ParameterValue: aws.String("0"),
						},
					}
					err = resourceModifyParameterGroup(ctx, conn, d.Get(names.AttrName).(string), paramsToModify)
					if err != nil {
						log.Printf("[WARN] Error attempting reserved-memory workaround to switch to reserved-memory-percent: %s", err)
						break
					}

					err = resourceResetParameterGroup(ctx, conn, d.Get(names.AttrName).(string), workaroundParams)
					if err != nil {
						log.Printf("[WARN] Error attempting reserved-memory workaround to reset reserved-memory-percent: %s", err)
					}

					break
				}

				// Retry any remaining parameter resets with reserved-memory potentially removed
				if len(paramsToModify) > 0 {
					err = resourceResetParameterGroup(ctx, conn, d.Get(names.AttrName).(string), paramsToModify)
				}
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting ElastiCache Parameter Group: %s", err)
			}
		}

		for len(toAdd) > 0 {
			var paramsToModify []*elasticache.ParameterNameValue
			if len(toAdd) <= maxParams {
				paramsToModify, toAdd = toAdd[:], nil
			} else {
				paramsToModify, toAdd = toAdd[:maxParams], toAdd[maxParams:]
			}

			err := resourceModifyParameterGroup(ctx, conn, d.Get(names.AttrName).(string), paramsToModify)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ElastiCache Parameter Group: %s", err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[INFO] Deleting ElastiCache Parameter Group: %s", d.Id())
	if err := deleteParameterGroup(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteParameterGroup(ctx context.Context, conn *elasticache.ElastiCache, name string) error {
	const (
		timeout = 3 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteCacheParameterGroupWithContext(ctx, &elasticache.DeleteCacheParameterGroupInput{
			CacheParameterGroupName: aws.String(name),
		})
	}, elasticache.ErrCodeInvalidCacheParameterGroupStateFault)

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting ElastiCache Parameter Group (%s): %s", name, err)
	}

	return err
}

var (
	parameterHash = sdkv2.SimpleSchemaSetFunc(names.AttrName, names.AttrValue)
)

func parameterChanges(o, n interface{}) (remove, addOrUpdate []*elasticache.ParameterNameValue) {
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	om := make(map[string]*elasticache.ParameterNameValue, os.Len())
	for _, raw := range os.List() {
		param := raw.(map[string]interface{})
		om[param[names.AttrName].(string)] = expandParameter(param)
	}
	nm := make(map[string]*elasticache.ParameterNameValue, len(addOrUpdate))
	for _, raw := range ns.List() {
		param := raw.(map[string]interface{})
		nm[param[names.AttrName].(string)] = expandParameter(param)
	}

	// Remove: key is in old, but not in new
	remove = make([]*elasticache.ParameterNameValue, 0, os.Len())
	for k := range om {
		if _, ok := nm[k]; !ok {
			remove = append(remove, om[k])
		}
	}

	// Add or Update: key is in new, but not in old or has changed value
	addOrUpdate = make([]*elasticache.ParameterNameValue, 0, ns.Len())
	for k, nv := range nm {
		ov, ok := om[k]
		if !ok || ok && (aws.StringValue(nv.ParameterValue) != aws.StringValue(ov.ParameterValue)) {
			addOrUpdate = append(addOrUpdate, nm[k])
		}
	}

	return remove, addOrUpdate
}

func resourceResetParameterGroup(ctx context.Context, conn *elasticache.ElastiCache, name string, parameters []*elasticache.ParameterNameValue) error {
	input := elasticache.ResetCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(name),
		ParameterNameValues:     parameters,
	}
	return retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, err := conn.ResetCacheParameterGroupWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidCacheParameterGroupStateFault, " has pending changes") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

func resourceModifyParameterGroup(ctx context.Context, conn *elasticache.ElastiCache, name string, parameters []*elasticache.ParameterNameValue) error {
	input := elasticache.ModifyCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(name),
		ParameterNameValues:     parameters,
	}
	_, err := conn.ModifyCacheParameterGroupWithContext(ctx, &input)
	return err
}

func findCacheParameterGroupByName(ctx context.Context, conn *elasticache.ElastiCache, name string) (*elasticache.CacheParameterGroup, error) {
	input := &elasticache.DescribeCacheParameterGroupsInput{
		CacheParameterGroupName: aws.String(name),
	}

	return findCacheParameterGroup(ctx, conn, input, tfslices.PredicateTrue[*elasticache.CacheParameterGroup]())
}

func findCacheParameterGroup(ctx context.Context, conn *elasticache.ElastiCache, input *elasticache.DescribeCacheParameterGroupsInput, filter tfslices.Predicate[*elasticache.CacheParameterGroup]) (*elasticache.CacheParameterGroup, error) {
	output, err := findCacheParameterGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findCacheParameterGroups(ctx context.Context, conn *elasticache.ElastiCache, input *elasticache.DescribeCacheParameterGroupsInput, filter tfslices.Predicate[*elasticache.CacheParameterGroup]) ([]*elasticache.CacheParameterGroup, error) {
	var output []*elasticache.CacheParameterGroup

	err := conn.DescribeCacheParameterGroupsPagesWithContext(ctx, input, func(page *elasticache.DescribeCacheParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CacheParameterGroups {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func expandParameter(tfMap map[string]interface{}) *elasticache.ParameterNameValue {
	return &elasticache.ParameterNameValue{
		ParameterName:  aws.String(tfMap[names.AttrName].(string)),
		ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
	}
}

func flattenParameters(apiObjects []*elasticache.Parameter) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		if apiObject.ParameterValue != nil {
			tfList = append(tfList, map[string]interface{}{
				names.AttrName:  strings.ToLower(aws.StringValue(apiObject.ParameterName)),
				names.AttrValue: aws.StringValue(apiObject.ParameterValue),
			})
		}
	}

	return tfList
}
