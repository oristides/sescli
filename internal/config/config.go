package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sescli/internal/presets"
)

const EnvPath = "SESCLI_CONFIG"

type Config struct {
	Defaults      Defaults            `json:"defaults"`
	DefaultPreset string              `json:"default_preset,omitempty"`
	Audience      string              `json:"audience,omitempty"`
	Profile       string              `json:"profile,omitempty"`
	Format        string              `json:"format,omitempty"`
	Limit         int                 `json:"limit,omitempty"`
	Page          int                 `json:"page,omitempty"`
	Presets       map[string][]string `json:"presets"`
}

type Defaults struct {
	Where    string `json:"where"`
	What     string `json:"what"`
	Audience string `json:"audience"`
	Format   string `json:"format"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
}

func Default() Config {
	defaults := presets.Defaults()
	return Config{
		Defaults: Defaults{
			Where:    defaults.Preset,
			What:     defaults.Profile,
			Audience: defaults.Audience,
			Format:   "json",
			Limit:    defaults.PerPage,
			Page:     defaults.Page,
		},
		DefaultPreset: defaults.Preset,
		Audience:      defaults.Audience,
		Profile:       defaults.Profile,
		Format:        "json",
		Limit:         defaults.PerPage,
		Page:          defaults.Page,
		Presets:       presets.ClonePresetMap(presets.DefaultInstallPresets()),
	}
}

func Path() (string, error) {
	if env := os.Getenv(EnvPath); env != "" {
		return env, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sescli", "config.json"), nil
}

func Load(path string) (Config, error) {
	if path == "" {
		var err error
		path, err = Path()
		if err != nil {
			return Default(), err
		}
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Default(), err
	}
	if cfg.Presets == nil {
		cfg.Presets = Default().Presets
	}
	cfg = normalize(cfg)
	return cfg, nil
}

func Write(path string, cfg Config, force bool) error {
	if path == "" {
		var err error
		path, err = Path()
		if err != nil {
			return err
		}
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return os.ErrExist
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	cfg = normalize(cfg)
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func Ensure(path string) (bool, error) {
	if path == "" {
		var err error
		path, err = Path()
		if err != nil {
			return false, err
		}
	}
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	return true, Write(path, Default(), false)
}

func Setup(input io.Reader, output io.Writer, path string, force bool) (Config, error) {
	cfg := Default()
	reader := bufio.NewReader(input)
	var err error
	if cfg.Defaults.Where, err = ask(reader, output, "Default venue preset", cfg.Defaults.Where); err != nil {
		return Config{}, err
	}
	if cfg.Defaults.Audience, err = ask(reader, output, "Default audience", cfg.Defaults.Audience); err != nil {
		return Config{}, err
	}
	if cfg.Defaults.What, err = ask(reader, output, "Default interests", cfg.Defaults.What); err != nil {
		return Config{}, err
	}
	if cfg.Defaults.Format, err = ask(reader, output, "Default output format", cfg.Defaults.Format); err != nil {
		return Config{}, err
	}
	limit, err := ask(reader, output, "Default limit", strconv.Itoa(cfg.Defaults.Limit))
	if err != nil {
		return Config{}, err
	}
	if parsed, err := strconv.Atoi(limit); err == nil && parsed > 0 {
		cfg.Defaults.Limit = parsed
	}
	cfg.DefaultPreset = cfg.Defaults.Where
	cfg.Audience = cfg.Defaults.Audience
	cfg.Profile = cfg.Defaults.What
	cfg.Format = cfg.Defaults.Format
	cfg.Limit = cfg.Defaults.Limit
	cfg.Page = cfg.Defaults.Page
	cfg = normalize(cfg)
	if err := Write(path, cfg, force); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ask(reader *bufio.Reader, output io.Writer, prompt, fallback string) (string, error) {
	fmt.Fprintf(output, "%s [%s]: ", prompt, fallback)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return fallback, nil
	}
	return line, nil
}

func normalize(cfg Config) Config {
	def := Default()
	if cfg.DefaultPreset != "" && cfg.DefaultPreset != cfg.Defaults.Where {
		cfg.Defaults.Where = cfg.DefaultPreset
	}
	if cfg.Profile != "" && cfg.Profile != cfg.Defaults.What {
		cfg.Defaults.What = cfg.Profile
	}
	if cfg.Audience != "" && cfg.Audience != cfg.Defaults.Audience {
		cfg.Defaults.Audience = cfg.Audience
	}
	if cfg.Format != "" && cfg.Format != cfg.Defaults.Format {
		cfg.Defaults.Format = cfg.Format
	}
	if cfg.Limit > 0 && cfg.Limit != cfg.Defaults.Limit {
		cfg.Defaults.Limit = cfg.Limit
	}
	if cfg.Page > 0 && cfg.Page != cfg.Defaults.Page {
		cfg.Defaults.Page = cfg.Page
	}
	if cfg.Defaults.Where == "" {
		cfg.Defaults.Where = firstNonEmpty(cfg.DefaultPreset, def.Defaults.Where)
	}
	if cfg.Defaults.What == "" {
		cfg.Defaults.What = firstNonEmpty(cfg.Profile, def.Defaults.What)
	}
	if cfg.Defaults.Audience == "" {
		cfg.Defaults.Audience = firstNonEmpty(cfg.Audience, def.Defaults.Audience)
	}
	if cfg.Defaults.Format == "" {
		cfg.Defaults.Format = firstNonEmpty(cfg.Format, def.Defaults.Format)
	}
	if cfg.Defaults.Limit == 0 {
		if cfg.Limit > 0 {
			cfg.Defaults.Limit = cfg.Limit
		} else {
			cfg.Defaults.Limit = def.Defaults.Limit
		}
	}
	if cfg.Defaults.Page == 0 {
		if cfg.Page > 0 {
			cfg.Defaults.Page = cfg.Page
		} else {
			cfg.Defaults.Page = def.Defaults.Page
		}
	}
	cfg.DefaultPreset = cfg.Defaults.Where
	cfg.Profile = cfg.Defaults.What
	cfg.Audience = cfg.Defaults.Audience
	cfg.Format = cfg.Defaults.Format
	cfg.Limit = cfg.Defaults.Limit
	cfg.Page = cfg.Defaults.Page
	if cfg.Presets == nil {
		cfg.Presets = presets.ClonePresetMap(def.Presets)
	}
	return cfg
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
