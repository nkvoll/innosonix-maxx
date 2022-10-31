package auto_ampenable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/nkvoll/innosonix-maxx/cmd/common"
	"github.com/nkvoll/innosonix-maxx/internal"
	"github.com/nkvoll/innosonix-maxx/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

var silenceLevel float64
var holdTime time.Duration

func init() {
	lf := AutoAmpenableCmd.Flags()

	lf.Float64Var(&silenceLevel, "silence-level", -100.0, "silence dB level")
	lf.DurationVar(&holdTime, "hold-time", 60*time.Second, "hold time before muting")
}

var AutoAmpenableCmd = &cobra.Command{
	Use:   "auto-ampenable",
	Short: "Automatically toggles ampenable for channels based on detected signals.",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.WithFields(log.Fields{
			"address":       common.Addr,
			"silence_level": silenceLevel,
			"hold_time":    holdTime,
			"token":         len(common.Token) > 0,
		}).Info("Starting")
		client, err := internal.NewClient(common.Addr, common.Token)

		if err != nil {
			log.Fatal(err)
		}

		ctx, cancelFunc := context.WithCancel(context.Background())

		signalsChan := make(chan map[int]bool)
		ampenabledsChan := make(chan map[int]bool)

		go func() {
			defer cancelFunc()

			// latest channel -> hasSignal value (as a momentary measurement)
			var latestSignals map[int]bool
			// latest channel -> ampenabled status
			var latestAmpenableds map[int]bool
			// channel -> time when we last detected silence (! presence of signal)
			silenceSince := make(map[int]*time.Time)
			// channel -> rateLimiter for the update settings REST API call
			updateSettingsRateLimiter := make(map[int]*rate.Limiter)
			// channel ->  rateLimiter for the delay logging
			holdTimeRemainingLoggingRateLimiters := make(map[int]*rate.Limiter)

			reconcile := func() error {
				if latestSignals == nil || latestAmpenableds == nil {
					log.WithFields(log.Fields{"signals": len(latestSignals), "ampenableds": len(latestAmpenableds)}).Println("Missing data")
					return nil
				}

				for i, hasSignal := range latestSignals {
					channelId := i + 1
					isAmpenabled := latestAmpenableds[i]

					if !hasSignal && isAmpenabled {
						if silenceSince[i] == nil {
							log.WithFields(log.Fields{"channel": channelId}).Println("No signal detected on channel.")
							silenceSince[i] = internal.Pointer(time.Now())
						}
					}

					if hasSignal && silenceSince[i] != nil {
						log.WithFields(log.Fields{"channel": channelId}).Println("Signal detected on channel, resetting hold time.")
						delete(silenceSince, i)
					}

					if hasSignal != isAmpenabled {
						if !hasSignal {
							minHoldTime := time.Now().Add(-holdTime)
							if minHoldTime.Before(*silenceSince[i]) {
								// rate limit to 1 log entry per channel per 10 seconds
								holdTimeRateLimiter, ok := holdTimeRemainingLoggingRateLimiters[i]
								if !ok {
									holdTimeRateLimiter = rate.NewLimiter(rate.Every(10*time.Second), 1)
									holdTimeRemainingLoggingRateLimiters[i] = holdTimeRateLimiter
								}
								if holdTimeRateLimiter.Allow() {
									log.WithFields(log.Fields{
										"channel":         channelId,
										"remaining_delay": silenceSince[i].Sub(minHoldTime),
									}).Println("No signal detected on channel, will mute after hold time.")
								}
								continue
							}
						}

						// rate limit to 1 request per channel per second
						rateLimiter, ok := updateSettingsRateLimiter[i]
						if !ok {
							rateLimiter = rate.NewLimiter(rate.Every(1*time.Second), 1)
							updateSettingsRateLimiter[i] = rateLimiter
						}

						if rateLimiter.Allow() {
							log.WithFields(log.Fields{"channel": channelId, "ampenable": hasSignal}).Println("Updating ampenable for channel.")

							res, err := client.PutSettingsChannelChannelIdAmpenableWithResponse(
								ctx,
								channelId,
								api.PutSettingsChannelChannelIdAmpenableJSONRequestBody(api.Boolean{Value: internal.Pointer(hasSignal)}),
							)
							if err != nil {
								return err
							}

							// update local cache
							switch res.StatusCode() {
							case http.StatusOK:
								latestAmpenableds[i] = hasSignal
							default:
								log.WithFields(log.Fields{
									"channel":     channelId,
									"ampenable":   hasSignal,
									"status_code": res.StatusCode(),
									"body":        string(res.Body),
								}).Println("Unexpected status code received when updating ampenable for channel.")
								continue
							}

							log.WithFields(log.Fields{"channel": channelId, "ampenable": hasSignal}).Println("Updated ampenable for channel.")
						}
					}
				}

				return nil
			}

			for {
				select {
				case signals := <-signalsChan:
					latestSignals = signals
				case ampenableds := <-ampenabledsChan:
					latestAmpenableds = ampenableds
				case <-ctx.Done():
					return
				}

				if err := reconcile(); err != nil {
					log.WithError(err).Printf("Error caught while reconciling (non-fatal).")
				}
			}
		}()

		go func() {
			defer cancelFunc()

			if err := streamSignals(ctx, common.Addr, signalsChan); err != nil {
				log.WithError(err).Printf("Error caught while streaming signals.")
			}
		}()

		go func() {
			defer cancelFunc()

			if err := streamAmpenableds(ctx, common.Addr, ampenabledsChan); err != nil {
				log.WithError(err).Printf("Error caught while ampenabled states.")
			}
		}()

		<-ctx.Done()

		return nil
	},
}

func streamSignals(ctx context.Context, addr string, ch chan<- map[int]bool) error {
	url := url.URL{Scheme: "ws", Host: addr, Path: internal.WEBSOCKET_LEVEL_PATH}

	return internal.StreamMessagesWithRetry(ctx, url, func(message []byte) error {
		response := api.WSLevelResponse{}
		if err := json.Unmarshal(message, &response); err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		// Ok, at this point it's pretty clear I don't actually know how to understand the
		// values in the nested levels array. The first level matches the channels in
		// number, but the second layer is less clear to me.

		// Some random samples captured when the channel is completely idle (No Dante input).
		//
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.3 -69.1 -162.5 -162.5 -162.5 -162.5]
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.3 -69.1 -162.5 -162.5 -162.5 -162.5]
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.3 -69.2 -162.5 -162.5 -162.5 -162.5]
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.3 -68.6 -162.5 -162.5 -162.5 -162.5]
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.6 -69 -162.5 -162.5 -162.5 -162.5]
		// [-162.5 -162.5 -162.5 -162.5 -162.5 -162.5 -69.3 -69.1 -162.5 -162.5 -162.5 -162.5]
		//
		// I initially attempted filtering for "any value above <silence-level>", but that works
		// poorly for very quiet music.
		//
		// Thus I'm currently averaging all of them and have what I feel is a relatively high
		// silence level value. What bothers me perhaps more than it should is that I think
		// I saw literally all channels lie at an average of somewhere closer to -95 for an
		// extended period of time, but I haven't been able to reproduce this.
		//
		// For my specific use case, I would prefer if I had more programmatic access to the
		// Dante "signal" indicator (e.g as found in Dante Controller), which seems to very
		// closely match whether I'm sending audio across the network or not. See
		// https://dev.audinate.com/GA/dante-controller/userguide/webhelp/content/receive_tab.htm
		signals := make(map[int]bool)
		for i, outputLevel := range response.Level {
			hasSignal := false
			sum := 0.0
			for _, l := range outputLevel {
				sum += l
				//if l > *silenceLevel {
				//	hasSignal = true
				//	break
				//}
			}
			avg := sum / float64(len(outputLevel))
			hasSignal = avg > silenceLevel
			signals[i] = hasSignal
		}

		select {
		case ch <- signals:
		case <-ctx.Done():
			log.WithFields(log.Fields{"url": url.String()}).Println("Closing stream")
			return nil
		}

		return nil
	})
}

func streamAmpenableds(ctx context.Context, addr string, ch chan<- map[int]bool) error {
	url := url.URL{Scheme: "ws", Host: addr, Path: internal.WEBSOCKET_DATAPOLL_PATH}

	return internal.StreamMessagesWithRetry(ctx, url, func(message []byte) error {
		response := api.WSDatapollResponse{}
		if err := json.Unmarshal(message, &response); err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		if response.Settings != nil {
			if response.Settings.Channel != nil {
				ampenables := make(map[int]bool)
				for i, channel := range *(response.Settings).Channel {
					ampenables[i] = *channel.Ampenable.Value
				}

				select {
				case ch <- ampenables:
				case <-ctx.Done():
					log.WithFields(log.Fields{"url": url.String()}).Println("Closing stream")
					return nil
				}
			}
		}
		return nil
	})
}
