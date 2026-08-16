package command

import (
	"casefile/internal/core"
	"casefile/internal/model"
	"casefile/internal/tui"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Issues",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	listCmd.Flags().String("severity", "", "Filter Issues by severity")
	listCmd.Flags().String("status", "", "Filter Issues by status")
	listCmd.Flags().String("file", "", "Filter Issues by a single file")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, _ []string) error {
	state, err := core.LoadState()
	if err != nil {
		return err
	}

	severity, _ := cmd.Flags().GetString("severity")
	status, _ := cmd.Flags().GetString("status")
	file, _ := cmd.Flags().GetString("file")

	db := state.DB()

	tx, err := db.BeginTransaction()
	if err != nil {
		return err
	}

	var issues []model.Issue

	rows, err := tx.Queryx(`
		SELECT *
		FROM T_Issue
		WHERE ($1 = '' OR severity = $1)
		    AND ($2 = '' OR status = $2)
			AND ($3 = '' OR file LIKE CONCAT('%', $3, '%'))
	`, severity, status, file)

	if err != nil {
		return err
	}

	for rows.Next() {
		var issue model.Issue
		err = rows.StructScan(&issue)
		if err != nil {
			return err
		}
		issues = append(issues, issue)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	// Print results.
	for _, issue := range issues {
		fmt.Printf("%s\n", tui.IssueCell(issue))
	}

	return nil
}
