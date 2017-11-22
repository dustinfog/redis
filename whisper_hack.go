package redis

type ExtCmder interface {
	Cmder
	SetArgs(args []interface{})
}

func (cmd *baseCmd) SetArgs(args []interface{}) {
	cmd._args = args
}

func (pipe *Pipeline) Cmds() []Cmder {
	return pipe.cmds
}
