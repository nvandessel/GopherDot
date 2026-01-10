package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressTracker tracks progress through a multi-step operation
type ProgressTracker struct {
	mu           sync.Mutex
	currentStep  int
	totalSteps   int
	steps        []string
	currentItem  int
	totalItems   int
	currentTask  string
	showProgress bool
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(steps []string) *ProgressTracker {
	return &ProgressTracker{
		steps:        steps,
		totalSteps:   len(steps),
		showProgress: true,
	}
}

// StartStep begins a new step
func (p *ProgressTracker) StartStep(stepNum int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if stepNum < 1 || stepNum > p.totalSteps {
		return
	}

	p.currentStep = stepNum
	p.currentItem = 0
	p.totalItems = 0

	// Print step header
	stepName := p.steps[stepNum-1]
	header := fmt.Sprintf("Step %d/%d: %s", stepNum, p.totalSteps, stepName)
	fmt.Println()
	fmt.Println(TitleStyle.Render(header))
}

// SetItemCount sets the total number of items for the current step
func (p *ProgressTracker) SetItemCount(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.totalItems = total
}

// NextItem advances to the next item
func (p *ProgressTracker) NextItem(taskName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentItem++
	p.currentTask = taskName
}

// Progress prints a progress message with optional item counter
func (p *ProgressTracker) Progress(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.showProgress {
		return
	}

	// Add item counter if we have items
	var output string
	if p.totalItems > 0 && p.currentItem > 0 {
		counter := SubtleStyle.Render(fmt.Sprintf("[%d/%d]", p.currentItem, p.totalItems))
		output = fmt.Sprintf("  %s %s", counter, msg)
	} else {
		output = fmt.Sprintf("  %s", msg)
	}

	fmt.Println(output)
}

// ProgressSuccess prints a success message with optional item counter
func (p *ProgressTracker) ProgressSuccess(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	icon := SuccessStyle.Render("✓")
	var output string
	if p.totalItems > 0 && p.currentItem > 0 {
		counter := SubtleStyle.Render(fmt.Sprintf("[%d/%d]", p.currentItem, p.totalItems))
		output = fmt.Sprintf("  %s %s %s", counter, icon, msg)
	} else {
		output = fmt.Sprintf("  %s %s", icon, msg)
	}

	fmt.Println(output)
}

// ProgressError prints an error message
func (p *ProgressTracker) ProgressError(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	icon := ErrorStyle.Render("✖")
	var output string
	if p.totalItems > 0 && p.currentItem > 0 {
		counter := SubtleStyle.Render(fmt.Sprintf("[%d/%d]", p.currentItem, p.totalItems))
		output = fmt.Sprintf("  %s %s %s", counter, icon, msg)
	} else {
		output = fmt.Sprintf("  %s %s", icon, msg)
	}

	fmt.Println(output)
}

// ProgressWarning prints a warning message
func (p *ProgressTracker) ProgressWarning(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	icon := WarningStyle.Render("⚠")
	var output string
	if p.totalItems > 0 && p.currentItem > 0 {
		counter := SubtleStyle.Render(fmt.Sprintf("[%d/%d]", p.currentItem, p.totalItems))
		output = fmt.Sprintf("  %s %s %s", counter, icon, msg)
	} else {
		output = fmt.Sprintf("  %s %s", icon, msg)
	}

	fmt.Println(output)
}

// ProgressSkip prints a skip message
func (p *ProgressTracker) ProgressSkip(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	icon := SubtleStyle.Render("⊘")
	var output string
	if p.totalItems > 0 && p.currentItem > 0 {
		counter := SubtleStyle.Render(fmt.Sprintf("[%d/%d]", p.currentItem, p.totalItems))
		output = fmt.Sprintf("  %s %s %s", counter, icon, msg)
	} else {
		output = fmt.Sprintf("  %s %s", icon, msg)
	}

	fmt.Println(output)
}

// StepSummary prints a summary for the current step
func (p *ProgressTracker) StepSummary(success, failed, skipped int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var parts []string

	if success > 0 {
		parts = append(parts, SuccessStyle.Render(fmt.Sprintf("%d succeeded", success)))
	}
	if failed > 0 {
		parts = append(parts, ErrorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	if skipped > 0 {
		parts = append(parts, SubtleStyle.Render(fmt.Sprintf("%d skipped", skipped)))
	}

	if len(parts) > 0 {
		fmt.Printf("  %s\n", strings.Join(parts, ", "))
	}
}

// progressBarModel is a Bubbletea model for showing a progress bar with spinner
type progressBarModel struct {
	progress   progress.Model
	spinner    spinner.Model
	percent    float64
	message    string
	done       bool
	err        error
	updateChan chan progressUpdate
	doneChan   chan error
	width      int
}

type progressUpdate struct {
	percent float64
	message string
}

type progressDoneMsg struct {
	err error
}

func newProgressBarModel(msg string, updateChan chan progressUpdate, doneChan chan error) progressBarModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(PrimaryColor)

	return progressBarModel{
		progress:   p,
		spinner:    s,
		message:    msg,
		updateChan: updateChan,
		doneChan:   doneChan,
		width:      40,
	}
}

func (m progressBarModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForUpdate(m.updateChan),
		waitForDone(m.doneChan),
	)
}

func waitForUpdate(ch chan progressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-ch
		if !ok {
			return nil
		}
		return update
	}
}

func waitForDone(ch chan error) tea.Cmd {
	return func() tea.Msg {
		err := <-ch
		return progressDoneMsg{err: err}
	}
}

func (m progressBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width - 10
		if m.width > 60 {
			m.width = 60
		}
		if m.width < 20 {
			m.width = 20
		}
		m.progress = progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(m.width),
		)

	case progressUpdate:
		m.percent = msg.percent
		if msg.message != "" {
			m.message = msg.message
		}
		return m, waitForUpdate(m.updateChan)

	case progressDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m progressBarModel) View() string {
	if m.done {
		if m.err != nil {
			return ErrorStyle.Render("✖") + " " + m.message + ": " + m.err.Error() + "\n"
		}
		return ""
	}

	// Show spinner with message
	str := fmt.Sprintf("%s %s", m.spinner.View(), m.message)

	// Show progress bar if we have progress
	if m.percent > 0 {
		str += "\n" + m.progress.ViewAs(m.percent)
	}

	return str + "\n"
}

// RunWithProgress runs a long-running task with a progress bar
// The task function receives channels to send progress updates
func RunWithProgress(msg string, task func(updateChan chan<- progressUpdate) error) error {
	updateChan := make(chan progressUpdate, 10)
	doneChan := make(chan error, 1)

	// Run the task in a goroutine
	go func() {
		err := task(updateChan)
		close(updateChan)
		doneChan <- err
	}()

	p := tea.NewProgram(newProgressBarModel(msg, updateChan, doneChan))
	m, err := p.Run()
	if err != nil {
		return err
	}

	if model, ok := m.(progressBarModel); ok && model.err != nil {
		return model.err
	}

	Success("%s", msg)
	return nil
}

// FormatProgress formats a progress message with an item counter
func FormatProgress(current, total int, msg string) string {
	if total > 0 && current > 0 {
		return fmt.Sprintf("[%d/%d] %s", current, total, msg)
	}
	return msg
}

// FormatProgressWithIcon formats a progress message with an icon and counter
func FormatProgressWithIcon(icon string, current, total int, msg string) string {
	if total > 0 && current > 0 {
		counter := fmt.Sprintf("[%d/%d]", current, total)
		return fmt.Sprintf("%s %s %s", icon, counter, msg)
	}
	return fmt.Sprintf("%s %s", icon, msg)
}
