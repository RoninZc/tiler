package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var SafeExitInst *SafeExit

func InitSafeExit() {
	SafeExitInst = new(SafeExit)
	go SafeExitInst.ListenSignal()
}

type SafeExit struct {
	funcs []func()
	mu    sync.Mutex
}

func (s *SafeExit) Register(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.funcs = append(s.funcs, f)
}

func (s *SafeExit) exit() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, f := range s.funcs {
		f()
	}
	os.Exit(0)
}

func (s *SafeExit) ListenSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for singal := range sigs {
		switch singal {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fmt.Printf("收到系统信号 %d, 正在停止任务, 请稍后", singal)
			s.exit()
		}
	}
}
