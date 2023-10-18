package main

import (
	"fmt"
	"log"
	"time"

	"joxgit.github.com/process/ps"
)

func main() {

	const appName = "notepad.exe"

	for {

		processes, err := ps.FilterProcesses(appName)

		if err != nil {
			log.Fatal(err)
		}

		for _, proc := range processes {
			cpuTime, err := proc.CPUTime()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%v %v\n", proc, cpuTime)

			if cpuTime.User+cpuTime.System > time.Duration(500*time.Millisecond) {

				fmt.Printf("Killing process %v\n", proc)

				err = proc.Kill()

				if err != nil {
					log.Fatal(err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}

}

//some helper info...
//https://github.com/elastic/go-sysinfo/blob/main/providers/windows/process_windows.go#L32
//github.com/elastic/go-windows
