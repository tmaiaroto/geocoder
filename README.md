Geocoder for Google Go
=====================

Originally authored by: github.com/jasonwinn/geocoder    
Updates have been made to this forked package and may be incompatible with the original package.

## What does it do

* Returns a Longitude and Latitude for a given string query
* Returns an address for a Longitude and Longitude

## API Key
Get a free API Key at [http://mapquestapi.com](http://mapquestapi.com)

## Why MapQuest API?
Google Maps Geocoding API has a limitation that prohibits querying their geocoding API unless you will be displaying the results on a Google Map.

## Install

* go get "github.com/tmaiaroto/geocoder"
* import "github.com/tmaiaroto/geocoder"
 

## Examples

### Geocode
```
  query := "Seattle WA"
  lat, lng := geocoder.Geocode(query)
  
  // 47.6064, -122.330803
 
```


### Reverse Geocode
```
  address := geocoder.ReverseGeocode(47.6064, -122.330803)

  address.Street 	        // 542 Marion St   
  address.City 		        // Seattle
  address.State 	        // WA
  address.PostalCode 	    // 98104 
  address.County 	        // King
  address.CountryCode       // US 

```