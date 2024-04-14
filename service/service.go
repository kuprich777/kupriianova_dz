package service

import (
	"DZ_ITOG/models"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func ConvertAmount(amount float64, baseCurrency string, targetCurrency string) (float64, error) {
	log.Infof("Starting conversion: %v from %s to %s", amount, baseCurrency, targetCurrency)

	if baseCurrency == targetCurrency {
		log.Info("No conversion needed as both currencies are the same.")
		return amount, nil
	}

	rates, err := FetchCurrencyRates(baseCurrency)
	if err != nil {
		log.Errorf("Failed to fetch currency rates: %v", err)
		return 0, err
	}

	if rate, ok := rates.Rates[targetCurrency]; ok {
		convertedAmount := amount * rate
		log.Infof("Conversion successful: %v %s is %v %s", amount, baseCurrency, convertedAmount, targetCurrency)
		return convertedAmount, nil
	} else {
		log.Errorf("Rate for target currency %s not found", targetCurrency)
		return 0, fmt.Errorf("rate for target currency %s not found", targetCurrency)
	}
}

const currencyAPIURL = "https://api.freecurrencyapi.com/v2/latest"
const apiKey = "fca_live_O0Kw1fj8ul0ZaVlBnTZxn48GBL3X0vr8TSL95HqI"
const backupAPIURL = "https://v6.exchangerate-api.com/v6/abe4f5b30553ba88e0824fdd/latest"
const backupAPIKey = "abe4f5b30553ba88e0824fdd"

var FetchCurrencyRates = fetchCurrencyRates

func fetchCurrencyRates(baseCurrency string) (models.CurrencyRates, error) {
	var rates models.CurrencyRates

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr, Timeout: time.Second * 10}

	url := fmt.Sprintf("https://api.freecurrencyapi.com/v2/latest?apikey=%s&base_currency=%s", apiKey, baseCurrency)
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Warnf("Primary API failed, attempting backup API: %v", err)
		url = fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s?apikey=%s", backupAPIKey, baseCurrency, backupAPIKey)
		resp, err = client.Get(url)
		if err != nil {
			return rates, fmt.Errorf("error fetching currency rates from backup API: %w", err)
		}
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return rates, fmt.Errorf("error decoding currency rates: %w", err)
	}
	log.Infof("API Response: %+v", rates)

	return rates, nil
}
