package transcribe

import "github.com/aws/aws-sdk-go-v2/service/transcribe/types"

func languageCodeSlice(in []types.LanguageCode) (out []string) {
	for _, v := range in {
		out = append(out, string(v))
	}

	return
}
