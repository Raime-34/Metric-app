package repository

type Repo[T any] interface {
	SetField(string, T)
	GetFields() map[string]T
	IncrementCounter()
}
