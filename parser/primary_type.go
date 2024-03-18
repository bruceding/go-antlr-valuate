package parser

type PrimaryType int

const (
	UNKNOWN PrimaryType = iota + 1
	STRING
	FLOAT
	BOOL
	ARRAY
	TIME
	MAP
	ARRAYORMAP
)

func (t PrimaryType) String() string {
	switch t {
	case FLOAT:
		return "float"
	case STRING:
		return "string"
	case BOOL:
		return "boolean"
	case ARRAY:
		return "array"
	case MAP:
		return "map"
	case TIME:
		return "time"
	case ARRAYORMAP:
		return "arrayormap"
	default:
		return "unknown"
	}
}
