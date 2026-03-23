package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ahmedfargh/server-manager/cmd/cli/config"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Blackwater server",
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Email").Show()
		if email == "" {
			pterm.Error.Println("Email cannot be empty")
			os.Exit(1)
		}

		password, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Password").WithMask("*").Show()
		if password == "" {
			pterm.Error.Println("Password cannot be empty")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Authenticating...")

		reqBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})

		url := fmt.Sprintf("%s/login", Host)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			spinner.Fail(fmt.Sprintf("Failed to connect to server: %v", err))
			os.Exit(1)
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			spinner.Fail(fmt.Sprintf("Login failed (Status %d): %s", resp.StatusCode, string(bodyBytes)))
			os.Exit(1)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			spinner.Fail("Failed to parse server response")
			os.Exit(1)
		}

		token, ok := result["token"].(string)
		if !ok || token == "" {
			spinner.Fail("Server response did not contain a token")
			os.Exit(1)
		}

		if err := config.SaveToken(token); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to save token locally: %v", err))
			os.Exit(1)
		}

		spinner.Success("Login successful!")
		pterm.Info.Println("JWT Token saved successfully.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
