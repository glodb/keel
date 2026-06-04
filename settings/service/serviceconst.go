package service

type ServiceType int

const (
	SERVICE_TYPE_HTTP   = ServiceType(1)
	SERVICE_TYPE_TCP    = ServiceType(2)
	SERVICE_TYPE_SIMPLE = ServiceType(4)
)
