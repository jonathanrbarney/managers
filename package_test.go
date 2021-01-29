// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
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
	startJob := func(managerState interface{}, request interface{}) interface{} {

		jobID = request.(string)
		jobStatus = "Processing"

		return nil
	}

	completeJob := func(managerState interface{}, request interface{}) interface{} {

		jobID = ""
		jobStatus = "Complete"

		return jobResult

	}

	updateStatus := func(managerState interface{}, request interface{}) interface{} {

		jobStatus = request.(string)
		value := managerState.(*int)
		*value = 9
		return nil

	}

	fail := func(managerState interface{}, request interface{}) interface{} {
		return errors.New("Test Error")
	}

	getResults := func(managerState interface{}, request interface{}) interface{} {

		fmt.Println()
		fmt.Println("ID", jobID)
		fmt.Println("Status:", jobStatus)
		fmt.Println("Result:", jobResult)

		v := managerState.(*int)
		fmt.Println("Value:", *v)
		return nil

	}

	// Attach all the processes to the manager
	manager.Attach("start", startJob)
	manager.Attach("complete", completeJob)
	manager.Attach("update", updateStatus)
	manager.Attach("get", getResults)
	manager.Attach("fail", fail)

	// Start the manager
	data := 4
	go manager.Start(&data)

	// Send some jobs!
	manager.Send("start", "1234")
	manager.Send("get", nil)
	Await("main", "update", "TEST")
	Await("main", "get", nil)

	// req := NewRequest("complete", nil)
	// req.Send("main")

	manager.Await("get", nil)
	err := manager.Await("fail", nil)
	if err != nil {
		t.Log("Successfully caught error.")
	}

}
