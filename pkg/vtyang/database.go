package vtyang

type DBNodeType string

const (
	Container DBNodeType = "container"
	List      DBNodeType = "list"
	Leaf      DBNodeType = "leaf"
	LeafList  DBNodeType = "leaf-list"
)

type DBNode struct {
	Name       string
	Type       DBNodeType
	Childs     []DBNode
	ListChilds [][]DBNode
	Value      DBValue
}

type DBValueType string

const (
	YString  DBValueType = "string"
	YInteger DBValueType = "integer"
	YBoolean DBValueType = "boolean"
)

type DBValue struct {
	Type DBValueType

	// Union
	Integer int
	String  string
	Boolean string
}

type DB struct {
	active bool
	root   DBNode
}
