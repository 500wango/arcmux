package setting

// Waffo Pancake hosted checkout configuration. Gateway is enabled once
// MerchantID + PrivateKey + ProductID are populated (no separate Enabled
// flag, matching Stripe / Creem). StoreID + ProductID are operator-bound
// via SaveWaffoPancakeConfig.
import (
	"strings"

	"github.com/500wango/arcmux/setting/operation_setting"
)

var (
	WaffoPancakeMerchantID string
	WaffoPancakePrivateKey string
	WaffoPancakeReturnURL  string
	WaffoPancakeCurrency   string = "USD"
	WaffoPancakeUnitPrice  float64 = 1.0
	WaffoPancakeMinTopUp   int     = 1
	WaffoPancakeStoreID    string
	WaffoPancakeProductID  string
)

// GetWaffoPancakeCurrency resolves the settlement currency for Pancake checkouts.
// If WaffoPancakeCurrency is explicitly set (e.g. "CNY" or "USD"), it is used;
// otherwise it defaults to "CNY" if the site displays in CNY, else "USD".
func GetWaffoPancakeCurrency() string {
	if strings.TrimSpace(WaffoPancakeCurrency) != "" {
		return strings.ToUpper(strings.TrimSpace(WaffoPancakeCurrency))
	}
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeCNY {
		return "CNY"
	}
	return "USD"
}
