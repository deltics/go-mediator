package mediator

import (
	"context"
	"errors"
	"testing"
)

type cmdRequest struct{}
type cmdRequestHandler struct{}
type cmdRequestCompatibleHandler struct{}

type cmdRequestWithResult struct {
	Result string
}
type cmdRequestWithResultHandler struct{}
type cmdRequestByValueHandler struct{}

const cmdRequestWithResultValue = "result!"

func (*cmdRequestCompatibleHandler) Execute(context.Context, cmdRequest) error { return nil }
func (*cmdRequestHandler) Execute(context.Context, cmdRequest) error           { return nil }
func (*cmdRequestWithResultHandler) Execute(ctx context.Context, req *cmdRequestWithResult) error {
	req.Result = cmdRequestWithResultValue
	return nil
}
func (*cmdRequestByValueHandler) Execute(ctx context.Context, req cmdRequestWithResult) error {
	req.Result = cmdRequestWithResultValue
	return nil
}

type cmdRequestHandlerWithValidator struct{}

func (*cmdRequestHandlerWithValidator) Execute(context.Context, cmdRequest) error { return nil }
func (*cmdRequestHandlerWithValidator) Validate(context.Context, cmdRequest) error {
	return errors.New("validation failed")
}

func TestThatTheRegistrationInterfaceRemovesTheHandler(t *testing.T) {
	// ARRANGE
	r := RegisterCommandHandler[cmdRequest](&cmdRequestHandler{})

	// ACT
	wanted := 1
	got := len(commandHandlers)
	if wanted != got {
		t.Errorf("wanted %d handlers, got %d", wanted, got)
	}
	r.Remove()

	// ASSERT
	wanted = 0
	got = len(commandHandlers)
	if wanted != got {
		t.Errorf("wanted %d handlers, got %d", wanted, got)
	}
}

func TestThatRegisterCommandHandlerPanicsWhenHandlerIsAlreadyRegisteredForAType(t *testing.T) {
	// ARRANGE (and ASSERT, since we're testing for a panic() :) )

	// Setup the panic test (deferred ASSERT)
	defer func() {
		if r := recover(); r == nil {
			t.Error("did not panic")
		}
	}()

	// Register a handler and remove it when done
	r := RegisterCommandHandler[cmdRequest](&cmdRequestHandler{})
	defer r.Remove()

	// ACT - attempt to register another handler for the same request type
	RegisterCommandHandler[cmdRequest](&cmdRequestCompatibleHandler{})

	// ASSERT (deferred)
}

func TestThatPerformReturnsExpectedErrorWhenRequestHandlerIsNotRegistered(t *testing.T) {
	// ARRANGE
	// no-op

	// ACT
	err := Perform(context.Background(), cmdRequest{})

	// ASSERT
	if _, ok := err.(*ErrNoHandler); !ok {
		t.Errorf("wanted *mediator.ErrNoHandler, got %T", err)
	}
}

func TestThatResultsCanBeReturnedViaFieldsInAByRefRequestType(t *testing.T) {
	// ARRANGE
	reg := RegisterCommandHandler[*cmdRequestWithResult](&cmdRequestWithResultHandler{})
	defer reg.Remove()

	// ACT
	request := &cmdRequestWithResult{}
	Perform(context.Background(), request)

	// ASSERT
	wanted := cmdRequestWithResultValue
	got := request.Result
	if request.Result != wanted {
		t.Errorf("wanted %q in request.Result, got %q", wanted, got)
	}
}

func TestThatResultsCannotBeReturnedViaFieldsInAByValueRequestType(t *testing.T) {
	// ARRANGE
	reg := RegisterCommandHandler[cmdRequestWithResult](&cmdRequestByValueHandler{})
	defer reg.Remove()

	// ACT
	request := cmdRequestWithResult{}
	Perform(context.Background(), request)

	// ASSERT
	wanted := ""
	got := request.Result
	if request.Result != wanted {
		t.Errorf("wanted %q in request.Result, got %q", wanted, got)
	}
}

func TestThatCommandValidatorErrorIsReturnedAsABadRequest(t *testing.T) {
	// ARRANGE
	reg := RegisterCommandHandler[cmdRequest](&cmdRequestHandlerWithValidator{})
	defer reg.Remove()

	// ACT
	err := Perform(context.Background(), cmdRequest{})

	// ASSERT
	if _, ok := err.(*ErrBadRequest); !ok {
		t.Errorf("wanted %T, got %T (%q)", new(ErrBadRequest), err, err)
	}
}
