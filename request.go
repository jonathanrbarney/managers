// Created by Clayton Brown. See "LICENSE" file in root for more info.

package managers

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
	// Result is a channel so that await commands can wait for the process
	// 	thread to finish it's computations.
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
		return errors.New(managerName + " manager is not created or has been deleted (occurred during request send).")
	}

	// Otherwise, send the request to the manager and return with no errors
	manager.Requests <- request
	return nil

}

// Await will wait until a request is completed and respond with the result
func (request *Request) Await(managerName string) (interface{}, error) {

	// First send the request
	err := request.Send(managerName)

	// If there is an error, respond with it
	if err != nil {
		return nil, err
	}

	// Otherwise wait for completion and return the result
	return request.Wait()

}

// Send will send a request to a specified manager
func (request *Request) SendManager(manager *Manager) {

	// Otherwise, send the request to the manager and return with no errors
	manager.Requests <- request

}

// Await will wait until a request is completed and respond with the result
func (request *Request) AwaitManager(manager *Manager) (interface{}, error) {

	// First send the request
	request.SendManager(manager)

	// Otherwise wait for completion and return the result
	return request.Wait()

}

/*
	GetData will either return data or an error depending on whether or
	not there is an error present in the data. Handy for use when you
	have nested response objects.
*/
func (response *Response) getData() (interface{}, error) {

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
		responseData, ok := data.(*Response)
		if ok {
			return responseData.getData()
		}

		return response.Data, nil
	}

	return nil, response.Error

}

// Wait is used for a job which has already been sent and the response wants to be waited on.
// 	Once the response is given, the data will be parsed and returned.
func (request *Request) Wait() (interface{}, error) {

	// Just wait for data to be put in the response
	response := <-request.Response
	return response.getData()

}
