package main

// Machine describe machine information
type Machine struct {
	HostName  string
	IPAddress string
	city      string
}

// Latency describe network latency from a source city to destination cities
type Latency struct {
	From   string
	Values map[string]int
}
