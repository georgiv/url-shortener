package api

type DbWorker interface {
	Persist() error
}
