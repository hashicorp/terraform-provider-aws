package rds

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validParamGroupNamePrefix,
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "immediate",
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceParameterHash,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateDBParameterGroupInput{
		DBParameterGroupFamily: aws.String(d.Get("family").(string)),
		DBParameterGroupName:   aws.String(name),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   Tags(tags.IgnoreAWS()),
	}

	output, err := conn.CreateDBParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creatingDB Parameter Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DBParameterGroup.DBParameterGroupName))

	// Set for update
	d.Set("arn", output.DBParameterGroup.DBParameterGroupArn)

	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbParameterGroup, err := FindDBParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(dbParameterGroup.DBParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", dbParameterGroup.Description)
	d.Set("family", dbParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbParameterGroup.DBParameterGroupName)

	input := &rds.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	configParams := d.Get("parameter").(*schema.Set)
	if configParams.Len() < 1 {
		// if we don't have any params in the ResourceData already, two possibilities
		// first, we don't have a config available to us. Second, we do, but it has
		// no parameters. We're going to assume the first, to be safe. In this case,
		// we're only going to ask for the user-modified values, because any defaults
		// the user may have _also_ set are indistinguishable from the hundreds of
		// defaults AWS sets. If the user hasn't set any parameters, this will return
		// an empty list anyways, so we just make some unnecessary requests. But in
		// the more common case (I assume) of an import, this will make fewer requests
		// and "do the right thing".
		input.Source = aws.String("user")
	}

	var parameters []*rds.Parameter
	err = conn.DescribeDBParametersPagesWithContext(ctx, input, func(page *rds.DescribeDBParametersOutput, lastPage bool) bool {
		parameters = append(parameters, page.Parameters...)
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	var userParams []*rds.Parameter
	if configParams.Len() < 1 {
		// if we have no config/no parameters in config, we've already asked for only
		// user-modified values, so we can just use the entire response.
		userParams = parameters
	} else {
		// if we have a config available to us, we have two possible classes of value
		// in the config. On the one hand, the user could have specified a parameter
		// that _actually_ changed things, in which case its Source would be set to
		// user. On the other, they may have specified a parameter that coincides with
		// the default value. In that case, the Source will be set to "system" or
		// "engine-default". We need to set the union of all "user" Source parameters
		// _and_ the "system"/"engine-default" Source parameters _that appear in the
		// config_ in the state, or the user gets a perpetual diff. See
		// terraform-providers/terraform-provider-aws#593 for more context and details.
		confParams := expandParameters(configParams.List())
		for _, param := range parameters {
			if param.Source == nil || param.ParameterName == nil {
				continue
			}
			if aws.StringValue(param.Source) == "user" {
				userParams = append(userParams, param)
				continue
			}
			var paramFound bool
			for _, cp := range confParams {
				if cp.ParameterName == nil {
					continue
				}
				if aws.StringValue(cp.ParameterName) == aws.StringValue(param.ParameterName) {
					userParams = append(userParams, param)
					break
				}
			}
			if !paramFound {
				log.Printf("[DEBUG] Not persisting %s to state, as its source is %q and it isn't in the config", aws.StringValue(param.ParameterName), aws.StringValue(param.Source))
			}
		}
	}

	if err := d.Set("parameter", flattenParameters(userParams)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for RDS DB Parameter Group (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		maxParamModifyChunk = 20
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Expand the "parameter" set to aws-sdk-go compat []rds.Parameter
		parameters := expandParameters(ns.Difference(os).List())

		if len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.

			for parameters != nil {
				var paramsToModify []*rds.Parameter
				paramsToModify, parameters = ResourceParameterModifyChunk(parameters, maxParamModifyChunk)

				input := &rds.ModifyDBParameterGroupInput{
					DBParameterGroupName: aws.String(d.Id()),
					Parameters:           paramsToModify,
				}

				_, err := conn.ModifyDBParameterGroupWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying DB Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}

		toRemove := map[string]*rds.Parameter{}

		for _, p := range expandParameters(os.List()) {
			if p.ParameterName != nil {
				toRemove[*p.ParameterName] = p
			}
		}

		for _, p := range expandParameters(ns.List()) {
			if p.ParameterName != nil {
				delete(toRemove, *p.ParameterName)
			}
		}

		// Reset parameters that have been removed
		var resetParameters []*rds.Parameter
		for _, v := range toRemove {
			resetParameters = append(resetParameters, v)
		}
		if len(resetParameters) > 0 {
			for resetParameters != nil {
				var paramsToReset []*rds.Parameter
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

				_, err := conn.ResetDBParameterGroupWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "resetting DB Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Parameter Group (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	conn := meta.(*conns.AWSClient).RDSClient()
	input := &rds_sdkv2.DeleteDBParameterGroupInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting RDS DB Parameter Group: %s", d.Id())
	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBParameterGroup(ctx, input)
		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil
		} else if errs.IsA[*types.InvalidDBParameterGroupStateFault](err) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDBParameterGroup(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Parameter Group (%s): %s", d.Id(), err)
	}
	return nil
}

func FindDBParameterGroupByName(ctx context.Context, conn *rds.RDS, name string) (*rds.DBParameterGroup, error) {
	input := &rds.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeDBParameterGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBParameterGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBParameterGroups) == 0 || output.DBParameterGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbParameterGroup := output.DBParameterGroups[0]

	// Eventual consistency check.
	if aws.StringValue(dbParameterGroup.DBParameterGroupName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return dbParameterGroup, nil
}

func resourceParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["name"].(string))))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["apply_method"].(string))))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))

	// This hash randomly affects the "order" of the set, which affects in what order parameters
	// are applied, when there are more than 20 (chunked).
	return create.StringHashcode(buf.String())
}

func ResourceParameterModifyChunk(all []*rds.Parameter, maxChunkSize int) ([]*rds.Parameter, []*rds.Parameter) {
	// Since the hash randomly affect the set "order," this attempts to prioritize important
	// parameters to go in the first chunk (i.e., charset)

	if len(all) <= maxChunkSize {
		return all[:], nil
	}

	var modifyChunk, remainder []*rds.Parameter

	// pass 1
	for i, p := range all {
		if len(modifyChunk) >= maxChunkSize {
			remainder = append(remainder, all[i:]...)
			return modifyChunk, remainder
		}

		if strings.Contains(aws.StringValue(p.ParameterName), "character_set") && aws.StringValue(p.ApplyMethod) != "pending-reboot" {
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

		if aws.StringValue(p.ApplyMethod) != "pending-reboot" {
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
