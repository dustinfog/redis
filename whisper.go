package redis

import (
	"errors"
	"time"
)

//todo Acquiring status for repeated use, default status to Acquiring
//todo may be lock for embeded in other object, needed to argument
type WhisperLockStatus int

const (
	WHISPER_LOCK_STATUS_FREE WhisperLockStatus = iota
	WHISPER_LOCK_STATUS_ACQUIRED
	WHISPER_LOCK_STATUS_RELEASING
)

type WhisperLock struct {
	statefulCmdable
	*Client
	name     string
	demander string
	expire   time.Duration
	status   WhisperLockStatus
}

func NewWhisperLock(client *Client, name, demander string, expire time.Duration) *WhisperLock {
	wl := &WhisperLock{
		Client:   client,
		name:     name,
		demander: demander,
		expire:   expire,
	}
	wl.setProcessor(wl.Process)
	return wl
}

func (wl *WhisperLock) Acquire() error {
	if wl.status != WHISPER_LOCK_STATUS_FREE {
		return errors.New("locked")
	}

	cmd := NewCmd("WHISPER.LOCK",
		wl.name,
		wl.demander,
		int64(wl.expire)/1000000,
	)
	wl.Client.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		//todo 重试
		return err
	}

	wl.status = WHISPER_LOCK_STATUS_ACQUIRED
	return nil
}

func (wl *WhisperLock) Keep(expire time.Duration) error {
	if wl.status != WHISPER_LOCK_STATUS_ACQUIRED {
		return errors.New("not locked")
	}

	cmd := NewCmd("WHISPER.KEEP",
		wl.name,
		wl.demander,
		int64(expire)/1000000,
	)
	wl.Client.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		return err
	}

	wl.expire = expire
	return nil
}

func (wl *WhisperLock) Release(delay bool) error {
	if wl.status != WHISPER_LOCK_STATUS_ACQUIRED {
		return errors.New("not locked")
	}

	if delay {
		wl.status = WHISPER_LOCK_STATUS_RELEASING
		return nil
	}

	cmd := NewCmd("WHISPER.RELEASE",
		wl.name,
		wl.demander,
	)
	wl.Client.Process(cmd)
	_, err := cmd.Result()
	if err != nil {
		return err
	}

	wl.status = WHISPER_LOCK_STATUS_FREE
	return nil
}

func (wl *WhisperLock) Process(cmd Cmder) error {
	var args []interface{}
	if wl.status == WHISPER_LOCK_STATUS_ACQUIRED {
		args = []interface{}{
			"WHISPER.KEEP",
			wl.name,
			wl.demander,
			0}
	} else if wl.status == WHISPER_LOCK_STATUS_FREE {
		args = []interface{}{
			"WHISPER.LOCK",
			wl.name,
			wl.demander,
			0}
	} else {
		args = []interface{}{
			"WHISPER.UNLOCK",
			wl.name,
			wl.demander,
		}
	}

	args = append(args, cmd.Args()...)

	switch c := cmd.(type) {
	case *Cmd:
		c._args = args
	case *SliceCmd:
		c._args = args
	case *StatusCmd:
		c._args = args
	case *IntCmd:
		c._args = args
	case *DurationCmd:
		c._args = args
	case *TimeCmd:
		c._args = args
	case *BoolCmd:
		c._args = args
	case *StringCmd:
		c._args = args
	case *FloatCmd:
		c._args = args
	case *StringSliceCmd:
		c._args = args
	case *StringIntMapCmd:
		c._args = args
	case *ZSliceCmd:
		c._args = args
	case *ScanCmd:
		c._args = args
	case *ClusterSlotsCmd:
		c._args = args
	case *GeoLocationCmd:
		c._args = args
	case *GeoPosCmd:
		c._args = args
	case *CommandsInfoCmd:
		c._args = args
	default:
		return errors.New("not support")
	}

	err := wl.Client.Process(cmd)
	if err != nil {
		//todo 重试
	} else {
		if wl.status == WHISPER_LOCK_STATUS_FREE {
			wl.status = WHISPER_LOCK_STATUS_ACQUIRED
		} else if wl.status == WHISPER_LOCK_STATUS_RELEASING {
			wl.status = WHISPER_LOCK_STATUS_FREE
		}
	}

	return err
}

func (wl *WhisperLock) Pipelined(fn func(Pipeliner) error) ([]Cmder, error) {
	return wl.Pipeline().Pipelined(fn)
}

func (wl *WhisperLock) Pipeline() Pipeliner {
	pipe := Pipeline{
		exec: wl.pipelineExecer(wl.pipelineProcessCmds),
	}
	pipe.setProcessor(pipe.Process)
	return &pipe
}

func (wl *WhisperLock) TxPipelined(fn func(Pipeliner) error) ([]Cmder, error) {
	return wl.TxPipeline().Pipelined(fn)
}

func (wl *WhisperLock) TxPipeline() Pipeliner {
	pipe := Pipeline{
		exec: wl.pipelineExecer(wl.txPipelineProcessCmds),
	}
	pipe.setProcessor(pipe.Process)
	return &pipe
}
