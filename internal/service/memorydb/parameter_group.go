// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_parameter_group", name="Parameter Group")
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
			names.AttrFamily: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validateResourceName(parameterGroupNameMaxLength),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validateResourceNamePrefix(parameterGroupNameMaxLength - id.UniqueIDSuffixLength),
			},
			names.AttrParameter: {
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
	}
}

var (
	parameterHash = sdkv2.SimpleSchemaSetFunc(names.AttrName, names.AttrValue)
)

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateParameterGroupInput{
		Description:        aws.String(d.Get(names.AttrDescription).(string)),
		Family:             aws.String(d.Get(names.AttrFamily).(string)),
		ParameterGroupName: aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	output, err := conn.CreateParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, output.ParameterGroup.ARN)

	// Update to apply parameter changes.
	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	group, err := findParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MemoryDB Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MemoryDB Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.ARN)
	d.Set(names.AttrDescription, group.Description)
	d.Set(names.AttrFamily, group.Family)
	d.Set(names.AttrName, group.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(group.Name)))

	userDefinedParameters := createUserDefinedParameterMap(d)
	parameters, err := listParameterGroupParameters(ctx, conn, d.Get(names.AttrFamily).(string), d.Id(), userDefinedParameters)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set(names.AttrParameter, flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		toRemove, toAdd := parameterChanges(o, n)

		// The API is limited to updating no more than 20 parameters at a time.
		const maxParams = 20

		// Removing a parameter from state is equivalent to resetting it
		// to its default state.
		for chunk := range slices.Chunk(toRemove, maxParams) {
			input := &memorydb.ResetParameterGroupInput{
				ParameterGroupName: aws.String(d.Id()),
				ParameterNames: tfslices.ApplyToAll(chunk, func(v awstypes.ParameterNameValue) string {
					return aws.ToString(v.ParameterName)
				}),
			}

			const (
				timeout = 30 * time.Second
			)
			_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterGroupStateFault](ctx, timeout, func() (any, error) {
				return conn.ResetParameterGroup(ctx, input)
			}, " has pending changes")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting MemoryDB Parameter Group (%s) parameters to defaults: %s", d.Id(), err)
			}
		}

		for chunk := range slices.Chunk(toAdd, maxParams) {
			input := &memorydb.UpdateParameterGroupInput{
				ParameterGroupName:  aws.String(d.Id()),
				ParameterNameValues: chunk,
			}

			_, err := conn.UpdateParameterGroup(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying MemoryDB Parameter Group (%s) parameters: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MemoryDBClient(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Parameter Group: (%s)", d.Id())
	_, err := conn.DeleteParameterGroup(ctx, &memorydb.DeleteParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findParameterGroupByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ParameterGroup, error) {
	input := &memorydb.DescribeParameterGroupsInput{
		ParameterGroupName: aws.String(name),
	}

	return findParameterGroup(ctx, conn, input)
}

func findParameterGroup(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeParameterGroupsInput) (*awstypes.ParameterGroup, error) {
	output, err := findParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findParameterGroups(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeParameterGroupsInput) ([]awstypes.ParameterGroup, error) {
	var output []awstypes.ParameterGroup

	pages := memorydb.NewDescribeParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ParameterGroups...)
	}

	return output, nil
}

func findParametersByParameterGroupName(ctx context.Context, conn *memorydb.Client, name string) ([]awstypes.Parameter, error) {
	input := &memorydb.DescribeParametersInput{
		ParameterGroupName: aws.String(name),
	}

	return findParameters(ctx, conn, input)
}

func findParameters(ctx context.Context, conn *memorydb.Client, input *memorydb.DescribeParametersInput) ([]awstypes.Parameter, error) {
	var output []awstypes.Parameter

	pages := memorydb.NewDescribeParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Parameters...)
	}

	return output, nil
}

// listParameterGroupParameters returns the user-defined MemoryDB parameters
// in the group with the given name and family.
//
// Parameters given in userDefined will be returned even if the value is equal
// to the default.
func listParameterGroupParameters(ctx context.Context, conn *memorydb.Client, family, name string, userDefined map[string]string) ([]awstypes.Parameter, error) {
	// There isn't an official API for defaults, and the mapping of family
	// to default parameter group name is a guess.
	defaultsFamily := "default." + strings.ReplaceAll(family, "_", "-")
	defaults, err := findParametersByParameterGroupName(ctx, conn, defaultsFamily)

	if err != nil {
		return nil, fmt.Errorf("reading MemoryDB Parameter Group (%s) parameters: %w", defaultsFamily, err)
	}

	defaultValueByName := map[string]string{}
	for _, v := range defaults {
		defaultValueByName[aws.ToString(v.Name)] = aws.ToString(v.Value)
	}

	current, err := findParametersByParameterGroupName(ctx, conn, name)

	if err != nil {
		return nil, fmt.Errorf("reading MemoryDB Parameter Group (%s) parameters: %w", name, err)
	}

	var apiObjects []awstypes.Parameter

	for _, v := range current {
		name := aws.ToString(v.Name)
		currentValue := aws.ToString(v.Value)
		defaultValue := defaultValueByName[name]
		_, isUserDefined := userDefined[name]

		if currentValue != defaultValue || isUserDefined {
			apiObjects = append(apiObjects, v)
		}
	}

	return apiObjects, nil
}

func parameterChanges(o, n any) (remove, addOrUpdate []awstypes.ParameterNameValue) {
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}
	os, ns := o.(*schema.Set), n.(*schema.Set)

	om := make(map[string]awstypes.ParameterNameValue, os.Len())
	for _, tfMapRaw := range os.List() {
		tfMap := tfMapRaw.(map[string]any)
		om[tfMap[names.AttrName].(string)] = expandParameterNameValue(tfMap)
	}
	nm := make(map[string]awstypes.ParameterNameValue, len(addOrUpdate))
	for _, tfMapRaw := range ns.List() {
		tfMap := tfMapRaw.(map[string]any)
		nm[tfMap[names.AttrName].(string)] = expandParameterNameValue(tfMap)
	}

	// Remove: key is in old, but not in new
	remove = make([]awstypes.ParameterNameValue, 0, os.Len())
	for k := range om {
		if _, ok := nm[k]; !ok {
			remove = append(remove, om[k])
		}
	}

	// Add or Update: key is in new, but not in old or has changed value
	addOrUpdate = make([]awstypes.ParameterNameValue, 0, ns.Len())
	for k, nv := range nm {
		ov, ok := om[k]
		if !ok || ok && (aws.ToString(nv.ParameterValue) != aws.ToString(ov.ParameterValue)) {
			addOrUpdate = append(addOrUpdate, nm[k])
		}
	}

	return remove, addOrUpdate
}

func flattenParameters(apiObjects []awstypes.Parameter) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		if apiObject.Value != nil {
			tfList = append(tfList, map[string]any{
				names.AttrName:  strings.ToLower(aws.ToString(apiObject.Name)),
				names.AttrValue: aws.ToString(apiObject.Value),
			})
		}
	}

	return tfList
}

func expandParameterNameValue(tfMap map[string]any) awstypes.ParameterNameValue {
	return awstypes.ParameterNameValue{
		ParameterName:  aws.String(tfMap[names.AttrName].(string)),
		ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
	}
}

func createUserDefinedParameterMap(d *schema.ResourceData) map[string]string {
	result := map[string]string{}

	for _, param := range d.Get(names.AttrParameter).(*schema.Set).List() {
		m, ok := param.(map[string]any)
		if !ok {
			continue
		}

		name, ok := m[names.AttrName].(string)
		if !ok || name == "" {
			continue
		}

		value, ok := m[names.AttrValue].(string)
		if !ok || value == "" {
			continue
		}

		result[name] = value
	}

	return result
}
