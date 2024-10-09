// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"
	"encoding/base64"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_mq_configuration", name="Configuration")
// @Tags(identifierAttribute="arn")
func resourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationCreate,
		ReadWithoutTimeout:   resourceConfigurationRead,
		UpdateWithoutTimeout: resourceConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext, // Delete is not available in the API

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if diff.HasChange(names.AttrDescription) {
					return diff.SetNewComputed("latest_revision")
				}
				if diff.HasChange("data") {
					o, n := diff.GetChange("data")
					os := o.(string)
					ns := n.(string)
					if !suppressXMLEquivalentConfig("data", os, ns, nil) {
						return diff.SetNewComputed("latest_revision")
					}
				}
				return nil
			},
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[types.AuthenticationStrategy](),
			},
			"data": {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      suppressXMLEquivalentConfig,
				DiffSuppressOnRefresh: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[types.EngineType](),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MQClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &mq.CreateConfigurationInput{
		EngineType:    types.EngineType(d.Get("engine_type").(string)),
		EngineVersion: aws.String(d.Get(names.AttrEngineVersion).(string)),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("authentication_strategy"); ok {
		input.AuthenticationStrategy = types.AuthenticationStrategy(v.(string))
	}

	output, err := conn.CreateConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MQ Configuration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	if v, ok := d.GetOk("data"); ok {
		input := &mq.UpdateConfigurationInput{
			ConfigurationId: aws.String(d.Id()),
			Data:            flex.StringValueToBase64String(v.(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MQ Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MQClient(ctx)

	configuration, err := findConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MQ Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MQ Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, configuration.Arn)
	d.Set("authentication_strategy", configuration.AuthenticationStrategy)
	d.Set(names.AttrDescription, configuration.LatestRevision.Description)
	d.Set("engine_type", configuration.EngineType)
	d.Set(names.AttrEngineVersion, configuration.EngineVersion)
	d.Set("latest_revision", configuration.LatestRevision.Revision)
	d.Set(names.AttrName, configuration.Name)

	revision := strconv.FormatInt(int64(aws.ToInt32(configuration.LatestRevision.Revision)), 10)
	configurationRevision, err := conn.DescribeConfigurationRevision(ctx, &mq.DescribeConfigurationRevisionInput{
		ConfigurationId:       aws.String(d.Id()),
		ConfigurationRevision: aws.String(revision),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MQ Configuration (%s) revision (%s): %s", d.Id(), revision, err)
	}

	data, err := base64.StdEncoding.DecodeString(aws.ToString(configurationRevision.Data))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "base64 decoding: %s", err)
	}

	d.Set("data", string(data))

	setTagsOut(ctx, configuration.Tags)

	return diags
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MQClient(ctx)

	if d.HasChanges("data", names.AttrDescription) {
		input := &mq.UpdateConfigurationInput{
			ConfigurationId: aws.String(d.Id()),
			Data:            aws.String(base64.StdEncoding.EncodeToString([]byte(d.Get("data").(string)))),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MQ Configuration (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationRead(ctx, d, meta)...)
}

func findConfigurationByID(ctx context.Context, conn *mq.Client, id string) (*mq.DescribeConfigurationOutput, error) {
	input := &mq.DescribeConfigurationInput{
		ConfigurationId: aws.String(id),
	}

	output, err := conn.DescribeConfiguration(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func suppressXMLEquivalentConfig(k, old, new string, d *schema.ResourceData) bool {
	os, err := CanonicalXML(old)
	if err != nil {
		log.Printf("[ERR] Error getting cannonicalXML from state (%s): %s", k, err)
		return false
	}
	ns, err := CanonicalXML(new)
	if err != nil {
		log.Printf("[ERR] Error getting cannonicalXML from config (%s): %s", k, err)
		return false
	}

	return os == ns
}
