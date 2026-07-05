package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var (
	defaultExclusions = "chrome,explorer,WindowsTerminal,ChatGPT,mmc,TrGUI,Codex,ClickUp,Spotify,ApplicationFrameHost,PickerHost"
	filePath          = "mfeederd.conf"
)

type Conf struct {
	exclusions []string
}

///// Conf methods

func (c *Conf) Exclusions() []string {
	return c.exclusions
}

/////

// LoadConfig loads the config file or creates a default one
func LoadConfig(ow bool) (*Conf, error) {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) && ow {
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
				trimmed = strings.Replace(trimmed, " ", "", -1)
				e := strings.Split(trimmed, ",")

				for i := range len(e) {
					if e[i] == "" || slices.Contains(cfg.exclusions, e[i]) {
						continue
					}

					cfg.exclusions = append(cfg.exclusions, e[i])
				}
			}
		}
	}

	if err = scan.Err(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func createDefault() (*Conf, error) {
	cfg := Conf{
		exclusions: strings.Split(defaultExclusions, ","),
	}

	err := os.WriteFile(filePath, []byte("EXCLUSIONS="+defaultExclusions+"\n"), 0644)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// RmExclusion removes an exclusion from the config file if it exists
// and overwrites the file with the new exclusions
func RmExclusion(exc string) error {
	exc = strings.TrimSpace(exc)

	if exc == "" {
		return errors.New("empty exclusion not allowed")
	}

	cfg, err := LoadConfig(false)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if cfg.exclusions == nil {
		return errors.New("no exclusions found")
	}
	if !slices.Contains(cfg.exclusions, exc) {
		return errors.New("exclusion not found, you should try an existing one")
	}

	excI := slices.Index(cfg.exclusions, exc)
	cfg.exclusions = slices.Delete(cfg.exclusions, excI, excI+1)

	err = overwriteConf(cfg)
	if err != nil {
		return err
	}

	return nil
}

// AddExclusion adds an exclusion from the config file if it doesn't exist
// and overwrites the file with the new exclusions
func AddExclusion(exc string) error {
	exc = strings.TrimSpace(exc)

	if exc == "" {
		return errors.New("empty exclusion not allowed")
	}

	cfg, err := LoadConfig(false)
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	if cfg.exclusions == nil {
		return errors.New("no exclusions found")
	}
	if slices.Contains(cfg.exclusions, exc) {
		return errors.New("exclusion already exists")
	}

	cfg.exclusions = append(cfg.exclusions, exc)

	err = overwriteConf(cfg)
	if err != nil {
		return err
	}

	return nil
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
		}
	}

	err = os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetExclusions() ([]string, error) {
	cfg, err := LoadConfig(false)
	if err != nil {
		return nil, err
	}

	return cfg.exclusions, nil
}
