package parse

type Parser interface {
	Fetch(interface{}) (interface{}, error)
}
