package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
)

type GeoResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

func geoLocationForCity(cityName string) (*GeoResponse, error) {
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

func main() {
	fmt.Println(geoLocationForCity("Bernhausen"))

}
