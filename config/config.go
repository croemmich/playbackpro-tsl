package config // import "github.com/croemmich/playbackpro-tsl/config"
import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

const name = "playbackpro-tsl"

var logger = logrus.WithField("component", "config")

func PbPNetAddress() string                 { return viper.GetString("PbPNetAddress") }
func PbPTimeout() time.Duration             { return viper.GetDuration("PbPNetTimeout") }
func PbPPollIntervalStopped() time.Duration { return viper.GetDuration("PbPPollIntervalStopped") }
func PbPPollIntervalPlaying() time.Duration { return viper.GetDuration("PbPPollIntervalPlaying") }
func PbPProxyListenAddress() string         { return viper.GetString("PbPProxyListenAddress") }

func TallyNetAddress() string              { return viper.GetString("TallyNetAddress") }
func TallyNetProtocol() string             { return viper.GetString("TallyNetProtocol") }
func TallyNetTimeout() time.Duration       { return viper.GetDuration("TallyNetTimeout") }
func TallyAddressClipName() int            { return viper.GetInt("TallyAddressClipName") }
func TallyAddressClipDuration() int        { return viper.GetInt("TallyAddressClipDuration") }
func TallyAddressElapsed() int             { return viper.GetInt("TallyAddressElapsed") }
func TallyAddressRemaining() int           { return viper.GetInt("TallyAddressRemaining") }
func TallyAddressPreviewClipName() int     { return viper.GetInt("TallyAddressPreviewClipName") }
func TallyAddressPreviewClipDuration() int { return viper.GetInt("TallyAddressPreviewClipDuration") }

func Configure() {
	setDefaults()
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + name)
	viper.AddConfigPath("/usr/local/etc/" + name)
	viper.AddConfigPath("$HOME/." + name)
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("no configuration file found")
		} else {
			logger.WithError(err).Panic("failed to read configuration file")
		}
	}
}

func setDefaults() {
	viper.SetDefault("PbPNetAddress", "127.0.0.1:4647")
	viper.SetDefault("PbPProxyListenAddress", "0.0.0.0:4648")
	viper.SetDefault("PbPNetTimeout", "5s")
	viper.SetDefault("PbPPollIntervalStopped", "2s")
	viper.SetDefault("PbPPollIntervalPlaying", "1s")

	viper.SetDefault("TallyNetProtocol", "tcp")
	viper.SetDefault("TallyNetAddress", "127.0.0.1:5727")
	viper.SetDefault("TallyNetTimeout", "5s")
	viper.SetDefault("TallyAddressClipName", 0)
	viper.SetDefault("TallyAddressClipDuration", 1)
	viper.SetDefault("TallyAddressElapsed", 2)
	viper.SetDefault("TallyAddressRemaining", 3)
	viper.SetDefault("TallyAddressPreviewClipName", 4)
	viper.SetDefault("TallyAddressPreviewClipDuration", 5)
}
