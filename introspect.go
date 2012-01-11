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

// This is done to give InterfaceData more go-friendly interface API;
// NumMethod(), Method(int), NumSignal(), Signal(int). See packages
// "reflect", "flag", etc. The "xml" package requires fields Method and
// Signal to have those names.
type xmlInterfaceData struct {
	Name   string `xml:"attr"`
	Method []methodData
	Signal []signalData
}

type interfaceData struct {
	Name    string
	Methods []methodData
	Signals []signalData
}

type introspect struct {
	Name      string `xml:"attr"`
	Interface []xmlInterfaceData
	Node      []*introspect
}

type Introspect interface {
	GetInterfaceData(name string) InterfaceData
}

// The InterfaceData type is analogous to the reflect.Type type for D-Bus
// interfaces. It provides access to name and type signature information for an
// interface's methods/signals.
type InterfaceData interface {
	// Get the interface name.
	GetName() string
	// Access the interface's method API.
	NumMethod() int
	Method(int) MethodData
	MethodByName(string) MethodData
	GetMethodData(name string) MethodData
	// Access the interface's signal API
	NumSignal() int
	Signal(int) SignalData
	SignalByName(string) SignalData
	GetSignalData(name string) SignalData
}

type MethodData interface {
	GetName() string
	GetInSignature() string
	GetOutSignature() string
}

type SignalData interface {
	GetName() string
	GetSignature() string
}

func NewIntrospect(xmlIntro string) (Introspect, error) {
	intro := new(introspect)
	buff := bytes.NewBufferString(xmlIntro)
	err := xml.Unmarshal(buff, intro)
	if err != nil {
		return nil, err
	}

	return intro, nil
}

func (p introspect) GetInterfaceData(name string) InterfaceData {
	for _, v := range p.Interface {
		if v.Name == name {
			return interfaceData{v.Name, v.Method, v.Signal} // Copy into an InterfaceData type.
		}
	}
	return nil
}

func (p interfaceData) NumMethod() int          { return len(p.Methods) }
func (p interfaceData) Method(i int) MethodData { return p.Methods[i] }
func (p interfaceData) MethodByName(name string) MethodData {
	return p.GetMethodData(name)
}
func (p interfaceData) GetMethodData(name string) MethodData {
	for _, v := range p.Methods {
		if v.GetName() == name {
			return v
		}
	}
	return nil
}

func (p interfaceData) NumSignal() int          { return len(p.Signals) }
func (p interfaceData) Signal(i int) SignalData { return p.Signals[i] }
func (p interfaceData) SignalByName(name string) SignalData {
	return p.GetSignalData(name)
}
func (p interfaceData) GetSignalData(name string) SignalData {
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
