// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	secretVersionStageCurrent  = "AWSCURRENT"
	secretVersionStagePrevious = "AWSPREVIOUS"
)

// @SDKResource("aws_secretsmanager_secret_version", name="Secret Version")
func resourceSecretVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretVersionCreate,
		ReadWithoutTimeout:   resourceSecretVersionRead,
		UpdateWithoutTimeout: resourceSecretVersionUpdate,
		DeleteWithoutTimeout: resourceSecretVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret_binary": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secret_string"},
				ValidateFunc:  verify.ValidBase64String,
			},
			"secret_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secret_binary"},
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_stages": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSecretVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID := d.Get("secret_id").(string)
	input := &secretsmanager.PutSecretValueInput{
		ClientRequestToken: aws.String(id.UniqueId()), // Needed because we're handling our own retries
		SecretId:           aws.String(secretID),
	}

	if v, ok := d.GetOk("secret_binary"); ok {
		var err error
		input.SecretBinary, err = itypes.Base64Decode(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	} else if v, ok := d.GetOk("secret_string"); ok {
		input.SecretString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("version_stages"); ok && v.(*schema.Set).Len() > 0 {
		input.VersionStages = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.PutSecretValue(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Secrets Manager Secret (%s) value: %s", secretID, err)
	}

	versionID := aws.ToString(output.VersionId)
	d.SetId(secretVersionCreateResourceID(secretID, versionID))

	_, err = tfresource.RetryWhenNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return findSecretVersionByTwoPartKey(ctx, conn, secretID, versionID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret Version (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSecretVersionRead(ctx, d, meta)...)
}

func resourceSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID, versionID, err := secretVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findSecretVersionByTwoPartKey(ctx, conn, secretID, versionID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Secrets Manager Secret Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ARN)
	d.Set("secret_binary", itypes.Base64EncodeOnce(output.SecretBinary))
	d.Set("secret_id", secretID)
	d.Set("secret_string", output.SecretString)
	d.Set("version_id", output.VersionId)
	d.Set("version_stages", output.VersionStages)

	return diags
}

func resourceSecretVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID, versionID, err := secretVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	o, n := d.GetChange("version_stages")
	os, ns := o.(*schema.Set), n.(*schema.Set)
	add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

	var listedVersionIDs bool
	for _, stage := range add {
		inputU := &secretsmanager.UpdateSecretVersionStageInput{
			MoveToVersionId: aws.String(versionID),
			SecretId:        aws.String(secretID),
			VersionStage:    aws.String(stage),
		}

		if !listedVersionIDs {
			if stage == secretVersionStageCurrent {
				inputL := &secretsmanager.ListSecretVersionIdsInput{
					SecretId: aws.String(secretID),
				}
				var versionStageCurrentVersionID string

				paginator := secretsmanager.NewListSecretVersionIdsPaginator(conn, inputL)
			listVersionIDs:
				for paginator.HasMorePages() {
					page, err := paginator.NextPage(ctx)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "listing Secrets Manager Secret (%s) version IDs: %s", secretID, err)
					}

					for _, version := range page.Versions {
						for _, versionStage := range version.VersionStages {
							if versionStage == secretVersionStageCurrent {
								versionStageCurrentVersionID = aws.ToString(version.VersionId)
								break listVersionIDs
							}
						}
					}
				}

				inputU.RemoveFromVersionId = aws.String(versionStageCurrentVersionID)
				listedVersionIDs = true
			}
		}

		_, err := conn.UpdateSecretVersionStage(ctx, inputU)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Secrets Manager Secret Version (%s) stage (%s): %s", d.Id(), stage, err)
		}
	}

	for _, stage := range del {
		// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
		if stage == secretVersionStageCurrent {
			log.Printf("[INFO] Skipping removal of AWSCURRENT staging label for secret %q version %q", secretID, versionID)
			continue
		}

		// If we added AWSCURRENT to this version then any AWSPREVIOUS label will have been moved to another version.
		if listedVersionIDs && stage == secretVersionStagePrevious {
			continue
		}

		input := &secretsmanager.UpdateSecretVersionStageInput{
			RemoveFromVersionId: aws.String(versionID),
			SecretId:            aws.String(secretID),
			VersionStage:        aws.String(stage),
		}

		_, err := conn.UpdateSecretVersionStage(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Secrets Manager Secret Version (%s) stage (%s): %s", d.Id(), stage, err)
		}
	}

	return append(diags, resourceSecretVersionRead(ctx, d, meta)...)
}

func resourceSecretVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	secretID, versionID, err := secretVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if v, ok := d.GetOk("version_stages"); ok && v.(*schema.Set).Len() > 0 {
		for _, stage := range flex.ExpandStringValueSet(v.(*schema.Set)) {
			// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
			if stage == secretVersionStageCurrent {
				log.Printf("[WARN] Cannot remove AWSCURRENT staging label, which may leave the secret %q version %q active", secretID, versionID)
				continue
			}

			input := &secretsmanager.UpdateSecretVersionStageInput{
				RemoveFromVersionId: aws.String(versionID),
				SecretId:            aws.String(secretID),
				VersionStage:        aws.String(stage),
			}

			log.Printf("[DEBUG] Deleting Secrets Manager Secret Version (%s) stage: %s", d.Id(), stage)
			_, err := conn.UpdateSecretVersionStage(ctx, input)

			if errs.IsA[*types.ResourceNotFoundException](err) ||
				errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "because it was deleted") ||
				errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "because it was marked for deletion") {
				return diags
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Secrets Manager Secret Version (%s) stage (%s): %s", d.Id(), stage, err)
			}
		}
	}

	_, err = tfresource.RetryUntilNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		output, err := findSecretVersionByTwoPartKey(ctx, conn, secretID, versionID)

		if err != nil {
			return nil, err
		}

		if len(output.VersionStages) == 0 || (len(output.VersionStages) == 1 && (output.VersionStages[0] == secretVersionStageCurrent || output.VersionStages[0] == secretVersionStagePrevious)) {
			return nil, &retry.NotFoundError{}
		}

		return output, nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret Version (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const secretVersionIDSeparator = "|"

func secretVersionCreateResourceID(secretID, versionID string) string {
	parts := []string{secretID, versionID}
	id := strings.Join(parts, secretVersionIDSeparator)

	return id
}

func secretVersionParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, secretVersionIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected SecretID%[2]sVersionID", id, secretVersionIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findSecretVersion(ctx context.Context, conn *secretsmanager.Client, input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	output, err := conn.GetSecretValue(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "because it was deleted") ||
		errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "because it was marked for deletion") {
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

func findSecretVersionByTwoPartKey(ctx context.Context, conn *secretsmanager.Client, secretID, versionID string) (*secretsmanager.GetSecretValueOutput, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:  aws.String(secretID),
		VersionId: aws.String(versionID),
	}

	return findSecretVersion(ctx, conn, input)
}
