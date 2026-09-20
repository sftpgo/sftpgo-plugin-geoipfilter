// Copyright (C) 2022-2023 Nicola Murino
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package filter

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"

	"github.com/oschwald/maxminddb-golang/v2"

	"github.com/sftpgo/sftpgo-plugin-geoipfilter/logger"
)

type countryRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

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
	reader, err := maxminddb.Open(f.DbFilePath)
	if err != nil {
		logger.AppLogger.Error("unable to load the database file, the previous database, if any, is preserved",
			"path", f.DbFilePath, "err", err)
		return err
	}
	logger.AppLogger.Debug("database loaded", "path", f.DbFilePath, "type", reader.Metadata.DatabaseType)

	f.mu.Lock()
	old := f.reader
	f.reader = reader
	f.mu.Unlock()

	if old != nil {
		if err := old.Close(); err != nil {
			logger.AppLogger.Warn("unable to close the previous database", "err", err)
		}
	}
	return nil
}

// Close closes the database file
func (f *Filter) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.reader != nil {
		if err := f.reader.Close(); err != nil {
			logger.AppLogger.Warn("unable to close the database file", "err", err)
		} else {
			logger.AppLogger.Debug("database file closed")
		}
		f.reader = nil
	}
}

func (f *Filter) getCountryCode(ip netip.Addr) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.reader == nil {
		return "", errors.New("database not loaded")
	}
	var record countryRecord
	if err := f.reader.Lookup(ip).Decode(&record); err != nil {
		return "", err
	}
	return record.Country.ISOCode, nil
}

// CheckIP returns an error if the specified IP is not allowed
func (f *Filter) CheckIP(ipAddr, _ string) error {
	ip, err := parseIPAddr(ipAddr)
	if err != nil {
		logger.AppLogger.Warn("error parsing the provided IP address, the IP will be allowed", "ip", ipAddr, "err", err)
		return nil
	}
	if isNonPublic(ip) {
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

var sharedAddressSpace = netip.MustParsePrefix("100.64.0.0/10")

func isNonPublic(ip netip.Addr) bool {
	return !ip.IsGlobalUnicast() || ip.IsPrivate() || sharedAddressSpace.Contains(ip)
}

func parseIPAddr(ipAddr string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(ipAddr)
	if err != nil {
		return netip.Addr{}, err
	}
	return addr.WithZone("").Unmap(), nil
}
