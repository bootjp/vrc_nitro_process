package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"gopkg.in/yaml.v2"
)

type Run struct {
	Name     string `yaml:"name"`
	Path     string `yaml:"path"`
	SleepSec int    `yaml:"sleep_sec"`
}
type Process []Run

const vrcLogName = "output_log"
const vrcRelativeLogPath = `\AppData\LocalLow\VRChat\VRChat\`

func lunch(run Run) error {
	cmd := &exec.Cmd{
		Path: os.Getenv("COMSPEC"),
		SysProcAttr: &syscall.SysProcAttr{
			CmdLine: fmt.Sprintf(`/S /C start %s`, run.Path),
			// Foreground: true,
		}, // when run non windows environment please comment out this line. because this line is window only system call.
	}

	out, err := cmd.Output()
	fmt.Printf("%s\n", out)
	return err
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

var debug bool

func setupDebugMode(home string) {
	debug = os.Getenv("DEBUG") == "true"
	debug = debug || strings.Contains(home, "bootjp")
	if debug {
		fmt.Println("ENABLE DEBUG MODE")
	}
}
func main() {
	d, err := os.Getwd()
	if err != nil {
		fmt.Println("home directory not detected")
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(d + "/setting.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	process := &Process{}
	err = yaml.Unmarshal(data, &process)
	if err != nil {
		log.Fatalf("error: %v", err)
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

				if event.Op&fsnotify.Create != fsnotify.Create {
					continue
				}

				if !strings.Contains(event.Name, vrcLogName) {
					continue
				}

				for _, r := range *process {
					go func(r Run) {
						log.SetPrefix(r.Name)
						if r.SleepSec > 0 {
							time.Sleep(time.Duration(r.SleepSec))
						}
						if err := lunch(r); err != nil {
							log.Fatal(err)
						}
					}(r)

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()
	home := UserHomeDir()
	setupDebugMode(home)
	if debug {
		for _, v := range *process {
			fmt.Println(v.Name)
			fmt.Println(v.Path)
			fmt.Println(v.SleepSec)
		}
	}
	err = watcher.Add(home + vrcRelativeLogPath)

	if err != nil {
		log.Fatal(err)
	}
	<-done

}
