// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	layerVersionPermissionResourceIDPartCount = 2
)

// @SDKResource("aws_lambda_layer_version_permission", name="Layer Version Permission")
func resourceLayerVersionPermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLayerVersionPermissionCreate,
		ReadWithoutTimeout:   resourceLayerVersionPermissionRead,
		DeleteWithoutTimeout: resourceLayerVersionPermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"layer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), ""),
					verify.ValidARN,
				),
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrincipal: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"statement_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version_number": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLayerVersionPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName := d.Get("layer_name").(string)
	versionNumber := d.Get("version_number").(int)
	id := errs.Must(flex.FlattenResourceId([]string{layerName, strconv.FormatInt(int64(versionNumber), 10)}, layerVersionPermissionResourceIDPartCount, true))
	input := &lambda.AddLayerVersionPermissionInput{
		Action:        aws.String(d.Get(names.AttrAction).(string)),
		LayerName:     aws.String(layerName),
		Principal:     aws.String(d.Get(names.AttrPrincipal).(string)),
		StatementId:   aws.String(d.Get("statement_id").(string)),
		VersionNumber: aws.Int64(int64(versionNumber)),
	}

	if v, ok := d.GetOk("organization_id"); ok {
		input.OrganizationId = aws.String(v.(string))
	}

	_, err := conn.AddLayerVersionPermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Lambda Layer Version Permission (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceLayerVersionPermissionRead(ctx, d, meta)...)
}

func resourceLayerVersionPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName, versionNumber, err := layerVersionPermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findLayerVersionPolicyByTwoPartKey(ctx, conn, layerName, versionNumber)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Layer Version Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Layer Version Permission (%s): %s", d.Id(), err)
	}

	policyDoc := &IAMPolicyDoc{}
	if err := json.Unmarshal([]byte(aws.ToString(output.Policy)), policyDoc); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("layer_name", layerName)
	d.Set(names.AttrPolicy, output.Policy)
	d.Set("revision_id", output.RevisionId)
	d.Set("version_number", versionNumber)

	if len(policyDoc.Statements) > 0 {
		d.Set("statement_id", policyDoc.Statements[0].Sid)

		if actions := policyDoc.Statements[0].Actions; actions != nil {
			var action string

			if t := reflect.TypeOf(actions); t.String() == "[]string" && len(actions.([]string)) > 0 {
				action = actions.([]string)[0]
			} else if t.String() == "string" {
				action = actions.(string)
			}

			d.Set(names.AttrAction, action)
		}

		if len(policyDoc.Statements[0].Conditions) > 0 {
			if values := policyDoc.Statements[0].Conditions[0].Values; values != nil {
				var organizationID string

				if t := reflect.TypeOf(values); t.String() == "[]string" && len(values.([]string)) > 0 {
					organizationID = values.([]string)[0]
				} else if t.String() == "string" {
					organizationID = values.(string)
				}

				d.Set("organization_id", organizationID)
			}
		}

		if len(policyDoc.Statements[0].Principals) > 0 {
			if identifiers := policyDoc.Statements[0].Principals[0].Identifiers; identifiers != nil {
				var principal string

				if t := reflect.TypeOf(identifiers); t.String() == "[]string" && len(identifiers.([]string)) > 0 && identifiers.([]string)[0] == "*" {
					principal = "*"
				} else if t.String() == "string" {
					policyPrincipalARN, err := arn.Parse(identifiers.(string))
					if err != nil {
						return sdkdiag.AppendFromErr(diags, err)
					}
					principal = policyPrincipalARN.AccountID
				}

				d.Set(names.AttrPrincipal, principal)
			}
		}
	}

	return diags
}

func resourceLayerVersionPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName, versionNumber, err := layerVersionPermissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Lambda Layer Permission Version %q", d.Id())
		return diags
	}

	log.Printf("[INFO] Deleting Lambda Layer Permission Version: %s", d.Id())
	_, err = conn.RemoveLayerVersionPermission(ctx, &lambda.RemoveLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		StatementId:   aws.String(d.Get("statement_id").(string)),
		VersionNumber: aws.Int64(versionNumber),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Layer Version Permission (%s): %s", d.Id(), err)
	}

	return diags
}

func layerVersionPermissionParseResourceID(id string) (string, int64, error) {
	parts, err := flex.ExpandResourceId(id, layerVersionPermissionResourceIDPartCount, true)

	if err != nil {
		return "", 0, err
	}

	layerName := parts[0]
	versionNumber, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		return "", 0, err
	}

	return layerName, versionNumber, nil
}

func findLayerVersionPolicy(ctx context.Context, conn *lambda.Client, input *lambda.GetLayerVersionPolicyInput) (*lambda.GetLayerVersionPolicyOutput, error) {
	output, err := conn.GetLayerVersionPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findLayerVersionPolicyByTwoPartKey(ctx context.Context, conn *lambda.Client, layerName string, versionNumber int64) (*lambda.GetLayerVersionPolicyOutput, error) {
	input := &lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	return findLayerVersionPolicy(ctx, conn, input)
}
