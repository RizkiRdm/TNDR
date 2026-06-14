package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/RizkiRdm/TNDR/internal/store"
	"github.com/RizkiRdm/TNDR/internal/tui/styles"
	"github.com/RizkiRdm/TNDR/internal/tui/tabs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	activeTab int
	tabs      []string
	store     *store.Store
	width     int
	height    int
	startedAt time.Time
}

func New(s *store.Store) model {
	return model{
		activeTab: 0,
		tabs:      []string{"Dashboard", "Cost", "Cache", "Config", "Logs"},
		store:     s,
		startedAt: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			return m, nil
		case "1":
			m.activeTab = 0
		case "2":
			m.activeTab = 1
		case "3":
			m.activeTab = 2
		case "4":
			m.activeTab = 3
		case "5":
			m.activeTab = 4
		}
	}
	// Delegate tab-specific keys
	switch m.activeTab {
	case 2: // Cache
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "c", "d", "enter":
				// Logic for Cache keys
			}
		}
	case 3: // Config
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "e":
				// Handle Edit
			case "r":
				// Handle Reload
			case "v":
				// Handle Validate
			}
		}
	case 4: // Logs
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "f", "p", "G", "c":
				// Handle Logs
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width < 80 || m.height < 24 {
		return "Terminal too small. Resize to continue."
	}

	doc := lipgloss.NewStyle().Padding(1, 2)

	header := fmt.Sprintf("TENDR - %s", m.tabs[m.activeTab])

	var content string
	switch m.activeTab {
	case 0:
		content = tabs.DashboardView(m.store, m.startedAt)
	case 1:
		content = tabs.CostView(m.store)
	case 2:
		content = tabs.CacheView(m.store)
	case 3:
		content = tabs.ConfigView(m.store)
	case 4:
		content = tabs.LogsView(m.store)
	}

	// Status bar with hints
	var hints []string
	hints = append(hints, "↑↓ navigate", "tab switch", "enter select", "q quit", "? help")
	
	switch m.activeTab {
	case 0:
		hints = append(hints, "r refresh")
	case 1:
		hints = append(hints, "r refresh", "e export", "d toggle view")
	case 2:
		hints = append(hints, "c clear", "d delete", "enter detail")
	case 3:
		hints = append(hints, "e edit", "r reload", "v validate")
	case 4:
		hints = append(hints, "f filter", "p pause", "G latest", "c clear")
	}

	statusBar := styles.StatusBarStyle.Render(strings.Join(hints, "   "))

	return doc.Render(header + "\n\n" + content + "\n\n" + statusBar)
}
