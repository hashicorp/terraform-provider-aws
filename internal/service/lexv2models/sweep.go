// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_lexv2models_bot_alias", &resource.Sweeper{
		Name: "aws_lexv2models_bot_alias",
		F:    sweepBotAliases,
	})

	resource.AddTestSweepers("aws_lexv2models_bot", &resource.Sweeper{
		Name:         "aws_lexv2models_bot",
		F:            sweepBots,
		Dependencies: []string{"aws_lexv2models_bot_alias"},
	})
}

func sweepBots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.LexV2ModelsClient(ctx)
	input := &lexmodelsv2.ListBotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lexmodelsv2.NewListBotsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex V2 Models Bot sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Lex V2 Models Bot: %w", err)
		}

		for _, b := range page.BotSummaries {
			id := aws.ToString(b.BotId)

			log.Printf("[INFO] Deleting Lex V2 Models Bot: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newBotResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lex V2 Models Bots for %s: %w", region, err)
	}

	return nil
}

func sweepBotAliases(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.LexV2ModelsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	botPages := lexmodelsv2.NewListBotsPaginator(conn, &lexmodelsv2.ListBotsInput{})
	for botPages.HasMorePages() {
		botPage, err := botPages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lex V2 Models Bot Alias sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error listing Lex V2 Models Bots: %w", err)
		}

		for _, b := range botPage.BotSummaries {
			botID := aws.ToString(b.BotId)
			aliasPages := lexmodelsv2.NewListBotAliasesPaginator(conn, &lexmodelsv2.ListBotAliasesInput{
				BotId: aws.String(botID),
			})
			for aliasPages.HasMorePages() {
				aliasPage, err := aliasPages.NextPage(ctx)
				if err != nil {
					return fmt.Errorf("error listing Lex V2 Models Bot Aliases for bot %s: %w", botID, err)
				}

				for _, a := range aliasPage.BotAliasSummaries {
					botAliasID := aws.ToString(a.BotAliasId)
					id, _ := intflex.FlattenResourceId([]string{botID, botAliasID}, botAliasResourceIDPartCount, false)

					log.Printf("[INFO] Deleting Lex V2 Models Bot Alias: %s", id)
					sweepResources = append(sweepResources, framework.NewSweepResource(newBotAliasResource, client,
						framework.NewAttribute(names.AttrID, id),
						framework.NewAttribute("bot_id", botID),
						framework.NewAttribute("bot_alias_id", botAliasID),
					))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Lex V2 Models Bot Aliases for %s: %w", region, err)
	}

	return nil
}
