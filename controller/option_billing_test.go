package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/500wango/arcmux/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updateBillingOptionForTest(t *testing.T, body string) (bool, string) {
	t.Helper()
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPut, "/api/option/", strings.NewReader(body))

	UpdateOption(context)

	var payload struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	return payload.Success, payload.Message
}

func TestUpdateOptionRejectsInvalidBillingExpression(t *testing.T) {
	success, message := updateBillingOptionForTest(t,
		`{"key":"billing_setting.billing_expr","value":"{\"unsafe-model\":\"tier(\\\"bad\\\", -1)\"}"}`)

	assert.False(t, success)
	assert.Contains(t, message, "must be non-negative")
}

func TestUpdateOptionRejectsInvalidQuotaPerUnit(t *testing.T) {
	for _, value := range []string{"0", "-1", "NaN", "+Inf", "2147483648"} {
		t.Run(value, func(t *testing.T) {
			success, message := updateBillingOptionForTest(t,
				`{"key":"QuotaPerUnit","value":"`+value+`"}`)
			assert.False(t, success)
			assert.Contains(t, message, "finite number")
		})
	}
}
