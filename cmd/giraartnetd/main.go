package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jpicht/gira"
	"github.com/jpicht/gira/data"
	"github.com/jsimonetti/go-artnet"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/jsimonetti/go-artnet/packet/code"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	var (
		fake    string
		debug   bool
		dump    bool
		verbose bool
	)

	pflag.StringVar(&fake, "fake-uiconfig", "", "Path to fake uiconfig")
	pflag.BoolVarP(&debug, "debug", "d", false, "debug logging")
	pflag.BoolVarP(&dump, "dump", "D", false, "dump raw universe data")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	pflag.Parse()

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.0000",
		DisableColors:   true,
		FullTimestamp:   true,
	}
	if debug {
		log.Level = logrus.DebugLevel
	} else {
		log.Level = logrus.InfoLevel
	}

	cfg, err := gira.LoadConfigFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	if cfg.IgnoreSSL {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if cfg.ArtNet.Network == "" {
		log.Fatal("missing artnet config")
	}

	var client data.Client

	if fake != "" {
		ui, err := gira.LoadFile[data.UIConfig](fake)
		if err != nil {
			log.Fatal(err)
		}
		client = data.NewFakeClient(ui)
	} else {
		client, err = data.NewRESTClient(*cfg)
		if err != nil {
			log.Fatal(err)
		}
	}

	ui, err := client.UIConfig()
	if err != nil {
		log.Fatal(err)
	}

	var (
		channels = channels{cfg: cfg}
		fixtures []fixture
	)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	for _, fn := range ui.Functions {
		if len(cfg.UIDs) > 0 && !slices.Contains(cfg.UIDs, fn.UID) {
			log.Infof("skipping %q (%s)", fn.DisplayName, fn.UID)
			continue
		}
		switch fn.ChannelType {
		case "de.gira.schema.channels.KNX.Dimmer":
			fixtures = append(fixtures, channels.addFunction(fn))
		case "de.gira.schema.channels.DimmerRGBW":
			fixtures = append(fixtures, channels.addFunction(fn))
		case "de.gira.schema.channels.Switch":
			fixtures = append(fixtures, channels.addFunction(fn))
		case "de.gira.schema.channels.SceneControl":
			log.Println("skipping scene", fn.DisplayName)
			continue
		default:
			enc.Encode(fn)
			log.Fatalf("not implemented: %q", fn.ChannelType)
		}
	}

	slices.SortFunc(fixtures, func(a, b fixture) int {
		return strings.Compare(a.UID, b.UID)
	})

	for offset, ch := range channels.channels {
		if offset != ch.Offset {
			continue
		}
		fmt.Printf("CH %d.%3d [%4s] %s\n", cfg.ArtNet.SubUni, 1+ch.Offset, ch.UID, ch.Name)
	}

	_, cidrnet, err := net.ParseCIDR(cfg.ArtNet.Network)
	if err != nil {
		log.Fatal(err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("error getting ips: %s", err)
	}

	var (
		ip      net.IP
		ok      bool
		current channelValues
		target  channelValues
		lock    sync.Mutex
	)

	for _, addr := range addrs {
		ip = addr.(*net.IPNet).IP
		if cidrnet.Contains(ip) {
			ok = true
			break
		}
	}

	if !ok {
		log.Fatalf("did not find interface for %q", cfg.ArtNet)
	}

	artnet.NewDefaultLogger()
	n := artnet.NewNode("gira-x1", code.StNode, ip, artnet.NewLogger(log.WithFields(nil)))

	n.Config.BaseAddress = artnet.Address{
		Net:    uint8(cfg.ArtNet.Net),
		SubUni: uint8(cfg.ArtNet.SubUni),
	}

	n.Config.Name = "GIRA bridge"
	n.Config.Manufacturer = "jpicht"
	n.Config.InputPorts = append(n.Config.InputPorts, artnet.InputPort{
		Address: artnet.Address{
			Net:    uint8(cfg.ArtNet.Net),
			SubUni: uint8(cfg.ArtNet.SubUni),
		},
		Type: 0, // DMX
	})

	n.RegisterCallback(code.OpDMX, func(p packet.ArtNetPacket) {
		dmx := p.(*packet.ArtDMXPacket)
		if dmx.Net != uint8(cfg.ArtNet.Net) || dmx.SubUni != uint8(cfg.ArtNet.SubUni) {
			log.Debugf("skip %d/%d", dmx.Net, dmx.SubUni)
			return
		}
		log.Debugf("take %d/%d", dmx.Net, dmx.SubUni)
		lock.Lock()
		copy(target[:], dmx.Data[:])
		if dump {
			for i := 0; i < 512; i++ {
				if i%16 == 0 {
					fmt.Printf("\n%02x: ", i)
				}
				fmt.Printf("%02x ", target[i])
			}
			fmt.Println()
		}
		lock.Unlock()
	})

	if err := n.Start(); err != nil {
		log.Fatal("cannot start", err)
	}
	defer n.Stop()

	for {
		time.Sleep(time.Second)
		var (
			buff channelValues
			tmp  *data.ValueBody
		)
		if func() bool {
			lock.Lock()
			defer lock.Unlock()
			tmp = channels.diff(current, target)
			if tmp == nil {
				return false
			}
			copy(buff[:], target[:])
			return true
		}() {
			if err := client.Set(tmp); err != nil {
				log.Error("could not update channels:", err)
				continue
			}

			lock.Lock()
			copy(current[:], buff[:])
			lock.Unlock()

			log.Infof("%d channels updated", len(tmp.Values))
			if verbose {
				for _, v := range tmp.Values {
					log.Infof("%s -> %s", v.UID, v.Value)
				}
			}
		}
	}
}
