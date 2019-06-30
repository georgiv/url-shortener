package db

type DbWorker interface {
	Persist() error
}

type db struct {

}
