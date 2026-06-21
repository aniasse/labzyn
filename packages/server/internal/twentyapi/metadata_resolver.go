package twentyapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aniasse/labzyn/server/internal/auth"
	"github.com/aniasse/labzyn/server/internal/core"
	"github.com/aniasse/labzyn/server/internal/metadata"
)

// MetadataResolver is the root resolver for the /metadata GraphQL endpoint.
type MetadataResolver struct {
	core   *core.Repo
	meta   *metadata.Repo
	secret string
}

type minimalMetadataT struct {
	ObjectMetadataItems []*objectMetadataItemMinimalT
	Views               []*minimalViewT
	CollectionHashes    []*collectionHashT
}

type objectMetadataItemMinimalT struct {
	Id            UUID
	NameSingular  string
	NamePlural    string
	LabelSingular string
	LabelPlural   string
	Icon          *string
	IsCustom      bool
	IsActive      bool
	IsSystem      bool
	IsRemote      bool
}

type minimalViewT struct {
	Id               UUID
	Type             string
	Key              *string
	ObjectMetadataId UUID
}

type collectionHashT struct {
	CollectionName string
	Hash           string
}

// MinimalMetadata returns the lightweight object/view catalog the frontend
// uses to bootstrap its metadata store (and render the sidebar).
func (r *MetadataResolver) MinimalMetadata(ctx context.Context) (*minimalMetadataT, error) {
	wsID, err := r.core.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	objs, err := r.meta.ListObjects(ctx, wsID)
	if err != nil {
		return nil, err
	}
	items := []*objectMetadataItemMinimalT{}
	views := []*minimalViewT{}
	for _, o := range objs {
		icon := o.Icon
		items = append(items, &objectMetadataItemMinimalT{
			Id:            UUID(o.ID),
			NameSingular:  o.NameSingular,
			NamePlural:    o.NamePlural,
			LabelSingular: o.LabelSingular,
			LabelPlural:   o.LabelPlural,
			Icon:          &icon,
			IsCustom:      !o.IsSystem,
			IsActive:      o.IsActive,
			IsSystem:      o.IsSystem,
			IsRemote:      false,
		})
		key := "INDEX"
		views = append(views, &minimalViewT{
			Id:               UUID(o.ID),
			Type:             "table",
			Key:              &key,
			ObjectMetadataId: UUID(o.ID),
		})
	}
	for _, so := range standardObjects() {
		o := so.Object
		icon := o.Icon
		items = append(items, &objectMetadataItemMinimalT{
			Id: UUID(o.ID), NameSingular: o.NameSingular, NamePlural: o.NamePlural,
			LabelSingular: o.LabelSingular, LabelPlural: o.LabelPlural, Icon: &icon,
			IsCustom: false, IsActive: true, IsSystem: true, IsRemote: false,
		})
	}

	return &minimalMetadataT{
		ObjectMetadataItems: items,
		Views:               views,
		CollectionHashes:    []*collectionHashT{},
	}, nil
}

// ── GraphQL output types (resolved via UseFieldResolvers) ────────────────────

type publicWorkspaceData struct {
	Id                  UUID
	AuthProviders       authProvidersT
	AuthBypassProviders *authBypassProvidersT
	Logo                *string
	DisplayName         *string
	WorkspaceUrls       workspaceUrlsT
}

type authProvidersT struct {
	Sso       []ssoIdentityProviderT
	Google    bool
	MagicLink bool
	Password  bool
	Microsoft bool
}

type authBypassProvidersT struct {
	Google    bool
	Password  bool
	Microsoft bool
}

type ssoIdentityProviderT struct {
	Id     string
	Name   string
	Type   string
	Status string
	Issuer string
}

type workspaceUrlsT struct {
	CustomUrl    *string
	SubdomainUrl string
}

type authTokenT struct {
	Token     string
	ExpiresAt DateTime
}

type authTokenPairT struct {
	AccessOrWorkspaceAgnosticToken authTokenT
	RefreshToken                   authTokenT
}

type loginTokenT struct{ LoginToken authTokenT }

type authTokensT struct{ Tokens authTokenPairT }

type analyticsT struct{ Success bool }

type checkUserExistT struct {
	Exists                   bool
	AvailableWorkspacesCount float64
	IsEmailVerified          bool
}

// ── Resolvers ────────────────────────────────────────────────────────────────

// GetPublicWorkspaceDataByDomain returns the public data of the workspace so the
// frontend can render the sign-in form. Single-workspace: returns the default.
func (r *MetadataResolver) GetPublicWorkspaceDataByDomain(
	ctx context.Context, args struct{ Origin *string },
) (*publicWorkspaceData, error) {
	wsID, err := r.core.OldestWorkspaceID(ctx)
	if err != nil {
		return nil, err
	}
	origin := "http://localhost:3001"
	if args.Origin != nil && *args.Origin != "" {
		origin = *args.Origin
	}
	displayName := "Labzyn"
	if wsID != "" {
		if ws, err := r.core.GetWorkspace(ctx, wsID); err == nil {
			displayName = ws.DisplayName
		}
	}
	return &publicWorkspaceData{
		Id:            UUID(wsID),
		AuthProviders: authProvidersT{Sso: []ssoIdentityProviderT{}, Password: true},
		DisplayName:   &displayName,
		WorkspaceUrls: workspaceUrlsT{SubdomainUrl: origin},
	}, nil
}

// CheckUserExists tells the sign-in form whether to ask for a password.
func (r *MetadataResolver) CheckUserExists(
	ctx context.Context, args struct {
		Email        string
		CaptchaToken *string
	},
) (*checkUserExistT, error) {
	_, err := r.core.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(args.Email)))
	exists := err == nil
	count := 0.0
	if exists {
		count = 1
	}
	return &checkUserExistT{Exists: exists, AvailableWorkspacesCount: count, IsEmailVerified: true}, nil
}

// GetLoginTokenFromCredentials verifies email/password and issues a login token.
func (r *MetadataResolver) GetLoginTokenFromCredentials(
	ctx context.Context,
	args struct {
		Email                   string
		Password                string
		CaptchaToken            *string
		Locale                  *string
		VerifyEmailRedirectPath *string
		Origin                  string
	},
) (*loginTokenT, error) {
	u, err := r.core.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(args.Email)))
	if err != nil || !auth.CheckPassword(u.PasswordHash, args.Password) {
		return nil, errors.New("Wrong password or email")
	}
	exp := time.Now().Add(15 * time.Minute)
	tok, err := auth.NewToken(r.secret, u.ID, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	return &loginTokenT{LoginToken: authTokenT{Token: tok, ExpiresAt: DateTime{exp}}}, nil
}

// GetAuthTokensFromLoginToken exchanges a login token for access/refresh tokens.
func (r *MetadataResolver) GetAuthTokensFromLoginToken(
	ctx context.Context, args struct{ LoginToken, Origin string },
) (*authTokensT, error) {
	uid, err := auth.ParseToken(r.secret, args.LoginToken)
	if err != nil {
		return nil, errors.New("Invalid login token")
	}
	accExp := time.Now().Add(24 * time.Hour)
	refExp := time.Now().Add(30 * 24 * time.Hour)
	acc, err := auth.NewToken(r.secret, uid, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	ref, err := auth.NewToken(r.secret, uid, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &authTokensT{Tokens: authTokenPairT{
		AccessOrWorkspaceAgnosticToken: authTokenT{Token: acc, ExpiresAt: DateTime{accExp}},
		RefreshToken:                   authTokenT{Token: ref, ExpiresAt: DateTime{refExp}},
	}}, nil
}

// TrackAnalytics is a no-op that satisfies the frontend's telemetry calls.
func (r *MetadataResolver) TrackAnalytics(
	ctx context.Context,
	args struct {
		Type       string
		Name       *string
		Event      *string
		Properties *JSON
	},
) (*analyticsT, error) {
	return &analyticsT{Success: true}, nil
}
