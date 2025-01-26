package request

type Model interface {
	ToBytes() ([]byte, error)
}
