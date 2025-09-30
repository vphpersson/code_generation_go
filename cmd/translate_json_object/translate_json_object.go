package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	motmedelEnv "github.com/Motmedel/utils_go/pkg/env"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	motmedelContextLogger "github.com/Motmedel/utils_go/pkg/log/context_logger"
	errorLogger "github.com/Motmedel/utils_go/pkg/log/error_logger"
	"github.com/vphpersson/code_generation/pkg/code_generation"
	"github.com/vphpersson/code_generation/pkg/translate"
)

func main() {
	logger := errorLogger.Logger{
		Logger: motmedelContextLogger.New(
			slog.NewJSONHandler(os.Stderr, nil),
			&motmedelLog.ErrorContextExtractor{},
		),
	}
	slog.SetDefault(logger.Logger)

	var path string
	flag.StringVar(&path, "path", "", "path to generate code from")

	var packageName string
	flag.StringVar(
		&packageName,
		"package-name",
		motmedelEnv.GetEnvWithDefault("GOPACKAGE", "main"),
		"The name of the package in the output.",
	)

	flag.Parse()

	if path == "" {
		logger.FatalWithExitingMessage("Empty path.", nil)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when reading the file.",
			motmedelErrors.NewWithTrace(fmt.Errorf("os read file: %w", err), path),
		)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when unmarshalling the file data.",
			motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), data),
		)
	}

	code, err := translate.Map(m)
	if err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when translating the map.",
			motmedelErrors.New(fmt.Errorf("translate map: %w", err), m),
		)
	}

	output, err := code_generation.MakeFileContent(
		code,
		packageName,
		"translate_json_object",
		nil,
	)
	if err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when generating the output file content.",
			motmedelErrors.New(fmt.Errorf("make file content: %w", err), code, packageName),
		)
	}

	fmt.Println(string(output))
}
