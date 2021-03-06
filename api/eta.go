package api

import (   //blank import import those below; we can use factored import declaration import two more things
	_ "encoding/json"
	"fmt"
	"github.com/wtg/shuttletracker"
	"github.com/wtg/shuttletracker/log"
	"math"
	"net/http"
	_ "reflect"
	_ "strconv"
)

// etas at /eta

func (api *API) EtaHandler(w http.ResponseWriter, r *http.Request) {
	//get routes and stops info from DB
	routes, err := api.ms.Routes()
	if err != nil {
		log.WithError(err).Error("unable to get eta")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stops, err := api.ms.Stops()
	if err != nil {
		log.WithError(err).Error("unable to get stops")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Since we have not data feed offline, use random data to test
	tmpcar :=shuttletracker.Point{42.73166,-73.68559} //A random car location in west campus
	tmproute:=routes[0]                                // use only west campus route


	StopIndices :=GetStopsIndices(routes[2].Points,stops)

	index:=FindAffinity(tmproute.Points, tmpcar)

	distances:= make(map[string]float64)

	// get the distance from the car to all the stops
	for k, v := range StopIndices {
		fmt.Println(v)
		distances[k]=GetDistance(tmproute.Points,index,v)
	}

	// get the distance from Blitman to union and back
	distances["BlitmanToUnion"]=GetDistance(tmproute.Points,18,276)
	distances["UnionToBlitman"]=GetDistance(tmproute.Points,276,18)

	//get the total distance of all the routes
	for i:=0; i<len(routes);i++{
		points:=routes[i].Points
		str:=routes[i].Name+"'s total distances"
		distances[str]=GetDistance(points,0,len(points)-1)
	}

	WriteJSON(w, distances)
}

//calculate the distance between two points on the route

func GetDistance(points []shuttletracker.Point,start int,end int) float64{
	if end==start{
		return 0.0
	}
	if end<start{
		return GetDistance(points,start,len(points)-1)+GetDistance(points,0,end)
	}
	var dis float64=0
	var lat,lon float64
	for j:=start; j<=end;j++{
		if j==start{
			lat=points[j].Latitude/360*2*math.Pi
			lon= points[j].Longitude / 360 * 2 * math.Pi
		}else{
			lat1:=points[j].Latitude/360*2*math.Pi
			lon1:=points[j].Longitude/360*2*math.Pi
			dlon:=lon1-lon
			dlat:=lat1-lat
			a:=math.Pow(math.Sin(dlat/2),2) + math.Cos(lat1)*math.Cos(lat)*math.Pow(math.Sin(dlon/2),2)
			c := 2*math.Asin(math.Sqrt(a))
			dis+=c*6371/1.6
			lat=lat1
			lon=lon1
		}
	}
	return dis
}

// find the index of the closest point of p on the route points

func FindAffinity(points []shuttletracker.Point, p shuttletracker.Point) int{
	diff:=math.Pow((points[0].Latitude-p.Latitude),2)+math.Pow((points[0].Longitude-p.Longitude),2)
	min:=diff
	minIndex:=0
	for i:=1;i<len(points);i++{
		diff=math.Pow((points[i].Latitude-p.Latitude),2)+math.Pow((points[i].Longitude-p.Longitude),2)
		if diff<min{
			min=diff
			minIndex=i
		}
	}
	return minIndex
}

// get the index of closest point of each stop in stops
func GetStopsIndices(points []shuttletracker.Point, stops []*shuttletracker.Stop) map[string]int{
	indices:=make(map[string]int)
	for i:=0; i<len(stops);i++{
		indices[*stops[i].Name]=FindAffinity(points,shuttletracker.Point{stops[i].Latitude,stops[i].Longitude})
	}
	return indices
}