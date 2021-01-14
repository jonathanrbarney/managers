// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
	"fmt"
	"sync"
)

// Internal managers struct used for direct requests
var managers = make(map[string]*Manager)
var managersLock = sync.Mutex{}

// getManager is an internal function to grab a manager
func getManager(managerName string) (*Manager, bool) {

	managersLock.Lock()
	defer managersLock.Unlock()
	manager, ok := managers[managerName]
	return manager, ok

}

// Manager is the type used to process and respond to requests.
type Manager struct {

	// Name is just a user defined name for the manager
	Name string

	// Requests is a channel used to keep track of everything the manager has
	// 	been asked to do.
	Requests chan *Request

	// Functions is a map of request type to respective processing function.
	//	These functions will take in a request interface and respond with a response interface.
	Functions map[string]func(request interface{}) interface{}

	// StateLock determines whether or not "Functions" can be read or editted.
	StateLock sync.Mutex
}

// Start will start the processing function for the manager
func (manager *Manager) Start() {

	// Big for loop for the manager to handle incomming requests
	for {

		// Extract the request and decide qhat to do based on what the route is
		request := <-manager.Requests

		// Response object data. Initialize to nil values.
		response := Response{
			Data:  nil,
			Error: nil,
		}

		// Internal kill command for the manager
		if request.Route == "state|kill-manager" {

			// Signify the request was processed and then break out of the processing loop.
			request.Response <- response
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
				response.Data = function(request.Data)
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
			request.Response <- response

		}

	}

	// Remove the manager from the array
	managersLock.Lock()
	defer managersLock.Unlock()
	delete(managers, manager.Name)

}

// Send will send a job to the manager and not wait for completion
func (manager *Manager) Send(route string, data interface{}) *Request {

	// Create a new request object
	request := NewRequest(route, data)

	// Send the job to the manager
	manager.Requests <- request

	// Respond with the request
	return request

}

// Await will send a job to the manager and await completion
func (manager *Manager) Await(route string, data interface{}) *Response {

	// Create and send the request to the manager
	request := manager.Send(route, data)

	// Wait for the request to complete
	response := request.wait()

	// Return the respons to the user
	return response

}

// Kill is a default request which will halt the manager
func (manager *Manager) Kill() {

	// Just send a kill request and wait for completion
	manager.Await("state|kill-manager", nil)

}

// Attach will attach a process to a manager
func (manager *Manager) Attach(route string, function func(request interface{}) interface{}) {

	// This is simple as just attaching the function
	manager.StateLock.Lock()
	defer manager.StateLock.Unlock()
	manager.Functions[route] = function

}

// getFunction returns the function of a given name
func (manager *Manager) getFunction(route string) (func(request interface{}) interface{}, bool) {

	// This is simple as just returning the function
	manager.StateLock.Lock()
	defer manager.StateLock.Unlock()
	function, ok := manager.Functions[route]
	return function, ok

}
