// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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

	conn := client.LexModelsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetBotsInput{}

	err = conn.GetBotsPagesWithContext(ctx, input, func(page *lexmodelbuildingservice.GetBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, bot := range page.Bots {
			input := &lexmodelbuildingservice.GetBotAliasesInput{
				BotName: bot.Name,
			}

			err := conn.GetBotAliasesPagesWithContext(ctx, input, func(page *lexmodelbuildingservice.GetBotAliasesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, botAlias := range page.BotAliases {
					r := ResourceBotAlias()
					d := r.Data(nil)

					d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(bot.Name), aws.StringValue(botAlias.Name)))
					d.Set("bot_name", bot.Name)
					d.Set(names.AttrName, botAlias.Name)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot Alias for %s: %w", region, err))
			}

			r := ResourceBotAlias()
			d := r.Data(nil)

			d.SetId(aws.StringValue(bot.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot Alias for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Bot Alias for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
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

	conn := client.LexModelsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetBotsInput{}

	err = conn.GetBotsPagesWithContext(ctx, input, func(page *lexmodelbuildingservice.GetBotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, bot := range page.Bots {
			r := ResourceBot()
			d := r.Data(nil)

			d.SetId(aws.StringValue(bot.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Lex Bot for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Bot for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
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

	conn := client.LexModelsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetIntentsInput{}

	err = conn.GetIntentsPagesWithContext(ctx, input, func(page *lexmodelbuildingservice.GetIntentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, intent := range page.Intents {
			r := ResourceIntent()
			d := r.Data(nil)

			d.SetId(aws.StringValue(intent.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Lex Intent for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Intent for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
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

	conn := client.LexModelsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &lexmodelbuildingservice.GetSlotTypesInput{}

	err = conn.GetSlotTypesPagesWithContext(ctx, input, func(page *lexmodelbuildingservice.GetSlotTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, slotType := range page.SlotTypes {
			r := ResourceSlotType()
			d := r.Data(nil)

			d.SetId(aws.StringValue(slotType.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Lex Slot Type for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Lex Slot Type for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Lex Slot Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
