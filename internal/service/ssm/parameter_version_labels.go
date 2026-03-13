// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ssm_parameter_version_labels", name="Parameter Version Labels", supportsTags=false)
func resourceParameterVersionLabels() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterVersionLabelsCreate,
		ReadWithoutTimeout:   resourceParameterVersionLabelsRead,
		UpdateWithoutTimeout: resourceParameterVersionLabelsUpdate,
		DeleteWithoutTimeout: resourceParameterVersionLabelsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceParameterVersionLabelsImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateDiagFunc: validation.AllDiag(
						// A label can have a maximum of 100 characters.
						validation.ToDiagFunc(validation.StringLenBetween(1, 100)),
						// Labels can contain letters (case sensitive), numbers, periods (.), hyphens (-), or underscores (_).
						// Labels can't begin with a number.
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[a-zA-Z_][a-zA-Z0-9._-]*$`), "must begin with a letter or underscore and contain only letters, numbers, periods (.), hyphens (-), or underscores (_)")),
						// Labels can't begin with " aws " or " ssm " (not case sensitive).
						func(v any, p cty.Path) diag.Diagnostics {
							if str, ok := v.(string); ok {
								if strings.HasPrefix(str, "aws") || strings.HasPrefix(str, "ssm") {
									return diag.Diagnostics{diag.Diagnostic{
										Severity:      diag.Error,
										Summary:       "Invalid label",
										Detail:        "Labels cannot start with 'aws' or 'ssm'",
										AttributePath: p,
									}}
								}
							}
							return nil
						},
					),
				},
			},
		},
	}
}

func resourceParameterVersionLabelsCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	version := d.Get(names.AttrVersion).(int)
	name := d.Get(names.AttrName).(string)
	labels := d.Get("labels").([]any)
	// we do not have parameter version, getting the latest
	if version == 0 {
		input := &ssm.GetParameterInput{
			Name: aws.String(name),
		}
		output, err := conn.GetParameter(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: %s", name, err)
		}
		if output.Parameter != nil {
			version = int(output.Parameter.Version)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: parameter not found", name)
		}
	}
	input := &ssm.LabelParameterVersionInput{
		Name:             aws.String(name),
		Labels:           tfslices.ApplyToAll(labels, func(l any) string { return l.(string) }),
		ParameterVersion: aws.Int64(int64(version)),
	}
	output, err := conn.LabelParameterVersion(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "labeling SSM Parameter (%s) version (%d) labels (%v): %s", name, version, labels, err)
	}
	version = int(output.ParameterVersion)
	d.SetId(fmt.Sprintf("%s:%d", name, version))
	return diags
}

func resourceParameterVersionLabelsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	version := d.Get(names.AttrVersion).(int)
	name := d.Get(names.AttrName).(string)
	labels, err := findParameterVersionLabels(ctx, conn, name, version)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("labels", labels)
	d.SetId(fmt.Sprintf("%s:%d", name, version))
	return diags
}

func resourceParameterVersionLabelsUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	version := d.Get(names.AttrVersion).(int)
	name := d.Get(names.AttrName).(string)
	labels := d.Get("labels").([]any)
	// we do not have parameter version, getting the latest
	if version == 0 {
		input := &ssm.GetParameterInput{
			Name: aws.String(name),
		}
		output, err := conn.GetParameter(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: %s", name, err)
		}
		if output.Parameter != nil {
			version = int(output.Parameter.Version)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: parameter not found", name)
		}
	}
	// get existing labels for given parameter version
	existingLabels, err := findParameterVersionLabels(ctx, conn, name, version)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	// determine which labels to remove
	var labelsToRemove []string
	for _, el := range existingLabels {
		found := false
		for _, l := range labels {
			if el == l.(string) {
				found = true
				break
			}
		}
		if !found {
			labelsToRemove = append(labelsToRemove, el)
		}
	}
	if len(labelsToRemove) > 0 {
		input := &ssm.UnlabelParameterVersionInput{
			Name:             aws.String(name),
			ParameterVersion: aws.Int64(int64(version)),
			Labels:           labelsToRemove,
		}
		_, err := conn.UnlabelParameterVersion(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unlabeling SSM Parameter (%s) version (%d) labels (%v): %s", name, version, labelsToRemove, err)
		}
	}
	input := &ssm.LabelParameterVersionInput{
		Name:             aws.String(name),
		Labels:           tfslices.ApplyToAll(labels, func(l any) string { return l.(string) }),
		ParameterVersion: aws.Int64(int64(version)),
	}
	output, err := conn.LabelParameterVersion(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "labeling SSM Parameter (%s) version (%d) labels (%v): %s", name, version, labels, err)
	}
	version = int(output.ParameterVersion)
	d.SetId(fmt.Sprintf("%s:%d", name, version))
	return diags
}

func resourceParameterVersionLabelsDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	version := d.Get(names.AttrVersion).(int)
	name := d.Get(names.AttrName).(string)
	labels := d.Get("labels").([]any)
	// we do not have parameter version, getting the latest
	if version == 0 {
		input := &ssm.GetParameterInput{
			Name: aws.String(name),
		}
		output, err := conn.GetParameter(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: %s", name, err)
		}
		if output.Parameter != nil {
			version = int(output.Parameter.Version)
		} else {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameter (%s) latest version: parameter not found", name)
		}
	}
	input := &ssm.UnlabelParameterVersionInput{
		Name:             aws.String(name),
		ParameterVersion: aws.Int64(int64(version)),
		Labels:           tfslices.ApplyToAll(labels, func(l any) string { return l.(string) }),
	}
	_, err := conn.UnlabelParameterVersion(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unlabeling SSM Parameter (%s) version (%d) labels (%v): %s", name, version, labels, err)
	}
	return diags
}

func findParameterVersionLabels(ctx context.Context, conn *ssm.Client, name string, version int) ([]string, error) {
	// we do not have parameter version, getting the latest
	if version == 0 {
		input := &ssm.GetParameterInput{
			Name: aws.String(name),
		}
		output, err := conn.GetParameter(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("reading SSM Parameter (%s) latest version: %w", name, err)
		}
		if output.Parameter != nil {
			version = int(output.Parameter.Version)
		} else {
			return nil, fmt.Errorf("reading SSM Parameter (%s) latest version: parameter not found", name)
		}
	}
	input := &ssm.GetParameterHistoryInput{
		Name:       aws.String(name),
		MaxResults: aws.Int32(10),
	}
	pages := ssm.NewGetParameterHistoryPaginator(conn, input)
	found := false
	var labels []string
	for pages.HasMorePages() && !found {
		output, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.ParameterNotFound](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		if err != nil {
			return nil, fmt.Errorf("reading SSM Parameter (%s) labels: %w", name, err)
		}
		for _, param := range output.Parameters {
			if param.Version != int64(version) {
				continue
			}
			labels = append(labels, param.Labels...)
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("reading SSM Parameter (%s) labels: version %d not found", name, version)
	}
	return labels, nil
}

func parameterVersionLabelsParseResourceID(id string) (string, int, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("unexpected format of ID (%s), expected name:version", id)
	}
	version := 0
	_, err := fmt.Sscanf(parts[1], "%d", &version)
	if err != nil {
		return "", 0, fmt.Errorf("unexpected format of version in ID (%s), expected integer: %w", id, err)
	}
	return parts[0], version, nil
}

func resourceParameterVersionLabelsImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	name, version, err := parameterVersionLabelsParseResourceID(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	conn := meta.(*conns.AWSClient).SSMClient(ctx)
	labels, err := findParameterVersionLabels(ctx, conn, name, version)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.Set(names.AttrName, name)
	d.Set(names.AttrVersion, version)
	d.Set("labels", labels)
	return []*schema.ResourceData{d}, nil
}
