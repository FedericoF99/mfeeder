package config

import (
	"bufio"
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"
)

var (
	defaultExclusions = "chrome,explorer,WindowsTerminal,ChatGPT,mmc,TrGUI,Codex,ClickUp,Spotify,ApplicationFrameHost,PickerHost"
	filePath          = "mfeederd.conf"
)

type Conf struct {
	exclusions    []string
	includeSystem bool
}

///// Conf methods

func (c *Conf) Exclusions() []string {
	return c.exclusions
}

func (c *Conf) IncludeSystem() bool {
	return c.includeSystem
}

/////

// LoadConfig loads the config file or creates a default one
func LoadConfig() (*Conf, error) {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		cfg, err := createDefault()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer cFile.Close()

	scan := bufio.NewScanner(cFile)
	cfg := Conf{}

	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "EXCLUSIONS=") {
			trimmed := strings.TrimPrefix(line, "EXCLUSIONS=")
			if trimmed != "" {
				cfg.exclusions = strings.Split(trimmed, ",")
			}
		} else if strings.HasPrefix(line, "SYSTEM=") {
			cfg.includeSystem = strings.TrimPrefix(line, "SYSTEM=") == "true"
		}
	}

	if err = scan.Err(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func createDefault() (*Conf, error) {
	cfg := Conf{
		exclusions:    strings.Split(defaultExclusions, ","),
		includeSystem: false,
	}

	err := os.WriteFile(filePath, []byte("EXCLUSIONS="+defaultExclusions+"\nSYSTEM=false\n"), 0644)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// RmExclusion removes an exclusion from the config file if it exists
// and overwrites the file with the new exclusions
func RmExclusion(exc string, cfg *Conf) error {
	if cfg == nil {
		return errors.New("no config found")
	}
	if cfg.exclusions == nil {
		return errors.New("no exclusions found")
	}
	if !slices.Contains(cfg.exclusions, exc) {
		return errors.New("exclusion not found, you should try an existing one")
	}

	excI := slices.Index(cfg.exclusions, exc)
	slices.Delete(cfg.exclusions, excI, excI)

	err := overwriteConf(cfg)
	if err != nil {
		return err
	}

	return nil
}

// AddExclusion adds an exclusion from the config file if it doesn't exist
// and overwrites the file with the new exclusions
func AddExclusion(exc string, cfg *Conf) error {
	if cfg == nil {
		return errors.New("no config found")
	}
	if cfg.exclusions == nil {
		return errors.New("no exclusions found")
	}
	if slices.Contains(cfg.exclusions, exc) {
		return errors.New("exclusion already exists")
	}

	cfg.exclusions = append(cfg.exclusions, exc)

	err := overwriteConf(cfg)
	if err != nil {
		return err
	}

	return nil
}

// ToggleSystem toggles the includeSystem flag in the config and overwrites the file
func ToggleSystem(cfg *Conf) error {
	cfg.includeSystem = !cfg.includeSystem
	return overwriteConf(cfg)
}

func overwriteConf(cfg *Conf) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(file), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "EXCLUSIONS=") {
			lines[i] = "EXCLUSIONS=" + strings.Join(cfg.exclusions, ",")
		} else if strings.HasPrefix(line, "SYSTEM=") {
			lines[i] = "SYSTEM=" + strconv.FormatBool(cfg.includeSystem)
		}
	}

	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return err
	}
	return nil
}
