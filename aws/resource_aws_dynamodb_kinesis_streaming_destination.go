package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDynamodbKinesisStreamingDestination() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsDynamodbKinesisStreamingDestinationCreate,
		ReadContext:   resourceAwsDynamodbKinesisStreamingDestinationRead,
		DeleteContext: resourceAwsDynamodbKinesisStreamingDestinationDelete,

		Schema: map[string]*schema.Schema{
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stream_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDynamodbKinesisStreamingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	streamArn := d.Get("stream_arn").(string)

	_, err := conn.EnableKinesisStreamingDestination(&dynamodb.EnableKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
		StreamArn: aws.String(streamArn),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error enabling Dynamodb Kinesis streaming destination: %s", err))
	}

	if err := waitForDynamodbKinesisStreamingDestinationToBeActive(ctx, tableName, streamArn, d.Timeout(schema.TimeoutCreate), conn); err != nil {
		return diag.FromErr(fmt.Errorf("failed waiting for Kinesis streaming destination to become active: %s", err))
	}

	d.SetId(createDynamodbKinesisStreamingDestinationResourceId(tableName, streamArn))

	return resourceAwsDynamodbKinesisStreamingDestinationRead(ctx, d, meta)
}

func resourceAwsDynamodbKinesisStreamingDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	streamArn := d.Get("stream_arn").(string)

	result, err := conn.DescribeKinesisStreamingDestination(&dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table (%s) not found, error code (404)", tableName)
			d.SetId("")
			return nil
		}

		return diag.FromErr(fmt.Errorf("error retrieving DynamoDB table item: %s", err))
	}

	for _, destination := range result.KinesisDataStreamDestinations {
		if *destination.StreamArn == streamArn {
			if *destination.DestinationStatus == dynamodb.DestinationStatusActive ||
				*destination.DestinationStatus == dynamodb.DestinationStatusEnabling {
				d.SetId(createDynamodbKinesisStreamingDestinationResourceId(tableName, streamArn))
			} else {
				log.Printf("[WARN] Dynamodb Kinesis streaming destination (%s, %s) not active: %s", tableName, streamArn, *destination.DestinationStatus)
				d.SetId("")
			}
			return nil
		}
	}

	log.Printf("[WARN] Dynamodb Kinesis streaming destination (%s, %s) not found", tableName, streamArn)
	d.SetId("")
	return nil
}

func resourceAwsDynamodbKinesisStreamingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	streamArn := d.Get("stream_arn").(string)

	_, err := conn.DisableKinesisStreamingDestination(&dynamodb.DisableKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
		StreamArn: aws.String(streamArn),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error disabling Dynamodb Kinesis streaming destination: %s", err))
	}

	if err := waitForDynamodbKinesisStreamingDestinationToBeDisabled(ctx, tableName, streamArn, d.Timeout(schema.TimeoutDelete), conn); err != nil {
		return diag.FromErr(fmt.Errorf("failed waiting for Kinesis streaming destination to be disabled: %s", err))
	}

	return nil
}

func createDynamodbKinesisStreamingDestinationResourceId(tableName  string, streamArn string) string {
	return fmt.Sprintf("%s_%s", tableName, streamArn)
}

func waitForDynamodbKinesisStreamingDestinationToBeActive(ctx context.Context, tableName string, streamArn string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusDisabled, dynamodb.DestinationStatusEnabling},
		Target:  []string{dynamodb.DestinationStatusActive},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeKinesisStreamingDestination(&dynamodb.DescribeKinesisStreamingDestinationInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			for _, destination := range result.KinesisDataStreamDestinations {
				if *destination.StreamArn == streamArn {
					return destination, *destination.DestinationStatus, nil
				}
			}
			return 42, "", fmt.Errorf("Kinesis streaming destination %s not found for table %s", streamArn, tableName)
		},
	}
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForDynamodbKinesisStreamingDestinationToBeDisabled(ctx context.Context, tableName string, streamArn string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusActive, dynamodb.DestinationStatusDisabling},
		Target:  []string{dynamodb.DestinationStatusDisabled},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeKinesisStreamingDestination(&dynamodb.DescribeKinesisStreamingDestinationInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			for _, destination := range result.KinesisDataStreamDestinations {
				if *destination.StreamArn == streamArn {
					return destination, *destination.DestinationStatus, nil
				}
			}
			return 42, "", fmt.Errorf("Kinesis streaming destination %s not found for table %s", streamArn, tableName)
		},
	}
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
