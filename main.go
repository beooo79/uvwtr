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

type GeoLocation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func geoLocationForCity(cityName string) (*GeoLocation, error) {
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

func getMapHandler(w http.ResponseWriter, req *http.Request) {
	tmpl := template.Must(template.ParseFS(templateHTML, "templates/map.html"))
	params := req.URL.Query()
	for k, v := range params {
		if k == "cityName" {
			geoResp, err := geoLocationForCity(v[0])
			if err == nil {
				model.CityName = geoResp.Name
				model.Latitude = geoResp.Latitude
				model.Longitude = geoResp.Longitude
			}
		}
	}
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
