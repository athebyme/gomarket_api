package parse

type BuilderEngine interface {
	Build() (interface{}, error)
}
