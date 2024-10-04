// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_policy", name="Policy")
// @Tags(identifierAttribute="arn")
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	name := d.Get(names.AttrName).(string)
	input := &iot.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
		Tags:           getTagsIn(ctx),
	}

	output, err := conn.CreatePolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.PolicyName))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	output, err := findPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.PolicyArn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set(names.AttrName, output.PolicyName)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &iot.CreatePolicyVersionInput{
			PolicyDocument: aws.String(policy),
			PolicyName:     aws.String(d.Id()),
			SetAsDefault:   true,
		}

		_, errCreate := conn.CreatePolicyVersion(ctx, input)

		// "VersionsLimitExceededException: The policy ... already has the maximum number of versions (5)"
		if errs.IsA[*awstypes.VersionsLimitExceededException](errCreate) {
			// Prune the lowest version and retry.
			policyVersions, err := findPolicyVersionsByName(ctx, conn, d.Id())

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s) versions: %s", d.Id(), err)
			}

			var versionIDs []int

			for _, v := range policyVersions {
				if v.IsDefaultVersion {
					continue
				}

				v, err := strconv.Atoi(aws.ToString(v.VersionId))

				if err != nil {
					continue
				}

				versionIDs = append(versionIDs, v)
			}

			if len(versionIDs) > 0 {
				// Sort ascending.
				slices.Sort(versionIDs)
				versionID := strconv.Itoa(versionIDs[0])

				if err := deletePolicyVersion(ctx, conn, d.Id(), versionID); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				_, errCreate = conn.CreatePolicyVersion(ctx, input)
			}
		}

		if errCreate != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Policy (%s): %s", d.Id(), errCreate)
		}
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	policyVersions, err := findPolicyVersionsByName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s) versions: %s", d.Id(), err)
	}

	// Delete all non-default versions of the policy.
	for _, v := range policyVersions {
		if v.IsDefaultVersion {
			continue
		}

		if err := deletePolicyVersion(ctx, conn, d.Id(), aws.ToString(v.VersionId)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	// Delete default policy version.
	if err := deletePolicy(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func findPolicyByName(ctx context.Context, conn *iot.Client, name string) (*iot.GetPolicyOutput, error) {
	input := &iot.GetPolicyInput{
		PolicyName: aws.String(name),
	}

	output, err := conn.GetPolicy(ctx, input)

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

func findPolicyVersionsByName(ctx context.Context, conn *iot.Client, name string) ([]awstypes.PolicyVersion, error) {
	input := &iot.ListPolicyVersionsInput{
		PolicyName: aws.String(name),
	}

	output, err := conn.ListPolicyVersions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PolicyVersions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PolicyVersions, nil
}

func deletePolicy(ctx context.Context, conn *iot.Client, name string) error {
	input := &iot.DeletePolicyInput{
		PolicyName: aws.String(name),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.DeleteConflictException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.DeletePolicy(ctx, input)
		})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IoT Policy (%s): %w", name, err)
	}

	return nil
}

func deletePolicyVersion(ctx context.Context, conn *iot.Client, name, versionID string) error {
	input := &iot.DeletePolicyVersionInput{
		PolicyName:      aws.String(name),
		PolicyVersionId: aws.String(versionID),
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.DeleteConflictException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.DeletePolicyVersion(ctx, input)
		})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting IoT Policy (%s) version (%s): %w", name, versionID, err)
	}

	return nil
}
