package main

import (
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ToastType string

const (
	ToastError   ToastType = "error"
	ToastSuccess ToastType = "success"
	ToastInfo    ToastType = "info"
	ToastWarning ToastType = "warning"
)

type Toast struct {
	Message string
	Type    ToastType
}

type showToastMsg struct {
	toast Toast
}

type dismissToastMsg struct{}

// creates a command to display a toast notification
func showToast(message string, toastType ToastType) tea.Cmd {
	return func() tea.Msg {
		// send system notification if available
		sendSystemNotification(message, toastType)

		return showToastMsg{
			toast: Toast{
				Message: message,
				Type:    toastType,
			},
		}
	}
}

// sends a notification to the system notification daemon
func sendSystemNotification(message string, toastType ToastType) {
	var urgency string
	var icon string

	switch toastType {
	case ToastError:
		urgency = "critical"
		icon = "error"
	case ToastWarning:
		urgency = "normal"
		icon = "warning"
	case ToastSuccess:
		urgency = "low"
		icon = "success"
	default:
		urgency = "normal"
		icon = "info"
	}

	// try notify-send
	cmd := exec.Command("notify-send",
		"-u", urgency,
		"-a", "got",
		"-i", icon,
		"got",
		message,
	)

	go cmd.Run()
}

// creates a command that dismisses the toast after a duration
func dismissToastAfter(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return dismissToastMsg{}
	})
}

// returns the appropriate style for a toast type
func toastStyle(toastType ToastType) lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Padding(0, 1).
		MarginTop(1).
		Bold(true)

	switch toastType {
	case ToastError:
		return baseStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("1"))
	case ToastSuccess:
		return baseStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("2"))
	case ToastWarning:
		return baseStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("3"))
	default:
		return baseStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("4"))
	}
}

// renders a toast notification
func renderToast(toast Toast) string {
	if toast.Message == "" {
		return ""
	}

	style := toastStyle(toast.Type)

	var icon string
	switch toast.Type {
	case ToastError:
		icon = "✗ "
	case ToastSuccess:
		icon = "✓ "
	case ToastWarning:
		icon = "⚠ "
	default:
		icon = "ℹ "
	}

	return style.Render(icon + toast.Message)
}
