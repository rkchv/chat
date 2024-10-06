package closer

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

// CloseFunc тип для callback-функций, которые будут вызываться
type CloseFunc func() error

var globalCloser = New()

// Add добавляет колбэк-функцию в список тех, что будут вызваны при всеобщем завершении
func Add(f ...CloseFunc) {
	globalCloser.Add(f...)
}

// Wait ожидает отработки всех функций добавленных через Add
func Wait() {
	globalCloser.Wait()
}

// CloseAll Запускает отработку всех функций, добавленных через Add
func CloseAll() {
	globalCloser.CloseAll()
}

// Closer Обертка для  мягкого (graceful) завершения процесса.
type Closer struct {
	mu    sync.Mutex
	once  sync.Once
	done  chan struct{}
	funcs []CloseFunc
}

// New возвращает новый объект Closer, если указать конкретные сигналы os, то отреагирует только на них
func New(sig ...os.Signal) *Closer {
	c := &Closer{done: make(chan struct{})}
	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.CloseAll()
		}()
	}
	return c
}

// Add добавляет колбэк-функцию в список тех, что будут вызваны при всеобщем завершении
func (c *Closer) Add(f ...CloseFunc) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

// Wait ожидает отработки всех функций добавленных через Add
func (c *Closer) Wait() {
	<-c.done
}

// CloseAll Запускает отработку всех функций, добавленных через Add
func (c *Closer) CloseAll() {
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f CloseFunc) {
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
