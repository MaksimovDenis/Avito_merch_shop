package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

var globalCloser = New()

func Add(f ...func() error) {
	globalCloser.Add(f...)
}

func Wait() {
	globalCloser.Wait()
}

func CloseAll() {
	globalCloser.CloseAll()
}

type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []func() error
}

func New(sig ...os.Signal) *Closer {
	closer := &Closer{done: make(chan struct{})}

	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			closer.CloseAll()
		}()
	}

	return closer
}

func (clr *Closer) Add(f ...func() error) {
	clr.mu.Lock()
	clr.funcs = append(clr.funcs, f...)
	clr.mu.Unlock()
}

func (clr *Closer) Wait() {
	<-clr.done
}

func (clr *Closer) CloseAll() {
	clr.once.Do(func() {
		defer close(clr.done)

		clr.mu.Lock()
		funcs := clr.funcs
		clr.funcs = nil
		clr.mu.Unlock()

		// call all Closer funcs async
		errs := make(chan error, len(funcs))

		for _, f := range funcs {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				log.Println("error returned from Closer")
			}
		}
	})
}
