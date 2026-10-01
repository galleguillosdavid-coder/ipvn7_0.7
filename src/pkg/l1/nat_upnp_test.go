package l1

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUPnPMapper_InvalidPort(t *testing.T) {
	mapper := NewUPnPMapper(100 * time.Millisecond)
	err := mapper.DiscoverAndForward(0, "test")
	if err == nil {
		t.Fatal("se esperaba error con puerto 0")
	}
	err = mapper.DiscoverAndForward(70000, "test")
	if err == nil {
		t.Fatal("se esperaba error con puerto > 65535")
	}
}

func TestUPnPMapper_MockGateway(t *testing.T) {
	var soapReceived bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/desc.xml" {
			xml := `<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <device>
    <serviceList>
      <service>
        <serviceType>urn:schemas-upnp-org:service:WANIPConnection:1</serviceType>
        <controlURL>/ctl</controlURL>
      </service>
    </serviceList>
  </device>
</root>`
			w.Header().Set("Content-Type", "text/xml")
			w.Write([]byte(xml))
			return
		}

		if r.URL.Path == "/ctl" && r.Method == "POST" {
			action := r.Header.Get("SOAPAction")
			if action != "\"urn:schemas-upnp-org:service:WANIPConnection:1#AddPortMapping\"" {
				http.Error(w, "invalid soap action", http.StatusBadRequest)
				return
			}
			soapReceived = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<s:Envelope xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\"><s:Body></s:Body></s:Envelope>"))
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	mapper := NewUPnPMapper(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	location := fmt.Sprintf("%s/desc.xml", ts.URL)
	err := mapper.addPortMapping(ctx, location, 7777, "ipvn7-unit-test")
	if err != nil {
		t.Fatalf("addPortMapping fallo: %v", err)
	}

	if !soapReceived {
		t.Fatal("el servidor mock no recibio la peticion SOAP AddPortMapping")
	}
}
