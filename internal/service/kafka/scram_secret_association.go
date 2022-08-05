package kafka

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	AssociatingSecret    = "associating"
	DisassociatingSecret = "disassociating"
	ScramSecretBatchSize = 10
)

func ResourceScramSecretAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceScramSecretAssociationCreate,
		Read:   resourceScramSecretAssociationRead,
		Update: resourceScramSecretAssociationUpdate,
		Delete: resourceScramSecretAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceScramSecretAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	clusterArn := d.Get("cluster_arn").(string)
	secretArnList := flex.ExpandStringSet(d.Get("secret_arn_list").(*schema.Set))

	output, err := associateClusterSecrets(conn, clusterArn, secretArnList)
	if err != nil {
		return fmt.Errorf("error associating scram secret(s) to MSK cluster (%s): %w", clusterArn, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	if len(output.UnprocessedScramSecrets) != 0 {
		return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, AssociatingSecret)
	}

	return resourceScramSecretAssociationRead(d, meta)
}

func resourceScramSecretAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	secretArnList, err := FindScramSecrets(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		log.Printf("[WARN] Scram secret(s) for MSK cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading MSK cluster (%s) scram secret(s): %w", d.Id(), err)
	}

	d.Set("cluster_arn", d.Id())
	if err := d.Set("secret_arn_list", flex.FlattenStringSet(secretArnList)); err != nil {
		return fmt.Errorf("error setting secret_arn_list: %w", err)
	}

	return nil
}

func resourceScramSecretAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	o, n := d.GetChange("secret_arn_list")
	oldSet, newSet := o.(*schema.Set), n.(*schema.Set)

	if newSet.Len() > 0 {
		if newSecrets := newSet.Difference(oldSet); newSecrets.Len() > 0 {
			output, err := associateClusterSecrets(conn, d.Id(), flex.ExpandStringSet(newSecrets))
			if err != nil {
				return fmt.Errorf("error associating scram secret(s) with MSK cluster (%s): %w", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, AssociatingSecret)
			}
		}
	}

	if oldSet.Len() > 0 {
		if deleteSecrets := oldSet.Difference(newSet); deleteSecrets.Len() > 0 {
			output, err := disassociateClusterSecrets(conn, d.Id(), flex.ExpandStringSet(deleteSecrets))
			if err != nil {
				return fmt.Errorf("error disassociating scram secret(s) from MSK cluster (%s): %w", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, DisassociatingSecret)
			}
		}
	}

	return resourceScramSecretAssociationRead(d, meta)
}

func resourceScramSecretAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	secretArnList, err := FindScramSecrets(conn, d.Id())

	if err != nil {
		if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error reading scram secret(s) for MSK cluster (%s): %w", d.Id(), err)
	}

	if len(secretArnList) > 0 {
		output, err := disassociateClusterSecrets(conn, d.Id(), secretArnList)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
				return nil
			}
			return fmt.Errorf("error disassociating scram secret(s) from MSK cluster (%s): %w", d.Id(), err)
		}
		if len(output.UnprocessedScramSecrets) != 0 {
			return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, DisassociatingSecret)
		}
	}

	return nil
}

func associateClusterSecrets(conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchAssociateScramSecretOutput, error) {
	output := &kafka.BatchAssociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchAssociateScramSecret(&kafka.BatchAssociateScramSecretInput{
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

func disassociateClusterSecrets(conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchDisassociateScramSecretOutput, error) {
	output := &kafka.BatchDisassociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchDisassociateScramSecret(&kafka.BatchDisassociateScramSecretInput{
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

func unprocessedScramSecretsError(clusterArn *string, secrets []*kafka.UnprocessedScramSecret, action string) error {
	var errors *multierror.Error

	for _, s := range secrets {
		secretArn, errMsg := aws.StringValue(s.SecretArn), aws.StringValue(s.ErrorMessage)
		errors = multierror.Append(errors, fmt.Errorf("error %s MSK cluster (%s) with scram secret (%s): %s", action, aws.StringValue(clusterArn), secretArn, errMsg))
	}

	return errors.ErrorOrNil()
}
