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

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System hardware monitoring commands",
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system RAM and CPU status",
	Run: func(cmd *cobra.Command, args []string) {
		token, err := config.LoadToken()
		if err != nil || token == "" {
			pterm.Error.Println("Not authenticated. Please run 'bwcli login' first.")
			os.Exit(1)
		}

		spinner, _ := pterm.DefaultSpinner.Start("Fetching system metrics...")

		// Fetch CPU
		cpuReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/cpu", Host), nil)
		cpuReq.Header.Add("Authorization", "Bearer "+token)

		client := &http.Client{}
		cpuResp, err := client.Do(cpuReq)
		if err != nil || cpuResp.StatusCode != http.StatusOK {
			spinner.Fail("Failed to fetch CPU data")
			os.Exit(1)
		}
		defer cpuResp.Body.Close()
		cpuBody, _ := io.ReadAll(cpuResp.Body)

		var cpuData map[string]interface{}
		json.Unmarshal(cpuBody, &cpuData)

		// Fetch RAM
		ramReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/ram", Host), nil)
		ramReq.Header.Add("Authorization", "Bearer "+token)

		ramResp, err := client.Do(ramReq)
		if err != nil || ramResp.StatusCode != http.StatusOK {
			spinner.Fail("Failed to fetch RAM data")
			os.Exit(1)
		}
		defer ramResp.Body.Close()
		ramBody, _ := io.ReadAll(ramResp.Body)

		var ramData map[string]interface{}
		json.Unmarshal(ramBody, &ramData)

		spinner.Success("System metrics fetched!")

		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("System Status")

		// Render CPU
		pterm.DefaultSection.Println("CPU Architecture & Cores")
		cpuCore := int(cpuData["Logical_core"].(float64))
		cpuArch := cpuData["Arch"].(string)
		cpuOs := cpuData["Os"].(string)

		pterm.Info.Printf("OS: %s | Arch: %s | Logical Cores: %d\n\n", cpuOs, cpuArch, cpuCore)

		if hardwareInfo, ok := cpuData["Cpu_Hard_Ware_Info"].(map[string]interface{}); ok {
			var tableData [][]string
			tableData = append(tableData, []string{"Core/Thread", "Model", "Physical Cores", "MHz"})
			for k, v := range hardwareInfo {
				hwMap := v.(map[string]interface{})
				
				modelName := "Unknown"
				if m, ok := hwMap["model_name"].(string); ok {
					modelName = m
				}
				cores := "Unknown"
				if c, ok := hwMap["physical_cores"].(string); ok {
					cores = c
				}
				mhz := "Unknown"
				if m, ok := hwMap["mhz"].(string); ok {
					mhz = m
				}
				
				tableData = append(tableData, []string{
					k,
					modelName,
					cores,
					mhz,
				})
			}
			pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		}

		// Render RAM
		pterm.DefaultSection.Println("Memory Metrics (MB)")
		
		var usedRam, freeRam, totalRam int
		if virtualInfo, ok := ramData["Vertiual_info"].(map[string]interface{}); ok {
			totalRam = int(virtualInfo["Total_memory"].(float64))
			usedRam = int(virtualInfo["Used_memory"].(float64))
			freeRam = int(virtualInfo["Free_memory"].(float64))
		}

		var usedSwap, freeSwap, totalSwap int
		if swapInfo, ok := ramData["SwapInfo"].(map[string]interface{}); ok {
			totalSwap = int(swapInfo["Total_memory"].(float64))
			usedSwap = int(swapInfo["Used_memory"].(float64))
			freeSwap = int(swapInfo["Free_memory"].(float64))
		}

		pterm.DefaultBarChart.WithHorizontal().WithBars([]pterm.Bar{
			{Label: "RAM Used", Value: usedRam, Style: pterm.NewStyle(pterm.FgLightRed)},
			{Label: "RAM Free", Value: freeRam, Style: pterm.NewStyle(pterm.FgLightGreen)},
			{Label: "RAM Total", Value: totalRam, Style: pterm.NewStyle(pterm.FgCyan)},
			{Label: "Swap Used", Value: usedSwap, Style: pterm.NewStyle(pterm.FgLightYellow)},
			{Label: "Swap Free", Value: freeSwap, Style: pterm.NewStyle(pterm.FgLightGreen)},
			{Label: "Swap Total", Value: totalSwap, Style: pterm.NewStyle(pterm.FgCyan)},
		}).Render()
	},
}

func init() {
	systemCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(systemCmd)
}
