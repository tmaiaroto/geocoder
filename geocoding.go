/* Exposes (partially) the mapquest geocoding api.

Reference: http://open.mapquestapi.com/geocoding/

Example:

lat, lng := Geocode("Seattle WA")
address := ReverseGeocode(47.603561, -122.329437)

*/

package geocoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tmaiaroto/geocoder/lib/httpclient"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	geocodeURL        = "http://www.mapquestapi.com/geocoding/v1/address?inFormat=kvp&outFormat=json&location="
	reverseGeocodeURL = "http://www.mapquestapi.com/geocoding/v1/reverse?location="
	batchGeocodeURL   = "http://open.mapquestapi.com/geocoding/v1/batch?key="
)

var HttpClient http.Client

func NewGeocoder() *http.Client {
	// TODO: Maybe make the timeout something configurable

	HttpClient := &http.Client{
		Transport: &httptimeout.TimeoutTransport{
			Transport: http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					//log.Printf("dial to %s://%s", netw, addr)
					return net.Dial(netw, addr) // Regular ass dial.
				},
			},
			// RoundTripTimeout: time.Millisecond * 200, // <--- what the author had
			// RoundTripTimeout: time.Nanosecond * 10, // <--- still was completing requests in this amount of time (holy smokes that's fast)!
			// I'm going to go a little tiny bit longer because I don't know what kind of machine this will run on.
			// Though the geocoding service is fast and the payload small, so requests should be fast.
			RoundTripTimeout: time.Millisecond * 300,
		},
	}

	//HttpClient = http.Client{}

	return HttpClient
}

// Geocode returns the latitude and longitude for a certain address
func Geocode(address string) (lat float64, lng float64) {
	// Query Provider
	//resp, err := http.Get(geocodeURL + url.QueryEscape(address) + "&key=" + apiKey)

	var buffer bytes.Buffer
	buffer.WriteString(geocodeURL)
	buffer.WriteString(url.QueryEscape(address))
	buffer.WriteString("&key=")
	buffer.WriteString(apiKey)
	// resp, err := HttpClient.Get(buffer.String())
	// with timeout
	req, err := http.NewRequest("GET", buffer.String(), nil)
	buffer.Reset()
	if err != nil {
		return 0.0, 0.0
	}
	req.Header.Add("Connection", "keep-alive")
	resp, err := HttpClient.Do(req)

	if err != nil {
		//panic(err)
		return 0.0, 0.0
	}

	defer resp.Body.Close()

	// Decode our JSON results
	var result geocodingResults
	err = decoder(resp).Decode(&result)

	if err != nil {
		//panic(err)
		return 0.0, 0.0
	}

	if len(result.Results[0].Locations) > 0 {
		lat = result.Results[0].Locations[0].LatLng.Lat
		lng = result.Results[0].Locations[0].LatLng.Lng
	}

	return lat, lng
}

func GeocodeLocation(address string) (Location, error) {
	loc := Location{}

	// Query Provider
	// buffer.WriteString() is a lot faster than concatenating with +
	var buffer bytes.Buffer
	buffer.WriteString(geocodeURL)
	buffer.WriteString(url.QueryEscape(address))
	buffer.WriteString("&key=")
	buffer.WriteString(apiKey)

	//resp, err := http.Get(buffer.String())
	//resp, err := HttpClient.Get(buffer.String())
	// even better?
	req, err := http.NewRequest("GET", buffer.String(), nil)
	buffer.Reset()
	if err != nil {
		return loc, err
	}
	req.Header.Add("Connection", "keep-alive")
	resp, err := HttpClient.Do(req)

	//resp, err := http.Get(geocodeURL + url.QueryEscape(address) + "&key=" + apiKey)
	if err != nil {
		return loc, err
	}
	defer resp.Body.Close()

	// Decode our JSON results
	var result geocodingResults
	err = decoder(resp).Decode(&result)
	if err != nil {
		return loc, err
	}

	if len(result.Results[0].Locations) > 0 {
		loc = result.Results[0].Locations[0]
	}

	return loc, err
}

// ReverseGeocode returns the address for a certain latitude and longitude
func ReverseGeocode(lat float64, lng float64) (Location, error) {
	var location Location

	// Query Provider
	var buffer bytes.Buffer
	buffer.WriteString(reverseGeocodeURL)
	buffer.WriteString(fmt.Sprintf("%f,%f&key=%s", lat, lng, apiKey))
	//resp, err := http.Get(buffer.String())
	//resp, err := HttpClient.Get(buffer.String())
	// with timeout
	req, err := http.NewRequest("GET", buffer.String(), nil)
	buffer.Reset()
	if err != nil {
		return location, err
	}
	req.Header.Add("Connection", "keep-alive")
	resp, err := HttpClient.Do(req)

	//resp, err := http.Get(reverseGeocodeURL + fmt.Sprintf("%f,%f&key=%s", lat, lng, apiKey))
	if err != nil {
		//panic(err)
		//log.Println(err)
		return location, err
	}
	defer resp.Body.Close()

	// Decode our JSON results
	var result geocodingResults
	err = decoder(resp).Decode(&result)
	if err != nil {
		//panic(err)
		log.Println(err)
	}

	// Assign the results to the Location struct
	if len(result.Results[0].Locations) > 0 {
		location = result.Results[0].Locations[0]
	}

	return location, err
}

// BatchGeocode allows multiple locations to be geocoded at the same time.
// A limit of 100 locations exists for one call. Therefore the json is
// embedded as the body of an http post.
func BatchGeocode(addresses []string) (latLngs []LatLng) {
	var next, start, end int
	n := len(addresses)
	latLngs = make([]LatLng, n)
	batches := n/100 + 1
	next = 0
	for batch := 0; batch < batches; batch++ {
		start = next
		next = (batch + 1) * 100
		if n < next {
			end = n
		} else {
			end = next
		}
		bgb := batchGeocodeBody{
			Locations:  addresses[start:end],
			MaxResults: 1,
			ThumbMaps:  false,
		}
		b, err := json.Marshal(bgb)
		if err != nil {
			panic(err)
		}
		body := bytes.NewBuffer(b)
		resp, err := http.Post(batchGeocodeURL+apiKey, "application/json", body)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		var result geocodingResults
		err = decoder(resp).Decode(&result)
		if err != nil {
			panic(err)
		}
		for i, r := range result.Results {
			if len(r.Locations) == 0 {
				latLngs[start+i] = LatLng{Lat: 0, Lng: 0}
			} else {
				latLngs[start+i] = r.Locations[0].LatLng
			}
		}
	}
	return
}

// geocodingResults contains the locations of a geocoding request
// MapQuest providers more JSON fields than this but this is all we are interested in.
type geocodingResults struct {
	Results []struct {
		Locations []Location `json:"locations"`
	} `json:"results"`
}

// batchGeocodeBody will be marshalled as json to send in body with http post
type batchGeocodeBody struct {
	Locations  []string `json:"locations"`
	MaxResults int
	ThumbMaps  bool `json:"thumbMaps"`
}
