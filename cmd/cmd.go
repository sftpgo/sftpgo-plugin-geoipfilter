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

package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/hashicorp/go-plugin"
	"github.com/sftpgo/sdk/plugin/ipfilter"
	"github.com/urfave/cli/v2"

	"github.com/sftpgo/sftpgo-plugin-geoipfilter/filter"
	"github.com/sftpgo/sftpgo-plugin-geoipfilter/logger"
)

const (
	version   = "1.0.4"
	envPrefix = "SFTPGO_PLUGIN_GEOIPFILTER_"
)

var (
	commitHash = ""
	buildDate  = ""
)

var (
	dbFile           string
	allowedCountries string
	deniedCountries  string

	rootCmd = &cli.App{
		Name:    "sftpgo-plugin-geoipfilter",
		Version: getVersionString(),
		Usage:   "SFTPGo Geoip filter plugin",
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Launch the SFTPGo plugin, it must be called from an SFTPGo instance",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "db-file",
						Usage:       "Path to the MaxMind GeoLite2 or GeoIP2 database",
						Destination: &dbFile,
						EnvVars:     []string{envPrefix + "DB_FILE"},
					},
					&cli.StringFlag{
						Name:        "allowed-countries",
						Usage:       "Comma separated allowed countries in ISO 3166-1 alpha-2 format",
						Destination: &allowedCountries,
						EnvVars:     []string{envPrefix + "ALLOWED_COUNTRIES"},
					},
					&cli.StringFlag{
						Name:        "denied-countries",
						Usage:       "Comma separated denied countries in ISO 3166-1 alpha-2 format",
						Destination: &deniedCountries,
						EnvVars:     []string{envPrefix + "DENIED_COUNTRIES"},
					},
				},
				Action: func(c *cli.Context) error {
					logger.AppLogger.Info("starting sftpgo-plugin-geoipfilter", "version", getVersionString())
					allowed := make(map[string]bool)
					denied := make(map[string]bool)
					var allowedList []string
					var deniedList []string
					for _, country := range strings.Split(allowedCountries, ",") {
						country = strings.TrimSpace(country)
						if country != "" {
							allowed[country] = true
							allowedList = append(allowedList, country)
						}
					}
					for _, country := range strings.Split(deniedCountries, ",") {
						country = strings.TrimSpace(country)
						if country != "" {
							denied[country] = true
							deniedList = append(deniedList, country)
						}
					}
					if len(allowed) == 0 && len(denied) == 0 {
						logger.AppLogger.Error("no country is set")
						return errors.New("please set allowed or denied countries or both")
					}
					logger.AppLogger.Debug("configured countries", "allowed", allowedList, "denied", deniedList)
					filter := &filter.Filter{
						DbFilePath:       dbFile,
						AllowedCountries: allowed,
						DeniedCountries:  denied,
					}
					if err := filter.Reload(); err != nil {
						return err
					}
					plugin.Serve(&plugin.ServeConfig{
						HandshakeConfig: ipfilter.Handshake,
						Plugins: map[string]plugin.Plugin{
							ipfilter.PluginName: &ipfilter.Plugin{Impl: filter},
						},
						GRPCServer: plugin.DefaultGRPCServer,
					})

					filter.Close()
					return errors.New("the plugin exited unexpectedly")
				},
			},
		},
	}
)

// Execute runs the root command
func Execute() error {
	return rootCmd.Run(os.Args)
}

func getVersionString() string {
	var sb strings.Builder
	sb.WriteString(version)
	if commitHash != "" {
		sb.WriteString("-")
		sb.WriteString(commitHash)
	}
	if buildDate != "" {
		sb.WriteString("-")
		sb.WriteString(buildDate)
	}
	return sb.String()
}
