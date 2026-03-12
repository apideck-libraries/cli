// Package tui implements the interactive TUI explorer for the Apideck CLI.
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/apideck-io/cli/internal/spec"
	"github.com/apideck-io/cli/internal/ui"
)

// focusedPanel tracks which panel has keyboard focus.
type focusedPanel int

const (
	panelList   focusedPanel = iota
	panelDetail focusedPanel = iota
)

// ExplorerModel is the bubbletea model for the API explorer TUI.
type ExplorerModel struct {
	apiSpec           *spec.APISpec
	groupKey          string
	resources         map[string]*spec.Resource
	items             []ListItem
	cursor            int
	expandedResources map[string]bool
	focus             focusedPanel
	width             int
	height            int
}

// NewExplorerModel creates an ExplorerModel for the given API group.
// If groupKey is empty, the first available group is used.
func NewExplorerModel(apiSpec *spec.APISpec, groupKey string) ExplorerModel {
	if groupKey == "" {
		for k := range apiSpec.APIGroups {
			groupKey = k
			break
		}
	}

	var resources map[string]*spec.Resource
	if g, ok := apiSpec.APIGroups[groupKey]; ok {
		resources = g.Resources
	}

	expanded := make(map[string]bool)
	items := BuildListItems(resources, expanded)

	return ExplorerModel{
		apiSpec:           apiSpec,
		groupKey:          groupKey,
		resources:         resources,
		items:             items,
		cursor:            0,
		expandedResources: expanded,
		focus:             panelList,
		width:             120,
		height:            40,
	}
}

// Init satisfies tea.Model. No initial commands needed.
func (m ExplorerModel) Init() tea.Cmd {
	return nil
}

// Update handles all incoming messages and key events.
func (m ExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case msg.Code == 'q' || (msg.Code == 'c' && msg.Mod == tea.ModCtrl):
			return m, tea.Quit

		// Navigate up
		case msg.Code == tea.KeyUp || msg.Code == 'k':
			if m.cursor > 0 {
				m.cursor--
			}

		// Navigate down
		case msg.Code == tea.KeyDown || msg.Code == 'j':
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		// Expand / collapse or select operation
		case msg.Code == tea.KeyEnter:
			if m.cursor >= 0 && m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if item.Kind == ListItemResource {
					if item.Expanded {
						m.expandedResources = CollapseResource(m.expandedResources, item.ResourceKey)
					} else {
						m.expandedResources = ExpandResource(m.expandedResources, item.ResourceKey)
					}
					m.items = BuildListItems(m.resources, m.expandedResources)
				}
			}

		// Focus right panel
		case msg.Code == tea.KeyRight || msg.Code == 'l':
			m.focus = panelDetail

		// Focus left panel
		case msg.Code == tea.KeyLeft || msg.Code == 'h':
			m.focus = panelList
		}
	}

	return m, nil
}

// View renders the full TUI layout as a tea.View.
func (m ExplorerModel) View() tea.View {
	return tea.NewView(m.render())
}

// render builds the string content of the view.
func (m ExplorerModel) render() string {
	totalWidth := m.width
	if totalWidth < 40 {
		totalWidth = 40
	}
	totalHeight := m.height
	if totalHeight < 10 {
		totalHeight = 10
	}

	// Title bar
	title := titleStyle.Render(fmt.Sprintf("Apideck Explorer — %s %s", m.apiSpec.Name, m.apiSpec.Version))
	group := dimStyle.Render(fmt.Sprintf(" (%s)", m.groupKey))
	titleBar := title + group
	titleBar = lipgloss.NewStyle().Width(totalWidth).Render(titleBar)

	// Help line at the bottom
	help := dimStyle.Render("↑/↓ navigate  enter expand/collapse  ←/→ focus panel  q quit")
	help = lipgloss.NewStyle().Width(totalWidth).Render(help)

	// Inner height for panels (title + help + borders consume some lines)
	innerHeight := totalHeight - 4 // title (1) + newline + help (1) + newline
	if innerHeight < 5 {
		innerHeight = 5
	}

	// Panel widths: list ~35%, detail ~65%
	leftWidth := totalWidth * 35 / 100
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := totalWidth - leftWidth - 4 // 4 for two borders padding
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Render left panel content
	listContent := m.renderList(leftWidth, innerHeight)
	// Render right panel content
	var selectedOp *spec.Operation
	if m.cursor >= 0 && m.cursor < len(m.items) {
		if m.items[m.cursor].Kind == ListItemOperation {
			selectedOp = m.items[m.cursor].Operation
		}
	}
	detailContent := RenderDetailPanel(selectedOp, rightWidth, innerHeight)

	// Wrap panels in borders
	var leftPanel, rightPanel string
	if m.focus == panelList {
		leftPanel = activePanelStyle.Width(leftWidth).Height(innerHeight).Render(listContent)
		rightPanel = panelStyle.Width(rightWidth).Height(innerHeight).Render(detailContent)
	} else {
		leftPanel = panelStyle.Width(leftWidth).Height(innerHeight).Render(listContent)
		rightPanel = activePanelStyle.Width(rightWidth).Height(innerHeight).Render(detailContent)
	}

	// Join panels side by side
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	return strings.Join([]string{titleBar, panels, help}, "\n")
}

// renderList renders the left-panel endpoint list.
func (m ExplorerModel) renderList(width, height int) string {
	if len(m.items) == 0 {
		return dimStyle.Render("No resources found.")
	}

	// Compute visible window around cursor.
	start := 0
	end := len(m.items)
	if end-start > height {
		// Centre cursor in view.
		start = m.cursor - height/2
		if start < 0 {
			start = 0
		}
		end = start + height
		if end > len(m.items) {
			end = len(m.items)
			start = end - height
			if start < 0 {
				start = 0
			}
		}
	}

	var rows []string
	for i := start; i < end; i++ {
		rows = append(rows, RenderListItem(m.items[i], i == m.cursor, width))
	}
	return strings.Join(rows, "\n")
}

// Run starts the TUI explorer program and blocks until the user quits.
func Run(apiSpec *spec.APISpec, groupKey string) error {
	m := NewExplorerModel(apiSpec, groupKey)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// Ensure ExplorerModel satisfies the tea.Model interface at compile time.
var _ tea.Model = ExplorerModel{}

// ui import is used for the title style color; confirm package reference.
var _ = ui.ColorPrimary
