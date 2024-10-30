package service

type LogUpdate interface {
	LogUpdate(query string) error
}
