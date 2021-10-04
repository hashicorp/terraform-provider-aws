package finder

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tflex "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lex"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func BotVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) (*lexmodelbuildingservice.GetBotOutput, error) {
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

// LatestBotVersionByName returns the latest published version of a bot or $LATEST if the bot has never been published.
// See https://docs.aws.amazon.com/lex/latest/dg/versioning-aliases.html.
func LatestBotVersionByName(conn *lexmodelbuildingservice.LexModelBuildingService, name string) (string, error) {
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

			if version == tflex.LexBotVersionLatest {
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
		return tflex.LexBotVersionLatest, nil
	}

	return strconv.Itoa(latestVersion), nil
}
