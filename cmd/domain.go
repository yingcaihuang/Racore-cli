package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"racore-cli/internal/api"

	"github.com/spf13/cobra"
)

// --- Parent command ---

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage CDN domains",
	Example: `  # List all domains
  racore-cli domain list

  # Filter by name
  racore-cli domain list --filter example.com

  # Create a domain (auto-matches wildcard SSL cert)
  racore-cli domain create --domain cdn.example.com --type oversea --source origin.example.com

  # Create with explicit cert
  racore-cli domain create --domain cdn.example.com --type static --source origin.example.com --cert-id 123

  # Enable/Disable domain
  racore-cli domain enable --domain cdn.example.com
  racore-cli domain disable --domain cdn.example.com

  # Delete domain (must be disabled first)
  racore-cli domain delete --domain cdn.example.com`,
}

// --- Top-level subcommands ---

var domainListCmd = &cobra.Command{
	Use:   "list",
	Short: "List CDN domains",
	Example: `  # List all domains
  racore-cli domain list

  # Filter by domain name
  racore-cli domain list --filter example.com`,
	RunE: runDomainList,
}

var domainCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new CDN domain",
	Example: `  # Create domain with auto cert matching
  racore-cli domain create --domain cdn.example.com --type oversea --source origin.example.com

  # Create with explicit cert ID
  racore-cli domain create --domain cdn.example.com --type static --source origin.example.com --cert-id 123

  # Type options: oversea, live, video, dynamic, static, download`,
	RunE: runDomainCreate,
}

var domainDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a CDN domain",
	Example: `  # Delete a domain (must be disabled first)
  racore-cli domain delete --domain cdn.example.com`,
	RunE: runDomainDelete,
}

var domainEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a CDN domain",
	Example: `  racore-cli domain enable --domain cdn.example.com`,
	RunE:    runDomainEnable,
}

var domainDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a CDN domain",
	Example: `  racore-cli domain disable --domain cdn.example.com`,
	RunE:    runDomainDisable,
}

// --- Two-level nested subcommands (parent commands) ---

var domainSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage domain origin source configuration",
	Example: `  # Query source configuration
  racore-cli domain source get --domain example.com

  # Set origin source
  racore-cli domain source set --domain example.com --config '{"source_type":"2","source_conf":[{"source":"origin.example.com","type":"1"}]}'`,
}

var domainSSLCmd = &cobra.Command{
	Use:   "ssl",
	Short: "Manage domain SSL configuration",
	Example: `  # Query SSL status
  racore-cli domain ssl get --domain example.com

  # Enable SSL with certificate
  racore-cli domain ssl set --domain example.com --config '{"is_ssl":"1","cert_id":"123"}'

  # Disable SSL
  racore-cli domain ssl set --domain example.com --config '{"is_ssl":"0","cert_id":""}'`,
}

var domainEnforceHTTPSCmd = &cobra.Command{
	Use:   "enforce-https",
	Short: "Manage HTTPS enforcement configuration",
	Example: `  # Query HTTPS redirect status
  racore-cli domain enforce-https get --domain example.com

  # Enable force HTTPS redirect (requires SSL enabled)
  racore-cli domain enforce-https set --domain example.com --config '{"https_redirect":"on"}'

  # Disable force HTTPS redirect
  racore-cli domain enforce-https set --domain example.com --config '{"https_redirect":"off"}'`,
}

var domainIPFilterCmd = &cobra.Command{
	Use:   "ip-filter",
	Short: "Manage IP filter configuration",
	Example: `  # Query IP filter
  racore-cli domain ip-filter get --domain example.com

  # Set IP whitelist
  racore-cli domain ip-filter set --domain example.com --config '{"state":"on","type":"white","value":["1.2.3.4","5.6.7.8"]}'

  # Set IP blacklist
  racore-cli domain ip-filter set --domain example.com --config '{"state":"on","type":"black","value":["10.0.0.1"]}'

  # Disable IP filter
  racore-cli domain ip-filter set --domain example.com --config '{"state":"off","value":[]}'`,
}

var domainRefererFilterCmd = &cobra.Command{
	Use:   "referer-filter",
	Short: "Manage referer filter configuration",
	Example: `  # Query referer filter
  racore-cli domain referer-filter get --domain example.com

  # Set referer whitelist
  racore-cli domain referer-filter set --domain example.com --config '{"state":"on","type":"white","values":["example.com","*.example.com"],"allow_empty":"on"}'

  # Set referer blacklist
  racore-cli domain referer-filter set --domain example.com --config '{"state":"on","type":"black","values":["bad.com"],"allow_empty":"off"}'

  # Disable referer filter
  racore-cli domain referer-filter set --domain example.com --config '{"state":"off","type":"off","values":["example.com"],"allow_empty":"on"}'`,
}

var domainUAFilterCmd = &cobra.Command{
	Use:   "ua-filter",
	Short: "Manage user-agent filter configuration",
	Example: `  # Query UA filter
  racore-cli domain ua-filter get --domain example.com

  # Set UA blacklist
  racore-cli domain ua-filter set --domain example.com --config '{"state":"on","type":"black","values":["curl/*","wget/*"]}'

  # Disable UA filter
  racore-cli domain ua-filter set --domain example.com --config '{"state":"off","values":[]}'`,
}

var domainOriginProtocolCmd = &cobra.Command{
	Use:   "origin-protocol",
	Short: "Manage origin protocol policy",
	Example: `  # Query origin protocol
  racore-cli domain origin-protocol get --domain example.com

  # Set origin protocol (match-viewer, http-only, https-only)
  racore-cli domain origin-protocol set --domain example.com --config '{"origin_protocol_policy":"https-only","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'

  # Follow viewer protocol
  racore-cli domain origin-protocol set --domain example.com --config '{"origin_protocol_policy":"match-viewer","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'`,
}

var domainHTTP2Cmd = &cobra.Command{
	Use:   "http2",
	Short: "Manage HTTP/2 configuration",
	Example: `  # Query HTTP/2 status
  racore-cli domain http2 get --domain example.com

  # Enable HTTP/2
  racore-cli domain http2 set --domain example.com --config '{"enable":"on"}'

  # Disable HTTP/2
  racore-cli domain http2 set --domain example.com --config '{"enable":"off"}'`,
}

var domainHTTP3Cmd = &cobra.Command{
	Use:   "http3",
	Short: "Manage HTTP/3 configuration",
	Example: `  # Query HTTP/3 status
  racore-cli domain http3 get --domain example.com

  # Enable HTTP/3
  racore-cli domain http3 set --domain example.com --config '{"enable":"on"}'

  # Disable HTTP/3
  racore-cli domain http3 set --domain example.com --config '{"enable":"off"}'`,
}

var domainTLSVersionCmd = &cobra.Command{
	Use:   "tls-version",
	Short: "Manage minimum TLS version",
	Example: `  # Query current TLS version
  racore-cli domain tls-version get --domain example.com

  # Set minimum TLS version
  # Valid: SSLv3, TLSv1, TLSv1_2016, TLSv1.1_2016, TLSv1.2_2018, TLSv1.2_2019, TLSv1.2_2021
  racore-cli domain tls-version set --domain example.com --config '{"min_tls_version":"TLSv1.2_2021"}'`,
}

var domainCompressCmd = &cobra.Command{
	Use:   "compress",
	Short: "Manage page compression configuration",
	Example: `  # Query compression status
  racore-cli domain compress get --domain example.com

  # Enable smart compression
  racore-cli domain compress set --domain example.com --config '{"enable":"on"}'

  # Disable smart compression
  racore-cli domain compress set --domain example.com --config '{"enable":"off"}'`,
}

var domainIPv6Cmd = &cobra.Command{
	Use:   "ipv6",
	Short: "Manage IPv6 configuration",
	Example: `  # Query IPv6 status
  racore-cli domain ipv6 get --domain example.com

  # Enable IPv6
  racore-cli domain ipv6 set --domain example.com --config '{"enable":"1"}'

  # Disable IPv6
  racore-cli domain ipv6 set --domain example.com --config '{"enable":"0"}'`,
}

var domainCachePolicyCmd = &cobra.Command{
	Use:   "cache-policy",
	Short: "Manage cache policy configuration",
	Example: `  # Query cache policy
  racore-cli domain cache-policy get --domain example.com

  # Set cache policy (AWS only)
  racore-cli domain cache-policy set --domain example.com --config '{"cache_policy_id":"658327ea-f89d-4fab-a63d-7e88639e58f6"}'`,
}

var domainOriginHostCmd = &cobra.Command{
	Use:   "origin-host",
	Short: "Manage origin host configuration",
	Example: `  # Query origin host
  racore-cli domain origin-host get --domain example.com

  # Set to origin server domain (type 1)
  racore-cli domain origin-host set --domain example.com --config '{"origin_host_type":"1"}'

  # Set to acceleration domain (type 2)
  racore-cli domain origin-host set --domain example.com --config '{"origin_host_type":"2"}'

  # Set to custom domain (type 3)
  racore-cli domain origin-host set --domain example.com --config '{"origin_host_type":"3","origin_host":"custom.example.com"}'`,
}

var domainOriginTimeoutCmd = &cobra.Command{
	Use:   "origin-timeout",
	Short: "Manage origin connection timeout policy",
	Example: `  # Query origin timeout
  racore-cli domain origin-timeout get --domain example.com

  # Set origin timeout (AWS only)
  racore-cli domain origin-timeout set --domain example.com --config '{"connection_timeout":"10","response_timeout":"30"}'`,
}

var domainGeoRestrictionCmd = &cobra.Command{
	Use:   "geo-restriction",
	Short: "Manage geo restriction configuration",
	Example: `  # Query geo restriction
  racore-cli domain geo-restriction get --domain example.com

  # Set geo restriction (AWS only)
  racore-cli domain geo-restriction set --domain example.com --config '{"restriction_type":"whitelist","items":["CN","US","JP"]}'`,
}

var domainRequestHeadersCmd = &cobra.Command{
	Use:   "request-headers",
	Short: "Manage HTTP request headers",
	Example: `  # Query request headers
  racore-cli domain request-headers get --domain example.com

  # Set custom request headers
  racore-cli domain request-headers set --domain example.com --config '{"headers":[{"key":"X-Custom","value":"test","action":"add"}]}'`,
}

var domainResponseHeadersCmd = &cobra.Command{
	Use:   "response-headers",
	Short: "Manage HTTP response headers",
	Example: `  # Query response headers
  racore-cli domain response-headers get --domain example.com

  # Set custom response headers
  racore-cli domain response-headers set --domain example.com --config '{"headers":[{"key":"X-Frame-Options","value":"SAMEORIGIN","action":"add"}]}'`,
}

var domainRequestHeaderPolicyCmd = &cobra.Command{
	Use:   "request-header-policy",
	Short: "Manage request header policy",
	Example: `  # Query request header policy (AWS only)
  racore-cli domain request-header-policy get --domain example.com

  # Set request header policy
  racore-cli domain request-header-policy set --domain example.com --config '{"origin_request_policy_id":"acba4595-bd28-49b8-b9fe-13317c0390fa"}'`,
}

var domainResponseHeaderPolicyCmd = &cobra.Command{
	Use:   "response-header-policy",
	Short: "Manage response header policy",
	Example: `  # Query response header policy (AWS only)
  racore-cli domain response-header-policy get --domain example.com

  # Set response header policy
  racore-cli domain response-header-policy set --domain example.com --config '{"response_headers_policy_id":"60669652-455b-4ae9-85a4-c4c02393f86c"}'`,
}

// --- Get/Set subcommands for each nested command ---

var domainSourceGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get source configuration",
	Example: `  racore-cli domain source get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/source"),
}

var domainSourceSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set source configuration",
	Example: `  racore-cli domain source set --domain example.com --config '{"source_type":"2","source_conf":[{"source":"origin.example.com","type":"1"}]}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/source", "put"),
}

var domainSSLGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get SSL configuration",
	Example: `  racore-cli domain ssl get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/ssl"),
}

var domainSSLSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set SSL configuration",
	Example: `  # Enable SSL with certificate
  racore-cli domain ssl set --domain example.com --config '{"is_ssl":"1","cert_id":"123"}'

  # Disable SSL
  racore-cli domain ssl set --domain example.com --config '{"is_ssl":"0","cert_id":""}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/ssl", "put"),
}

var domainEnforceHTTPSGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get HTTPS enforcement configuration",
	Example: `  racore-cli domain enforce-https get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/enforce/https"),
}

var domainEnforceHTTPSSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTPS enforcement configuration",
	Example: `  # Enable HTTPS redirect
  racore-cli domain enforce-https set --domain example.com --config '{"https_redirect":"on"}'

  # Disable HTTPS redirect
  racore-cli domain enforce-https set --domain example.com --config '{"https_redirect":"off"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/enforce/https", "put"),
}

var domainIPFilterGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get IP filter configuration",
	Example: `  racore-cli domain ip-filter get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/ip/filter"),
}

var domainIPFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set IP filter configuration",
	Example: `  # Enable IP whitelist
  racore-cli domain ip-filter set --domain example.com --config '{"state":"on","type":"white","value":["1.2.3.4"]}'

  # Disable IP filter
  racore-cli domain ip-filter set --domain example.com --config '{"state":"off","value":[]}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/ip/filter", "post"),
}

var domainRefererFilterGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get referer filter configuration",
	Example: `  racore-cli domain referer-filter get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/referer/filter"),
}

var domainRefererFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set referer filter configuration",
	Example: `  # Enable referer whitelist
  racore-cli domain referer-filter set --domain example.com --config '{"state":"on","type":"white","values":["example.com"],"allow_empty":"on"}'

  # Disable referer filter
  racore-cli domain referer-filter set --domain example.com --config '{"state":"off","type":"off","values":["example.com"],"allow_empty":"on"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/referer/filter", "post"),
}

var domainUAFilterGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get user-agent filter configuration",
	Example: `  racore-cli domain ua-filter get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/user/agent/filter"),
}

var domainUAFilterSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set user-agent filter configuration",
	Example: `  # Enable UA blacklist
  racore-cli domain ua-filter set --domain example.com --config '{"state":"on","type":"black","values":["curl/*"]}'

  # Disable UA filter
  racore-cli domain ua-filter set --domain example.com --config '{"state":"off","values":[]}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/user/agent/filter", "post"),
}

var domainOriginProtocolGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get origin protocol policy",
	Example: `  racore-cli domain origin-protocol get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/origin/protocol/policy"),
}

var domainOriginProtocolSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin protocol policy",
	Example: `  racore-cli domain origin-protocol set --domain example.com --config '{"origin_protocol_policy":"https-only","origin_protocol_http_port":"80","origin_protocol_https_port":"443"}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/protocol/policy", "put"),
}

var domainHTTP2GetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get HTTP/2 configuration",
	Example: `  racore-cli domain http2 get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/http2"),
}

var domainHTTP2SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP/2 configuration",
	Example: `  # Enable HTTP/2
  racore-cli domain http2 set --domain example.com --config '{"enable":"on"}'

  # Disable HTTP/2
  racore-cli domain http2 set --domain example.com --config '{"enable":"off"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/http2", "put"),
}

var domainHTTP3GetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get HTTP/3 configuration",
	Example: `  racore-cli domain http3 get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/http3"),
}

var domainHTTP3SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP/3 configuration",
	Example: `  # Enable HTTP/3
  racore-cli domain http3 set --domain example.com --config '{"enable":"on"}'

  # Disable HTTP/3
  racore-cli domain http3 set --domain example.com --config '{"enable":"off"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/http3", "put"),
}

var domainTLSVersionGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get minimum TLS version",
	Example: `  racore-cli domain tls-version get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/min/tls/version"),
}

var domainTLSVersionSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set minimum TLS version",
	Example: `  racore-cli domain tls-version set --domain example.com --config '{"min_tls_version":"TLSv1.2_2021"}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/min/tls/version", "put"),
}

var domainCompressGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get page compression configuration",
	Example: `  racore-cli domain compress get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/page/compress"),
}

var domainCompressSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set page compression configuration",
	Example: `  # Enable compression
  racore-cli domain compress set --domain example.com --config '{"enable":"on"}'

  # Disable compression
  racore-cli domain compress set --domain example.com --config '{"enable":"off"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/page/compress", "put"),
}

var domainIPv6GetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get IPv6 configuration",
	Example: `  racore-cli domain ipv6 get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/ipv6"),
}

var domainIPv6SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set IPv6 configuration",
	Example: `  # Enable IPv6
  racore-cli domain ipv6 set --domain example.com --config '{"enable":"1"}'

  # Disable IPv6
  racore-cli domain ipv6 set --domain example.com --config '{"enable":"0"}'`,
	RunE: runDomainConfigSet("/API/cdn/domain/ipv6", "put"),
}

var domainCachePolicyGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get cache policy configuration",
	Example: `  racore-cli domain cache-policy get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/cache/conf"),
}

var domainCachePolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set cache policy configuration",
	Example: `  racore-cli domain cache-policy set --domain example.com --config '{"cache_policy_id":"658327ea-..."}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/cache/conf", "put"),
}

var domainOriginHostGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get origin host configuration",
	Example: `  racore-cli domain origin-host get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/origin/host"),
}

var domainOriginHostSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin host configuration",
	Example: `  racore-cli domain origin-host set --domain example.com --config '{"origin_host_type":"3","origin_host":"custom.example.com"}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/host", "put"),
}

var domainOriginTimeoutGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get origin connection timeout policy",
	Example: `  racore-cli domain origin-timeout get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/origin/connection/policy"),
}

var domainOriginTimeoutSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set origin connection timeout policy",
	Example: `  racore-cli domain origin-timeout set --domain example.com --config '{"connection_timeout":"10","response_timeout":"30"}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/origin/connection/policy", "put"),
}

var domainGeoRestrictionGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get geo restriction configuration",
	Example: `  racore-cli domain geo-restriction get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/geo/restriction"),
}

var domainGeoRestrictionSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set geo restriction configuration",
	Example: `  racore-cli domain geo-restriction set --domain example.com --config '{"restriction_type":"whitelist","items":["CN","US"]}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/geo/restriction", "put"),
}

var domainRequestHeadersGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get HTTP request headers",
	Example: `  racore-cli domain request-headers get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/http/request/headers"),
}

var domainRequestHeadersSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP request headers",
	Example: `  racore-cli domain request-headers set --domain example.com --config '{"headers":[{"key":"X-Custom","value":"test","action":"add"}]}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/http/request/headers", "post"),
}

var domainResponseHeadersGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get HTTP response headers",
	Example: `  racore-cli domain response-headers get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/http/response/headers"),
}

var domainResponseHeadersSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set HTTP response headers",
	Example: `  racore-cli domain response-headers set --domain example.com --config '{"headers":[{"key":"X-Frame-Options","value":"SAMEORIGIN","action":"add"}]}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/http/response/headers", "post"),
}

var domainRequestHeaderPolicyGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get request header policy",
	Example: `  racore-cli domain request-header-policy get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/request/header/policy"),
}

var domainRequestHeaderPolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set request header policy",
	Example: `  racore-cli domain request-header-policy set --domain example.com --config '{"origin_request_policy_id":"acba4595-..."}'`,
	RunE:  runDomainConfigSet("/API/cdn/domain/request/header/policy", "put"),
}

var domainResponseHeaderPolicyGetCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get response header policy",
	Example: `  racore-cli domain response-header-policy get --domain example.com`,
	RunE:    runDomainConfigGet("/API/cdn/domain/response/header/policy"),
}

var domainResponseHeaderPolicySetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set response header policy",
	Example: `  racore-cli domain response-header-policy set --domain example.com --config '{"response_headers_policy_id":"60669652-..."}'`,
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

// findMatchingCert tries to find a certificate that matches the given domain.
// It checks for exact matches and wildcard matches (*.parent.domain).
// Returns (certID, certName) if found, or ("", "") if no match.
func findMatchingCert(client *api.Client, domain string) (string, string, error) {
	rawResp, err := client.Get("/API/cdn/sslcert", nil)
	if err != nil {
		return "", "", err
	}

	var resp struct {
		Code int `json:"code"`
		Data []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Domain string `json:"domain"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return "", "", err
	}
	if resp.Code != 1 {
		return "", "", nil // silently skip if can't fetch certs
	}

	// Build list of wildcard patterns to try
	// For "v5.bbc.cfai.work": try "*.bbc.cfai.work", "*.cfai.work"
	parts := strings.Split(domain, ".")
	var wildcards []string
	for i := 1; i < len(parts)-1; i++ {
		wildcard := "*." + strings.Join(parts[i:], ".")
		wildcards = append(wildcards, wildcard)
	}

	for _, cert := range resp.Data {
		// Exact match
		if strings.EqualFold(cert.Domain, domain) {
			return cert.ID, cert.Domain, nil
		}
		// Wildcard match
		for _, wc := range wildcards {
			if strings.EqualFold(cert.Domain, wc) {
				return cert.ID, cert.Domain, nil
			}
		}
	}

	return "", "", nil
}

func executeDomainCreate(domain, domainType, source, sourceType, certID string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	// Normalize type value - API accepts lowercase English: oversea, live, video, dynamic, static, download
	typeNorm := map[string]string{
		"cdn": "static",
		"web": "static",
		"vod": "video",
	}
	lowerType := strings.ToLower(domainType)
	if mapped, ok := typeNorm[lowerType]; ok {
		domainType = mapped
	} else {
		domainType = lowerType
	}

	// Auto-match certificate if not explicitly provided
	certName := ""
	if certID == "" {
		matchedID, matchedName, err := findMatchingCert(client, domain)
		if err == nil && matchedID != "" {
			certID = matchedID
			certName = matchedName
		}
	}

	isSSL := "0"
	if certID != "" {
		isSSL = "1"
	}

	body := map[string]interface{}{
		"domain":      domain,
		"type":        domainType,
		"source_type": sourceType,
		"cache_type":  "1",
		"source_conf": []map[string]string{
			{
				"source": source,
				"type":   "1",
			},
		},
		"is_ssl":  isSSL,
		"cert_id": certID,
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

	result := fmt.Sprintf("Domain %s created successfully.", domain)
	if certName != "" {
		result += fmt.Sprintf("\nSSL: auto-matched certificate '%s' (ID: %s)", certName, certID)
	}
	return result, nil
}

func executeDomainDelete(domain string) (string, error) {
	client, err := newAuthenticatedClient()
	if err != nil {
		return "", err
	}

	body := map[string]interface{}{"domains": []string{domain}}
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

	body := map[string]interface{}{"domains": []string{domain}}
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

	body := map[string]interface{}{"domains": []string{domain}}
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
	certID, _ := cmd.Flags().GetString("cert-id")
	sourceType, _ := cmd.Flags().GetString("source-type")

	if domain == "" || domainType == "" || source == "" {
		return fmt.Errorf("--domain, --type, and --source flags are required")
	}

	result, err := executeDomainCreate(domain, domainType, source, sourceType, certID)
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
	domainCreateCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainCreateCmd.Flags().String("type", "", "CDN type: oversea, live, video, dynamic, static, download, web")
	domainCreateCmd.Flags().String("source", "", "Origin source address (IP or domain)")
	domainCreateCmd.Flags().String("cert-id", "", "SSL certificate ID to bind (enables HTTPS)")
	domainCreateCmd.Flags().String("source-type", "2", "Origin type: 1 (IP) or 2 (domain)")
	domainDeleteCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainEnableCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainDisableCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")

	// Get subcommand flags (--domain)
	domainSourceGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainSSLGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainEnforceHTTPSGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainIPFilterGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRefererFilterGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainUAFilterGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginProtocolGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainHTTP2GetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainHTTP3GetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainTLSVersionGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainCompressGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainIPv6GetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainCachePolicyGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginHostGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginTimeoutGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainGeoRestrictionGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRequestHeadersGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainResponseHeadersGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRequestHeaderPolicyGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainResponseHeaderPolicyGetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")

	// Set subcommand flags (--domain and --config)
	domainSourceSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainSourceSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainSSLSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainSSLSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainEnforceHTTPSSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainEnforceHTTPSSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainIPFilterSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainIPFilterSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainRefererFilterSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRefererFilterSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainUAFilterSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainUAFilterSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainOriginProtocolSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginProtocolSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainHTTP2SetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainHTTP2SetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainHTTP3SetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainHTTP3SetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainTLSVersionSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainTLSVersionSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainCompressSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainCompressSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainIPv6SetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainIPv6SetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainCachePolicySetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainCachePolicySetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainOriginHostSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginHostSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainOriginTimeoutSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainOriginTimeoutSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainGeoRestrictionSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainGeoRestrictionSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainRequestHeadersSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRequestHeadersSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainResponseHeadersSetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainResponseHeadersSetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainRequestHeaderPolicySetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainRequestHeaderPolicySetCmd.Flags().String("config", "", "Configuration JSON payload")
	domainResponseHeaderPolicySetCmd.Flags().String("domain", "", "CDN acceleration domain name (e.g., cdn.example.com)")
	domainResponseHeaderPolicySetCmd.Flags().String("config", "", "Configuration JSON payload")

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
