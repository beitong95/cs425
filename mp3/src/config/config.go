package config
// copy from mp0 sample

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

// need to be the same as the json structure in config.json
type configParam struct {
	IntroducerIPAddresses []string
	Port                  string
}

// Opens the filename and attempts to deserialize it into a struct
// directly copy from mp0
func parseJSON(fileName string) (configParam, error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return configParam{}, err
	}

	//Necessities for go to be able to read JSON
	fileString := string(file)

	fileReader := strings.NewReader(fileString)

	decoder := json.NewDecoder(fileReader)

	var configParams configParam

	// Finally decode into json object
	err = decoder.Decode(&configParams)
	if err != nil {
		return configParam{}, err
	}

	return configParams, nil
}

// IntroducerIPAddresses gets the introducer ip address from config.json
func IntroducerIPAddresses() ([]string, error) {
	configParams, err := parseJSON(os.Getenv("CONFIG"))
	if err != nil {
		return make([]string, 0), err
	}
	return configParams.IntroducerIPAddresses, nil
}

// Port gets the port number from config.json
func Port() (string, error) {
	configParams, err := parseJSON(os.Getenv("CONFIG"))
	if err != nil {
		return "", err
	}
	return configParams.Port, nil
}
