# SFTPGO geoipfilter plugin

![Build](https://github.com/sftpgo/sftpgo-plugin-geoipfilter/workflows/Build/badge.svg?branch=main&event=push)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPLv3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

This plugin allows to accept/deny connections based on the the geographical location of the clients' IP addresses.

## Supported databases

This plugin can read MaxMind [GeoLite2](http://dev.maxmind.com/geoip/geoip2/geolite2/) and [GeoIP2](http://www.maxmind.com/en/geolocation_landing) databases. You need to download a country database in MMDB format.

## Configuration

The plugin can be configured within the `plugins` section of the SFTPGo configuration file. To start the plugin you have to use the `serve` subcommand. Here is the usage.

```shell
NAME:
   sftpgo-plugin-geoipfilter serve - Launch the SFTPGo plugin, it must be called from an SFTPGo instance

USAGE:
   sftpgo-plugin-geoipfilter serve [command options] [arguments...]

OPTIONS:
   --db-file value            Path to the MaxMind GeoLite2 or GeoIP2 database [$SFTPGO_PLUGIN_GEOIPFILTER_DB_FILE]
   --allowed-countries value  Comma separated allowed countries in ISO 3166-1 alpha-2 format [$SFTPGO_PLUGIN_GEOIPFILTER_ALLOWED_COUNTRIES]
   --denied-countries value   Comma separated denied countries in ISO 3166-1 alpha-2 format [$SFTPGO_PLUGIN_GEOIPFILTER_DENIED_COUNTRIES]
   --help, -h                 show help (default: false)
```

The `db-file` and at least one bewteen `allowed-countries` and `denied-countries` are required. Each flag can also be set using environment variables.

This is an example configuration.

```json
...
  "plugins": [
    {
      "type": "ipfilter",
      "cmd": "<path to sftpgo-plugin-metadata>",
      "args": ["serve", "--db-file", "GeoLite2-Country.mmdb", "--allowed-countries", "IT,US"],
      "sha256sum": "",
      "auto_mtls": true
    }
  ]
...
```

With the example above, only IP address from Italy and USA are allowed.

The plugin will not start if it fails to load the country database, this will prevent SFTPGo from starting.
If you want to reload the country database without restarting the plugin, simply send a reload command to SFTPGo which will propagate the reload to the plugin.
IPs in private ranges will always be allowed. If the plugin fails to get the country the IP will be allowed.

Countries should be specified in [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) format.
