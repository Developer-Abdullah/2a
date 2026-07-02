package sse
import "sync"

type Broker struct { clients map[chan string]bool; mu sync.RWMutex }
func NewBroker() *Broker { return &Broker{ clients: make(map[chan string]bool) } }
func (b *Broker) AddClient(client chan string) { b.mu.Lock(); defer b.mu.Unlock(); b.clients[client] = true }
func (b *Broker) RemoveClient(client chan string) { b.mu.Lock(); defer b.mu.Unlock(); delete(b.clients, client); close(client) }
func (b *Broker) Broadcast(msg string) {
b.mu.RLock()
defer b.mu.RUnlock()
for client := range b.clients {
select { case client <- msg: default: }
}
}
