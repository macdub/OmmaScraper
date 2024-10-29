package Data

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"slices"
	"sync"
	"time"
)

type OmmaLicenseType int

const (
	DISPENSARY OmmaLicenseType = iota
	EDUCATIONAL
	GROWER
	GROWER_INDOOR
	GROWER_OUTDOOR
	PROCESSOR
	LABORATORY
	TRANSPORTER
	WASTE_DISPOSAL
)

func (o OmmaLicenseType) String() string {
	return [...]string{
		"Dispensary",
		"Education Facility",
		"Grower",
		"Grower Indoor",
		"Grower Outdoor",
		"Processor",
		"Testing Laboratory",
		"Transporter",
		"Waste Disposal Facility"}[o]
}

var licenseTypeMap_ map[OmmaLicenseType]string
var licenseTypeMap_once sync.Once

func LicenseTypeMap() map[OmmaLicenseType]string {
	licenseTypeMap_once.Do(func() {
		licenseTypeMap_ = map[OmmaLicenseType]string{
			DISPENSARY:     "Dispensary",
			EDUCATIONAL:    "Education%20Facility",
			GROWER:         "Grower",
			GROWER_INDOOR:  "Grower%20Indoor",
			GROWER_OUTDOOR: "Grower%20Outdoor",
			PROCESSOR:      "Processor",
			LABORATORY:     "Testing%20Laboratory",
			TRANSPORTER:    "Transporter",
			WASTE_DISPOSAL: "Waste%20Disposal%20Facility",
		}
	})

	return licenseTypeMap_
}

func LicenseTypeMapKeys(exclude []OmmaLicenseType) []OmmaLicenseType {
	var licenseTypes []OmmaLicenseType
	for key, _ := range LicenseTypeMap() {
		if slices.Contains(exclude, key) {
			continue
		}

		licenseTypes = append(licenseTypes, key)
	}

	return licenseTypes
}

type OmmaLicense struct {
	LicenseNumber     string    `bson:"licenseNumber"`
	LegalName         string    `bson:"legalName"`
	TradeName         string    `bson:"tradeName"`
	LicenseType       string    `bson:"licenseType"`
	StreetAddress     string    `bson:"streetAddress"`
	City              string    `bson:"city"`
	County            string    `bson:"county"`
	LicenseExpiryDate string    `bson:"licenseExpiryDate"`
	Zip               string    `bson:"zip"`
	Phone             string    `bson:"phone"`
	Email             string    `bson:"email"`
	Hours             string    `bson:"hours"`
	DataSourceName    string    `bson:"dataSourceName,omitempty"`
	DiscloseAddress   bool      `bson:"discloseAddress"`
	Expiration        time.Time `bson:"Expiration"`
	AsOfDate          time.Time `bson:"AsOfDate"`
}

type OmmaResponse struct {
	ErrorCode   int32         `json:"errorCode"`
	ErrorMsg    string        `json:"errorMessage"`
	Method      string        `json:"method"`
	ResultCount int32         `json:"resultCount"`
	Result      []OmmaLicense `json:"result"`
}

type OmmaTime struct {
	time.Time
}

func (t *OmmaTime) UnmarshalJSON(data []byte) error {
	date, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}

	t.Time = date
	return nil
}

func (t *OmmaTime) UnmarshalBSONValue(bt bsontype.Type, b []byte) error {
	rawValue := bson.RawValue{Type: bt, Value: b}

	var result time.Time
	if err := rawValue.Unmarshal(&result); err != nil {
		return err
	}

	t.Time = result
	return nil
}

func (t *OmmaTime) MarshalJSON() ([]byte, error) {
	data, err := t.Time.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (t *OmmaTime) MarshalBSON() ([]byte, error) {
	data, err := t.Time.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return data, nil
}
