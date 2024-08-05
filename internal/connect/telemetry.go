package connect

import (
	"log/slog"

	"github.com/SUSE/telemetry"
)

func pushTelemetry(content []byte) {
	// Skip this altogether if the telemetry client cannot register with the
	// upstream telemetry server.
	if telemetry.Status() != telemetry.CLIENT_REGISTERED {
		return
	}

	telemetryType := telemetry.TelemetryType("SCC-CONNECT-DATA")
	class := telemetry.MANDATORY_TELEMETRY
	tags := telemetry.Tags{"scc", "connect"}
	// TODO: SUBMIT is only to be delivered if we don't trust a background agent
	// exists. Moreover, only the last Generate needs to have it (i.e. in case
	// we have multiple of them).
	flags := telemetry.GENERATE | telemetry.SUBMIT

	err := telemetry.Generate(
		telemetryType,
		class,
		content,
		tags,
		flags)
	if err != nil {
		slog.Error(
			"Generate() failed",
			slog.String("type", telemetryType.String()),
			slog.String("class", class.String()),
			slog.Any("content", content),
			slog.String("tags", tags.String()),
			slog.String("flags", flags.String()),
			slog.String("error", err.Error()),
		)
	}
}
