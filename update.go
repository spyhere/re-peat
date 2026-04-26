package main

import (
	"github.com/spyhere/re-peat/internal/autoupdate"
	"github.com/spyhere/re-peat/internal/common"
	"github.com/spyhere/re-peat/internal/prompt"
	"github.com/spyhere/re-peat/internal/state"
)

func checkForUpdate(appState *state.AppState) {
	appState.Lg.Info("Checking for updates")
	rel, ok, err := autoupdate.ShouldUpdate(tag, appState.Cfgs.LastUpdateCheck)
	if err != nil {
		appState.Lg.Error("Autoupdate", err)
		return
	}
	if !ok {
		return
	}
	upd := prompt.UpdateInfo{
		HtmlUrl:     rel.HtmlUrl,
		TagName:     rel.TagName.String(),
		Name:        rel.Name,
		PublishedAt: rel.PublishedAt,
		Body:        rel.Body,
		Size:        rel.Asset.Size,
	}
	if appState.Prompter.AskUpdate(upd) {
		if err = common.OpenBrowserLink(rel.Asset.BrowserDownloadUrl); err != nil {
			appState.Lg.Error("Autoupdate", err)
			return
		}
	}
	appState.Cfgs.MarkUpdateChecked()
}
