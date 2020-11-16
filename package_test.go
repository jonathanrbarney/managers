// Created by Clayton Brown. See "LICENSE" file in root for more info.

package manager

import (
	"fmt"
	"testing"
)

// Test for everything
func Test(t *testing.T) {

	// Create the manager
	manager := NewManager("main", 8)

	// Create a couple variables to monitor
	jobID := ""
	jobStatus := ""
	jobResult := ""

	// Greate a couple functions
	startJob := func(request interface{}) interface{} {

		jobID = request.(string)
		jobStatus = "Processing"

		return nil
	}

	completeJob := func(request interface{}) interface{} {

		jobID = ""
		jobStatus = "Complete"

		return jobResult

	}

	updateStatus := func(request interface{}) interface{} {

		jobStatus = request.(string)
		return nil

	}

	getResults := func(request interface{}) interface{} {

		fmt.Println()
		fmt.Println("ID", jobID)
		fmt.Println("Status:", jobStatus)
		fmt.Println("Result:", jobResult)
		return nil

	}

	// Attach all the processes to the manager
	manager.Attach("start", startJob)
	manager.Attach("complete", completeJob)
	manager.Attach("update", updateStatus)
	manager.Attach("get", getResults)

	// Start the manager
	go manager.Start()

	// Send some jobs!
	manager.Send("start", "1234")
	manager.Send("get", nil)
	Await("main", "update", "TEST")
	Await("main", "get", nil)
	
	req := NewRequest("complete", nil)
	req.Send("main")

	manager.Await("get", nil)

	manager.Kill()

}