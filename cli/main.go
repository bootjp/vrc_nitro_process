package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"

	"gopkg.in/yaml.v2"
)

type Run struct {
	Name string
	Path string
}
type Process []Run

const vrcLogName = "output_log"

func lunch(run Run) error {
	cmd := &exec.Cmd{
		Path:   os.Getenv("COMSPEC"),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: `/S /C ` + run.Path,
			// Foreground: true,
		}, // when run non windows environment please comment out this line. because this line is window only system call.
	}

	out, err := cmd.Output()
	fmt.Println(out)
	return err
}

func main() {
	if len(os.Args) != 1 {
		fmt.Println("please specify yml file.")
		//os.Exit(1)
		// todo remove this line, because for debugging.

	}
	d, _ := os.Getwd()
	p := d + "/example.yml"
	fmt.Println(p)

	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	process := &Process{}
	err = yaml.Unmarshal(data, &process)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for _, v := range *process {
		fmt.Println(v.Name)
		fmt.Println(v.Path)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = watcher.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//// ignore hidden file
				//path := strings.Split(event.Name, "/")
				//
				//if path[len(path)-1] == "." {
				//	continue
				//}
				if event.Op&fsnotify.Create != fsnotify.Create {
					continue
				}

				if !strings.Contains(event.Name, vrcLogName) {
					continue
				}
				log.Println("create log file:", event.Name)
				for _, r := range *process {
					if err := lunch(r); err != nil {
						log.Fatal(err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add(`/Volumes/HomeNas`)
	//err = watcher.Add(`C:\Users\bootjp\AppData\LocalLow\VRChat\VRChat\`)
	if err != nil {
		log.Fatal(err)
	}
	<-done

}
