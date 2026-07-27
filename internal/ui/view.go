package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"broom/internal/scan"
)

// ---- Main View dispatcher ----

func (m Model) View() string {
	switch m.state {
	case stateWelcome:
		return m.welcomeView()
	case stateScanning:
		return m.scanningView()
	case stateSelect:
		return m.selectView()
	case stateConfirm:
		return m.confirmView()
	case stateCleaning:
		return m.cleaningView()
	case stateDone:
		return m.doneView()
	}
	return ""
}

// ---- Welcome ----

func (m Model) welcomeView() string {
	w := m.width
	if w == 0 {
		w = 80
	}

	header := styleLogo.Render(logo)
	dryTag := ""
	if m.dryRun {
		dryTag = "  " + styleWarning.Render("[DRY RUN — nothing will be deleted]")
	}
	sub := styleMuted.Render(tagline) + dryTag

	categories := []string{
		"  • node_modules & Python venvs",
		"  • Rust target dirs & Cargo cache",
		"  • Build output (.next, dist, build…)",
		"  • LLM model files (HuggingFace, LM Studio, Ollama)",
		"  • Package caches (npm, uv, pnpm, Homebrew…)",
		"  • Docker / Podman images & volumes",
		"  • Xcode DerivedData & simulators",
		"  • App caches (Puppeteer, Playwright…)",
	}

	body := strings.Join([]string{
		header,
		sub,
		"",
		styleBold.Render("Scans for:"),
		strings.Join(categories, "\n"),
		"",
		styleSuccess.Render("[ press enter to start scan ]"),
		styleMuted.Render("  q to quit"),
	}, "\n")

	box := styleBox.Width(w - 4).Render(body)
	return lipgloss.Place(w, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---- Scanning ----

func (m Model) scanningView() string {
	var sb strings.Builder
	sb.WriteString(styleLogo.Render(logo) + "\n\n")

	scannerNames := []string{
		"docker/podman", "llm models", "caches",
		"node_modules", "python venvs", "rust artifacts", "build output",
	}
	for _, name := range scannerNames {
		p, done := m.scanProgress[name], false
		if prog, ok := m.scanProgress[name]; ok {
			done = prog.Done
			p = prog
		}
		if done {
			sb.WriteString(styleSuccess.Render("✓ ") + fmt.Sprintf("%-20s", name) +
				styleMuted.Render(fmt.Sprintf("found %d", p.Found)) + "\n")
		} else {
			sb.WriteString(m.spinner.View() + " " + fmt.Sprintf("%-20s", name) +
				styleMuted.Render("scanning…") + "\n")
		}
	}

	if len(m.allItems) > 0 {
		sb.WriteString("\n" + styleWarning.Render(
			fmt.Sprintf("Found so far: %s across %d items",
				humanBytes(m.totalFoundBytes()), len(m.allItems))))
	}

	return styleBox.Width(m.width - 4).Render(sb.String())
}

// ---- Select ----

func (m Model) selectView() string {
	var sb strings.Builder

	// header bar
	total := humanBytes(m.totalFoundBytes())
	selCount := m.selectedCount()
	selSize := humanBytes(m.selectedBytes())

	headerLine := styleLogo.Render(logo) + "  " +
		styleMuted.Render(fmt.Sprintf("%d items  •  %s recoverable", len(m.allItems), total))
	sb.WriteString(headerLine + "\n")

	// filter bar
	if m.filterMode {
		sb.WriteString(styleWarning.Render("/") + " " + m.filter + "█\n")
	} else if m.filter != "" {
		sb.WriteString(styleMuted.Render("filter: ") + styleWarning.Render(m.filter) +
			styleMuted.Render("  (/ to edit, esc to clear)") + "\n")
	} else {
		sb.WriteString(styleMuted.Render("/ to filter") + "\n")
	}
	sb.WriteString("\n")

	// list
	listH := m.listHeight()
	visItems := m.visibleItemRows()

	// build cursor map: itemIdx → cursor position among item rows
	cursorSet := map[int]bool{}
	if m.cursor < len(visItems) {
		cursorSet[visItems[m.cursor].itemIdx] = true
	}

	// render visible rows (with scroll window)
	rendered := 0
	skipped := 0
	for _, row := range m.visRows {
		if rendered >= listH {
			break
		}
		switch row.kind {
		case rowHeader:
			if skipped < m.viewOffset {
				skipped++
				continue
			}
			sb.WriteString(styleCategoryHeader.Render(string(row.header)) + "\n")
			rendered++

		case rowItem:
			if skipped < m.viewOffset {
				skipped++
				continue
			}
			item := m.allItems[row.itemIdx]
			checkbox := "[ ]"
			checkStyle := styleMuted
			if item.Selected {
				checkbox = "[✓]"
				checkStyle = styleSelected
			}

			line := checkStyle.Render(checkbox) + " "
			nameW := m.width - 32
			if nameW < 20 {
				nameW = 20
			}
			name := truncate(item.DisplayName, nameW)
			namePad := fmt.Sprintf("%-*s", nameW, name)

			if cursorSet[row.itemIdx] {
				line += styleCursor.Render(" "+namePad+" ") + " "
			} else {
				line += namePad + " "
			}
			line += styleSize.Render(item.HumanSize()) + " "
			line += styleAge.Render(item.AgeString())

			sb.WriteString(line + "\n")
			rendered++
		}
	}

	// status bar
	sb.WriteString("\n")
	if selCount > 0 {
		sb.WriteString(styleStatusBar.Render(
			fmt.Sprintf("Selected: %d items  •  %s", selCount, selSize)))
	} else {
		sb.WriteString(styleMuted.Render("space to select, a to select all"))
	}
	sb.WriteString("\n")

	// help
	help := "↑↓/jk navigate  space toggle  a all  A none  / filter  enter proceed  q quit"
	sb.WriteString(styleHelp.Render(help))

	return sb.String()
}

// ---- Confirm ----

func (m Model) confirmView() string {
	selCount := m.selectedCount()
	selSize := humanBytes(m.selectedBytes())

	warning := styleDanger.Render("  This cannot be undone.")
	if m.dryRun {
		warning = styleWarning.Render("  DRY RUN — nothing will actually be deleted.")
	}
	lines := []string{
		styleLogo.Render(logo),
		"",
		styleBold.Render("Ready to delete:"),
		"",
		styleWarning.Render(fmt.Sprintf("  %d items  •  %s", selCount, selSize)),
		"",
		warning,
		"",
		styleBold.Render("  Press y to confirm, n to go back"),
	}

	body := strings.Join(lines, "\n")
	box := styleBox.Width(50).Render(body)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---- Cleaning ----

func (m Model) cleaningView() string {
	pct := m.cleanProgress()
	barW := 40
	filled := int(pct * float64(barW))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)

	var sb strings.Builder
	sb.WriteString(styleLogo.Render(logo) + "\n\n")
	sweepLabel := styleBold.Render("Sweeping…")
	if m.dryRun {
		sweepLabel = styleBold.Render("Simulating… ") + styleWarning.Render("[DRY RUN]")
	}
	sb.WriteString(sweepLabel + "\n\n")
	sb.WriteString(stylePrimary.Render(bar) +
		styleMuted.Render(fmt.Sprintf(" %d%%", int(pct*100))) + "\n\n")

	// show last few cleaned items
	start := m.cleanIdx - 3
	if start < 0 {
		start = 0
	}
	for i := start; i < m.cleanIdx && i < len(m.cleanQueue); i++ {
		sb.WriteString(styleSuccess.Render("✓ ") + truncate(m.cleanQueue[i].DisplayName, 50) + "\n")
	}
	if m.cleanIdx < len(m.cleanQueue) {
		sb.WriteString(m.spinner.View() + " " + truncate(m.cleanQueue[m.cleanIdx].DisplayName, 50) + "\n")
	}

	sb.WriteString("\n" + styleWarning.Render("Freed so far: "+humanBytes(m.freed)))

	return styleBox.Width(m.width - 4).Render(sb.String())
}

// ---- Done ----

func (m Model) doneView() string {
	var sb strings.Builder

	if len(m.allItems) == 0 {
		sb.WriteString(styleSuccess.Render("✨ Nothing to clean — your machine is already tidy!"))
	} else {
		freed := m.freed
		if m.dryRun {
			// in dry-run, sum selected item sizes since nothing was actually deleted
			for _, item := range m.cleanQueue {
				freed += item.SizeBytes
			}
		} else if freed == 0 && m.finalFree > m.initialFree {
			freed = m.finalFree - m.initialFree
		}

		doneLabel := styleSuccess.Render("✨ All done!")
		freedLabel := styleBold.Render("Freed  ") + styleWarning.Render(humanBytes(freed))
		if m.dryRun {
			doneLabel = styleWarning.Render("✨ Dry run complete — nothing was deleted.")
			freedLabel = styleBold.Render("Would free  ") + styleWarning.Render(humanBytes(freed))
		}
		sb.WriteString(doneLabel + "\n\n")
		sb.WriteString(freedLabel + "\n\n")

		if m.initialFree > 0 {
			sb.WriteString(styleMuted.Render("Before  ") + humanBytes(m.initialFree) + " free\n")
		}
		if m.finalFree > 0 {
			sb.WriteString(styleMuted.Render("After   ") + humanBytes(m.finalFree) + " free\n")
		}
		if len(m.cleanErrs) > 0 {
			sb.WriteString("\n" + styleWarning.Render(
				fmt.Sprintf("%d errors (some items may not have been deleted)", len(m.cleanErrs))))
		}
	}

	sb.WriteString("\n" + styleMuted.Render("press any key to quit"))

	box := styleBox.Width(50).Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// ---- Helpers ----

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// categoryIcon returns an emoji for a category.
func categoryIcon(cat scan.Category) string {
	icons := map[scan.Category]string{
		scan.CategoryDocker:       "🐳",
		scan.CategoryLLMModel:     "🤖",
		scan.CategoryPackageCache: "📦",
		scan.CategoryNodeModules:  "⬡",
		scan.CategoryPythonVenv:   "🐍",
		scan.CategoryRustArtifact: "🦀",
		scan.CategoryBuildOutput:  "🏗",
		scan.CategoryXcode:        "🍎",
		scan.CategoryAppCache:     "🗑",
	}
	if icon, ok := icons[cat]; ok {
		return icon + " "
	}
	return ""
}
