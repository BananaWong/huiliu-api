/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package common

import "github.com/MoeForge/moego/regexp"

var (
	anyURLPattern    = regexp.MustCompile(`(?i)\b(?:https?|wss?)://[^\s"'<>{}]+`)
	anyIPv4Pattern   = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	anyDomainPattern = regexp.MustCompile(`(?i)\b(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z][a-z0-9-]{0,61}[a-z]\b`)
)

// gatewayOwnErrorCodes are errors this gateway raises itself after a channel was
// picked; their wording is ours and tells the caller what to fix, so it is kept.
var gatewayOwnErrorCodes = map[string]bool{
	"insufficient_user_quota":        true,
	"pre_consume_token_quota_failed": true,
	"model_price_error":              true,
	"access_denied":                  true,
	"get_channel_failed":             true,
	"sensitive_words_detected":       true,
	"request_rejected":               true,
	"response_rejected":              true,
	"read_request_body_failed":       true,
	"convert_request_failed":         true,
	"bad_request_body":               true,
	"query_data_error":               true,
	"update_data_error":              true,
}

// MaskAllHosts masks every URL, domain name and IPv4 address in text.
// MaskSensitiveInfo masks only the channel's own address, but a relay's error
// often names the provider behind it.
func MaskAllHosts(text string) string {
	text = anyURLPattern.ReplaceAllString(text, "***")
	text = anyIPv4Pattern.ReplaceAllString(text, "***")
	return anyDomainPattern.ReplaceAllString(text, "***")
}

// UpstreamUnavailableStatus reports whether an upstream status means the provider
// could not serve the request (its own credentials, balance, rate limit, timeout or
// failure) rather than a problem in the caller's request.
func UpstreamUnavailableStatus(status int) bool {
	switch status {
	case 401, 402, 403, 408, 429:
		return true
	}
	return status >= 500
}

// HideUpstreamWording returns what a caller may read of an error a channel produced:
// unavailable (a fixed sentence) when the upstream could not serve the request,
// otherwise the text with every host masked. Errors the gateway raises itself keep
// their wording, with hosts masked.
func HideUpstreamWording(text string, status int, errorCode string, unavailable string) string {
	if UpstreamUnavailableStatus(status) && !gatewayOwnErrorCodes[errorCode] {
		return unavailable
	}
	return MaskAllHosts(text)
}
