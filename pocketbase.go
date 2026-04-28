package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const collectionName = "signups"

var (
	errEmailAlreadySignedUp = errors.New("email already signed up")
	errInvalidEmail         = errors.New("invalid email")
)

type pocketBaseClient struct {
	baseURL string
	email   string
	pass    string
	token   string
	client  *http.Client
	timeout time.Duration
}

func newPocketBaseClient(cfg config) *pocketBaseClient {
	return &pocketBaseClient{
		baseURL: cfg.PocketBaseURL,
		email:   cfg.SuperuserEmail,
		pass:    cfg.SuperuserPass,
		token:   cfg.SuperuserToken,
		client:  &http.Client{Timeout: cfg.RequestTimeout},
		timeout: cfg.RequestTimeout,
	}
}

func (pb *pocketBaseClient) ensureSignupsCollection(ctx context.Context) error {
	token, err := pb.authToken(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pb.baseURL+"/api/collections/"+collectionName, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return pb.createSignupsCollection(ctx, token)
	default:
		return fmt.Errorf("check collection returned HTTP %d", resp.StatusCode)
	}
}

func (pb *pocketBaseClient) createSignupsCollection(ctx context.Context, token string) error {
	body := map[string]any{
		"name": collectionName,
		"type": "base",
		// Keep the direct PocketBase API private. The app server owns validation
		// and writes records with a server-side superuser token.
		"listRule":   nil,
		"viewRule":   nil,
		"createRule": nil,
		"updateRule": nil,
		"deleteRule": nil,
		"fields": []map[string]any{
			{
				"name":     "email",
				"type":     "email",
				"required": true,
			},
			{
				"name": "source",
				"type": "text",
			},
		},
		"indexes": []string{
			"CREATE UNIQUE INDEX idx_signups_email ON signups (email)",
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return nil
	}

	// If two app instances start at once, the loser may see a validation error
	// because the collection now exists. Re-check before failing startup.
	if resp.StatusCode == http.StatusBadRequest {
		return pb.expectCollectionExists(ctx, token)
	}

	return fmt.Errorf("create collection returned HTTP %d", resp.StatusCode)
}

func (pb *pocketBaseClient) expectCollectionExists(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pb.baseURL+"/api/collections/"+collectionName, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("recheck collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("collection recheck returned HTTP %d", resp.StatusCode)
}

func (pb *pocketBaseClient) createSignup(ctx context.Context, email string) error {
	token, err := pb.authToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]string{
		"email":  email,
		"source": "ovek-example",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections/"+collectionName+"/records", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return fmt.Errorf("create signup: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return signupBadRequestError(body)
	default:
		return fmt.Errorf("create signup returned HTTP %d", resp.StatusCode)
	}
}

func signupBadRequestError(body []byte) error {
	bodyText := strings.ToLower(string(body))
	switch {
	case strings.Contains(bodyText, "validation_invalid_email"),
		strings.Contains(bodyText, "invalid email"):
		return errInvalidEmail
	case strings.Contains(bodyText, "validation_not_unique"),
		strings.Contains(bodyText, "already"),
		strings.Contains(bodyText, "unique"),
		strings.Contains(bodyText, "idx_signups_email"):
		return errEmailAlreadySignedUp
	default:
		return fmt.Errorf("create signup returned HTTP %d", http.StatusBadRequest)
	}
}

func (pb *pocketBaseClient) authToken(ctx context.Context) (string, error) {
	if strings.TrimSpace(pb.token) != "" {
		return pb.token, nil
	}
	if strings.TrimSpace(pb.email) == "" || strings.TrimSpace(pb.pass) == "" {
		return "", errors.New("set PB_SUPERUSER_TOKEN or PB_SUPERUSER_EMAIL and PB_SUPERUSER_PASSWORD")
	}

	body := map[string]string{
		"identity": pb.email,
		"password": pb.pass,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pb.baseURL+"/api/collections/_superusers/auth-with-password", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := pb.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("authenticate PocketBase superuser: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authenticate PocketBase superuser returned HTTP %d", resp.StatusCode)
	}

	var auth struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return "", err
	}
	if auth.Token == "" {
		return "", errors.New("PocketBase auth response did not include a token")
	}

	return auth.Token, nil
}
