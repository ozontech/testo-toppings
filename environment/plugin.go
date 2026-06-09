package environment

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ozontech/testo"
	"github.com/ozontech/testo/testoplugin"
)

// DefaultEnviroments is the default list of .env files to load.
// By default, it loads the .env file from the current directory when no files are specified.
var DefaultEnviroments = []string{".env"}

var _ testoplugin.Plugin = (*PluginEnvironment)(nil)

type PluginEnvironment struct {
	*testo.T
	filenames []string
}

// Plugin implements [testoplugin.Plugin].
// Loads environment variables from .env files before all tests run.
// Later files overwrite earlier files for the same key.
// If no files are specified, loads the default .env file from the current directory.
func (p *PluginEnvironment) Plugin(
	_ testoplugin.Plugin,
	options ...testoplugin.Option,
) testoplugin.Spec {

	for _, opt := range options {
		if o, ok := opt.Value.(option); ok {
			o(p)
		}
	}

	if len(p.filenames) == 0 {
		p.filenames = append(p.filenames, DefaultEnviroments...)
	}

	return testoplugin.Spec{
		Hooks: p.hooks(),
	}
}

func (p *PluginEnvironment) hooks() testoplugin.Hooks {
	envs := map[string]string{}

	// Read environment variables from all specified files.
	// Later files overwrite earlier files for the same key.
	for _, fileName := range p.filenames {
		err := appendValuesFromEnv(envs, fileName)
		if err != nil {
			p.T.Errorf("Failed to load .env file %s: %v", fileName, err)
		}
	}

	return testoplugin.Hooks{
		BeforeAll: testoplugin.Hook{
			Func: func() {
				// Set all environment variables before running any tests
				for k, v := range envs {
					p.T.Setenv(k, v)
				}
			},
		},
	}
}

// appendValuesFromEnv appends environment variables from a .env file to the envs map.
// Lines starting with # or empty lines are ignored.
// Values are trimmed and outer quotes are removed (e.g., "value" -> value).
// Later lines in the file take precedence over earlier lines for the same key.
func appendValuesFromEnv(envs map[string]string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("line %d in %s: expected KEY=VALUE format", lineNum, path)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes from value
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		envs[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	return nil
}
