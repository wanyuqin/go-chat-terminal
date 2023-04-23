package service

type Server interface {
	Run() error
	Stop() error
}
