package logging

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// fileExtEmpty is the file extension for empty data.
	// Empty data may be expected, depending on the RPC.
	fileExtEmpty = "empty"

	// fileExtJson is the file extension for JSON data.
	fileExtJson = "json"

	// fileExtMsgpack is the file extension for MessagePack data.
	fileExtMsgpack = "msgpack"
)

var protocolDataSkippedLog sync.Once

// ProtocolData emits raw protocol data to a file, if given a directory.
//
// The directory must exist and be writable, prior to invoking this function.
//
// File names are in the format: {TIME}_{RPC}_{MESSAGE}_{FIELD}.{EXT}
func ProtocolData(ctx context.Context, dataDir string, rpc string, message string, field string, data interface{}) {
	if dataDir == "" {
		// Write a log, only once, that explains how to enable this functionality.
		protocolDataSkippedLog.Do(func() {
			ProtocolTrace(ctx, "Skipping protocol data file writing because no data directory is set. "+
				fmt.Sprintf("Use the %s environment variable to enable this functionality.", EnvTfLogSdkProtoDataDir))
		})

		return
	}

	var fileContents []byte
	var fileExtension string

	switch data := data.(type) {
	case *tfprotov5.DynamicValue:
		fileExtension, fileContents = protocolDataDynamicValue5(ctx, data)
	case *tfprotov6.DynamicValue:
		fileExtension, fileContents = protocolDataDynamicValue6(ctx, data)
	default:
		ProtocolError(ctx, fmt.Sprintf("Skipping unknown protocol data type: %T", data))
		return
	}

	fileName := fmt.Sprintf("%d_%s_%s_%s.%s", time.Now().Unix(), rpc, message, field, fileExtension)
	filePath := path.Join(dataDir, fileName)
	logFields := map[string]interface{}{KeyProtocolDataFile: filePath} // should not be persisted using With()

	ProtocolTrace(ctx, "Writing protocol data file", logFields)

	err := os.WriteFile(filePath, fileContents, 0644)

	if err != nil {
		ProtocolError(ctx, fmt.Sprintf("Unable to write protocol data file: %s", err), logFields)
		return
	}

	ProtocolTrace(ctx, "Wrote protocol data file", logFields)
}

func protocolDataDynamicValue5(_ context.Context, value *tfprotov5.DynamicValue) (string, []byte) {
	if value == nil {
		return fileExtEmpty, nil
	}

	// (tfprotov5.DynamicValue).Unmarshal() prefers JSON first, so prefer to
	// output JSON if found.
	if len(value.JSON) > 0 {
		return fileExtJson, value.JSON
	}

	if len(value.MsgPack) > 0 {
		return fileExtMsgpack, value.MsgPack
	}

	return fileExtEmpty, nil
}

func protocolDataDynamicValue6(_ context.Context, value *tfprotov6.DynamicValue) (string, []byte) {
	if value == nil {
		return fileExtEmpty, nil
	}

	// (tfprotov6.DynamicValue).Unmarshal() prefers JSON first, so prefer to
	// output JSON if found.
	if len(value.JSON) > 0 {
		return fileExtJson, value.JSON
	}

	if len(value.MsgPack) > 0 {
		return fileExtMsgpack, value.MsgPack
	}

	return fileExtEmpty, nil
}
