Documentation
=============

**This link is not accurate** Look at the API on [GoPkgDoc](http://gopkgdoc.appspot.com/pkg/github.com/norisatir/go-dbus).

The go-dbus provides object oriented D-Bus bindings. The method API is similar
to package "reflect".

Installation
============

    goinstall github.com/norisatir/go-dbus

Usage
=====

Interfaces
----------

Interfaces are obtained with

    _iface := conn.Object(dest, path).InterfaceByName(iname)

They can also be iterated

    for obj, i := conn.Object(dest, path), 0; i < obj.NumInterface(); i++ {
        iface := obj.Interface(i)
        //...
    }

Methods
-------

Methods are obtained with

    meth := conn.Object(dest, path).InterfaceByName(iface).MethodByName(mname)

They can also be iterated

    for iface, i := conn.Object(dest, path).InterfaceByName(iname), 0; i < iface.NumMethod(); i++ {
        meth := iface.Method(i)
        //...
    }

They are called with

    out, err := conn.Call(meth)

Signals
-------

Signals are obtained with

    sig := conn.Object(dest, path).InterfaceByName(iface).SignalByName(sname)

They can also be iterated

    for obj, i := conn.Object(dest, path).Interface(iface), 0; i < obj.NumSignal(); i++ {
        sig := obj.Signal(i)
        //...
    }

They are emitted with

    err = conn.Emit(sig)

**TODO** Add signal handling usage.

Introspect
----------

Object introspection is done automatically. It provides a way to inspect the
methods and signals of an object. To retrieve the Introspect type describing an object

    intro := conn.Object(dest, path).Introspect()

Introspection can be performed on Interface and Method types as well.

    idata := conn.Object(dest, path).InterfaceByName(iname).Introspect()
    mdata := conn.Object(dest, path).InterfaceByName(iname).Introspect()

Introspection is done under the hood during `conn.Object(dest, path)`. So these
methods are 'cheap' and require no network communication.

An example
----------

```go
// Issue OSD notifications according to the Desktop Notifications Specification 1.1
//      http://people.canonical.com/~agateau/notifications-1.1/spec/index.html
// See also
//      https://wiki.ubuntu.com/NotifyOSD#org.freedesktop.Notifications.Notify
package main

import "github.com/norisatir/go-dbus"
import "log"

func main() {
    var (
        err error
        conn *dbus.Connection
        out []interface{}
    )

    // Connect to Session or System buses.
    if conn, err = dbus.Connect(dbus.SessionBus); err != nil {
        log.Fatal("Connection error:", err)
    }
    if err = conn.Initialize(); err != nil {
        log.Fatal("Initialization error:", err)
    }

	method := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications").
		InterfaceByName("org.freedesktop.Notifications").
		MethodByName("Notify")

    // Introspect objects.
	m := method.Introspect()
    log.Printf("%s in:%s out:%s", m.GetName(), m.GetInSignature(), m.GetOutSignature())

    // Call object methods.
    out, err = conn.Call(method,
		"dbus-tutorial", uint32(0), "",
        "dbus-tutorial", "You've been notified!",
		[]interface{}{}, map[string]interface{}{}, int32(-1))
    if err != nil {
        log.Fatal("Notification error:", err)
    }
    log.Print("Notification id:", out[0])
}
```
