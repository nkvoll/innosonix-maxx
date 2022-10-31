package mute

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/davecgh/go-spew/spew"
	"github.com/nkvoll/innosonix-maxx/cmd/common"
	"github.com/nkvoll/innosonix-maxx/internal"
	"github.com/spf13/cobra"
)

// MuteCmd is a quick and dirty placeholder showing how to interact with the REST API.
var MuteCmd = &cobra.Command{
	Use:   "mute",
	Short: "Show channel muting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		client, err := internal.NewClient(common.Addr, common.Token)
		if err != nil {
			return err
		}

		settingsResponse, err := client.GetSettingsWithResponse(ctx)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{"settings": spew.Sdump(settingsResponse.JSON200)}).Debug("Got settings")

		if settingsResponse.JSON200.Channel != nil {
			for _, channel := range *settingsResponse.JSON200.Channel {
				log.WithFields(log.Fields{"channel": *channel.ChannelId, "muted": *channel.Dsp.Mute.Value}).Printf("Mute status")
			}
		}

		// note this is not the same mute as the channel mutes above
		// boolValue := false
		// res, err := client.PutSettingsChannelChannelIdAmpenableWithResponse(
		// 	ctx,
		// 	1,
		// 	api.PutSettingsChannelChannelIdAmpenableJSONRequestBody(api.Boolean{Value: &boolValue}),
		// )
		// if err != nil {
		// 	panic(err)
		// }

		// log.Println(string(res.Body))

		// patchRes, err := client.PutSettingsChannelChannelIdDspPatchWithResponse(ctx, 1, []api.SettingsPatch{
		// 	{
		// 		PatchId: internal.Pointer(2),
		// 		Mute:    internal.Pointer(false),
		// 	},
		// })
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// log.Println(string(patchRes.Body))

		return nil
	},
}
