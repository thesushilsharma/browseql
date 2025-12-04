package ui

import (
    // "fmt"
    "strings"
    "github.com/charmbracelet/lipgloss"
)

func formatTable(headers []string, data [][]string, width int) string {
    if len(data) == 0 {
        return "No data available"
    }

    // Calculate column widths
    colWidths := calculateColumnWidths(headers, data, width)
    
    var builder strings.Builder
    
    // Build header
    builder.WriteString(buildHeader(headers, colWidths))
    builder.WriteString(buildSeparator(colWidths))
    
    // Build rows
    for i, row := range data {
        if i == 0 { // Skip header row in data
            continue
        }
        builder.WriteString(buildRow(row, colWidths))
        if i < len(data)-1 {
            builder.WriteString("\n")
        }
    }
    
    return builder.String()
}

func calculateColumnWidths(headers []string, data [][]string, maxWidth int) []int {
    colWidths := make([]int, len(headers))
    
    // Find maximum width for each column
    for i, header := range headers {
        colWidths[i] = len(header)
    }
    
    for _, row := range data {
        for i, cell := range row {
            if len(cell) > colWidths[i] {
                colWidths[i] = len(cell)
            }
        }
    }
    
    // Ensure columns don't exceed max width
    totalWidth := 0
    for i, width := range colWidths {
        if width > 30 { // Max column width
            colWidths[i] = 30
        }
        totalWidth += colWidths[i] + 3 // +3 for padding and borders
    }
    
    return colWidths
}

func buildHeader(headers []string, colWidths []int) string {
    var builder strings.Builder
    builder.WriteString("┌")
    
    for i, width := range colWidths {
        builder.WriteString(strings.Repeat("─", width+2))
        if i < len(colWidths)-1 {
            builder.WriteString("┬")
        }
    }
    builder.WriteString("┐\n")
    
    builder.WriteString("│")
    for i, header := range headers {
        padded := padString(header, colWidths[i])
        style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
        builder.WriteString(" " + style.Render(padded) + " │")
    }
    builder.WriteString("\n")
    
    return builder.String()
}

func buildSeparator(colWidths []int) string {
    var builder strings.Builder
    builder.WriteString("├")
    for i, width := range colWidths {
        builder.WriteString(strings.Repeat("─", width+2))
        if i < len(colWidths)-1 {
            builder.WriteString("┼")
        }
    }
    builder.WriteString("┤\n")
    return builder.String()
}

func buildRow(row []string, colWidths []int) string {
    var builder strings.Builder
    builder.WriteString("│")
    for i, cell := range row {
        padded := padString(cell, colWidths[i])
        builder.WriteString(" " + padded + " │")
    }
    return builder.String()
}

func padString(s string, width int) string {
    if len(s) > width {
        return s[:width-3] + "..."
    }
    return s + strings.Repeat(" ", width-len(s))
}