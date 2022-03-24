package filter

import (
	"fmt"
	"net"
	"sync"

	"github.com/oschwald/maxminddb-golang"

	"github.com/sftpgo/sftpgo-plugin-geoipfilter/logger"
)

// Filter implements the ipfilter interface
type Filter struct {
	DbFilePath       string
	AllowedCountries map[string]bool
	DeniedCountries  map[string]bool
	mu               sync.RWMutex
	reader           *maxminddb.Reader
}

// Reload loads the database file
func (f *Filter) Reload() error {
	f.Close()
	reader, err := maxminddb.Open(f.DbFilePath)
	if err != nil {
		logger.AppLogger.Error("unable to load the database file", "path", f.DbFilePath, "err", err)
		return err
	}
	logger.AppLogger.Debug("database loaded", "path", f.DbFilePath)
	f.mu.Lock()
	defer f.mu.Unlock()

	f.reader = reader
	return nil
}

// Close closes the database file
func (f *Filter) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.reader != nil {
		err := f.reader.Close()
		logger.AppLogger.Debug("closed db file", "err", err)
		f.reader = nil
	}
}

func (f *Filter) getCountryCode(ip net.IP) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.reader == nil {
		return "", nil
	}
	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	err := f.reader.Lookup(ip, &record)
	if err != nil {
		return "", err
	}
	return record.Country.ISOCode, nil
}

// CheckIP returns an error if the specified IP is not allowed
func (f *Filter) CheckIP(ipAddr string) error {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		logger.AppLogger.Warn("error parsing the provided IP address", "ip", ipAddr)
		return nil
	}
	if ip.IsPrivate() {
		return nil
	}
	country, err := f.getCountryCode(ip)
	if err != nil {
		logger.AppLogger.Warn("unable to lookup the provided IP address, the IP will be allowed", "ip", ipAddr, "err", err)
		return nil
	}
	if country == "" {
		logger.AppLogger.Warn("unable to get country, the IP will be allowed", "ip", ipAddr)
		return nil
	}
	if f.DeniedCountries[country] {
		logger.AppLogger.Debug("country denied", "ip", ipAddr, "country", country)
		return fmt.Errorf("country %s is denied, ip %s", country, ipAddr)
	}
	if len(f.AllowedCountries) == 0 {
		return nil
	}
	if f.AllowedCountries[country] {
		return nil
	}
	logger.AppLogger.Debug("country not allowed", "ip", ipAddr, "country", country)
	return fmt.Errorf("country %s is not allowed, ip %s", country, ipAddr)
}
