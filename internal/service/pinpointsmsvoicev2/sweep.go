// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_pinpointsmsvoicev2_phone_number", &resource.Sweeper{
		Name:         "aws_pinpointsmsvoicev2_phone_number",
		F:            sweepPhoneNumbers,
		Dependencies: []string{"aws_pinpointsmsvoicev2_pool"},
	})
	resource.AddTestSweepers("aws_pinpointsmsvoicev2_pool", &resource.Sweeper{
		Name: "aws_pinpointsmsvoicev2_pool",
		F:    sweepPools,
	})
	awsv2.Register("aws_pinpointsmsvoicev2_account_default_protect_configuration", sweepAccountDefaultProtectConfigurations)
	awsv2.Register("aws_pinpointsmsvoicev2_protect_configuration", sweepProtectConfigurations, "aws_pinpointsmsvoicev2_account_default_protect_configuration", "aws_pinpointsmsvoicev2_protect_configuration_association")
	awsv2.Register("aws_pinpointsmsvoicev2_protect_configuration_association", sweepProtectConfigurationAssociations)
	resource.AddTestSweepers("aws_pinpointsmsvoicev2_sender_id", &resource.Sweeper{
		Name:         "aws_pinpointsmsvoicev2_sender_id",
		F:            sweepSenderIDs,
		Dependencies: []string{"aws_pinpointsmsvoicev2_pool"},
	})
}

func sweepPhoneNumbers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}
	conn := client.PinpointSMSVoiceV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping End User Messaging SMS Phone Number sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing End User Messaging SMS Phone Numbers (%s): %w", region, err)
		}

		for _, v := range page.PhoneNumbers {
			id := aws.ToString(v.PhoneNumberId)

			if v := v.DeletionProtectionEnabled; v {
				log.Printf("[INFO] Skipping End User Messaging SMS Phone Number %s: DeletionProtectionEnabled=%t", id, v)
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newPhoneNumberResource, client,
				framework.NewAttribute(names.AttrID, id)))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping End User Messaging SMS Phone Numbers (%s): %w", region, err)
	}

	return nil
}

func sweepPools(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &pinpointsmsvoicev2.DescribePoolsInput{}
	conn := client.PinpointSMSVoiceV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := pinpointsmsvoicev2.NewDescribePoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping End User Messaging SMS Pool sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing End User Messaging SMS Pools (%s): %w", region, err)
		}

		for _, v := range page.Pools {
			id := aws.ToString(v.PoolId)

			sweepResources = append(sweepResources, framework.NewSweepResource(newPoolResource, client,
				framework.NewAttribute(names.AttrID, id)))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping End User Messaging SMS Pools (%s): %w", region, err)
	}

	return nil
}

func sweepSenderIDs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &pinpointsmsvoicev2.DescribeSenderIdsInput{}
	conn := client.PinpointSMSVoiceV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := pinpointsmsvoicev2.NewDescribeSenderIdsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping End User Messaging SMS Sender ID sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing End User Messaging SMS Sender IDs (%s): %w", region, err)
		}

		for _, v := range page.SenderIds {
			senderID := aws.ToString(v.SenderId)
			isoCountryCode := aws.ToString(v.IsoCountryCode)

			if v := v.DeletionProtectionEnabled; v {
				log.Printf("[INFO] Skipping End User Messaging SMS Sender ID %s: DeletionProtectionEnabled=%t", senderID, v)
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newSenderIDResource, client,
				framework.NewAttribute("sender_id", senderID),
				framework.NewAttribute("iso_country_code", isoCountryCode)))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping End User Messaging SMS Sender IDs (%s): %w", region, err)
	}

	return nil
}

func sweepAccountDefaultProtectConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.PinpointSMSVoiceV2Client(ctx)

	_, err := findAccountDefaultProtectConfiguration(ctx, conn)
	if retry.NotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return []sweep.Sweepable{
		framework.NewSweepResource(newAccountDefaultProtectConfigurationResource, client,
			framework.NewAttribute(names.AttrID, client.Region(ctx))),
	}, nil
}

func sweepProtectConfigurationAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input pinpointsmsvoicev2.DescribeConfigurationSetsInput
	conn := client.PinpointSMSVoiceV2Client(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listConfigurationSets(ctx, conn, &input) {
		if err != nil {
			return nil, err
		}

		if aws.ToString(v.ProtectConfigurationId) == "" {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newProtectConfigurationAssociationResource, client,
			framework.NewAttribute("configuration_set_name", aws.ToString(v.ConfigurationSetName)),
			framework.NewAttribute("protect_configuration_id", aws.ToString(v.ProtectConfigurationId))))
	}

	return sweepResources, nil
}

func sweepProtectConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input pinpointsmsvoicev2.DescribeProtectConfigurationsInput
	conn := client.PinpointSMSVoiceV2Client(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listProtectConfigurations(ctx, conn, &input) {
		if err != nil {
			return nil, err
		}

		sweepResources = append(sweepResources, protectConfigurationSweeper{
			conn:   conn,
			config: v,
			sweepable: framework.NewSweepResource(newProtectConfigurationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.ProtectConfigurationId))),
		})
	}

	return sweepResources, nil
}

// protectConfigurationSweeper removes the settings that block deletion before deleting the protect configuration.
type protectConfigurationSweeper struct {
	conn      *pinpointsmsvoicev2.Client
	config    awstypes.ProtectConfigurationInformation
	sweepable sweep.Sweepable
}

func (s protectConfigurationSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	id := aws.ToString(s.config.ProtectConfigurationId)

	if s.config.DeletionProtectionEnabled {
		input := pinpointsmsvoicev2.UpdateProtectConfigurationInput{
			DeletionProtectionEnabled: aws.Bool(false),
			ProtectConfigurationId:    aws.String(id),
		}
		if _, err := s.conn.UpdateProtectConfiguration(ctx, &input); err != nil {
			return fmt.Errorf("disabling deletion protection for End User Messaging SMS Protect Configuration (%s): %w", id, err)
		}
	}

	input := pinpointsmsvoicev2.DescribeConfigurationSetsInput{
		Filters: []awstypes.ConfigurationSetFilter{
			{
				Name:   awstypes.ConfigurationSetFilterNameProtectConfigurationId,
				Values: []string{id},
			},
		},
	}
	configurationSets, err := findConfigurationSets(ctx, s.conn, &input)
	if err != nil && !retry.NotFound(err) {
		return fmt.Errorf("listing End User Messaging SMS Configuration Sets for Protect Configuration (%s): %w", id, err)
	}
	for _, v := range configurationSets {
		input := pinpointsmsvoicev2.DisassociateProtectConfigurationInput{
			ConfigurationSetName:   v.ConfigurationSetName,
			ProtectConfigurationId: aws.String(id),
		}
		_, err := s.conn.DisassociateProtectConfiguration(ctx, &input)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || isProtectConfigurationNotAssociatedError(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("disassociating End User Messaging SMS Protect Configuration (%s) from Configuration Set (%s): %w", id, aws.ToString(v.ConfigurationSetName), err)
		}
	}

	if s.config.AccountDefault {
		var input pinpointsmsvoicev2.DeleteAccountDefaultProtectConfigurationInput
		if _, err := s.conn.DeleteAccountDefaultProtectConfiguration(ctx, &input); err != nil {
			return fmt.Errorf("removing End User Messaging SMS Protect Configuration (%s) as account default: %w", id, err)
		}
	}

	return s.sweepable.Delete(ctx, optFns...)
}
