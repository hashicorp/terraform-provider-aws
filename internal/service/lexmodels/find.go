package lexmodels

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBotVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) (*lexmodelbuildingservice.GetBotOutput, error) {
	input := &lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(name),
		VersionOrAlias: aws.String(version),
	}

	output, err := conn.GetBot(input)

	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindSlotTypeVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) (*lexmodelbuildingservice.GetSlotTypeOutput, error) {
	input := &lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(name),
		Version: aws.String(version),
	}

	output, err := conn.GetSlotType(input)

	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindLatestBotVersionByName returns the latest published version of a bot or $LATEST if the bot has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func FindLatestBotVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name string) (string, error) {
	input := &lexmodelbuildingservice.GetBotVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	err := conn.GetBotVersionsPages(input, func(page *lexmodelbuildingservice.GetBotVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, bot := range page.Bots {
			version := aws.StringValue(bot.Version)

			if version == BotVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if latestVersion == 0 {
		return BotVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}

// FindLatestIntentVersionByName returns the latest published version of an intent or $LATEST if the intent has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func FindLatestIntentVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name string) (string, error) {
	input := &lexmodelbuildingservice.GetIntentVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	err := conn.GetIntentVersionsPages(input, func(page *lexmodelbuildingservice.GetIntentVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, intent := range page.Intents {
			version := aws.StringValue(intent.Version)

			if version == IntentVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if latestVersion == 0 {
		return IntentVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}

// FindLatestSlotTypeVersionByName returns the latest published version of a slot or $LATEST if the slot has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func FindLatestSlotTypeVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name string) (string, error) {
	input := &lexmodelbuildingservice.GetSlotTypeVersionsInput{
		Name: aws.String(name),
	}
	var latestVersion int

	err := conn.GetSlotTypeVersionsPages(input, func(page *lexmodelbuildingservice.GetSlotTypeVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, slot := range page.SlotTypes {
			version := aws.StringValue(slot.Version)

			if version == SlotTypeVersionLatest {
				continue
			}

			if version, err := strconv.Atoi(version); err != nil {
				continue
			} else if version > latestVersion {
				latestVersion = version
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if latestVersion == 0 {
		return SlotTypeVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}
