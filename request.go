// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

import (
	"errors"
)

// Request is the generic type used to communicate information to and from managers.
// 	None of the data in request needs to be private as none of them have race conditions.
// 	unless maliciously used by others.
type Request struct {

	// Route is used to determine what this request is asking for.
	Route string

	// Data is the information being transfered during the request.
	Data interface{}

	// Response is what is sent back when the process is finished
	// Response is a channel so that await commands can wait for the process
	// 	thread to finish it's computations. This is not necessary for a user to see.
	response chan responseStruct
}

// responseStruct is the default type returned by objects
type responseStruct struct {
	Data  interface{}
	Error error
}

/////////////
// METHODS //
/////////////

// Binding for manager.SendRequest() with the overhead of fetching manager by name.
func (request *Request) Send(managerName string) error {

	// Get the required manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return errors.New(managerName + " manager is not created or has been deleted (occurred during request send).")
	}

	// Otherwise, send the request to the manager and return with no errors
	manager.SendRequest(request)
	return nil

}

// Binding for manager.AwaitRequest() with the overhead of fetching manager by name.
func (request *Request) Await(managerName string) (interface{}, error) {

	// Get the required manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return nil, errors.New(managerName + " manager is not created or has been deleted (occurred during request send).")
	}

	// Call the binding
	return manager.AwaitRequest(request)

}

// Inverse binding for manager.SendRequest().
func (request *Request) SendManager(manager *Manager) {

	// Call the manager binding
	manager.SendRequest(request)

}

// Inverse binding for manager.AwaitRequest().
func (request *Request) AwaitManager(manager *Manager) (interface{}, error) {

	// Call the manager binding
	return manager.AwaitRequest(request)

}

// Wait is used for a job which has already been sent and the response wants to be waited on.
// 	Once the response is given, the data will be parsed and returned.
func (request *Request) Wait() (interface{}, error) {

	// Just wait for data to be put in the response
	response := <-request.response
	return response.getData()

}

// Check to see if the request has been carried out yet. As long as there are responses,
// 	the request "has data"
func (request *Request) HasData() bool {
	return len(request.response) > 0
}

////////////////////////
// INTERNAL FUNCTIONS //
////////////////////////

// Internal function for storing a response
func (request *Request) storeResponse(response responseStruct) {
	request.response <- response
}

/*
	GetData will either return data or an error depending on whether or
	not there is an error present in the data. Handy for use when you
	have nested response objects.
*/
func (response *responseStruct) getData() (interface{}, error) {

	// If there is no error, we need to extract the actual data.
	if response.Error == nil {
		data := response.Data

		/*
			If the data is a channel, we want to wait until data is passed into that channel
			and then use that data as the main data response.
		*/
		responseChannel, ok := data.(chan interface{})
		if ok {
			data = <-responseChannel
		}

		// Check that data is not a response struct. If it is, repeat the process and return the smallest child.
		responseData, ok := data.(*responseStruct)
		if ok {
			return responseData.getData()
		}

		return response.Data, nil
	}

	// If there is an error, we just return it.
	return nil, response.Error

}
