// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
	"fmt"
	"sync"
)

// Internal managers struct used for direct requests
var managersMap = make(map[string]*Manager)
var managersLock = sync.Mutex{}

// getManager is an internal function to grab a manager
func getManager(managerName string) (*Manager, bool) {

	managersLock.Lock()
	defer managersLock.Unlock()
	manager, ok := managersMap[managerName]
	return manager, ok

}

// deleteManager is an internal function for deleting a manager from the managers map
func deleteManager(managerName string) {
	managersLock.Lock()
	defer managersLock.Unlock()
	delete(managersMap, managerName)
}

// Manager is the type used to process and respond to requests.
type Manager struct {

	// Name is just a user defined name for the manager. This is the only public
	// 	value because it is not read anywhere internally, so we can pass it on to the users
	// 	to handle.
	Name string

	// Requests is a channel used to keep track of everything the manager has
	// 	been asked to do.
	requests chan *Request

	// Whether or not the manager is currently processing
	running bool

	// Functions is a map of request type to respective processing function.
	//	These functions will take in a request interface and respond with a response interface.
	functions map[string]func(managerState interface{}, request interface{}) interface{}

	// stateLock determines whether or not "Functions" and "Running" can be read or editted.
	stateLock sync.Mutex
}

// Start will start the processing function for the manager
func (manager *Manager) Start(managerState interface{}) {

	// Freeze the state so that the manager can be set to running
	manager.stateLock.Lock()
	manager.running = true
	manager.stateLock.Unlock()

	// Big for loop for the manager to handle incomming requests
	for {

		// Extract the request and decide qhat to do based on what the route is
		request := <-manager.requests

		// Response object data. Initialize to nil values.
		response := responseStruct{
			Data:  nil,
			Error: nil,
		}

		// Internal kill command for the manager
		if request.Route == "state|kill-manager" {

			// Signify the request was processed and then break out of the processing loop.
			request.storeResponse(response)
			break

			// User defined commands
		} else {

			// Check to see if that route was added.
			//	If it wasn't, create an error.
			//	If it was, process the job .
			function, ok := manager.getFunction(request.Route)
			if !ok {
				response.Error = errors.New("No function named " + request.Route + " added to " + manager.Name + " manager.")
			} else {
				response.Data = function(managerState, request.Data)
				if err, ok := response.Data.(error); ok {
					response.Error = err
				}
			}

			// If there is an error, just let the user know about it.
			if response.Error != nil {
				fmt.Println("Error in manager, " + manager.Name + ":")
				fmt.Println(response.Error)
			}

			// Add the response to the request
			request.storeResponse(response)

		}

	}

	// Freeze the state so that the manager can be set to not running
	manager.stateLock.Lock()
	manager.running = false
	manager.stateLock.Unlock()

}

// Send will send a job to the manager and not wait for completion
func (manager *Manager) IsRunning() bool {
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	return manager.running
}

// Send will send a job to the manager and not wait for completion
func (manager *Manager) Send(route string, data interface{}) *Request {

	// Create a new request object
	request := NewRequest(route, data)

	// Send the job to the manager
	manager.requests <- request

	// Respond with the request
	return request

}

// SendRequest will queue a premade request to the manager. This is mainly just to ensure
// 	that the .requests field can stay hidden and unaccessible to users. However, it can also
//  be utilized if a user wishes to interact with it in a different way.
func (manager *Manager) SendRequest(request *Request) {
	manager.requests <- request
}

// AwaitRequest will queue a premade request to the manager. This is mainly just to ensure
// 	that the .requests field can stay hidden.
func (manager *Manager) AwaitRequest(request *Request) (interface{}, error) {
	manager.SendRequest(request)
	return request.Wait()
}

// Await will send a job to the manager and await completion
func (manager *Manager) Await(route string, data interface{}) (interface{}, error) {

	// Create and send the request to the manager
	request := manager.Send(route, data)

	// Wait for the request to complete
	return request.Wait()

}

// Kill is a default request which will halt the manager
func (manager *Manager) Kill() error {

	// Just send a kill request and wait for completion
	_, err := manager.Await("state|kill-manager", nil)
	return err

}

// Remove is the function which will remove the manager from the public map.
// 	Once this is done, the manager should be deleted/removed from memory.
func (manager *Manager) Remove() error {

	// Can only remove if the manager is not running
	if manager.IsRunning() {
		return errors.New("Unable to remove manager " + manager.Name + " because it is currently running.")
	}
	
	deleteManager(manager.Name)
	return nil
}

// Kill is a default request which will halt the manager AND remove it from the map
func (manager *Manager) KillAndRemove() error {

	// Just send a kill request and wait for completion
	_, err := manager.Await("state|kill-manager", nil)
	if err != nil {
		return err
	}

	// Afterwards, remove the manager from the manager map
	return manager.Remove()

}

// Attach will attach a process to a manager
func (manager *Manager) Attach(route string, function func(managerState interface{}, request interface{}) interface{}) {

	// This is simple as just attaching the function
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	manager.functions[route] = function

}

// Detach will remove a specified route from a manager
func (manager *Manager) Detach(route string) {
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	delete(manager.functions, route)
}

// getFunction returns the function of a given name
func (manager *Manager) getFunction(route string) (func(managerState interface{}, request interface{}) interface{}, bool) {

	// This is simple as just returning the function
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	function, ok := manager.functions[route]
	return function, ok

}