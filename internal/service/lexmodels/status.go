package lexmodels

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	serviceStatusCreated  = "CREATED"
	serviceStatusNotFound = "NOTFOUND"
	serviceStatusUnknown  = "UNKNOWN"
)

func statusBotVersion(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBotVersionByName(conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusSlotType(conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSlotTypeVersionByName(conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, serviceStatusCreated, nil
	}
}

func statusIntent(conn *lexmodelbuildingservice.LexModelBuildingService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetIntentVersions(&lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, serviceStatusNotFound, nil
		}
		if err != nil {
			return nil, serviceStatusUnknown, err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil, serviceStatusNotFound, nil
		}

		return output, serviceStatusCreated, nil
	}
}

func statusBotAlias(conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil, serviceStatusNotFound, nil
		}
		if err != nil {
			return nil, serviceStatusUnknown, err
		}
		if output == nil {
			return nil, serviceStatusNotFound, nil
		}

		return output, serviceStatusCreated, nil
	}
}
