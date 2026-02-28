package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	maxOutputTokens int32 = 4096
	timeout               = 60 * time.Second
)

// Provider abstracts the LLM backend.
type Provider interface {
	Classify(ctx context.Context, text string) (string, error)
	Rewrite(ctx context.Context, model, systemPrompt, text string) (string, error)
	Prompt(ctx context.Context, model, text string) (string, error)
	DefaultModel(contentType string) string
}

// System prompts
const classifyPrompt = `Classify the following input as exactly one of: CODE, CONFIG, or PROSE.

CODE: programming language source code (functions, classes, statements, scripts)
CONFIG: configuration files (YAML, JSON, TOML, INI, XML, env files, Dockerfiles, Makefiles, CI configs)
PROSE: natural language text (emails, documentation, comments, chat messages, articles)

Respond with exactly one word: CODE, CONFIG, or PROSE. Nothing else.`

const rewriteCodePrompt = `You are a code improvement assistant. Rewrite the provided code to be better:
- Fix bugs, typos, and logical errors
- Improve naming, readability, and idiomatic style
- Preserve the original language, structure, and intent
- Do NOT change functionality unless fixing a clear bug
- Do NOT add comments explaining your changes
- Do NOT wrap output in markdown code fences unless the input had them
- Output ONLY the rewritten code, nothing else — no preamble, no explanation`

const rewriteConfigPrompt = `You are a configuration improvement assistant. Rewrite the provided configuration to be better:
- Fix syntax errors, typos, and formatting issues
- Improve organization and readability
- Add missing but clearly implied fields if appropriate
- Preserve the original format (YAML stays YAML, JSON stays JSON, etc.)
- Do NOT change semantics unless fixing a clear error
- Do NOT wrap output in markdown code fences unless the input had them
- Output ONLY the rewritten configuration, nothing else — no preamble, no explanation`

const rewriteProsePrompt = `You are a writing improvement assistant. Rewrite the provided text to be better:
- Fix grammar, spelling, and punctuation errors
- Improve clarity, flow, and readability
- Preserve the original meaning, tone, and voice
- Preserve the original formatting (line breaks, paragraphs, lists)
- Do NOT add new information or change the message
- Do NOT wrap output in markdown code fences
- Output ONLY the rewritten text, nothing else — no preamble, no explanation`

func main() {
	mode := flag.String("mode", "auto", "Content type: auto, code, config, or prose")
	model := flag.String("model", "", "Override model selection")
	providerFlag := flag.String("provider", "", "LLM provider: gemini or claude (default: WRITEBETTERER_PROVIDER env var, then gemini)")
	promptMode := flag.Bool("prompt", false, "Send input as a direct prompt (bypass classify/rewrite)")
	verbose := flag.Bool("v", false, "Print detected mode/model/stats to stderr")
	flag.Parse()

	// Validate mode
	if !*promptMode {
		switch *mode {
		case "auto", "code", "config", "prose":
		default:
			fmt.Fprintf(os.Stderr, "error: invalid mode %q (must be auto, code, config, or prose)\n", *mode)
			os.Exit(1)
		}
	}

	// Resolve provider name
	providerName := *providerFlag
	if providerName == "" {
		providerName = os.Getenv("WRITEBETTERER_PROVIDER")
	}
	if providerName == "" {
		providerName = "gemini"
	}

	// Read input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	text := string(input)
	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "error: empty input")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create provider
	var provider Provider
	switch providerName {
	case "gemini":
		provider, err = newGeminiProvider(ctx)
	case "claude":
		provider, err = newClaudeProvider()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown provider %q (must be gemini or claude)\n", providerName)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating %s provider: %v\n", providerName, err)
		os.Exit(1)
	}

	// Prompt mode: bypass classify/rewrite, send input directly
	if *promptMode {
		selectedModel := *model
		if selectedModel == "" {
			selectedModel = provider.DefaultModel("prompt")
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "provider: %s | mode: prompt | model: %s\n", providerName, selectedModel)
		}

		result, err := provider.Prompt(ctx, selectedModel, text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(result)
		return
	}

	// Determine content type and model
	contentType := *mode
	selectedModel := *model

	if contentType == "auto" {
		classified, err := provider.Classify(ctx, text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error classifying input: %v\n", err)
			os.Exit(1)
		}
		contentType = classified
	}

	if selectedModel == "" {
		selectedModel = provider.DefaultModel(contentType)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "provider: %s | mode: %s | model: %s\n", providerName, contentType, selectedModel)
	}

	// Select the appropriate rewrite prompt
	var systemPrompt string
	switch contentType {
	case "code":
		systemPrompt = rewriteCodePrompt
	case "config":
		systemPrompt = rewriteConfigPrompt
	default:
		systemPrompt = rewriteProsePrompt
	}

	// Call the API for rewriting
	result, err := provider.Rewrite(ctx, selectedModel, systemPrompt, text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error rewriting: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

// resolveKey resolves an API key from an environment variable, with optional
// 1Password fallback. If the env var is empty, the defaultOPRef is used.
// If the env var contains an op:// reference, it is resolved via `op read`.
func resolveKey(envVar, defaultOPRef string) (string, error) {
	key := os.Getenv(envVar)

	switch {
	case key == "":
		return opRead(defaultOPRef)
	case strings.HasPrefix(key, "op://"):
		return opRead(key)
	default:
		return key, nil
	}
}

func opRead(ref string) (string, error) {
	out, err := exec.Command("op", "read", ref).Output()
	if err != nil {
		return "", fmt.Errorf("failed to read from 1Password (op read %q): %v", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// parseClassification normalises a raw classification response to a content type.
func parseClassification(raw string) string {
	classification := strings.ToUpper(strings.TrimSpace(raw))

	switch classification {
	case "CODE":
		return "code"
	case "CONFIG":
		return "config"
	case "PROSE":
		return "prose"
	default:
		return "prose"
	}
}
