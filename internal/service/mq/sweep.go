//go:build sweep
// +build sweep

package mq

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_mq_broker", &resource.Sweeper{
		Name: "aws_mq_broker",
		F:    sweepBrokers,
	})
}

func sweepBrokers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).MQConn

	resp, err := conn.ListBrokers(&mq.ListBrokersInput{
		MaxResults: aws.Int64(100),
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MQ Broker sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing MQ brokers: %s", err)
	}

	if len(resp.BrokerSummaries) == 0 {
		log.Print("[DEBUG] No MQ brokers found to sweep")
		return nil
	}
	log.Printf("[DEBUG] %d MQ brokers found", len(resp.BrokerSummaries))

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(resp.BrokerSummaries), func(i, j int) {
		resp.BrokerSummaries[i], resp.BrokerSummaries[j] = resp.BrokerSummaries[j], resp.BrokerSummaries[i]
	})

	for _, bs := range resp.BrokerSummaries {
		log.Printf("[INFO] Deleting MQ broker %s", aws.StringValue(bs.BrokerId))
		_, err := conn.DeleteBroker(&mq.DeleteBrokerInput{
			BrokerId: bs.BrokerId,
		})
		if tfawserr.ErrMessageContains(err, mq.ErrCodeBadRequestException, "while in state [CREATION_IN_PROGRESS") {
			log.Printf("[WARN] Broker in state CREATION_IN_PROGRESS and must complete creation before deletion")
			if _, err = WaitBrokerCreated(conn, aws.StringValue(bs.BrokerId)); err != nil {
				return err
			}

			log.Printf("[WARN] Retrying deletion of broker %s", aws.StringValue(bs.BrokerId))
			_, err = conn.DeleteBroker(&mq.DeleteBrokerInput{
				BrokerId: bs.BrokerId,
			})
		}
		if err != nil {
			return err
		}
		if _, err = WaitBrokerDeleted(conn, aws.StringValue(bs.BrokerId)); err != nil {
			return err
		}
	}

	return nil
}
