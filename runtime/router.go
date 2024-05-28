package runtime

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	v1 "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/router"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMsgRouterService implements router.Service.
func NewMsgRouterService(msgRouter baseapp.MessageRouter) router.Service {
	return &msgRouterService{
		router: msgRouter,
	}
}

var _ router.Service = (*msgRouterService)(nil)

type msgRouterService struct {
	// TODO: eventually authenticate modules to use the message router
	router baseapp.MessageRouter
}

// CanInvoke returns an error if the given message cannot be invoked.
func (m *msgRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")

	handler := m.router.HandlerByTypeURL(typeURL)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *msgRouterService) InvokeTyped(ctx context.Context, msg, resp protoiface.MessageV1) error {
	messageName := msgTypeURL(msg)
	handler := m.router.HandlerByTypeURL(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	resp, err := handler(sdkCtx, msg)

	resp = resp

	return err
}

// InvokeUntyped execute a message and returns a response.
func (m *msgRouterService) InvokeUntyped(ctx context.Context, msg protoiface.MessageV1) (protoiface.MessageV1, error) {
	messageName := msgTypeURL(msg)
	respName := m.router.ResponseNameByMsgName(messageName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %s (%T)", messageName, msg)
	}

	// get response type
	typ := proto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	msgResp, ok := reflect.New(typ.Elem()).Interface().(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("could not create response message %s", respName)
	}

	return msgResp, m.InvokeTyped(ctx, msg, msgResp)
}

// NewQueryRouterService implements router.Service.
func NewQueryRouterService(queryRouter *baseapp.GRPCQueryRouter) router.Service {
	return &queryRouterService{
		router: queryRouter,
	}
}

var _ router.Service = (*queryRouterService)(nil)

type queryRouterService struct {
	router *baseapp.GRPCQueryRouter
}

// CanInvoke returns an error if the given request cannot be invoked.
func (m *queryRouterService) CanInvoke(ctx context.Context, typeURL string) error {
	if typeURL == "" {
		return fmt.Errorf("missing type url")
	}

	typeURL = strings.TrimPrefix(typeURL, "/")

	handlers := m.router.Route(typeURL)
	if handlers == nil {
		return fmt.Errorf("unknown request: %s", typeURL)
	}

	return nil
}

// InvokeTyped execute a message and fill-in a response.
// The response must be known and passed as a parameter.
// Use InvokeUntyped if the response type is not known.
func (m *queryRouterService) InvokeTyped(ctx context.Context, req, resp protoiface.MessageV1) error {
	reqName := msgTypeURL(req)
	handlers := m.router.Route(reqName)
	if handlers == nil {
		return fmt.Errorf("unknown request: %s", reqName)
	}

	var err error

	qreq := v1.QueryRequest{
		Data: req,
	}

	resp, err = handlers(sdk.UnwrapSDKContext(ctx), &qreq)

	return err
}

// InvokeUntyped execute a message and returns a response.
func (m *queryRouterService) InvokeUntyped(ctx context.Context, req protoiface.MessageV1) (protoiface.MessageV1, error) {
	reqName := msgTypeURL(req)
	respName := m.router.ResponseNameByRequestName(reqName)
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for request %s (%T)", reqName, req)
	}

	// get response type
	typ := proto.MessageType(respName)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", respName)
	}
	reqResp, ok := reflect.New(typ.Elem()).Interface().(protoiface.MessageV1)
	if !ok {
		return nil, fmt.Errorf("could not create response request %s", respName)
	}

	return reqResp, m.InvokeTyped(ctx, req, reqResp)
}

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg proto.Message) string {
	if m, ok := msg.(protov2.Message); ok {
		return string(m.ProtoReflect().Descriptor().FullName())
	}

	return proto.MessageName(msg)
}
