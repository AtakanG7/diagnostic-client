package models

import "time"

type FileNode struct {
	Path        string    `json:"path"`
	ParentPath  string    `json:"parent_path"`
	Name        string    `json:"name"`
	IsDirectory bool      `json:"is_directory"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	IsGzipped   bool      `json:"is_gzipped"`
	IsScraped   bool      `json:"is_scraped"`
}

type LogEntry struct {
	ID        int64     `json:"-"`
	Filename  string    `json:"filename"`
	Line      string    `json:"line"`
	LineNum   int       `json:"line_num"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
}

type NetworkPacket struct {
	Timestamp   time.Time `json:"timestamp"`
	Protocol    string    `json:"protocol"`
	SrcIP       string    `json:"src_ip"`
	DstIP       string    `json:"dst_ip"`
	SrcPort     int       `json:"src_port"`
	DstPort     int       `json:"dst_port"`
	Length      int       `json:"length"`
	PayloadSize int       `json:"payload_size"`
	TCPFlags    string    `json:"tcp_flags,omitempty"`
}

type NetworkStats struct {
	PacketCount        int64            `json:"packet_count"`
	TotalBytes         int64            `json:"total_bytes"`
	AvgPacketSize      float64          `json:"avg_packet_size"`
	UniqueSources      int64            `json:"unique_sources"`
	UniqueDestinations int64            `json:"unique_destinations"`
	ProtocolCount      int64            `json:"protocol_count"`
	ProtocolStats      map[string]int64 `json:"protocol_stats"`
	Packets            []NetworkPacket  `json:"packets"`
}

type TopNetworkStats struct {
	TopSources      map[string]int64 `json:"top_sources"`
	TopDestinations map[string]int64 `json:"top_destinations"`
	TopProtocols    map[string]int64 `json:"top_protocols"`
	TopPorts        map[string]int64 `json:"top_ports"`
}
