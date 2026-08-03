// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findBotVersionByName(ctx context.Context, conn *lexmodelbuildingservice.Client, name, version string) (*lexmodelbuildingservice.GetBotOutput, error) {
	input := &lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(name),
		VersionOrAlias: aws.String(version),
	}

	output, err := conn.GetBot(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func findSlotTypeVersionByName(ctx context.Context, conn *lexmodelbuildingservice.Client, name, version string) (*lexmodelbuildingservice.GetSlotTypeOutput, error) {
	input := &lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(name),
		Version: aws.String(version),
	}

	output, err := conn.GetSlotType(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// findLatestBotVersionByName returns the latest published version of a bot or $LATEST if the bot has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func findLatestBotVersionByName(ctx context.Context, conn *lexmodelbuildingservice.Client, name string) (string, error) {
	input := &lexmodelbuildingservice.GetBotVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	pages := lexmodelbuildingservice.NewGetBotVersionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return "", err
		}

		for _, bot := range page.Bots {
			version := aws.ToString(bot.Version)

			if version == BotVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}
	}

	if latestVersion == 0 {
		return BotVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}

// findLatestIntentVersionByName returns the latest published version of an intent or $LATEST if the intent has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func findLatestIntentVersionByName(ctx context.Context, conn *lexmodelbuildingservice.Client, name string) (string, error) {
	input := &lexmodelbuildingservice.GetIntentVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	pages := lexmodelbuildingservice.NewGetIntentVersionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return "", err
		}

		for _, intent := range page.Intents {
			version := aws.ToString(intent.Version)

			if version == IntentVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}
	}

	if latestVersion == 0 {
		return IntentVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}

// findLatestSlotTypeVersionByName returns the latest published version of a slot or $LATEST if the slot has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func findLatestSlotTypeVersionByName(ctx context.Context, conn *lexmodelbuildingservice.Client, name string) (string, error) {
	input := &lexmodelbuildingservice.GetSlotTypeVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	pages := lexmodelbuildingservice.NewGetSlotTypeVersionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return "", err
		}

		for _, slot := range page.SlotTypes {
			version := aws.ToString(slot.Version)

			if version == SlotTypeVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}
	}

	if latestVersion == 0 {
		return SlotTypeVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}
