package parser

import (
	"encoding/xml"
	"os"
	"strconv"
	"strings"

	"mission/domain"
)

type KmlParser struct{}

func NewKmlParser() *KmlParser {
	return &KmlParser{}
}

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

type Document struct {
	Placemark Placemark `xml:"Placemark"`
}

type Placemark struct {
	LineString LineString `xml:"LineString"`
}

type LineString struct {
	Coordinates string `xml:"coordinates"`
}

func (k *KmlParser) ParseMission(filePath string) ([]domain.Waypoint, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var kml KML
	if err := decoder.Decode(&kml); err != nil {
		return nil, err
	}

	var waypoints []domain.Waypoint
	coordText := strings.TrimSpace(kml.Document.Placemark.LineString.Coordinates)
	if coordText == "" {
		return waypoints, nil
	}

	points := strings.Fields(coordText)
	for _, pointStr := range points {
		parts := strings.Split(pointStr, ",")
		if len(parts) >= 2 {
			lon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			lat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			alt := 0.0
			if len(parts) >= 3 {
				alt, _ = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			}

			if err1 == nil && err2 == nil {
				waypoints = append(waypoints, domain.Waypoint{
					Longitude: lon,
					Latitude:  lat,
					Altitude:  alt,
				})
			}
		}
	}

	return waypoints, nil
}
