package systemd

import (
	"reflect"

	"launchpad.net/go-dbus/v1"
)

const (
	busName           string          = "org.freedesktop.systemd1"
	busPath           dbus.ObjectPath = "/org/freedesktop/systemd1"
	managerInterface  string          = "org.freedesktop.systemd1.Manager"
	unitInterface     string          = "org.freedesktop.systemd1.Unit"
	propertyInterface string          = "org.freedesktop.DBus.Properties"
)

type SystemD struct {
	conn *dbus.Connection
}

type Unit struct {
	objectPath dbus.ObjectPath
	conn       *dbus.Connection
}

func New(conn *dbus.Connection) *SystemD {
	return &SystemD{
		conn: conn,
	}
}

func (sd *SystemD) Unit(service string) (*Unit, error) {
	mgrObject := sd.conn.Object(busName, busPath)

	reply, err := mgrObject.Call(managerInterface, "LoadUnit", service)
	if err != nil {
		return nil, err
	} else if reply.Type == dbus.TypeError {
		return nil, reply.AsError()
	}

	var unitPath dbus.ObjectPath
	if err := reply.Args(&unitPath); err != nil {
		return nil, err
	}

	unit := &Unit{
		conn:       sd.conn,
		objectPath: unitPath,
	}

	return unit, nil
}

func (sd *SystemD) Stop(service, mode string) error {
	return sd.startstop("StopUnit", service, mode)
}

func (sd *SystemD) Start(service, mode string) error {
	return sd.startstop("StartUnit", service, mode)
}

func (sd *SystemD) startstop(action string, args ...interface{}) error {
	mgrObject := sd.conn.Object(busName, busPath)

	reply, err := mgrObject.Call(managerInterface, action, args...)
	if err != nil {
		return err
	} else if reply.Type == dbus.TypeError {
		return reply.AsError()
	}

	return nil
}

func (u *Unit) Status() (string, error) {
	return u.getStringProperty("ActiveState")
}

func (u *Unit) Description() (string, error) {
	return u.getStringProperty("Description")
}

func (u *Unit) getStringProperty(propertyName string) (string, error) {
	propertyValue, err := u.getProperty(propertyName)
	if err != nil {
		return "", err
	}

	return reflect.ValueOf(propertyValue.Value).String(), nil
}

func (u *Unit) getProperty(propertyName string) (propertyValue dbus.Variant, err error) {
	unitObj := u.conn.Object(busName, u.objectPath)

	reply, err := unitObj.Call(propertyInterface, "Get", unitInterface, propertyName)
	if err != nil {
		return dbus.Variant{}, err
	} else if reply.Type == dbus.TypeError {
		return dbus.Variant{}, reply.AsError()
	}

	if err := reply.Args(&propertyValue); err != nil {
		return dbus.Variant{}, err
	}

	return propertyValue, nil
}
