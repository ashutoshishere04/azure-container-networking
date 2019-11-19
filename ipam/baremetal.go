// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package ipam

import (
	"errors"
	"net"
	"runtime"
	"github.com/Azure/azure-container-networking/log"
)

const (
	envname                   = "baremetal"
)

// baremetal IPAM configuration source.
type baremetalSource struct {
	name       string
	sink       addressConfigSink
	fileLoaded bool
	filePath   string
}

// Creates the baremetal source.
func newBaremetalSource(options map[string]interface{}) (*baremetalSource, error) {
	var filePath string
	if runtime.GOOS == windows {
		filePath = defaultWindowsFilePath
	} else {
		filePath = defaultLinuxFilePath
	}

	return &baremetalSource{
		name: envname,
		filePath: filePath,
	}, nil
}

// Starts the baremetal source.
func (source *baremetalSource) start(sink addressConfigSink) error {
	source.sink = sink
	return nil
}

// Stops the baremetal source.
func (source *baremetalSource) stop() {
	source.sink = nil
}

// Refreshes configuration.
func (source *baremetalSource) refresh() error {
	if source == nil {
		return errors.New("baremetalSource is nil")
	}

	if source.fileLoaded {
		return nil
	}

	// Query the list of local interfaces.
	localInterfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// Query the list of Azure Network Interfaces
	sdnInterfaces, err := getSDNInterfaces(source.filePath)
	if err != nil {
		return err
	}

	// Configure the local default address space.
	local, err := source.sink.newAddressSpace(LocalDefaultAddressSpaceId, LocalScope)
	if err != nil {
		return err
	}

	if err = populateAddressSpace(local, sdnInterfaces, localInterfaces); err != nil {
		return err
	}

	// Set the local address space as active.
	if err = source.sink.setAddressSpace(local); err != nil {
		return err
	}

	log.Printf("[ipam] [baremetal] Address space successfully populated from config file")
	source.fileLoaded = true

	return nil
}
