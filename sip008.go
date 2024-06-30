// Package ansip provides a list of API for [SIP008](https://shadowsocks.org/doc/sip008.html)
package ansip

import (
	"github.com/gofrs/uuid/v5"
)

const SIP008Version = 1

// SIP008 is the format of SIP008.
type SIP008 struct {
	// Version marks SIP008 protocol Version. It always is SIP008Version.
	Version int `json:"Version"`

	Servers []Server `json:"servers"`

	// BytesUsed represents data used by the user in bytes.
	BytesUsed uint64 `json:"bytes_used"`

	// BytesRemaining represents data remaining to be used by the user in bytes.
	// If no data limit is in place, the field must be omitted.
	// If a data limit is enforced, the bytes_used field must also be included.
	// In other words, this field must not be specified without the BytesUsed field.
	BytesRemaining uint64 `json:"bytes_remaining"`
}

// NewSIP008 returns a pointer to SIP008 with version SIP008Version.
func NewSIP008(servers []Server, bytesUsed uint64, bytesRemaining uint64) *SIP008 {
	return &SIP008{
		Version:        SIP008Version,
		Servers:        servers,
		BytesUsed:      bytesUsed,
		BytesRemaining: bytesRemaining,
	}
}

// Server records the information of shadowsocks configuration.
type Server struct {
	// Id is a server UUID to distinguish between servers when updating.
	Id uuid.UUID `json:"id"`

	// Remarks is a readable name of server.
	Remarks string `json:"remarks"`

	// Server is the server address.
	Server string `json:"server"`

	// ServerPort is the port of server.
	ServerPort uint16 `json:"server_port"`

	Password   string `json:"password"`
	Method     string `json:"method"`
	Plugin     string `json:"plugin"`
	PluginOpts string `json:"plugin_opts"`
}
