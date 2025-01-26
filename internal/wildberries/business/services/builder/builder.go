package builder

type Builder interface {
	Build() (interface{}, error)
	Clear()
}
