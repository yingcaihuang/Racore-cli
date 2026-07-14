package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// statsQueryParams represents the common request body for POST stats commands.
type statsQueryParams struct {
	StartTime string   `json:"start_time"`
	EndTime   string   `json:"end_time"`
	Domains   []string `json:"domains,omitempty"`
	Interval  string   `json:"interval,omitempty"`
}

// statsTopQueryParams represents the request body for top-url, top-referer, top-ua endpoints.
// These endpoints do NOT accept a domains parameter.
type statsTopQueryParams struct {
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Scope     string `json:"scope,omitempty"`
	Sorted    string `json:"sorted,omitempty"`
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "CDN statistics and analytics",
	Example: `  # Query flow statistics
  racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com

  # Query request count
  racore-cli stats request --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com

  # Top URLs (by scope)
  racore-cli stats top-url --scope yesterday
  racore-cli stats top-referer --scope month

  # ISO country codes
  racore-cli stats iso-country`,
}

// --- stats flow ---

var statsFlowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Query CDN flow statistics",
	Example: `  racore-cli stats flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")
		interval, _ := cmd.Flags().GetString("interval")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsFlow(startTime, endTime, domains, interval)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsFlow(startTime, endTime string, domains []string, interval string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
		Interval:  interval,
	}
	rawResp, err := client.Post("/API/cdn/statistics/flow", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats request ---

var statsRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Query CDN request statistics",
	Example: `  racore-cli stats request --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")
		interval, _ := cmd.Flags().GetString("interval")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsRequest(startTime, endTime, domains, interval)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsRequest(startTime, endTime string, domains []string, interval string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
		Interval:  interval,
	}
	rawResp, err := client.Post("/API/cdn/statistics/request", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats hit-flow ---

var statsHitFlowCmd = &cobra.Command{
	Use:   "hit-flow",
	Short: "Query CDN hit flow statistics",
	Example: `  racore-cli stats hit-flow --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsHitFlow(startTime, endTime, domains)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsHitFlow(startTime, endTime string, domains []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
	}
	rawResp, err := client.Post("/API/cdn/statistics/hit/flow", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats hit-request ---

var statsHitRequestCmd = &cobra.Command{
	Use:   "hit-request",
	Short: "Query CDN hit request statistics",
	Example: `  racore-cli stats hit-request --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsHitRequest(startTime, endTime, domains)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsHitRequest(startTime, endTime string, domains []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
	}
	rawResp, err := client.Post("/API/cdn/statistics/hit/request", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats http-code ---

var statsHTTPCodeCmd = &cobra.Command{
	Use:   "http-code",
	Short: "Query CDN HTTP status code statistics",
	Example: `  racore-cli stats http-code --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsHTTPCode(startTime, endTime, domains)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsHTTPCode(startTime, endTime string, domains []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
	}
	rawResp, err := client.Post("/API/cdn/statistics/http/code", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats http-code-detail ---

var statsHTTPCodeDetailCmd = &cobra.Command{
	Use:   "http-code-detail",
	Short: "Query CDN HTTP status code detail statistics",
	Example: `  racore-cli stats http-code-detail --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsHTTPCodeDetail(startTime, endTime, domains)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsHTTPCodeDetail(startTime, endTime string, domains []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
	}
	rawResp, err := client.Post("/API/cdn/statistics/http/code/detail", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats district ---

var statsDistrictCmd = &cobra.Command{
	Use:   "district",
	Short: "Query CDN district statistics",
	Example: `  racore-cli stats district --start-time "2026-07-01" --end-time "2026-07-14" --domains example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		domainsFlag, _ := cmd.Flags().GetString("domains")

		var domains []string
		if domainsFlag != "" {
			domains = strings.Split(domainsFlag, ",")
		}

		result, err := executeStatsDistrict(startTime, endTime, domains)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsDistrict(startTime, endTime string, domains []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Domains:   domains,
	}
	rawResp, err := client.Post("/API/cdn/statistics/district", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats iso-country ---

var statsISOCountryCmd = &cobra.Command{
	Use:     "iso-country",
	Short:   "List ISO country codes",
	Example: `  racore-cli stats iso-country`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := executeStatsISOCountry()
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsISOCountry() (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	rawResp, err := client.Get("/API/cdn/statistics/iso/country", nil)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats top-domain ---

var statsTopDomainCmd = &cobra.Command{
	Use:   "top-domain",
	Short: "Query top domains by traffic",
	Example: `  racore-cli stats top-domain --start-time "2026-07-01" --end-time "2026-07-14"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")

		result, err := executeStatsTopDomain(startTime, endTime)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsTopDomain(startTime, endTime string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
	}
	rawResp, err := client.Post("/API/cdn/statistics/top/domain", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats top-url ---

var statsTopURLCmd = &cobra.Command{
	Use:   "top-url",
	Short: "Query top URLs by traffic",
	Example: `  # By scope
  racore-cli stats top-url --scope yesterday

  # By date range
  racore-cli stats top-url --start-time "2026-07-01 00:00" --end-time "2026-07-14 00:00"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		scope, _ := cmd.Flags().GetString("scope")
		sorted, _ := cmd.Flags().GetString("sorted")

		result, err := executeStatsTopURL(startTime, endTime, scope, sorted)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsTopURL(startTime, endTime, scope, sorted string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsTopQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Scope:     scope,
		Sorted:    sorted,
	}
	rawResp, err := client.Post("/API/cdn/domain/top/url", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats top-referer ---

var statsTopRefererCmd = &cobra.Command{
	Use:   "top-referer",
	Short: "Query top referers by traffic",
	Example: `  racore-cli stats top-referer --scope month --sorted referer_count`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		scope, _ := cmd.Flags().GetString("scope")
		sorted, _ := cmd.Flags().GetString("sorted")

		result, err := executeStatsTopReferer(startTime, endTime, scope, sorted)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsTopReferer(startTime, endTime, scope, sorted string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsTopQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Scope:     scope,
		Sorted:    sorted,
	}
	rawResp, err := client.Post("/API/cdn/domain/top/referer", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- stats top-ua ---

var statsTopUACmd = &cobra.Command{
	Use:   "top-ua",
	Short: "Query top user agents by traffic",
	Example: `  racore-cli stats top-ua --scope yesterday`,
	RunE: func(cmd *cobra.Command, args []string) error {
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		scope, _ := cmd.Flags().GetString("scope")
		sorted, _ := cmd.Flags().GetString("sorted")

		result, err := executeStatsTopUA(startTime, endTime, scope, sorted)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func executeStatsTopUA(startTime, endTime, scope, sorted string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := statsTopQueryParams{
		StartTime: startTime,
		EndTime:   endTime,
		Scope:     scope,
		Sorted:    sorted,
	}
	rawResp, err := client.Post("/API/cdn/domain/top/ua", body)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}
	if err := checkAPIError(resp, rawResp); err != nil {
		return "", err
	}

	prettyJSON, _ := json.MarshalIndent(json.RawMessage(resp.Data), "", "  ")
	return string(prettyJSON), nil
}

// --- init: register all stats subcommands ---

func init() {
	// stats flow flags
	statsFlowCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsFlowCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsFlowCmd.Flags().String("domains", "", "Comma-separated list of domains to query")
	statsFlowCmd.Flags().String("interval", "", "Time interval for aggregation (e.g., 5min, 1hour, 1day)")

	// stats request flags
	statsRequestCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsRequestCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsRequestCmd.Flags().String("domains", "", "Comma-separated list of domains to query")
	statsRequestCmd.Flags().String("interval", "", "Time interval for aggregation (e.g., 5min, 1hour, 1day)")

	// stats hit-flow flags
	statsHitFlowCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHitFlowCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHitFlowCmd.Flags().String("domains", "", "Comma-separated list of domains to query")

	// stats hit-request flags
	statsHitRequestCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHitRequestCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHitRequestCmd.Flags().String("domains", "", "Comma-separated list of domains to query")

	// stats http-code flags
	statsHTTPCodeCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHTTPCodeCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHTTPCodeCmd.Flags().String("domains", "", "Comma-separated list of domains to query")

	// stats http-code-detail flags
	statsHTTPCodeDetailCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHTTPCodeDetailCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsHTTPCodeDetailCmd.Flags().String("domains", "", "Comma-separated list of domains to query")

	// stats district flags
	statsDistrictCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsDistrictCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsDistrictCmd.Flags().String("domains", "", "Comma-separated list of domains to query")

	// stats top-domain flags
	statsTopDomainCmd.Flags().String("start-time", "", "Start time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")
	statsTopDomainCmd.Flags().String("end-time", "", "End time (yyyy-mm-dd or yyyy-mm-dd hh:mm)")

	// stats top-url flags
	statsTopURLCmd.Flags().String("start-time", "", "Start time for the query (yyyy-mm-dd hh:mm)")
	statsTopURLCmd.Flags().String("end-time", "", "End time for the query (yyyy-mm-dd hh:mm)")
	statsTopURLCmd.Flags().String("scope", "", "Time scope (today, yesterday, week, month, last_month)")
	statsTopURLCmd.Flags().String("sorted", "", "Sort order (url_size or url_count)")

	// stats top-referer flags
	statsTopRefererCmd.Flags().String("start-time", "", "Start time for the query (yyyy-mm-dd hh:mm)")
	statsTopRefererCmd.Flags().String("end-time", "", "End time for the query (yyyy-mm-dd hh:mm)")
	statsTopRefererCmd.Flags().String("scope", "", "Time scope (today, yesterday, week, month, last_month)")
	statsTopRefererCmd.Flags().String("sorted", "", "Sort order (referer_size or referer_count)")

	// stats top-ua flags
	statsTopUACmd.Flags().String("start-time", "", "Start time for the query (yyyy-mm-dd hh:mm)")
	statsTopUACmd.Flags().String("end-time", "", "End time for the query (yyyy-mm-dd hh:mm)")
	statsTopUACmd.Flags().String("scope", "", "Time scope (today, yesterday, week, month, last_month)")
	statsTopUACmd.Flags().String("sorted", "", "Sort order (ua_size or ua_count)")

	// Register subcommands under statsCmd
	statsCmd.AddCommand(statsFlowCmd)
	statsCmd.AddCommand(statsRequestCmd)
	statsCmd.AddCommand(statsHitFlowCmd)
	statsCmd.AddCommand(statsHitRequestCmd)
	statsCmd.AddCommand(statsHTTPCodeCmd)
	statsCmd.AddCommand(statsHTTPCodeDetailCmd)
	statsCmd.AddCommand(statsDistrictCmd)
	statsCmd.AddCommand(statsISOCountryCmd)
	statsCmd.AddCommand(statsTopDomainCmd)
	statsCmd.AddCommand(statsTopURLCmd)
	statsCmd.AddCommand(statsTopRefererCmd)
	statsCmd.AddCommand(statsTopUACmd)

	// Register statsCmd under root
	rootCmd.AddCommand(statsCmd)
}
