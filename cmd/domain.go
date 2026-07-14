package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// --- Parent command ---

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage CDN domains",
}

// --- Top-level subcommands ---

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDN domains",
	RunE:  runDomainList,
}

var domainCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CDN domain",
	RunE:  runDomainCreate,
}

var domainDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CDN domain",
	RunE:  runDomainDelete,
}

var domainEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a CDN domain",
	RunE:  runDomainEnable,
}

var domainDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a CDN domain",
	RunE:  runDomainDisable,
}

// --- Two-level nested subcommands (parent commands) ---

var domainSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage domain origin source configuration",
}

var domainSSLCmd = &cobra.Command{
	Use:   "ssl",
	Short: "Manage domain SSL configuration",
}

var domainEnforceHTTPSCmd = &cobra.Command{
	Use:   "enforce-https",
	Short: "Manage HTTPS enforcement configuration",
}

var domainIPFilterCmd = &cobra.Command{
	Use:   "ip-filter",
	Short: "Manage IP filter configuration",
}

var domainRefererFilterCmd = &cobra.Command{
	Use:   "referer-filter",
	Short: "Manage referer filter configuration",
}

var domainUAFilterCmd = &cobra.Command{
	Use:   "ua-filter",
	Short: "Manage user-agent filter configuration",
}

var domainOriginProtocolCmd = &cobra.Command{
	Use:   "origin-protocol",
	Short: "Manage origin protocol policy",
}

var domainHTTP2Cmd = &cobra.Command{
	Use:   "http2",
	Short: "Manage HTTP/2 configuration",
}

var domainHTTP3Cmd = &cobra.Command{
	Use:   "http3",
	Short: "Manage HTTP/3 configuration",
}

var domainTLSVersionCmd = &cobra.Command{
	Use:   "tls-version",
	Short: "Manage minimum TLS version",
}

var domainCompressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Manage page compression configuration",
}

var domainIPv6Cmd = &cobra.Command{
	Use:   "ipv6",
	Short: "Manage IPv6 configuration",
}

var domainCachePolicyCmd = &cobra.Command{
	Use:   "cache-policy",
	Short: "Manage cache policy configuration",
}

var domainOriginHostCmd = &cobra.Command{
	Use:   "origin-host",
	Short: "Manage origin host configuration",
}

var domainOriginTimeoutCmd = &cobra.Command{
	Use:   "origin-timeout",
	Short: "Manage origin connection timeout policy",
}

var domainGeoRestrictionCmd = &cobra.Command{
	Use:   "geo-restriction",
	Short: "Manage geo restriction configuration",
}

var domainRequestHeadersCmd = &cobra.Command{
	Use:   "request-headers",
	Short: "Manage HTTP request headers",
}

var domainResponseHeadersCmd = &cobra.Command{
	Use:   "response-headers",
	Short: "Manage HTTP response headers",
}

var domainRequestHeaderPolicyCmd = &cobra.Command{
	Use:   "request-header-policy",
	Short: "Manage request header policy",
}

var domainResponseHeaderPolicyCmd = &cobra.Command{
	Use:   "response-header-policy",
	Short: "Manage response header policy",
}

// --- Get/Set subcommands for each nested command ---

var domainSourceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get source configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/source"),
}

var domainSourceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set source configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/source", "put"),
}

var domainSSLGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get SSL configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/ssl"),
}

var domainSSLSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set SSL configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/ssl", "put"),
}

var domainEnforceHTTPSGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HTTPS enforcement configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/enforce/https"),
}

var domainEnforceHTTPSSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTPS enforcement configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/enforce/https", "put"),
}

var domainIPFilterGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get IP filter configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/ip/filter"),
}

var domainIPFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set IP filter configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/ip/filter", "post"),
}

var domainRefererFilterGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get referer filter configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/referer/filter"),
}

var domainRefererFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set referer filter configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/referer/filter", "post"),
}

var domainUAFilterGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get user-agent filter configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/user/agent/filter"),
}

var domainUAFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set user-agent filter configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/user/agent/filter", "post"),
}

var domainOriginProtocolGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get origin protocol policy",
	RunE:  runDomainConfigGet("/API/cdn/domain/origin/protocol/policy"),
}

var domainOriginProtocolSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin protocol policy",
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/protocol/policy", "put"),
}

var domainHTTP2GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HTTP/2 configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/http2"),
}

var domainHTTP2SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP/2 configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/http2", "put"),
}

var domainHTTP3GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HTTP/3 configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/http3"),
}

var domainHTTP3SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP/3 configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/http3", "put"),
}

var domainTLSVersionGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get minimum TLS version",
	RunE:  runDomainConfigGet("/API/cdn/domain/min/tls/version"),
}

var domainTLSVersionSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set minimum TLS version",
	RunE:  runDomainConfigSet("/API/cdn/domain/min/tls/version", "put"),
}

var domainCompressGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get page compression configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/page/compress"),
}

var domainCompressSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set page compression configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/page/compress", "put"),
}

var domainIPv6GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get IPv6 configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/ipv6"),
}

var domainIPv6SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set IPv6 configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/ipv6", "put"),
}

var domainCachePolicyGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get cache policy configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/cache/conf"),
}

var domainCachePolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set cache policy configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/cache/conf", "put"),
}

var domainOriginHostGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get origin host configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/origin/host"),
}

var domainOriginHostSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin host configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/host", "put"),
}

var domainOriginTimeoutGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get origin connection timeout policy",
	RunE:  runDomainConfigGet("/API/cdn/domain/origin/connection/policy"),
}

var domainOriginTimeoutSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin connection timeout policy",
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/connection/policy", "put"),
}

var domainGeoRestrictionGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get geo restriction configuration",
	RunE:  runDomainConfigGet("/API/cdn/domain/geo/restriction"),
}

var domainGeoRestrictionSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set geo restriction configuration",
	RunE:  runDomainConfigSet("/API/cdn/domain/geo/restriction", "put"),
}

var domainRequestHeadersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HTTP request headers",
	RunE:  runDomainConfigGet("/API/cdn/domain/http/request/headers"),
}

var domainRequestHeadersSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP request headers",
	RunE:  runDomainConfigSet("/API/cdn/domain/http/request/headers", "post"),
}

var domainResponseHeadersGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get HTTP response headers",
	RunE:  runDomainConfigGet("/API/cdn/domain/http/response/headers"),
}

var domainResponseHeadersSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP response headers",
	RunE:  runDomainConfigSet("/API/cdn/domain/http/response/headers", "post"),
}

var domainRequestHeaderPolicyGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get request header policy",
	RunE:  runDomainConfigGet("/API/cdn/domain/request/header/policy"),
}

var domainRequestHeaderPolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set request header policy",
	RunE:  runDomainConfigSet("/API/cdn/domain/request/header/policy", "put"),
}

var domainResponseHeaderPolicyGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get response header policy",
	RunE:  runDomainConfigGet("/API/cdn/domain/response/header/policy"),
}

var domainResponseHeaderPolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set response header policy",
	RunE:  runDomainConfigSet("/API/cdn/domain/response/header/policy", "put"),
}

// --- Execute functions ---

func executeDomainList(filter string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := make(map[string]string)
	if filter != "" {
		queryParams["domain"] = filter
	}

	rawResp, err := client.Get("/API/cdn/domain", queryParams)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Name   string `json:"name"`
			Cname  string `json:"cname"`
			Type   string `json:"type"`
			Status string `json:"status"`
		} `json:"data"`
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

	headers := []string{"DOMAIN", "CNAME", "TYPE", "STATUS"}
	rows := make([][]string, 0, len(resp.Data))
	for _, d := range resp.Data {
		rows = append(rows, []string{d.Name, d.Cname, d.Type, d.Status})
	}
	return formatTable(headers, rows), nil
}

func executeDomainCreate(domain, domainType, source string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]string{
		"domain": domain,
		"type":   domainType,
		"source": source,
	}
	rawResp, err := client.Post("/API/cdn/domain", body)
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

	return fmt.Sprintf("Domain %s created successfully.", domain), nil
}

func executeDomainDelete(domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]string{"domain": domain}
	rawResp, err := client.Delete("/API/cdn/domain", body)
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

	return fmt.Sprintf("Domain %s deleted successfully.", domain), nil
}

func executeDomainEnable(domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]string{"domain": domain}
	rawResp, err := client.Put("/API/cdn/domain/state/open", body)
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

	return fmt.Sprintf("Domain %s enabled successfully.", domain), nil
}

func executeDomainDisable(domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]string{"domain": domain}
	rawResp, err := client.Put("/API/cdn/domain/state/close", body)
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

	return fmt.Sprintf("Domain %s disabled successfully.", domain), nil
}

// executeDomainConfigGet is a generic function for GET config endpoints.
func executeDomainConfigGet(endpoint, domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	queryParams := map[string]string{"domain": domain}
	rawResp, err := client.Get(endpoint, queryParams)
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

	// Pretty-print the data field as JSON
	prettyData, err := json.MarshalIndent(resp.Data, "", "  ")
	if err != nil {
		return string(resp.Data), nil
	}
	return string(prettyData), nil
}

// executeDomainConfigSet is a generic function for PUT/POST config endpoints.
func executeDomainConfigSet(endpoint, domain, configJSON, method string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	// Parse the config JSON into a map
	var configMap map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		return "", fmt.Errorf("invalid JSON config: %w", err)
	}

	// Add domain to the body
	configMap["domain"] = domain

	var rawResp json.RawMessage
	if method == "post" {
		rawResp, err = client.Post(endpoint, configMap)
	} else {
		rawResp, err = client.Put(endpoint, configMap)
	}
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

	return "Configuration updated successfully.", nil
}

// --- Run functions for top-level subcommands ---

func runDomainList(cmd *cobra.Command, args []string) error {
	filter, _ := cmd.Flags().GetString("filter")
	result, err := executeDomainList(filter)
	if err != nil {
		return err
	}
	fmt.Print(result)
	return nil
}

func runDomainCreate(cmd *cobra.Command, args []string) error {
	domain, _ := cmd.Flags().GetString("domain")
	domainType, _ := cmd.Flags().GetString("type")
	source, _ := cmd.Flags().GetString("source")

	if domain == "" || domainType == "" || source == "" {
		return fmt.Errorf("--domain, --type, and --source flags are required")
	}

	result, err := executeDomainCreate(domain, domainType, source)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func runDomainDelete(cmd *cobra.Command, args []string) error {
	domain, _ := cmd.Flags().GetString("domain")
	if domain == "" {
		return fmt.Errorf("--domain flag is required")
	}

	result, err := executeDomainDelete(domain)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func runDomainEnable(cmd *cobra.Command, args []string) error {
	domain, _ := cmd.Flags().GetString("domain")
	if domain == "" {
		return fmt.Errorf("--domain flag is required")
	}

	result, err := executeDomainEnable(domain)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func runDomainDisable(cmd *cobra.Command, args []string) error {
	domain, _ := cmd.Flags().GetString("domain")
	if domain == "" {
		return fmt.Errorf("--domain flag is required")
	}

	result, err := executeDomainDisable(domain)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

// --- Generic run function factories for get/set subcommands ---

// runDomainConfigGet returns a RunE function for a config GET endpoint.
func runDomainConfigGet(endpoint string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		if domain == "" {
			return fmt.Errorf("--domain flag is required")
		}

		result, err := executeDomainConfigGet(endpoint, domain)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	}
}

// runDomainConfigSet returns a RunE function for a config PUT/POST endpoint.
func runDomainConfigSet(endpoint, method string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		config, _ := cmd.Flags().GetString("config")
		if domain == "" || config == "" {
			return fmt.Errorf("--domain and --config flags are required")
		}

		result, err := executeDomainConfigSet(endpoint, domain, config, method)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	}
}

// --- Init: register flags and wire command tree ---

func init() {
	// Top-level subcommand flags
	domainListCmd.Flags().StringP("filter", "f", "", "Filter domains by name")
	domainCreateCmd.Flags().String("domain", "", "Domain name")
	domainCreateCmd.Flags().String("type", "", "Domain type")
	domainCreateCmd.Flags().String("source", "", "Origin source")
	domainDeleteCmd.Flags().String("domain", "", "Domain name")
	domainEnableCmd.Flags().String("domain", "", "Domain name")
	domainDisableCmd.Flags().String("domain", "", "Domain name")

	// Get subcommand flags (--domain)
	domainSourceGetCmd.Flags().String("domain", "", "Domain name")
	domainSSLGetCmd.Flags().String("domain", "", "Domain name")
	domainEnforceHTTPSGetCmd.Flags().String("domain", "", "Domain name")
	domainIPFilterGetCmd.Flags().String("domain", "", "Domain name")
	domainRefererFilterGetCmd.Flags().String("domain", "", "Domain name")
	domainUAFilterGetCmd.Flags().String("domain", "", "Domain name")
	domainOriginProtocolGetCmd.Flags().String("domain", "", "Domain name")
	domainHTTP2GetCmd.Flags().String("domain", "", "Domain name")
	domainHTTP3GetCmd.Flags().String("domain", "", "Domain name")
	domainTLSVersionGetCmd.Flags().String("domain", "", "Domain name")
	domainCompressGetCmd.Flags().String("domain", "", "Domain name")
	domainIPv6GetCmd.Flags().String("domain", "", "Domain name")
	domainCachePolicyGetCmd.Flags().String("domain", "", "Domain name")
	domainOriginHostGetCmd.Flags().String("domain", "", "Domain name")
	domainOriginTimeoutGetCmd.Flags().String("domain", "", "Domain name")
	domainGeoRestrictionGetCmd.Flags().String("domain", "", "Domain name")
	domainRequestHeadersGetCmd.Flags().String("domain", "", "Domain name")
	domainResponseHeadersGetCmd.Flags().String("domain", "", "Domain name")
	domainRequestHeaderPolicyGetCmd.Flags().String("domain", "", "Domain name")
	domainResponseHeaderPolicyGetCmd.Flags().String("domain", "", "Domain name")

	// Set subcommand flags (--domain and --config)
	domainSourceSetCmd.Flags().String("domain", "", "Domain name")
	domainSourceSetCmd.Flags().String("config", "", "Configuration JSON")
	domainSSLSetCmd.Flags().String("domain", "", "Domain name")
	domainSSLSetCmd.Flags().String("config", "", "Configuration JSON")
	domainEnforceHTTPSSetCmd.Flags().String("domain", "", "Domain name")
	domainEnforceHTTPSSetCmd.Flags().String("config", "", "Configuration JSON")
	domainIPFilterSetCmd.Flags().String("domain", "", "Domain name")
	domainIPFilterSetCmd.Flags().String("config", "", "Configuration JSON")
	domainRefererFilterSetCmd.Flags().String("domain", "", "Domain name")
	domainRefererFilterSetCmd.Flags().String("config", "", "Configuration JSON")
	domainUAFilterSetCmd.Flags().String("domain", "", "Domain name")
	domainUAFilterSetCmd.Flags().String("config", "", "Configuration JSON")
	domainOriginProtocolSetCmd.Flags().String("domain", "", "Domain name")
	domainOriginProtocolSetCmd.Flags().String("config", "", "Configuration JSON")
	domainHTTP2SetCmd.Flags().String("domain", "", "Domain name")
	domainHTTP2SetCmd.Flags().String("config", "", "Configuration JSON")
	domainHTTP3SetCmd.Flags().String("domain", "", "Domain name")
	domainHTTP3SetCmd.Flags().String("config", "", "Configuration JSON")
	domainTLSVersionSetCmd.Flags().String("domain", "", "Domain name")
	domainTLSVersionSetCmd.Flags().String("config", "", "Configuration JSON")
	domainCompressSetCmd.Flags().String("domain", "", "Domain name")
	domainCompressSetCmd.Flags().String("config", "", "Configuration JSON")
	domainIPv6SetCmd.Flags().String("domain", "", "Domain name")
	domainIPv6SetCmd.Flags().String("config", "", "Configuration JSON")
	domainCachePolicySetCmd.Flags().String("domain", "", "Domain name")
	domainCachePolicySetCmd.Flags().String("config", "", "Configuration JSON")
	domainOriginHostSetCmd.Flags().String("domain", "", "Domain name")
	domainOriginHostSetCmd.Flags().String("config", "", "Configuration JSON")
	domainOriginTimeoutSetCmd.Flags().String("domain", "", "Domain name")
	domainOriginTimeoutSetCmd.Flags().String("config", "", "Configuration JSON")
	domainGeoRestrictionSetCmd.Flags().String("domain", "", "Domain name")
	domainGeoRestrictionSetCmd.Flags().String("config", "", "Configuration JSON")
	domainRequestHeadersSetCmd.Flags().String("domain", "", "Domain name")
	domainRequestHeadersSetCmd.Flags().String("config", "", "Configuration JSON")
	domainResponseHeadersSetCmd.Flags().String("domain", "", "Domain name")
	domainResponseHeadersSetCmd.Flags().String("config", "", "Configuration JSON")
	domainRequestHeaderPolicySetCmd.Flags().String("domain", "", "Domain name")
	domainRequestHeaderPolicySetCmd.Flags().String("config", "", "Configuration JSON")
	domainResponseHeaderPolicySetCmd.Flags().String("domain", "", "Domain name")
	domainResponseHeaderPolicySetCmd.Flags().String("config", "", "Configuration JSON")

	// Wire get/set into nested parent commands
	domainSourceCmd.AddCommand(domainSourceGetCmd, domainSourceSetCmd)
	domainSSLCmd.AddCommand(domainSSLGetCmd, domainSSLSetCmd)
	domainEnforceHTTPSCmd.AddCommand(domainEnforceHTTPSGetCmd, domainEnforceHTTPSSetCmd)
	domainIPFilterCmd.AddCommand(domainIPFilterGetCmd, domainIPFilterSetCmd)
	domainRefererFilterCmd.AddCommand(domainRefererFilterGetCmd, domainRefererFilterSetCmd)
	domainUAFilterCmd.AddCommand(domainUAFilterGetCmd, domainUAFilterSetCmd)
	domainOriginProtocolCmd.AddCommand(domainOriginProtocolGetCmd, domainOriginProtocolSetCmd)
	domainHTTP2Cmd.AddCommand(domainHTTP2GetCmd, domainHTTP2SetCmd)
	domainHTTP3Cmd.AddCommand(domainHTTP3GetCmd, domainHTTP3SetCmd)
	domainTLSVersionCmd.AddCommand(domainTLSVersionGetCmd, domainTLSVersionSetCmd)
	domainCompressCmd.AddCommand(domainCompressGetCmd, domainCompressSetCmd)
	domainIPv6Cmd.AddCommand(domainIPv6GetCmd, domainIPv6SetCmd)
	domainCachePolicyCmd.AddCommand(domainCachePolicyGetCmd, domainCachePolicySetCmd)
	domainOriginHostCmd.AddCommand(domainOriginHostGetCmd, domainOriginHostSetCmd)
	domainOriginTimeoutCmd.AddCommand(domainOriginTimeoutGetCmd, domainOriginTimeoutSetCmd)
	domainGeoRestrictionCmd.AddCommand(domainGeoRestrictionGetCmd, domainGeoRestrictionSetCmd)
	domainRequestHeadersCmd.AddCommand(domainRequestHeadersGetCmd, domainRequestHeadersSetCmd)
	domainResponseHeadersCmd.AddCommand(domainResponseHeadersGetCmd, domainResponseHeadersSetCmd)
	domainRequestHeaderPolicyCmd.AddCommand(domainRequestHeaderPolicyGetCmd, domainRequestHeaderPolicySetCmd)
	domainResponseHeaderPolicyCmd.AddCommand(domainResponseHeaderPolicyGetCmd, domainResponseHeaderPolicySetCmd)

	// Wire all subcommands into domain parent
	domainCmd.AddCommand(
		domainListCmd,
		domainCreateCmd,
		domainDeleteCmd,
		domainEnableCmd,
		domainDisableCmd,
		domainSourceCmd,
		domainSSLCmd,
		domainEnforceHTTPSCmd,
		domainIPFilterCmd,
		domainRefererFilterCmd,
		domainUAFilterCmd,
		domainOriginProtocolCmd,
		domainHTTP2Cmd,
		domainHTTP3Cmd,
		domainTLSVersionCmd,
		domainCompressCmd,
		domainIPv6Cmd,
		domainCachePolicyCmd,
		domainOriginHostCmd,
		domainOriginTimeoutCmd,
		domainGeoRestrictionCmd,
		domainRequestHeadersCmd,
		domainResponseHeadersCmd,
		domainRequestHeaderPolicyCmd,
		domainResponseHeaderPolicyCmd,
	)

	// Register domain command with root
	rootCmd.AddCommand(domainCmd)
}
