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

	// Create a pointer to a new manager for clients to use.
	newManager := &Manager{
		Name:      name,
		Requests:  make(chan *Request, bufferSize),
		Running: false,
		Functions: make(map[string]func(managerState interface{}, request interface{}) interface{}),
		StateLock: sync.Mutex{},
	}

	// Mutex management
	managersLock.Lock()
	defer managersLock.Unlock()

	// Check that the manager name doesn't already exist
	if _, exists := managersMap[name]; exists {
		return nil, errors.New("Manager with name " + name + " already exists!")
	}

	// Add it to the managers map and return it
	managersMap[name] = newManager
	return newManager, nil

}

// NewRequest will return a new request with the given Route and input Data
func NewRequest(route string, data interface{}) *Request {

	// Create a new request and return it with the give values
	return &Request{
		Route:    route,
		Data:     data,
		Response: make(chan Response, 1),
	}

}

// Send will create and send a request to a defined manager
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

// Await will create and send a request to a defined manager and respond with the completed data
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

// Attach a function to a manager
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

// Start a manager
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

func Kill(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during kil).")
	}


	// Just send a kill request and wait for completion
	return manager.Kill()

}

func Remove(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during remove).")
	}

	return manager.Remove()

}

func KillAndRemove(managerName string) error {

	manager, exists := getManager(managerName)
	if !exists {
		return errors.New(managerName + " manager doesn't exist or has been deleted (occurred during killAndRemove).")
	}

	return manager.KillAndRemove()

}