package main

import (
	"flag"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"ommaScraper/Data"
	"ommaScraper/Utils"
	"slices"
	"strings"
	"sync"
	"time"
)

func main() {
	doAll := flag.Bool("all", false, "Scrape all data")
	init := flag.Bool("init", false, "Initialize database")
	group := flag.String("group", "", "Scrape group data")
	mongoConfigPath := flag.String("mongoConfig", "Config/mongo.json", "Mongodb connection JSON file")
	flag.Parse()

	if *group != "" {
		typeList := []string{
			"grower",
			"grower-indoor",
			"grower-outdoor",
			"dispensary",
			"education",
			"processor",
			"laboratory",
			"waste",
		}

		if !slices.Contains(typeList, strings.ToLower(*group)) {
			log.Fatalf("Invalid group '%s'. Valid options: %v", *group, typeList)
		}

		*group = strings.ToLower(*group)
	}

	log.Printf("Loading MongoDB configuration: %s", *mongoConfigPath)
	mongoCfg, err := Utils.LoadMongoConfig(*mongoConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up new mongo client")
	mongo, err := Utils.NewMongoClient(fmt.Sprintf("mongodb://%s:%d", mongoCfg.Hostname, mongoCfg.Port))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connecting to mongodb")
	err = mongo.Connect(mongoCfg.Database, mongoCfg.Collection)
	if err != nil {
		log.Fatal(err)
	}
	defer func(mongo *Utils.MongoClient) {
		err = mongo.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(mongo)

	if *init {
		if err = Utils.InitDatabase(mongo, true); err != nil {
			log.Fatalf("Error initializing database: %s", err.Error())
		}
		return
	}

	today := time.Now()
	log.Printf("Today is %s", today.Format(time.DateOnly))

	var ltm []Data.OmmaLicenseType
	if *doAll {
		// exclude the indoor and outdoor growers as they are returned by the basic grower query
		ltm = Data.LicenseTypeMapKeys([]Data.OmmaLicenseType{}) //Data.GROWER_INDOOR, Data.GROWER_OUTDOOR})
	} else {
		ltm = append(ltm, groupToOmmaLicenseType(*group))
	}

	var wg sync.WaitGroup
	for i := 0; i < len(ltm); i++ {
		wg.Add(1)
		log.Printf("Creating thread %d for %s", i, ltm[i].String())
		go doWork(i, &wg, ltm[i], mongo)
	}

	wg.Wait()

	endTime := time.Now()
	log.Printf("Completed in %s", endTime.Sub(today))
}

func doWork(threadId int, wg *sync.WaitGroup, group Data.OmmaLicenseType, dbh *Utils.MongoClient) {
	defer wg.Done()

	StartTime := time.Now()
	log.Printf("[%2d] Processing group '%s'", threadId, group.String())
	data, err := Utils.QueryOmmaByLicenseType(group)
	if err != nil {
		log.Printf("[%2d] Error processing group '%s': %s", threadId, group.String(), err.Error())
		return
	}
	log.Printf("[%2d] Got %d records in %s", threadId, len(data), time.Since(StartTime))

	log.Printf("[%2d] Updating database %d records for %s", threadId, len(data), group.String())
	for i, record := range data {
		record.AsOfDate = time.Now()
		record.Expiration, err = time.Parse("2006-01-02", record.LicenseExpiryDate)
		data[i] = record
	}

	updateErrors := dbh.UpdateMany(bson.M{"licenseType": group.String()}, data)
	if len(updateErrors) > 0 {
		var sErrors string
		for _, e := range updateErrors {
			sErrors += e.Error() + "\n"
		}
		log.Printf("[%2d] Error updating database records: %s", threadId, sErrors)
	}

	log.Printf("[%2d] Completed processing group '%s' in %s", threadId, group, time.Since(StartTime))
}

func groupToOmmaLicenseType(group string) Data.OmmaLicenseType {
	switch group {
	case "dispensary":
		return Data.DISPENSARY
	case "education":
		return Data.EDUCATIONAL
	case "grower":
		return Data.GROWER
	case "grower-indoor":
		return Data.GROWER_INDOOR
	case "grower-outdoor":
		return Data.GROWER_INDOOR
	case "processor":
		return Data.PROCESSOR
	case "laboratory":
		return Data.LABORATORY
	case "waste":
		return Data.WASTE_DISPOSAL
	default:
		return -1
	}
}
