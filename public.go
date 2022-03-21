// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
	"sync"
)

/*
NewManager will return a blank manager for use.
Buffer size is the number of requests for the manager to hold onto until it starts blocking
requests. The appropriate number will depend on how many requests you expect the manager
to recieve and how long each request takes to process.
*/
func NewManager(name string, bufferSize int) (*Manager, error) {

	// Create a pointer to a new manager for clients to use. The requests and functions
	// 	will be prepopulated for the user.
	newManager := &Manager{
		Name:      name,
		requests:  make(chan *Request, bufferSize),
		running:   false,
		functions: make(map[string]func(managerState interface{}, request interface{}) interface{}),
		stateLock: sync.Mutex{},
	}

	// Mutex management
	managersLock.Lock()
	defer managersLock.Unlock()

	// Check that the manager name doesn't already exist. If it does, we
	// 	will obviously return an error.
	if _, exists := managersMap[name]; exists {
		return nil, errors.New("Manager with name " + name + " already exists!")
	}

	// Add it to the managers map and return it
	managersMap[name] = newManager
	return newManager, nil

}

// NewRequest will return a new request with the given Route and input Data.
// 	The response channel will be appropriately generated as well. We default to
// 	length 1 channel because that's all we really need.
func NewRequest(route string, data interface{}) *Request {
	return &Request{
		Route:    route,
		Data:     data,
		response: make(chan responseStruct, 1),
	}
}

//////////////
// REQUESTS //
//////////////

// Binding for manager.Send() with the overhead of fetching manager by name.
func Send(managerName string, route string, data interface{}) (*Request, error) {

	// Get the manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return nil, errors.New(managerName + " manager is not created or has been deleted (occurred during public send).")
	}

	// Send a job to the manager and return with no errors
	return manager.Send(route, data), nil

}

// Binding for manager.SendRequest() with the overhead of fetching manager by name.
func SendRequest(managerName string, request *Request) error {

	// Get the manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return errors.New(managerName + " manager is not created or has been deleted (occurred during public sendRequest).")
	}

	// Send a job to the manager and return with no errors
	manager.SendRequest(request)

	return nil

}

// Binding for manager.Await() with the overhead of fetching manager by name.
func Await(managerName string, route string, data interface{}) (interface{}, error) {

	// Get the manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return nil, errors.New(managerName + " manager is not created or has been deleted (occurred during public await).")
	}

	// Send a job to the manager and return with no errors
	return manager.Await(route, data)

}

// Binding for manager.AwaitRequest() with the overhead of fetching manager by name.
func AwaitRequest(managerName string, request *Request) (interface{}, error) {
	// Get the manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return nil, errors.New(managerName + " manager is not created or has been deleted (occurred during public sendRequest).")
	}

	// Send a job to the manager and return with no errors
	return manager.AwaitRequest(request)
}

/////////////////////
// MANAGER CONTROL //
/////////////////////

// Binding for manager.Attach() with the overhead of fetching manager by name.
func Attach(managerName string, route string, f func(interface{}, interface{}) interface{}) error {

	// First grab the manager
	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during public attach).")
	}

	// Then attach the function
	manager.Attach(route, f)

	// If here, nothing went wrong
	return nil

}

// Binding for manager.Detach() with the overhead of fetching manager by name.
func Detach(managerName string, route string) error {

	// First grab the manager
	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during public attach).")
	}

	// Then detach the route
	manager.Detach(route)

	// If here, everything went well
	return nil

}

// Binding for manager.Start() with the overhead of fetching manager by name. The only
// 	difference is that the manager will automatically start detached. (Non-blocking call)
func Start(managerName string, managerState interface{}) error {

	// First grab the manager
	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during start).")
	}

	// Then start the manager
	go manager.Start(managerState)
	return nil

}

/////////////////////////////////////
// PUBLIC MANAGER CONTROL BINDINGS //
/////////////////////////////////////

// Simple function for fetching a manager by name
func GetManager(managerName string) (*Manager, error) {

	// First grab the manager
	manager, exists := getManager(managerName)
	if !exists {
		return nil, errors.New(managerName + " manager doesn't exist or has been deleted (occurred during getManager).")
	}

	return manager, nil

}

// Binding for manager.Kill() with the overhead of fetching manager by name.
func Kill(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during kil).")
	}

	// Just send a kill request and wait for completion
	return manager.Kill()

}

// Binding for manager.Remove() with the overhead of fetching manager by name.
func Remove(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during remove).")
	}

	return manager.Remove()

}

// Binding for manager.KillAndRemove() with the overhead of fetching manager by name.
func KillAndRemove(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during killAndRemove).")
	}

	return manager.KillAndRemove()

}
