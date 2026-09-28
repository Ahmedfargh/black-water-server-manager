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

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH key management and authorization",
}

// Subcommand: List SSH keys
var sshListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all managed SSH keys",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetching SSH keys...")

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/ssh/keys/list?limit=100", Host), nil)
		req.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			spinner.Fail("Failed to fetch SSH keys")
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var result struct {
			Keys  []map[string]interface{} `json:"keys"`
			Total int                      `json:"total"`
			Page  int                      `json:"page"`
			Limit int                      `json:"limit"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			spinner.Fail("Failed to parse server response")
			os.Exit(1)
		}

		spinner.Success(fmt.Sprintf("Found %d SSH key(s)!", len(result.Keys)))

		if len(result.Keys) == 0 {
			pterm.Info.Println("No SSH keys found. Use 'bwcli ssh gen' or 'bwcli ssh import' to add keys.")
			return
		}

		var tableData [][]string
		tableData = append(tableData, []string{"ID", "Name", "Type", "Fingerprint", "Comment", "Authorized in Host"})

		for _, k := range result.Keys {
			id := fmt.Sprintf("%.0f", k["id"].(float64))
			name, _ := k["name"].(string)
			keyType, _ := k["key_type"].(string)
			fingerprint, _ := k["fingerprint"].(string)
			comment, _ := k["comment"].(string)
			added, _ := k["added_to_authorized_keys"].(bool)

			authStr := pterm.FgLightRed.Sprint("Disabled")
			if added {
				authStr = pterm.FgLightGreen.Sprint("Authorized")
			}

			tableData = append(tableData, []string{id, name, keyType, fingerprint, comment, authStr})
		}

		pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	},
}

// Subcommand: Generate SSH Key
var (
	genName      string
	genType      string
	genComment   string
	genAuthorize bool
	genSaveFile  string
	genRsaBits   int
)

var sshGenCmd = &cobra.Command{
	Use:     "gen",
	Aliases: []string{"generate"},
	Short:   "Generate a new ED25519 or RSA key pair",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		if genName == "" {
			pterm.Error.Println("Key name is required (--name)")
			os.Exit(1)
		}

		reqBody := map[string]interface{}{
			"name":                   genName,
			"key_type":               genType,
			"comment":                genComment,
			"add_to_authorized_keys": genAuthorize,
			"rsa_bits":               genRsaBits,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		spinner, _ := pterm.DefaultSpinner.Start("Generating SSH key pair...")

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/ssh/keys/generate", Host), bytes.NewBuffer(jsonBytes))
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			spinner.Fail(fmt.Sprintf("Failed to generate SSH key: %s", string(body)))
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		var res struct {
			Message       string                 `json:"message"`
			Key           map[string]interface{} `json:"key"`
			PrivateKeyPEM string                 `json:"private_key_pem"`
		}
		json.Unmarshal(body, &res)

		spinner.Success("SSH Key Pair generated successfully!")

		pterm.DefaultSection.Println("Key Details")
		pterm.Info.Printf("Name: %s\nType: %s\nFingerprint: %s\n\n", res.Key["name"], res.Key["key_type"], res.Key["fingerprint"])

		pterm.DefaultSection.Println("Public Key")
		pterm.Println(pterm.FgLightCyan.Sprint(res.Key["public_key"]))

		pterm.DefaultSection.Println("Private Key (PEM)")
		pterm.Warning.Println("⚠️  Save this private key immediately. It will not be shown again!")
		pterm.Println(pterm.FgLightYellow.Sprint(res.PrivateKeyPEM))

		if genSaveFile != "" {
			if err := os.WriteFile(genSaveFile, []byte(res.PrivateKeyPEM), 0600); err != nil {
				pterm.Error.Printf("Failed to save private key to file %s: %v\n", genSaveFile, err)
			} else {
				pterm.Success.Printf("Private key saved to %s (permissions: 0600)\n", genSaveFile)
			}
		}
	},
}

// Subcommand: Import Public Key
var (
	importName      string
	importPubKey    string
	importComment   string
	importAuthorize bool
)

var sshImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an existing SSH public key",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		if importName == "" || importPubKey == "" {
			pterm.Error.Println("Both --name and --pubkey are required")
			os.Exit(1)
		}

		reqBody := map[string]interface{}{
			"name":                   importName,
			"public_key":             importPubKey,
			"comment":                importComment,
			"add_to_authorized_keys": importAuthorize,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		spinner, _ := pterm.DefaultSpinner.Start("Importing SSH public key...")

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/ssh/keys/import", Host), bytes.NewBuffer(jsonBytes))
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			spinner.Fail(fmt.Sprintf("Failed to import key: %s", string(body)))
			os.Exit(1)
		}
		defer resp.Body.Close()

		spinner.Success("SSH Public Key imported successfully!")
	},
}

// Subcommand: Toggle Authorization
var (
	toggleKeyID uint
)

var sshToggleCmd = &cobra.Command{
	Use:     "toggle-auth",
	Aliases: []string{"toggle"},
	Short:   "Toggle SSH key in host authorized_keys",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		if toggleKeyID == 0 {
			pterm.Error.Println("Please specify key ID with --id")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Toggling authorization...")

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/ssh/keys/%d/toggle-auth", Host, toggleKeyID), nil)
		req.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			spinner.Fail(fmt.Sprintf("Failed to toggle authorization: %s", string(body)))
			os.Exit(1)
		}
		defer resp.Body.Close()

		var res struct {
			Message               string `json:"message"`
			AddedToAuthorizedKeys bool   `json:"added_to_authorized_keys"`
		}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &res)

		if res.AddedToAuthorizedKeys {
			spinner.Success(fmt.Sprintf("Key ID %d is now AUTHORIZED on the host!", toggleKeyID))
		} else {
			spinner.Success(fmt.Sprintf("Key ID %d was REMOVED from host authorized_keys.", toggleKeyID))
		}
	},
}

// Subcommand: Delete Key
var (
	deleteKeyID uint
)

var sshDeleteCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"delete", "remove"},
	Short:   "Delete an SSH key and remove from authorized_keys",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		if deleteKeyID == 0 {
			pterm.Error.Println("Please specify key ID with --id")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Deleting SSH key...")

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/ssh/keys/%d", Host, deleteKeyID), nil)
		req.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			spinner.Fail(fmt.Sprintf("Failed to delete SSH key: %s", string(body)))
			os.Exit(1)
		}
		defer resp.Body.Close()

		spinner.Success(fmt.Sprintf("Key ID %d deleted successfully!", deleteKeyID))
	},
}

func init() {
	// Flags for generate
	sshGenCmd.Flags().StringVarP(&genName, "name", "n", "", "Key name / label (required)")
	sshGenCmd.Flags().StringVarP(&genType, "type", "t", "ed25519", "Key type: ed25519 or rsa")
	sshGenCmd.Flags().StringVarP(&genComment, "comment", "c", "", "Key comment")
	sshGenCmd.Flags().BoolVarP(&genAuthorize, "authorize", "a", true, "Automatically add to host authorized_keys")
	sshGenCmd.Flags().StringVarP(&genSaveFile, "save-file", "s", "", "Path to save generated private key PEM")
	sshGenCmd.Flags().IntVarP(&genRsaBits, "bits", "b", 4096, "RSA key size in bits (if type is rsa)")

	// Flags for import
	sshImportCmd.Flags().StringVarP(&importName, "name", "n", "", "Key name / label (required)")
	sshImportCmd.Flags().StringVarP(&importPubKey, "pubkey", "k", "", "Public key string (e.g. ssh-ed25519 AAAA...) (required)")
	sshImportCmd.Flags().StringVarP(&importComment, "comment", "c", "", "Key comment")
	sshImportCmd.Flags().BoolVarP(&importAuthorize, "authorize", "a", true, "Automatically add to host authorized_keys")

	// Flags for toggle & delete
	sshToggleCmd.Flags().UintVarP(&toggleKeyID, "id", "i", 0, "SSH key ID (required)")
	sshDeleteCmd.Flags().UintVarP(&deleteKeyID, "id", "i", 0, "SSH key ID (required)")

	sshCmd.AddCommand(sshListCmd)
	sshCmd.AddCommand(sshGenCmd)
	sshCmd.AddCommand(sshImportCmd)
	sshCmd.AddCommand(sshToggleCmd)
	sshCmd.AddCommand(sshDeleteCmd)

	rootCmd.AddCommand(sshCmd)
}
