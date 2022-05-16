// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
	"math/rand"
	"testing"
	"time"
)

const PROCESS_DELAY = 30 // In milliseconds
const TEST_DURATION = 10 // In seconds
const JOB_INTERVAL = 25  // In milliseconds

// Test for public bindings. These are the binding you can use to ask for a
// 	manager to do something without having a handle on the actual manager object.
// 	In general, this is slower and it's preferred to not do things this way if you don't need to.
func Test_Public_Bindings(t *testing.T) {

	// Create a few public managers. Each will have their own internal state
	// 	and will automatically have the test functionsattached.
	createPublicManager(t, "Manager 1", 256)
	createPublicManager(t, "Manager 2", 256)
	createPublicManager(t, "Manager 3", 256)
	createPublicManager(t, "Manager 4", 256)

	//Test error handling for double creation
	_, err := NewManager("Manager 4", 256)
	if err == nil {
		t.Fail()
	}

	// For each of the managers, send a bunch of information to them
	go managerTest(t, nil, "Manager 1", "mr")
	go managerTest(t, nil, "Manager 2", "rr")
	go managerTest(t, nil, "Manager 3", "rrl")
	managerTest(t, nil, "Manager 4", "mr")

	// Ensure all managers have finished processing before shutting down
	<-time.Tick(3 * time.Second)

	if err := Detach("Manager 1", "get"); err != nil {
		t.Fail()
	}
	if err := Detach("Manager 1", "setStatus"); err != nil {
		t.Fail()
	}
	if err := Detach("Manager 1", "setValue"); err != nil {
		t.Fail()
	}
	if err := Detach("Manager 1", "square"); err != nil {
		t.Fail()
	}

	// Test fetching manager by name
	if m, err := GetManager("Manager 1"); err != nil {
		t.Fail()
	} else {
		// try a function that doesnt exist
		if _, err2 := m.Await("ThisFunctionDoesntExist", nil); err2 == nil {
			t.Fail()
		}
	}
	//test misnamed manager
	testDNEManager(t)

	//Try to remove before killing
	if err := Remove("Manager 1"); err == nil {
		t.Fail()
	}

	if err := Kill("Manager 1"); err != nil {
		t.Fail()
	}
	if err := Remove("Manager 1"); err != nil {
		t.Fail()
	}

	if err := Kill("Manager 2"); err != nil {
		t.Fail()
	}
	if err := Remove("Manager 2"); err != nil {
		t.Fail()
	}

	if err := KillAndRemove("Manager 3"); err != nil {
		t.Fail()
	}
	if err := KillAndRemove("Manager 4"); err != nil {
		t.Fail()
	}

	if len(managersMap) != 0 {
		t.Fail()
	}

}

// Test for internal manager bindings
func Test_Manager(t *testing.T) {

	// Create a few handled managers
	m1 := createHandledManager(t, "Manager 1", 256)
	m2 := createHandledManager(t, "Manager 2", 256)
	m3 := createHandledManager(t, "Manager 3", 256)
	m4 := createHandledManager(t, "Manager 4", 256)

	go managerTest(t, m1, "", "mr")
	go managerTest(t, m2, "", "rr")
	go managerTest(t, m3, "", "rrl")
	managerTest(t, m4, "", "mr")

	// Ensure all managers have finished processing before shutting down
	<-time.Tick(3 * time.Second)

	m1.Detach("get")
	m1.Detach("setStatus")
	m1.Detach("setValue")
	m1.Detach("square")

	if err := m1.Kill(); err != nil {
		t.Fail()
	}
	if err := m1.Remove(); err != nil {
		t.Fail()
	}

	if err := m2.Kill(); err != nil {
		t.Fail()
	}
	if err := m2.Remove(); err != nil {
		t.Fail()
	}

	if err := m3.KillAndRemove(); err != nil {
		t.Fail()
	}
	if err := m4.KillAndRemove(); err != nil {
		t.Fail()
	}

	if len(managersMap) != 0 {
		t.Fail()
	}

}

// Test that is complete is working
func Test_hasData(t *testing.T) {
	m := createHandledManager(t, "Manager", 256)
	go m.Start(&State{Status: "Starting Up", Value: 0})
	<-time.Tick(5 * time.Millisecond)
	r := m.Send("setStatus", "Status 1")
	if m.IsRunning() != true {
		t.Error("Didn't show manager as running")
	}
	if r.HasData() == true {
		t.Error("Didn't show request as incomplete")
	}
	<-time.Tick(5 * PROCESS_DELAY * time.Millisecond)
	if r.HasData() == false {
		t.Error("Didn't show request as complete")
	}
}

/////////////////////////
// INTERNAL TEST SETUP //
/////////////////////////
func createPublicManager(t *testing.T, managerName string, bufferSize int) {

	//Test error handling
	if err := Start(managerName, &State{Status: "Starting Up", Value: 0}); err == nil {
		t.Fail()
	}
	if err := Attach(managerName, "test", func(any, any) any { return nil }); err == nil {
		t.Fail()
	}
	if err := Detach(managerName, "test"); err == nil {
		t.Fail()
	}

	// First, create a manager with the specified name
	_, err := NewManager(managerName, bufferSize)
	if err != nil {
		t.Fail()
	}

	// Then we want to attach all the routes
	if err := Attach(managerName, "get", getTestState); err != nil {
		t.Fail()
	}
	if err := Attach(managerName, "setStatus", setTestStatus); err != nil {
		t.Fail()
	}
	if err := Attach(managerName, "setValue", setTestValue); err != nil {
		t.Fail()
	}
	if err := Attach(managerName, "square", getTestSquare); err != nil {
		t.Fail()
	}

	// Now we create the original state and use that to start the manager
	state := &State{Status: "Starting Up", Value: 0}
	if err := Start(managerName, state); err != nil {
		t.Fail()
	}

}

func createHandledManager(t *testing.T, managerName string, bufferSize int) *Manager {

	// First, create a manager with the specified name
	manager, err := NewManager(managerName, bufferSize)
	if err != nil {
		t.Fail()
	}

	// Then we want to attach all the routes
	manager.Attach("get", getTestState)
	manager.Attach("setStatus", setTestStatus)
	manager.Attach("setValue", setTestValue)
	manager.Attach("square", getTestSquare)

	// Now we create the original state and use that to start the manager
	state := &State{Status: "Starting Up", Value: 0}
	go manager.Start(state)

	// If here, nothing went wrong
	return manager

}

func managerTest(t *testing.T, manager *Manager, managerName string, requestsMode string) {

	/////////////////////////
	// WRAPPER DEFINITIONS //
	/////////////////////////

	// Used to send a bunch of requests to manager through the manager. If "Manager" is nil, we do it publically
	performManagerRequest := func(route string, await bool, request any) any {
		if manager == nil {
			if await {
				response, err := Await(managerName, route, request)
				if _, ok := request.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				_, err := Send(managerName, route, request)
				if _, ok := request.(error); err != nil && !ok {
					t.Fail()
				}
				return nil
			}
		} else {
			if await {
				response, err := manager.Await(route, request)
				if _, ok := request.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				manager.Send(route, request)
				return nil
			}
		}
	}

	// Used to send a bunch of requests to manager through requests.
	performRequestRequest := func(route string, await bool, requestData any) any {
		request := NewRequest(route, requestData)
		if manager == nil {
			if await {
				response, err := request.Await(managerName)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				err := request.Send(managerName)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return nil
			}
		} else {
			if await {
				response, err := request.AwaitManager(manager)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				request.SendManager(manager)
				return nil
			}
		}
	}

	// Used to send a bunch of requests to manager through requests.
	performRequestRequestLookup := func(route string, await bool, requestData any) any {
		request := NewRequest(route, requestData)
		if manager == nil {
			if await {
				response, err := AwaitRequest(managerName, request)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				err := SendRequest(managerName, request)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return nil
			}
		} else {
			if await {
				response, err := request.AwaitManager(manager)
				if _, ok := requestData.(error); err != nil && !ok {
					t.Fail()
				}
				return response
			} else {
				request.SendManager(manager)
				return nil
			}
		}
	}

	// Wrapper for the above wrappers
	performRequest := func(route string, await bool, data any) any {
		switch requestsMode {
		case "rr":
			return performRequestRequest(route, await, data)
		case "rrl":
			return performRequestRequestLookup(route, await, data)
		default:
			return performManagerRequest(route, await, data)
		}
	}

	/////////////////
	// ACTUAL TEST //
	/////////////////

	// Create the reference state so we can ensure the actual state lines up correctly
	referenceState := &State{Status: "Starting Up", Value: 0}

	// Create tickers for the test duration and for the job interval
	durationTicker := time.NewTicker(TEST_DURATION * time.Second)
	intervalTicker := time.NewTicker(JOB_INTERVAL * time.Millisecond)

	for {
		select {

		// The test is over if the duration ticker finishes
		case <-durationTicker.C:
			return

		// Otherwise, we are going to handle a request
		case <-intervalTicker.C:

			// Select a random operation to perform
			choice := []string{"get", "setStatus", "setValue", "square"}[rand.Intn(4)]

			switch choice {

			// If "get" case, check all different types we can get
			case "get":
				respChan := make(chan any)
				go func() {
					respChan <- "test"
				}()
				response := performRequest("get", true, respChan)
				response = performRequest("get", true, errors.New("test error"))
				response = performRequest("get", true, &responseStruct{"test", nil})

				response = performRequest("get", true, nil)
				state := response.(*State)
				if state.Status != referenceState.Status || state.Value != referenceState.Value {
					t.Fail()
				}

			// If "setStatus" case, just send the request and update local memory
			case "setStatus":
				newStatus := []string{
					"Status 1", "Status 2", "Status 3", "Status 4",
					"Status 5", "Status 6", "Status 7", "Status 8",
					"Status 9", "Status 10", "Status 11", "Status 12",
				}[rand.Intn(12)]
				performRequest("setStatus", false, newStatus)
				referenceState.Status = newStatus

			// If "setValue" case, just send the request and update local memory
			case "setValue":
				newValue := rand.Intn(1_000_000)
				performRequest("setValue", false, newValue)
				referenceState.Value = newValue

			case "square":
				response := performRequest("square", true, nil)
				square := response.(int)
				if square != referenceState.Value*referenceState.Value {
					t.Fail()
				}
			}

		}
	}

}

///////////////
// TEST DATA //
///////////////
type State struct {
	Status string
	Value  int
}

////////////////////
// TEST FUNCTIONS //
////////////////////
func getTestState(managerState any, request any) any {

	if request != nil {
		return request
	}

	// Put an arbitrary delay
	time.NewTicker(PROCESS_DELAY * time.Millisecond)

	// Extract the manager state and return it
	return managerState

}

func setTestStatus(managerState any, request any) any {

	// Put an arbitrary delay
	time.NewTicker(PROCESS_DELAY * time.Millisecond)

	// Extract the manager state and the request state
	state := managerState.(*State)
	newStatus := request.(string)

	state.Status = newStatus

	// We need a nil return because the function is expecting one
	return nil

}

func setTestValue(managerState any, request any) any {

	// Put an arbitrary delay
	time.NewTicker(PROCESS_DELAY * time.Millisecond)

	// Extract the manager state and the request state
	state := managerState.(*State)
	newValue := request.(int)

	state.Value = newValue

	// We need a nil return because the function is expecting one
	return nil

}

func getTestSquare(managerState any, request any) any {

	// Put an arbitrary delay
	time.NewTicker(PROCESS_DELAY * time.Millisecond)

	// Return the square of the state value
	state := managerState.(*State)
	return state.Value * state.Value

}

func testDNEManager(t *testing.T) {

	// Test error handling
	if _, err := GetManager("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}
	if err := Kill("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}
	if err := Remove("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}
	if err := KillAndRemove("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}

	//Test error handling for requests to no manager
	if _, err := Send("This Manager Doesn't Exist", "get", nil); err == nil {
		t.Fail()
	}
	if _, err := Await("This Manager Doesn't Exist", "get", nil); err == nil {
		t.Fail()
	}
	r := NewRequest("get", nil)
	if err := SendRequest("This Manager Doesn't Exist", r); err == nil {
		t.Fail()
	}
	if _, err := AwaitRequest("This Manager Doesn't Exist", r); err == nil {
		t.Fail()
	}
	if err := r.Send("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}
	if _, err := r.Await("This Manager Doesn't Exist"); err == nil {
		t.Fail()
	}
}
