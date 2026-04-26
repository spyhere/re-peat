package main

import (
	"fmt"
	"time"

	"github.com/spyhere/re-peat/internal/logging"
	"github.com/spyhere/re-peat/internal/state"
)

const logDumpCooldown = time.Minute * 5

func notifyAboutErrors(appState *state.AppState) {
	commonI18n := appState.I18n.Common
	appState.NotifyCrashReportsOnStartup()
	for range appState.Lg.DumpDoneCh {
		body := fmt.Sprintf(commonI18n.LogsDumpedBody, logging.LogReportFileName)
		appState.Prompter.Tell(commonI18n.LogsDumpedTitle, body, commonI18n.InfoDialogOk)
		// Intentionally block DumpDoneCh to stop spamming with the same error logs (dump + notification blocked)
		time.Sleep(logDumpCooldown)
	}
}
