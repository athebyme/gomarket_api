package builder

type Proxy interface {
	Build() (interface{}, error)
}
