package servicehandler

type ServiceBase interface {
	Run() error
	Stop()
}
