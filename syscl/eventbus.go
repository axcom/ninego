package syscl

/*
go get -u github.com/lockp111/go-eventbus
在多个模块间，为事件提供传输通道的组件或者说库，称之为事件总线，也可以称之为消息总线。
事件总线涉及三个要素：事件源、订阅者、通道
1、当一个模块处理逻辑时会产生事件，我们称这个模块为事件源。
2、有的模块等待事件的产生，然后去执行特定的任务，我们称这个模块为订阅者。
3、维护这些事件源和订阅者的关系的组件我们称之为通道，也就是总线。

New()
Create a new bus struct reference

bus := classes.NewBus()
On(topic string, e ...Event)

Subscribe event

type ready struct{
}

func (e ready) Dispatch(msg interface{}){
    fmt.Println("I am ready!")
}

bus.On("ready", &ready{})

You can also subscribe multiple events for example:

type run struct{
}

func (e run) Dispatch(msg interface{}){
    fmt.Println("I am run!")
}

bus.On("ready", &ready{}, &ready{}).On("run", &run{})
Off(topic string, e ...Event)
Unsubscribe event

e := &ready{}
bus.On("ready", e)
bus.Off("ready", e)
You can also unsubscribe multiple events for example:

e1 := &ready{}
e2 := &ready{}
bus.On("ready", e1, e2)
bus.Off("ready", e1, e2)
You can unsubscribe all events for example:

bus.On("ready", &ready{}, &ready{})
bus.Off("ready")
You can unsubscribe all topics for example:

bus.On("ready", &ready{}, &ready{})
bus.On("run", &run{})
bus.Off(ALL)
Trigger(topic string, msg ...interface{})
Dispatch events

bus.Trigger("ready")
You can also dispatch multiple events for example:

bus.Trigger("ready", &struct{"1"}, &struct{"2"})
You can also dispatch all events for example:

bus.Trigger(ALL, &struct{"1"})
*/

import (
	"reflect"
	"sync"
)

// ALL - The key use to listen all the topics
const ALL = "*"

// Event interface
type Event interface {
	Dispatch(data interface{})
}

// event struct
type event struct {
	Event
	topic     string
	tag       reflect.Value
	isUnique  bool
	hasCalled bool
}

func newEvent(e Event, topic string, isUnique bool) *event {
	return &event{e, topic, reflect.ValueOf(e), isUnique, false}
}

// Bus struct
type Bus struct {
	mux    sync.Mutex
	events map[string][]*event
}

// New - return a new Bus object
func NewBus() *Bus {
	return &Bus{
		events: make(map[string][]*event),
	}
}

// On - register topic event and return error
func (b *Bus) On(topic string, e ...Event) *Bus {
	b.addEvents(topic, false, e)
	return b
}

// Once - register once event and return error
func (b *Bus) Once(topic string, e ...Event) *Bus {
	b.addEvents(topic, true, e)
	return b
}

// Off - remove topic event
func (b *Bus) Off(topic string, e ...Event) *Bus {
	b.removeEvents(topic, e)
	return b
}

// Clean - clear all events
func (b *Bus) Clean() *Bus {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.events = make(map[string][]*event)
	return b
}

// Trigger - dispatch event
func (b *Bus) Trigger(topic string, msg ...interface{}) *Bus {
	if len(msg) != 0 {
		for _, d := range msg {
			b.dispatch(topic, d)
		}
	} else {
		b.dispatch(topic, nil)
	}

	return b
}

func (b *Bus) addEvents(topic string, isUnique bool, es []Event) {
	if len(es) == 0 {
		return
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	for _, e := range es {
		b.events[topic] = append(b.events[topic], newEvent(e, topic, isUnique))
	}
}

func (b *Bus) removeEvents(topic string, es []Event) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if len(es) == 0 {
		delete(b.events, topic)
		return
	}

	events := b.events[topic]
	if len(events) == 0 {
		return
	}

	for _, e := range es {
		tag := reflect.ValueOf(e)
		for i := 0; i < len(events); i++ {
			if events[i].tag == tag {
				events = append(events[:i], events[i+1:]...)
				i--
			}
		}
	}

	if len(events) == 0 {
		delete(b.events, topic)
		return
	}

	b.events[topic] = events
}

func (b *Bus) getEvents(topic string) []*event {
	b.mux.Lock()
	defer b.mux.Unlock()

	events := make([]*event, 0, len(b.events[topic])+len(b.events[ALL]))
	for _, e := range b.events[topic] {
		if e.isUnique {
			if e.hasCalled {
				continue
			}
			e.hasCalled = true
		}
		events = append(events, e)
	}

	if topic != ALL {
		for _, e := range b.events[ALL] {
			if e.isUnique {
				if e.hasCalled {
					continue
				}
				e.hasCalled = true
			}
			events = append(events, e)
		}
	}
	return events
}

func (b *Bus) dispatch(topic string, data interface{}) {
	var (
		events  = b.getEvents(topic)
		removes = make(map[string][]Event)
	)

	for _, e := range events {
		e.Dispatch(data)
		if e.isUnique && e.hasCalled {
			removes[e.topic] = append(removes[e.topic], e.Event)
		}
	}

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}
