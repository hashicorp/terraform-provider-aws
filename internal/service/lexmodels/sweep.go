// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
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
	input := &lexmodelbuildingservice.GetBotsInput{}
	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lexmodelbuildingservice.NewGetBotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex Bot Alias sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lex Bots (%s): %w", region, err)
		}

		for _, v := range page.Bots {
			botName := aws.ToString(v.Name)
			input := &lexmodelbuildingservice.GetBotAliasesInput{
				BotName: aws.String(botName),
			}

			pages := lexmodelbuildingservice.NewGetBotAliasesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					log.Printf("[WARN] Skipping Lex Bot Alias sweep for %s: %s", region, err)
					return nil
				}

				if err != nil {
					return fmt.Errorf("error listing Lex Bot Aliases (%s): %w", region, err)
				}

				for _, v := range page.BotAliases {
					botAliasName := aws.ToString(v.Name)
					r := resourceBotAlias()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s:%s", botName, botAliasName))
					d.Set("bot_name", botName)
					d.Set(names.AttrName, botAliasName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lex Bot Aliases (%s): %w", region, err)
	}

	return nil
}

func sweepBots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &lexmodelbuildingservice.GetBotsInput{}
	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lexmodelbuildingservice.NewGetBotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex Bot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lex Bots (%s): %w", region, err)
		}

		for _, v := range page.Bots {
			r := resourceBot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lex Bots (%s): %w", region, err)
	}

	return nil
}

func sweepIntents(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &lexmodelbuildingservice.GetIntentsInput{}
	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lexmodelbuildingservice.NewGetIntentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex Intent sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lex Intents (%s): %w", region, err)
		}

		for _, v := range page.Intents {
			r := resourceIntent()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lex Intents (%s): %w", region, err)
	}

	return nil
}

func sweepSlotTypes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &lexmodelbuildingservice.GetSlotTypesInput{}
	conn := client.LexModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lexmodelbuildingservice.NewGetSlotTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex Slot Type sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lex Slot Types (%s): %w", region, err)
		}

		for _, v := range page.SlotTypes {
			r := resourceSlotType()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lex Slot Types (%s): %w", region, err)
	}

	return nil
}
