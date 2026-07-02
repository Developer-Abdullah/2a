package billing

import (
	"os"
	"strconv"
)

const defaultGraceDays = 14

// FromEnv builds a provider registry from environment variables, including only the providers whose
// credentials are actually configured. It also returns the configured grace-period length in days.
//
//	PAYMOB_API_KEY, PAYMOB_INTEGRATION_ID, PAYMOB_IFRAME_ID, PAYMOB_HMAC_SECRET, PAYMOB_BASE_URL
//	MYFATOORAH_API_TOKEN, MYFATOORAH_WEBHOOK_SECRET, MYFATOORAH_BASE_URL
//	GRACE_PERIOD_DAYS (default 14)
func FromEnv() (*Registry, int) {
	var providers []Provider

	if key := os.Getenv("PAYMOB_API_KEY"); key != "" {
		providers = append(providers, NewPaymobProvider(
			key,
			os.Getenv("PAYMOB_INTEGRATION_ID"),
			os.Getenv("PAYMOB_IFRAME_ID"),
			os.Getenv("PAYMOB_HMAC_SECRET"),
			os.Getenv("PAYMOB_BASE_URL"),
		))
	}

	if token := os.Getenv("MYFATOORAH_API_TOKEN"); token != "" {
		providers = append(providers, NewMyFatoorahProvider(
			token,
			os.Getenv("MYFATOORAH_WEBHOOK_SECRET"),
			os.Getenv("MYFATOORAH_BASE_URL"),
		))
	}

	grace := defaultGraceDays
	if v := os.Getenv("GRACE_PERIOD_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			grace = n
		}
	}

	return NewRegistry(providers...), grace
}
