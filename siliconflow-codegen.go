package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	modelsEndpoint = "https://api.siliconflow.com/v1/models"
	baseURL        = "https://api.siliconflow.com/v1"

	opencodeProviderKey = "siliconflow"
	opencodeType        = "openai"
	opencodeAPIKey      = "${SILICONFLOW_API_KEY}"

	crushSchemaURL   = "https://charm.land/crush.json"
	crushProviderKey = "siliconflow"
	crushType        = "openai"
	crushAPIKey      = "$SILICONFLOW_API_KEY"

	qwencodeAPIKey = "SILICONFLOW_API_KEY"
)

var (
	inputModalities  = []string{"text", "image", "video", "audio"}
	outputModalities = []string{"text"}
)

type apiResponse struct {
	Data []apiModel `json:"data"`
}

type apiModel struct {
	ID string `json:"id"`
}

type openCodeConfig struct {
	SiliconFlow openCodeProvider `json:"siliconflow"`
}

type openCodeProvider struct {
	Type    string                   `json:"type"`
	BaseURL string                   `json:"baseURL"`
	APIKey  string                   `json:"apiKey"`
	Models  map[string]openCodeModel `json:"models"`
}

type openCodeModel struct {
	Name       string     `json:"name"`
	Modalities modalities `json:"modalities"`
}

type modalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

type crushConfig struct {
	Schema    string                   `json:"$schema"`
	Providers map[string]crushProvider `json:"providers"`
}

type crushProvider struct {
	Type    string       `json:"type"`
	BaseURL string       `json:"base_url"`
	APIKey  string       `json:"api_key"`
	Models  []crushModel `json:"models"`
}

type crushModel struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	CostPer1MIn         float64 `json:"cost_per_1m_in,omitempty"`
	CostPer1MOut        float64 `json:"cost_per_1m_out,omitempty"`
	CostPer1MInCached   float64 `json:"cost_per_1m_in_cached,omitempty"`
	CostPer1MOutCached  float64 `json:"cost_per_1m_out_cached,omitempty"`
	ContextWindow       int     `json:"context_window,omitempty"`
	DefaultMaxTokens    int     `json:"default_max_tokens,omitempty"`
	CanReason           bool    `json:"can_reason,omitempty"`
	SupportsAttachments bool    `json:"supports_attachments,omitempty"`
}

type qwencodeConfig struct {
	OpenAI []qwencodeModel `json:"openai"`
}

type qwencodeModel struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	EnvKey           string                   `json:"envKey"`
	BaseURL          string                   `json:"baseUrl"`
	GenerationConfig qwencodeGenerationConfig `json:"generationConfig,omitempty"`
}

type qwencodeGenerationConfig struct {
	Modalities map[string]bool `json:"modalities"`
}

func main() {
	genOpenCode := flag.Bool("gen-opencode", false, "generate an OpenCode-compatible SiliconFlow provider config instead of printing the raw API response")
	genCrush := flag.Bool("gen-crush", false, "generate a Crush-compatible SiliconFlow provider config instead of printing the raw API response")
	genQwencode := flag.Bool("gen-qwencode", false, "generate a Qwencode-compatible SiliconFlow provider config instead of printing the raw API response")
	genClaude := flag.Bool("claude", false, "list SiliconFlow models and print an export command for ANTHROPIC_MODEL=<selected>")
	flag.Parse()

	requestedActions := 0
	if *genOpenCode {
		requestedActions++
	}
	if *genCrush {
		requestedActions++
	}
	if *genQwencode {
		requestedActions++
	}
	if *genClaude {
		requestedActions++
	}
	if requestedActions > 1 {
		fmt.Fprintln(os.Stderr, "ERROR: only one of --gen-opencode, --gen-crush, --gen-qwencode, or --claude can be used at a time")
		os.Exit(1)
	}

	apiKey := os.Getenv("SILICONFLOW_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: SILICONFLOW_API_KEY is not set")
		os.Exit(1)
	}

	body, err := fetchModels(apiKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch {
	case *genOpenCode:
		config, err := generateOpenCodeConfig(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, err := os.Stdout.Write(config); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing output: %v\n", err)
			os.Exit(1)
		}
	case *genCrush:
		config, err := generateCrushConfig(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, err := os.Stdout.Write(config); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing output: %v\n", err)
			os.Exit(1)
		}
	case *genQwencode:
		config, err := generateQwencodeConfig(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if _, err := os.Stdout.Write(config); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: writing output: %v\n", err)
			os.Exit(1)
		}
	case *genClaude:
		ids, err := parseModelIDs(body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var input io.Reader = os.Stdin
		var output io.Writer = os.Stderr
		if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
			defer tty.Close()
			input = tty
			output = tty
		}

		selected, err := promptForModel(ids, input, output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Setenv("ANTHROPIC_MODEL", selected)

		claudePath, err := exec.LookPath("claude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: 'claude' CLI not found in PATH. Please install it or ensure it is in your PATH.\n")
			fmt.Fprintf(os.Stderr, "To set it manually in your shell, run: export ANTHROPIC_MODEL=%q\n", selected)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Launching Claude Code with ANTHROPIC_MODEL=%s...\n", selected)

		cmd := exec.Command(claudePath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "ERROR: running 'claude': %v\n", err)
			os.Exit(1)
		}
	default:
		printRawResponse(body)
	}
}

func fetchModels(apiKey string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, modelsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("ERROR: creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ERROR: reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"ERROR: API returned status %d\nResponse: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	return body, nil
}

func generateOpenCodeConfig(body []byte) ([]byte, error) {
	ids, err := parseModelIDs(body)
	if err != nil {
		return nil, err
	}

	models := make(map[string]openCodeModel, len(ids))
	for _, id := range ids {
		models[id] = openCodeModel{
			Name: id,
			Modalities: modalities{
				Input:  append([]string(nil), inputModalities...),
				Output: append([]string(nil), outputModalities...),
			},
		}
	}

	if len(models) == 0 {
		return nil, errors.New("ERROR: no models found in API response")
	}

	config := openCodeConfig{
		SiliconFlow: openCodeProvider{
			Type:    opencodeType,
			BaseURL: baseURL,
			APIKey:  opencodeAPIKey,
			Models:  models,
		},
	}

	// encoding/json sorts map keys, but sort the values here too so generated
	// modalities remain stable even if the source slices are changed later.
	for id, model := range config.SiliconFlow.Models {
		sort.Strings(model.Modalities.Input)
		sort.Strings(model.Modalities.Output)
		config.SiliconFlow.Models[id] = model
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ERROR: encoding OpenCode config: %w", err)
	}

	return append(encoded, '\n'), nil
}

func generateCrushConfig(body []byte) ([]byte, error) {
	ids, err := parseModelIDs(body)
	if err != nil {
		return nil, err
	}

	models := make([]crushModel, 0, len(ids))
	for _, id := range ids {
		models = append(models, crushModel{
			ID:   id,
			Name: id,
		})
	}

	if len(models) == 0 {
		return nil, errors.New("ERROR: no models found in API response")
	}

	config := crushConfig{
		Schema: crushSchemaURL,
		Providers: map[string]crushProvider{
			crushProviderKey: {
				Type:    crushType,
				BaseURL: baseURL,
				APIKey:  crushAPIKey,
				Models:  models,
			},
		},
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ERROR: encoding Crush config: %w", err)
	}

	return append(encoded, '\n'), nil
}

func generateQwencodeConfig(body []byte) ([]byte, error) {
	ids, err := parseModelIDs(body)
	if err != nil {
		return nil, err
	}

	models := make([]qwencodeModel, 0, len(ids))
	for _, id := range ids {
		models = append(models, qwencodeModel{
			ID:      id,
			Name:    id,
			EnvKey:  qwencodeAPIKey,
			BaseURL: baseURL,
			GenerationConfig: qwencodeGenerationConfig{
				Modalities: map[string]bool{
					"image": true,
				},
			},
		})
	}

	if len(models) == 0 {
		return nil, errors.New("ERROR: no models found in API response")
	}

	config := qwencodeConfig{
		OpenAI: models,
	}

	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ERROR: encoding Qwencode config: %w", err)
	}

	return append(encoded, '\n'), nil
}

func parseModelIDs(body []byte) ([]string, error) {
	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("ERROR: parsing API response: %w", err)
	}

	seen := make(map[string]struct{}, len(response.Data))
	for _, model := range response.Data {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}

		seen[id] = struct{}{}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	return ids, nil
}

func promptForModel(models []string, input io.Reader, output io.Writer) (string, error) {
	if len(models) == 0 {
		return "", errors.New("ERROR: no SiliconFlow models found in API response")
	}

	cols := chooseColumnCount(len(models), longestID(models), terminalWidth())
	if err := writeModelGrid(output, models, cols); err != nil {
		return "", err
	}

	reader := bufio.NewReader(input)
	prompt := fmt.Sprintf("\nmodel [1-%d, blank to quit]> ", len(models))
	for {
		if _, err := fmt.Fprint(output, prompt); err != nil {
			return "", fmt.Errorf("ERROR: writing prompt: %w", err)
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("ERROR: reading selection: %w", err)
		}

		choice := strings.TrimSpace(line)
		if choice == "" {
			return "", errors.New("ERROR: model selection cancelled")
		}

		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(models) {
			fmt.Fprintf(output, "ERROR: enter a number between 1 and %d\n", len(models))
			continue
		}

		return models[index-1], nil
	}
}

func printRawResponse(body []byte) {
	if len(body) == 0 {
		return
	}

	os.Stdout.Write(body)
	if body[len(body)-1] != '\n' {
		os.Stdout.WriteString("\n")
	}
}

func longestID(models []string) int {
	max := 0
	for _, model := range models {
		if len(model) > max {
			max = len(model)
		}
	}
	return max
}

func terminalWidth() int {
	if value := os.Getenv("COLUMNS"); value != "" {
		if width, err := strconv.Atoi(value); err == nil && width > 0 {
			return width
		}
	}
	return 100
}

// chooseColumnCount picks a column count that lays the models out in roughly
// square rows within the terminal width, leaving a small margin per column.
func chooseColumnCount(count, idWidth, width int) int {
	if width <= 0 {
		width = 100
	}
	margin := 4 // room for the "NNN) " prefix and trailing space
	cell := idWidth + margin
	if cell < 8 {
		cell = 8
	}
	maxCols := width / cell
	if maxCols < 1 {
		maxCols = 1
	}
	if maxCols > count {
		maxCols = count
	}

	bestCols := 1
	bestMaxDim := count
	for c := 1; c <= maxCols; c++ {
		rows := (count + c - 1) / c
		maxDim := rows
		if c > rows {
			maxDim = c
		}
		// Prefer the column count that minimises the larger of rows/cols
		// (i.e. a roughly square grid), breaking ties by the smaller
		// column count so the model IDs are easier to scan.
		if maxDim < bestMaxDim {
			bestMaxDim = maxDim
			bestCols = c
		}
	}
	return bestCols
}

func writeModelGrid(output io.Writer, models []string, cols int) error {
	if _, err := fmt.Fprintln(output, "Available SiliconFlow models (enter the number to set ANTHROPIC_MODEL):"); err != nil {
		return fmt.Errorf("ERROR: writing header: %w", err)
	}

	if cols < 1 {
		cols = 1
	}
	rows := (len(models) + cols - 1) / cols
	idWidth := longestID(models)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			idx := col*rows + row
			if idx >= len(models) {
				continue
			}
			if _, err := fmt.Fprintf(output, "%3d) %-*s  ", idx+1, idWidth, models[idx]); err != nil {
				return fmt.Errorf("ERROR: writing model list: %w", err)
			}
		}
		if _, err := fmt.Fprintln(output); err != nil {
			return fmt.Errorf("ERROR: writing model list: %w", err)
		}
	}

	return nil
}
