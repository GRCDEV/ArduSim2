package kml

import (
	"encoding/xml"
	"os"
	"strconv"
	"strings"
	"ui/internal/domain"
)

type kmlRoot struct {
	XMLName  xml.Name    `xml:"kml"`
	Document kmlDocument `xml:"Document"`
}

type kmlDocument struct {
	Placemark kmlPlacemark `xml:"Placemark"`
}

type kmlPlacemark struct {
	LineString kmlLineString `xml:"LineString"`
}

type kmlLineString struct {
	Coordinates string `xml:"coordinates"`
}

// GetFirstCoordinate returns the first Latitude, Longitude, and Altitude from a KML file.
func GetFirstCoordinate(filePath string) (*domain.Coordinate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var kml kmlRoot
	if err := decoder.Decode(&kml); err != nil {
		return nil, err
	}

	coordText := strings.TrimSpace(kml.Document.Placemark.LineString.Coordinates)
	if coordText == "" {
		return nil, nil
	}

	// KML coordinates are typically: lon,lat,alt
	points := strings.Fields(coordText)
	if len(points) == 0 {
		return nil, nil
	}

	parts := strings.Split(points[0], ",")
	if len(parts) < 2 {
		return nil, nil
	}

	lon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	alt := 0.0
	if len(parts) >= 3 {
		alt, _ = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	}

	if err1 != nil || err2 != nil {
		return nil, nil
	}

	return &domain.Coordinate{Lat: lat, Lon: lon, Alt: alt}, nil
}
