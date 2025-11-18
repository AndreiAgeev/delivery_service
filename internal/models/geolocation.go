package models

const (
	YandexGeocoderURL      string = "https://geocode-maps.yandex.ru/v1/"
	OpenrouteDirectionsURL string = "https://api.openrouteservice.org/v2/directions/foot-walking"
)

type YandexResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					Point struct {
						Pos string `json:"pos"`
					} `json:"Point"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

type OpenrouteRequest struct {
	Coordinates  [][2]float64 `json:"coordinates"`
	Instructions string       `json:"instructions"`
	Maneuvers    string       `json:"maneuvers"`
}

type OpenrouteResponse struct {
	Routes struct {
		Summary struct {
			Distance float64 `json:"distance"`
		} `json:"summary"`
	} `json:"routes"`
}

type GeoCache struct {
	PickupCoordinates   [2]float64 `json:"pickup_coordinates"`
	DeliveryCoordinates [2]float64 `json:"delivery_coordinates"`
	Distance            float64    `json:"distance"`
	DeliveryCost        float64    `json:"delivery_cost"`
}
