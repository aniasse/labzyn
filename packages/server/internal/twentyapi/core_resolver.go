package twentyapi

import (
	"context"
	"errors"

	"github.com/aniasse/labzyn/server/internal/core"
)

// CoreResolver is the root resolver for /graphql (currentUser, records later).
type CoreResolver struct {
	core   *core.Repo
	secret string
}

func (*CoreResolver) Ping() *string { s := "labzyn-core"; return &s }

// ── User graph (resolved via UseFieldResolvers) ──────────────────────────────

type userT struct {
	Id                      UUID
	FirstName               string
	LastName                string
	Email                   string
	HasPassword             bool
	CanAccessFullAdminPanel bool
	CanImpersonate          bool
	SupportUserHash         *string
	OnboardingStatus        *string
	WorkspaceMember         *workspaceMemberT
	WorkspaceMembers        []*workspaceMemberT
	DeletedWorkspaceMembers []*deletedWorkspaceMemberT
	CurrentUserWorkspace    *userWorkspaceT
	CurrentWorkspace        *workspaceT
	AvailableWorkspaces     availableWorkspacesT
	UserVars                *JSON
}

type fullNameT struct {
	FirstName string
	LastName  string
}

type workspaceMemberT struct {
	Id               UUID
	Name             fullNameT
	ColorScheme      string
	AvatarUrl        *string
	Locale           *string
	UserEmail        string
	UserWorkspaceId  *UUID
	TimeZone         *string
	DateFormat       *string
	TimeFormat       *string
	CalendarStartDay *int32
	NumberFormat     *string
}

type deletedWorkspaceMemberT struct {
	Id        UUID
	Name      fullNameT
	AvatarUrl *string
	UserEmail string
}

type userWorkspaceT struct {
	Id                                   UUID
	PermissionFlags                      []string
	ObjectsPermissions                   []*objectPermissionT
	TwoFactorAuthenticationMethodSummary []*twoFactorSummaryT
}

type twoFactorSummaryT struct {
	TwoFactorAuthenticationMethodId string
	Status                          string
	Strategy                        string
}

type objectPermissionT struct {
	ObjectMetadataId                  UUID
	CanReadObjectRecords              *bool
	CanUpdateObjectRecords            *bool
	CanSoftDeleteObjectRecords        *bool
	CanDestroyObjectRecords           *bool
	RestrictedFields                  *JSON
	RowLevelPermissionPredicates      []*rowLevelPredicateT
	RowLevelPermissionPredicateGroups []*rowLevelPredicateGroupT
}

type rowLevelPredicateT struct {
	Id                                         string
	FieldMetadataId                            string
	ObjectMetadataId                           string
	Operand                                    string
	SubFieldName                               *string
	WorkspaceMemberFieldMetadataId             *string
	WorkspaceMemberSubFieldName                *string
	RowLevelPermissionPredicateGroupId         *string
	PositionInRowLevelPermissionPredicateGroup *float64
	RoleId                                     string
	Value                                      *JSON
}

type rowLevelPredicateGroupT struct {
	Id                                         string
	ParentRowLevelPermissionPredicateGroupId   *string
	LogicalOperator                            string
	PositionInRowLevelPermissionPredicateGroup *float64
	RoleId                                     string
	ObjectMetadataId                           string
}

type workspaceT struct {
	Id                                UUID
	DisplayName                       *string
	Logo                              *string
	InviteHash                        *string
	AllowImpersonation                bool
	ActivationStatus                  string
	IsPublicInviteLinkEnabled         bool
	IsGoogleAuthEnabled               bool
	IsMicrosoftAuthEnabled            bool
	IsPasswordAuthEnabled             bool
	IsGoogleAuthBypassEnabled         bool
	IsMicrosoftAuthBypassEnabled      bool
	IsPasswordAuthBypassEnabled       bool
	Subdomain                         string
	CustomDomain                      *string
	HasValidSignedEnterpriseKey       bool
	HasValidEnterpriseValidityToken   bool
	WorkspaceCustomApplication        *applicationT
	InstalledApplications             []*applicationT
	IsCustomDomainEnabled             bool
	WorkspaceUrls                     workspaceUrlsT
	FeatureFlags                      []*featureFlagT
	MetadataVersion                   float64
	CurrentBillingSubscription        *billingSubscriptionT
	BillingSubscriptions              []*billingSubscriptionT
	BillingEntitlements               []*billingEntitlementT
	WorkspaceMembersCount             *float64
	DefaultRole                       *roleT
	FastModel                         *string
	SmartModel                        *string
	AiAdditionalInstructions          *string
	EnabledAiModelIds                 []string
	UseRecommendedModels              bool
	IsTwoFactorAuthenticationEnforced bool
	TrashRetentionDays                *float64
	EventLogRetentionDays             *float64
	EditableProfileFields             []string
	IsInternalMessagesImportEnabled   bool
}

type applicationT struct {
	Id                  UUID
	Name                string
	UniversalIdentifier string
	Logo                *string
}

type featureFlagT struct {
	Key   string
	Value bool
}

type billingEntitlementT struct {
	Key   string
	Value bool
}

type roleT struct {
	Id                            UUID
	Label                         string
	Description                   *string
	Icon                          *string
	CanUpdateAllSettings          bool
	CanAccessAllTools             bool
	IsEditable                    bool
	CanReadAllObjectRecords       bool
	CanUpdateAllObjectRecords     bool
	CanSoftDeleteAllObjectRecords bool
	CanDestroyAllObjectRecords    bool
	CanBeAssignedToUsers          bool
	CanBeAssignedToAgents         bool
	CanBeAssignedToApiKeys        bool
}

type availableWorkspacesT struct {
	AvailableWorkspacesForSignIn []*availableWorkspaceT
	AvailableWorkspacesForSignUp []*availableWorkspaceT
}

type availableWorkspaceT struct {
	Id                  UUID
	DisplayName         *string
	LoginToken          *string
	InviteHash          *string
	PersonalInviteToken *string
	WorkspaceUrls       workspaceUrlsT
	Logo                *string
	Sso                 []*ssoConnectionT
}

type ssoConnectionT struct {
	Type   string
	Id     UUID
	Issuer string
	Name   string
	Status string
}

type billingSubscriptionT struct {
	Id                       UUID
	Status                   string
	Interval                 *string
	Metadata                 *JSON
	CurrentPeriodEnd         *float64
	Phases                   []*billingPhaseT
	BillingSubscriptionItems []*billingItemT
}

type billingPhaseT struct {
	Start_date *float64
	End_date   *float64
	Items      []*billingPhaseItemT
}

type billingPhaseItemT struct {
	Price    *string
	Quantity *float64
}

type billingItemT struct {
	Id                         UUID
	HasReachedCurrentPeriodCap bool
	Quantity                   *float64
	StripePriceId              *string
	BillingProduct             *billingProductT
}

type billingProductT struct {
	Name        string
	Description string
	Images      []string
	Metadata    *billingProductMetadataT
}

type billingProductMetadataT struct {
	ProductKey      *string
	PlanKey         *string
	PriceUsageBased *string
}

// ── currentUser resolver ─────────────────────────────────────────────────────

// CurrentUser loads the authenticated user, their workspace member, and the
// current workspace — enough for Twenty's app to boot. Served on /metadata.
func (r *MetadataResolver) CurrentUser(ctx context.Context) (*userT, error) {
	uid := userIDFromCtx(ctx)
	if uid == "" {
		return nil, errors.New("unauthenticated")
	}
	u, err := r.core.GetUserByID(ctx, uid)
	if err != nil {
		return nil, errors.New("user not found")
	}
	wsID, err := r.core.DefaultWorkspaceID(ctx, uid)
	if err != nil {
		return nil, err
	}

	displayName := "Labzyn"
	subdomain := "labzyn"
	if wsID != "" {
		if ws, err := r.core.GetWorkspace(ctx, wsID); err == nil {
			displayName = ws.DisplayName
			subdomain = ws.Subdomain
		}
	}

	member := workspaceMemberT{
		Id:          UUID(u.ID),
		Name:        fullNameT{FirstName: u.FirstName, LastName: u.LastName},
		ColorScheme: "Light",
		UserEmail:   u.Email,
		Locale:      ptr("en"),
	}
	uwID := UUID(u.ID)
	member.UserWorkspaceId = &uwID

	role := adminRole(UUID(wsID))
	ws := &workspaceT{
		Id:                        UUID(wsID),
		DisplayName:               &displayName,
		AllowImpersonation:        true,
		ActivationStatus:          "ACTIVE",
		IsPublicInviteLinkEnabled: true,
		IsPasswordAuthEnabled:     true,
		Subdomain:                 subdomain,
		InstalledApplications:     []*applicationT{},
		WorkspaceUrls:             workspaceUrlsT{SubdomainUrl: "http://localhost:3001"},
		FeatureFlags:              []*featureFlagT{},
		MetadataVersion:           1,
		BillingSubscriptions:      []*billingSubscriptionT{},
		BillingEntitlements:       []*billingEntitlementT{},
		WorkspaceMembersCount:     f64(1),
		DefaultRole:               &role,
		UseRecommendedModels:      true,
		EditableProfileFields:     []string{},
		EnabledAiModelIds:         []string{},
	}

	uw := &userWorkspaceT{
		Id:                                   UUID(u.ID),
		PermissionFlags:                      allPermissionFlags(),
		ObjectsPermissions:                   []*objectPermissionT{},
		TwoFactorAuthenticationMethodSummary: []*twoFactorSummaryT{},
	}

	return &userT{
		Id:                      UUID(u.ID),
		FirstName:               u.FirstName,
		LastName:                u.LastName,
		Email:                   u.Email,
		HasPassword:             true,
		OnboardingStatus:        ptr("COMPLETED"),
		WorkspaceMember:         &member,
		WorkspaceMembers:        []*workspaceMemberT{&member},
		DeletedWorkspaceMembers: []*deletedWorkspaceMemberT{},
		CurrentUserWorkspace:    uw,
		CurrentWorkspace:        ws,
		AvailableWorkspaces: availableWorkspacesT{
			AvailableWorkspacesForSignIn: []*availableWorkspaceT{},
			AvailableWorkspacesForSignUp: []*availableWorkspaceT{},
		},
		UserVars: &JSON{Value: map[string]any{}},
	}, nil
}

func adminRole(wsID UUID) roleT {
	return roleT{
		Id:                            wsID,
		Label:                         "Admin",
		CanUpdateAllSettings:          true,
		CanAccessAllTools:             true,
		IsEditable:                    false,
		CanReadAllObjectRecords:       true,
		CanUpdateAllObjectRecords:     true,
		CanSoftDeleteAllObjectRecords: true,
		CanDestroyAllObjectRecords:    true,
		CanBeAssignedToUsers:          true,
		CanBeAssignedToAgents:         true,
		CanBeAssignedToApiKeys:        true,
	}
}

func allPermissionFlags() []string {
	return []string{
		"WORKSPACE", "WORKSPACE_MEMBERS", "ROLES", "DATA_MODEL",
		"API_KEYS_AND_WEBHOOKS", "WORKFLOWS", "SECURITY", "ADMIN_PANEL",
		"IMPERSONATE", "SEND_EMAIL_INVITATIONS", "IMPORT_CSV", "EXPORT_CSV",
	}
}

func ptr(s string) *string   { return &s }
func f64(v float64) *float64 { return &v }
