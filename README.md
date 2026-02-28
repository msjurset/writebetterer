# writebetterer

A CLI tool that reads text from stdin, sends it to an LLM, and writes improved text to stdout. It auto-detects whether the input is code, configuration, or prose, then rewrites it accordingly. It can also be used as a general-purpose prompt tool.

## Installation

```sh
make deploy
```

This builds the binary, installs it to `~/.local/bin/`, copies it to Typinator, and installs the man page.

To also set up zsh completions:

```sh
make install-completion
```

## Usage

```sh
# Rewrite prose (auto-detected)
echo "Hello wrold, how are you todya?" | writebetterer

# Rewrite code explicitly
cat main.go | writebetterer -mode code -provider claude

# Ask a question directly
echo "What is the capital of France?" | writebetterer -prompt

# Verbose output
echo "Fix this sentense." | writebetterer -v
```

## Flags

| Flag | Description |
|------|-------------|
| `-mode` | Content type: `auto` (default), `code`, `config`, or `prose` |
| `-model` | Override the default model selection |
| `-provider` | LLM provider: `gemini` (default) or `claude` |
| `-prompt` | Send input as a direct prompt, bypassing classify/rewrite |
| `-v` | Print provider, mode, and model info to stderr |

## Providers

- **Gemini** — uses Flash for classification/prose, Pro for code/config/prompt
- **Claude** — uses Haiku for classification/prose, Sonnet for code/config/prompt

## Configuration

| Environment Variable | Description |
|----------------------|-------------|
| `WRITEBETTERER_PROVIDER` | Default provider (`gemini` or `claude`) |
| `GEMINI_API_KEY` | Gemini API key (or `op://` reference) |
| `ANTHROPIC_API_KEY` | Claude API key (or `op://` reference) |

API keys can be literal values or 1Password `op://` references. If unset, keys are resolved from 1Password automatically.

## License

MIT
