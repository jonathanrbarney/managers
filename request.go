// Created by Clayton Brown. See "LICENSE" file in root for more info.

package manager

import (
	"errors"
)

// Request is the generic type used to communicate information to and from managers.
type Request struct {

	// Route is used to determine what this request is asking for.
	Route string

	// Data is the information being transfered during the request.
	Data interface{}

	// Result is what is sent back when the process is finished
	Response chan Response
}

// Response is the default type returned by objects
type Response struct {
	Data  interface{}
	Error error
}

// Send will send a request to a specified manager
func (request *Request) Send(managerName string) error {

	// Get the required manager
	manager, ok := getManager(managerName)

	// If the manager doesn't exist, respond with an error
	if !ok {
		return errors.New(managerName + " manager is not created.")
	}

	// Otherwise, send the request to the manager and return with no errors
	manager.Requests <- request
	return nil

}

// Await will wait until a request is completed and respond with the result
func (request *Request) Await(managerName string) (*Response, error) {

	// First send the request
	err := request.Send(managerName)

	// If there is an error, respond with it
	if err != nil {
		return nil, err
	}

	// Otherwise wait for completion and return the result
	return request.wait(), nil

}

// wait is an internal command which will wait until a response is given.
func (request *Request) wait() *Response {

	// Just wait for data to be put in the response
	response := <-request.Response
	return &response

}
