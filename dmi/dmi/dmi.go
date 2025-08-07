package dmi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spatial-go/geoos/geoencoding/geojson"
)

type Config struct {
	API_KEY string `env:"API_KEY,required"`
}

type DMIService struct {
	API_KEY string
}

func New(cfg Config) *DMIService {

	return &DMIService{
		API_KEY: cfg.API_KEY,
	}
}

func (s *DMIService) GetLightningData() {
	//https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=
	//Fetch data -> we also need models

	res, err := http.Get("https://dmigw.govcloud.dk/v2/lightningdata/collections/observation/items?api-key=" + s.API_KEY)
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
