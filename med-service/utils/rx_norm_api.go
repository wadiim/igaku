package utils

import (
	"gorm.io/gorm"
	"github.com/google/uuid"

	"encoding/json"
	"io"
	"log"
	"net/http"

	"igaku/med-service/errors"
	"igaku/med-service/models"
)

const RxNormDomainURL = "https://rxnav.nlm.nih.gov/REST"

type RxNormAPI interface {
	GetAllDiseases(db *gorm.DB) ([]models.Disease, error)
}

type rxNormAPI struct {
	URL string
}

func NewRxNormAPI() RxNormAPI {
	return &rxNormAPI{URL: RxNormDomainURL}
}

func (api *rxNormAPI) fetchFromEndpoint(endpoint string) ([]byte, error) {
	res, err := http.Get(api.URL + endpoint)
	if err != nil {
		log.Printf("Failed to fetch data from endpoint: %v", err)
		return nil, &errors.RxNormUnavailableError{}
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func (api *rxNormAPI) GetAllDiseases(db *gorm.DB) ([]models.Disease, error) {
	endpoint := "/rxclass/allClasses.json?classTypes=DISEASE"
	data, err := api.fetchFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var classData RxClassData
	err = json.Unmarshal(data, &classData)
	if err != nil {
		return nil, err
	}

	diseases := make([]models.Disease, len(classData.ConceptList.Concepts))

	for idx, concept := range classData.ConceptList.Concepts {
		diseases[idx] = models.Disease{
			ID: uuid.New(),
			RxNormID: concept.ClassId,
			Name: concept.ClassName,
		}
	}

	return diseases, nil
}
