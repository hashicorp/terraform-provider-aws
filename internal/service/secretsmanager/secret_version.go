// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_secretsmanager_secret_version")
func ResourceSecretVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretVersionCreate,
		ReadWithoutTimeout:   resourceSecretVersionRead,
		UpdateWithoutTimeout: resourceSecretVersionUpdate,
		DeleteWithoutTimeout: resourceSecretVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secret_binary"},
			},
			"secret_binary": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Sensitive:     true,
				ConflictsWith: []string{"secret_string"},
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
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.PutSecretValueInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("secret_string"); ok {
		input.SecretString = aws.String(v.(string))
	}

	if v, ok := d.GetOk("secret_binary"); ok {
		vs := []byte(v.(string))

		if !verify.IsBase64Encoded(vs) {
			return sdkdiag.AppendErrorf(diags, "expected base64 in secret_binary")
		}

		var err error
		input.SecretBinary, err = base64.StdEncoding.DecodeString(v.(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "decoding secret binary value: %s", err)
		}
	}

	if v, ok := d.GetOk("version_stages"); ok {
		input.VersionStages = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Putting Secrets Manager Secret %q value", secretID)
	output, err := conn.PutSecretValueWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Secrets Manager Secret value: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", secretID, aws.StringValue(output.VersionId)))

	return append(diags, resourceSecretVersionRead(ctx, d, meta)...)
}

func resourceSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	secretID, versionID, err := DecodeSecretVersionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version (%s): %s", d.Id(), err)
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:  aws.String(secretID),
		VersionId: aws.String(versionID),
	}

	var output *secretsmanager.GetSecretValueOutput

	err = retry.RetryContext(ctx, PropagationTimeout, func() *retry.RetryError {
		var err error

		output, err = conn.GetSecretValueWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
			return retry.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetSecretValueWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Secrets Manager Secret Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
		log.Printf("[WARN] Secrets Manager Secret Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Version (%s): empty response", d.Id())
	}

	d.Set("secret_id", secretID)
	d.Set("secret_string", output.SecretString)
	d.Set("secret_binary", verify.Base64Encode(output.SecretBinary))
	d.Set("version_id", output.VersionId)
	d.Set("arn", output.ARN)

	if err := d.Set("version_stages", flex.FlattenStringList(output.VersionStages)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting version_stages: %s", err)
	}

	return diags
}

func resourceSecretVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	secretID, versionID, err := DecodeSecretVersionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret Version (%s): %s", d.Id(), err)
	}

	o, n := d.GetChange("version_stages")
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	stagesToAdd := ns.Difference(os).List()
	stagesToRemove := os.Difference(ns).List()

	for _, stage := range stagesToAdd {
		input := &secretsmanager.UpdateSecretVersionStageInput{
			MoveToVersionId: aws.String(versionID),
			SecretId:        aws.String(secretID),
			VersionStage:    aws.String(stage.(string)),
		}

		log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
		_, err := conn.UpdateSecretVersionStageWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
		}
	}

	for _, stage := range stagesToRemove {
		// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
		if stage.(string) == "AWSCURRENT" {
			log.Printf("[INFO] Skipping removal of AWSCURRENT staging label for secret %q version %q", secretID, versionID)
			continue
		}
		input := &secretsmanager.UpdateSecretVersionStageInput{
			RemoveFromVersionId: aws.String(versionID),
			SecretId:            aws.String(secretID),
			VersionStage:        aws.String(stage.(string)),
		}
		log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
		_, err := conn.UpdateSecretVersionStageWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
		}
	}

	return append(diags, resourceSecretVersionRead(ctx, d, meta)...)
}

func resourceSecretVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	secretID, versionID, err := DecodeSecretVersionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Secrets Manager Secret Version (%s): %s", d.Id(), err)
	}

	if v, ok := d.GetOk("version_stages"); ok {
		for _, stage := range v.(*schema.Set).List() {
			// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
			if stage.(string) == "AWSCURRENT" {
				log.Printf("[WARN] Cannot remove AWSCURRENT staging label, which may leave the secret %q version %q active", secretID, versionID)
				continue
			}
			input := &secretsmanager.UpdateSecretVersionStageInput{
				RemoveFromVersionId: aws.String(versionID),
				SecretId:            aws.String(secretID),
				VersionStage:        aws.String(stage.(string)),
			}
			log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
			_, err := conn.UpdateSecretVersionStageWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
					return diags
				}
				if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
					return diags
				}
				return sdkdiag.AppendErrorf(diags, "updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
			}
		}
	}

	return diags
}

func DecodeSecretVersionID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format SecretID|VersionID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
