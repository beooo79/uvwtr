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

func geoLocationForCity(cityName string) (*GeoResponse, error) {
	fmt.Println("+++loc for CITY+++", cityName)

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
		return &geoResp, nil
	} else {
		fmt.Println("No results found")
		return nil, errors.New("geolocation for city " + cityName + " not found")
	}
}

func getMapHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFS(templateHTML, "templates/map.html"))
	tmpl.Execute(w, model)
}

func getLocationHandler(w http.ResponseWriter, req *http.Request) {
	params := req.URL.Query()
	var cityName string
	for k, v := range params {
		if k == "cityName" {
			cityName = v[0]
		}

		w.Header().Set("Content-Type", "application/json")
		geoResp, _ := geoLocationForCity(cityName)
		json.NewEncoder(w).Encode(geoResp)
	}
}

func main() {
	fmt.Println(geoLocationForCity("Bernhausen"))

	// http server init
	mux := http.NewServeMux()

	// http endpoints
	mux.HandleFunc("GET /loc", getLocationHandler)
	mux.HandleFunc("GET /uv", getMapHandler)

	// start http server
	log.Fatal(http.ListenAndServe(":3000", mux))
}
