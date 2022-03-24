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

	err = f.CheckIP("1.1.1.1") // AU
	assert.NoError(t, err)
	err = f.CheckIP("2.2.2.2") // FR
	assert.Error(t, err)

	err = f.CheckIP("")
	assert.NoError(t, err)
	// private ranges
	err = f.CheckIP("10.8.0.1")
	assert.NoError(t, err)
	err = f.CheckIP("192.168.255.1")
	assert.NoError(t, err)
	err = f.CheckIP("172.16.127.128")
	assert.NoError(t, err)

	f.AllowedCountries = make(map[string]bool)
	f.DeniedCountries = map[string]bool{"FR": true}

	err = f.CheckIP("2.2.2.2") // FR
	assert.Error(t, err)
	err = f.CheckIP("1.1.1.1")
	assert.NoError(t, err)
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
