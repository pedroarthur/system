//
// The Collector handles starting and stopping of
// all of the resources, and flushes stats on shutdown.
//
package collector

import "system/memory"
import "system/disk"
import "system/cpu"
import sdClient "github.com/statsd/client"
import "system/resource"
import "github.com/statsd/client-namespace"
import "github.com/statsd/client-interface"
import "github.com/segmentio/go-log"
import "sync"
import "time"

// Collector.
type Collector struct {
	Resources []resource.Resource
	client    statsd.Client
	wg        sync.WaitGroup
}

type CollectionParameters struct {
	StatsdAddress         string
	Namespace             string
	Extended              bool
	MemoryMonitorInterval time.Duration
	CpuMonitorInterval    time.Duration
	DiskMonitorInterval   time.Duration
}

func BuildCollector(parameters CollectionParameters) (*Collector, error) {
	client, err := sdClient.Dial(parameters.StatsdAddress)

	if err != nil {
		return nil, err
	}

	c := New(namespace.New(client, parameters.Namespace))
	c.Add(memory.New(parameters.MemoryMonitorInterval, parameters.Extended))
	c.Add(cpu.New(parameters.CpuMonitorInterval, parameters.Extended))
	c.Add(disk.New(parameters.DiskMonitorInterval))

	return c, nil
}

// New collector with the given statsd client.
func New(client statsd.Client) *Collector {
	return &Collector{
		client: client,
	}
}

// Start the collector with the resources
// which have been provided. Each resource gets
// its own prefixed statsd client.
func (c *Collector) Start() error {
	log.Info("starting collector with %d resources", len(c.Resources))

	for _, r := range c.Resources {
		log.Info("starting %s", r.Name())
		c.wg.Add(1)
		err := r.Start(namespace.New(c.client, r.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop the resource collectors gracefully
// and then flush all metrics.
func (c *Collector) Stop() error {
	log.Info("stopping collector")

	for _, r := range c.Resources {
		go func(r resource.Resource) {
			log.Info("stopping %s", r.Name())
			err := r.Stop()
			if err != nil {
				log.Error("failed to gracefully stop %s: %s", r.Name(), err)
			}
			c.wg.Done()
		}(r)
	}

	c.wg.Wait()

	log.Info("flushing stats")
	return c.client.Flush()
}

// Add the given resource for collection.
func (c *Collector) Add(r resource.Resource) {
	c.Resources = append(c.Resources, r)
}
