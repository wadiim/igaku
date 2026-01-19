package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	commonsModels "igaku/commons/models"
	"igaku/med-service/errors"
)

const RxClassDomainURL = "https://rxnav.nlm.nih.gov"

type RxClassAPI interface {
	GetSubstances(diseaseID string) ([]commonsModels.Substance, error)
	GetDrugsByName(name string) ([]commonsModels.Drug, error)
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

func (api *rxClassAPI) GetSubstances(diseaseID string) ([]commonsModels.Substance, error) {
	endpoint := fmt.Sprintf(
		"/REST/rxclass/classMembers.json?classId=%s&relaSource=MEDRT&rela=may_treat",
		diseaseID,
	)
	substancesData, err := api.fetchFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	type MinConcept struct {
		Rxcui string `json:"rxcui"`
		Name  string `json:"name"`
		Tty   string `json:"tty"`
	}

	type NodeAttr struct {
		AttrName  string `json:"attrName"`
		AttrValue string `json:"attrValue"`
	}

	type DrugMember struct {
		MinConcept MinConcept  `json:"minConcept"`
		NodeAttrs  []NodeAttr  `json:"nodeAttr"`
	}

	type DrugMemberGroup struct {
		DrugMembers []DrugMember `json:"drugMember"`
	}

	type RxClass struct {
		DrugMemberGroup DrugMemberGroup `json:"drugMemberGroup"`
	}

	var rx RxClass
	if err := json.Unmarshal(substancesData, &rx); err != nil {
		return nil, &errors.SubstanceNotFoundError{}
	}

	var substances []commonsModels.Substance
	for _, dm := range rx.DrugMemberGroup.DrugMembers {
		for _, a := range dm.NodeAttrs {
			if a.AttrName == "Relation" && a.AttrValue == "DIRECT" {
				substances = append(substances, commonsModels.Substance{
					ID:   dm.MinConcept.Rxcui,
					Name: dm.MinConcept.Name,
					Type: dm.MinConcept.Tty,
				})
				break
			}
		}
	}

	return substances, nil
}

func (api *rxClassAPI) GetDrugsByName(name string) ([]commonsModels.Drug, error) {
	type ConceptProperty struct {
		Rxcui    string `json:"rxcui"`
		Name     string `json:"name"`
		Synonym  string `json:"synonym"`
		Tty      string `json:"tty"`
		Language string `json:"language"`
		Suppress string `json:"suppress"`
		Umlscui  string `json:"umlscui"`
	}

	type ConceptGroup struct {
		Tty               string               `json:"tty,omitempty"`
		ConceptProperties []ConceptProperty    `json:"conceptProperties,omitempty"`
	}

	type DrugGroup struct {
		Name         *string        `json:"name"`
		ConceptGroup []ConceptGroup `json:"conceptGroup"`
	}

	type DrugGroupResponse struct {
		DrugGroup DrugGroup `json:"drugGroup"`
	}

	var drugs []commonsModels.Drug
	endpoint := fmt.Sprintf("/REST/drugs.json?name=%s", name)
	data, err := api.fetchFromEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	var resp DrugGroupResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, &errors.DrugNotFoundError{}
	}

	for _, cg := range resp.DrugGroup.ConceptGroup {
		for _, cp := range cg.ConceptProperties {
			drugs = append(drugs, commonsModels.Drug{
				ID:        cp.Rxcui,
				Name:      cp.Name,
				Substance: name,
			})
		}
	}

	return drugs, nil
}
