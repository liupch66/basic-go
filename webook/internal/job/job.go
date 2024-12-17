package job

type Job interface {
	Name() string
	// Run 要是加入链路追踪，这里还需要接收一个 context 参数
	Run() error
}
