/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

// HuiliuHealthSetting controls the passive channel health job: it reads the request
// logs, disables a channel that keeps failing and puts it back after a backoff so real
// requests can try it again. Only channels with automatic disabling (auto_ban) on are
// touched. Every field is an option (`huiliu_health_setting.<field>`), so the job can be
// switched off or retuned without a release.
type HuiliuHealthSetting struct {
	Enabled             bool    `json:"enabled"`
	WindowMinutes       int     `json:"window_minutes"`
	ConsecutiveFailures int     `json:"consecutive_failures"`
	MinRequests         int     `json:"min_requests"`
	MaxFailureRate      float64 `json:"max_failure_rate"`
	BaseBackoffMinutes  int     `json:"base_backoff_minutes"`
	MaxBackoffMinutes   int     `json:"max_backoff_minutes"`
	ResetAfterMinutes   int     `json:"reset_after_minutes"`
}

var huiliuHealthSetting = HuiliuHealthSetting{
	Enabled:             true,
	WindowMinutes:       30,
	ConsecutiveFailures: 3,
	MinRequests:         10,
	MaxFailureRate:      0.5,
	BaseBackoffMinutes:  30,
	MaxBackoffMinutes:   480,
	ResetAfterMinutes:   360,
}

func init() {
	config.GlobalConfig.Register("huiliu_health_setting", &huiliuHealthSetting)
}

// GetHuiliuHealthSetting returns the setting with out-of-range values replaced by the defaults.
func GetHuiliuHealthSetting() HuiliuHealthSetting {
	s := huiliuHealthSetting
	if s.WindowMinutes < 5 {
		s.WindowMinutes = 30
	}
	if s.ConsecutiveFailures < 2 {
		s.ConsecutiveFailures = 3
	}
	if s.MinRequests < 2 {
		s.MinRequests = 10
	}
	if s.MaxFailureRate <= 0 || s.MaxFailureRate > 1 {
		s.MaxFailureRate = 0.5
	}
	if s.BaseBackoffMinutes < 1 {
		s.BaseBackoffMinutes = 30
	}
	if s.MaxBackoffMinutes < s.BaseBackoffMinutes {
		s.MaxBackoffMinutes = s.BaseBackoffMinutes
	}
	if s.ResetAfterMinutes < 1 {
		s.ResetAfterMinutes = 360
	}
	return s
}
