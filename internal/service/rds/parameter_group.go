// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_parameter_group", name="DB Parameter Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
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
				ValidateFunc:  validParamGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validParamGroupNamePrefix,
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_method": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ApplyMethodImmediate,
							ValidateDiagFunc: enum.ValidateIgnoreCase[types.ApplyMethod](),
						},
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
				Set: resourceParameterHash,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &rds.CreateDBParameterGroupInput{
		DBParameterGroupFamily: aws.String(d.Get(names.AttrFamily).(string)),
		DBParameterGroupName:   aws.String(name),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                   getTagsInV2(ctx),
	}

	output, err := conn.CreateDBParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Parameter Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DBParameterGroup.DBParameterGroupName))

	// Set for update.
	d.Set(names.AttrARN, output.DBParameterGroup.DBParameterGroupArn)

	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbParameterGroup, err := findDBParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbParameterGroup.DBParameterGroupArn)
	d.Set(names.AttrDescription, dbParameterGroup.Description)
	d.Set(names.AttrFamily, dbParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbParameterGroup.DBParameterGroupName)

	input := &rds.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	configParams := d.Get(names.AttrParameter).(*schema.Set)
	if configParams.Len() < 1 {
		// If we don't have any params in the ResourceData already, two possibilities
		// first, we don't have a config available to us. Second, we do, but it has
		// no parameters. We're going to assume the first, to be safe. In this case,
		// we're only going to ask for the user-modified values, because any defaults
		// the user may have _also_ set are indistinguishable from the hundreds of
		// defaults AWS sets. If the user hasn't set any parameters, this will return
		// an empty list anyways, so we just make some unnecessary requests. But in
		// the more common case (I assume) of an import, this will make fewer requests
		// and "do the right thing".
		input.Source = aws.String(parameterSourceUser)
	}

	parameters, err := findDBParameters(ctx, conn, input, tfslices.PredicateTrue[*types.Parameter]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	var userParams []types.Parameter
	if configParams.Len() < 1 {
		// If we have no config/no parameters in config, we've already asked for only
		// user-modified values, so we can just use the entire response.
		userParams = parameters
	} else {
		// If we have a config available to us, we have two possible classes of value
		// in the config. On the one hand, the user could have specified a parameter
		// that _actually_ changed things, in which case its Source would be set to
		// user. On the other, they may have specified a parameter that coincides with
		// the default value. In that case, the Source will be set to "system" or
		// "engine-default". We need to set the union of all "user" Source parameters
		// _and_ the "system"/"engine-default" Source parameters _that appear in the
		// config_ in the state, or the user gets a perpetual diff. See
		// terraform-providers/terraform-provider-aws#593 for more context and details.
		for _, parameter := range parameters {
			if parameter.Source == nil || parameter.ParameterName == nil {
				continue
			}

			if aws.ToString(parameter.Source) == parameterSourceUser {
				userParams = append(userParams, parameter)
				continue
			}

			var paramFound bool
			for _, cp := range expandParameters(configParams.List()) {
				if cp.ParameterName == nil {
					continue
				}

				if aws.ToString(cp.ParameterName) == aws.ToString(parameter.ParameterName) {
					userParams = append(userParams, parameter)
					paramFound = true
					break
				}
			}
			if !paramFound {
				log.Printf("[DEBUG] Not persisting %s to state, as its source is %q and it isn't in the config", aws.ToString(parameter.ParameterName), aws.ToString(parameter.Source))
			}
		}
	}

	if err := d.Set(names.AttrParameter, flattenParameters(userParams)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		maxParamModifyChunk = 20
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if parameters := expandParameters(ns.Difference(os).List()); len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []types.Parameter
				paramsToModify, parameters = parameterGroupModifyChunk(parameters, maxParamModifyChunk)

				input := &rds.ModifyDBParameterGroupInput{
					DBParameterGroupName: aws.String(d.Id()),
					Parameters:           paramsToModify,
				}

				_, err := conn.ModifyDBParameterGroup(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying RDS DB Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}

		toRemove := map[string]types.Parameter{}

		for _, p := range expandParameters(os.List()) {
			if p.ParameterName != nil {
				toRemove[aws.ToString(p.ParameterName)] = p
			}
		}

		for _, p := range expandParameters(ns.List()) {
			if p.ParameterName != nil {
				delete(toRemove, aws.ToString(p.ParameterName))
			}
		}

		// Reset parameters that have been removed.
		if resetParameters := tfmaps.Values(toRemove); len(resetParameters) > 0 {
			for resetParameters != nil {
				var paramsToReset []types.Parameter
				if len(resetParameters) <= maxParamModifyChunk {
					paramsToReset, resetParameters = resetParameters[:], nil
				} else {
					paramsToReset, resetParameters = resetParameters[:maxParamModifyChunk], resetParameters[maxParamModifyChunk:]
				}

				input := &rds.ResetDBParameterGroupInput{
					DBParameterGroupName: aws.String(d.Id()),
					Parameters:           paramsToReset,
					ResetAllParameters:   aws.Bool(false),
				}

				_, err := conn.ResetDBParameterGroup(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "resetting RDS DB Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Parameter Group: %s", d.Id())
	const (
		timeout = 3 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.InvalidDBParameterGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteDBParameterGroup(ctx, &rds.DeleteDBParameterGroupInput{
			DBParameterGroupName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findDBParameterGroupByName(ctx context.Context, conn *rds.Client, name string) (*types.DBParameterGroup, error) {
	input := &rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(name),
	}
	output, err := findDBParameterGroup(ctx, conn, input, tfslices.PredicateTrue[*types.DBParameterGroup]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBParameterGroup(ctx context.Context, conn *rds.Client, input *rds.DescribeDBParameterGroupsInput, filter tfslices.Predicate[*types.DBParameterGroup]) (*types.DBParameterGroup, error) {
	output, err := findDBParameterGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBParameterGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBParameterGroupsInput, filter tfslices.Predicate[*types.DBParameterGroup]) ([]types.DBParameterGroup, error) {
	var output []types.DBParameterGroup

	pages := rds.NewDescribeDBParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBParameterGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findDBParameters(ctx context.Context, conn *rds.Client, input *rds.DescribeDBParametersInput, filter tfslices.Predicate[*types.Parameter]) ([]types.Parameter, error) {
	var output []types.Parameter

	pages := rds.NewDescribeDBParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Parameters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func resourceParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m[names.AttrName].(string))))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["apply_method"].(string))))
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrValue].(string)))

	// This hash randomly affects the "order" of the set, which affects in what order parameters
	// are applied, when there are more than 20 (chunked).
	return create.StringHashcode(buf.String())
}

func parameterGroupModifyChunk(all []types.Parameter, maxChunkSize int) ([]types.Parameter, []types.Parameter) {
	// Since the hash randomly affect the set "order," this attempts to prioritize important
	// parameters to go in the first chunk (i.e., charset).

	if len(all) <= maxChunkSize {
		return all[:], nil
	}

	var modifyChunk, remainder []types.Parameter

	// pass 1
	for i, p := range all {
		if len(modifyChunk) >= maxChunkSize {
			remainder = append(remainder, all[i:]...)
			return modifyChunk, remainder
		}

		if strings.Contains(aws.ToString(p.ParameterName), "character_set") && p.ApplyMethod != types.ApplyMethodPendingReboot {
			modifyChunk = append(modifyChunk, p)
			continue
		}

		remainder = append(remainder, p)
	}

	all = remainder
	remainder = nil

	// pass 2 - avoid pending reboot
	for i, p := range all {
		if len(modifyChunk) >= maxChunkSize {
			remainder = append(remainder, all[i:]...)
			return modifyChunk, remainder
		}

		if p.ApplyMethod != types.ApplyMethodPendingReboot {
			modifyChunk = append(modifyChunk, p)
			continue
		}

		remainder = append(remainder, p)
	}

	all = remainder
	remainder = nil

	// pass 3 - everything else
	for i, p := range all {
		if len(modifyChunk) >= maxChunkSize {
			remainder = append(remainder, all[i:]...)
			return modifyChunk, remainder
		}

		modifyChunk = append(modifyChunk, p)
	}

	return modifyChunk, remainder
}
