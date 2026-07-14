package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache management operations",
	Example: `  # Purge URLs
  racore-cli cache purge --urls "https://example.com/path1,https://example.com/path2"

  # Check purge status
  racore-cli cache purge-status --task-id abc123

  # Prefetch URLs
  racore-cli cache prefetch --urls "https://example.com/bigfile.zip"

  # List cache policies
  racore-cli cache list-policies --domain example.com
  racore-cli cache list-policies --domain example.com --type custom`,
}

var cachePurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge cached URLs from CDN edge nodes",
	Example: `  racore-cli cache purge --urls "https://example.com/path1,https://example.com/path2"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlsStr, _ := cmd.Flags().GetString("urls")
		if urlsStr == "" {
			return fmt.Errorf("--urls flag is required")
		}
		urls := strings.Split(urlsStr, ",")
		result, err := executeCachePurge(urls)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cachePurgeStatusCmd = &cobra.Command{
	Use:     "purge-status",
	Short:   "Check the status of a purge task",
	Example: `  racore-cli cache purge-status --task-id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" {
			return fmt.Errorf("--task-id flag is required")
		}
		result, err := executeCachePurgeStatus(taskID)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cachePrefetchCmd = &cobra.Command{
	Use:   "prefetch",
	Short: "Prefetch URLs to CDN edge nodes",
	Example: `  racore-cli cache prefetch --urls "https://example.com/file.zip" --region us-east-1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		urlsStr, _ := cmd.Flags().GetString("urls")
		if urlsStr == "" {
			return fmt.Errorf("--urls flag is required")
		}
		urls := strings.Split(urlsStr, ",")
		region, _ := cmd.Flags().GetString("region")
		country, _ := cmd.Flags().GetString("country")
		result, err := executeCachePrefetch(urls, region, country)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cachePrefetchStatusCmd = &cobra.Command{
	Use:     "prefetch-status",
	Short:   "Check the status of a prefetch task",
	Example: `  racore-cli cache prefetch-status --task-id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" {
			return fmt.Errorf("--task-id flag is required")
		}
		result, err := executeCachePrefetchStatus(taskID)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cachePrewarmRegionsCmd = &cobra.Command{
	Use:     "prewarm-regions",
	Short:   "Get available prewarm regions for a URL",
	Example: `  racore-cli cache prewarm-regions --url "https://example.com/"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		if url == "" {
			return fmt.Errorf("--url flag is required")
		}
		result, err := executeCachePrewarmRegions(url)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cachePrewarmPopCmd = &cobra.Command{
	Use:     "prewarm-pop",
	Short:   "Get prewarm PoP nodes for a region",
	Example: `  racore-cli cache prewarm-pop --region us-east-1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		region, _ := cmd.Flags().GetString("region")
		if region == "" {
			return fmt.Errorf("--region flag is required")
		}
		result, err := executeCachePrewarmPop(region)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cacheListPoliciesCmd = &cobra.Command{
	Use:   "list-policies",
	Short: "List available cache policies",
	Example: `  # List managed policies
  racore-cli cache list-policies --domain example.com

  # List custom policies
  racore-cli cache list-policies --domain example.com --type custom`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		if domain == "" {
			return fmt.Errorf("--domain flag is required")
		}
		policyType, _ := cmd.Flags().GetString("type")
		result, err := executeCacheListPolicies(domain, policyType)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cacheListOriginRequestPoliciesCmd = &cobra.Command{
	Use:     "list-origin-request-policies",
	Short:   "List available origin request policies",
	Example: `  racore-cli cache list-origin-request-policies --domain example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		if domain == "" {
			return fmt.Errorf("--domain flag is required")
		}
		policyType, _ := cmd.Flags().GetString("type")
		result, err := executeCacheListOriginRequestPolicies(domain, policyType)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

var cacheListResponseHeaderPoliciesCmd = &cobra.Command{
	Use:     "list-response-header-policies",
	Short:   "List available response header policies",
	Example: `  racore-cli cache list-response-header-policies --domain example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		if domain == "" {
			return fmt.Errorf("--domain flag is required")
		}
		policyType, _ := cmd.Flags().GetString("type")
		result, err := executeCacheListResponseHeaderPolicies(domain, policyType)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

func init() {
	// Register flags
	cachePurgeCmd.Flags().String("urls", "", "Comma-separated list of URLs to purge from CDN cache")
	cachePurgeStatusCmd.Flags().String("task-id", "", "Purge task ID returned from purge command")
	cachePrefetchCmd.Flags().String("urls", "", "Comma-separated list of URLs to prefetch to edge nodes")
	cachePrefetchCmd.Flags().String("region", "", "Target region for prefetch (e.g., us-east-1)")
	cachePrefetchCmd.Flags().String("country", "", "Target country code for prefetch")
	cachePrefetchStatusCmd.Flags().String("task-id", "", "Prefetch task ID returned from prefetch command")
	cachePrewarmRegionsCmd.Flags().String("url", "", "URL to query available prewarm regions for")
	cachePrewarmPopCmd.Flags().String("region", "", "Region to list prewarm PoP nodes for")
	cacheListPoliciesCmd.Flags().String("domain", "", "CDN domain name (required)")
	cacheListPoliciesCmd.Flags().String("type", "managed", "Policy type: managed or custom")
	cacheListOriginRequestPoliciesCmd.Flags().String("domain", "", "CDN domain name (required)")
	cacheListOriginRequestPoliciesCmd.Flags().String("type", "managed", "Policy type: managed or custom")
	cacheListResponseHeaderPoliciesCmd.Flags().String("domain", "", "CDN domain name (required)")
	cacheListResponseHeaderPoliciesCmd.Flags().String("type", "managed", "Policy type: managed or custom")

	// Add subcommands to cache parent
	cacheCmd.AddCommand(cachePurgeCmd)
	cacheCmd.AddCommand(cachePurgeStatusCmd)
	cacheCmd.AddCommand(cachePrefetchCmd)
	cacheCmd.AddCommand(cachePrefetchStatusCmd)
	cacheCmd.AddCommand(cachePrewarmRegionsCmd)
	cacheCmd.AddCommand(cachePrewarmPopCmd)
	cacheCmd.AddCommand(cacheListPoliciesCmd)
	cacheCmd.AddCommand(cacheListOriginRequestPoliciesCmd)
	cacheCmd.AddCommand(cacheListResponseHeaderPoliciesCmd)

	// Register cache command group with root
	rootCmd.AddCommand(cacheCmd)
}

// Execute functions

func executeCachePurge(urls []string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]interface{}{"urls": urls}
	rawResp, err := client.Post("/API/cdn/purge", body)
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

	var data struct {
		TaskID string `json:"task_id"`
	}
	json.Unmarshal(resp.Data, &data)
	return fmt.Sprintf("Purge submitted. Task ID: %s", data.TaskID), nil
}

func executeCachePurgeStatus(taskID string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"task_id": taskID}
	rawResp, err := client.Get("/API/cdn/purge/detail", queryParams)
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

func executeCachePrefetch(urls []string, region, country string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]interface{}{"urls": urls}
	if region != "" {
		body["region"] = region
	}
	if country != "" {
		body["country"] = country
	}

	rawResp, err := client.Post("/API/cdn/prefetch", body)
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

	var data struct {
		TaskID string `json:"task_id"`
	}
	json.Unmarshal(resp.Data, &data)
	return fmt.Sprintf("Prefetch submitted. Task ID: %s", data.TaskID), nil
}

func executeCachePrefetchStatus(taskID string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"task_id": taskID}
	rawResp, err := client.Get("/API/cdn/prefetch/detail", queryParams)
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

func executeCachePrewarmRegions(url string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"url": url}
	rawResp, err := client.Get("/API/aws/prewarm/get/region", queryParams)
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

func executeCachePrewarmPop(region string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"region": region}
	rawResp, err := client.Get("/API/aws/prewarm/pop", queryParams)
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

func executeCacheListPolicies(domain, policyType string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"domain": domain, "type": policyType}
	rawResp, err := client.Get("/API/cdn/list/cache/policies", queryParams)
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

func executeCacheListOriginRequestPolicies(domain, policyType string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"domain": domain, "type": policyType}
	rawResp, err := client.Get("/API/cdn/aws/origin/request/policies", queryParams)
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

func executeCacheListResponseHeaderPolicies(domain, policyType string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"domain": domain, "type": policyType}
	rawResp, err := client.Get("/API/cdn/aws/response/headers/policies", queryParams)
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
