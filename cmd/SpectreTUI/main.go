package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/model"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
)

// exitCode returns the exit code for a response.
func exitCode(resp schema.Response) int {
	return resp.ExitCode
}

// run executes the CLI logic. It is parameterized for end-to-end testing.
func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, runPrompt func(schema.Payload) (schema.Response, error)) int {
	fs := flag.NewFlagSet("SpectreTUI", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", "", "Path to payload JSON file (reads from stdin if not specified)")
	outputPath := fs.String("output", "", "Path to output JSON file (writes to stdout if not specified)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 1
	}

	var data []byte
	var err error

	if *inputPath != "" {
		data, err = os.ReadFile(*inputPath)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	} else {
		data, err = io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	}

	payload, err := schema.Parse(data)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	resp, err := runPrompt(payload)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	respBytes = append(respBytes, '\n')

	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, respBytes, 0644); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	} else {
		if _, err := stdout.Write(respBytes); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	}

	return exitCode(resp)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, model.RunPrompt))
}
