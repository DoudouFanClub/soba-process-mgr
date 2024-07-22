package main

import processmanager "soba-process-mgr/process_manager"

func main() {
	mgr := processmanager.CreateProcessManager("launch_config.json", "shutdown_config.json")
	mgr.StartWorkers()
}