package subscription

import (
	"sync"
	"whatsapp-ki-maya/models"
)

type Service interface {
	SetServer(server Server)
	Run()
	SendToServer() chan models.Message
	SendToClients() chan models.Message
	Subscribe() chan Client
	Unsubscribe() chan Client
}

type service struct {
	Subscribers map[string]map[Client]bool
	Server      Server
	updates     chan models.Message
	subscribe   chan Client
	unsubscribe chan Client
	mutex       sync.Mutex
}

func NewService() Service {
	return &service{
		Subscribers: make(map[string]map[Client]bool),
		updates:     make(chan models.Message, 100),
		subscribe:   make(chan Client, 100),
		unsubscribe: make(chan Client, 100),
		mutex:       sync.Mutex{},
	}
}

// SetServer Set the server (eg: whatsapp or telegram)
func (s *service) SetServer(server Server) {
	s.Server = server
	s.Server.Listen(s)
	go s.Server.Serve()
}

// SendToServer send message to the server
// this will automatically send the message to the respective chat
func (s *service) SendToServer() chan models.Message {
	return s.Server.Update() // returning update of server because there is only one server
}

// SendToClients send message to clients
// this will automatically send the message to the clients which are subscribed to the particular chat
func (s *service) SendToClients() chan models.Message {
	// we need to handle the message using the updater function which will be listening to this channel
	return s.updates
}

// Subscribe is used to subscribe clients to particular chat
// this is also handled via a go routine which listens to this channel
func (s *service) Subscribe() chan Client {
	s.mutex.Lock()
	return s.subscribe
}

func (s *service) Unsubscribe() chan Client {
	s.mutex.Lock()
	return s.unsubscribe
}

func (s *service) Run() {
	go s.subscriber()
	go s.unSubscriber()
	go s.updater()
}

// subscriber is a method which runs as goroutine to handle the subscription requests in the channel
func (s *service) subscriber() {
	for {
		c := <-s.subscribe
		if s.Subscribers[c.Subscription()] == nil {
			s.Subscribers[c.Subscription()] = make(map[Client]bool)
		}
		s.Subscribers[c.Subscription()][c] = true
		s.mutex.Unlock()
	}
}

// unSubscriber is same as subscriber just does the opposite of that
func (s *service) unSubscriber() {
	for {
		c := <-s.unsubscribe
		delete(s.Subscribers[c.Subscription()], c)
		s.mutex.Unlock()
	}
}

// updater is a go routine which handles the updates which comes to the subscription service
// it sends this update to all the subscribed clients
func (s *service) updater() {
	for {
		u := <-s.updates
		for c := range s.Subscribers[u.GetChatId()] {
			c.Update() <- u
		}
	}
}