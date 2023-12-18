package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
// если хочется, то для красоты можно разнести логику по разным файликам

type ACL struct {
	Rights map[string][]string
	Mu     *sync.RWMutex
}

func NewACL(acl string) (*ACL, error) {
	var mp map[string][]string
	err := json.Unmarshal([]byte(acl), &mp)

	for consumer, rights := range mp {
		for i := range rights {
			rights[i] = strings.TrimSuffix(rights[i], "/*")
		}
		mp[consumer] = rights
	}

	if err != nil {
		return nil, err
	}
	return &ACL{Rights: mp, Mu: &sync.RWMutex{}}, nil
}

func (a *ACL) CheckPermission(consumer, method string) bool {
	var rights []string
	var ok bool
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	if rights, ok = a.Rights[consumer]; !ok {
		return false
	}
	for _, right := range rights {
		if strings.HasPrefix(method, right) {
			return true
		}
	}
	return false
}

func EventFromContext(ctx context.Context) (*Event, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("couldn't metadat from context")
	}

	var val []string
	if val, ok = md["consumer"]; !ok {
		return nil, fmt.Errorf("couldn't get consumer from context")
	}
	if (len(val)) == 0 {
		return nil, fmt.Errorf("empty consumer")
	}

	methodName, ok := grpc.Method(ctx)
	if !ok {
		return nil, fmt.Errorf("no method")
	}

	clientPeer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer")
	}

	event := &Event{
		Method:    methodName,
		Consumer:  val[0],
		Host:      clientPeer.Addr.String(),
		Timestamp: time.Now().Unix(),
	}

	return event, nil
}

func ACLInterception(ctx context.Context, acl *ACL) (bool, error) {
	event, err := EventFromContext(ctx)
	if err != nil {
		return false, err
	}

	res := acl.CheckPermission(event.Consumer, event.Method)

	return res, nil
}

func ACLUnaryInterceptor(acl *ACL) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		res, err := ACLInterception(ctx, acl)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if !res {
			return nil, status.Error(codes.Unauthenticated, "no rights")
		}
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func ACLStreamInterceptor(acl *ACL) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		res, err := ACLInterception(stream.Context(), acl)
		if err != nil {
			return err
		}
		if !res {
			return status.Error(codes.Unauthenticated, "no rights")
		}
		err = handler(srv, stream)
		return err
	}
}

func AdminUnaryInterceptor(a *AdminServerStruct) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		res, err := EventFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		a.notifyObservers(res)
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func AdminStreamInterceptor(a *AdminServerStruct) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		res, err := EventFromContext(stream.Context())
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		a.notifyObservers(res)

		err = handler(srv, stream)
		return err
	}
}

func StartMyMicroservice(ctx context.Context, addr string, acl string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	aclManager, err := NewACL(acl)
	if err != nil {
		lis.Close()
		return err
	}

	admin := NewAdminServerStruct()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(ACLUnaryInterceptor(aclManager), AdminUnaryInterceptor(admin)),
		grpc.ChainStreamInterceptor(ACLStreamInterceptor(aclManager), AdminStreamInterceptor(admin)),
	)

	RegisterAdminServer(server, admin)
	RegisterBizServer(server, NewBizLogic())

	go func() {
		<-ctx.Done()
		fmt.Println("Stopping")
		server.GracefulStop()
		lis.Close()
		fmt.Println("Stopped")
	}()

	go func() {
		fmt.Println("starting server at", addr)
		if err := server.Serve(lis); err != nil {
			log.Fatalln("failed to serve:", err)
		}
	}()

	return nil
}

type AdminServerStruct struct {
	UnimplementedAdminServer
	Observers map[string]chan interface{}
	Mu        *sync.Mutex
}

func NewAdminServerStruct() *AdminServerStruct {
	return &AdminServerStruct{
		Observers: make(map[string]chan interface{}, 0),
		Mu:        &sync.Mutex{},
	}
}

func (a *AdminServerStruct) addNewObserver(ch chan interface{}) string {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	stamp := uuid.New().String()
	a.Observers[stamp] = ch
	return stamp
}

func (a *AdminServerStruct) deleteObserver(key string) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	delete(a.Observers, key)
}

func (a *AdminServerStruct) notifyObservers(event *Event) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	for _, v := range a.Observers {
		localEvent := Event{
			Consumer:  event.Consumer,
			Method:    event.Method,
			Host:      event.Host,
			Timestamp: event.Timestamp,
		}
		go func(c chan interface{}) {
			c <- localEvent
		}(v)
	}
}

func (a *AdminServerStruct) Logging(n *Nothing, log Admin_LoggingServer) error {
	logChan := make(chan interface{})
	stamp := a.addNewObserver(logChan)
	defer a.deleteObserver(stamp)

	ctx := log.Context()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("aborted")
		case val := <-logChan:
			event := Event{
				Consumer:  val.(Event).Consumer,
				Host:      val.(Event).Host,
				Method:    val.(Event).Method,
				Timestamp: val.(Event).Timestamp,
			}
			err := log.Send(&event)
			if err != nil {
				return fmt.Errorf("error in logging: %s", err.Error())
			}
		}
	}
}

func (a *AdminServerStruct) Statistics(s *StatInterval, stat Admin_StatisticsServer) error {
	statChan := make(chan interface{})
	stamp := a.addNewObserver(statChan)
	defer a.deleteObserver(stamp)

	stats := Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
		Timestamp:  time.Now().Unix(),
	}
	ticker := time.NewTicker(time.Duration(s.IntervalSeconds) * time.Second)
	ctx := stat.Context()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("aborted")
		case <-ticker.C:
			err := stat.Send(&stats)
			if err != nil {
				return fmt.Errorf("error in statistics: %s", err.Error())
			}
			stats.ByConsumer = make(map[string]uint64)
			stats.ByMethod = make(map[string]uint64)
		case val := <-statChan:
			event := Event{
				Consumer:  val.(Event).Consumer,
				Host:      val.(Event).Host,
				Method:    val.(Event).Method,
				Timestamp: val.(Event).Timestamp,
			}
			stats.ByConsumer[event.Consumer] += 1
			stats.ByMethod[event.Method] += 1
		}
	}
}

type BizLogic struct {
	UnimplementedBizServer
}

func NewBizLogic() *BizLogic {
	bzl := &BizLogic{}
	return bzl
}
func (b *BizLogic) Check(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nothing, nil
}

func (b *BizLogic) Add(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nothing, nil
}

func (b *BizLogic) Test(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nothing, nil
}
