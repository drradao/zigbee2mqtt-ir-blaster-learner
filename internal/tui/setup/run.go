package setup

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/config"
)

// Run starts the setup form with alt-screen and returns the result when it exits.
func Run(existing config.Config, mode Mode, configPath string) (Result, error) {
	m := New(existing, mode, configPath)
	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return Result{}, err
	}
	return final.(*Model).Result(), nil
}
