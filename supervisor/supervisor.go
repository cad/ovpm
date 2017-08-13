// Package supervisor provides a generic API to watch and manage Unix processes.
package supervisor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
)

// State is a FSM state.
type State uint

// States
const (
	UNKNOWN State = iota
	RUNNING
	STOPPED
	STARTING
	STOPPING
	FAILED
	EXITED
)

func (s State) String() string {
	switch s {
	case RUNNING:
		return "RUNNING"
	case STOPPED:
		return "STOPPED"
	case STARTING:
		return "STARTING"
	case STOPPING:
		return "STOPPING"
	case FAILED:
		return "FAILED"
	case EXITED:
		return "EXITED"

	default:
		return "UNKNOWN"
	}
}

type transition struct {
	currState State
	nextState State
}

// Transition Table
var tt = []transition{
	transition{currState: STOPPED, nextState: STARTING},
	transition{currState: STARTING, nextState: RUNNING},
	transition{currState: STARTING, nextState: STARTING},
	transition{currState: STARTING, nextState: FAILED},

	transition{currState: RUNNING, nextState: STOPPING},
	transition{currState: RUNNING, nextState: EXITED},
	transition{currState: STOPPING, nextState: STOPPED},
	transition{currState: STOPPING, nextState: STOPPING},
}

// Supervisable is an interface that represents a process.
type Supervisable interface {
	Start()
	Stop()
	Restart()
	Status() State
}

// Process represents a unix process to be supervised.
type Process struct {
	lock            *sync.RWMutex
	state           State
	maxRetry        uint
	cmd             *exec.Cmd
	executable      string
	wdir            string
	args            []string
	done            chan error
	stop            chan bool
	out             *bytes.Buffer
	stateChangeCond *sync.Cond
	// stdin     io.WriteCloser
	// stdoutLog Logger
	// stderrLog Logger
}

// NewProcess returns a new process to be supervised.
func NewProcess(executable string, dir string, args []string) (*Process, error) {
	// initialize process and set the state to STOPPED without transitioning to it.
	p := Process{}
	if !isExist(executable) {
		return &p, fmt.Errorf("executable can not be found: %s", executable)
	}
	p.maxRetry = 3
	p.executable = executable
	p.wdir = dir
	p.args = args
	p.lock = &sync.RWMutex{}
	p.stateChangeCond = sync.NewCond(&sync.RWMutex{})
	p.state = STOPPED
	p.done = make(chan error)
	p.stop = make(chan bool)
	p.out = new(bytes.Buffer)

	return &p, nil
}

// isExist returns wether the given executable binary is found on the filesystem or not.
func isExist(executable string) bool {
	if _, err := os.Stat(executable); !os.IsNotExist(err) {
		return true
	}
	return false
}

// waitFor blocks until the FSM transitions to the given state.
func (p *Process) waitFor(state State) {
	for p.state != state {
		p.stateChangeCond.L.Lock()
		p.stateChangeCond.Wait()
		p.stateChangeCond.L.Unlock()
	}
}

// Start will run the process.
func (p *Process) Start() {
	p.transitionTo(STARTING)
}

// Stop will cause the process to stop.
func (p *Process) Stop() {
	p.transitionTo(STOPPING)
}

// Restart will cause a running process to restart.
func (p *Process) Restart() {
	if p.state == RUNNING {
		p.Stop()
	}
	p.waitFor(STOPPED)
	p.Start()
}

// Status returns the current state of the FSM.
func (p *Process) Status() State {
	return p.state
}

func (p *Process) permittable(state State) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, t := range tt {
		if p.state == t.currState && t.nextState == state {
			return true
		}
	}
	return false
}

func (p *Process) setState(state State) {
	p.lock.Lock()
	p.state = state
	p.lock.Unlock()
}

func (p *Process) transitionTo(state State) {
	if p.permittable(state) {
		p.stateChangeCond.L.Lock()
		logrus.WithField("cmd", p.executable).Debugf("transition: '%s' -> '%s'", p.state, state)
		if p.out.Len() > 0 {
			logrus.Debugf("STDOUT(err): %s", p.out.String())
		}
		p.setState(state)
		go p.run(state)()
		p.stateChangeCond.L.Unlock()
		p.stateChangeCond.Broadcast()
		return
	}
	logrus.Errorf("transition to '%s' from '%s' is not permitted!", p.state, state)
	return
}

func (p *Process) newCommand() *exec.Cmd {
	cmd := exec.Command(p.executable)
	cmd.Stdout = p.out
	cmd.Stderr = p.out
	cmd.Dir = p.wdir
	cmd.Args = append([]string{p.executable}, p.args...)

	currUsr, err := user.Current()
	if err != nil {
		logrus.Errorf("can not get current running user: %v", err)
	}

	uid, err := strconv.Atoi(currUsr.Uid)
	if err != nil {
		panic(fmt.Sprintf("can not convert string to int %s: %v", currUsr.Uid, err))
	}

	gid, err := strconv.Atoi(currUsr.Gid)
	if err != nil {
		panic(fmt.Sprintf("can not convert string to int %s: %v", currUsr.Gid, err))
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}}
	return cmd
}

func (p *Process) run(state State) func() {
	switch state {
	case STOPPED:
		return func() {
			p.done = make(chan error)
			p.stop = make(chan bool)
		}
	case STARTING:
		return func() {
			// Prepare the command and start.
			var err error
			for i := uint(1); i <= p.maxRetry; i++ {
				// Prepare the command to run.
				p.lock.Lock()
				p.cmd = p.newCommand()
				p.lock.Unlock()

				logrus.Debugf("running %s", p.executable)
				err = p.cmd.Start()
				if err != nil {
					logrus.Debugf("process can not be started: %v", err)
					logrus.Debugf("retrying... (%d/%d)", i, p.maxRetry)
					continue
				}
				break
			}
			// Max retry reached process still not started.
			if err != nil {
				p.transitionTo(FAILED)
				return
			}

			// Process started successfully.
			logrus.Debugf("process is started %s PID %d", p.executable, p.cmd.Process.Pid)

			// Process Observer
			go func() {
				err := p.cmd.Wait()
				if err != nil {
					logrus.Error(err)
					close(p.done)
					return
				}
				p.done <- err
				close(p.done)
			}()
			p.transitionTo(RUNNING)
		}
	case RUNNING:
		return func() {
			// Stop Observer
			go func() {
				select {
				// process is ordered to stop.
				case <-p.stop:
					p.transitionTo(STOPPING)
					return
				// process exited on it's own
				case err := <-p.done:
					if p.state == RUNNING {
						logrus.Infof("process exited: %v", err)
						p.transitionTo(EXITED)
						return
					}
				}
			}()
		}
	case STOPPING:
		return func() {
			gracefullyStopped := false

			// first try to kill the process, gracefully
			err := p.cmd.Process.Signal(os.Interrupt)
			if err != nil {
				logrus.Errorf("interrupt signal returned error: %v", err)
			}
			for i := uint(1); i <= p.maxRetry; i++ {
				select {
				case <-time.After(3 * time.Second):
					logrus.Debugf("retrying... (%d/%d)", i, p.maxRetry)
					err := p.cmd.Process.Signal(os.Interrupt)
					if err != nil {
						logrus.Errorf("interrupt signal returned error: %v", err)
					}
				case err = <-p.done:
					if err == nil {
						gracefullyStopped = true
						break
					}
					logrus.Debugf("process stopped with error: %v", err)
					break

				}
			}

			// process didn't exit and retry count is full
			// hard killing
			if !gracefullyStopped {
				err := p.cmd.Process.Kill()
				if err != nil {
					logrus.Fatal("can not kill process!")
				}
				<-p.done
			}
			logrus.Debugf("process stopped %s", p.executable)
			p.transitionTo(STOPPED)
		}
	case FAILED:
		return func() {
			logrus.Fatalf("failed to launch process: %s", p.executable)
		}
	case EXITED:
		return func() {
			logrus.Errorf("process exited unexpectedly: %s", p.executable)
			os.Exit(1)
		}
	default: // UNKNOWN
		return nil
	}
}
