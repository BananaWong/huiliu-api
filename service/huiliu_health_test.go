/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package service

import (
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func errLog(ch int, req string, at int64, status int) *model.Log {
	return &model.Log{Type: model.LogTypeError, ChannelId: ch, RequestId: req, CreatedAt: at, ModelName: "gpt-image-2.6",
		Content: "upstream said no", Other: `{"status_code":` + strconv.Itoa(status) + `}`, UseTime: 2}
}

func okLog(ch int, req string, at int64, useTime int) *model.Log {
	return &model.Log{Type: model.LogTypeConsume, ChannelId: ch, RequestId: req, CreatedAt: at, ModelName: "gpt-image-2.6", UseTime: useTime}
}

func defaultHealthSetting() operation_setting.HuiliuHealthSetting {
	return operation_setting.HuiliuHealthSetting{Enabled: true, WindowMinutes: 30, ConsecutiveFailures: 3, MinRequests: 10,
		MaxFailureRate: 0.5, BaseBackoffMinutes: 30, MaxBackoffMinutes: 480, ResetAfterMinutes: 360}
}

func TestOutcomesFromLogsCountsRequestsNotRetries(t *testing.T) {
	logs := []*model.Log{
		// one request retried four times on channel 7: one failure, not four
		errLog(7, "r1", 100, 503), errLog(7, "r1", 101, 503), errLog(7, "r1", 102, 503), errLog(7, "r1", 103, 503),
		// a request that failed once, then succeeded on the same channel
		errLog(7, "r2", 110, 502), okLog(7, "r2", 112, 40),
		// caller errors do not count against the channel
		errLog(7, "r3", 120, 400), errLog(7, "r4", 121, 413),
		// a request that reached no channel is ignored
		errLog(0, "r5", 122, 503),
	}
	out := OutcomesFromLogs(logs)
	require.Len(t, out, 2)
	assert.False(t, out[0].Success)
	assert.Equal(t, 503, out[0].StatusCode)
	assert.Equal(t, int64(103), out[0].At)
	assert.True(t, out[1].Success)
	assert.Equal(t, 40, out[1].UseTime)
}

func TestCountsAgainstChannel(t *testing.T) {
	for _, s := range []int{0, 401, 402, 403, 404, 408, 429, 500, 502, 503, 504, 524} {
		assert.True(t, countsAgainstChannel(s), s)
	}
	for _, s := range []int{400, 413, 415, 422, 499} {
		assert.False(t, countsAgainstChannel(s), s)
	}
}

func TestSummarizeOutcomes(t *testing.T) {
	out := []HealthOutcome{
		{ChannelID: 1, At: 10, Success: true, UseTime: 20},
		{ChannelID: 1, At: 20, Success: false, StatusCode: 503, Error: "a"},
		{ChannelID: 1, At: 30, Success: true, UseTime: 40},
		{ChannelID: 1, At: 40, Success: false, StatusCode: 429, Error: "b"},
		{ChannelID: 1, At: 50, Success: false, StatusCode: 401, Error: "c"},
	}
	s := SummarizeOutcomes(out, 0)
	assert.Equal(t, 5, s.Requests)
	assert.Equal(t, 2, s.Successes)
	assert.Equal(t, 3, s.Failures)
	assert.Equal(t, 2, s.ConsecutiveFailures)
	assert.Equal(t, 0.4, s.SuccessRate)
	assert.Equal(t, 20, s.P50Seconds)
	assert.Equal(t, 40, s.P90Seconds)
	assert.Equal(t, int64(30), s.LastSuccessAt)
	assert.Equal(t, 401, s.LastStatusCode)
	assert.Equal(t, "c", s.LastError)

	later := SummarizeOutcomes(out, 35)
	assert.Equal(t, 2, later.Requests)
	assert.Equal(t, 0.0, later.SuccessRate)
}

func TestDecideHealthActionDisableAndBackoff(t *testing.T) {
	s := defaultHealthSetting()
	now := int64(1_000_000)

	// healthy: nothing to do
	a := DecideHealthAction(common.ChannelStatusEnabled, map[string]any{}, HealthStats{Requests: 5, Successes: 5, SuccessRate: 1}, now, s)
	assert.Empty(t, a.Kind)

	// three failures in a row: first strike
	a = DecideHealthAction(common.ChannelStatusEnabled, map[string]any{}, HealthStats{Requests: 3, Failures: 3, ConsecutiveFailures: 3, LastStatusCode: 503}, now, s)
	assert.Equal(t, "disable", a.Kind)
	assert.Equal(t, "huiliu-health#1: 3 consecutive failures, last status 503", a.Reason)

	// too many failures overall
	a = DecideHealthAction(common.ChannelStatusEnabled, map[string]any{}, HealthStats{Requests: 10, Successes: 5, Failures: 5, SuccessRate: 0.5, ConsecutiveFailures: 1}, now, s)
	assert.Equal(t, "disable", a.Kind)
	assert.Contains(t, a.Reason, "5 of 10 requests failed")

	// failing again soon after this job re-enabled it: next strike
	recovered := map[string]any{"status_reason": "huiliu-health#2: recovered", "status_time": float64(now - 600)}
	a = DecideHealthAction(common.ChannelStatusEnabled, recovered, HealthStats{Requests: 3, Failures: 3, ConsecutiveFailures: 3}, now, s)
	assert.Contains(t, a.Reason, "huiliu-health#3:")

	// failing long after the last recovery starts over
	old := map[string]any{"status_reason": "huiliu-health#4: recovered", "status_time": float64(now - 7*3600)}
	a = DecideHealthAction(common.ChannelStatusEnabled, old, HealthStats{Requests: 3, Failures: 3, ConsecutiveFailures: 3}, now, s)
	assert.Contains(t, a.Reason, "huiliu-health#1:")
}

func TestDecideHealthActionReenable(t *testing.T) {
	s := defaultHealthSetting()
	now := int64(1_000_000)
	disabled := func(strikes int, ago int64) map[string]any {
		return map[string]any{"status_reason": "huiliu-health#" + strconv.Itoa(strikes) + ": 3 consecutive failures", "status_time": float64(now - ago)}
	}
	// 1st strike waits 30 minutes, 2nd 60, 3rd 120; capped at 480
	assert.Empty(t, DecideHealthAction(common.ChannelStatusAutoDisabled, disabled(1, 29*60), HealthStats{}, now, s).Kind)
	a := DecideHealthAction(common.ChannelStatusAutoDisabled, disabled(1, 30*60), HealthStats{}, now, s)
	assert.Equal(t, "enable", a.Kind)
	assert.Equal(t, "huiliu-health#1: recovered", a.Reason)
	assert.Empty(t, DecideHealthAction(common.ChannelStatusAutoDisabled, disabled(2, 59*60), HealthStats{}, now, s).Kind)
	assert.Equal(t, int64(480*60), BackoffSeconds(9, s))
	assert.Equal(t, now-600+120*60, NextRetryAt(common.ChannelStatusAutoDisabled, disabled(3, 600), s))

	// never touches a channel someone else disabled
	other := map[string]any{"status_reason": "Your credit balance is too low", "status_time": float64(now - 99999)}
	assert.Empty(t, DecideHealthAction(common.ChannelStatusAutoDisabled, other, HealthStats{}, now, s).Kind)
	assert.Empty(t, DecideHealthAction(common.ChannelStatusManuallyDisabled, disabled(1, 99999), HealthStats{}, now, s).Kind)
	assert.Zero(t, NextRetryAt(common.ChannelStatusAutoDisabled, other, s))
}
