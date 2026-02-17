package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// UnmarshalJSON deserializes JSON data from either a string or file
// If data is not empty, it will be unmarshalled directly
// If data is empty and filename is provided, the file will be read first
func UnmarshalJSON(ctx context.Context, data string, filename string, v interface{}) error {
	var b []byte

	if strings.TrimSpace(data) != "" {
		b = []byte(data)
	} else if strings.TrimSpace(filename) != "" {
		var err error
		b, err = os.ReadFile(filename)
		if err != nil {
			tflog.Error(ctx, "Failed to read file", map[string]interface{}{
				"filename": filename,
				"error":    err.Error(),
			})
			return fmt.Errorf("failed to read file '%s': %w", filename, err)
		}
	} else {
		return fmt.Errorf("either data string or filename must be provided")
	}

	if err := json.Unmarshal(b, v); err != nil {
		tflog.Error(ctx, "Failed to unmarshal JSON", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// DeprecatedFileLog provides backwards compatibility for file-based logging
// New code should use tflog from terraform-plugin-log instead
// This function logs to stderr if UCMDB_PROVIDER_LOG environment variable is set
func DeprecatedFileLog(ctx context.Context, level string, info string, message interface{}) {
	logFile, hasLogFile := os.LookupEnv("UCMDB_PROVIDER_LOG")

	if !hasLogFile {
		return
	}

	logFile = strings.TrimSpace(logFile)
	if logFile == "" {
		return
	}

	// Use tflog instead of file writing for better integration
	switch level {
	case "ERROR":
		tflog.Error(ctx, info, map[string]interface{}{"detail": message})
	case "WARN":
		tflog.Warn(ctx, info, map[string]interface{}{"detail": message})
	case "INFO":
		tflog.Info(ctx, info, map[string]interface{}{"detail": message})
	case "DEBUG":
		tflog.Debug(ctx, info, map[string]interface{}{"detail": message})
	default:
		tflog.Info(ctx, info, map[string]interface{}{"detail": message})
	}
}
