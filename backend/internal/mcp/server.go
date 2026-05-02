package mcp

import (
	"context"
	"errors"
	"time"

	"fraud-detection/internal/domain"
	"fraud-detection/internal/transaction"
	"fraud-detection/internal/user"

	"github.com/google/uuid"
	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Deps struct {
	Users        *user.Service
	Transactions *transaction.Service
}

func RunStdio(deps Deps) error {
	s := server.NewMCPServer("fraud-detection", "1.0.0")

	s.AddTool(
		mcpsdk.NewTool("get_recent_frauds",
			mcpsdk.WithDescription("Lists transactions flagged as fraud (fraud_status=true) within created_at range. "+
				"Use DD.MM.YYYY only (local calendar day). Leave `to` empty for end=now."),
			mcpsdk.WithString("from", mcpsdk.Required(), mcpsdk.Description("Range start: DD.MM.YYYY (e.g. 01.05.2026)")),
			mcpsdk.WithString("to", mcpsdk.Description("Range end DD.MM.YYYY; omit or empty for current time.")),
			mcpsdk.WithNumber("limit", mcpsdk.Description("Max rows (default 500, cap 500)")),
		),
		func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			return handleGetRecentFrauds(ctx, deps, req)
		},
	)

	s.AddTool(
		mcpsdk.NewTool("check_user_status",
			mcpsdk.WithDescription("Returns user profile including account status and risk_rate (0–100)."),
			mcpsdk.WithString("user_id", mcpsdk.Required(), mcpsdk.Description("User UUID")),
		),
		func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
			return handleCheckUserStatus(ctx, deps, req)
		},
	)

	return server.ServeStdio(s)
}

func handleGetRecentFrauds(_ context.Context, deps Deps, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	fromStr, err := req.RequireString("from")
	if err != nil {
		return mcpsdk.NewToolResultError("from is required"), nil
	}
	toStr := req.GetString("to", "")

	from, err := parseRangeBound(fromStr, boundFrom)
	if err != nil {
		return mcpsdk.NewToolResultError("invalid from: " + err.Error()), nil
	}
	to, err := parseRangeBound(toStr, boundTo)
	if err != nil {
		return mcpsdk.NewToolResultError("invalid to: " + err.Error()), nil
	}
	limit := req.GetInt("limit", 0)

	txs, err := deps.Transactions.FindFraudByCreatedAtRange(from, to, limit)
	if err != nil {
		if errors.Is(err, transaction.ErrInvalidTimeRange) {
			return mcpsdk.NewToolResultError("invalid time range: to must not be before from"), nil
		}
		return mcpsdk.NewToolResultErrorFromErr("query failed", err), nil
	}

	out, err := mcpsdk.NewToolResultJSON(txs)
	if err != nil {
		return mcpsdk.NewToolResultErrorFromErr("encode result", err), nil
	}
	return out, nil
}

type userStatusPayload struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Surname   string            `json:"surname"`
	Email     *string           `json:"email,omitempty"`
	Status    domain.UserStatus `json:"status"`
	RiskRate  float64           `json:"risk_rate"`
	CreatedAt time.Time         `json:"created_at"`
}

func handleCheckUserStatus(_ context.Context, deps Deps, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	idStr, err := req.RequireString("user_id")
	if err != nil {
		return mcpsdk.NewToolResultError("user_id is required"), nil
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return mcpsdk.NewToolResultError("invalid user_id"), nil
	}

	u, err := deps.Users.GetUserByID(uid)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return mcpsdk.NewToolResultError("user not found"), nil
		}
		return mcpsdk.NewToolResultErrorFromErr("lookup failed", err), nil
	}

	payload := userStatusPayload{
		ID:        u.ID,
		Name:      u.Name,
		Surname:   u.Surname,
		Email:     u.Email,
		Status:    u.UserStatus,
		RiskRate:  u.RiskRate,
		CreatedAt: u.CreatedAt,
	}
	out, err := mcpsdk.NewToolResultJSON(payload)
	if err != nil {
		return mcpsdk.NewToolResultErrorFromErr("encode result", err), nil
	}
	return out, nil
}
