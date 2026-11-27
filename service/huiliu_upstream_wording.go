/*
This file is part of a modified version of new-api (https://github.com/QuantumNous/new-api),
licensed under the GNU Affero General Public License v3.0. See HUILIU_CHANGES.md.
*/

package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// HideUpstreamWordingForCaller rewrites the final error of a request that reached a
// channel so the response carries no upstream wording (see common.HideUpstreamWording).
// It runs after the error logs were written, so administrators keep the original text.
func HideUpstreamWordingForCaller(c *gin.Context, e *types.NewAPIError) {
	if e == nil || !operation_setting.GetHuiliuPrivacySetting().HideUpstreamWording {
		return
	}
	if common.GetContextKeyInt(c, constant.ContextKeyChannelId) == 0 {
		return
	}
	unavailable := i18n.T(c, i18n.MsgRelayUpstreamUnavailable)
	code := string(e.GetErrorCode())
	e.SetMessage(common.HideUpstreamWording(e.Error(), e.StatusCode, code, unavailable))
	switch relayError := e.RelayError.(type) {
	case types.OpenAIError:
		relayError.Message = common.HideUpstreamWording(relayError.Message, e.StatusCode, code, unavailable)
		relayError.Metadata = nil
		e.RelayError = relayError
	case types.ClaudeError:
		relayError.Message = common.HideUpstreamWording(relayError.Message, e.StatusCode, code, unavailable)
		e.RelayError = relayError
	}
	e.Metadata = nil
}
