package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "CDN log download",
}

var logListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDN log files",
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")

		result, err := executeLogList(domain, startDate, endDate)
		if err != nil {
			return err
		}
		fmt.Print(result)
		return nil
	},
}

func init() {
	logListCmd.Flags().String("domain", "", "Domain name")
	logListCmd.Flags().String("start-date", "", "Start date for log query")
	logListCmd.Flags().String("end-date", "", "End date for log query")
	_ = logListCmd.MarkFlagRequired("domain")

	logCmd.AddCommand(logListCmd)
	rootCmd.AddCommand(logCmd)
}

func executeLogList(domain, startDate, endDate string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"domain": domain}
	if startDate != "" {
		queryParams["start_time"] = startDate
	}
	if endDate != "" {
		queryParams["end_time"] = endDate
	}

	rawResp, err := client.Get("/API/cdn/domain/log", queryParams)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Filename  string `json:"filename"`
			URL       string `json:"url"`
			Timepoint string `json:"timepoint"`
			Size      int64  `json:"size"`
			MD5       string `json:"md5"`
		} `json:"data"`
		Limit   int    `json:"limit"`
		Page    int    `json:"page"`
		Total   int    `json:"total"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if resp.Code != 1 {
		msg := resp.Message
		if msg == "" {
			msg = "the server returned an error with no details. Check that the resource exists and parameters are correct"
		}
		return "", fmt.Errorf("API error (code %d): %s", resp.Code, msg)
	}

	headers := []string{"FILENAME", "TIMEPOINT", "SIZE", "MD5", "URL"}
	rows := make([][]string, 0, len(resp.Data))
	for _, l := range resp.Data {
		rows = append(rows, []string{l.Filename, l.Timepoint, fmt.Sprintf("%d", l.Size), l.MD5, l.URL})
	}
	return formatTable(headers, rows), nil
}
