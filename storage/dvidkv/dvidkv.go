package dvidkv

import (
	"bytes"
	"fmt"
	"github.com/blang/semver"
	"github.com/connectome-neuprint/neuPrintHTTP/storage"
	"io/ioutil"
	"net/http"
	"time"
)

func init() {
	version, _ := semver.Make(VERSION)
	e := Engine{NAME, version}
	storage.RegisterEngine(e)
}

const (
	// VERSION of database that is supported
	VERSION = "0.1.0"
	NAME    = "dvidkv"
)

type Engine struct {
	name    string
	version semver.Version
}

func (e Engine) GetName() string {
	return e.name
}

type dvidConfig struct {
	Dataset  string `json:"dataset"`
	Server   string `json:"server"`
	Branch   string `json:"branch"`
	Instance string `json:"instance"`
}

// NewStore creates an store instance that works with dvid.
// DVID requires  data instance name, server, branch, and dataset
func (e Engine) NewStore(data interface{}, typename, instance string) (storage.SimpleStore, error) {
	dbversion, _ := semver.Make(VERSION)

	config, ok := data.(dvidConfig)
	if !ok {
		return nil, fmt.Errorf("incorrect configuration for neo4j")
	}

	endPoint := "http://" + config.Server + "/api/" + config.Branch + "/" + config.Instance + "/key/"
	return Store{dbversion, typename, instance, config, endPoint}, nil
}

// Store is the neo4j storage instance
type Store struct {
	version  semver.Version
	typename string
	instance string
	config   dvidConfig
	endPoint string
}

// GetDatabsae returns database information
func (store Store) GetDatabase() (loc string, desc string, err error) {
	return store.config.Server, NAME, nil
}

// GetVersion returns the version of the driver
func (store Store) GetVersion() (string, error) {
	return store.version.String(), nil
}

type databaseInfo struct {
	Branch   string `json:"branch"`
	Instance string `json:"instance"`
}

// GetDatasets returns information on the datasets supported
func (store Store) GetDatasets() (map[string]interface{}, error) {
	var datasetmap map[string]interface{}
	datasetmap[store.config.Dataset] = databaseInfo{store.config.Branch, store.config.Instance}

	return datasetmap, nil
}

func (store Store) GetInstance() string {
	return store.instance
}

func (store Store) GetType() string {
	return store.typename
}

// *** KeyValue Query Interfacde ****

// Set puts data into DVID
func (s Store) Set(key, val []byte) error {
	dvidClient := http.Client{
		Timeout: time.Second * 60,
	}

	req, err := http.NewRequest(http.MethodPost, s.endPoint, bytes.NewBuffer(val))
	if err != nil {
		return fmt.Errorf("request failed")
	}

	res, err := dvidClient.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// Get retrieve data from DVID
func (s Store) Get(key []byte) ([]byte, error) {
	dvidClient := http.Client{
		Timeout: time.Second * 60,
	}

	req, err := http.NewRequest(http.MethodGet, s.endPoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed")
	}
	res, err := dvidClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("request failed")
	}

	return body, err
}
