package tui

import (
	"fmt"

	"casefile/internal/model"

	"github.com/fatih/color"
)

func IssueCell(issue model.Issue) string {
	var (
		severityColors = map[model.Severity]*color.Color{
			model.SeverityLow:      color.New(color.FgBlue),
			model.SeverityMedium:   color.New(color.FgYellow),
			model.SeverityHigh:     color.New(color.FgHiRed),
			model.SeverityCritical: color.New(color.FgRed, color.Bold),
		}
		statusColors = map[model.Status]*color.Color{
			model.StatusOpen:   color.New(color.FgGreen),
			model.StatusClosed: color.New(color.FgHiBlack),
		}
		dimColor  = color.New(color.FgHiBlack)
		idColor   = color.New(color.FgHiBlack)
		titleBold = color.New(color.Bold)
	)

	dot := dimColor.Sprint("•")

	header := fmt.Sprintf(
		"%s    %s",
		idColor.Sprintf("#%04d", issue.ID),
		titleBold.Sprint(issue.Title),
	)

	detail := fmt.Sprintf(
		"         %s %s %s %s %s",
		statusColors[issue.Status].Sprint(issue.Status),
		dot,
		severityColors[issue.Severity].Sprint(issue.Severity),
		dot,
		dimColor.Sprintf("%s:%d", issue.File, issue.Line),
	)

	return header + "\n" + detail
}
