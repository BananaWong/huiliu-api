/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// HuiliuPrivacySetting controls what callers can learn about the providers behind a
// channel. With HideUpstreamWording on, an error a channel produced reaches the caller,
// and the caller's own error logs, without the upstream's wording: a fixed sentence when
// the upstream could not serve the request, otherwise the text with every host masked.
// Administrators still read the original text in the error logs. The option is
// `huiliu_privacy_setting.hide_upstream_wording`.
type HuiliuPrivacySetting struct {
	HideUpstreamWording bool `json:"hide_upstream_wording"`
}

var huiliuPrivacySetting = HuiliuPrivacySetting{HideUpstreamWording: true}

func init() {
	config.GlobalConfig.Register("huiliu_privacy_setting", &huiliuPrivacySetting)
}

func GetHuiliuPrivacySetting() HuiliuPrivacySetting {
	return huiliuPrivacySetting
}
