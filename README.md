# OMMA Scraper
The Oklahoma Medical Marijuana Authority (omma) provides public access to their business license database. This queries their database and inserts the records into a MongoDB database. Each record has its expiration data converted from a string to an ISODate and includes an "AsOfDate" to denote the date when the record was added to the database.

## Usage
```
Usage: OmmaScraper [--mongoConfig <PATH> (--all || --group <GROUP>) --init]

Options:
    --mongoConfig <PATH>        Path to configuration JSON file for MongoDB
    --all                       Process all available license types
    --group <GROUP>             Process a single license type
                                valid types: [grower, grower-indoor, grower-outdoor, dispensary, education, processor, laboratory, waste]
    --init                      Initialize the database and process all the license types
```

## Configuration
### mongo.json
Used to provide the connection details for the target MongoDB instance
```json
{
  "hostname": "localhost",
  "port": 27017,
  "database": "",
  "collection": ""
}
```

## Data Model
```go
// OMMA License
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
```

```go
// OMMA REST Response
ErrorCode   int32         `json:"errorCode"`
ErrorMsg    string        `json:"errorMessage"`
Method      string        `json:"method"`
ResultCount int32         `json:"resultCount"`
Result      []OmmaLicense `json:"result"`
```