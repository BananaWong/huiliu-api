/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package common

import (
	"github.com/QuantumNous/new-api/relaykit/types"

	"github.com/MoeForge/moego/regexp"
)

var (
	anyURLPattern    = regexp.MustCompile(`(?i)\b(?:https?|wss?)://[^\s"'<>{}]+`)
	anyIPv4Pattern   = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	anyDomainPattern = regexp.MustCompile(`(?i)\b(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z][a-z0-9-]{0,61}[a-z]\b`)
)

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
// otherwise the text with every host masked. Errors the gateway builds itself keep
// their wording whatever their status, with hosts masked: those have the type
// types.ErrorTypeNewAPIError (types.NewError, types.NewErrorWithStatusCode), while
// every other type carries what the upstream sent back or how reaching it failed. The
// type, not the code, tells them apart, because a relay upstream answers with codes
// such as insufficient_user_quota that read like the gateway's own.
func HideUpstreamWording(text string, status int, errorType string, unavailable string) string {
	if errorType != string(types.ErrorTypeNewAPIError) && UpstreamUnavailableStatus(status) {
		return unavailable
	}
	return MaskAllHosts(text)
}
