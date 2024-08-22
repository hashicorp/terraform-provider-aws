// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_lex_bot_alias", &resource.Sweeper{
		Name: "aws_lex_bot_alias",
		F:    sweepBotAliases,
	})

	resource.AddTestSweepers("aws_lex_bot", &resource.Sweeper{
		Name:         "aws_lex_bot",
		F:            sweepBots,
		Dependencies: []string{"aws_lex_bot_alias"},
	})

	resource.AddTestSweepers("aws_lex_intent", &resource.Sweeper{
		Name:         "aws_lex_intent",
		F:            sweepIntents,
		Dependencies: []string{"aws_lex_bot"},
	})

	resource.AddTestSweepers("aws_lex_slot_type", &resource.Sweeper{
		Name:         "aws_lex_slot_type",
		F:            sweepSlotTypes,
		Dependencies: []string{"aws_lex_intent"},
	})
}

func sweepBotAliases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetBotsInput{}

	pages := lexmodelbuildingservice.NewGetBotsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot Alias for %s: %w", region, err))
		}

		for _, bot := range page.Bots {
			input := &lexmodelbuildingservice.GetBotAliasesInput{
				BotName: bot.Name,
			}

			pages := lexmodelbuildingservice.NewGetBotAliasesPaginator(conn, input)

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot Alias for %s: %w", region, err))
				}

				for _, botAlias := range page.BotAliases {
					r := resourceBotAlias()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s:%s", aws.ToString(bot.Name), aws.ToString(botAlias.Name)))
					d.Set("bot_name", bot.Name)
					d.Set(names.AttrName, botAlias.Name)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}

			r := resourceBotAlias()
			d := r.Data(nil)

			d.SetId(aws.ToString(bot.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Bot Alias for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Bot Alias sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepBots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetBotsInput{}

	pages := lexmodelbuildingservice.NewGetBotsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot for %s: %w", region, err))
		}

		for _, bot := range page.Bots {
			r := resourceBot()
			d := r.Data(nil)

			d.SetId(aws.ToString(bot.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Bot for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Bot sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepIntents(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetIntentsInput{}

	pages := lexmodelbuildingservice.NewGetIntentsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Lex Intent for %s: %w", region, err))
		}

		for _, intent := range page.Intents {
			r := resourceIntent()
			d := r.Data(nil)

			d.SetId(aws.ToString(intent.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Intent for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Intent sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSlotTypes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetSlotTypesInput{}

	pages := lexmodelbuildingservice.NewGetSlotTypesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Lex Slot Type for %s: %w", region, err))
		}

		for _, slotType := range page.SlotTypes {
			r := resourceSlotType()
			d := r.Data(nil)

			d.SetId(aws.ToString(slotType.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Slot Type for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Slot Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
