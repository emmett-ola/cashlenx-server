package cli_auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
)

const (
	DeviceID   = "cashlenx-cli"
	DeviceName = "CashLenX CLI"
	IPAddress  = "127.0.0.1"
	UserAgent  = "cashlenx-cli"
)

type StoredSession struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func Save(accessToken, refreshToken string, user model.UserEntity) error {
	session := StoredSession{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.Id.Hex(),
		Username:     user.Username,
		Role:         user.Role,
		UpdatedAt:    time.Now().UTC(),
	}
	return writeSession(session)
}

func Clear() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func RequireUser() (*provider.Claims, error) {
	session, err := readSession()
	if err != nil {
		return nil, err
	}

	if session.AccessToken != "" {
		claims, err := auth.Service.ValidateToken(session.AccessToken)
		if err == nil && claims != nil {
			return claims, nil
		}
	}

	if session.RefreshToken == "" {
		return nil, notLoggedInError()
	}

	accessToken, refreshToken, user, err := auth.Service.RefreshToken(session.RefreshToken, DeviceID, DeviceName, IPAddress, UserAgent)
	if err != nil {
		_ = Clear()
		return nil, errors.NewUnauthorizedError("CLI session expired; run `cashlenx open auth login`")
	}
	if err := Save(accessToken, refreshToken, user); err != nil {
		return nil, err
	}
	return auth.Service.ValidateToken(accessToken)
}

func RequireAdmin() (*provider.Claims, error) {
	claims, err := RequireUser()
	if err != nil {
		return nil, err
	}
	if claims.Role != model.UserRoleAdmin {
		return nil, errors.NewForbiddenError("admin role required")
	}
	return claims, nil
}

func RequireUserID(target *string) error {
	claims, err := RequireUser()
	if err != nil {
		return err
	}
	if target != nil {
		if *target != "" && *target != claims.UserID {
			return errors.NewForbiddenError("user flag must match logged-in user")
		}
		*target = claims.UserID
	}
	return nil
}

func CurrentSession() (StoredSession, error) {
	return readSession()
}

func readSession() (StoredSession, error) {
	path, err := sessionPath()
	if err != nil {
		return StoredSession{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return StoredSession{}, notLoggedInError()
	}
	if err != nil {
		return StoredSession{}, err
	}
	var session StoredSession
	if err := json.Unmarshal(data, &session); err != nil {
		return StoredSession{}, err
	}
	if session.AccessToken == "" && session.RefreshToken == "" {
		return StoredSession{}, notLoggedInError()
	}
	return session, nil
}

func writeSession(session StoredSession) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func sessionPath() (string, error) {
	if path := os.Getenv("CASHLENX_CLI_AUTH_FILE"); path != "" {
		return path, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cashlenx", "cli_auth.json"), nil
}

func notLoggedInError() error {
	return errors.NewUnauthorizedError(fmt.Sprintf("not logged in; run `%s` first", "cashlenx open auth login --username <username> --password <password>"))
}
