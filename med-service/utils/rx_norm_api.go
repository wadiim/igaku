package utils

import (
	"gorm.io/gorm"
	"github.com/google/uuid"

	"encoding/xml"
	"io"
	"log"
	"net/http"

	models "igaku/commons/models"
	errors "igaku/med-service/errors"
)

type RxNormAPI struct {
	URL string
}

func (api *RxNormAPI) fetchDiseaseData() ([]byte, error) {
	res, err := http.Get(api.URL)
	if err != nil {
		log.Printf("Failed to fetch disease data: %v", err)
		return nil, &errors.RxNormUnavailableError{}
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func (api *RxNormAPI) transformDiseaseData(diseaseData []byte) []models.Disease {
	type RxClassMinConcept struct {
		ClassId string `xml:"classId"`
		ClassName string `xml:"className"`
		ClassType string `xml:"classType"`
	}

	type RxClassMinConceptList struct {
		XMLName xml.Name `xml:"rxclassMinConceptList"`
		Concepts []RxClassMinConcept `xml:"rxclassMinConcept"`
	}

	type RxClassData struct {
		XMLName xml.Name `xml:"rxclassdata"`
		ConceptList RxClassMinConceptList `xml:"rxclassMinConceptList"`
	}

	var classData RxClassData
	xml.Unmarshal(diseaseData, &classData)

	diseases := make([]models.Disease, len(classData.ConceptList.Concepts))

	for idx, concept := range classData.ConceptList.Concepts {
		diseases[idx] = models.Disease{
			ID: uuid.New(),
			RxNormID: concept.ClassId,
			Name: concept.ClassName,
		}
	}

	return diseases
}

func (api *RxNormAPI) GetAllDiseases(db *gorm.DB) ([]models.Disease){
	diseaseData, err := api.fetchDiseaseData()
	if err != nil {
		log.Printf("%v", err)
	}

	return api.transformDiseaseData(diseaseData)
}
