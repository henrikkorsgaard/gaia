package dmi

type Config struct {
	API_KEY string `env:"API_KEY,required"`
}

type DMIService struct {
	API_KEY string
}

func New(cfg Config) DMIService {

	return DMIService{
		API_KEY: cfg.API_KEY,
	}
}

func (s *DMIService) GetData() {
	//Fetch data -> we also need models

}
