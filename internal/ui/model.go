package ui

import (
	"context"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"

	"broom/internal/clean"
	"broom/internal/platform"
	"broom/internal/scan"
)

type appState int

const (
	stateWelcome appState = iota
	stateScanning
	stateSelect
	stateConfirm
	stateCleaning
	stateDone
)

// ---- Bubble Tea messages ----

type itemMsg scan.Item
type progressMsg scan.Progress
type scanDoneMsg struct{}
type cleanResultMsg struct {
	idx int
	err error
}
type cleanDoneMsg struct{}

// ---- List row types (select screen) ----

type rowKind int

const (
	rowHeader rowKind = iota
	rowItem
)

type listRow struct {
	kind     rowKind
	header   scan.Category
	itemIdx  int // index into m.allItems, valid when kind==rowItem
}

// ---- Main model ----

type Model struct {
	state  appState
	plat   platform.Platform
	dryRun bool
	width  int
	height int

	// scanning
	spinner      spinner.Model
	scanProgress map[string]scan.Progress
	allItems     []scan.Item
	scanFinished bool

	// scanning channels (kept for polling)
	itemCh    <-chan scan.Item
	progressCh <-chan scan.Progress

	// select screen
	rows       []listRow
	visRows    []listRow // filtered view
	cursor     int       // position in visRows (item rows only)
	viewOffset int
	filter     string
	filterMode bool

	// confirm screen
	confirmInput string

	// cleaning
	cleanQueue []scan.Item
	cleanIdx   int
	cleanErrs  []error
	freed      int64

	// done
	initialFree int64
	finalFree   int64
}

func New(dryRun bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = stylePrimary

	plat := platform.New()
	free, _ := plat.DiskFreeBytes(homeOrRoot())

	return Model{
		state:        stateWelcome,
		plat:         plat,
		dryRun:       dryRun,
		spinner:      s,
		scanProgress: map[string]scan.Progress{},
		initialFree:  free,
	}
}

func homeOrRoot() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return h
}

// ---- Init ----

func (m Model) Init() tea.Cmd {
	return nil
}

// ---- Update ----

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case itemMsg:
		item := scan.Item(msg)
		m.allItems = append(m.allItems, item)
		return m, pollItems(m.itemCh, m.progressCh)

	case progressMsg:
		p := scan.Progress(msg)
		m.scanProgress[p.Scanner] = p
		return m, pollItems(m.itemCh, m.progressCh)

	case scanDoneMsg:
		m.scanFinished = true
		m.buildRows("")
		if len(m.allItems) == 0 {
			m.state = stateDone
			return m, nil
		}
		m.state = stateSelect
		return m, nil

	case cleanResultMsg:
		if msg.err != nil {
			m.cleanErrs = append(m.cleanErrs, msg.err)
		} else {
			m.freed += m.cleanQueue[msg.idx].SizeBytes
		}
		m.cleanIdx++
		if m.cleanIdx >= len(m.cleanQueue) {
			return m, func() tea.Msg { return cleanDoneMsg{} }
		}
		return m, doClean(m.cleanQueue, m.cleanIdx, m.dryRun)

	case cleanDoneMsg:
		free, _ := m.plat.DiskFreeBytes(homeOrRoot())
		m.finalFree = free
		m.state = stateDone
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {

	case stateWelcome:
		switch msg.String() {
		case "enter", " ":
			return m.startScanning()
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case stateScanning:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case stateSelect:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}
		return m.handleSelectKey(msg)

	case stateConfirm:
		switch msg.String() {
		case "y", "Y":
			return m.startCleaning()
		case "n", "N", "q", "esc":
			m.state = stateSelect
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}

	case stateDone:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visItems := m.visibleItemRows()
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(visItems)-1 {
			m.cursor++
			m.scrollToCursor()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.scrollToCursor()
		}
	case " ":
		if m.cursor < len(visItems) {
			idx := visItems[m.cursor].itemIdx
			m.allItems[idx].Selected = !m.allItems[idx].Selected
		}
	case "a":
		for _, r := range visItems {
			m.allItems[r.itemIdx].Selected = true
		}
	case "A":
		for _, r := range visItems {
			m.allItems[r.itemIdx].Selected = false
		}
	case "/":
		m.filterMode = true
	case "enter":
		if m.selectedCount() > 0 {
			m.state = stateConfirm
		}
	}
	m.buildRows(m.filter)
	return m, nil
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.filterMode = false
	case "ctrl+c":
		return *m, tea.Quit
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
		}
	default:
		if len(msg.Runes) > 0 {
			m.filter += string(msg.Runes)
		}
	}
	m.cursor = 0
	m.viewOffset = 0
	m.buildRows(m.filter)
	return *m, nil
}

// ---- Scanning ----

func (m Model) startScanning() (Model, tea.Cmd) {
	m.state = stateScanning
	roots := m.plat.SearchRoots()
	caches := m.plat.GlobalCaches()

	scanners := []scan.Scanner{
		&scan.DockerScanner{},
		&scan.LLMScanner{},
		&scan.CacheScanner{Caches: caches},
		&scan.NodeScanner{},
		&scan.PythonScanner{},
		&scan.RustScanner{},
		&scan.BuildScanner{},
	}

	ctx := context.Background()
	itemCh, progressCh := scan.Run(ctx, scanners, roots, caches)
	m.itemCh = itemCh
	m.progressCh = progressCh

	return m, tea.Batch(
		m.spinner.Tick,
		pollItems(itemCh, progressCh),
	)
}

// pollItems waits for the next item or progress event.
func pollItems(items <-chan scan.Item, progress <-chan scan.Progress) tea.Cmd {
	return func() tea.Msg {
		select {
		case item, ok := <-items:
			if !ok {
				return scanDoneMsg{}
			}
			return itemMsg(item)
		case p, ok := <-progress:
			if !ok {
				return scanDoneMsg{}
			}
			return progressMsg(p)
		}
	}
}

// ---- Cleaning ----

func (m Model) startCleaning() (Model, tea.Cmd) {
	m.cleanQueue = []scan.Item{}
	for _, item := range m.allItems {
		if item.Selected {
			m.cleanQueue = append(m.cleanQueue, item)
		}
	}
	m.cleanIdx = 0
	m.freed = 0
	m.state = stateCleaning
	if len(m.cleanQueue) == 0 {
		return m, func() tea.Msg { return cleanDoneMsg{} }
	}
	return m, doClean(m.cleanQueue, 0, m.dryRun)
}

func doClean(queue []scan.Item, idx int, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if !dryRun {
			err = clean.Item(queue[idx])
		}
		return cleanResultMsg{idx: idx, err: err}
	}
}

// ---- Row building ----

// buildRows rebuilds m.visRows with optional fuzzy filter applied.
func (m *Model) buildRows(filter string) {
	// group items by category
	byCategory := map[scan.Category][]int{}
	for i, item := range m.allItems {
		byCategory[item.Category] = append(byCategory[item.Category], i)
	}

	m.visRows = []listRow{}
	for _, cat := range scan.CategoryOrder {
		indices, ok := byCategory[cat]
		if !ok {
			continue
		}
		// sort by size desc within category
		sort.Slice(indices, func(a, b int) bool {
			return m.allItems[indices[a]].SizeBytes > m.allItems[indices[b]].SizeBytes
		})

		var matched []listRow
		for _, idx := range indices {
			if filter == "" || fuzzyMatch(filter, m.allItems[idx].FilterValue()) {
				matched = append(matched, listRow{kind: rowItem, itemIdx: idx})
			}
		}
		if len(matched) == 0 {
			continue
		}
		m.visRows = append(m.visRows, listRow{kind: rowHeader, header: cat})
		m.visRows = append(m.visRows, matched...)
	}
	// keep cursor in bounds
	vis := m.visibleItemRows()
	if m.cursor >= len(vis) {
		m.cursor = max(0, len(vis)-1)
	}
}

// visibleItemRows returns only rowItem rows from visRows.
func (m Model) visibleItemRows() []listRow {
	var out []listRow
	for _, r := range m.visRows {
		if r.kind == rowItem {
			out = append(out, r)
		}
	}
	return out
}

func (m *Model) scrollToCursor() {
	visible := m.listHeight()
	if m.cursor < m.viewOffset {
		m.viewOffset = m.cursor
	}
	if m.cursor >= m.viewOffset+visible {
		m.viewOffset = m.cursor - visible + 1
	}
}

func (m Model) listHeight() int {
	reserved := 10 // header + status bar + help
	h := m.height - reserved
	if h < 5 {
		return 5
	}
	return h
}

// ---- Helpers ----

func (m Model) selectedCount() int {
	n := 0
	for _, item := range m.allItems {
		if item.Selected {
			n++
		}
	}
	return n
}

func (m Model) selectedBytes() int64 {
	var total int64
	for _, item := range m.allItems {
		if item.Selected {
			total += item.SizeBytes
		}
	}
	return total
}

func (m Model) totalFoundBytes() int64 {
	var total int64
	for _, item := range m.allItems {
		total += item.SizeBytes
	}
	return total
}

func (m Model) cleanProgress() float64 {
	if len(m.cleanQueue) == 0 {
		return 1
	}
	return float64(m.cleanIdx) / float64(len(m.cleanQueue))
}

func humanBytes(b int64) string {
	return humanize.Bytes(uint64(b))
}

func fuzzyMatch(pattern, target string) bool {
	pattern = toLower(pattern)
	target = toLower(target)
	pi := 0
	for _, c := range target {
		if pi < len(pattern) && rune(pattern[pi]) == c {
			pi++
		}
	}
	return pi == len(pattern)
}

func toLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
