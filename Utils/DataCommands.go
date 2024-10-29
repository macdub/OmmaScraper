package Utils

import (
	"context"
	"fmt"
	"log"
	"ommaScraper/Data"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func InitDatabase(m *MongoClient, destroy bool) error {
	log.Println("Collecting data from OMMA")
	data, err := getAllFromOmma()
	if err != nil {
		return err
	}

	log.Printf("Checking if collection (%s) exists ... ", m.Collection.Name())
	count, err := m.Collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	if count > 0 && destroy {
		log.Println("Destroy set. Dropping collection.")
		err = m.Collection.Drop(context.Background())
		if err != nil {
			return err
		}
	}

	if count > 0 && !destroy {
		return fmt.Errorf("collection is already populated")
	}

	log.Printf("Attempting to insert %d records\n", len(data))
	insertCount := 0
	for _, rec := range data {
		rec.AsOfDate = time.Now()
		rec.Expiration, err = time.Parse("2006-01-02", rec.LicenseExpiryDate)
		if err != nil {
			log.Printf("Error parsing LicenseExpiryDate %s", rec.LicenseExpiryDate)
		}

		_, err = m.Collection.InsertOne(context.Background(), rec)
		if err != nil {
			log.Printf("Error inserting record to collection: %s\n", err.Error())
		}

		insertCount++
	}

	log.Printf("Inserted %d / %d records\n", insertCount, len(data))
	return nil
}

func getAllFromOmma() ([]Data.OmmaLicense, error) {
	r := &OmmaRest{AllUrl: "https://omma.us.thentiacloud.net/rest/public/profile/search/?keyword=%2A&skip=0&take=20&lang=en&type=all"}
	data, err := r.FetchAll()
	if err != nil {
		return nil, err
	}

	return data.Result, nil
}

func QueryOmmaByLicenseNumber(id string) (*Data.OmmaLicense, error) {
	r := &OmmaRest{KeyUrl: "https://omma.us.thentiacloud.net/rest/public/profile/search/?keyword={{ .LicenseNo }}&skip=0&take=20&lang=en&type=all"}

	if data, err := doOmmaQuery(r, id); err != nil {
		return nil, err
	} else {
		return &data.Result[0], nil
	}
}

func QueryOmmaByLicenseType(licenseType Data.OmmaLicenseType) ([]Data.OmmaLicense, error) {
	typeString := Data.LicenseTypeMap()[licenseType]
	r := &OmmaRest{
		KeyUrl: "https://omma.us.thentiacloud.net/rest/public/profile/search/?type={{ .LicenseNo }}",
	}

	data, err := doOmmaQuery(r, typeString)
	if err != nil {
		return nil, err
	}

	data.Result = filter(data.Result, func(s Data.OmmaLicense) bool {
		return s.LicenseType == licenseType.String()
	})

	return data.Result, nil
}

func filter[T any](sliceT []T, test func(T) bool) (ret []T) {
	for _, t := range sliceT {
		if test(t) {
			ret = append(ret, t)
		}
	}
	return
}

func doOmmaQuery(r *OmmaRest, query string) (*Data.OmmaResponse, error) {
	data, err := r.Fetch(query)
	if err != nil {
		return nil, err
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("omma fetch error: [%d] %s", data.ErrorCode, data.ErrorCode)
	}

	if data.ResultCount < 1 || len(data.Result) < 1 {
		return nil, fmt.Errorf("omma fetch error: no results for query '%s'\n", query)
	}

	return data, nil
}
