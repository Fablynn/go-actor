package logger

import (
	"fmt"
	"go-actor/library/builder"
	"go-actor/library/queue"
	"os"
	"path"
	"sync"
	"time"
)

type LogWriter struct {
	sync.WaitGroup
	lpath   string
	lname   string
	logs    *queue.Queue[*MetaData]
	builder *builder.Builder
	exit    chan struct{}
	notify  chan struct{}
}

func NewLogWriter(lpath, lname string, size int) *LogWriter {
	w := &LogWriter{
		lpath:   lpath,
		lname:   lname,
		logs:    queue.NewQueue[*MetaData](),
		notify:  make(chan struct{}, 1),
		exit:    make(chan struct{}),
		builder: builder.NewBuilder(size),
	}
	w.Add(1)
	go w.run()
	return w
}

func (l *LogWriter) Write(data *MetaData) error {
	l.logs.Push(data)
	select {
	case l.notify <- struct{}{}:
	default:
	}
	return nil
}

func (l *LogWriter) Close() error {
	close(l.exit)
	l.Wait()
	return l.builder.Close()
}

func (l *LogWriter) run() {
	tt := time.NewTicker(1 * time.Second)
	defer func() {
		tt.Stop()
		l.handle()
		l.Done()
	}()
	for {
		select {
		case <-l.notify:
			l.handle()
		case <-tt.C:
			l.builder.Flush()
		case <-l.exit:
			return
		}
	}
}

func (l *LogWriter) handle() {
	for mm := l.logs.Pop(); mm != nil; mm = l.logs.Pop() {
		filename := mm.GetFileName(l.lname)
		l.builder.Set(path.Join(l.lpath, filename))
		l.builder.Write(mm.Bytes())
		mm.Done()
	}
}

type StdWriter struct{}

func (d *StdWriter) Write(data *MetaData) error {
	defer data.Done()
	_, err := fmt.Fprint(os.Stdout, data.tt.Format("2006-01-02 15:04:05")+"\t"+data.String())
	return err
}

func (d *StdWriter) Close() error {
	return nil
}
