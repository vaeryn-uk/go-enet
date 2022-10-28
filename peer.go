package enet

// #include <enet/enet.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// Peer is a peer which data packets may be sent or received from
type Peer interface {
	GetAddress() Address

	Disconnect(data uint32)
	DisconnectNow(data uint32)
	DisconnectLater(data uint32)

	SendBytes(data []byte, channel uint8, flags PacketFlags) error
	SendString(str string, channel uint8, flags PacketFlags) error
	SendPacket(packet Packet, channel uint8) error

	// SetData sets an arbitrary value against a peer. This is useful to attach some
	// application-specific data for future use, such as an identifier.
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	//
	// For simplicity of implementation, this only allows byte slices up to 255
	// length. If given a slice longer than this, a panic is raised. See
	// MaxPeerDataLength.
	SetData(data []byte)

	// GetData returns an application-specific value that's been set
	// against this peer. This returns nil if no data has been set.
	//
	// http://enet.bespin.org/structENetPeer.html#a1873959810db7ac7a02da90469ee384e
	GetData() []byte
}

// MaxPeerDataLength is the maximum number of bytes we can support being stored
// alongside a peer. See Peer.SetData.
const MaxPeerDataLength = 0xff

type enetPeer struct {
	cPeer *C.struct__ENetPeer
}

func (peer enetPeer) GetAddress() Address {
	return &enetAddress{
		cAddr: peer.cPeer.address,
	}
}

func (peer enetPeer) Disconnect(data uint32) {
	C.enet_peer_disconnect(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) DisconnectNow(data uint32) {
	C.enet_peer_disconnect_now(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) DisconnectLater(data uint32) {
	C.enet_peer_disconnect_later(
		peer.cPeer,
		(C.enet_uint32)(data),
	)
}

func (peer enetPeer) SendBytes(data []byte, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket(data, flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendString(str string, channel uint8, flags PacketFlags) error {
	packet, err := NewPacket([]byte(str), flags)
	if err != nil {
		return err
	}
	return peer.SendPacket(packet, channel)
}

func (peer enetPeer) SendPacket(packet Packet, channel uint8) error {
	C.enet_peer_send(
		peer.cPeer,
		(C.enet_uint8)(channel),
		packet.(enetPacket).cPacket,
	)
	return nil
}

func (peer enetPeer) SetData(data []byte) {
	if len(data) > MaxPeerDataLength {
		panic(fmt.Sprintf("cannot store data with len > %d", MaxPeerDataLength))
	}

	if data == nil {
		peer.cPeer.data = nil
		return
	}

	// First byte is how long our slice is.
	b := make([]byte, len(data)+1)
	b[0] = byte(len(data))
	copy(b[1:], data)
	peer.cPeer.data = unsafe.Pointer(&b[0])
}

func (peer enetPeer) GetData() []byte {
	ptr := unsafe.Pointer(peer.cPeer.data)

	if ptr == nil {
		return nil
	}

	return C.GoBytes(
		// Skip the first byte as this is the slice length.
		unsafe.Add(ptr, 1),
		// Read this many bytes.
		C.int(*(*byte)(ptr)),
	)
}
