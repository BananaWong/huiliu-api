/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package service

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

// Channel health is computed from the request logs: a consume log is a request a
// channel served, an error log (written when ERROR_LOG_ENABLED is on) is an attempt that
// failed. One request retried on the same channel writes several error logs, so outcomes
// are counted per request id and channel. Errors the caller caused (bad parameters,
// content refused) do not count against a channel.

// HealthReasonPrefix marks a status this job set; it only re-enables channels whose
// status reason carries it, never a channel an administrator or another rule disabled.
const HealthReasonPrefix = "huiliu-health#"

const (
	healthLoopInterval = 5 * time.Minute
	healthLogLimit     = 200000
	healthErrorPreview = 200
)

// HealthOutcome is the result of one request on one channel.
type HealthOutcome struct {
	ChannelID  int
	Model      string
	At         int64
	Success    bool
	UseTime    int
	StatusCode int
	Error      string
}

// HealthStats summarises outcomes over a window.
type HealthStats struct {
	Requests            int     `json:"requests"`
	Successes           int     `json:"successes"`
	Failures            int     `json:"failures"`
	SuccessRate         float64 `json:"success_rate"`
	P50Seconds          int     `json:"p50_seconds"`
	P90Seconds          int     `json:"p90_seconds"`
	LastSuccessAt       int64   `json:"last_success_at"`
	LastFailureAt       int64   `json:"last_failure_at"`
	LastError           string  `json:"last_error,omitempty"`
	LastStatusCode      int     `json:"last_status_code,omitempty"`
	ConsecutiveFailures int     `json:"consecutive_failures"`
}

// countsAgainstChannel reports whether a failed attempt with this upstream status says
// something about the channel. 0 is a network error or timeout.
func countsAgainstChannel(status int) bool {
	switch {
	case status == 0, status >= 500:
		return true
	case status == 401, status == 402, status == 403, status == 404, status == 408, status == 429:
		return true
	default:
		return false
	}
}

// OutcomesFromLogs turns consume and error logs into per-request outcomes per channel.
// A request with any consume log on a channel succeeded there; otherwise it failed if one
// of its errors counts against the channel; requests with only caller errors are dropped.
func OutcomesFromLogs(logs []*model.Log) []HealthOutcome {
	type key struct {
		channel int
		request string
	}
	byKey := map[key]*HealthOutcome{}
	order := make([]key, 0, len(logs))
	for i, l := range logs {
		if l == nil || l.ChannelId <= 0 || (l.Type != model.LogTypeConsume && l.Type != model.LogTypeError) {
			continue
		}
		k := key{l.ChannelId, l.RequestId}
		if l.RequestId == "" {
			k.request = "#" + strconv.Itoa(i)
		}
		success := l.Type == model.LogTypeConsume
		status := 0
		if !success {
			status = logStatusCode(l.Other)
			if !countsAgainstChannel(status) {
				continue
			}
		}
		o, ok := byKey[k]
		if !ok {
			o = &HealthOutcome{ChannelID: l.ChannelId, Model: l.ModelName}
			byKey[k] = o
			order = append(order, k)
		}
		if l.CreatedAt > o.At {
			o.At = l.CreatedAt
		}
		if success {
			o.Success = true
			o.UseTime = l.UseTime
			o.StatusCode = 0
			o.Error = ""
		} else if !o.Success {
			o.StatusCode = status
			o.Error = previewError(l.Content)
			if l.UseTime > o.UseTime {
				o.UseTime = l.UseTime
			}
		}
	}
	out := make([]HealthOutcome, 0, len(order))
	for _, k := range order {
		out = append(out, *byKey[k])
	}
	slices.SortStableFunc(out, func(a, b HealthOutcome) int { return int(a.At - b.At) })
	return out
}

func logStatusCode(other string) int {
	if other == "" {
		return 0
	}
	var m map[string]any
	if err := common.UnmarshalJsonStr(other, &m); err != nil {
		return 0
	}
	switch v := m["status_code"].(type) {
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

func previewError(s string) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > healthErrorPreview {
		return string(r[:healthErrorPreview]) + "…"
	}
	return s
}

// SummarizeOutcomes computes stats for outcomes at or after since (outcomes sorted by time).
func SummarizeOutcomes(outcomes []HealthOutcome, since int64) HealthStats {
	var s HealthStats
	var times []int
	for _, o := range outcomes {
		if o.At < since {
			continue
		}
		s.Requests++
		if o.Success {
			s.Successes++
			s.ConsecutiveFailures = 0
			times = append(times, o.UseTime)
			s.LastSuccessAt = max(s.LastSuccessAt, o.At)
			continue
		}
		s.Failures++
		s.ConsecutiveFailures++
		if o.At >= s.LastFailureAt {
			s.LastFailureAt = o.At
			s.LastError = o.Error
			s.LastStatusCode = o.StatusCode
		}
	}
	if s.Requests > 0 {
		s.SuccessRate = math.Round(float64(s.Successes)/float64(s.Requests)*1000) / 1000
	}
	if len(times) > 0 {
		slices.Sort(times)
		s.P50Seconds = times[(len(times)-1)/2]
		s.P90Seconds = times[int(math.Ceil(float64(len(times))*0.9))-1]
	}
	return s
}

// HealthMark is what this job wrote into a channel's status reason.
type HealthMark struct {
	Strikes   int
	Recovered bool
	At        int64
}

// ParseHealthMark reads the job's marker from a channel's other_info, if any.
func ParseHealthMark(info map[string]any) (HealthMark, bool) {
	reason, _ := info["status_reason"].(string)
	rest, ok := strings.CutPrefix(reason, HealthReasonPrefix)
	if !ok {
		return HealthMark{}, false
	}
	num, text, _ := strings.Cut(rest, ":")
	strikes, err := strconv.Atoi(strings.TrimSpace(num))
	if err != nil || strikes < 1 {
		return HealthMark{}, false
	}
	var at int64
	switch v := info["status_time"].(type) {
	case float64:
		at = int64(v)
	case int64:
		at = v
	case int:
		at = int64(v)
	}
	return HealthMark{Strikes: strikes, Recovered: strings.TrimSpace(text) == "recovered", At: at}, true
}

// BackoffSeconds is how long the job keeps a channel disabled after its strikes-th disable.
func BackoffSeconds(strikes int, s operation_setting.HuiliuHealthSetting) int64 {
	minutes := float64(s.BaseBackoffMinutes) * math.Pow(2, float64(max(strikes, 1)-1))
	return int64(math.Min(minutes, float64(s.MaxBackoffMinutes))) * 60
}

// HealthAction is what the job does to one channel.
type HealthAction struct {
	Kind   string // "", "disable" or "enable"
	Reason string
}

// DecideHealthAction decides on one channel. stats must cover the decision window, starting
// no earlier than the channel's last re-enable by this job, so old failures cannot disable
// it again right after it comes back.
func DecideHealthAction(status int, info map[string]any, stats HealthStats, now int64, s operation_setting.HuiliuHealthSetting) HealthAction {
	mark, marked := ParseHealthMark(info)
	switch status {
	case common.ChannelStatusEnabled:
		tooMany := stats.ConsecutiveFailures >= s.ConsecutiveFailures
		tooOften := stats.Requests >= s.MinRequests && 1-stats.SuccessRate >= s.MaxFailureRate
		if !tooMany && !tooOften {
			return HealthAction{}
		}
		strikes := 1
		if marked && mark.Recovered && now-mark.At < int64(s.ResetAfterMinutes)*60 {
			strikes = mark.Strikes + 1
		}
		why := fmt.Sprintf("%d consecutive failures", stats.ConsecutiveFailures)
		if !tooMany {
			why = fmt.Sprintf("%d of %d requests failed", stats.Failures, stats.Requests)
		}
		if stats.LastStatusCode > 0 {
			why += fmt.Sprintf(", last status %d", stats.LastStatusCode)
		}
		return HealthAction{Kind: "disable", Reason: fmt.Sprintf("%s%d: %s", HealthReasonPrefix, strikes, why)}
	case common.ChannelStatusAutoDisabled:
		if !marked || mark.Recovered || mark.At == 0 {
			return HealthAction{}
		}
		if now < mark.At+BackoffSeconds(mark.Strikes, s) {
			return HealthAction{}
		}
		return HealthAction{Kind: "enable", Reason: fmt.Sprintf("%s%d: recovered", HealthReasonPrefix, mark.Strikes)}
	default:
		return HealthAction{}
	}
}

// NextRetryAt is when the job will put an auto-disabled channel back, or 0.
func NextRetryAt(status int, info map[string]any, s operation_setting.HuiliuHealthSetting) int64 {
	mark, ok := ParseHealthMark(info)
	if status != common.ChannelStatusAutoDisabled || !ok || mark.Recovered || mark.At == 0 {
		return 0
	}
	return mark.At + BackoffSeconds(mark.Strikes, s)
}

// LoadHealthOutcomes reads the logs since the given time and returns per-request outcomes.
func LoadHealthOutcomes(since int64) ([]HealthOutcome, error) {
	var logs []*model.Log
	err := model.LOG_DB.
		Select("id", "created_at", "type", "model_name", "use_time", "channel_id", "request_id", "other", "content").
		Where("created_at >= ? AND type IN ? AND channel_id > ?", since, []int{model.LogTypeConsume, model.LogTypeError}, 0).
		Order("id desc").Limit(healthLogLimit).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return OutcomesFromLogs(logs), nil
}

var healthRunMu sync.Mutex

// RunHealthCheckOnce applies the job to every managed channel once.
func RunHealthCheckOnce() {
	healthRunMu.Lock()
	defer healthRunMu.Unlock()
	s := operation_setting.GetHuiliuHealthSetting()
	if !s.Enabled {
		return
	}
	now := time.Now().Unix()
	outcomes, err := LoadHealthOutcomes(now - int64(s.WindowMinutes)*60)
	if err != nil {
		common.SysError(common.LogText("channel health: failed to read logs: %v", err))
		return
	}
	byChannel := map[int][]HealthOutcome{}
	for _, o := range outcomes {
		byChannel[o.ChannelID] = append(byChannel[o.ChannelID], o)
	}
	channels, err := model.GetAllChannels(0, 0, true, false)
	if err != nil {
		common.SysError(common.LogText("channel health: failed to read channels: %v", err))
		return
	}
	changed := false
	for _, ch := range channels {
		if ch == nil || !ch.GetAutoBan() || ch.ChannelInfo.IsMultiKey {
			continue
		}
		info := ch.GetOtherInfo()
		since := now - int64(s.WindowMinutes)*60
		if mark, ok := ParseHealthMark(info); ok && mark.Recovered && mark.At > since {
			since = mark.At
		}
		action := DecideHealthAction(ch.Status, info, SummarizeOutcomes(byChannel[ch.Id], since), now, s)
		switch action.Kind {
		case "disable":
			if model.UpdateChannelStatus(ch.Id, "", common.ChannelStatusAutoDisabled, action.Reason) {
				changed = true
				common.SysLog(common.LogText("channel health: disabled channel #%d (%s): %s", ch.Id, ch.Name, action.Reason))
			}
		case "enable":
			if model.UpdateChannelStatus(ch.Id, "", common.ChannelStatusEnabled, action.Reason) {
				changed = true
				common.SysLog(common.LogText("channel health: re-enabled channel #%d (%s)", ch.Id, ch.Name))
			}
		}
	}
	if changed {
		model.InitChannelCache()
	}
}

// StartHealthLoop runs the job every five minutes on the master node.
func StartHealthLoop() {
	if !common.IsMasterNode {
		return
	}
	go func() {
		ticker := time.NewTicker(healthLoopInterval)
		defer ticker.Stop()
		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						common.SysError(common.LogText("channel health: panic: %v", r))
					}
				}()
				RunHealthCheckOnce()
			}()
		}
	}()
}
