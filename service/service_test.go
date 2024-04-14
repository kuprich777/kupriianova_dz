package service

import (
	"DZ_ITOG/models"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestConvertAmount(t *testing.T) {
	t.Run("same currency conversion", func(t *testing.T) {
		amount, err := ConvertAmount(100, "USD", "USD")
		if err != nil {
			t.Errorf("ConvertAmount returned an error for same currency conversion: %v", err)
		}
		if amount != 100 {
			t.Errorf("ConvertAmount changed the amount for same currency conversion: got %v want %v", amount, 100)
		}
	})

	t.Run("different currency conversion", func(t *testing.T) {
		originalFetch := FetchCurrencyRates
		defer func() { FetchCurrencyRates = originalFetch }()

		FetchCurrencyRates = func(baseCurrency string) (models.CurrencyRates, error) {
			return models.CurrencyRates{Rates: map[string]float64{"EUR": 0.9}}, nil
		}

		amount, err := ConvertAmount(100, "USD", "EUR")
		if err != nil {
			t.Errorf("ConvertAmount returned an error: %v", err)
		}
		if amount != 90 {
			t.Errorf("ConvertAmount returned incorrect conversion: got %v want %v", amount, 90)
		}
	})
}
func TestFetchCurrencyRates(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.freecurrencyapi.com/v2/latest?apikey=fca_live_O0Kw1fj8ul0ZaVlBnTZxn48GBL3X0vr8TSL95HqI&base_currency=USD",
		httpmock.NewStringResponder(200, `{"rates":{"USD":1,"EUR":0.9}}`))

	rates, err := FetchCurrencyRates("USD")
	if err != nil {
		t.Fatalf("FetchCurrencyRates returned an error: %v", err)
	}

	expectedRate := 0.9
	if rates.Rates["EUR"] != expectedRate {
		t.Errorf("Expected EUR rate of %v, got %v", expectedRate, rates.Rates["EUR"])
	}
}
