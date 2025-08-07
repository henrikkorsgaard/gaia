package dmi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spatial-go/geoos/geoencoding/geojson"
)

type Config struct {
	LIGHTNING_KEY string `env:"LIGHTNING_KEY,required"`
	CLIMATE_KEY   string `env:"CLIMATE_KEY,required"`
	METOBS_KEY    string `env:"METOBS_KEY,required"`
}

type DMIService struct {
	Config
}

func New(cfg Config) *DMIService {

	return &DMIService{
		Config: cfg,
	}
}

func (s *DMIService) GetLightningData() {
	//https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=
	//Fetch data -> we also need models

	res, err := http.Get("https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=" + s.LIGHTNING_KEY)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	var fc geojson.FeatureCollection
	err = json.Unmarshal(body, &fc)
	if err != nil {
		fmt.Println(err)
	}

	for _, f := range fc.Features {
		fmt.Printf("%+v", f)
	}
}

func (s *DMIService) GetClimateData() {
	//https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=
	//Fetch data -> we also need models

	res, err := http.Get("https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=" + s.LIGHTNING_KEY)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	var fc geojson.FeatureCollection
	err = json.Unmarshal(body, &fc)
	if err != nil {
		fmt.Println(err)
	}

	for _, f := range fc.Features {
		fmt.Printf("%+v", f)
	}
}

func (s *DMIService) GetMetObsData() {
	//https://dmigw.govcloud.dk/v2/metObs/collections/observation/items?api-key=0bd05afc-b284-470d-8a75-f46864fd347c
	//Fetch data -> we also need models

	res, err := http.Get("https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=" + s.LIGHTNING_KEY)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	var fc geojson.FeatureCollection
	err = json.Unmarshal(body, &fc)
	if err != nil {
		fmt.Println(err)
	}

	for _, f := range fc.Features {
		fmt.Printf("%+v", f)
	}
}
