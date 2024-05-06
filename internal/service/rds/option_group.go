// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_option_group", name="DB Option Group")
// @Tags(identifierAttribute="arn")
func ResourceOptionGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOptionGroupCreate,
		ReadWithoutTimeout:   resourceOptionGroupRead,
		UpdateWithoutTimeout: resourceOptionGroupUpdate,
		DeleteWithoutTimeout: resourceOptionGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"major_engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validOptionGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validOptionGroupNamePrefix,
			},
			"option": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"option_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"option_settings": {
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
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrVersion: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"option_group_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceOptionGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get("name_prefix").(string))
	input := &rds.CreateOptionGroupInput{
		EngineName:             aws.String(d.Get("engine_name").(string)),
		MajorEngineVersion:     aws.String(d.Get("major_engine_version").(string)),
		OptionGroupDescription: aws.String(d.Get("option_group_description").(string)),
		OptionGroupName:        aws.String(name),
		Tags:                   getTagsIn(ctx),
	}

	_, err := conn.CreateOptionGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Option Group (%s): %s", name, err)
	}

	d.SetId(strings.ToLower(name))

	return append(diags, resourceOptionGroupUpdate(ctx, d, meta)...)
}

func resourceOptionGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	option, err := FindOptionGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Option Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Option Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, option.OptionGroupArn)
	d.Set("engine_name", option.EngineName)
	d.Set("major_engine_version", option.MajorEngineVersion)
	d.Set(names.AttrName, option.OptionGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(option.OptionGroupName)))
	if err := d.Set("option", flattenOptions(option.Options, expandOptionConfiguration(d.Get("option").(*schema.Set).List()))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting option: %s", err)
	}
	d.Set("option_group_description", option.OptionGroupDescription)

	return diags
}

func resourceOptionGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChange("option") {
		o, n := d.GetChange("option")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		optionsToInclude := expandOptionConfiguration(ns.Difference(os).List())
		optionsToIncludeNames := flattenOptionNames(ns.Difference(os).List())
		optionsToRemove := []*string{}
		optionsToRemoveNames := flattenOptionNames(os.Difference(ns).List())

		for _, optionToRemoveName := range optionsToRemoveNames {
			if optionInList(*optionToRemoveName, optionsToIncludeNames) {
				continue
			}
			optionsToRemove = append(optionsToRemove, optionToRemoveName)
		}

		// Ensure there is actually something to update
		// InvalidParameterValue: At least one option must be added, modified, or removed.
		if len(optionsToInclude) > 0 || len(optionsToRemove) > 0 {
			input := &rds.ModifyOptionGroupInput{
				ApplyImmediately: aws.Bool(true),
				OptionGroupName:  aws.String(d.Id()),
			}

			if len(optionsToInclude) > 0 {
				input.OptionsToInclude = optionsToInclude
			}

			if len(optionsToRemove) > 0 {
				input.OptionsToRemove = optionsToRemove
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
				return conn.ModifyOptionGroupWithContext(ctx, input)
			}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying RDS DB Option Group (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceOptionGroupRead(ctx, d, meta)...)
}

func resourceOptionGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Option Group: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteOptionGroupWithContext(ctx, &rds.DeleteOptionGroupInput{
			OptionGroupName: aws.String(d.Id()),
		})
	}, rds.ErrCodeInvalidOptionGroupStateFault)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeOptionGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Option Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindOptionGroupByName(ctx context.Context, conn *rds.RDS, name string) (*rds.OptionGroup, error) {
	input := &rds.DescribeOptionGroupsInput{
		OptionGroupName: aws.String(name),
	}
	output, err := findOptionGroup(ctx, conn, input, tfslices.PredicateTrue[*rds.OptionGroup]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.OptionGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findOptionGroup(ctx context.Context, conn *rds.RDS, input *rds.DescribeOptionGroupsInput, filter tfslices.Predicate[*rds.OptionGroup]) (*rds.OptionGroup, error) {
	output, err := findOptionGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findOptionGroups(ctx context.Context, conn *rds.RDS, input *rds.DescribeOptionGroupsInput, filter tfslices.Predicate[*rds.OptionGroup]) ([]*rds.OptionGroup, error) {
	var output []*rds.OptionGroup

	err := conn.DescribeOptionGroupsPagesWithContext(ctx, input, func(page *rds.DescribeOptionGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OptionGroupsList {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeOptionGroupNotFoundFault) {
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

func optionInList(optionName string, list []*string) bool {
	for _, opt := range list {
		if aws.StringValue(opt) == optionName {
			return true
		}
	}
	return false
}

func flattenOptionNames(configured []interface{}) []*string {
	var optionNames []*string
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})
		optionNames = append(optionNames, aws.String(data["option_name"].(string)))
	}

	return optionNames
}
