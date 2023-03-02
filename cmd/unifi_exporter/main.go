// Command unifi_exporter provides a Prometheus exporter for a Ubiquiti UniFi
// Controller API and UniFi devices.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/visago/unifi"
	"github.com/visago/unifi_exporter"
)

const (
	// userAgent is ther user agent reported to the UniFi Controller API.
	userAgent = "github.com/visago/unifi_exporter"
)

func main() {
	metricsPath := flag.String("metrics.path", "/metrics", "Metrics path")
	listenAddr := flag.String("metrics.listen", ":9130", "Metrics listening port")
	unifiAddr := flag.String("unifi.addr", "https://127.0.0.1:8443/", "URL to Unifi Controller")
	username := flag.String("unifi.username", "admin", "Username for Unifi")
	password := flag.String("unifi.password", "password", "Password for Unifi")
	insecure := flag.Bool("unifi.insecure", true, "Insecure mode")
	site := flag.String("unifi.site", "", "Site for Unifi")
	timeoutString := flag.String("unifi.timeout", "5s", "Timeout Unifi connection")

	flag.Parse()

	timeout := 5 * time.Second
	timeout, err := time.ParseDuration(*timeoutString)
	if err != nil {
		log.Fatalf("failed to parse duration %s: %v", *timeoutString, err)
	}

	clientFn := newClient(
		*unifiAddr,
		*username,
		*password,
		*insecure,
		timeout,
	)
	c, err := clientFn()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	sites, err := c.Sites()
	if err != nil {
		log.Fatalf("failed to retrieve list of sites: %v", err)
	}

	useSites, err := pickSites(*site, sites)
	if err != nil {
		log.Fatalf("failed to select a site: %v", err)
	}

	e, err := unifiexporter.New(useSites, clientFn)
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	prometheus.MustRegister(e)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Printf("Starting UniFi exporter on %s for site(s): %s", *listenAddr, sitesString(useSites))

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.Fatalf("cannot start UniFi exporter: %s", err)
	}
}

// pickSites attempts to find a site with a description matching the value
// specified in choose.  If choose is empty, all sites are returned.
func pickSites(choose string, sites []*unifi.Site) ([]*unifi.Site, error) {
	if choose == "" {
		return sites, nil
	}

	var pick *unifi.Site
	for _, s := range sites {
		if s.Description == choose {
			pick = s
			break
		}
	}
	if pick == nil {
		return nil, fmt.Errorf("site with description %q was not found in UniFi Controller", choose)
	}

	return []*unifi.Site{pick}, nil
}

// sitesString returns a comma-separated string of site descriptions, meant
// for displaying to users.
func sitesString(sites []*unifi.Site) string {
	ds := make([]string, 0, len(sites))
	for _, s := range sites {
		ds = append(ds, s.Description)
	}

	return strings.Join(ds, ", ")
}

// newClient returns a unifiexporter.ClientFunc using the input parameters.
func newClient(addr, username, password string, insecure bool, timeout time.Duration) unifiexporter.ClientFunc {
	return func() (*unifi.Client, error) {
		httpClient := &http.Client{Timeout: timeout}
		if insecure {
			httpClient = unifi.InsecureHTTPClient(timeout)
		}

		c, err := unifi.NewClient(addr, httpClient)
		if err != nil {
			return nil, fmt.Errorf("cannot create UniFi Controller client: %v", err)
		}
		c.UserAgent = userAgent

		if err := c.Login(username, password); err != nil {
			return nil, fmt.Errorf("failed to authenticate to UniFi Controller: %v", err)
		}

		return c, nil
	}
}
