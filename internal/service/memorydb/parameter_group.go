// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_memorydb_parameter_group", name="Parameter Group")
// @Tags(identifierAttribute="arn")
func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

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
				Set: ParameterHash,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &memorydb.CreateParameterGroupInput{
		Description:        aws.String(d.Get(names.AttrDescription).(string)),
		Family:             aws.String(d.Get(names.AttrFamily).(string)),
		ParameterGroupName: aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating MemoryDB Parameter Group: %s", input)
	output, err := conn.CreateParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MemoryDB Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, output.ParameterGroup.ARN)

	log.Printf("[INFO] MemoryDB Parameter Group ID: %s", d.Id())

	// Update to apply parameter changes.
	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		toRemove, toAdd := ParameterChanges(o, n)

		log.Printf("[DEBUG] Updating MemoryDB Parameter Group (%s)", d.Id())
		log.Printf("[DEBUG] Parameters to remove: %#v", toRemove)
		log.Printf("[DEBUG] Parameters to add or update: %#v", toAdd)

		// The API is limited to updating no more than 20 parameters at a time.
		const maxParams = 20

		for len(toRemove) > 0 {
			// Removing a parameter from state is equivalent to resetting it
			// to its default state.

			var paramsToReset []*memorydb.ParameterNameValue
			if len(toRemove) <= maxParams {
				paramsToReset, toRemove = toRemove[:], nil
			} else {
				paramsToReset, toRemove = toRemove[:maxParams], toRemove[maxParams:]
			}

			err := resetParameterGroupParameters(ctx, conn, d.Get(names.AttrName).(string), paramsToReset)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting MemoryDB Parameter Group (%s) parameters to defaults: %s", d.Id(), err)
			}
		}

		for len(toAdd) > 0 {
			var paramsToModify []*memorydb.ParameterNameValue
			if len(toAdd) <= maxParams {
				paramsToModify, toAdd = toAdd[:], nil
			} else {
				paramsToModify, toAdd = toAdd[:maxParams], toAdd[maxParams:]
			}

			err := modifyParameterGroupParameters(ctx, conn, d.Get(names.AttrName).(string), paramsToModify)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying MemoryDB Parameter Group (%s) parameters: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	group, err := FindParameterGroupByName(ctx, conn, d.Id())

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
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(group.Name)))

	userDefinedParameters := createUserDefinedParameterMap(d)

	parameters, err := listParameterGroupParameters(ctx, conn, d.Get(names.AttrFamily).(string), d.Id(), userDefinedParameters)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing parameters for MemoryDB Parameter Group (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrParameter, flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "failed to set parameter: %s", err)
	}

	return diags
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)

	log.Printf("[DEBUG] Deleting MemoryDB Parameter Group: (%s)", d.Id())
	_, err := conn.DeleteParameterGroupWithContext(ctx, &memorydb.DeleteParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MemoryDB Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

// resetParameterGroupParameters resets the given parameters to their default values.
func resetParameterGroupParameters(ctx context.Context, conn *memorydb.MemoryDB, name string, parameters []*memorydb.ParameterNameValue) error {
	var parameterNames []*string
	for _, parameter := range parameters {
		parameterNames = append(parameterNames, parameter.ParameterName)
	}

	input := memorydb.ResetParameterGroupInput{
		ParameterGroupName: aws.String(name),
		ParameterNames:     parameterNames,
	}

	return retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		_, err := conn.ResetParameterGroupWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, memorydb.ErrCodeInvalidParameterGroupStateFault, " has pending changes") {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
}

// modifyParameterGroupParameters updates the given parameters.
func modifyParameterGroupParameters(ctx context.Context, conn *memorydb.MemoryDB, name string, parameters []*memorydb.ParameterNameValue) error {
	input := memorydb.UpdateParameterGroupInput{
		ParameterGroupName:  aws.String(name),
		ParameterNameValues: parameters,
	}
	_, err := conn.UpdateParameterGroupWithContext(ctx, &input)
	return err
}

// listParameterGroupParameters returns the user-defined MemoryDB parameters
// in the group with the given name and family.
//
// Parameters given in userDefined will be returned even if the value is equal
// to the default.
func listParameterGroupParameters(ctx context.Context, conn *memorydb.MemoryDB, family, name string, userDefined map[string]string) ([]*memorydb.Parameter, error) {
	query := func(ctx context.Context, parameterGroupName string) ([]*memorydb.Parameter, error) {
		input := memorydb.DescribeParametersInput{
			ParameterGroupName: aws.String(parameterGroupName),
		}

		output, err := conn.DescribeParametersWithContext(ctx, &input)
		if err != nil {
			return nil, err
		}

		return output.Parameters, nil
	}

	// There isn't an official API for defaults, and the mapping of family
	// to default parameter group name is a guess.

	defaultsFamily := "default." + strings.ReplaceAll(family, "_", "-")

	defaults, err := query(ctx, defaultsFamily)
	if err != nil {
		return nil, fmt.Errorf("list defaults for family %s: %w", defaultsFamily, err)
	}

	defaultValueByName := map[string]string{}
	for _, defaultPV := range defaults {
		defaultValueByName[aws.StringValue(defaultPV.Name)] = aws.StringValue(defaultPV.Value)
	}

	current, err := query(ctx, name)
	if err != nil {
		return nil, err
	}

	var result []*memorydb.Parameter

	for _, parameter := range current {
		name := aws.StringValue(parameter.Name)
		currentValue := aws.StringValue(parameter.Value)
		defaultValue := defaultValueByName[name]
		_, isUserDefined := userDefined[name]

		if currentValue != defaultValue || isUserDefined {
			result = append(result, parameter)
		}
	}

	return result, nil
}

// ParameterHash was copy-pasted from ElastiCache.
func ParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrName].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrValue].(string)))

	return create.StringHashcode(buf.String())
}

// ParameterChanges was copy-pasted from ElastiCache.
func ParameterChanges(o, n interface{}) (remove, addOrUpdate []*memorydb.ParameterNameValue) {
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	om := make(map[string]*memorydb.ParameterNameValue, os.Len())
	for _, raw := range os.List() {
		param := raw.(map[string]interface{})
		om[param[names.AttrName].(string)] = expandParameterNameValue(param)
	}
	nm := make(map[string]*memorydb.ParameterNameValue, len(addOrUpdate))
	for _, raw := range ns.List() {
		param := raw.(map[string]interface{})
		nm[param[names.AttrName].(string)] = expandParameterNameValue(param)
	}

	// Remove: key is in old, but not in new
	remove = make([]*memorydb.ParameterNameValue, 0, os.Len())
	for k := range om {
		if _, ok := nm[k]; !ok {
			remove = append(remove, om[k])
		}
	}

	// Add or Update: key is in new, but not in old or has changed value
	addOrUpdate = make([]*memorydb.ParameterNameValue, 0, ns.Len())
	for k, nv := range nm {
		ov, ok := om[k]
		if !ok || ok && (aws.StringValue(nv.ParameterValue) != aws.StringValue(ov.ParameterValue)) {
			addOrUpdate = append(addOrUpdate, nm[k])
		}
	}

	return remove, addOrUpdate
}

func flattenParameters(list []*memorydb.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		if i.Value != nil {
			result = append(result, map[string]interface{}{
				names.AttrName:  strings.ToLower(aws.StringValue(i.Name)),
				names.AttrValue: aws.StringValue(i.Value),
			})
		}
	}
	return result
}

func expandParameterNameValue(param map[string]interface{}) *memorydb.ParameterNameValue {
	return &memorydb.ParameterNameValue{
		ParameterName:  aws.String(param[names.AttrName].(string)),
		ParameterValue: aws.String(param[names.AttrValue].(string)),
	}
}

func createUserDefinedParameterMap(d *schema.ResourceData) map[string]string {
	result := map[string]string{}

	for _, param := range d.Get(names.AttrParameter).(*schema.Set).List() {
		m, ok := param.(map[string]interface{})
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
