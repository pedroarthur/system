package main

import "system/collector"
import . "github.com/tj/go-gracefully"
import "github.com/segmentio/go-log"
import "github.com/tj/docopt"
import "time"
import "os"

const Version = "0.2.0"

const Usage = `
  Usage:
    system-stats
      [--statsd-address addr]
      [--memory-interval i]
      [--disk-interval i]
      [--cpu-interval i]
      [--extended]
      [--name name]
    system-stats -h | --help
    system-stats --version

  Options:
    --statsd-address addr   statsd address [default: :8125]
    --memory-interval i     memory reporting interval [default: 10s]
    --disk-interval i       disk reporting interval [default: 30s]
    --cpu-interval i        cpu reporting interval [default: 5s]
    --extended              output additional extended metrics
    --name name             node name defaulting to hostname [default: hostname]
    -h, --help              output help information
    -v, --version           output version
`

func main() {
	args, err := docopt.Parse(Usage, nil, true, Version, false)
	log.Check(err)

	log.Info("starting system %s", Version)

	name := args["--name"].(string)

	if "hostname" == name {
		host, err := os.Hostname()
		log.Check(err)
		name = host
	}

	parameters := collector.CollectionParameters{
		StatsdAddress:         args["--statsd-address"].(string),
		Namespace:             name,
		Extended:              args["--extended"].(bool),
		CpuMonitorInterval:    interval(args, "--cpu-interval"),
		MemoryMonitorInterval: interval(args, "--memory-interval"),
		DiskMonitorInterval:   interval(args, "--disk-interval"),
	}

	c, _ := collector.BuildCollector(parameters)

	c.Start()
	Shutdown()
	c.Stop()
}

func interval(args map[string]interface{}, name string) time.Duration {
	d, err := time.ParseDuration(args[name].(string))
	log.Check(err)
	return d
}
