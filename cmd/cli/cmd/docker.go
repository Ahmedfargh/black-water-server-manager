package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ahmedfargh/server-manager/cmd/cli/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker management commands",
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all docker containers",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetching containers...")

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/docker/containers", Host), nil)
		req.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			spinner.Fail("Failed to fetch Docker containers")
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var containers []map[string]interface{}
		json.Unmarshal(body, &containers)

		spinner.Success(fmt.Sprintf("Found %d containers!", len(containers)))

		// Render table
		if len(containers) == 0 {
			pterm.Info.Println("No Docker containers found.")
			return
		}

		var tableData [][]string
		tableData = append(tableData, []string{"ID (Short)", "Names", "Image", "State", "Status"})

		for _, c := range containers {
			id := ""
			if cID, ok := c["id"].(string); ok && len(cID) >= 12 {
				id = cID[:12]
			} else if ok {
				id = cID
			}

			// Format names
			namesStr := ""
			if names, ok := c["names"].([]interface{}); ok {
				for i, n := range names {
					if nameStr, ok := n.(string); ok {
						if i > 0 {
							namesStr += ", "
						}
						namesStr += nameStr
					}
				}
			}

			image, _ := c["image"].(string)
			state, _ := c["state"].(string)
			status, _ := c["status"].(string)

			// Simple color coding for state
			if state == "running" {
				state = pterm.FgLightGreen.Sprint(state)
			} else if state == "exited" {
				state = pterm.FgLightRed.Sprint(state)
			} else {
				state = pterm.FgLightYellow.Sprint(state)
			}

			tableData = append(tableData, []string{id, namesStr, image, state, status})
		}

		pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	},
}

func init() {
	dockerCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(dockerCmd)
}
