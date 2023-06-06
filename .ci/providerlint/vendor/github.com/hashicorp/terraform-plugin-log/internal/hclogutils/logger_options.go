package hclogutils

import (
	"github.com/hashicorp/go-hclog"
)

// LoggerOptionsCopy will safely copy LoggerOptions. Manually implemented
// to save importing a dependency such as github.com/mitchellh/copystructure.
func LoggerOptionsCopy(src *hclog.LoggerOptions) *hclog.LoggerOptions {
	if src == nil {
		return nil
	}

	return &hclog.LoggerOptions{
		AdditionalLocationOffset: src.AdditionalLocationOffset,
		Color:                    src.Color,
		DisableTime:              src.DisableTime,
		Exclude:                  src.Exclude,
		IncludeLocation:          src.IncludeLocation,
		IndependentLevels:        src.IndependentLevels,
		JSONFormat:               src.JSONFormat,
		Level:                    src.Level,
		Mutex:                    src.Mutex,
		Name:                     src.Name,
		Output:                   src.Output,
		TimeFormat:               src.TimeFormat,
		TimeFn:                   src.TimeFn,
	}
}
