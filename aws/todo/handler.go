package main

import (
	"github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/yunspace/serverless-golang/aws/event/apigateway"
	"github.com/yunspace/serverless-golang/examples/todo"
	"github.com/yunspace/serverless-golang/api"
)

var mockTodoService api.CRUDAPI

func init() {
	mockTodoService = todo.NewMockTodoService()
}

/// Create
func Create(evt *apigatewayproxyevt.Event, ctx *runtime.Context) (interface{}, error) {
	return apigateway.HandleCreate(evt, todo.TodoAPIGAtewayEventUnmarshaler, mockTodoService.Create)
}

func Get(evt *apigatewayproxyevt.Event, _ *runtime.Context) (interface{}, error) {
	return apigateway.HandleGet(evt, mockTodoService.Get)
}

func List(evt *apigatewayproxyevt.Event, _ *runtime.Context) (interface{}, error) {
	return apigateway.HandleList(evt, mockTodoService.List)
}

func Update(evt *apigatewayproxyevt.Event, _ *runtime.Context) (interface{}, error) {
	return apigateway.HandleUpdate(evt, todo.TodoAPIGAtewayEventUnmarshaler, mockTodoService.Update)
}

func Delete(evt *apigatewayproxyevt.Event, _ *runtime.Context) (interface{}, error) {
	return apigateway.HandleDelete(evt, mockTodoService.Delete)
}