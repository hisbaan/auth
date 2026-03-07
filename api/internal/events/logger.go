package events

import (
	"auth/internal/jet/postgres/public/model"
	"auth/internal/utils/httputil"
	"context"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

type EventWriter interface {
	Create(event model.Events) error
}

func Log(ctx context.Context, writer EventWriter, eventType EventType, userID *ulid.ULID, data any) {
	if writer == nil {
		return
	}
	info, ok := httputil.ClientInfoFromContext(ctx)
	if !ok || info.IP == "" {
		info = httputil.ClientInfo{IP: "0.0.0.0", UserAgent: "unknown"}
	} else if info.UserAgent == "" {
		info.UserAgent = "unknown"
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	var userIDBytes *[]byte
	if userID != nil {
		bytes := userID.Bytes()
		userIDBytes = &bytes
	}

	_ = writer.Create(model.Events{
		ID:        ulid.Make().Bytes(),
		UserID:    userIDBytes,
		Type:      string(eventType),
		Data:      string(payload),
		IPAddress: info.IP,
		UserAgent: info.UserAgent,
		CreatedAt: time.Now(),
	})
}
