package scipipe

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type fileWriter struct {
	task
	FilePath chan string
	In       chan []byte
}

func NewFileWriter() *fileWriter {
	return &fileWriter{
		FilePath: make(chan string),
	}
}

func NewFileWriterFromPath(pl *Pipeline, path string) *fileWriter {
	t := &fileWriter{
		FilePath: make(chan string),
	}
	pl.AddTask(t)
	go func() {
		t.FilePath <- path
	}()
	return t
}

func (self *fileWriter) Run() {
	f, err := os.Create(<-self.FilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for line := range self.In {
		w.WriteString(fmt.Sprint(string(line), "\n"))
	}
	w.Flush()
}