package consumer

type IConsumer interface {
	Listen() error
	Stop() error
}
