package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// https://github.com/strideynet/bsky-furry-feed/blob/main/bluesky/pds_client.go
// DefaultPDSHost is now the vPDS - be cautious - making calls for user data who
// aren't on the same PDS as the authenticated account may fail. Use BGS or AppView.
const DefaultPDSHost = "https://bsky.social"

type tokenInfo struct {
	authInfo  *xrpc.AuthInfo
	expiresAt time.Time
}

var timeFormat = "2006-01-02T15:04:05.999999999Z"

func FormatTime(t time.Time) string {
	return t.UTC().Format(timeFormat)
}

// https://github.com/strideynet/bsky-furry-feed/blob/2e8d7bd35dd43a66cb96d972e3a1608fcbafad9e/feed/feed.go#L22s
type Meta struct {
	// ID is the rkey that is used to identify the Feed in generation requests.
	ID string
	// DisplayName is the short name of the feed used in the BlueSky client.
	DisplayName string
	// Description is a long description of the feed used in the BlueSky client.
	Description string
}

func tokenInfoFromAuthInfo(authInfo *xrpc.AuthInfo) (tokenInfo, error) {
	var claims jwt.RegisteredClaims
	if _, _, err := jwt.NewParser().ParseUnverified(authInfo.AccessJwt, &claims); err != nil {
		// Temp hack: ignore ErrTokenUnverifiable which has been triggered by BlueSky potentially changing the
		// signing algo.
		if !errors.Is(err, jwt.ErrTokenUnverifiable) {
			return tokenInfo{}, fmt.Errorf("failed to parse jwt: %w", err)
		}
	}

	return tokenInfo{
		authInfo:  authInfo,
		expiresAt: claims.ExpiresAt.Time,
	}, nil
}

type PDSClient struct {
	pdsHost     string
	tokenInfo   tokenInfo
	tokenInfoMu sync.Mutex
}

type Credentials struct {
	Identifier string
	Password   string
}

func CredentialsFromEnv() (*Credentials, error) {
	godotenv.Load()
	identifier := os.Getenv("BLUESKY_USERNAME")
	if identifier == "" {
		return nil, fmt.Errorf("BLUESKY_USERNAME environment variable not set")
	}
	password := os.Getenv("BLUESKY_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("BLUESKY_PASSWORD environment variable not set")
	}

	return &Credentials{Identifier: identifier, Password: password}, nil
}

func ClientFromCredentials(ctx context.Context, pdsHost string, credentials *Credentials) (*PDSClient, error) {
	c := &PDSClient{
		pdsHost: pdsHost,
	}

	sess, err := atproto.ServerCreateSession(
		ctx,
		c.baseXRPCClient(),
		&atproto.ServerCreateSession_Input{
			Identifier: credentials.Identifier,
			Password:   credentials.Password,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	ti, err := tokenInfoFromAuthInfo(&xrpc.AuthInfo{
		AccessJwt:  sess.AccessJwt,
		RefreshJwt: sess.RefreshJwt,
		Did:        sess.Did,
		Handle:     sess.Handle,
	})
	if err != nil {
		return nil, err
	}

	// c.tokenInfoMu does not need to be locked here on first initialization.
	c.tokenInfo = ti

	return c, nil
}

const UserAgent = ""

func (c *PDSClient) baseXRPCClient() *xrpc.Client {
	ua := UserAgent
	return &xrpc.Client{
		Host:      c.pdsHost,
		UserAgent: &ua,
	}
}

func (c *PDSClient) xrpcClient(ctx context.Context) (*xrpc.Client, error) {
	c.tokenInfoMu.Lock()
	defer c.tokenInfoMu.Unlock()

	if time.Now().After(c.tokenInfo.expiresAt.Add(-10 * time.Minute)) {
		if err := c.refreshToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	xc := c.baseXRPCClient()
	xc.Auth = &xrpc.AuthInfo{
		AccessJwt:  c.tokenInfo.authInfo.AccessJwt,
		RefreshJwt: c.tokenInfo.authInfo.RefreshJwt,
		Handle:     c.tokenInfo.authInfo.Handle,
		Did:        c.tokenInfo.authInfo.Did,
	}
	return xc, nil
}

func (c *PDSClient) refreshToken(ctx context.Context) error {
	xc := c.baseXRPCClient()
	xc.Auth = &xrpc.AuthInfo{
		AccessJwt: c.tokenInfo.authInfo.RefreshJwt,
	}

	sess, err := atproto.ServerRefreshSession(ctx, xc)
	if err != nil {
		return fmt.Errorf("refresh session: %w", err)
	}

	ti, err := tokenInfoFromAuthInfo(&xrpc.AuthInfo{
		AccessJwt:  sess.AccessJwt,
		RefreshJwt: sess.RefreshJwt,
		Did:        sess.Did,
		Handle:     sess.Handle,
	})
	if err != nil {
		return err
	}

	c.tokenInfo = ti
	return nil
}

func (c *PDSClient) ResolveHandle(ctx context.Context, handle string) (*atproto.IdentityResolveHandle_Output, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}
	return atproto.IdentityResolveHandle(ctx, xc, handle)
}

// GetProfile fetches an actor's profile. actor can be a DID or a handle.
func (c *PDSClient) GetProfile(
	ctx context.Context, actor string,
) (*bsky.ActorDefs_ProfileViewDetailed, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}
	return bsky.ActorGetProfile(ctx, xc, actor)
}

// RepoPutRecord_Input This exists because the go code gen is incorrect for swapRecord and misses
// an omitEmpty on SwapRecord.
// putting feed record: putting record: XRPC ERROR 400: InvalidSwap: Record was at bafyreigkeuzjkpot7yzpseezz4hat2jmlobypfhtaaisxbdlwafwxp4ywa
type RepoPutRecord_Input struct {
	// collection: The NSID of the record collection.
	Collection string `json:"collection" cborgen:"collection"`
	// record: The record to write.
	Record *util.LexiconTypeDecoder `json:"record" cborgen:"record"`
	// repo: The handle or DID of the repo.
	Repo string `json:"repo" cborgen:"repo"`
	// rkey: The key of the record.
	Rkey string `json:"rkey" cborgen:"rkey"`
	// swapCommit: Compare and swap with the previous commit by cid.
	SwapCommit *string `json:"swapCommit,omitempty" cborgen:"swapCommit,omitempty"`
	// swapRecord: Compare and swap with the previous record by cid.
	SwapRecord *string `json:"swapRecord,omitempty" cborgen:"swapRecord,omitempty"`
	// validate: Validate the record?
	Validate *bool `json:"validate,omitempty" cborgen:"validate,omitempty"`
}

// PutRecord creates or updates a record in the actor's repository.
func (c *PDSClient) PutRecord(
	ctx context.Context, collection, rkey string, record repo.CborMarshaler,
) error {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return fmt.Errorf("get xrpc client: %w", err)
	}

	var out atproto.RepoPutRecord_Output
	if err := xc.Do(ctx, xrpc.Procedure, "application/json", "com.atproto.repo.putRecord", nil, &RepoPutRecord_Input{
		Collection: collection,
		Repo:       xc.Auth.Did,
		Rkey:       rkey,
		Record: &util.LexiconTypeDecoder{
			Val: record,
		},
	}, &out); err != nil {
		return err
	}
	return nil
}

func (c *PDSClient) UploadBlob(
	ctx context.Context, blob io.Reader,
) (*util.LexBlob, error) {
	xc, err := c.xrpcClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get xrpc client: %w", err)
	}

	// set encoding: 'image/png'
	out, err := atproto.RepoUploadBlob(ctx, xc, blob)
	if err != nil {
		return nil, fmt.Errorf("uploading blob: %w", err)
	}
	return out.Blob, nil
}

func getBlueskyClient(ctx context.Context) (*PDSClient, error) {
	creds, err := CredentialsFromEnv()
	if err != nil {
		return nil, err
	}
	return ClientFromCredentials(ctx, DefaultPDSHost, creds)
}

func run(cctx *context.Context) error {

	log := &slog.Logger{}

	client, err := getBlueskyClient(*cctx)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("./octocat.png", os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("reading avatar: %w", err)
	}
	blob, err := client.UploadBlob(*cctx, f)
	if err != nil {
		return fmt.Errorf("uploading avatar: %w", err)
	}

	hostname := os.Getenv("HOSTNAME")

	meta := Meta{"rkey", "githubfeed", "A feed of potsts with GitHub links"}

	log.Info("Creating Feed", slog.String("rkey", meta.ID))

	err = client.PutRecord(*cctx, "app.bsky.feed.generator", meta.ID, &bsky.FeedGenerator{
		Avatar:      blob,
		Did:         fmt.Sprintf("did:web:%s", hostname),
		CreatedAt:   FormatTime(time.Now().UTC()),
		Description: &meta.Description,
		DisplayName: meta.DisplayName,
	})

	if err != nil {
		return fmt.Errorf("putting feed record: %w", err)
	}

}
