package builder

import (
	"go-actor/library/util"
	"os"
)

type Builder struct {
	n    int
	size int
	buff []byte
	fp   *os.File
}

func NewBuilder(size int) *Builder {
	return &Builder{buff: make([]byte, size), size: size}
}

func (d *Builder) Set(filename string) error {
	if d.fp != nil && !util.IsSameFile(d.fp, filename) {
		if err := d.Flush(); err != nil {
			return err
		}
		if err := d.fp.Close(); err != nil {
			return err
		}
		d.fp = nil
	}
	if d.fp == nil {
		fp, err := util.CreateFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR)
		if err != nil {
			return err
		}
		d.fp = fp
	}
	return nil
}

func (d *Builder) Write(buf []byte) error {
	diff := d.size - d.n
	if ll := len(buf); ll <= diff {
		copy(d.buff[d.n:], buf)
		d.n += ll
		return nil
	}
	if diff > 0 {
		copy(d.buff[d.n:], buf[:diff])
		d.n += diff
	}
	if err := d.Flush(); err != nil {
		return err
	}
	return d.Write(buf[diff:])
}

func (d *Builder) Flush() error {
	if d.fp == nil || d.n <= 0 {
		return nil
	}
	if _, err := d.fp.Write(d.buff[:d.n]); err != nil {
		return err
	}
	d.n = 0
	return nil
}

func (d *Builder) Close() error {
	if d.fp == nil {
		return nil
	}
	return d.fp.Close()
}
