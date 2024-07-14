package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

//go:embed templates
var templateHTML embed.FS

var model *Model

type Model struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	UVIndex   int     `json:"uvIndex"`
	CityName  string  `json:"cityName"`
}

type GeoResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

type MetResponse struct {
	Latitude             float64           `json:"latitude"`
	Longitude            float64           `json:"longitude"`
	GenerationTimeMs     float64           `json:"generationtime_ms"`
	UtcOffsetSeconds     int               `json:"utc_offset_seconds"`
	Timezone             string            `json:"timezone"`
	TimezoneAbbreviation string            `json:"timezone_abbreviation"`
	Elevation            float64           `json:"elevation"`
	HourlyUnits          map[string]string `json:"hourly_units"`
	Hourly               HourlyData        `json:"hourly"`
	DailyUnits           map[string]string `json:"daily_units"`
	Daily                DailyData         `json:"daily"`
}

type HourlyData struct {
	Time          []string  `json:"time"`
	Temperature2M []float64 `json:"temperature_2m"`
	WeatherCode   []int     `json:"weather_code"`
	Cape          []float64 `json:"cape"`
}

type DailyData struct {
	Time               []string  `json:"time"`
	UvIndexMax         []float64 `json:"uv_index_max"`
	UvIndexClearSkyMax []float64 `json:"uv_index_clear_sky_max"`
}

type GeoLocation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func geoLocationForCity(cityName string) (*GeoLocation, error) {
	client := &http.Client{}

	encodedCityName := url.QueryEscape(html.EscapeString(cityName))
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=de&format=json", encodedCityName)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var geoResp GeoResponse
	json.Unmarshal(body, &geoResp)

	if len(geoResp.Results) > 0 {
		return &GeoLocation{
			Name:      cityName,
			Latitude:  geoResp.Results[0].Latitude,
			Longitude: geoResp.Results[0].Longitude,
		}, nil
	} else {
		fmt.Println("No results found")
		return nil, errors.New("geolocation for city " + cityName + " not found")
	}
}

func updateFromMeteo(geoLocation *GeoLocation) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m,weather_code,cape&daily=uv_index_max,uv_index_clear_sky_max&timezone=Europe%sBerlin&forecast_days=1", geoLocation.Latitude, geoLocation.Longitude, "%2F")
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var metResp MetResponse
	json.Unmarshal(body, &metResp)

	model.UVIndex = int(metResp.Daily.UvIndexMax[0])
}

func getMapHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFS(templateHTML, "templates/map.html"))
	params := req.URL.Query()
	model.CityName = ""
	model.Latitude = 0
	model.Longitude = 0
	for k, v := range params {
		if k == "cityName" {
			geoResp, err := geoLocationForCity(v[0])
			if err == nil {
				model.CityName = geoResp.Name
				model.Latitude = geoResp.Latitude
				model.Longitude = geoResp.Longitude
				updateFromMeteo(geoResp)
			}
		}
	}
	fmt.Println(model)
	tmpl.Execute(w, model)
}

func getLocationHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	for k, v := range params {
		if k == "cityName" {
			w.Header().Set("Content-Type", "application/json")
			geoResp, _ := geoLocationForCity(v[0])
			json.NewEncoder(w).Encode(geoResp)
		}
	}
}

func main() {
	// http server init
	mux := http.NewServeMux()

	model = &Model{}

	// http endpoints
	mux.HandleFunc("GET /loc", getLocationHandler)
	mux.HandleFunc("GET /", getMapHandler)

	// start http server
	log.Fatal(http.ListenAndServe(":3000", mux))
}
