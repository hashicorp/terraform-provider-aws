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

	writeProtocolFile(ctx, dataDir, rpc, message, field, fileExtension, fileContents)
}

// ProtocolPrivateData emits raw protocol private data to a file, if given a
// directory. This data is "private" in the sense that it is provider-owned,
// rather than something managed by Terraform.
//
// The directory must exist and be writable, prior to invoking this function.
//
// File names are in the format: {TIME}_{RPC}_{MESSAGE}_{FIELD}(.empty)
func ProtocolPrivateData(ctx context.Context, dataDir string, rpc string, message string, field string, data []byte) {
	if dataDir == "" {
		// Write a log, only once, that explains how to enable this functionality.
		protocolDataSkippedLog.Do(func() {
			ProtocolTrace(ctx, "Skipping protocol data file writing because no data directory is set. "+
				fmt.Sprintf("Use the %s environment variable to enable this functionality.", EnvTfLogSdkProtoDataDir))
		})

		return
	}

	var fileExtension string

	if len(data) == 0 {
		fileExtension = fileExtEmpty
	}

	writeProtocolFile(ctx, dataDir, rpc, message, field, fileExtension, data)
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

func writeProtocolFile(ctx context.Context, dataDir string, rpc string, message string, field string, fileExtension string, fileContents []byte) {
	fileName := fmt.Sprintf("%d_%s_%s_%s", time.Now().Unix(), rpc, message, field)

	if fileExtension != "" {
		fileName += "." + fileExtension
	}

	filePath := path.Join(dataDir, fileName)
	ctx = ProtocolSetField(ctx, KeyProtocolDataFile, filePath)

	ProtocolTrace(ctx, "Writing protocol data file")

	err := os.WriteFile(filePath, fileContents, 0644)

	if err != nil {
		ProtocolError(ctx, "Unable to write protocol data file", map[string]any{KeyError: err.Error()})
		return
	}

	ProtocolTrace(ctx, "Wrote protocol data file")
}
