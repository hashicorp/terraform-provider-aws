package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	ScramSecretBatchSize = 10
)

func ResourceScramSecretAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScramSecretAssociationCreate,
		ReadWithoutTimeout:   resourceScramSecretAssociationRead,
		UpdateWithoutTimeout: resourceScramSecretAssociationUpdate,
		DeleteWithoutTimeout: resourceScramSecretAssociationDelete,
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

func resourceScramSecretAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	clusterArn := d.Get("cluster_arn").(string)
	secretArnList := flex.ExpandStringSet(d.Get("secret_arn_list").(*schema.Set))

	output, err := associateClusterSecrets(ctx, conn, clusterArn, secretArnList)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating scram secret(s) to MSK cluster (%s): %s", clusterArn, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	if len(output.UnprocessedScramSecrets) != 0 {
		return sdkdiag.AppendErrorf(diags, "associating scram secret(s) to MSK cluster (%s): %s", clusterArn, unprocessedScramSecretsError(output.UnprocessedScramSecrets))
	}

	return append(diags, resourceScramSecretAssociationRead(ctx, d, meta)...)
}

func resourceScramSecretAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	secretArnList, err := FindScramSecrets(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		log.Printf("[WARN] Scram secret(s) for MSK cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK cluster (%s) scram secret(s): %s", d.Id(), err)
	}

	d.Set("cluster_arn", d.Id())
	if err := d.Set("secret_arn_list", flex.FlattenStringSet(secretArnList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting secret_arn_list: %s", err)
	}

	return diags
}

func resourceScramSecretAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	o, n := d.GetChange("secret_arn_list")
	oldSet, newSet := o.(*schema.Set), n.(*schema.Set)

	if newSet.Len() > 0 {
		if newSecrets := newSet.Difference(oldSet); newSecrets.Len() > 0 {
			output, err := associateClusterSecrets(ctx, conn, d.Id(), flex.ExpandStringSet(newSecrets))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating scram secret(s) with MSK cluster (%s): %s", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return sdkdiag.AppendErrorf(diags, "associating scram secret(s) to MSK cluster (%s): %s", d.Id(), unprocessedScramSecretsError(output.UnprocessedScramSecrets))
			}
		}
	}

	if oldSet.Len() > 0 {
		if deleteSecrets := oldSet.Difference(newSet); deleteSecrets.Len() > 0 {
			output, err := disassociateClusterSecrets(ctx, conn, d.Id(), flex.ExpandStringSet(deleteSecrets))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating scram secret(s) from MSK cluster (%s): %s", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return sdkdiag.AppendErrorf(diags, "disassociating scram secret(s) from MSK cluster (%s): %s", d.Id(), unprocessedScramSecretsError(output.UnprocessedScramSecrets))
			}
		}
	}

	return append(diags, resourceScramSecretAssociationRead(ctx, d, meta)...)
}

func resourceScramSecretAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	secretArnList, err := FindScramSecrets(ctx, conn, d.Id())

	if err != nil {
		if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading scram secret(s) for MSK cluster (%s): %s", d.Id(), err)
	}

	if len(secretArnList) > 0 {
		output, err := disassociateClusterSecrets(ctx, conn, d.Id(), secretArnList)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
				return diags
			}
			return sdkdiag.AppendErrorf(diags, "disassociating scram secret(s) from MSK cluster (%s): %s", d.Id(), err)
		}
		if len(output.UnprocessedScramSecrets) != 0 {
			return sdkdiag.AppendErrorf(diags, "disassociating scram secret(s) from MSK cluster (%s): %s", d.Id(), unprocessedScramSecretsError(output.UnprocessedScramSecrets))
		}
	}

	return diags
}

func associateClusterSecrets(ctx context.Context, conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchAssociateScramSecretOutput, error) {
	output := &kafka.BatchAssociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchAssociateScramSecretWithContext(ctx, &kafka.BatchAssociateScramSecretInput{
			ClusterArn:    aws.String(clusterArn),
			SecretArnList: secretArnList[i:end],
		})
		if err != nil {
			return nil, err
		}

		output.ClusterArn = resp.ClusterArn
		output.UnprocessedScramSecrets = append(output.UnprocessedScramSecrets, resp.UnprocessedScramSecrets...)
	}
	return output, nil
}

func disassociateClusterSecrets(ctx context.Context, conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchDisassociateScramSecretOutput, error) {
	output := &kafka.BatchDisassociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchDisassociateScramSecretWithContext(ctx, &kafka.BatchDisassociateScramSecretInput{
			ClusterArn:    aws.String(clusterArn),
			SecretArnList: secretArnList[i:end],
		})
		if err != nil {
			return nil, err
		}

		output.ClusterArn = resp.ClusterArn
		output.UnprocessedScramSecrets = append(output.UnprocessedScramSecrets, resp.UnprocessedScramSecrets...)
	}
	return output, nil
}

func unprocessedScramSecretsError(secrets []*kafka.UnprocessedScramSecret) error {
	var errors *multierror.Error

	for _, s := range secrets {
		secretArn, errMsg := aws.StringValue(s.SecretArn), aws.StringValue(s.ErrorMessage)
		errors = multierror.Append(errors, fmt.Errorf("scram secret (%s): %s", secretArn, errMsg))
	}

	return errors.ErrorOrNil()
}
