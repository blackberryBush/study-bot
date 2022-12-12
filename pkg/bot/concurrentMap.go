package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

type Chattable struct {
	data   tgbotapi.Chattable
	option AnswerOptions
}

type AnswerOptions struct {
	taskID  int
	correct int
}

func NewChattable(data tgbotapi.Chattable, options ...int) *Chattable {
	switch {
	case len(options) == 1:
		return &Chattable{data: data, option: AnswerOptions{0, options[0]}}
	case len(options) >= 2:
		return &Chattable{data: data, option: AnswerOptions{options[0], options[1]}}
	default:
		return &Chattable{data: data}
	}
}

type ItemToSend struct {
	queue int
	data  chan Chattable
}

func NewItemToSend() *ItemToSend {
	return &ItemToSend{
		queue: 0,
		data:  make(chan Chattable, 1),
	}
}

type KitToSend struct {
	mx sync.RWMutex
	m  map[int64]ItemToSend
}

func NewKitToSend() *KitToSend {
	return &KitToSend{
		mx: sync.RWMutex{},
		m:  make(map[int64]ItemToSend),
	}
}

func (c *KitToSend) Load(key int64) (ItemToSend, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *KitToSend) Store(key int64, value ItemToSend) {
	c.mx.Lock()
	c.m[key] = value
	c.mx.Unlock()
}

func (c *KitToSend) StoreData(key int64, value chan Chattable) {
	c.mx.Lock()
	temp := c.m[key]
	temp.data = value
	c.m[key] = temp
	c.mx.Unlock()
}

func (c *KitToSend) Delete(key int64) {
	c.mx.Lock()
	delete(c.m, key)
	c.mx.Unlock()
}

func (c *KitToSend) Range(f func(key int64, value ItemToSend) bool) {
	tmp := make(map[int64]ItemToSend)
	c.mx.RLock()
	for i, v := range c.m {
		tmp[i] = v
	}
	c.mx.RUnlock()
	for i, v := range tmp {
		if !f(i, v) {
			break
		}
	}
}

func (c *KitToSend) QueueInc(key int64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if item, ok := c.m[key]; ok {
		item.queue++
		c.m[key] = item
	}
}

func (c *KitToSend) QueueDec(key int64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if item, ok := c.m[key]; ok {
		item.queue--
		c.m[key] = item
	}
}
