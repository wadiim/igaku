package utils

type Ordering string

const (
	Asc	Ordering = "asc"
	Desc	Ordering = "desc"
)

var OrderingsMap = map[string]Ordering{
	"asc": Asc,
	"desc": Desc,
}
