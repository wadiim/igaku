package utils


type RxClassMinConcept struct {
	ClassId string `json:"classId"`
	ClassName string `json:"className"`
	ClassType string `json:"classType"`
}

type RxClassMinConceptList struct {
	Concepts []RxClassMinConcept `json:"rxclassMinConcept"`
}

type RxClassData struct {
	ConceptList RxClassMinConceptList `json:"rxclassMinConceptList"`
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
