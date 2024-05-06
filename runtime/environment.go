package runtime

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewEnvironment creates a new environment for the application
// For setting custom services that aren't set by default, use the EnvOption
// Note: Depinject always provide an environment with all services (mandatory and optional)
func NewEnvironment(
	kvService store.KVStoreService,
	logger log.Logger,
	opts ...EnvOption,
) appmodule.Environment {
	env := appmodule.Environment{
		Logger:             logger,
		EventService:       EventService{},
		HeaderService:      HeaderService{},
		BranchService:      BranchService{},
		GasService:         GasService{},
		TransactionService: TransactionService{},
		KVStoreService:     kvService,
		RouterService:      NewRouterService(kvService, &failingQueryRouter{}, &failingMsgRouter{}, map[string]bool{}),
		MemStoreService:    failingMemStore{},
	}

	for _, opt := range opts {
		opt(&env)
	}

	return env
}

type EnvOption func(*appmodule.Environment)

func EnvWithRouterService(
	queryServiceRouter *baseapp.GRPCQueryRouter,
	msgServiceRouter *baseapp.MsgServiceRouter,
	allowList map[string]bool,
) EnvOption {
	return func(env *appmodule.Environment) {
		env.RouterService = NewRouterService(env.KVStoreService, queryServiceRouter, msgServiceRouter, allowList)
	}
}

func EnvWithMemStoreService(memStoreService store.MemoryStoreService) EnvOption {
	return func(env *appmodule.Environment) {
		env.MemStoreService = memStoreService
	}
}

type failingMsgRouter struct {
	*baseapp.MsgServiceRouter
}

func (failingMsgRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	panic("message router not set")
}

func (failingMsgRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	panic("message router not set")
}

func (failingMsgRouter) ResponseNameByMsgName(msgName string) string {
	panic("message router not set")
}

func (failingMsgRouter) HybridHandlerByMsgName(msgName string) func(ctx context.Context, req, resp protoiface.MessageV1) error {
	panic("message router not set")
}

type failingQueryRouter struct {
	*baseapp.GRPCQueryRouter
}

func (failingQueryRouter) HybridHandlerByRequestName(name string) []func(ctx context.Context, req protoiface.MessageV1, resp protoiface.MessageV1) error {
	panic("query router not set")
}

func (failingQueryRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	panic("query router not set")
}

func (failingQueryRouter) ResponseNameByRequestName(requestName string) string {
	panic("query router not set")
}

func (failingQueryRouter) Route(path string) baseapp.GRPCQueryHandler {
	panic("query router not set")
}

func (failingQueryRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	panic("query router not set")
}

type failingMemStore struct {
	store.MemoryStoreService
}

func (failingMemStore) OpenMemoryStore(context.Context) store.KVStore {
	panic("memory store not set")
}
