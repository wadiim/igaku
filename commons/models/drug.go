package models

type Drug struct {
	ID	string
	Name	string
	Substance	string
}

type DrugOrderableField string

const (
	DrugID		DrugOrderableField = "id"
	DrugName	DrugOrderableField = "name"
	SubstanceName		DrugOrderableField = "substance"
)

var DrugOrderableFieldsMap = map[string]DrugOrderableField{
	"id": DrugID,
	"name": DrugName,
	"substance": SubstanceName,
}
