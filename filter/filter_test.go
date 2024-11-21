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
	"net"
	"testing"

	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	f := Filter{
		DbFilePath: "../GeoLite2-Country.mmdb",
		AllowedCountries: map[string]bool{
			"AU": true,
		},
	}
	err := f.Reload()
	require.NoError(t, err)
	defer f.Close()

	err = f.CheckIP("1.1.1.1", "") // AU
	assert.NoError(t, err)
	err = f.CheckIP("2.2.2.2", "") // FR
	assert.Error(t, err)

	err = f.CheckIP("", "")
	assert.NoError(t, err)
	// private ranges
	err = f.CheckIP("10.8.0.1", "")
	assert.NoError(t, err)
	err = f.CheckIP("192.168.255.1", "")
	assert.NoError(t, err)
	err = f.CheckIP("172.16.127.128", "")
	assert.NoError(t, err)

	f.AllowedCountries = make(map[string]bool)
	f.DeniedCountries = map[string]bool{"FR": true}

	err = f.CheckIP("2.2.2.2", "") // FR
	assert.Error(t, err)
	err = f.CheckIP("1.1.1.1", "")
	assert.NoError(t, err)
}

func TestParseIPAddr(t *testing.T) {
	ipAddr, err := parseIPAddr("192.168.1.1")
	require.NoError(t, err)
	assert.True(t, ipAddr.IsPrivate())
	ipAddr, err = parseIPAddr("fe80::c523:8ca6:67cc:e1ea%Ethernet 3")
	require.NoError(t, err)
	assert.Equal(t, "fe80::c523:8ca6:67cc:e1ea", ipAddr.String())
	assert.True(t, ipAddr.IsLinkLocalUnicast())
	ipAddr, err = parseIPAddr("212.71.246.123")
	require.NoError(t, err)
	assert.False(t, ipAddr.IsPrivate())
	ipAddr, err = parseIPAddr("2345:0425:2CA1:0000:0000:0567:5673:23b5")
	require.NoError(t, err)
	assert.False(t, ipAddr.IsPrivate())
	ipAddr, err = parseIPAddr("fd7a:115c:a1e0:ab12:4843:cd96:626b:430b")
	require.NoError(t, err)
	assert.True(t, ipAddr.IsPrivate())
}

func BenchmarkIPLookup(b *testing.B) {
	reader, err := maxminddb.Open("../GeoLite2-Country.mmdb")
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	ip1 := net.ParseIP("1.1.1.1")
	ip2 := net.ParseIP("2.2.2.2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var record struct {
			Country struct {
				ISOCode string `maxminddb:"iso_code"`
			} `maxminddb:"country"`
		}
		err = reader.Lookup(ip1, &record)
		if err != nil {
			panic(err)
		}
		err = reader.Lookup(ip2, &record)
		if err != nil {
			panic(err)
		}
	}
}
