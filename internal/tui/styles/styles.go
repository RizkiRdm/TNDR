package styles

import "github.com/charmbracelet/lipgloss"

const (
	ColorBg       = "#0D0D0D"
	ColorSurface  = "#141414"
	ColorSurface2 = "#1C1C1C"
	ColorBorder   = "#2A2A2A"
	ColorBorder2  = "#3A3A3A"
	ColorBorder3  = "#525252"

	ColorTextPrimary   = "#E8E8E8"
	ColorTextSecondary = "#A0A0A0"
	ColorTextTertiary  = "#606060"
	ColorTextMuted     = "#404040"

	ColorSuccess = "#4CAF7D"
	ColorFailed  = "#E05252"
	ColorWarning = "#F5A623"
	ColorRetrying = "#7B68EE"
	ColorIdle     = "#3A3A3A"
	ColorActive   = "#F5A623"

	ColorAccent    = "#F5A623"
	ColorAccentDim = "#7A5210"
)

var (
	DocStyle = lipgloss.NewStyle().Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorTextPrimary))

	TabStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(lipgloss.Color(ColorTextTertiary))

	ActiveTabStyle = TabStyle.Copy().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true).
		Underline(true)

	StatusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTextTertiary)).
		Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder))
)
