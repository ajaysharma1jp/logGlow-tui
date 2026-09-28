package main

import "github.com/charmbracelet/lipgloss"

var (
	primaryColor   = lipgloss.Color("#fabf34")
	secondaryColor = lipgloss.Color("#fbe8bc")
	textColor      = lipgloss.Color("#010101dc")
	subtleColor    = lipgloss.Color("#241")
)

var (
	headerStyle    = lipgloss.NewStyle().Bold(true).Foreground(textColor).Background(primaryColor).Padding(0, 2).MarginTop(1)
	containerStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Padding(0, 1)
	searchBoxStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(secondaryColor).Padding(0, 1)
	helpStyle      = lipgloss.NewStyle().Foreground(subtleColor).Italic(true)
)
var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	debugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)
