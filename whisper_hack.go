package redis

type ExtCmder interface {
	Cmder
	SetArgs(args []interface{})
}

func (cmd *baseCmd) SetArgs(args []interface{}) {
	cmd._args = args
}

func (pipe *Pipeline) SetProcessor(fn func(Cmder) error) {
	pipe.setProcessor(fn)
}
