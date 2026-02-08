package utils

import (
	"github.com/google/uuid"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"igaku/med-service/errors"
	"igaku/med-service/models"
)

const RxClassDomainURL = "https://rxnav.nlm.nih.gov/REST"

type RxClassAPI interface {
	GetSubstances(diseaseID string) ([]models.Substance, error)
	GetDrugsByName(name string) ([]models.Drug, error)
}

type rxClassAPI struct {
	URL string
}

func NewRxClassAPI() RxClassAPI {
	return &rxClassAPI{URL: RxClassDomainURL}
}

func (api *rxClassAPI) fetchFromEndpoint(endpoint string) ([]byte, error) {
	res, err := http.Get(api.URL + endpoint)
	if err != nil {
		log.Printf("Failed to fetch data from endpoint: %v", err)
		return nil, &errors.RxClassUnavailableError{}
	}
	defer res.Body.Close()

	return io.ReadAll(res.Body)
}

func (api *rxClassAPI) GetSubstances(diseaseID string) ([]models.Substance, error) {
	endpoint := fmt.Sprintf(
		"/rxclass/classMembers.json?classId=%s&relaSource=MEDRT&rela=may_treat&trans=1",
		diseaseID,
	)
	data, err := api.fetchFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var rx RxClass
	err = json.Unmarshal(data, &rx)
	if err != nil {
		return nil, &errors.SubstanceNotFoundError{}
	}

	substances := make([]models.Substance, 0, len(rx.DrugMemberGroup.DrugMembers))

	for _, dm := range rx.DrugMemberGroup.DrugMembers {
		substances = append(substances, models.Substance{
			ID:    uuid.New(),
			RXCUI: dm.MinConcept.Rxcui,
			Name:  dm.MinConcept.Name,
			TTY:   dm.MinConcept.Tty,
		})
	}

	return substances, nil
}

func (api *rxClassAPI) GetDrugsByName(name string) ([]models.Drug, error) {
	endpoint := fmt.Sprintf("/drugs.json?name=%s", name)
	data, err := api.fetchFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var resp DrugGroupResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, &errors.DrugNotFoundError{}
	}

	total := 0
	for _, cg := range resp.DrugGroup.ConceptGroup {
		total += len(cg.ConceptProperties)
	}
	if total == 0 {
		return nil, &errors.DrugNotFoundError{}
	}

	drugs := make([]models.Drug, 0, total)

	for _, cg := range resp.DrugGroup.ConceptGroup {
		for _, cp := range cg.ConceptProperties {
			drugs = append(drugs, models.Drug{
				ID:        uuid.New(),
				RXCUI:     cp.Rxcui,
				Name:      cp.Name,
				Substance: name,
			})
		}
	}

	return drugs, nil
}
