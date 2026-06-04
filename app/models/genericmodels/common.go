package genericmodels

// Address represents the structure for storing address details.
type Address struct {
	Line       []string `bson:"line,omitempty" json:"line,omitempty" validate:"required,addressLines" field:"line"`             // Address line details (e.g., street or building information).
	City       string   `bson:"city,omitempty" json:"city,omitempty" validate:"required,max=100" field:"city"`                  // City name with a maximum length of 100 characters.
	District   string   `bson:"district,omitempty" json:"district,omitempty" validate:"omitempty,max=100" field:"district"`     // Optional district name with a maximum length of 100 characters.
	State      string   `bson:"state,omitempty" json:"state,omitempty" validate:"required,max=100" field:"state"`               // State name with a maximum length of 100 characters.
	PostalCode string   `bson:"postalCode,omitempty" json:"postalCode,omitempty" validate:"required,max=30" field:"postalCode"` // Postal code with a maximum length of 30 characters.
	Country    string   `bson:"country,omitempty" json:"country,omitempty" validate:"required,max=100" field:"country"`         // Country name with a maximum length of 100 characters.
}

// Address represents the structure for storing address details.
type UpdateAddress struct {
	Line       []string `bson:"line" validate:"omitempty,addressLines" field:"line"`
	City       string   `bson:"city" validate:"omitempty,max=100" field:"city"`
	District   string   `bson:"district" validate:"omitempty,max=100" field:"district"`
	State      string   `bson:"state" validate:"omitempty,max=100" field:"state"`
	PostalCode string   `bson:"postalCode" validate:"omitempty,max=30" field:"postalCode"`
	Country    string   `bson:"country" validate:"omitempty,max=100" field:"country"`
}

// Location represents the geographical coordinates of a place.
type Location struct {
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180" field:"longitude"` // Longitude value, must be between -180 and 180 degrees.
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90" field:"latitude"`     // Latitude value, must be between -90 and 90 degrees.
}

type GeoLocation struct {
	Type        string     `json:"type"`        // The type of GeoJSON object, e.g., "Point"
	Coordinates [2]float64 `json:"coordinates"` // An array of [longitude, latitude]
}
