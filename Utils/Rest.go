package Utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"text/template"

	"ommaScraper/Data"
)

type OmmaRest struct {
	KeyUrl string
	AllUrl string
}

func (o *OmmaRest) Fetch(id string) (*Data.OmmaResponse, error) {
	wr := &bytes.Buffer{}
	templ := template.Must(template.New("licenseNo").Parse(o.KeyUrl))
	err := templ.ExecuteTemplate(wr, "licenseNo", map[string]interface{}{"LicenseNo": id})
	if err != nil {
		log.Printf("failure executing template: %s\n", err.Error())
		return nil, err
	}

	resp, err := http.Get(wr.String())
	if err != nil {
		log.Printf("OmmaRest Fetch Error: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("OmmaRest Fetch Error: %s", err.Error())
		return nil, err
	}

	data := &Data.OmmaResponse{}
	err = json.Unmarshal(body, data)
	return data, nil
}

func (o *OmmaRest) FetchAll() (*Data.OmmaResponse, error) {
	resp, err := http.Get(o.AllUrl)
	if err != nil {
		log.Printf("OmmaRest FetchAll Error: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("OmmaRest FetchAll Error: %s", err.Error())
		return nil, err
	}

	data := &Data.OmmaResponse{}
	err = json.Unmarshal(body, data)
	return data, nil
}
