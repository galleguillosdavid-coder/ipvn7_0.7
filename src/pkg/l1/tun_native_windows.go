//go:build windows

package l1

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"

	"ipvn7/pkg/l0"
)

const (
	WintunRingCapacity = 0x400000 // 4 MiB (potencia de dos canónica WireGuard/Wintun)
)

var (
	modKernel32       = syscall.NewLazyDLL("kernel32.dll")
	procRtlMoveMemory = modKernel32.NewProc("RtlMoveMemory")
)

// NativeTunAdapter encapsula la sesión real con el driver de kernel Wintun de Windows
type NativeTunAdapter struct {
	name      string
	ipv4      net.IP
	ipv6      net.IP
	mtu       int
	closed    atomic.Bool
	closeOnce sync.Once

	dll       *syscall.DLL
	adapter   uintptr
	session   uintptr
	readEvent syscall.Handle

	// Punteros a procedimientos dinámicos del driver
	procCloseAdapter        *syscall.Proc
	procEndSession          *syscall.Proc
	procReceivePacket       *syscall.Proc
	procReleaseReceivePkt   *syscall.Proc
	procAllocateSendPacket  *syscall.Proc
	procSendPacket          *syscall.Proc

	rxPackets atomic.Uint64
	txPackets atomic.Uint64
}

func createNativeTun(id *l0.Identity, devName string) (TunAdapter, error) {
	dllName := "wintun.dll"
	if _, err := os.Stat(dllName); err != nil {
		return nil, errors.New("driver wintun.dll no detectado, usando adaptador de espacio de usuario con aislamiento")
	}

	dll, err := syscall.LoadDLL(dllName)
	if err != nil {
		return nil, fmt.Errorf("error cargando wintun.dll: %w", err)
	}

	procCreateAdapter, err := dll.FindProc("WintunCreateAdapter")
	if err != nil {
		dll.Release()
		return nil, fmt.Errorf("wintun: WintunCreateAdapter no encontrado: %w", err)
	}
	procCloseAdapter, _ := dll.FindProc("WintunCloseAdapter")
	procStartSession, _ := dll.FindProc("WintunStartSession")
	procEndSession, _ := dll.FindProc("WintunEndSession")
	procGetReadWaitEvent, _ := dll.FindProc("WintunGetReadWaitEvent")
	procReceivePacket, _ := dll.FindProc("WintunReceivePacket")
	procReleaseReceivePkt, _ := dll.FindProc("WintunReleaseReceivePacket")
	procAllocateSendPacket, _ := dll.FindProc("WintunAllocateSendPacket")
	procSendPacket, _ := dll.FindProc("WintunSendPacket")

	namePtr, err := syscall.UTF16PtrFromString(devName)
	if err != nil {
		dll.Release()
		return nil, err
	}
	tunnelTypePtr, _ := syscall.UTF16PtrFromString("ipvn7")

	// Crear o abrir adaptador L3 físico
	adapterHandle, _, errCreate := procCreateAdapter.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(tunnelTypePtr)),
		0, // GUID nil -> auto-generado determinista
	)
	if adapterHandle == 0 {
		// Intentar abrir si ya existía la interfaz
		procOpenAdapter, errOpenProc := dll.FindProc("WintunOpenAdapter")
		if errOpenProc == nil {
			adapterHandle, _, _ = procOpenAdapter.Call(uintptr(unsafe.Pointer(namePtr)))
		}
		if adapterHandle == 0 {
			dll.Release()
			return nil, fmt.Errorf("wintun: fallo creando interfaz de red de kernel: %v", errCreate)
		}
	}

	// Iniciar sesión con anillo de memoria compartida
	sessionHandle, _, errSession := procStartSession.Call(adapterHandle, uintptr(WintunRingCapacity))
	if sessionHandle == 0 {
		if procCloseAdapter != nil {
			procCloseAdapter.Call(adapterHandle)
		}
		dll.Release()
		return nil, fmt.Errorf("wintun: fallo al iniciar sesión de buffers L3: %v", errSession)
	}

	var readEvt syscall.Handle
	if procGetReadWaitEvent != nil {
		evt, _, _ := procGetReadWaitEvent.Call(sessionHandle)
		readEvt = syscall.Handle(evt)
	}

	return &NativeTunAdapter{
		name:                   devName,
		ipv4:                   id.IPv4(),
		ipv6:                   id.IPv6(),
		mtu:                    DefaultMTU,
		dll:                    dll,
		adapter:                adapterHandle,
		session:                sessionHandle,
		readEvent:              readEvt,
		procCloseAdapter:       procCloseAdapter,
		procEndSession:         procEndSession,
		procReceivePacket:      procReceivePacket,
		procReleaseReceivePkt:  procReleaseReceivePkt,
		procAllocateSendPacket: procAllocateSendPacket,
		procSendPacket:         procSendPacket,
	}, nil
}

func (a *NativeTunAdapter) ReadPacket() ([]byte, error) {
	if a.closed.Load() {
		return nil, errors.New("adaptador TUN cerrado")
	}

	for {
		if a.closed.Load() {
			return nil, errors.New("adaptador TUN cerrado")
		}

		var size uint32
		ptr, _, _ := a.procReceivePacket.Call(a.session, uintptr(unsafe.Pointer(&size)))
		if ptr != 0 && size > 0 {
			packet := make([]byte, size)
			procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&packet[0])), ptr, uintptr(size))
			a.procReleaseReceivePkt.Call(a.session, ptr)
			a.rxPackets.Add(1)
			return packet, nil
		}

		// Esperar evento con timeout de 50ms para permitir cancelaciones limpias
		if a.readEvent != 0 {
			_, _ = syscall.WaitForSingleObject(a.readEvent, 50)
		} else {
			return nil, errors.New("wintun: evento de lectura no inicializado")
		}
	}
}

func (a *NativeTunAdapter) WritePacket(data []byte) error {
	if a.closed.Load() {
		return errors.New("adaptador TUN cerrado")
	}
	if len(data) > a.mtu {
		return fmt.Errorf("paquete excede MTU de %d bytes (tamaño: %d)", a.mtu, len(data))
	}

	ptr, _, _ := a.procAllocateSendPacket.Call(a.session, uintptr(len(data)))
	if ptr == 0 {
		return errors.New("wintun: anillo de transmisión lleno, paquete descartado")
	}

	if len(data) > 0 {
		procRtlMoveMemory.Call(ptr, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data)))
	}

	a.procSendPacket.Call(a.session, ptr)
	a.txPackets.Add(1)
	return nil
}

func (a *NativeTunAdapter) Close() error {
	a.closeOnce.Do(func() {
		a.closed.Store(true)
		if a.session != 0 && a.procEndSession != nil {
			a.procEndSession.Call(a.session)
			a.session = 0
		}
		if a.adapter != 0 && a.procCloseAdapter != nil {
			a.procCloseAdapter.Call(a.adapter)
			a.adapter = 0
		}
		if a.dll != nil {
			a.dll.Release()
		}
	})
	return nil
}

func (a *NativeTunAdapter) MTU() int {
	return a.mtu
}

func (a *NativeTunAdapter) IPv4() net.IP {
	return a.ipv4
}

func (a *NativeTunAdapter) IPv6() net.IP {
	return a.ipv6
}

func (a *NativeTunAdapter) Mode() AdapterMode {
	return ModeKernelVirtual
}

func (a *NativeTunAdapter) IsUserspace() bool {
	return false
}
