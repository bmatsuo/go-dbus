package dbus

import (
	"bytes"
	"encoding/xml"
	"strings"
)

type annotationData struct {
	Name  string `xml:"attr"`
	Value string `xml:"attr"`
}

type argData struct {
	Name      string `xml:"attr"`
	Type      string `xml:"attr"`
	Direction string `xml:"attr"`
}

type methodData struct {
	Name       string `xml:"attr"`
	Arg        []argData
	Annotation annotationData
}

type signalData struct {
	Name string `xml:"attr"`
	Arg  []argData
}

// This is done to give InterfaceIntrospect more go-friendly interface API;
// NumMethod(), Method(int), NumSignal(), Signal(int). See package "reflect".
// The "xml" package requires fields Method and Signal to have those names.
type xmlInterfaceData struct {
	Name   string `xml:"attr"`
	Method []methodData
	Signal []signalData
}

type xmlIntrospect struct {
	Name      string `xml:"attr"`
	Interface []xmlInterfaceData
	Node      []*introspect
}

type interfaceData struct {
	Name    string
	Methods []methodData
	Signals []signalData
}

type introspect struct {
	Name       string
	Interfaces []xmlInterfaceData
	Nodes      []*introspect
}

// The Introspect type is analogous to the reflect.Type type for D-Bus objects.
// It describes the API of a D-Bus object.
type Introspect interface {
	GetName() string
	NumInterface() int
	Interface(i int) InterfaceIntrospect
	InterfaceByName(string) InterfaceIntrospect
	GetInterfaceData(name string) InterfaceIntrospect
}

// The InterfaceIntrospect type is analogous to the reflect.Type type for D-Bus
// interfaces. It provides access to name and type signature information for an
// interface's methods/signals.
type InterfaceIntrospect interface {
	// Get the interface name.
	GetName() string
	// Access the interface's method API.
	NumMethod() int
	Method(int) MethodIntrospect
	MethodByName(string) MethodIntrospect
	GetMethodData(name string) MethodIntrospect
	// Access the interface's signal API
	NumSignal() int
	Signal(int) SignalIntrospect
	SignalByName(string) SignalIntrospect
	GetSignalData(name string) SignalIntrospect
}

type MethodIntrospect interface {
	GetName() string
	GetInSignature() string
	GetOutSignature() string
}

type SignalIntrospect interface {
	GetName() string
	GetSignature() string
}

func NewIntrospect(xmlIntro string) (Introspect, error) {
	intro := new(xmlIntrospect)
	buff := bytes.NewBufferString(xmlIntro)
	err := xml.Unmarshal(buff, intro)
	if err != nil {
		return nil, err
	}

	return introspect{intro.Name, intro.Interface, intro.Node}, nil
}

func (p introspect) GetName() string   { return p.Name }
func (p introspect) NumInterface() int { return len(p.Interfaces) }
func (p introspect) Interface(i int) InterfaceIntrospect {
	iface := p.Interfaces[i]
	return interfaceData{iface.Name, iface.Method, iface.Signal}
}
func (p introspect) InterfaceByName(name string) InterfaceIntrospect {
	return p.GetInterfaceData(name)
}
func (p introspect) GetInterfaceData(name string) InterfaceIntrospect {
	for _, v := range p.Interfaces {
		if v.Name == name {
			return interfaceData{v.Name, v.Method, v.Signal} // Copy to InterfaceIntrospect.
		}
	}
	return nil
}

func (p interfaceData) NumMethod() int                { return len(p.Methods) }
func (p interfaceData) Method(i int) MethodIntrospect { return p.Methods[i] }
func (p interfaceData) MethodByName(name string) MethodIntrospect {
	return p.GetMethodData(name)
}
func (p interfaceData) GetMethodData(name string) MethodIntrospect {
	for _, v := range p.Methods {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) NumSignal() int                { return len(p.Signals) }
func (p interfaceData) Signal(i int) SignalIntrospect { return p.Signals[i] }
func (p interfaceData) SignalByName(name string) SignalIntrospect {
	return p.GetSignalData(name)
}
func (p interfaceData) GetSignalData(name string) SignalIntrospect {
	for _, v := range p.Signals {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) GetName() string { return p.Name }

func (p methodData) GetInSignature() (sig string) {
	for _, v := range p.Arg {
		if strings.ToUpper(v.Direction) == "IN" {
			sig += v.Type
		}
	}
	return
}

func (p methodData) GetOutSignature() (sig string) {
	for _, v := range p.Arg {
		if strings.ToUpper(v.Direction) == "OUT" {
			sig += v.Type
		}
	}
	return
}

func (p methodData) GetName() string { return p.Name }

func (p signalData) GetSignature() (sig string) {
	for _, v := range p.Arg {
		sig += v.Type
	}
	return
}

func (p signalData) GetName() string { return p.Name }
