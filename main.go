package main

import (
	"context"
	"fmt"

	"github.com/hectormalot/omgo"
)

func main() {
	c, _ := omgo.NewClient()

	// Get the current weather for amsterdam
	city := "Würzburg"
	loc, _ := omgo.NewLocation(49.791608493420576, 9.949478410503758)
	res, _ := c.CurrentWeather(context.Background(), loc, nil)
	fmt.Printf("The temperature in %s is %f", city, res.Temperature)
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
