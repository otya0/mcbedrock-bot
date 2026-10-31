package join

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

// signIn returns a token source for the Microsoft account.
//
// If a token is saved at tokenPath, it is used. Otherwise the login URL and
// code are written to prompt, and signIn waits until the login in the browser
// is finished. The new token is saved at tokenPath. Tokens refreshed after
// they expire are saved again as well.
//
// The file at tokenPath is the key to the account: never share it.
func signIn(tokenPath string, prompt io.Writer) (oauth2.TokenSource, error) {
	tok, err := readToken(tokenPath)
	if errors.Is(err, fs.ErrNotExist) {
		tok, err = auth.RequestLiveTokenWriter(prompt)
		if err != nil {
			return nil, fmt.Errorf("login: %w", err)
		}
		if err := writeToken(tokenPath, tok); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &savingTokenSource{
		src:  auth.RefreshTokenSource(tok),
		path: tokenPath,
		last: tok.AccessToken,
	}, nil
}

// savingTokenSource saves a token again after it is refreshed.
// Without it, every expiry would mean logging in again.
type savingTokenSource struct {
	mu   sync.Mutex
	src  oauth2.TokenSource
	path string
	last string
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.AccessToken != s.last {
		s.last = t.AccessToken
		// If saving fails, this connection can still use the token, so the
		// error is ignored. The only cost is logging in again next time.
		_ = writeToken(s.path, t)
	}
	return t, nil
}

// readToken reads a saved token. It returns fs.ErrNotExist if there is no file.
func readToken(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	t := new(oauth2.Token)
	if err := json.Unmarshal(b, t); err != nil {
		return nil, fmt.Errorf("token file %s is broken: %w", path, err)
	}
	return t, nil
}

// writeToken saves a token, readable and writable only by the owner (0600).
func writeToken(path string, t *oauth2.Token) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	return nil
}
