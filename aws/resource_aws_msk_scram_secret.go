package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsMskScramSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskScramSecretCreate,
		Read:   resourceAwsMskScramSecretRead,
		Update: resourceAwsMskScramSecretUpdate,
		Delete: resourceAwsMskScramSecretDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_arn_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"scram_secrets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsMskScramSecretCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	existingSecrets, err := readSecrets(conn, d.Get("cluster_arn").(string))
	if err != nil {
		return fmt.Errorf("failed lookup secrets %s", err)
	}

	createSecrets := filterNewSecrets(expandStringList(d.Get("secret_arn_list").([]interface{})), existingSecrets)

	out, err := associateSecrets(conn, d.Get("cluster_arn").(string), createSecrets)
	if err != nil {
		return fmt.Errorf("error associating credentials with MSK cluster: %s", err)
	}
	d.SetId(d.Get("cluster_arn").(string))

	if len(out.UnprocessedScramSecrets) != 0 {
		return fmt.Errorf("there were unprocessed secrets during association: %s", out.UnprocessedScramSecrets)
	}

	return resourceAwsMskScramSecretRead(d, meta)
}

func resourceAwsMskScramSecretRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	scramSecrets, err := readSecrets(conn, d.Get("cluster_arn").(string))
	if err != nil {
		return fmt.Errorf("failed lookup secrets %s", err)
	}

	allSecrets := filterExistingSecrets(expandStringList(d.Get("secret_arn_list").([]interface{})), scramSecrets)

	d.SetId(d.Get("cluster_arn").(string))
	d.Set("arn", d.Get("cluster_arn").(string))
	d.Set("scram_secrets", aws.StringValueSlice(allSecrets))

	return nil
}

func resourceAwsMskScramSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	existingSecrets := expandStringList(d.Get("scram_secrets").([]interface{}))

	updateSecrets, deleteSecrets := filterSecretsForDeletion(expandStringList(d.Get("secret_arn_list").([]interface{})), existingSecrets)

	out, err := associateSecrets(conn, d.Get("cluster_arn").(string), updateSecrets)
	if err != nil {
		return fmt.Errorf("error associating credentials with MSK cluster: %s", err)
	}

	if len(out.UnprocessedScramSecrets) != 0 {
		return fmt.Errorf("there were unprocessed secrets during association: %s", out.UnprocessedScramSecrets)
	}

	if len(deleteSecrets) > 0 {
		deleteOutput, err := conn.BatchDisassociateScramSecret(&kafka.BatchDisassociateScramSecretInput{
			ClusterArn:    aws.String(d.Get("cluster_arn").(string)),
			SecretArnList: deleteSecrets,
		})
		if err != nil {
			return fmt.Errorf("error disassociating credentials with MSK cluster: %s", err)
		}

		if len(deleteOutput.UnprocessedScramSecrets) != 0 {
			return fmt.Errorf("there were unprocessed secrets during association: %s", out.UnprocessedScramSecrets)
		}
	}

	return resourceAwsMskScramSecretRead(d, meta)
}

func resourceAwsMskScramSecretDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	_, err := conn.BatchDisassociateScramSecret(&kafka.BatchDisassociateScramSecretInput{
		ClusterArn:    aws.String(d.Get("cluster_arn").(string)),
		SecretArnList: expandStringList(d.Get("secret_arn_list").([]interface{})),
	})
	if err != nil {
		return fmt.Errorf("error disassociating credentials with MSK cluster: %s", err)
	}

	return nil
}

func readSecrets(conn *kafka.Kafka, clusterArn string) ([]*string, error) {
	input := &kafka.ListScramSecretsInput{
		ClusterArn: aws.String(clusterArn),
	}

	var scramSecrets []*string
	err := conn.ListScramSecretsPages(input,
		func(page *kafka.ListScramSecretsOutput, lastPage bool) bool {
			scramSecrets = append(scramSecrets, page.SecretArnList...)
			return !lastPage
		})
	if err != nil {
		return nil, err
	}

	return scramSecrets, nil
}

func associateSecrets(conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchAssociateScramSecretOutput, error) {
	batch := 10

	output := &kafka.BatchAssociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += batch {
		end := i + batch
		if end > len(secretArnList) {
			end = len(secretArnList)
		}
		out, err := conn.BatchAssociateScramSecret(&kafka.BatchAssociateScramSecretInput{
			ClusterArn:    aws.String(clusterArn),
			SecretArnList: secretArnList[i:end],
		})
		if err != nil {
			return nil, err
		}
		for _, secret := range out.UnprocessedScramSecrets {
			if secret.ErrorCode != nil {
				output.UnprocessedScramSecrets = append(output.UnprocessedScramSecrets, secret)
			}
		}
	}
	return output, nil
}

func filterExistingSecrets(existingSecrets, newSecrets []*string) []*string {
	finalSecrets := []*string{}
	for _, existingSecret := range existingSecrets {
		if contains(newSecrets, existingSecret) {
			finalSecrets = append(finalSecrets, existingSecret)
		}
	}
	return finalSecrets
}

func filterNewSecrets(existingSecrets, newSecrets []*string) []*string {
	finalSecrets := []*string{}
	for _, existingSecret := range existingSecrets {
		if !contains(newSecrets, existingSecret) {
			finalSecrets = append(finalSecrets, existingSecret)
		}
	}
	return finalSecrets
}

func filterSecretsForDeletion(newSecrets, existingSecrets []*string) ([]*string, []*string) {
	var updateSecrets, deleteSecrets []*string
	for _, existingSecret := range existingSecrets {
		if !contains(newSecrets, existingSecret) {
			deleteSecrets = append(deleteSecrets, existingSecret)
		}
	}
	for _, newSecret := range newSecrets {
		if !contains(existingSecrets, newSecret) {
			updateSecrets = append(updateSecrets, newSecret)
		}
	}
	return updateSecrets, deleteSecrets
}

func contains(slice []*string, val *string) bool {
	for _, item := range slice {
		if *item == *val {
			return true
		}
	}
	return false
}
