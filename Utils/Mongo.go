package Utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"ommaScraper/Data"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
	Uri        string
}

func NewMongoClient(uri string) (*MongoClient, error) {
	return &MongoClient{Uri: uri}, nil
}

func (m *MongoClient) Connect(db string, collection string) error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(m.Uri))
	if err != nil {
		log.Printf("error connecting to mongodb '%s': %s", m.Uri, err.Error())
		return err
	}

	m.Client = client
	m.useDatabase(db)
	err = m.useCollection(collection)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) Close() error {
	err := m.Client.Disconnect(context.TODO())
	if err != nil {
		log.Printf("error closing mongodb connection: %s", err.Error())
		return err
	}

	return nil
}

func (m *MongoClient) useDatabase(db string) {
	m.Database = m.Client.Database(db)
}

func (m *MongoClient) useCollection(collection string) error {
	if m.Database == nil {
		return fmt.Errorf("no database set")
	}
	m.Collection = m.Database.Collection(collection)

	return nil
}

func (m *MongoClient) FindOne(filter bson.D) ([]byte, error) {
	var result bson.M
	err := m.Collection.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("no document found for filter: %s", filter)
		return nil, nil
	}

	if err != nil {
		log.Printf("error finding document: %s", err.Error())
		return nil, err
	}

	jsonData, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Printf("error marshalling document: %s", err.Error())
		return nil, err
	}

	return jsonData, nil
}

func (m *MongoClient) Find(filter interface{}, projection interface{}) (*mongo.Cursor, error) {
	opts := options.Find().SetProjection(projection)
	cursor, err := m.Collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Printf("error finding document: %s\n", err.Error())
		return nil, err
	}

	return cursor, nil
}

func (m *MongoClient) UpdateOne(filter interface{}, update interface{}) error {
	opts := options.Update().SetUpsert(true)

	_, err := m.Collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) UpdateMany(filter interface{}, records []Data.OmmaLicense) []error {
	var updateErrors []error
	for _, record := range records {
		filter.(bson.M)["licenseNumber"] = record.LicenseNumber
		err := m.UpdateOne(filter, bson.M{"$set": record})
		if err != nil {
			updateErrors = append(updateErrors, err)
		}
	}

	return updateErrors
}

type MongoConfig struct {
	Hostname   string `json:"hostname"`
	Port       int    `json:"port"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

func LoadMongoConfig(filename string) (*MongoConfig, error) {
	cfg := &MongoConfig{}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if !json.Valid(data) {
		return nil, errors.New("invalid json")
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
