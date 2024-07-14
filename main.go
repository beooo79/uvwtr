package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hectormalot/omgo"
)

type GeoResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

func main() {

	cityName := "Würzburg"
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=de&format=json", cityName)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var geoResp GeoResponse
	fmt.Println(string(url))
	fmt.Println(string(body))
	json.Unmarshal(body, &geoResp)

	if len(geoResp.Results) > 0 {
		fmt.Printf("Latitude: %f, Longitude: %f\n", geoResp.Results[0].Latitude, geoResp.Results[0].Longitude)
	} else {
		fmt.Println("No results found")
		return
	}

	c, _ := omgo.NewClient()

	// Get the current weather
	loc, _ := omgo.NewLocation(geoResp.Results[0].Latitude, geoResp.Results[0].Longitude)
	res, _ := c.CurrentWeather(context.Background(), loc, nil)
	fmt.Printf("The temperature in %s is %f (code: %f)", cityName, res.Temperature, res.WeatherCode)
	fmt.Println()

	// Get the humidity and cloud cover forecast,
	// including the last 2 days and non-metric units
	opts := omgo.Options{
		TemperatureUnit:   "celsius",
		WindspeedUnit:     "kmh",
		PrecipitationUnit: "mm",
		Timezone:          "Europe/Berlin",
		PastDays:          2,
		HourlyMetrics:     []string{"cloudcover, relativehumidity_2m"},
		DailyMetrics:      []string{"temperature_2m_max"},
	}

	fore, _ := c.Forecast(context.Background(), loc, &opts)
	fmt.Println("forecast", fore)
	// res.HourlyMetrics["cloudcover"] contains an array of cloud coverage predictions
	// res.HourlyMetrics["relativehumidity_2m"] contains an array of relative humidity predictions
	// res.HourlyTimes contains the timestamps for each prediction
	// res.DailyMetrics["temperature_2m_max"] contains daily maximum values for the temperature_2m metric
	// res.DailyTimes contains the timestamps for all daily predictions
}
