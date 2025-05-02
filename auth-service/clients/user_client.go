package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"igaku/auth-service/errors"
	"igaku/commons/models"
)

type UserClient interface {
	FindByUsername(username string) (*models.User, error)
	Persist(user *models.User) error
}

type userClient struct {
	url string
}

// TODO: Use custom errors
func (c *userClient) FindByUsername(username string) (*models.User, error) {
	url := fmt.Sprintf("%s/internal-accounts/find-by-username/%s", c.url, username)
	httpClient := http.Client{}

	res, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusNotFound:
		return nil, &errors.UserNotFoundError{}
	case http.StatusOK:
		var user models.User
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if err := json.Unmarshal(body, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user: %w", err)
		}
		return &user, nil
	default:
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
}

// TODO: Consider modifying this function so that it takes only username and
// password as parameters.
func (c *userClient) Persist(user *models.User) error {
	url := fmt.Sprintf("%s/internal-accounts/persist", c.url)

	payload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"unexpected response [%d]: %s",
			resp.StatusCode, string(body),
		)
	}

	return nil
}

func NewUserClient(url string) UserClient {
	return &userClient{url: url}
}
