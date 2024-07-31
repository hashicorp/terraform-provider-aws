// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	scramSecretBatchSize = 10
)

// @SDKResource("aws_msk_scram_secret_association", name="SCRAM Secret Association)
func resourceSCRAMSecretAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSCRAMSecretAssociationCreate,
		ReadWithoutTimeout:   resourceSCRAMSecretAssociationRead,
		UpdateWithoutTimeout: resourceSCRAMSecretAssociationUpdate,
		DeleteWithoutTimeout: resourceSCRAMSecretAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"secret_arn_list": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
		},
	}
}

func resourceSCRAMSecretAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterARN := d.Get("cluster_arn").(string)

	if err := associateSRAMSecrets(ctx, conn, clusterARN, flex.ExpandStringValueSet(d.Get("secret_arn_list").(*schema.Set))); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK SCRAM Secret Association (%s): %s", clusterARN, err)
	}

	d.SetId(clusterARN)

	return append(diags, resourceSCRAMSecretAssociationRead(ctx, d, meta)...)
}

func resourceSCRAMSecretAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	scramSecrets, err := findSCRAMSecretsByClusterARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK SCRAM Secret Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK SCRAM Secret Association (%s): %s", d.Id(), err)
	}

	d.Set("cluster_arn", d.Id())
	d.Set("secret_arn_list", scramSecrets)

	return diags
}

func resourceSCRAMSecretAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	o, n := d.GetChange("secret_arn_list")
	os, ns := o.(*schema.Set), n.(*schema.Set)

	if add := flex.ExpandStringValueSet(ns.Difference(os)); len(add) > 0 {
		if err := associateSRAMSecrets(ctx, conn, d.Id(), add); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK SCRAM Secret Association (%s): %s", d.Id(), err)
		}
	}

	if del := flex.ExpandStringValueSet(os.Difference(ns)); len(del) > 0 {
		if err := disassociateSRAMSecrets(ctx, conn, d.Id(), del); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK SCRAM Secret Association (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSCRAMSecretAssociationRead(ctx, d, meta)...)
}

func resourceSCRAMSecretAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	err := disassociateSRAMSecrets(ctx, conn, d.Id(), flex.ExpandStringValueSet(d.Get("secret_arn_list").(*schema.Set)))

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK SCRAM Secret Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findSCRAMSecretsByClusterARN(ctx context.Context, conn *kafka.Client, clusterARN string) ([]string, error) {
	input := &kafka.ListScramSecretsInput{
		ClusterArn: aws.String(clusterARN),
	}
	var output []string

	pages := kafka.NewListScramSecretsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SecretArnList...)
	}

	return output, nil
}

func associateSRAMSecrets(ctx context.Context, conn *kafka.Client, clusterARN string, secretARNs []string) error {
	for _, chunk := range tfslices.Chunks(secretARNs, scramSecretBatchSize) {
		input := &kafka.BatchAssociateScramSecretInput{
			ClusterArn:    aws.String(clusterARN),
			SecretArnList: chunk,
		}

		output, err := conn.BatchAssociateScramSecret(ctx, input)

		if err == nil {
			err = unprocessedScramSecretsError(output.UnprocessedScramSecrets, false)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func disassociateSRAMSecrets(ctx context.Context, conn *kafka.Client, clusterARN string, secretARNs []string) error {
	for _, chunk := range tfslices.Chunks(secretARNs, scramSecretBatchSize) {
		input := &kafka.BatchDisassociateScramSecretInput{
			ClusterArn:    aws.String(clusterARN),
			SecretArnList: chunk,
		}

		output, err := conn.BatchDisassociateScramSecret(ctx, input)

		if err == nil {
			err = unprocessedScramSecretsError(output.UnprocessedScramSecrets, true)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func unprocessedScramSecretsError(apiObjects []types.UnprocessedScramSecret, ignoreInvalidSecretARN bool) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if ignoreInvalidSecretARN && aws.ToString(apiObject.ErrorCode) == "InvalidSecretArn" {
			continue
		}

		err := unprocessedScramSecretError(&apiObject)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.ToString(apiObject.SecretArn), err))
		}
	}

	return errors.Join(errs...)
}

func unprocessedScramSecretError(apiObject *types.UnprocessedScramSecret) error {
	return fmt.Errorf("%s: %s", aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))
}
