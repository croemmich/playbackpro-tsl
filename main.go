package main // import "github.com/croemmich/playbackpro-tsl"
import (
	"github.com/croemmich/playbackpro-tsl/config"
	"github.com/croemmich/playbackpro-tsl/playbackpro"
	"github.com/croemmich/playbackpro-tsl/tsl"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"time"
)

var logger = logrus.WithField("component", "main")

func main() {
	config.Configure()

	running := true
	done := make(chan interface{})
	go func() {
		// create playback pro connection
		pbp := playbackpro.NewPlaybackPro(config.PbPNetAddress(), config.PbPTimeout())
		defer pbp.Close()

		// start the playback pro proxy
		err := pbp.StartProxy(config.PbPProxyListenAddress())
		if err != nil {
			logger.WithError(err).Error("error starting proxy")
		}
		defer pbp.StopProxy()

		// create the tally connection
		tally := tsl.NewTally(config.TallyNetProtocol(), config.TallyNetAddress(), config.TallyNetTimeout())
		defer tally.Close()

		// monitor playback pro and write updates to tally
		for running {
			playing, err := pbp.GetPlaybackStatus()
			if err != nil {
				logger.WithError(err).Error("failed to get PlaybackPro status")
				sendTallyProgram(tally, "PBP Conn Err", tsl.Clear)
				sendTallyPreview(tally, "PBP Conn Err", tsl.Clear)
				time.Sleep(config.PbPTimeout())
				continue
			}
			if config.TallyAddressPreviewClipName() >= 0 {
				name, _ := pbp.GetPreviewClipName()
				if name != "N/A" {
					_ = tally.Send(config.TallyAddressPreviewClipName(), name, tsl.Tally1|tsl.BrightnessFull)
				} else {
					_ = tally.Send(config.TallyAddressPreviewClipName(), "", tsl.Clear)
				}
			}
			if config.TallyAddressPreviewClipDuration() >= 0 {
				duration, _ := pbp.GetPreviewClipDuration()
				if duration != "N/A" {
					_ = tally.Send(config.TallyAddressPreviewClipDuration(), truncateTime(duration), tsl.Tally1|tsl.BrightnessFull)
				} else {
					_ = tally.Send(config.TallyAddressPreviewClipDuration(), "", tsl.Clear)
				}
			}
			if playing {
				if config.TallyAddressRemaining() >= 0 {
					remaining, _ := pbp.GetProgramTimeRemaining()
					_ = tally.Send(config.TallyAddressRemaining(), truncateTime(remaining), tsl.Tally1|tsl.BrightnessFull)
				}
				if config.TallyAddressElapsed() >= 0 {
					elapsed, _ := pbp.GetProgramTimeElapsed()
					_ = tally.Send(config.TallyAddressElapsed(), truncateTime(elapsed), tsl.Tally1|tsl.BrightnessFull)
				}
				if config.TallyAddressClipName() >= 0 {
					name, _ := pbp.GetProgramClipName()
					_ = tally.Send(config.TallyAddressClipName(), name, tsl.Tally1|tsl.BrightnessFull)
				}
				if config.TallyAddressClipDuration() >= 0 {
					duration, _ := pbp.GetProgramClipDuration()
					_ = tally.Send(config.TallyAddressClipDuration(), truncateTime(duration), tsl.Tally1|tsl.BrightnessFull)
				}
				time.Sleep(config.PbPPollIntervalPlaying())
			} else {
				sendTallyProgram(tally, "", tsl.Clear)
				sendTallyPreview(tally, "", tsl.Clear)
				time.Sleep(config.PbPPollIntervalStopped())
			}
		}
		close(done)
	}()

	// wait for os signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	<-sigc
	running = false
	<-done
}

func truncateTime(time string) string {
	if len(time) > 8 {
		time = time[0:8]
		time = strings.TrimPrefix(time, "00:")
	}
	return time
}

func sendTallyProgram(tally *tsl.Tally, display string, control tsl.ControlByte) {
	if config.TallyAddressClipName() >= 0 {
		_ = tally.Send(config.TallyAddressClipName(), display, control)
	}
	if config.TallyAddressClipDuration() >= 0 {
		_ = tally.Send(config.TallyAddressClipDuration(), display, control)
	}
	if config.TallyAddressElapsed() >= 0 {
		_ = tally.Send(config.TallyAddressElapsed(), display, control)
	}
	if config.TallyAddressRemaining() >= 0 {
		_ = tally.Send(config.TallyAddressRemaining(), display, control)
	}
}

func sendTallyPreview(tally *tsl.Tally, display string, control tsl.ControlByte) {
	if config.TallyAddressPreviewClipName() >= 0 {
		_ = tally.Send(config.TallyAddressPreviewClipName(), display, control)
	}
	if config.TallyAddressPreviewClipDuration() >= 0 {
		_ = tally.Send(config.TallyAddressPreviewClipDuration(), display, control)
	}
}
