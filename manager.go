// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
	"fmt"
	"sync"
)

///////////////////////////
// PUBLIC MANAGERS STATE //
///////////////////////////

// Internal managers struct used for public requests. This data is just
// 	for storing a link from a manager name to the manager. The main use case
// 	for this is to allow managers to be created and used without tracking the
//  the handle to the manager.
var managersMap = make(map[string]*Manager)
var managersLock = sync.Mutex{}

// getManager is an internal function to grab a manager from the managersMap.
// 	This function uses the managersLock to ensure thread safety.
func getManager(managerName string) (*Manager, bool) {

	managersLock.Lock()
	defer managersLock.Unlock()
	manager, ok := managersMap[managerName]
	return manager, ok

}

// deleteManager is an internal function for deleting a manager from the managersMap.
// 	This function uses the managersLock to ensure thread safety.
func deleteManager(managerName string) {
	managersLock.Lock()
	defer managersLock.Unlock()
	delete(managersMap, managerName)
}

/////////////
// MANAGER //
/////////////

// Manager is the struct used to process and respond to requests. The object itself is
// 	quite simple. See below descriptions for what each attribute does.
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

	// stateLock determines whether or not values in the Manager can be read or editted.
	// 	The only exception is the Name, which the "managers" package doesn't care about.
	// 	We will let clients control access to this.
	stateLock sync.Mutex
}

// Start will start the processing function for the manager. The for loop below is the
// 	loop which handles the process. It's very straightforward. Just loop through and process
// 	each request as they come through until a kill request is sent.
func (manager *Manager) Start(managerState interface{}) {

	// Freeze the state so that the manager can be set to running. Then unfreeze so
	// 	the rest of the data can be read (like the functions)
	manager.stateLock.Lock()
	manager.running = true
	manager.stateLock.Unlock()

	// Big for loop for the manager to handle incomming requests
	for {

		// Wait for a request to come in before parsing it
		// 	and deciding what to do based on the route.
		request := <-manager.requests

		// Response object data. Initialize to nil values. The response
		// 	will be populated with data as the route function is processed.
		response := responseStruct{
			Data:  nil,
			Error: nil,
		}

		// Internal kill command for the manager. When manager.Kill() is called, it
		// 	will send this route. This will just store an arbitrary response and then
		// 	break out of the processing loop.
		if request.Route == "state|kill-manager" {

			// Signify the request was processed and then break out of the processing loop.
			request.storeResponse(response)
			break

			// User defined commands will end up here
		} else {

			// Check to see if that route was added.
			//	If it wasn't, create an error.
			//	If it was, process the job .
			function, ok := manager.getFunction(request.Route)
			if !ok {
				response.Error = errors.New("No function named " + request.Route + " added to " + manager.Name + " manager.")
			} else {

				// If here, it's time to process the job. We simply send the managerState to the
				// 	processing function along with the requested data.
				response.Data = function(managerState, request.Data)

				// If there is an error with the process, set the error appropriately. Also
				// 	remove the original response data as it was an error.
				if err, ok := response.Data.(error); ok {
					response.Data = nil
					response.Error = err
				}
			}

			// If there is an error, just let the user know about it.
			// TODO: Maybe find a better way to handle this? I think this is ok for now though.
			if response.Error != nil {
				fmt.Println("Error in manager, " + manager.Name + ":")
				fmt.Println(response.Error)
			}

			// Add the response to the request. All this does is send the response in the
			// 	response channel on the request. This allows the "Wait" function on the
			// 	request to respond appropriately.
			request.storeResponse(response)

		}

	}

	// Freeze the state so that the manager can be set to not running
	manager.stateLock.Lock()
	manager.running = false
	manager.stateLock.Unlock()

}

// IsRunning will just return the value of manager.running. Simple binding so that we
// 	can ensure thread safety of manager attributes.
func (manager *Manager) IsRunning() bool {
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	return manager.running
}

///////////////////////
// REQUEST FUNCTIONS //
///////////////////////

// Send will send a job to the manager and not wait for completion. See Request.Send()
// 	for a more detailed description of how this works.
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

// Await will send a job to the manager and await completion. See Request.Await()
// 	for a more detailed description of how this works.
func (manager *Manager) Await(route string, data interface{}) (interface{}, error) {

	// Create and send the request to the manager
	request := manager.Send(route, data)

	// Wait for the request to complete
	return request.Wait()

}

// AwaitRequest will queue a premade request to the manager. This is mainly just to ensure
// 	that the .requests field can stay hidden.
func (manager *Manager) AwaitRequest(request *Request) (interface{}, error) {
	manager.SendRequest(request)
	return request.Wait()
}

/////////////
// CONTROL //
/////////////

// Kill is an internal request which will halt the manager. This is blocking and will wait
// 	for the manager to actually stop processing. Just detach in a go-routine if you'd like to
// 	kill without waiting for a success.
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

///////////////
// FUNCTIONS //
///////////////

// Attach will attach a function to a manager at a specific route. Once a function is
// 	attached, requests sent to the manager are able to find and use the function.
func (manager *Manager) Attach(route string, function func(managerState interface{}, request interface{}) interface{}) {

	// This is simple as just attaching the function
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	manager.functions[route] = function

}

// Detach will remove a specified route from a manager. Once a function is detached,
// 	requests sent to the manager will no longer be accessible.
func (manager *Manager) Detach(route string) {
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	delete(manager.functions, route)
}

// getFunction returns the function of a given name. This is just an internal function
// 	to handle race conditions.
func (manager *Manager) getFunction(route string) (func(managerState interface{}, request interface{}) interface{}, bool) {

	// This is simple as just returning the function
	manager.stateLock.Lock()
	defer manager.stateLock.Unlock()
	function, ok := manager.functions[route]
	return function, ok

}
