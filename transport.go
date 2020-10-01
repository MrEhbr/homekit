package homekit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/brutella/dnssd"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/db"
	"github.com/brutella/hc/event"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/hap/http"
	"github.com/brutella/hc/log"
	"github.com/brutella/hc/util"
)

type Transport struct {
	cfg        *Config
	hapContext hap.Context
	device     hap.SecuredDevice
	hapHTTP    *http.Server

	container *accessory.Container
	database  db.Database

	responder dnssd.Responder
	handle    dnssd.ServiceHandle

	emitter event.Emitter
}

func NewTransport(port int, path, pin string, acc ...*accessory.Accessory) (*Transport, error) {
	if len(acc) == 0 {
		return nil, errors.New("no accessories")
	}

	if len(acc) > 1 && acc[0].Type != accessory.TypeBridge {
		return nil, errors.New("multiple accessories passed and first not bridge")
	}

	if path == "" {
		path = fmt.Sprintf("./%s_storage", acc[0].Info.Name.GetValue())
	}

	database, err := NewDatabase(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	cfg := NewConfig(acc[0].Info.Name.GetValue(), database)
	_ = cfg.Load()

	if port > 0 {
		cfg.ServePort = port
	}

	if pin != "" {
		cfg.Pin = pin
	}

	hapPin, err := hc.NewPin(cfg.Pin)
	if err != nil {
		return nil, err
	}

	device, err := hap.NewSecuredDevice(cfg.ID, hapPin, database)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure device: %w", err)
	}

	container := accessory.NewContainer()
	for _, v := range acc {
		if err := container.AddAccessory(v); err != nil {
			return nil, fmt.Errorf("failed to add accessory: %w", err)
		}
	}

	cfg.CategoryID = uint8(container.AccessoryType())

	responder, err := dnssd.NewResponder()
	if err != nil {
		return nil, fmt.Errorf("failed to create responder: %w", err)
	}

	service, err := dnssd.NewService(dnssd.Config{
		Name:   util.RemoveAccentsFromString(strings.ReplaceAll(cfg.Name, " ", "_")),
		Type:   "_hap._tcp",
		Domain: "local",
		IPs:    []net.IP{},
		Text:   cfg.MDNSRecords(),
		Port:   cfg.ServePort,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create dns service: %w", err)
	}

	handler, err := responder.Add(service)
	if err != nil {
		return nil, fmt.Errorf("failed to add service to dns responder: %w", err)
	}

	transport := &Transport{
		cfg:        cfg,
		hapContext: hap.NewContextForSecuredDevice(device),
		device:     device,
		container:  container,
		database:   database,
		emitter:    event.NewEmitter(),
		responder:  responder,
		handle:     handler,
	}

	transport.hapHTTP = http.NewServer(http.Config{
		Port:      ":" + strconv.Itoa(cfg.ServePort),
		Context:   transport.hapContext,
		Database:  transport.database,
		Container: transport.container,
		Device:    transport.device,
		Mutex:     &sync.Mutex{},
		Emitter:   transport.emitter,
	})

	cfg.Discoverable = !transport.IsPaired()

	cfg.updateConfigHash(transport.container.ContentHash())

	if err := cfg.Save(); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	transport.emitter.AddListener(transport)

	return transport, nil
}

func (t *Transport) Run(ctx context.Context) error {
	go func() {
		if err := t.responder.Respond(ctx); err != nil {
			log.Info.Printf("responder error: %s", err)
		}

		log.Info.Println("mdns stopped")
	}()

	log.Info.Printf("listening on port: %s", t.hapHTTP.Port())

	return t.hapHTTP.ListenAndServe(ctx)
}

func (t *Transport) IsPaired() bool {
	enteties, err := t.database.Entities()

	return len(enteties) > 1 && err == nil
}

func (t *Transport) AddAccessory(acc *accessory.Accessory) error {
	if err := t.container.AddAccessory(acc); err != nil {
		return fmt.Errorf("failed to add accessory: %w", err)
	}

	for _, svc := range acc.GetServices() {
		for _, ch := range svc.GetCharacteristics() {
			onConnChange := func(conn net.Conn, c *characteristic.Characteristic, new, old interface{}) {
				if c.Events {
					t.notifyListener(acc, c, conn)
				}
			}
			ch.OnValueUpdateFromConn(onConnChange)

			onChange := func(c *characteristic.Characteristic, new, old interface{}) {
				if c.Events {
					t.notifyListener(acc, c, nil)
				}
			}
			ch.OnValueUpdate(onChange)
		}
	}

	t.updateConfig()

	return nil
}

func (t *Transport) notifyListener(a *accessory.Accessory, c *characteristic.Characteristic, except net.Conn) {
	conns := t.hapContext.ActiveConnections()
	for _, conn := range conns {
		if conn == except {
			continue
		}

		resp, err := hap.NewCharacteristicNotification(a, c)
		if err != nil {
			log.Info.Printf("failed to create notification: %s", err)
		}

		var buffer = &bytes.Buffer{}
		_ = resp.Write(buffer)
		data, _ := ioutil.ReadAll(buffer)

		data = bytes.Replace(data, []byte("HTTP/1.0"), []byte("EVENT/1.0"), 1)
		log.Debug.Printf("send notification to %s, data: %q", conn.RemoteAddr().String(), string(data))
		_, _ = conn.Write(data)
	}
}

func (t *Transport) RemoveAccessory(acc *accessory.Accessory) {
	t.container.RemoveAccessory(acc)
	t.updateConfig()
}

func (t *Transport) Handle(ev interface{}) {
	switch ev.(type) {
	case event.DevicePaired:
		log.Debug.Println("paired with device")
		t.UpdateReachability()
	case event.DeviceUnpaired:
		log.Debug.Println("unpaired with device")
		t.UpdateReachability()
	}
}

func (t *Transport) UpdateReachability() {
	t.cfg.Discoverable = !t.IsPaired()
	t.handle.UpdateText(t.cfg.MDNSRecords(), t.responder)
}

func (t *Transport) updateConfig() {
	t.cfg.updateConfigHash(t.container.ContentHash())

	if err := t.cfg.Save(); err != nil {
		log.Info.Printf("failed to save config: %s", err)
	}

	t.handle.UpdateText(t.cfg.MDNSRecords(), t.responder)
}
