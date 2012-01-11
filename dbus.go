package dbus

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type StandardBus int

const (
	SessionBus StandardBus = iota
	SystemBus
)

const dbusXMLIntro = `
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node>
  <interface name="org.freedesktop.DBus.Introspectable">
    <method name="Introspect">
      <arg name="data" direction="out" type="s"/>
    </method>
  </interface>
  <interface name="org.freedesktop.DBus">
    <method name="RequestName">
      <arg direction="in" type="s"/>
      <arg direction="in" type="u"/>
      <arg direction="out" type="u"/>
    </method>
    <method name="ReleaseName">
      <arg direction="in" type="s"/>
      <arg direction="out" type="u"/>
    </method>
    <method name="StartServiceByName">
      <arg direction="in" type="s"/>
      <arg direction="in" type="u"/>
      <arg direction="out" type="u"/>
    </method>
    <method name="Hello">
      <arg direction="out" type="s"/>
    </method>
    <method name="NameHasOwner">
      <arg direction="in" type="s"/>
      <arg direction="out" type="b"/>
    </method>
    <method name="ListNames">
      <arg direction="out" type="as"/>
    </method>
    <method name="ListActivatableNames">
      <arg direction="out" type="as"/>
    </method>
    <method name="AddMatch">
      <arg direction="in" type="s"/>
    </method>
    <method name="RemoveMatch">
      <arg direction="in" type="s"/>
    </method>
    <method name="GetNameOwner">
      <arg direction="in" type="s"/>
      <arg direction="out" type="s"/>
    </method>
    <method name="ListQueuedOwners">
      <arg direction="in" type="s"/>
      <arg direction="out" type="as"/>
    </method>
    <method name="GetConnectionUnixUser">
      <arg direction="in" type="s"/>
      <arg direction="out" type="u"/>
    </method>
    <method name="GetConnectionUnixProcessID">
      <arg direction="in" type="s"/>
      <arg direction="out" type="u"/>
    </method>
    <method name="GetConnectionSELinuxSecurityContext">
      <arg direction="in" type="s"/>
      <arg direction="out" type="ay"/>
    </method>
    <method name="ReloadConfig">
    </method>
    <signal name="NameOwnerChanged">
      <arg type="s"/>
      <arg type="s"/>
      <arg type="s"/>
    </signal>
    <signal name="NameLost">
      <arg type="s"/>
    </signal>
    <signal name="NameAcquired">
      <arg type="s"/>
    </signal>
  </interface>
</node>`

type signalHandler struct {
	mr   MatchRule
	proc func(*Message)
}

// A connection to a single D-Bus bus. See StandardBus.
type Connection struct {
	addressMap        map[string]string
	uniqName          string
	methodCallReplies map[uint32](func(msg *Message))
	signalMatchRules  []signalHandler
	conn              net.Conn
	buffer            *bytes.Buffer
	proxy             Interface
}

// An Object type is analogous to the reflect.Value type for D-Bus remote objects.
type Object struct {
	dest  string
	path  string
	intro Introspect
}

// The Interface type is analogous to the reflect.Value type for D-Bus
// interfaces. Methods/Signals accessed through an Interface can be passed to
// Call/Emit. See also, InterfaceData.
type Interface interface {
	// The name of the interface.
	GetName() string
	// The object the interface belongs to.
	Object() *Object
	// Access interface methods. Like InterfaceData methods but returns a Method,
	// not MethodData.
	NumMethod() int
	Method(i int) Method
	MethodByName(string) Method
	// Access interface signals. Like InterfaceData methods but returns a Signal,
	// not SignalData.
	NumSignal() int
	Signal(i int) Signal
	SignalByName(string) Signal
	// Access underlying InterfaceData, which is analogous to a reflect.Type.
	Introspect() InterfaceData
}

type _interface struct {
	obj   *Object
	name  string
	intro InterfaceData
}

type Method interface {
	Introspect() MethodIntrospect
	Interface() Interface
}

type method struct {
	iface Interface
	MethodIntrospect
}

func (m *method) Interface() Interface         { return m.iface }
func (m *method) Introspect() MethodIntrospect { return m.MethodIntrospect }

type Signal interface {
	Introspect() SignalIntrospect
	Interface() Interface
}

type signal struct {
	iface Interface
	SignalIntrospect
}

func (s *signal) Interface() Interface         { return s.iface }
func (s *signal) Introspect() SignalIntrospect { return s.SignalIntrospect }

func (iface *_interface) GetName() string           { return iface.name }
func (iface *_interface) Object() *Object           { return iface.obj }
func (iface *_interface) Introspect() InterfaceData { return iface.intro }

func (iface *_interface) NumMethod() int      { return iface.intro.NumMethod() }
func (iface *_interface) Method(i int) Method { return &method{iface, iface.intro.Method(i)} }
func (iface *_interface) MethodByName(name string) Method {
	data := iface.intro.MethodByName(name)
	if nil == data {
		panic("invalid method")
	}
	return &method{iface, data}
}

func (iface *_interface) NumSignal() int      { return iface.intro.NumSignal() }
func (iface *_interface) Signal(i int) Signal { return &signal{iface, iface.intro.Signal(i)} }
func (iface *_interface) SignalByName(name string) Signal {
	data := iface.intro.SignalByName(name)
	if nil == data {
		panic("invalid signal")
	}
	return &signal{iface, data}
}

func Connect(busType StandardBus) (*Connection, error) {
	var address string

	switch busType {
	case SessionBus:
		address = os.Getenv("DBUS_SESSION_BUS_ADDRESS")

	case SystemBus:
		if address = os.Getenv("DBUS_SYSTEM_BUS_ADDRESS"); len(address) == 0 {
			address = "unix:path=/var/run/dbus/system_bus_socket"
		}

	default:
		return nil, errors.New("Unknown bus")
	}

	if len(address) == 0 {
		return nil, errors.New("Unknown bus address")
	}
	transport := address[:strings.Index(address, ":")]

	bus := new(Connection)
	bus.addressMap = make(map[string]string)
	for _, pair := range strings.Split(address[len(transport)+1:], ",") {
		pair := strings.Split(pair, "=")
		bus.addressMap[pair[0]] = pair[1]
	}

	var ok bool
	if address, ok = bus.addressMap["path"]; ok {
	} else if address, ok = bus.addressMap["abstract"]; ok {
		address = "@" + address
	} else {
		return nil, errors.New("Unknown address key")
	}

	var err error
	if bus.conn, err = net.Dial(transport, address); err != nil {
		return nil, err
	}

	return bus, nil
}

func (p *Connection) Initialize() error {
	p.methodCallReplies = make(map[uint32]func(*Message))
	p.signalMatchRules = make([]signalHandler, 0)
	p.proxy = p._GetProxy()
	p.buffer = bytes.NewBuffer([]byte{})
	err := p._Auth()
	if err != nil {
		return err
	}
	go p._RunLoop()
	p._SendHello()
	return nil
}

func (p *Connection) _Auth() error {
	auth := new(authState)
	auth.AddAuthenticator(new(AuthExternal))

	return auth.Authenticate(p.conn)
}

func (p *Connection) _MessageReceiver(msgChan chan *Message) {
	for {
		msg, e := p._PopMessage()
		if e == nil {
			msgChan <- msg
			continue // might be another msg in p.buffer
		}
		p._UpdateBuffer()
	}
}

func (p *Connection) _RunLoop() {
	msgChan := make(chan *Message)
	go p._MessageReceiver(msgChan)
	for {
		select {
		case msg := <-msgChan:
			p._MessageDispatch(msg)
		}
	}
}

func (p *Connection) _MessageDispatch(msg *Message) {
	if msg == nil {
		return
	}

	switch msg.Type {
	case METHOD_RETURN:
		rs := msg.replySerial
		if replyFunc, ok := p.methodCallReplies[rs]; ok {
			replyFunc(msg)
			delete(p.methodCallReplies, rs)
		}
	case SIGNAL:
		for _, handler := range p.signalMatchRules {
			if handler.mr._Match(msg) {
				handler.proc(msg)
			}
		}
	case ERROR:
		fmt.Println("ERROR")
	}
}

func (p *Connection) _PopMessage() (*Message, error) {
	msg, n, err := _Unmarshal(p.buffer.Bytes())
	if err != nil {
		return nil, err
	}
	p.buffer.Read(make([]byte, n)) // remove first n bytes
	return msg, nil
}

func (p *Connection) _UpdateBuffer() error {
	//	_, e := p.buffer.ReadFrom(p.conn);
	buff := make([]byte, 4096)
	n, e := p.conn.Read(buff)
	p.buffer.Write(buff[0:n])
	return e
}

func (p *Connection) _SendSync(msg *Message, callback func(*Message)) error {
	seri := uint32(msg.serial)
	recvChan := make(chan int)
	p.methodCallReplies[seri] = func(rmsg *Message) {
		callback(rmsg)
		recvChan <- 0
	}

	buff, _ := msg._Marshal()
	p.conn.Write(buff)
	<-recvChan // synchronize
	return nil
}

func (p *Connection) _SendHello() { p.Call(p.proxy.MethodByName("Hello")) }

func (p *Connection) _GetIntrospect(dest string, path string) Introspect {
	msg := NewMessage()
	msg.Type = METHOD_CALL
	msg.Path = path
	msg.Dest = dest
	msg.Iface = "org.freedesktop.DBus.Introspectable"
	msg.Member = "Introspect"

	var intro Introspect

	p._SendSync(msg, func(reply *Message) {
		if v, ok := reply.Params[0].(string); ok {
			if i, err := NewIntrospect(v); err == nil {
				intro = i
			}
		}
	})

	return intro
}

// Get the D-Bus destination for the object.
func (obj *Object) GetDestination() string { return obj.dest }

// Get the destination-relative path of the object.
func (obj *Object) GetPath() string { return obj.path }

// Get the ?full? Introspect name of the object.
func (obj *Object) GetName() string { return obj.intro.GetName() }

// The number of interfaces implemented by the object.
func (obj *Object) NumInterface() int { return obj.intro.NumInterface() }

// Retrieve an interface by index.
func (obj *Object) Interface(i int) Interface {
	if obj == nil || obj.intro == nil {
		return nil
	}
	data := obj.intro.Interface(i)
	if nil == data {
		return nil
	}
	name := data.GetName()
	return &_interface{obj, name, data}
}

// Retrieve an interface by name.
func (obj *Object) InterfaceByName(name string) Interface {
	if obj == nil || obj.intro == nil {
		return nil
	}
	data := obj.intro.GetInterfaceData(name)
	if nil == data {
		return nil
	}
	return &_interface{obj, name, data}
}

// The Introspect type describing the object.
func (obj *Object) Introspect() Introspect { return obj.intro }

func (p *Connection) _GetProxy() Interface {
	obj := new(Object)
	obj.path = "/org/freedesktop/DBus"
	obj.dest = "org.freedesktop.DBus"
	obj.intro, _ = NewIntrospect(dbusXMLIntro)

	iface := new(_interface)
	iface.obj = obj
	iface.name = "org.freedesktop.DBus"
	iface.intro = obj.intro.GetInterfaceData("org.freedesktop.DBus")

	return iface
}

// Call a method with the given arguments.
func (p *Connection) Call(method Method, args ...interface{}) ([]interface{}, error) {
	iface, data := method.Interface(), method.Introspect()
	msg := NewMessage()

	obj := iface.Object()
	msg.Type = METHOD_CALL
	msg.Path = obj.path
	msg.Iface = iface.GetName()
	msg.Dest = obj.dest
	msg.Member = data.GetName()
	msg.Sig = data.GetInSignature()
	if len(args) > 0 {
		msg.Params = args[:]
	}

	var ret []interface{}
	p._SendSync(msg, func(reply *Message) {
		ret = reply.Params
	})

	return ret, nil
}

// Emit a signal with the given arguments.
func (p *Connection) Emit(signal Signal, args ...interface{}) error {
	iface, data := signal.Interface(), signal.Introspect()
	msg := NewMessage()

	obj := iface.Object()
	msg.Type = SIGNAL
	msg.Path = obj.path
	msg.Iface = iface.GetName()
	msg.Dest = obj.dest
	msg.Member = data.GetName()
	msg.Sig = data.GetSignature()
	msg.Params = args[:]

	buff, _ := msg._Marshal()
	_, err := p.conn.Write(buff)

	return err
}

// Retrieve a specified object.
func (p *Connection) Object(dest string, path string) *Object {

	obj := new(Object)
	obj.path = path
	obj.dest = dest
	obj.intro = p._GetIntrospect(dest, path)

	return obj
}

// Handle received signals.
func (p *Connection) Handle(rule *MatchRule, handler func(*Message)) {
	p.signalMatchRules = append(p.signalMatchRules, signalHandler{*rule, handler})
	p.Call(p.proxy.MethodByName("AddMatch"), rule._ToString())
}
