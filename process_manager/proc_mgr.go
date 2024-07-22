package processmanager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ProcessManager struct {
	LaunchConfigPath   string
	ShutdownConfigPath string
	LaunchProcesses    []PackageExecDetails
	ShutdownProcesses  []PackageExecDetails
	wg                 sync.WaitGroup
}

func CreateProcessManager(launchConfigPath string, shutdownConfigPath string) ProcessManager {
	return ProcessManager{
		LaunchConfigPath:   launchConfigPath,
		ShutdownConfigPath: shutdownConfigPath,
		LaunchProcesses:    readPackageDetails(launchConfigPath),
		ShutdownProcesses:  readPackageDetails(shutdownConfigPath),
	}
}

func readPackageDetails(filename string) []PackageExecDetails {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return make([]PackageExecDetails, 0)
	}
	defer file.Close()

	// Read File Contents
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	var processes []PackageExecDetails
	err = json.Unmarshal(data, &processes)
	if err != nil {
		fmt.Println("Error interpreting json:", err)
	}

	return processes
}

func (pm *ProcessManager) execProcess(details PackageExecDetails, done <-chan bool) {
	pm.wg.Add(1)
	defer pm.wg.Done()

	if details.ParentProcess.KeepAlive {
		for {
			select {
			case doneStatus := <-done:
				if doneStatus {
					return
				}
			default:
				args := strings.Split(details.ParentProcess.LaunchCommand, " ")
				cmd := exec.Command("cmd", "/k")
				cmd.Args = append(cmd.Args, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Start()
				if err != nil {
					fmt.Println("Error starting process:", err)
					continue
				}

				err = cmd.Wait()
				if err != nil {
					fmt.Println("Process exited with error:", err)
				}
				time.Sleep(10 * time.Second)
			}
		}
	} else {
		args := strings.Split(details.ParentProcess.LaunchCommand, " ")
		cmd := exec.Command("cmd", "/k")
		cmd.Args = append(cmd.Args, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			fmt.Println("Error starting process:", err)
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Println("Process exited with error:", err)
		}

		time.Sleep(time.Duration(details.ParentProcess.ExtraDelay) * time.Second)

		for _, childProcess := range details.ChildProcesseses {
			go pm.execProcess(childProcess, done)
		}
	}
}

func (pm *ProcessManager) worker(baseProcess PackageExecDetails, done <-chan bool) {
	pm.execProcess(baseProcess, done)
}

func (pm *ProcessManager) stopWorkers() {
	for _, mainProcess := range pm.ShutdownProcesses {
		pm.worker(mainProcess, nil)
	}
}

func (pm *ProcessManager) StartWorkers() {
	var persistentProcesseses int = 0
	for _, mainProcess := range pm.LaunchProcesses {
		if mainProcess.ParentProcess.KeepAlive {
			persistentProcesseses++
		}
	}

	done := make(chan bool, persistentProcesseses)
	for _, mainProcess := range pm.LaunchProcesses {
		go pm.worker(mainProcess, done)
	}

	time.Sleep(10 * time.Second)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if scanner.Scan() {
			if userCommand := scanner.Text(); userCommand == "quit" {
				for i := 0; i < persistentProcesseses; i++ {
					done <- true
				}
				close(done)
				pm.stopWorkers()
				break
			} else {
				fmt.Println("You typed:", userCommand, "did you mean to type 'quit'?")
			}
		} else if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from stdin:", err)
			break
		}
	}

	pm.wg.Wait()
}
