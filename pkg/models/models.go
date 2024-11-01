package models

import "time"

type FileNode struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	HasChildren bool      `json:"has_children"`
	ChildCount  int       `json:"children_count"`
	IsScraped   bool      `json:"is_scraped"`
	ParentPath  string    `json:"-"`
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
