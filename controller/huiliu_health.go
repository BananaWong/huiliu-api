/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package controller

import (
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

type channelHealthModel struct {
	Model string              `json:"model"`
	Day   service.HealthStats `json:"day"`
}

type channelHealthRow struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Status       int                  `json:"status"`
	Managed      bool                 `json:"managed"`
	Group        string               `json:"group"`
	StatusReason string               `json:"status_reason,omitempty"`
	StatusTime   int64                `json:"status_time,omitempty"`
	NextRetryAt  int64                `json:"next_retry_at,omitempty"`
	Hour         service.HealthStats  `json:"hour"`
	Day          service.HealthStats  `json:"day"`
	Models       []channelHealthModel `json:"models"`
}

// GetChannelHealth lists every channel with its request outcomes over the last hour and day.
func GetChannelHealth(c *gin.Context) {
	now := time.Now().Unix()
	outcomes, err := service.LoadHealthOutcomes(now - 86400)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channels, err := model.GetAllChannels(0, 0, true, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	byChannel := map[int][]service.HealthOutcome{}
	for _, o := range outcomes {
		byChannel[o.ChannelID] = append(byChannel[o.ChannelID], o)
	}
	s := operation_setting.GetHuiliuHealthSetting()
	rows := make([]channelHealthRow, 0, len(channels))
	for _, ch := range channels {
		if ch == nil {
			continue
		}
		info := ch.GetOtherInfo()
		reason, _ := info["status_reason"].(string)
		var statusTime int64
		if v, ok := info["status_time"].(float64); ok {
			statusTime = int64(v)
		}
		list := byChannel[ch.Id]
		perModel := map[string][]service.HealthOutcome{}
		for _, o := range list {
			perModel[o.Model] = append(perModel[o.Model], o)
		}
		models := make([]channelHealthModel, 0, len(perModel))
		for name, items := range perModel {
			models = append(models, channelHealthModel{Model: name, Day: service.SummarizeOutcomes(items, now-86400)})
		}
		slices.SortFunc(models, func(a, b channelHealthModel) int { return strings.Compare(a.Model, b.Model) })
		rows = append(rows, channelHealthRow{
			ID:           ch.Id,
			Name:         ch.Name,
			Status:       ch.Status,
			Managed:      s.Enabled && ch.GetAutoBan() && !ch.ChannelInfo.IsMultiKey,
			Group:        ch.Group,
			StatusReason: reason,
			StatusTime:   statusTime,
			NextRetryAt:  service.NextRetryAt(ch.Status, info, s),
			Hour:         service.SummarizeOutcomes(list, now-3600),
			Day:          service.SummarizeOutcomes(list, now-86400),
			Models:       models,
		})
	}
	common.ApiSuccess(c, gin.H{"enabled": s.Enabled, "window_minutes": s.WindowMinutes, "channels": rows})
}

// Model status for signed-in users: no channel names and no upstream error text.
type modelStatusRow struct {
	Model         string  `json:"model"`
	Status        string  `json:"status"` // normal, degraded or unavailable
	Requests      int     `json:"requests"`
	SuccessRate   float64 `json:"success_rate"`
	P50Seconds    int     `json:"p50_seconds"`
	LastSuccessAt int64   `json:"last_success_at"`
}

type modelStatusCache struct {
	at   int64
	rows []modelStatusRow
}

var (
	modelStatusMu    sync.Mutex
	modelStatusByGrp = map[string]modelStatusCache{}
)

const modelStatusCacheSeconds = 60

// GetModelStatus reports, for each model the user's group can call, whether it works now.
func GetModelStatus(c *gin.Context) {
	group, err := model.GetUserGroup(c.GetInt("id"), false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	now := time.Now().Unix()
	modelStatusMu.Lock()
	cached, ok := modelStatusByGrp[group]
	modelStatusMu.Unlock()
	if ok && now-cached.at < modelStatusCacheSeconds {
		common.ApiSuccess(c, cached.rows)
		return
	}
	rows, err := buildModelStatus(group, now)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	modelStatusMu.Lock()
	modelStatusByGrp[group] = modelStatusCache{at: now, rows: rows}
	modelStatusMu.Unlock()
	common.ApiSuccess(c, rows)
}

func buildModelStatus(group string, now int64) ([]modelStatusRow, error) {
	channels, err := model.GetAllChannels(0, 0, true, false)
	if err != nil {
		return nil, err
	}
	type modelChannels struct {
		enabled, healthDisabled int
		ids                     []int
	}
	byModel := map[string]*modelChannels{}
	for _, ch := range channels {
		if ch == nil || !channelListHasGroup(ch.Group, group) {
			continue
		}
		_, healthMarked := service.ParseHealthMark(ch.GetOtherInfo())
		for name := range strings.SplitSeq(ch.Models, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			mc := byModel[name]
			if mc == nil {
				mc = &modelChannels{}
				byModel[name] = mc
			}
			mc.ids = append(mc.ids, ch.Id)
			switch {
			case ch.Status == common.ChannelStatusEnabled:
				mc.enabled++
			case ch.Status == common.ChannelStatusAutoDisabled && healthMarked:
				mc.healthDisabled++
			}
		}
	}
	outcomes, err := service.LoadHealthOutcomes(now - 86400)
	if err != nil {
		return nil, err
	}
	rows := make([]modelStatusRow, 0, len(byModel))
	for name, mc := range byModel {
		var mine []service.HealthOutcome
		for _, o := range outcomes {
			if o.Model == name && slices.Contains(mc.ids, o.ChannelID) {
				mine = append(mine, o)
			}
		}
		day := service.SummarizeOutcomes(mine, now-86400)
		hour := service.SummarizeOutcomes(mine, now-3600)
		status := "normal"
		switch {
		case mc.enabled == 0:
			status = "unavailable"
		case mc.healthDisabled > 0, hour.Requests >= 3 && hour.SuccessRate < 0.8:
			status = "degraded"
		}
		rows = append(rows, modelStatusRow{
			Model:         name,
			Status:        status,
			Requests:      day.Requests,
			SuccessRate:   day.SuccessRate,
			P50Seconds:    day.P50Seconds,
			LastSuccessAt: day.LastSuccessAt,
		})
	}
	slices.SortFunc(rows, func(a, b modelStatusRow) int { return strings.Compare(a.Model, b.Model) })
	return rows, nil
}

func channelListHasGroup(groups, group string) bool {
	for g := range strings.SplitSeq(groups, ",") {
		if strings.TrimSpace(g) == group {
			return true
		}
	}
	return false
}
