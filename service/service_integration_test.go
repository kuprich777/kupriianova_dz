package service

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestIntegrationFetchCurrencyRates(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.freecurrencyapi.com/v2/latest?apikey=fca_live_O0Kw1fj8ul0ZaVlBnTZxn48GBL3X0vr8TSL95HqI&base_currency=USD",
		httpmock.NewStringResponder(500, `{"error":"Internal Server Error"}`))


	httpmock.RegisterResponder("GET", "https://v6.exchangerate-api.com/v6/abe4f5b30553ba88e0824fdd/latest/USD",
		httpmock.NewStringResponder(200, `{"conversion_rates":{"USD":1,"EUR":0.9}}`))

	rates, err := FetchCurrencyRates("USD")
	if err != nil {
		t.Fatalf("Failed to fetch currency rates: %v", err)
	}

	expected := 0.9
	if rates.Rates["EUR"] != expected {
		t.Errorf("Expected EUR rate of %v, got %v", expected, rates.Rates["EUR"])
	}
}
