package components

import (
	"fmt"

	"github.com/scipipe/scipipe"
)

// FileSplitter is a process that will split a file into multiple files, each
// with LinesPerSplit number of lines per file
type FileSplitter struct {
	name          string
	InFile        *scipipe.InPort
	OutSplitFile  *scipipe.OutPort
	LinesPerSplit int
	workflow      *scipipe.Workflow
}

// NewFileSplitter returns an initialized FileSplitter process that will split a
// file into multiple files, each with linesPerSplit number of lines per file
func NewFileSplitter(wf *scipipe.Workflow, name string, linesPerSplit int) *FileSplitter {
	fs := &FileSplitter{
		name:          name,
		InFile:        scipipe.NewInPort("in_file"),
		OutSplitFile:  scipipe.NewOutPort("out_split_file"),
		LinesPerSplit: linesPerSplit,
		workflow:      wf,
	}
	wf.AddProc(fs)
	return fs
}

// Name returns the name of the FileSplitter process
func (proc *FileSplitter) Name() string {
	return proc.name
}

// Run runs the FileSplitter process
func (proc *FileSplitter) Run() {
	defer proc.OutSplitFile.Close()

	fileReader := NewFileReader(proc.workflow, proc.Name()+"_file_reader")

	for ft := range proc.InFile.Chan {
		scipipe.Audit.Println("FileSplitter      Now processing input file ", ft.GetPath(), "...")

		go func() {
			defer close(fileReader.FilePath)
			fileReader.FilePath <- ft.GetPath()
		}()

		go fileReader.Run()

		i := 1
		splitIdx := 1
		splitFt := newSplitIPFromIndex(ft.GetPath(), splitIdx)
		if !splitFt.Exists() {
			splitfile := splitFt.OpenWriteTemp()
			for line := range fileReader.OutLine {
				// If we have not yet reached the number of lines per split ...
				/// ... then just continue to write ...
				if i < splitIdx*proc.LinesPerSplit {
					splitfile.Write(line)
					i++
				} else {
					splitfile.Close()
					splitFt.Atomize()
					scipipe.Audit.Println("FileSplitter      Created split file", splitFt.GetPath())
					proc.OutSplitFile.Send(splitFt)
					splitIdx++

					splitFt = newSplitIPFromIndex(ft.GetPath(), splitIdx)
					splitfile = splitFt.OpenWriteTemp()
				}
			}
			splitfile.Close()
			splitFt.Atomize()
			scipipe.Audit.Println("FileSplitter      Created split file", splitFt.GetPath())
			proc.OutSplitFile.Send(splitFt)
		} else {
			scipipe.Audit.Printf("Split file already exists: %s, so skipping.\n", splitFt.GetPath())
		}
	}
}

// IsConnected tells whether all the ports of the FileSplitter process are connected
func (proc *FileSplitter) IsConnected() bool {
	return proc.InFile.IsConnected() &&
		proc.OutSplitFile.IsConnected()
}

func newSplitIPFromIndex(basePath string, splitIdx int) *scipipe.IP {
	return scipipe.NewIP(basePath + fmt.Sprintf(".split_%v", splitIdx))
}
