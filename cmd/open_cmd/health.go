package open_cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health",
	Long:  `Check if the CashLenX system is running and healthy`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if the server is running by making a request to the health endpoint
		client := http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get("http://localhost:8080/api/open/health")
		if err != nil {
			fmt.Println("❌ System is not healthy: Server is not reachable")
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅ System is healthy")
			fmt.Printf("Status Code: %d %s\n", resp.StatusCode, resp.Status)
		} else {
			fmt.Printf("❌ System is not healthy: %d %s\n", resp.StatusCode, resp.Status)
		}
	},
}
