//go:build sweep
// +build sweep

package datasync

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_datasync_agent", &resource.Sweeper{
		Name: "aws_datasync_agent",
		F:    sweepAgents,
	})

	resource.AddTestSweepers("aws_datasync_location_efs", &resource.Sweeper{
		Name: "aws_datasync_location_efs",
		F:    sweepLocationEFSs,
	})

	resource.AddTestSweepers("aws_datasync_location_fsx_windows_file_system", &resource.Sweeper{
		Name: "aws_datasync_location_fsx_windows_file_system",
		F:    sweepLocationFSxWindows,
	})

	resource.AddTestSweepers("aws_datasync_location_fsx_lustre_file_system", &resource.Sweeper{
		Name: "aws_datasync_location_fsx_lustre_file_system",
		F:    sweepLocationFSxLustres,
	})

	resource.AddTestSweepers("aws_datasync_location_nfs", &resource.Sweeper{
		Name: "aws_datasync_location_nfs",
		F:    sweepLocationNFSs,
	})

	resource.AddTestSweepers("aws_datasync_location_s3", &resource.Sweeper{
		Name: "aws_datasync_location_s3",
		F:    sweepLocationS3s,
	})

	resource.AddTestSweepers("aws_datasync_location_smb", &resource.Sweeper{
		Name: "aws_datasync_location_smb",
		F:    sweepLocationSMBs,
	})

	resource.AddTestSweepers("aws_datasync_location_hdfs", &resource.Sweeper{
		Name: "aws_datasync_location_hdfs",
		F:    sweepLocationHDFSs,
	})

	resource.AddTestSweepers("aws_datasync_task", &resource.Sweeper{
		Name: "aws_datasync_task",
		F:    sweepTasks,
	})
}

func sweepAgents(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListAgentsInput{}
	for {
		output, err := conn.ListAgents(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Agent sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Agents: %s", err)
		}

		if len(output.Agents) == 0 {
			log.Print("[DEBUG] No DataSync Agents to sweep")
			return nil
		}

		for _, agent := range output.Agents {
			name := aws.StringValue(agent.Name)

			log.Printf("[INFO] Deleting DataSync Agent: %s", name)
			input := &datasync.DeleteAgentInput{
				AgentArn: agent.AgentArn,
			}

			_, err := conn.DeleteAgent(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "does not exist") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Agent (%s): %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationEFSs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location EFS sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location EFSs: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location EFSs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "efs://") {
				log.Printf("[INFO] Skipping DataSync Location EFS: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location EFS: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location EFS (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationFSxWindows(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location FSX Windows sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving DataSync Location FSX Windows: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location FSX Windows File System to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "fsxw://") {
				log.Printf("[INFO] Skipping DataSync Location FSX Windows File System: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location FSX Windows File System: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location FSX Windows (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationFSxLustres(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location FSX Lustre sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving DataSync Location FSX Lustre: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location FSX Lustre File System to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "fsxl://") {
				log.Printf("[INFO] Skipping DataSync Location FSX Lustre File System: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location FSX Lustre File System: %s", uri)
			r := ResourceLocationFSxLustreFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(location.LocationArn))
			err = r.Delete(d, client)
			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location Lustre File System (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationNFSs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location Nfs sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location Nfss: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location Nfss to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "nfs://") {
				log.Printf("[INFO] Skipping DataSync Location Nfs: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location Nfs: %s", uri)

			r := ResourceLocationNFS()
			d := r.Data(nil)
			d.SetId(aws.StringValue(location.LocationArn))
			err = r.Delete(d, client)
			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location Nfs (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationS3s(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location S3 sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location S3s: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location S3s to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "s3://") {
				log.Printf("[INFO] Skipping DataSync Location S3: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location S3: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location S3 (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationSMBs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location SMB sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location SMBs: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location SMBs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "smb://") {
				log.Printf("[INFO] Skipping DataSync Location SMB: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location SMB: %s", uri)

			r := ResourceLocationSMB()
			d := r.Data(nil)
			d.SetId(aws.StringValue(location.LocationArn))
			err = r.Delete(d, client)
			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location SMB (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepLocationHDFSs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location HDFS sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location HDFSs: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location HDFSs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "hdfs://") {
				log.Printf("[INFO] Skipping DataSync Location HDFS: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location HDFS: %s", uri)

			r := ResourceLocationHDFS()
			d := r.Data(nil)
			d.SetId(aws.StringValue(location.LocationArn))
			err = r.Delete(d, client)
			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location HDFS (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepTasks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListTasksInput{}
	for {
		output, err := conn.ListTasks(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Task sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Tasks: %w", err)
		}

		if len(output.Tasks) == 0 {
			log.Print("[DEBUG] No DataSync Tasks to sweep")
			return nil
		}

		for _, task := range output.Tasks {
			name := aws.StringValue(task.Name)

			log.Printf("[INFO] Deleting DataSync Task: %s", name)
			input := &datasync.DeleteTaskInput{
				TaskArn: task.TaskArn,
			}

			_, err := conn.DeleteTask(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Task (%s): %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}
